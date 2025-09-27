package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	var (
		backendsFlag = flag.String("backends", "", "comma-separated backend addresses (host:port)")
		listen       = flag.String("listen", ":8080", "TCP listen address")
		strat        = flag.String("strategy", "rr", "load balancing strategy: rr | lc")
	)
	flag.Parse()

	if *backendsFlag == "" {
		fmt.Fprintln(os.Stderr, "error: -backends is required")
		flag.Usage()
		os.Exit(2)
	}

	addrs := splitAndTrim(*backendsFlag)
	var strategy Strategy
	switch *strat {
	case "rr":
		strategy = NewRoundRobinStrategy()
	case "lc":
		strategy = NewLeastConnStrategy()
	default:
		log.Fatalf("unknown strategy %q", *strat)
	}

	lb := NewTCPBalancer(addrs, strategy)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutting down...")
		lb.Close()
		os.Exit(0)
	}()

	log.Printf("LB TCP starting on %s, strategy=%T, backends=%v", *listen, strategy, addrs)
	if err := lb.ListenAndServe(*listen); err != nil {
		log.Fatalf("TCP LB failed: %v", err)
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
