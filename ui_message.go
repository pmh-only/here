package main

import (
	"encoding/base64"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionTimeoutMsg struct{}

func sessionTimeoutCmd(timeout time.Duration) tea.Cmd {
	return tea.Tick(timeout, func(time.Time) tea.Msg {
		return sessionTimeoutMsg{}
	})
}

type clipboardMsg struct {
	success bool
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		osc52 := fmt.Sprintf("\033]52;c;%s\a", base64.StdEncoding.EncodeToString([]byte(text)))
		fmt.Print(osc52)

		return clipboardMsg{success: true}
	}
}
