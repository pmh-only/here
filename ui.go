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

	model := MainUIModel{
		list:     list,
		mappings: mappings,
	}

	model.list.SetFilteringEnabled(false)
	model.list.Title = "Here: simple & nodeps tunnel"
	model.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "pause"),
			),
			key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit ID"),
			),
		}
	}

	model.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "pause subdomain"),
			),
			key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit tunnel ID"),
			),
		}
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
