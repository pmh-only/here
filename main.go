package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

type HereServer struct {
	ssh  *ssh.Server
	http http.Server

	mappings       map[string]MappingModel
	mappingsMu     sync.RWMutex
	forwardHandler *ssh.ForwardedTCPHandler
}

var here = HereServer{}

func main() {
	here.mappings = map[string]MappingModel{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go here.sshServerStart()
	go here.httpServerStart()

	<-ctx.Done()

	log.Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := here.http.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", "err", err)
	}

	if here.ssh != nil {
		if err := here.ssh.Close(); err != nil {
			log.Error("SSH server shutdown error", "err", err)
		}
	}

	here.mappingsMu.Lock()
	here.mappings = nil
	here.mappingsMu.Unlock()

	log.Info("Shutdown complete")
	os.Exit(0)
}
