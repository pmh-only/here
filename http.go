package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

func (here *HereServer) httpServerStart() {
	here.http = http.Server{}
	here.http.Addr = here.getHTTPAddr()
	here.http.Handler = http.HandlerFunc(here.onHTTPConnection)

	log.Printf("HereServer http server is listening on '%s'", here.getHTTPAddr())
	log.Fatal(here.http.ListenAndServe())
}

func (here *HereServer) onHTTPConnection(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(strings.TrimSpace(r.Host))
	host = strings.TrimSuffix(host, ".")
	host = strings.TrimSuffix(host, here.getHostSuffix())
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}

	here.mappingsMu.RLock()
	mapping, isExist := here.mappings[host]
	here.mappingsMu.RUnlock()

	if !isExist {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	payload := gossh.Marshal(&mapping.channel)
	ch, reqs, err := mapping.connection.OpenChannel("forwarded-tcpip", payload)

	if err != nil {
		log.Printf("Failed to open channel for %s: %v", host, err)

		// Remove stale mapping
		here.mappingsMu.Lock()
		delete(here.mappings, host)
		here.mappingsMu.Unlock()

		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	go gossh.DiscardRequests(reqs)

	conn := &sshNetConn{ch}

	ctx := r.Context()
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = host
			req.Host = r.Host

			// Extract client IP from RemoteAddr
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}

			// Set X-Real-IP header (immediate client)
			req.Header.Set("X-Real-IP", clientIP)

			// Handle X-Forwarded-For header
			if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
				// Append to existing X-Forwarded-For chain
				req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
			} else {
				// Create new X-Forwarded-For header
				req.Header.Set("X-Forwarded-For", clientIP)
			}

			// Set X-Forwarded-Proto header (protocol used by client)
			proto := "http"
			if r.TLS != nil {
				proto = "https"
			}
			req.Header.Set("X-Forwarded-Proto", proto)

			// Set X-Forwarded-Host header (original host requested by client)
			if r.Host != "" {
				req.Header.Set("X-Forwarded-Host", r.Host)
			}
		},
		Transport: &http.Transport{
			Proxy:               nil,
			DialContext:         oneShotDial(conn),
			ForceAttemptHTTP2:   false,
			DisableCompression:  false,
			MaxIdleConns:        1,
			MaxConnsPerHost:     1,
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		},
		ModifyResponse: nil,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, e error) {
			log.Println("proxy error:", e)
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
		},
		FlushInterval: 100 * time.Millisecond,
	}

	proxy.ServeHTTP(w, r)
}
