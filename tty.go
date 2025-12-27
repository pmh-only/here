package main

import (
	"fmt"
	"io"

	"github.com/gliderlabs/ssh"
)

// ANSI color codes
const (
	colorReset     = "\033[0m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorMagenta   = "\033[35m"
	colorCyan      = "\033[36m"
	colorWhite     = "\033[37m"
	colorBold      = "\033[1m"
	colorBoldGreen = "\033[1;32m"
	colorBoldBlue  = "\033[1;34m"
	colorBoldCyan  = "\033[1;36m"
)

// Helper functions for colored output
func colorize(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, colorReset)
}

func colorBoldText(color, text string) string {
	return fmt.Sprintf("%s%s%s%s", colorBold, color, text, colorReset)
}

func writeSuccess(s ssh.Session, message string) {
	io.WriteString(s, colorize(colorBoldGreen, "✓ "+message)+"\n")
}

func writeError(s ssh.Session, message string) {
	io.WriteString(s, colorize(colorRed, "✗ "+message)+"\n")
}

func writeInfo(s ssh.Session, icon, message string) {
	io.WriteString(s, colorize(colorYellow, icon+"  "+message)+"\n")
}

func writePrompt(s ssh.Session, message string) {
	io.WriteString(s, colorBoldText(colorCyan, message))
}

// readLine reads from SSH session until newline or error, with echo
func readLine(s ssh.Session) (string, error) {
	var line []byte
	buf := make([]byte, 1)

	for {
		n, err := s.Read(buf)
		if err != nil {
			return "", err
		}

		if n > 0 {
			// Handle Ctrl+C (ETX - 0x03) - interrupt
			if buf[0] == 3 {
				s.Write([]byte(fmt.Sprintf("%s^C%s\r\n", colorRed, colorReset)))
				return "", fmt.Errorf("interrupted by user (Ctrl+C)")
			}
			// Handle Ctrl+D (EOT - 0x04) - end of transmission
			if buf[0] == 4 {
				s.Write([]byte(fmt.Sprintf("%s^D%s\r\n", colorYellow, colorReset)))
				return "", fmt.Errorf("terminated by user (Ctrl+D)")
			}
			// Check for newline (Enter key) - handle both \r and \n
			if buf[0] == '\n' || buf[0] == '\r' {
				// Echo newline for proper display
				s.Write([]byte{'\r', '\n'})
				break
			}
			// Handle backspace (0x7f or 0x08)
			if buf[0] == 127 || buf[0] == 8 {
				if len(line) > 0 {
					line = line[:len(line)-1]
					// Echo backspace: move back, space, move back
					s.Write([]byte{8, 32, 8})
				}
				continue
			}
			// Echo the character back to the user
			s.Write(buf[:1])
			line = append(line, buf[0])
		}
	}

	return string(line), nil
}

// readPassword reads from SSH session until newline, WITHOUT echoing (for passwords)
func readPassword(s ssh.Session) (string, error) {
	var line []byte
	buf := make([]byte, 1)

	for {
		n, err := s.Read(buf)
		if err != nil {
			return "", err
		}

		if n > 0 {
			// Handle Ctrl+C (ETX - 0x03) - interrupt
			if buf[0] == 3 {
				s.Write([]byte(fmt.Sprintf("%s^C%s\r\n", colorRed, colorReset)))
				return "", fmt.Errorf("interrupted by user (Ctrl+C)")
			}
			// Handle Ctrl+D (EOT - 0x04) - end of transmission
			if buf[0] == 4 {
				s.Write([]byte(fmt.Sprintf("%s^D%s\r\n", colorYellow, colorReset)))
				return "", fmt.Errorf("terminated by user (Ctrl+D)")
			}
			// Check for newline (Enter key) - handle both \r and \n
			if buf[0] == '\n' || buf[0] == '\r' {
				// Echo newline for proper display
				s.Write([]byte{'\r', '\n'})
				break
			}
			// Handle backspace
			if buf[0] == 127 || buf[0] == 8 {
				if len(line) > 0 {
					line = line[:len(line)-1]
				}
				continue
			}
			// Don't echo - just collect the character
			line = append(line, buf[0])
		}
	}

	return string(line), nil
}
