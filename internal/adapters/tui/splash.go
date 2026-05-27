package tui

import (
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/pkg/ui"
)

func (m *Model) renderSplash() string {
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorStack)).
		Bold(true).
		Render(m.spinnerView() + " Opening MonoStack...")
	subtitle := m.styles.InfoText.Render("A dense AWS workspace, rendered lightly.")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderBrandWordmark(false),
		"",
		mutedSplashText("  Terminal-first AWS control for dense daily workflows."),
		"",
		status,
		mutedSplashText(" "+ui.SpinnerFrames[m.splashFrame%len(ui.SpinnerFrames)]+" starting up"),
		"",
		subtitle,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func mutedSplashText(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSubtle)).
		Render(text)
}

func renderBrandWordmark(compact bool) string {
	mono := ui.BrandMonoStyle
	stack := ui.BrandStackStyle
	subtle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSubtle))

	if compact {
		return lipgloss.JoinHorizontal(lipgloss.Bottom,
			mono.Render("Mono"),
			stack.Render("Stack"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top,
			mono.Render("Mono"),
			stack.Render("Stack"),
		),
		subtle.Render("terminal aws control deck"),
	)
}
