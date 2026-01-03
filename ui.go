package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	items := []list.Item{}
	mappings, ok := s.Context().Value(MappingContextKey).([]MappingDisplayModel)
	if !ok {
		log.Error("failed to parse mappings")
		tea.Println("oh no.. something went wrong.")

		s.Close()
		return MainUIModel{}, nil
	}

	for idx, mapping := range mappings {
		items = append(items,
			MainUIListItem{
				title: fmt.Sprintf("#%d (-R%d)", idx, mapping.TargetPort),
				desc: fmt.Sprint(
					srv.getHostPrefix(),
					mapping.SourceSubdomain,
					srv.getHostSuffix()),
			},
		)
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)

	model := MainUIModel{list}
	model.list.Title = "Here: simple & nodeps tunnel"

	return model, []tea.ProgramOption{tea.WithAltScreen()}
}
