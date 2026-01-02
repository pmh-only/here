package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go srv.sshServerStart()
	go srv.httpServerStart()

	<-ctx.Done()
	os.Exit(0)
}
