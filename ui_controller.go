package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)
	if !ok {
		log.Error("failed to parse mappings")
		tea.Println("oh no.. something went wrong.")

		s.Close()
		return MainUIModel{}, nil
	}

	items := createTunnelIndex(mappings)
	list := list.New(items, list.NewDefaultDelegate(), 0, 0)

	passwordNeeded := here.isPasswordRequired()

	model := MainUIModel{
		list:           list,
		mappings:       mappings,
		passwordNeeded: passwordNeeded,
		authenticated:  false,
		session:        s,
	}

	model.list.SetFilteringEnabled(false)
	model.list.Title = "Here: simple & nodeps tunnel"
	model.list.AdditionalShortHelpKeys = func() []key.Binding {
		bindings := []key.Binding{
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "copy url"),
			),
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "pause/resume"),
			),
			key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit domain"),
			),
		}
		if passwordNeeded {
			bindings = append(bindings, key.NewBinding(
				key.WithKeys("l"),
				key.WithHelp("l", "login"),
			))
		}
		return bindings
	}

	model.list.AdditionalFullHelpKeys = func() []key.Binding {
		bindings := []key.Binding{
			key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "copy tunnel url"),
			),
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "pause/resume subdomain"),
			),
			key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "change subdomain"),
			),
		}
		if passwordNeeded {
			bindings = append(bindings, key.NewBinding(
				key.WithKeys("l"),
				key.WithHelp("l", "login with password"),
			))
		}
		return bindings
	}

	return model, []tea.ProgramOption{tea.WithAltScreen()}
}

func createTunnelIndex(mappings []MappingDisplayModel) []list.Item {
	items := []list.Item{}

	for idx, mapping := range mappings {
		title := fmt.Sprintf("Tunnel #%d (-R%d)", idx, mapping.TargetPort)
		desc := fmt.Sprint(here.getHostPrefix(), mapping.SourceSubdomain, here.getHostSuffix())

		if mapping.Actual.IsPaused {
			title += " [PAUSED]"
		}

		if mapping.IsConflict {
			desc += " [CONFLICT]"
		}

		items = append(items,
			MainUIListItem{
				title,
				desc,
			},
		)
	}

	return items
}
