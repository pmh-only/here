package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
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

	assignedPort := payload.BindPort
	if assignedPort == 0 {
		assignedPort = rand.Uint32N(10000) + 10000
	}

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
		remoteForwardChannelsRaw = []OverrideModel{}
	}

	remoteForwardChannels, ok := remoteForwardChannelsRaw.([]OverrideModel)
	if !ok {
		log.Println("Invalid RemoteForwardChannels type in context")
		return false, []byte{}
	}

	remoteForwardChannels = append(remoteForwardChannels, OverrideModel{
		payload.BindPort,
		RemoteForwardChannel{
			DestAddr:   payload.BindAddr,
			DestPort:   assignedPort,
			OriginAddr: originAddr,
			OriginPort: uint32(originPort),
		},
	})

	ctx.SetValue("RemoteForwardChannels", remoteForwardChannels)

	log.Printf("Assigned port %d for forward request", assignedPort)

	return true, gossh.Marshal(RemoteForwardSuccess{
		assignedPort,
	})
}

// displayWelcomeBanner shows the welcome message
func displayWelcomeBanner(s ssh.Session) {
	io.WriteString(s, fmt.Sprintf("\n%s%s╔═══════════════════════════════╗%s\n", colorBold, colorCyan, colorReset))
	io.WriteString(s, fmt.Sprintf("%s%s║   Welcome to HereServer!      ║%s\n", colorBold, colorCyan, colorReset))
	io.WriteString(s, fmt.Sprintf("%s%s╚═══════════════════════════════╝%s\n\n", colorBold, colorCyan, colorReset))
}

// selectMode prompts user to choose between anonymous or login mode
func (here *HereServer) selectMode(s ssh.Session) (authenticated bool, err error) {
	timeoutDuration := here.getUnauthenticatedTimeout()

	io.WriteString(s, colorBoldText(colorYellow, "Select mode:")+"\n")
	io.WriteString(s, fmt.Sprintf("  %s1)%s Anonymous (%s%v%s session timeout)\n",
		colorBoldCyan, colorReset, colorYellow, timeoutDuration, colorReset))
	io.WriteString(s, fmt.Sprintf("  %s2)%s Login (%sunlimited%s session, requires password)\n",
		colorBoldCyan, colorReset, colorGreen, colorReset))
	io.WriteString(s, fmt.Sprintf("\n%sEnter choice (1 or 2):%s ", colorBold, colorReset))

	choice, err := readLine(s)
	if err != nil {
		log.Printf("Mode selection cancelled from %s: %v", s.RemoteAddr(), err)
		writeError(s, "Session cancelled")
		return false, err
	}

	choice = strings.TrimSpace(choice)

	switch choice {
	case "2":
		return here.authenticateUser(s)
	case "1":
		log.Printf("User from %s selected anonymous mode", s.RemoteAddr())
		writeSuccess(s, "Anonymous mode selected\n")
		return false, nil
	default:
		log.Printf("Invalid mode choice '%s' from %s", choice, s.RemoteAddr())
		writeError(s, "Invalid choice. Please select 1 or 2.")
		return false, fmt.Errorf("invalid choice")
	}
}

// authenticateUser handles password authentication
func (here *HereServer) authenticateUser(s ssh.Session) (bool, error) {
	writePrompt(s, "\nPassword: ")

	password, err := readPassword(s)
	if err != nil {
		log.Printf("Password entry cancelled from %s: %v", s.RemoteAddr(), err)
		writeError(s, "Authentication cancelled")
		return false, err
	}

	password = strings.TrimSpace(password)

	if password != here.getSSHPassword() {
		log.Printf("Invalid password attempt from %s", s.RemoteAddr())
		writeError(s, "Authentication failed")
		return false, fmt.Errorf("invalid password")
	}

	log.Printf("Successful authentication from %s", s.RemoteAddr())
	writeSuccess(s, "Authentication successful!\n")
	return true, nil
}

// displayServiceInfo shows the registered tunnels
func (here *HereServer) displayServiceInfo(s ssh.Session, ids []string, channels []OverrideModel) {
	io.WriteString(s, fmt.Sprintf("%s%sYou requested %d service(s):%s\n",
		colorBold, colorMagenta, len(channels), colorReset))

	for i, channel := range channels {
		io.WriteString(s, fmt.Sprintf("  %s#%d%s (-R%d) -> %s%s%s%s%s\n",
			colorBoldBlue, i, colorReset,
			channel.DestPort,
			colorBoldGreen,
			here.getHostPerfix(),
			ids[i],
			here.getHostSuffix(),
			colorReset))
	}
}

