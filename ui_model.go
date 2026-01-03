package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type MainUIListItem struct {
	title, desc string
}

func (i MainUIListItem) Title() string       { return i.title }
func (i MainUIListItem) Description() string { return i.desc }
func (i MainUIListItem) FilterValue() string { return i.title }

type MainUIModel struct {
	list         list.Model
	mappings     []MappingDisplayModel
	form         *huh.Form
	editMode     bool
	editingIndex int
	newSubdomain string
}

func (m MainUIModel) Init() tea.Cmd {
	return nil
}

func (m MainUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle window size for both list and form
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if !m.editMode && msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if !m.editMode {
			switch msg.String() {
			case "e":
				selectedIdx := m.list.Index()
				m.editingIndex = selectedIdx
				m.editMode = true

				m.newSubdomain = m.mappings[selectedIdx].SourceSubdomain
				m.form = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Key("subdomain").
							Title("Edit tunnel ID").
							Value(&m.newSubdomain).
							Placeholder("Enter new tunnel ID"),
					),
				).WithWidth(60)

				return m, m.form.Init()

			case "p":
				selectedIdx := m.list.Index()
				mapping := m.mappings[selectedIdx]

				if _, exist := here.mappings[mapping.SourceSubdomain]; mapping.IsConflict && exist {
					statusMessage := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f14242ff")).
						Render("\nthis tunnel has subdomain conflict.")

					return m, m.list.NewStatusMessage(statusMessage)
				}

				m.mappings[selectedIdx].TogglePause()
				m.mappings[selectedIdx].IsConflict = false

				items := createTunnelIndex(m.mappings)
				listCmd := m.list.SetItems(items)
				statusCmd := m.list.NewStatusMessage("Paused/Resumed " + m.mappings[selectedIdx].SourceSubdomain)

				return m, tea.Batch(listCmd, statusCmd)
			}
		}
	}

	if m.editMode && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			newSubdomain := m.form.GetString("subdomain")
			log.Print("Form completed with value:", newSubdomain)

			if newSubdomain != "" && newSubdomain != m.mappings[m.editingIndex].SourceSubdomain {
				m.mappings[m.editingIndex].RenameSubdomain(newSubdomain)
				items := createTunnelIndex(m.mappings)
				listCmd := m.list.SetItems(items)
				statusCmd := m.list.NewStatusMessage("Renamed tunnel to: " + newSubdomain)
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

func (m MainUIModel) View() string {
	if m.editMode {
		return docStyle.Render(m.list.View() + "\n\n" + m.form.View())
	}
	return docStyle.Render(m.list.View())
}
