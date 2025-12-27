package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (here *HereServer) sshServerStart() {
	here.ssh = ssh.Server{}

	hostKey, err := here.getSSHHostKey()
	if err != nil {
		log.Fatal(err)
	}

	here.ssh.AddHostKey(hostKey)
	here.ssh.Addr = here.getSSHAddr()
	here.ssh.Handler = here.onSSHConnection

	here.forwardHandler = &ssh.ForwardedTCPHandler{}
	here.ssh.RequestHandlers = map[string]ssh.RequestHandler{
		"tcpip-forward": here.onSSHForwardRequest,
	}

	log.Printf("HereServer ssh server is listening on '%s'", here.getSSHAddr())
	log.Fatal(here.ssh.ListenAndServe())
}

func (here *HereServer) onSSHForwardRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	var payload RemoteForwardRequest
	if err := gossh.Unmarshal(req.Payload, &payload); err != nil {
		return false, []byte{}
	}

	log.Printf("Received Forward Request: %s:%d", payload.BindAddr, payload.BindPort)

	originAddr, orignPortStr, err := net.SplitHostPort(ctx.RemoteAddr().String())
	if err != nil {
		log.Printf("Failed to parse remote address: %v", err)
		return false, []byte{}
	}

	originPort, err := strconv.Atoi(orignPortStr)
	if err != nil {
		log.Printf("Failed to parse port number: %v", err)
		return false, []byte{}
	}

	remoteForwardChannelsRaw := ctx.Value("RemoteForwardChannels")
	if remoteForwardChannelsRaw == nil {
		remoteForwardChannelsRaw = []RemoteForwardChannel{}
	}

	remoteForwardChannels, ok := remoteForwardChannelsRaw.([]RemoteForwardChannel)
	if !ok {
		log.Println("Invalid RemoteForwardChannels type in context")
		return false, []byte{}
	}

	remoteForwardChannels = append(remoteForwardChannels, RemoteForwardChannel{
		DestAddr:   payload.BindAddr,
		DestPort:   payload.BindPort,
		OriginAddr: originAddr,
		OriginPort: uint32(originPort),
	})

	ctx.SetValue("RemoteForwardChannels", remoteForwardChannels)

	return true, gossh.Marshal(RemoteForwardSuccess{
		payload.BindPort,
	})
}

func (here *HereServer) onSSHConnection(s ssh.Session) {
	remoteForwardChannelsRaw := s.Context().Value("RemoteForwardChannels")
	if remoteForwardChannelsRaw == nil {
		remoteForwardChannelsRaw = []RemoteForwardChannel{}
	}

	remoteForwardChannels, ok := remoteForwardChannelsRaw.([]RemoteForwardChannel)
	if !ok {
		log.Println("Invalid RemoteForwardChannels type")
		io.WriteString(s, "Internal server error\n")
		return
	}

	io.WriteString(s, "Welcome to HereServer!\n")
	io.WriteString(s, fmt.Sprintf("Args: %s\n\n", strings.Join(s.Command(), ", ")))

	ids := here.registerForwards(s.Context(), remoteForwardChannels)
	io.WriteString(s, fmt.Sprintf("You requested %d service(s):\n", len(remoteForwardChannels)))

	for i, remoteForwardChannels := range remoteForwardChannels {
		io.WriteString(s, fmt.Sprintf("%d -> %s\n", remoteForwardChannels.DestPort, ids[i]))
	}

	<-s.Context().Done()

	log.Printf("Session context finished, cleaning up %d mappings", len(ids))

	// Clean up mappings when session ends (prevents memory leak)
	here.mappingsMu.Lock()
	for _, id := range ids {
		delete(here.mappings, id)
	}
	here.mappingsMu.Unlock()
}

func (here *HereServer) registerForwards(ctx ssh.Context, remoteForwardChannels []RemoteForwardChannel) []string {
	conn, ok := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		log.Println("Failed to get SSH connection from context")
		return []string{}
	}

	ids := []string{}

	for _, remoteForwardChannels := range remoteForwardChannels {
		id, err := uuid.NewV7()
		if err != nil {
			log.Printf("Failed to generate UUID: %v", err)
			continue
		}

		ids = append(ids, id.String())

		here.mappingsMu.Lock()
		here.mappings[id.String()] = MappingModel{
			conn,
			remoteForwardChannels,
		}
		here.mappingsMu.Unlock()
	}

	return ids
}
