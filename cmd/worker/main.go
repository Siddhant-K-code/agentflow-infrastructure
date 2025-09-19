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

	// Initialize worker
	worker, err := aor.NewWorker(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	// Start worker
	if err := worker.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	if err := worker.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}