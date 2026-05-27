package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monostack/internal/pkg/ui"
)

func (m *Model) refreshLogViewport() {
	if m.logViewport.Width == 0 || m.logViewport.Height == 0 {
		return
	}
	m.logViewport.SetContent(m.renderCommandLog(m.logViewport.Width))
	if m.commandLogCursor < m.logViewport.YOffset {
		m.logViewport.YOffset = m.commandLogCursor
	} else if m.commandLogCursor >= m.logViewport.YOffset+m.logViewport.Height {
		m.logViewport.YOffset = m.commandLogCursor - m.logViewport.Height + 1
	}
}

func (m Model) renderCommandLog(width int) string {
	if len(m.commandLogs) == 0 {
		return m.styles.InfoText.Render("  No commands executed yet.")
	}

	contentWidth := width - 6
	if contentWidth < 0 {
		contentWidth = 0
	}

	var sb strings.Builder
	for i, entry := range m.commandLogs {
		timeStr := entry.Time.Format("15:04:05")
		status := m.styles.SuccessBadge.Render("SUCCESS")
		if entry.Error != nil {
			status = m.styles.ErrorBadge.Render("FAILED")
		}

		headLine := fmt.Sprintf("  [%s] %s > %s : %s", timeStr, entry.Action, entry.Target, status)
		if m.isIndexSelected(selectionCommandLogs, i) || i == m.commandLogCursor {
			headLine = m.styles.SelectedListItem.Render(headLine)
		}
		sb.WriteString(headLine)
		sb.WriteString("\n")

		if entry.Error != nil {
			for _, line := range strings.Split(lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorError)).Width(contentWidth).Render("Error: "+entry.Error.Error()), "\n") {
				if line != "" {
					if m.isIndexSelected(selectionCommandLogs, i) || i == m.commandLogCursor {
						line = m.styles.SelectedListItem.Render(line)
					}
					sb.WriteString("    " + line + "\n")
				}
			}
		}

		if entry.Output != "" {
			for _, line := range strings.Split(lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSubtle)).Width(contentWidth).Render(strings.TrimSpace(entry.Output)), "\n") {
				if line != "" {
					if m.isIndexSelected(selectionCommandLogs, i) || i == m.commandLogCursor {
						line = m.styles.SelectedListItem.Render(line)
					}
					sb.WriteString("    " + line + "\n")
				}
			}
		}

		if i < len(m.commandLogs)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) renderLogsModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Command Logs") + "\n\n")

	if len(m.commandLogs) == 0 {
		builder.WriteString(m.styles.InfoText.Render("No commands executed yet.") + "\n\n")
	} else {
		builder.WriteString(m.logViewport.View())
		builder.WriteString("\n")
	}

	builder.WriteString(m.styles.InfoText.Render("[o] Close  |  [space] Select  |  [y] Copy  |  [jk] Scroll"))
	return builder.String()
}
