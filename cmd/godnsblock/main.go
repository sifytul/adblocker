package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"adblocker/internal/config"
	"adblocker/internal/server"
)

func main() {
	// Configuration
	//
	configFile := flag.String("config", "configs/config.yaml", "Path to configuration file")
	listenAddr := flag.String("listen", "", "Override listen address")
	showVersion := flag.Bool("version", false, "Show version and exit")

	flag.Parse()

	if *showVersion {
		log.Println("GoDNSBlock v0.3.0")
		return
	}

	log.Printf("Loading configuration from %s", *configFile)
	cfg, err := config.LoadFromFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		log.Println("Using default configuration")
		cfg = config.DefaultConfig()
	}

	if *listenAddr != "" {
		cfg.Server.ListenAddress = *listenAddr
	}
	// create DNS server
	dnsServer := server.NewDNSServer(cfg)
	
	// Load blocklists 
	log.Printf("Loading blocklist...")
	for _, source := range cfg.Blocklist.Sources {
		if err := dnsServer.LoadBlocklist(source); err != nil {
			log.Printf("Failed to load blocklist %s: %v", source, err)
		} else {
			log.Printf("Loaded blocklist: %s", source)
		}
	}

	// start server
	log.Printf("Starting DNS server on %s", cfg.Server.ListenAddress)

	if err := dnsServer.Start(); err != nil {
		log.Printf("Failed to start DNS server: %s", err)
	}

	log.Printf("DNS server is running. Press Ctrl+C to stop.")


	// Wait for interrupt signal to gracefully shutdown the server
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	dnsServer.Stop()
}