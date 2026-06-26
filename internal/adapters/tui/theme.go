package tui

import (
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/pkg/ui"
)

type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

type ThemeColors struct {
	Bg        string
	Fg        string
	Highlight string
	Selected  string
	Accent    string
	Mono      string
	Stack     string
	Success   string
	Error     string
	Warning   string
	Subtle    string
	Border    string
	Cyan      string
	Orange    string
	Indigo    string
	Rose      string
	Emerald   string
	Amber     string
}

func DarkColors() ThemeColors {
	return ThemeColors{
		Bg:        ui.ColorBg,
		Fg:        ui.ColorFg,
		Highlight: ui.ColorHighlight,
		Selected:  ui.ColorSelected,
		Accent:    ui.ColorAccent,
		Mono:      ui.ColorMono,
		Stack:     ui.ColorStack,
		Success:   ui.ColorSuccess,
		Error:     ui.ColorError,
		Warning:   ui.ColorWarning,
		Subtle:    ui.ColorSubtle,
		Border:    ui.ColorBorder,
		Cyan:      ui.ColorCyan,
		Orange:    ui.ColorOrange,
		Indigo:    ui.ColorIndigo,
		Rose:      ui.ColorRose,
		Emerald:   ui.ColorEmerald,
		Amber:     ui.ColorAmber,
	}
}

func LightColors() ThemeColors {
	return ThemeColors{
		Bg:        "#f8f9fc",
		Fg:        "#383a42",
		Highlight: "#4078f2",
		Selected:  "#d4dbfa",
		Accent:    "#a626a4",
		Mono:      "#1a1b26",
		Stack:     "#c18401",
		Success:   "#50a14f",
		Error:     "#e45649",
		Warning:   "#c18401",
		Subtle:    "#a0a1a7",
		Border:    "#d0d3db",
		Cyan:      "#0184bc",
		Orange:    "#d75f00",
		Indigo:    "#6366f1",
		Rose:      "#fb7185",
		Emerald:   "#34d399",
		Amber:     "#fbbf24",
	}
}

func NewTheme(t Theme) ThemeColors {
	switch t {
	case ThemeLight:
		return LightColors()
	default:
		return DarkColors()
	}
}

func (m *Model) buildStyles(t ThemeColors) styles {
	return styles{
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Mono)).
			Bold(true).
			Padding(0, 1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Subtle)).
			PaddingLeft(1),

		Footer: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Bg)).
			Foreground(lipgloss.Color(t.Subtle)).
			Padding(0, 1),

		ActiveTab: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Stack)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true).
			Padding(0, 1).
			MarginRight(1),

		InactiveTab: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Bg)).
			Foreground(lipgloss.Color(t.Subtle)).
			Padding(0, 1).
			MarginRight(1),

		BorderPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Mono)).
			Padding(0, 0),

		FocusedPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Stack)).
			Padding(0, 0),

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Border)).
			Background(lipgloss.Color(t.Bg)).
			Padding(1, 2),

		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Highlight)).
			Bold(true),

		ToastSuccess: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Success)).
			Background(lipgloss.Color(t.Bg)).
			Foreground(lipgloss.Color(t.Success)).
			Bold(true).
			Padding(0, 1),

		ToastError: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Error)).
			Background(lipgloss.Color(t.Bg)).
			Foreground(lipgloss.Color(t.Error)).
			Bold(true).
			Padding(0, 1),

		ToastInfo: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Cyan)).
			Background(lipgloss.Color(t.Bg)).
			Foreground(lipgloss.Color(t.Cyan)).
			Bold(true).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Stack)).
			Bold(true).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Highlight)),

		ListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Fg)),

		SelectedListItem: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Highlight)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true),

		MultiSelectedMarker: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Success)).
			Bold(true),

		CPCategory: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Accent)).
			Bold(true),

		CPItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Fg)),

		CPSelected: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Highlight)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true),

		CPKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Stack)),

		InputLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Accent)).
			Bold(true),

		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(t.Stack)).
			Padding(0, 1).
			Width(44).
			MarginLeft(2),

		InputUnfocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(t.Border)).
			Padding(0, 1).
			Width(44).
			MarginLeft(2),

		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Stack)).
			Background(lipgloss.Color(t.Bg)).
			Padding(1, 2),

		SuccessBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Success)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true).
			Padding(0, 1),

		WarningBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Warning)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true).
			Padding(0, 1),

		ErrorBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Error)).
			Foreground(lipgloss.Color(t.Bg)).
			Bold(true).
			Padding(0, 1),

		InfoText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Subtle)),

		BrandMono: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Mono)).
			Bold(true),

		BrandStack: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Stack)).
			Bold(true),

		FooterKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Accent)).
			Bold(true),

		FooterAction: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Fg)),
	}
}

func (m *Model) ApplyTheme(t Theme) {
	if m.config != nil {
		m.config.Theme = string(t)
	}
	m.styles = m.buildStyles(NewTheme(t))
	m.commandPaletteInput.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.themeColors().Accent)).
		Bold(true)
}

func (m Model) themeColors() ThemeColors {
	if m.config != nil && m.config.Theme == string(ThemeLight) {
		return LightColors()
	}
	return DarkColors()
}

func (m Model) currentTheme() Theme {
	if m.config != nil && m.config.Theme == string(ThemeLight) {
		return ThemeLight
	}
	return ThemeDark
}
