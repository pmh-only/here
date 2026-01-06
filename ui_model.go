package main

import (
	"crypto/subtle"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

type MainUIListItem struct {
	title, desc string
}

func (i MainUIListItem) Title() string       { return i.title }
func (i MainUIListItem) Description() string { return i.desc }
func (i MainUIListItem) FilterValue() string { return i.title }

type MainUIModel struct {
	list           list.Model
	mappings       []MappingDisplayModel
	form           *huh.Form
	editMode       bool
	loginMode      bool
	editingIndex   int
	newSubdomain   string
	password       string
	authenticated  bool
	passwordNeeded bool
	session        ssh.Session
	renderer       *lipgloss.Renderer
	height         int
	width          int
}

func (m MainUIModel) Init() tea.Cmd {
	if m.passwordNeeded && !m.authenticated {
		return sessionTimeoutCmd(here.getUnauthenticatedTimeout())
	}
	return nil
}

func (m MainUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case sessionTimeoutMsg:
		if !m.authenticated && m.passwordNeeded {
			log.Info("Session timeout reached for unauthenticated user")
			m.session.Context().SetValue("timeout", true)
			return m, tea.Quit
		}
		return m, nil
	case tea.WindowSizeMsg:
		h, v := m.renderer.NewStyle().Margin(1, 2).GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
	case tea.KeyMsg:
		if !m.editMode && !m.loginMode && msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.editMode && !m.loginMode {
			switch msg.String() {
			case "l":
				if !m.passwordNeeded {
					return m, nil
				}
				if m.authenticated {
					statusMessage := m.renderer.NewStyle().
						Foreground(lipgloss.Color("#42f157")).
						Render("Already authenticated.")
					return m, m.list.NewStatusMessage(statusMessage)
				}

				m.loginMode = true
				m.password = ""
				m.form = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Key("password").
							Title("Login").
							Value(&m.password).
							EchoMode(huh.EchoModePassword).
							Placeholder("Enter password").
							WithTheme(createFormTheme(m.renderer)),
					).WithTheme(createFormTheme(m.renderer)),
				).WithWidth(60).WithTheme(createFormTheme(m.renderer))

				return m, m.form.Init()

			case "e":
				selectedIdx := m.list.Index()
				m.editingIndex = selectedIdx
				m.editMode = true

				m.newSubdomain = m.mappings[selectedIdx].SourceSubdomain
				m.form = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Key("subdomain").
							Title("Edit tunnel subdomain").
							Value(&m.newSubdomain).
							Placeholder("Enter new tunnel subdomain").
							WithTheme(createFormTheme(m.renderer)),
					).WithTheme(createFormTheme(m.renderer)),
				).WithWidth(60).WithTheme(createFormTheme(m.renderer))

				return m, m.form.Init()

			case "p":
				selectedIdx := m.list.Index()
				mapping := m.mappings[selectedIdx]

				if _, exist := here.mappings[mapping.SourceSubdomain]; mapping.IsConflict && exist {
					statusMessage := m.renderer.NewStyle().
						Foreground(lipgloss.Color("#f14242ff")).
						Render("this tunnel has subdomain conflict.")

					return m, m.list.NewStatusMessage(statusMessage)
				}

				m.mappings[selectedIdx].TogglePause()
				m.mappings[selectedIdx].IsConflict = false

				items := createTunnelIndex(m.mappings)
				listCmd := m.list.SetItems(items)
				statusCmd := m.list.NewStatusMessage(m.renderer.NewStyle().
					Foreground(lipgloss.Color("#42f157")).
					Render("Paused/Resumed " + m.mappings[selectedIdx].SourceSubdomain))

				return m, tea.Batch(listCmd, statusCmd)

			case "c":
				return m, tea.Batch(copyToClipboard(fmt.Sprint(
					here.getHostPrefix(),
					m.mappings[m.list.Index()].SourceSubdomain,
					here.getHostSuffix())), m.list.NewStatusMessage(m.renderer.NewStyle().
					Foreground(lipgloss.Color("#42f157")).
					Render("Copied Tunnel URL")))
			}
		}
	}

	if m.loginMode && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			password := m.form.GetString("password")
			log.Info("Login form completed")

			expectedPassword := here.getSSHPassword()
			if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1 {
				m.authenticated = true
				statusCmd := m.list.NewStatusMessage(m.renderer.NewStyle().
					Foreground(lipgloss.Color("#42f157")).
					Render("Login successful! Session timeout removed."))
				cmds = append(cmds, statusCmd)
			} else {
				statusCmd := m.list.NewStatusMessage(m.renderer.NewStyle().
					Foreground(lipgloss.Color("#f14242ff")).
					Render("Login failed: incorrect password"))
				cmds = append(cmds, statusCmd)
			}

			m.loginMode = false
			m.form = nil
			return m, tea.Batch(cmds...)
		}

		if m.form.State == huh.StateAborted {
			m.loginMode = false
			m.form = nil
			return m, tea.Batch(cmds...)
		}

		return m, tea.Batch(cmds...)
	}

	if m.editMode && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			newSubdomain := m.form.GetString("subdomain")
			log.Info("Form completed with value", "subdomain", newSubdomain)

			if newSubdomain != "" && newSubdomain != m.mappings[m.editingIndex].SourceSubdomain {
				m.mappings[m.editingIndex].RenameSubdomain(newSubdomain)
				items := createTunnelIndex(m.mappings)
				listCmd := m.list.SetItems(items)
				statusCmd := m.list.NewStatusMessage(m.renderer.NewStyle().
					Foreground(lipgloss.Color("#42f157")).
					Render("Renamed tunnel to: " + newSubdomain))
				cmds = append(cmds, listCmd, statusCmd)
			}

			m.editMode = false
			m.form = nil
			return m, tea.Batch(cmds...)
		}

		if m.form.State == huh.StateAborted {
			m.editMode = false
			m.form = nil
			return m, tea.Batch(cmds...)
		}

		return m, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// ---

type EmptyUIModel struct {
	renderer *lipgloss.Renderer
}

func (m EmptyUIModel) Init() tea.Cmd {
	return nil
}

func (m EmptyUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}
