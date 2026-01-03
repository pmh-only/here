package main

import (
	"math/rand/v2"
	"net"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	gossh "golang.org/x/crypto/ssh"
)

const MappingContextKey = "mappings"

func (here *HereServer) sshServerStart() {
	server, err := wish.NewServer(
		wish.WithAddress(here.getSSHAddr()),
		wish.WithHostKeyPEM(here.getSSHHostKey()),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					next(s)
					here.onSSHClose(s)
				}
			},
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	here.ssh = server
	here.forwardHandler = &ssh.ForwardedTCPHandler{}
	here.ssh.RequestHandlers = map[string]ssh.RequestHandler{
		"tcpip-forward": here.onSSHForwardRequest,
	}

	log.Info("HereServer ssh server is listening", "Addr", here.getSSHAddr())
	log.Fatal(here.ssh.ListenAndServe())
}

func (here *HereServer) onSSHForwardRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	var payload RemoteForwardRequest
	if err := gossh.Unmarshal(req.Payload, &payload); err != nil {
		return false, []byte{}
	}

	log.Info("Received Forward Request", "BindAdr", payload.BindAddr, "BindPort", payload.BindPort)
	assignedPort := payload.BindPort
	if assignedPort == 0 {
		assignedPort = rand.Uint32N(10000) + 10000
	}

	originAddr, orignPortStr, err := net.SplitHostPort(ctx.RemoteAddr().String())
	if err != nil {
		log.Error("Failed to parse remote address", err)
		return false, []byte{}
	}

	originPort, err := strconv.Atoi(orignPortStr)
	if err != nil {
		log.Error("Failed to parse port number", err)
		return false, []byte{}
	}

	conn, ok := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		log.Error("Failed to get SSH connection from context")
		return false, []byte{}
	}

	mappingModel := MappingModel{
		conn,
		RemoteForwardChannel{
			DestAddr:   payload.BindAddr,
			DestPort:   assignedPort,
			OriginAddr: originAddr,
			OriginPort: uint32(originPort),
		},
		false,
	}

	id := randStringRunes(10)
	if payload.BindAddr != "localhost" {
		id = payload.BindAddr
	}

	mapping := MappingDisplayModel{
		SourceSubdomain: id,
		TargetPort:      payload.BindPort,
		Actual:          mappingModel,
	}

	_, idAlreadyExist := here.mappings[id]

	if idAlreadyExist {
		log.Info("id conflict found", "id", id)
		mapping.IsConflict = true
		mapping.Actual.IsPaused = true
	}

	if !idAlreadyExist {
		here.mappingsMu.Lock()
		here.mappings[id] = mappingModel
		here.mappingsMu.Unlock()
	}

	log.Info("Assigned port for forward request", "port", assignedPort)

	mappings, ok := ctx.Value(MappingContextKey).([]MappingDisplayModel)
	if !ok || mappings == nil {
		mappings = []MappingDisplayModel{}
	}

	mappings = append(mappings, mapping)

	ctx.SetValue(MappingContextKey, mappings)

	return true, gossh.Marshal(RemoteForwardSuccess{
		assignedPort,
	})
}

func (here *HereServer) onSSHClose(s ssh.Session) {
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)
	if !ok {
		log.Error("failed to parse mappings on close session")
	}

	here.mappingsMu.Lock()
	for _, mapping := range mappings {
		delete(here.mappings, mapping.SourceSubdomain)
	}
	here.mappingsMu.Unlock()
}
