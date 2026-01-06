package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	renderer := bubbletea.MakeRenderer(s)
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)
	if !ok || mappings == nil {
		log.Error("failed to parse mappings")

		tea.Println("Oops. hereserver failed to detect tunnels.")
		tea.Println("Please read the docs and retry with -R flags")

		s.Close()
		return EmptyUIModel{}, nil
	}

	items := createTunnelIndex(mappings)

	delegate := list.NewDefaultDelegate()
	delegate.Styles = createListItemStyles(renderer)

	list := list.New(items, delegate, 0, 0)
	list.Styles = createListStyles(renderer)
	list.Help.Styles = createHelpStyles(renderer)

	passwordNeeded := here.isPasswordRequired()

	model := MainUIModel{
		list:           list,
		mappings:       mappings,
		passwordNeeded: passwordNeeded,
		authenticated:  false,
		session:        s,
		renderer:       renderer,
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
