package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"adblocker/internal/config"
	"adblocker/internal/logger"
	"adblocker/internal/server"
)

func main() {
    // Parse command-line flags
    configFile := flag.String("config", "configs/config.yaml", "Path to configuration file")
    listenAddr := flag.String("listen", "", "Override listen address")
    showVersion := flag.Bool("version", false, "Show version and exit")
    logLevel := flag.String("log-level", "", "Override log level")
    
    flag.Parse()
    
    if *showVersion {
        fmt.Println("GoDNSBlock v0.3.0")
        return
    }
    
    // Load configuration
    fmt.Printf("Loading configuration from %s\n", *configFile)
    cfg, err := config.LoadFromFile(*configFile)
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        fmt.Println("Using default configuration")
        cfg = config.DefaultConfig()
    }
    
    // Override with flags
    if *listenAddr != "" {
        cfg.Server.ListenAddress = *listenAddr
    }
    if *logLevel != "" {
        cfg.Logging.Level = *logLevel
    }
    
    // Initialize logger
    log, err := logger.NewLogger(cfg.Logging.Level, cfg.Logging.OutputFile)
    if err != nil {
        fmt.Printf("Failed to create logger: %v\n", err)
        os.Exit(1)
    }
    
    log.Info("GoDNSBlock v0.3.0 starting...")
    log.Info("Configuration loaded successfully")
    
    // Create DNS server
    dnsServer := server.NewDNSServer(cfg, log)
    
    // Load blocklists
    log.Info("Loading blocklists...")
    for _, source := range cfg.Blocklist.Sources {
        if err := dnsServer.LoadBlocklist(source); err != nil {
            log.Error("Failed to load blocklist %s: %v", source, err)
        } else {
            log.Info("Loaded blocklist: %s", source)
        }
    }
    
    // Start server
    log.Info("Starting DNS server on %s", cfg.Server.ListenAddress)
    if err := dnsServer.Start(); err != nil {
        log.Error("Failed to start server: %v", err)
        os.Exit(1)
    }
    
    log.Info("DNS server is running. Press Ctrl+C to stop.")
    
    // Start statistics ticker (print stats every 60 seconds)
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for range ticker.C {
            stats := dnsServer.GetStats()
            log.Info("Statistics: %s", stats.String())
        }

        // Cache statistics (if enabled)
        if cfg.Cache.Enabled {
            cacheStats := dnsServer.GetCacheStats()
            log.Info("Cache statistics: %s", cacheStats.String())
        }
    }()
    
    // Wait for interrupt signal
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    
    // Shutdown
    ticker.Stop()
    log.Info("Shutting down...")
    
    // Print final statistics
    finalStats := dnsServer.GetStats()
    log.Info("Final statistics: %s", finalStats.String())

    if cfg.Cache.Enabled {
        finalCacheStats := dnsServer.GetCacheStats()
        log.Info("Final cache stats: %s", finalCacheStats.String())
    }
    
    dnsServer.Stop()
    log.Info("Shutdown complete")
}