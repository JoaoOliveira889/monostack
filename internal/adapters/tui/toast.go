package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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
	maxToasts      = 3
	toastDuration  = 4 * time.Second
	toastFadeDelay = 3 * time.Second
)

type lineCell struct {
	r     rune
	style string
}

func parseANSILine(s string) []lineCell {
	var cells []lineCell
	currentStyle := ""
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n; {
		if runes[i] == '\x1b' {
			start := i
			i++ // skip '\x1b'
			if i < n && runes[i] == '[' {
				i++ // skip '['
				for i < n && !((runes[i] >= 'a' && runes[i] <= 'z') || (runes[i] >= 'A' && runes[i] <= 'Z')) {
					i++
				}
				if i < n {
					i++ // include the terminating letter
				}
			} else {
				for i < n && runes[i] != 'm' && runes[i] != '\a' {
					i++
				}
				if i < n {
					i++
				}
			}
			seq := string(runes[start:i])
			if seq == "\x1b[0m" {
				currentStyle = ""
			} else {
				currentStyle += seq
			}
			continue
		}

		r := runes[i]
		w := runewidth.RuneWidth(r)
		if w <= 0 {
			i++
			continue
		}

		cells = append(cells, lineCell{r: r, style: currentStyle})
		for j := 1; j < w; j++ {
			cells = append(cells, lineCell{r: 0, style: currentStyle})
		}
		i++
	}
	return cells
}

func rebuildANSILine(cells []lineCell) string {
	var sb strings.Builder
	lastStyle := ""
	for _, cell := range cells {
		if cell.style != lastStyle {
			if lastStyle != "" {
				sb.WriteString("\x1b[0m")
			}
			sb.WriteString(cell.style)
			lastStyle = cell.style
		}
		if cell.r != 0 {
			sb.WriteRune(cell.r)
		}
	}
	if lastStyle != "" {
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

func overlayString(bg, fg string, startRow, startCol int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	for i, fgLine := range fgLines {
		row := startRow + i
		if row < 0 || row >= len(bgLines) {
			continue
		}

		bgCells := parseANSILine(bgLines[row])
		fgCells := parseANSILine(fgLine)

		neededLen := startCol + len(fgCells)
		if len(bgCells) < neededLen {
			for len(bgCells) < neededLen {
				bgCells = append(bgCells, lineCell{r: ' ', style: ""})
			}
		}

		for j, fgCell := range fgCells {
			targetIdx := startCol + j
			if targetIdx >= 0 && targetIdx < len(bgCells) {
				bgCells[targetIdx] = fgCell
			}
		}

		bgLines[row] = rebuildANSILine(bgCells)
	}

	return strings.Join(bgLines, "\n")
}


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

func (m *Model) renderToasts() string {
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
		var prefix string
		switch toast.Type {
		case ToastSuccess:
			style = m.styles.ToastSuccess
			prefix = "✓ "
		case ToastError:
			style = m.styles.ToastError
			prefix = "✗ "
		case ToastInfo:
			style = m.styles.ToastInfo
			prefix = "ℹ "
		}

		text := style.Render(" " + prefix + toast.Message + " ")
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

	return lipgloss.JoinVertical(lipgloss.Right, rendered...)
}
