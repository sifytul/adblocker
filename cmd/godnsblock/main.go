package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"adblocker/internal/server"
)

func main() {
	// Configuration
	listenAddr := "0.0.0.0:5335"
	upstreamDNS := "8.8.8.8:53"

	dnsServer := server.NewDNSServer(listenAddr, upstreamDNS)

	if err := dnsServer.Start(); err != nil {
		log.Fatalf("Failed to start DNS server: %s", err)
	}

	log.Printf("DNS server is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal to gracefully shutdown the server
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	dnsServer.Stop()


}