// monitorInput watches for Ctrl+C/Ctrl+D during active tunnel
func monitorInput(s ssh.Session) <-chan error {
	inputDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := s.Read(buf)
			if err != nil {
				inputDone <- err
				return
			}
			if n > 0 {
				if buf[0] == 3 { // Ctrl+C
					s.Write([]byte(fmt.Sprintf("\r\n%s^C%s\r\n", colorRed, colorReset)))
					inputDone <- fmt.Errorf("interrupted by user (Ctrl+C)")
					return
				}
				if buf[0] == 4 { // Ctrl+D
					s.Write([]byte(fmt.Sprintf("\r\n%s^D%s\r\n", colorYellow, colorReset)))
					inputDone <- fmt.Errorf("terminated by user (Ctrl+D)")
					return
				}
			}
		}
	}()
	return inputDone
}

// handleSessionTimeout manages timeout for authenticated/anonymous sessions
func (here *HereServer) handleSessionTimeout(s ssh.Session, authenticated bool, inputDone <-chan error) {
	if !authenticated {
		timeoutDuration := here.getUnauthenticatedTimeout()
		writeInfo(s, "\n⏱", fmt.Sprintf("Note: Anonymous session will timeout after %s%v%s",
			colorBold, timeoutDuration, colorReset))
		io.WriteString(s, colorize(colorCyan, "⌨  Press Ctrl+C or Ctrl+D to exit")+"\n")
		log.Printf("Anonymous session from %s will timeout after %v", s.RemoteAddr(), timeoutDuration)

		timeoutTimer := time.NewTimer(timeoutDuration)
		defer timeoutTimer.Stop()

		select {
		case <-s.Context().Done():
			log.Printf("Session from %s ended before timeout", s.RemoteAddr())
		case <-timeoutTimer.C:
			log.Printf("Session from %s timed out after %v", s.RemoteAddr(), timeoutDuration)
			writeInfo(s, "⏱", fmt.Sprintf("Session timed out after %v", timeoutDuration))
			s.Close()
		case err := <-inputDone:
			log.Printf("Session from %s closed by user: %v", s.RemoteAddr(), err)
			writeSuccess(s, "Session closed")
			s.Close()
		}
	} else {
		io.WriteString(s, fmt.Sprintf("\n%s✓ Authenticated session%s - %sno timeout%s\n",
			colorBoldGreen, colorReset, colorGreen, colorReset))
		io.WriteString(s, colorize(colorCyan, "⌨  Press Ctrl+C or Ctrl+D to exit")+"\n")
		log.Printf("Authenticated session from %s has no timeout", s.RemoteAddr())

		select {
		case <-s.Context().Done():
			log.Printf("Session from %s ended", s.RemoteAddr())
		case err := <-inputDone:
			log.Printf("Session from %s closed by user: %v", s.RemoteAddr(), err)
			writeSuccess(s, "Session closed")
			s.Close()
		}
	}
}

func (here *HereServer) onSSHConnection(s ssh.Session) {
	// Get remote forward channels from context
	remoteForwardChannelsRaw := s.Context().Value("RemoteForwardChannels")
	if remoteForwardChannelsRaw == nil {
		remoteForwardChannelsRaw = []OverrideModel{}
	}

	remoteForwardChannels, ok := remoteForwardChannelsRaw.([]OverrideModel)
	if !ok {
		log.Println("Invalid RemoteForwardChannels type")
		writeError(s, "Internal server error")
		return
	}

	// Display welcome banner
	displayWelcomeBanner(s)

	// Handle authentication/mode selection if password is configured
	authenticated := false
	if here.isPasswordRequired() {
		var err error
		authenticated, err = here.selectMode(s)
		if err != nil {
			return
		}
	}

	// Register forwards and display service information
	ids := here.registerForwards(s.Context(), remoteForwardChannels)
	here.displayServiceInfo(s, ids, remoteForwardChannels)

	// Monitor for Ctrl+C/Ctrl+D and handle session timeout
	inputDone := monitorInput(s)
	here.handleSessionTimeout(s, authenticated, inputDone)

	// Cleanup mappings when session ends
	log.Printf("Session context finished, cleaning up %d mappings", len(ids))
	here.mappingsMu.Lock()
	for _, id := range ids {
		delete(here.mappings, id)
	}
	here.mappingsMu.Unlock()
}

func (here *HereServer) registerForwards(ctx ssh.Context, remoteForwardChannels []OverrideModel) []string {
	conn, ok := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		log.Println("Failed to get SSH connection from context")
		return []string{}
	}

	ids := []string{}

	for _, remoteForwardChannels := range remoteForwardChannels {
		id := randStringRunes(10)
		if remoteForwardChannels.Override.DestAddr != "localhost" {
			id = remoteForwardChannels.Override.DestAddr
		}

		ids = append(ids, id)

		here.mappingsMu.Lock()
		here.mappings[id] = MappingModel{
			conn,
			remoteForwardChannels.Override,
		}
		here.mappingsMu.Unlock()
	}

	return ids
}
