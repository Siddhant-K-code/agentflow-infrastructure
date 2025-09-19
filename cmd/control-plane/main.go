package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/aor"
	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize control plane
	cp, err := aor.NewControlPlane(cfg)
	if err != nil {
		log.Fatalf("Failed to create control plane: %v", err)
	}

	// Start control plane
	if err := cp.Start(ctx); err != nil {
		log.Fatalf("Failed to start control plane: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down control plane...")
	if err := cp.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
