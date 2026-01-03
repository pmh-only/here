package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	textInput    textinput.Model
	editMode     bool
	editingIndex int
}

func (m MainUIModel) Init() tea.Cmd {
	return nil
}

func (m MainUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle edit mode
		if m.editMode {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.editMode = false
				m.textInput.Blur()
				return m, nil

			case "enter":
				newID := m.textInput.Value()
				if newID != "" {
					m.mappings[m.editingIndex].RenameSubdomain(newID)
					items := createTunnelIndex(m.mappings)
					listCmd := m.list.SetItems(items)
					m.editMode = false
					m.textInput.Blur()
					statusCmd := m.list.NewStatusMessage("Renamed tunnel to: " + newID)
					return m, tea.Batch(listCmd, statusCmd)
				}

				m.editMode = false
				m.textInput.Blur()
				return m, nil
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Normal mode
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "e":
			selectedIdx := m.list.Index()
			m.editingIndex = selectedIdx
			m.editMode = true

			// Initialize text input with current subdomain
			m.textInput = textinput.New()
			m.textInput.Placeholder = "Enter new tunnel ID"
			m.textInput.SetValue(m.mappings[selectedIdx].SourceSubdomain)
			m.textInput.Focus()
			m.textInput.Width = 40

			return m, nil

		case "p":
			selectedIdx := m.list.Index()

			if m.mappings[selectedIdx].IsConflict {
				statusMessage := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#f14242ff")).
					Render("\nthis tunnel has subdomain conflict.")

				return m, m.list.NewStatusMessage(statusMessage)
			}

			m.mappings[selectedIdx].TogglePause()

			items := createTunnelIndex(m.mappings)
			listCmd := m.list.SetItems(items)
			statusCmd := m.list.NewStatusMessage("Paused " + m.mappings[selectedIdx].SourceSubdomain)

			return m, tea.Batch(listCmd, statusCmd)
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainUIModel) View() string {
	if m.editMode {
		return docStyle.Render(m.list.View() + "\n\n" + "Edit tunnel ID:\n" + m.textInput.View() + "\n(press enter to save, esc to cancel)")
	}
	return docStyle.Render(m.list.View())
}
