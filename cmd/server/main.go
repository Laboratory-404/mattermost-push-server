package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Laboratory-404/mattermost-push-server/pkg/config"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push/apns"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push/fcm"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push/web"
	"github.com/Laboratory-404/mattermost-push-server/pkg/server"
)

var (
	version = "2.0.0"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to YAML configuration file")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Mattermost Push Proxy Server v%s\n", version)
		os.Exit(0)
	}

	log.Printf("==================================================")
	log.Printf(" Mattermost Push Proxy Server v%s starting...", version)
	log.Printf("==================================================")

	// Load Configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("[Main] Warning loading config file '%s': %v. Using defaults/env vars.", *configPath, err)
		cfg = config.DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("[Main] Configuration validation error: %v", err)
	}

	// Initialize Providers
	fcmProv, err := fcm.NewProvider(cfg.FCM)
	if err != nil {
		log.Printf("[Main] Warning initializing FCM provider: %v", err)
	}

	apnsProv, err := apns.NewProvider(cfg.APNS)
	if err != nil {
		log.Printf("[Main] Warning initializing APNs provider: %v", err)
	}

	webProv := web.NewProvider()

	// Build Push Dispatcher Service
	pushService := push.NewService(fcmProv, apnsProv, webProv)

	// Build & Start HTTP Server
	srv := server.NewServer(cfg, pushService)

	// Listen for OS signals for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("[Main] Server error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-shutdownChan
	log.Printf("[Main] Received signal '%s', initiating shutdown sequence...", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("[Main] Server shutdown error: %v", err)
	} else {
		log.Printf("[Main] Server stopped cleanly.")
	}
}
