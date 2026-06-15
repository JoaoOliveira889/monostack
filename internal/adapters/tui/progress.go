package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/pkg/ui"
)

type progressBar struct {
	label   string
	percent float64
	width   int
}

type progressTracker struct {
	mu        sync.Mutex
	operation string
	total     int64
	current   int64
	done      bool
	result    tea.Msg
	destPath  string
}

func newProgressBar(label string, width int) progressBar {
	if width < 10 {
		width = 10
	}
	return progressBar{label: label, width: width}
}

func (p progressBar) view() string {
	barWidth := p.width - lipgloss.Width(p.label) - 10
	if barWidth < 6 {
		barWidth = 6
	}

	filled := int(float64(barWidth) * p.percent / 100.0)
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	filledBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSuccess)).
		Render(strings.Repeat("█", filled))
	emptyBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBg)).
		Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s [%s%s] %3.0f%%", p.label, filledBar, emptyBar, p.percent)
}

func progressFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (m *Model) progressTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		if m.progressTracker == nil {
			return progressMsg{Percent: 100}
		}

		tracker := m.progressTracker
		tracker.mu.Lock()
		defer tracker.mu.Unlock()

		if tracker.done {
			return progressDoneMsg{
				Operation: tracker.operation,
				Result:    tracker.result,
			}
		}

		if tracker.destPath != "" && tracker.total > 0 {
			tracker.current = progressFileSize(tracker.destPath)
		}

		percent := float64(0)
		if tracker.total > 0 {
			percent = float64(tracker.current) / float64(tracker.total) * 100
			if percent > 99 {
				percent = 99
			}
		}

		return progressMsg{Operation: tracker.operation, Percent: percent}
	})
}
