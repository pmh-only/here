package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	os.Exit(0)
}
