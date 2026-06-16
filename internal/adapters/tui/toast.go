package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
)

type Toast struct {
	Message string
	Type    ToastType
	Expire  time.Time
	Height  int
}

const (
	maxToasts      = 5
	toastDuration  = 4 * time.Second
	toastFadeDelay = 3 * time.Second
)

func (m *Model) pushToast(msg string, t ToastType) {
	now := time.Now()
	toast := Toast{
		Message: msg,
		Type:    t,
		Expire:  now.Add(toastDuration),
		Height:  1,
	}
	m.toasts = append(m.toasts, toast)

	if len(m.toasts) > maxToasts {
		m.toasts = m.toasts[len(m.toasts)-maxToasts:]
	}
}

func (m *Model) pushSuccessToast(msg string) {
	m.pushToast(msg, ToastSuccess)
}

func (m *Model) pushErrorToast(msg string) {
	m.pushToast(msg, ToastError)
}

func (m *Model) pushInfoToast(msg string) {
	m.pushToast(msg, ToastInfo)
}

func (m *Model) pruneExpiredToasts(now time.Time) {
	var alive []Toast
	for _, t := range m.toasts {
		if now.Before(t.Expire) {
			alive = append(alive, t)
		}
	}
	m.toasts = alive
}

func (m *Model) toastTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return toastTickMsg{}
	})
}

func (m *Model) renderToasts(width int) string {
	if len(m.toasts) == 0 {
		return ""
	}

	now := time.Now()
	m.pruneExpiredToasts(now)
	if len(m.toasts) == 0 {
		return ""
	}

	var rendered []string
	for _, toast := range m.toasts {
		remaining := time.Until(toast.Expire)
		if remaining <= 0 {
			continue
		}

		opacity := 1.0
		if remaining < time.Second {
			opacity = float64(remaining) / float64(time.Second)
			if opacity < 0 {
				opacity = 0
			}
		}

		var style lipgloss.Style
		switch toast.Type {
		case ToastSuccess:
			style = m.styles.ToastSuccess
		case ToastError:
			style = m.styles.ToastError
		case ToastInfo:
			style = m.styles.ToastInfo
		}

		text := style.Render(" " + toast.Message + " ")
		if opacity < 1.0 {
			text = lipgloss.NewStyle().
				Faint(true).
				Render(text)
		}

		rendered = append(rendered, text)
	}

	if len(rendered) == 0 {
		return ""
	}

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Right).
		Padding(0, 2).
		Render(lipgloss.JoinVertical(lipgloss.Right, rendered...))
}
