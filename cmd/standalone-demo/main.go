package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/demo"
)

func main() {
	// Create standalone demo server
	server := demo.NewStandaloneDemoServer()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down demo server...")
		os.Exit(0)
	}()

	// Start the server
	log.Println("Starting AgentFlow Standalone Demo Server...")
	if err := server.Start(8080); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
