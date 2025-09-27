package main

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Backend struct {
	Address     string
	alive       bool
	activeConns int64
	mu          sync.Mutex
}

func (b *Backend) SetAlive(a bool) {
	b.mu.Lock()
	b.alive = a
	b.mu.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.alive
}

func (b *Backend) IncConns() { atomic.AddInt64(&b.activeConns, 1) }
func (b *Backend) DecConns() { atomic.AddInt64(&b.activeConns, -1) }
func (b *Backend) ActiveConns() int64 { return atomic.LoadInt64(&b.activeConns) }

type TCPBalancer struct {
	backends []*Backend
	strategy Strategy
	rrCounter uint64
	mu        sync.RWMutex
	listener  net.Listener
}

func NewTCPBalancer(addrs []string, strategy Strategy) *TCPBalancer {
	b := &TCPBalancer{strategy: strategy}
	for _, addr := range addrs {
		b.backends = append(b.backends, &Backend{Address: addr, alive: true})
	}
	return b
}

func (lb *TCPBalancer) ListenAndServe(listenAddr string) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	lb.listener = ln
	for {
		clientConn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		backend := lb.strategy.Select(lb)
		if backend == nil {
			log.Println("no backends available")
			clientConn.Close()
			continue
		}
		go lb.handleConnection(clientConn, backend)
	}
}

func (lb *TCPBalancer) handleConnection(clientConn net.Conn, backend *Backend) {
	backend.IncConns()
	defer backend.DecConns()

	serverConn, err := net.Dial("tcp", backend.Address)
	if err != nil {
		log.Printf("failed to connect to backend %s: %v", backend.Address, err)
		clientConn.Close()
		return
	}

	// Bidirectional copy
	// Why close the destination, not the source?

	// io.Copy(dst, src) reads from src and writes to dst until EOF or error.
	// Closing the destination (serverConn or clientConn) signals that no more data will be written to it. 
	// This is important for protocols like TCP, where closing the write side sends a FIN to the peer.
	// If you close the source (clientConn or serverConn) inside the goroutine, you risk interrupting the other goroutine 
	// that may still be reading from it. This could cause partial reads or errors.
	// Analogy:
	// Imagine two people passing notes through a tube. If you close the tube on the sending side, the receiver can't get all the notes. 
	// But if you close the tube on the receiving side after all notes are received, communication ends cleanly.

	// Gotcha:
	// If you closed both ends in both goroutines, you could prematurely terminate the data transfer.

	// Summary:

	// Close the destination after copying to signal end of writing.
	// Leave the source open until all copying is done to avoid interrupting the other direction.

    var wg sync.WaitGroup
    wg.Add(2)
    go func() {
        io.Copy(serverConn, clientConn)
        serverConn.Close()
        wg.Done()
    }()
    go func() {
        io.Copy(clientConn, serverConn)
        clientConn.Close()
        wg.Done()
    }()
    wg.Wait()
}

func (lb *TCPBalancer) GetAliveBackends() []*Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	alive := make([]*Backend, 0, len(lb.backends))
	for _, b := range lb.backends {
		if b.IsAlive() {
			alive = append(alive, b)
		}
	}
	return alive
}

func (lb *TCPBalancer) Close() {
	if lb.listener != nil {
		lb.listener.Close()
	}
}
