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
const RendererContextKey = "renderer"

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
					here.onSSHConnection(s)
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

func (here *HereServer) onSSHConnection(s ssh.Session) {
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)

	if !ok || mappings == nil {
		log.Error("failed to parse mappings")
		fmt.Fprintf(s, `Usage: ssh %[1]s [OPTION]...

Expose local or internal network services via SSH remote forwarding.

Options:
  -R [NAME:]ID:HOST:PORT
        Expose a service through here.

        NAME        Optional custom subdomain name. If omitted, a random
                    subdomain is assigned.
        ID          Placeholder identifier. The value is ignored by the server
                    and does not need to be unique.
        HOST        Target hostname or IP address, resolved locally.
        PORT        Target port on the target host.

        This option may be specified multiple times to expose
        multiple services in a single session.

Examples:
  ssh %[1]s -R0:localhost:8080
        Expose a local service on port 8080.

  ssh %[1]s -R0:localhost:8080 -R1:localhost:9000
        Expose multiple local services.

  ssh %[1]s -R0:service.local:80
        Expose a service on an internal network host.

  ssh %[1]s -R myfancy-service:0:localhost:8080
        Expose a service using a custom subdomain.

`, here.getSSHDomain())

		s.Close()
	}
}

func (here *HereServer) onSSHClose(s ssh.Session) {
	renderer := bubbletea.MakeRenderer(s)
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)
	if !ok || mappings == nil {
		log.Error("failed to parse mappings on close session")
		s.Close()
		return
	}

	here.mappingsMu.Lock()
	for _, mapping := range mappings {
		delete(here.mappings, mapping.SourceSubdomain)
	}
	here.mappingsMu.Unlock()

	if s.Context().Value("timeout") != nil && here.isSSHCloseMessageEnabled() {
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
	}

	s.Close()
}
