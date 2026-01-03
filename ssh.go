package main

import (
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"

	"github.com/charmbracelet/lipgloss"
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
	if !ok || conn == nil {
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

	here.mappingsMu.Lock()
	_, idAlreadyExist := here.mappings[id]

	if idAlreadyExist {
		log.Info("id conflict found", "id", id)
		mapping.IsConflict = true
		mapping.Actual.IsPaused = true
	} else {
		here.mappings[id] = mappingModel
	}
	here.mappingsMu.Unlock()

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
	if !ok || mappings == nil {
		log.Error("failed to parse mappings on close session")
		return
	}

	here.mappingsMu.Lock()
	for _, mapping := range mappings {
		delete(here.mappings, mapping.SourceSubdomain)
	}
	here.mappingsMu.Unlock()

	if s.Context().Value("timeout") != nil {
		renderer := bubbletea.MakeRenderer(s)
		fmt.Fprint(s, renderer.NewStyle().
			Padding(2).
			Width(50).
			Background(lipgloss.Color("#212121")).
			Foreground(lipgloss.Color("#f7f784")).
			Render(fmt.Sprintf(`Your tunneling session has been terminated after %v because the user is not authenticated.

If you require additional time, login is necessary; however, sign-up is not available at the moment. We are planning to integrate the server with a web console to make it easier for users to obtain extended timeouts, but this may take some time to complete.

In the meantime, you are welcome to self-host the server to use it with an unlimited timeout. Please refer to the self-hosting guide here: https://github.com/pmh-only/here

thanks. -p`,
				here.getUnauthenticatedTimeout()))+"\n")
		return
	}
}
