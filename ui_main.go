package main

import (
	"github.com/charmbracelet/bubbles/list"
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
	list list.Model
}

func (m MainUIModel) Init() tea.Cmd {
	return nil
}

func (m MainUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
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
	return docStyle.Render(m.list.View(), "\nhi")
}
