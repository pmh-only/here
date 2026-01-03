package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m MainUIModel) View() string {
	if m.loginMode {
		m.list.SetSize(m.width, m.height-strings.Count(m.form.View(), "\n")-2)
		return m.docStyle.Render(m.list.View() + "\n\n" + m.form.View())
	}

	if m.editMode {
		m.list.SetSize(m.width, m.height-strings.Count(m.form.View(), "\n")-2)
		return m.docStyle.Render(m.list.View() + "\n\n" + m.form.View())
	}

	m.list.SetSize(m.width, m.height)
	login := ""

	if !m.authenticated && here.isPasswordRequired() {
		m.list.SetSize(m.width, m.height-1)
		login += "\n  " + m.renderer.
			NewStyle().
			Background(lipgloss.Color("#212121")).
			Foreground(lipgloss.Color("#f7f784")).
			PaddingLeft(1).
			PaddingRight(1).
			Render(fmt.Sprintf("Note: Unauthenticated session will timeout after %v", here.getUnauthenticatedTimeout()))
	}

	return m.docStyle.Render(m.list.View() + login)
}
