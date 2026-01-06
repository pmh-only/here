package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func createListStyles(renderer *lipgloss.Renderer) list.Styles {
	styles := list.Styles{}

	var bullet = "•"

	verySubduedColor := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	subduedColor := lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

	styles.TitleBar = renderer.NewStyle().Padding(0, 0, 1, 2) //nolint:mnd

	styles.Title = renderer.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	styles.Spinner = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#8E8E8E", Dark: "#747373"})

	styles.FilterPrompt = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#ECFD65"})

	styles.FilterCursor = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

	styles.DefaultFilterCharacterMatch = renderer.NewStyle().Underline(true)

	styles.StatusBar = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 1, 2) //nolint:mnd

	styles.StatusEmpty = renderer.NewStyle().Foreground(subduedColor)

	styles.StatusBarActiveFilter = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	styles.StatusBarFilterCount = renderer.NewStyle().Foreground(verySubduedColor)

	styles.NoItems = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"})

	styles.ArabicPagination = renderer.NewStyle().Foreground(subduedColor)

	styles.PaginationStyle = renderer.NewStyle().PaddingLeft(2) //nolint:mnd

	styles.HelpStyle = renderer.NewStyle().Padding(1, 0, 0, 2) //nolint:mnd

	styles.ActivePaginationDot = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"}).
		SetString(bullet)

	styles.InactivePaginationDot = renderer.NewStyle().
		Foreground(verySubduedColor).
		SetString(bullet)

	styles.DividerDot = renderer.NewStyle().
		Foreground(verySubduedColor).
		SetString(" " + bullet + " ")

	return styles
}

func createListItemStyles(renderer *lipgloss.Renderer) list.DefaultItemStyles {
	styles := list.DefaultItemStyles{}

	styles.NormalTitle = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)

	styles.NormalDesc = styles.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	styles.SelectedTitle = renderer.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)

	styles.SelectedDesc = styles.SelectedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	styles.DimmedTitle = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)

	styles.DimmedDesc = styles.DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})

	styles.FilterMatch = renderer.NewStyle().Underline(true)

	return styles
}

func createHelpStyles(renderer *lipgloss.Renderer) help.Styles {
	keyStyle := renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle := renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle := renderer.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	return help.Styles{
		ShortKey:       keyStyle,
		ShortDesc:      descStyle,
		ShortSeparator: sepStyle,
		Ellipsis:       sepStyle,
		FullKey:        keyStyle,
		FullDesc:       descStyle,
		FullSeparator:  sepStyle,
	}
}

func createFormTheme(renderer *lipgloss.Renderer) *huh.Theme {
	theme := huh.Theme{}

	var (
		indigo  = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
		fuchsia = lipgloss.Color("#F780E2")
		green   = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	)

	theme.Form.Base = renderer.NewStyle()
	theme.Group.Base = renderer.NewStyle()
	theme.FieldSeparator = renderer.NewStyle().SetString("\n\n")

	theme.Focused.Base = renderer.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.ThickBorder()).BorderLeft(true)
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = renderer.NewStyle()
	theme.Focused.TextInput.Cursor = renderer.NewStyle()
	theme.Focused.TextInput.Placeholder = renderer.NewStyle()
	theme.Focused.TextInput.Prompt = renderer.NewStyle()

	theme.Help = createHelpStyles(renderer)

	theme.Focused.Base = theme.Focused.Base.BorderForeground(lipgloss.Color("238"))
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = theme.Focused.Title.Foreground(indigo).Bold(true)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(green)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(fuchsia)

	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.Card = theme.Blurred.Base

	return &theme
}
