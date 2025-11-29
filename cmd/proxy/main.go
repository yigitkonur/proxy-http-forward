// Package main is the entry point for the go-native-squid-proxy.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yigitkonur/go-native-squid-proxy/pkg/config"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/log"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/proxy"
)

var (
	// Version information (set via ldflags)
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("go-native-squid-proxy %s\n", version)
		fmt.Printf("  commit:  %s\n", commit)
		fmt.Printf("  built:   %s\n", buildDate)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := log.New(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Infow("starting go-native-squid-proxy",
		"version", version,
		"commit", commit,
		"build_date", buildDate,
	)

	// Create and start the proxy server
	server := proxy.New(cfg, sugar)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-quit:
		sugar.Infow("received shutdown signal", "signal", sig.String())
	case err := <-errChan:
		sugar.Errorw("server error", "error", err)
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.ShutdownWithContext(ctx); err != nil {
		sugar.Errorw("shutdown error", "error", err)
		os.Exit(1)
	}

	sugar.Info("server stopped gracefully")
}
