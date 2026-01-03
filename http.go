package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	gossh "golang.org/x/crypto/ssh"
)

func (here *HereServer) httpServerStart() {
	here.http = http.Server{}
	here.http.Addr = here.getHTTPAddr()
	here.http.Handler = http.HandlerFunc(here.onHTTPConnection)

	log.Info("HereServer http server is listening", "Addr", here.getHTTPAddr())
	log.Fatal(here.http.ListenAndServe())
}

func (here *HereServer) onHTTPConnection(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(strings.TrimSpace(r.Host))
	host = strings.TrimSuffix(host, ".")
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}

	host = strings.TrimSuffix(host, here.getHostSuffix())

	here.mappingsMu.RLock()
	mapping, isExist := here.mappings[host]
	here.mappingsMu.RUnlock()

	if !isExist {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if mapping.IsPaused {
		http.Error(w, "Tunnel Paused", http.StatusServiceUnavailable)
		return
	}

	payload := gossh.Marshal(mapping.Actual)
	ch, reqs, err := mapping.Connection.OpenChannel("forwarded-tcpip", payload)

	if err != nil {
		log.Error("Failed to open channel for %s: %v", host, err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		gossh.DiscardRequests(reqs)
		cancel()
	}()

	conn := &sshNetConn{ch}

	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	transport := &http.Transport{
		Proxy:               nil,
		DialContext:         oneShotDial(conn),
		ForceAttemptHTTP2:   false,
		DisableCompression:  false,
		MaxIdleConns:        1,
		MaxConnsPerHost:     1,
		MaxIdleConnsPerHost: 1,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
	}
	defer transport.CloseIdleConnections()

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = host
		},
		Transport:      transport,
		ModifyResponse: nil,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Error(err)
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
		},
		FlushInterval: 100 * time.Millisecond,
	}

	proxy.ServeHTTP(w, r)
}
