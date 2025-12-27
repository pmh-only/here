package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gliderlabs/ssh"
)

type HereServer struct {
	ssh  ssh.Server
	http http.Server

	mappings       map[string]MappingModel
	mappingsMu     sync.RWMutex
	forwardHandler *ssh.ForwardedTCPHandler
}

func main() {
	srv := HereServer{}
	srv.mappings = map[string]MappingModel{}

	// Create context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start servers
	go srv.sshServerStart()
	go srv.httpServerStart()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.http.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown SSH server
	if err := srv.ssh.Shutdown(shutdownCtx); err != nil {
		log.Printf("SSH server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
