package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CommandPaletteItem struct {
	Label       string
	Key         string
	Category    string
	description string
}

const (
	cpCategoryNav      = "Navigation"
	cpCategoryGlobal   = "Global"
	cpCategoryS3       = "S3 Explorer"
	cpCategorySQS      = "SQS Queues"
	cpCategorySNS      = "SNS Topics"
	cpCategorySecrets  = "Secrets"
	cpCategorySettings = "Settings"
)

func (m *Model) buildCommandPaletteItems() []CommandPaletteItem {
	items := []CommandPaletteItem{
		{Label: "S3 Explorer", Key: "1", Category: cpCategoryNav},
		{Label: "SQS Queues", Key: "2", Category: cpCategoryNav},
		{Label: "SNS Topics", Key: "3", Category: cpCategoryNav},
		{Label: "Secrets", Key: "4", Category: cpCategoryNav},
		{Label: "Settings", Key: "5", Category: cpCategoryNav},

		{Label: "Quit", Key: "q / ctrl+c", Category: cpCategoryGlobal},
		{Label: "Command Logs", Key: "o", Category: cpCategoryGlobal},
		{Label: "Help", Key: "?", Category: cpCategoryGlobal},
		{Label: "Filter List", Key: "/", Category: cpCategoryGlobal},
		{Label: "Sort", Key: "ctrl+s", Category: cpCategoryGlobal},
		{Label: "Copy ARN", Key: "ctrl+y", Category: cpCategoryGlobal},
		{Label: "Profile Switcher", Key: "p", Category: cpCategoryGlobal},
		{Label: "Toggle Theme", Key: "T", Category: cpCategoryGlobal},

		{Label: "Open Bucket / Inspect", Key: "enter", Category: cpCategoryS3},
		{Label: "Navigate Back", Key: "esc / h", Category: cpCategoryS3},
		{Label: "Create Bucket", Key: "c", Category: cpCategoryS3},
		{Label: "Create Folder", Key: "f", Category: cpCategoryS3},
		{Label: "Upload Object", Key: "u", Category: cpCategoryS3},
		{Label: "Delete Object/Bucket", Key: "d", Category: cpCategoryS3},
		{Label: "Download File", Key: "w", Category: cpCategoryS3},
		{Label: "Preview Object", Key: "v", Category: cpCategoryS3},
		{Label: "Object Versions", Key: "V", Category: cpCategoryS3},
		{Label: "Open in Browser", Key: "b", Category: cpCategoryS3},

		{Label: "Inspect Queue", Key: "enter", Category: cpCategorySQS},
		{Label: "View Routes", Key: "l", Category: cpCategorySQS},
		{Label: "Create Queue", Key: "c", Category: cpCategorySQS},
		{Label: "Send Message", Key: "s", Category: cpCategorySQS},
		{Label: "Delete Queue", Key: "d", Category: cpCategorySQS},
		{Label: "Purge Queue", Key: "m", Category: cpCategorySQS},
		{Label: "Purge All", Key: "M", Category: cpCategorySQS},
		{Label: "Subscribe SNS Topics", Key: "b", Category: cpCategorySQS},
		{Label: "Peek Messages", Key: "v", Category: cpCategorySQS},

		{Label: "Inspect Topic", Key: "enter", Category: cpCategorySNS},
		{Label: "View Routes", Key: "l", Category: cpCategorySNS},
		{Label: "Create Topic", Key: "c", Category: cpCategorySNS},
		{Label: "Publish Message", Key: "s", Category: cpCategorySNS},
		{Label: "Delete Topic/Sub", Key: "d", Category: cpCategorySNS},
		{Label: "Batch Subscribe", Key: "b", Category: cpCategorySNS},
		{Label: "Import YAML", Key: "i", Category: cpCategorySNS},
		{Label: "Edit Filter", Key: "e", Category: cpCategorySNS},

		{Label: "Inspect Secret", Key: "enter", Category: cpCategorySecrets},
		{Label: "View Versions", Key: "l", Category: cpCategorySecrets},
		{Label: "Create Secret", Key: "c", Category: cpCategorySecrets},
		{Label: "Update Secret", Key: "u", Category: cpCategorySecrets},
		{Label: "Reveal Value", Key: "v", Category: cpCategorySecrets},
		{Label: "Delete Secret", Key: "d", Category: cpCategorySecrets},
		{Label: "Recover Secret", Key: "R", Category: cpCategorySecrets},
		{Label: "Promote Version", Key: "r", Category: cpCategorySecrets},

		{Label: "Edit Field", Key: "enter", Category: cpCategorySettings},
		{Label: "Save Config", Key: "s", Category: cpCategorySettings},
		{Label: "Export Snapshot", Key: "E", Category: cpCategorySettings},
		{Label: "Load Snapshot", Key: "L", Category: cpCategorySettings},
	}

	return items
}

func (m *Model) openCommandPalette() tea.Cmd {
	m.showCommandPalette = true
	m.commandPaletteItems = m.buildCommandPaletteItems()
	m.commandPaletteFiltered = m.commandPaletteItems
	m.commandPaletteCursor = 0
	m.commandPaletteInput.SetValue("")
	m.commandPaletteInput.Focus()
	return textinput.Blink
}

func (m *Model) closeCommandPalette() {
	m.showCommandPalette = false
	m.commandPaletteInput.Blur()
	m.commandPaletteInput.SetValue("")
}

func (m *Model) filterCommandPalette(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		m.commandPaletteFiltered = m.commandPaletteItems
		m.commandPaletteCursor = 0
		return
	}

	var filtered []CommandPaletteItem
	for _, item := range m.commandPaletteItems {
		if strings.Contains(strings.ToLower(item.Label), query) ||
			strings.Contains(strings.ToLower(item.Category), query) ||
			strings.Contains(strings.ToLower(item.Key), query) {
			filtered = append(filtered, item)
		}
	}
	m.commandPaletteFiltered = filtered
	if m.commandPaletteCursor >= len(m.commandPaletteFiltered) {
		m.commandPaletteCursor = 0
	}
	if len(m.commandPaletteFiltered) > 0 && m.commandPaletteCursor < 0 {
		m.commandPaletteCursor = 0
	}
}

func (m *Model) executeCommandPaletteItem() tea.Cmd {
	if len(m.commandPaletteFiltered) == 0 || m.commandPaletteCursor >= len(m.commandPaletteFiltered) {
		m.closeCommandPalette()
		return nil
	}

	m.closeCommandPalette()
	return m.dispatchCommand(m.commandPaletteFiltered[m.commandPaletteCursor])
}

func (m *Model) dispatchCommand(item CommandPaletteItem) tea.Cmd {
	switch item.Label {
	case "S3 Explorer":
		return m.activatePanel(panelS3)
	case "SQS Queues":
		return m.activatePanel(panelSQS)
	case "SNS Topics":
		return m.activatePanel(panelSNS)
	case "Secrets":
		return m.activatePanel(panelSecrets)
	case "Settings":
		return m.activatePanel(panelConfig)
	case "Quit":
		return tea.Quit
	case "Command Logs":
		m.showLogsModal = true
		m.refreshLogViewport()
		return nil
	case "Help":
		if m.showHelpModal {
			m.showHelpModal = false
		} else {
			m.showHelpModal = true
		}
		return nil
	case "Profile Switcher":
		return m.listProfilesCmd()
	case "Toggle Theme":
		if m.currentTheme() == ThemeDark {
			m.ApplyTheme(ThemeLight)
			m.pushInfoToast("Light theme enabled")
		} else {
			m.ApplyTheme(ThemeDark)
			m.pushInfoToast("Dark theme enabled")
		}
		return nil
	default:
		return nil
	}
}

func (m *Model) renderCommandPalette() string {
	categories := make(map[string][]CommandPaletteItem)
	var catOrder []string

	for _, item := range m.commandPaletteFiltered {
		if _, exists := categories[item.Category]; !exists {
			catOrder = append(catOrder, item.Category)
		}
		categories[item.Category] = append(categories[item.Category], item)
	}

	var lines []string
	globalIndex := 0

	for _, cat := range catOrder {
		lines = append(lines, m.styles.CPCategory.Render(" " + cat))
		for _, item := range categories[cat] {
			prefix := "  "
			if globalIndex == m.commandPaletteCursor {
				prefix = m.styles.Cursor.Render("> ")
			}
			keyStr := m.styles.CPKey.Render(item.Key)
			label := item.Label

			if globalIndex == m.commandPaletteCursor {
				lines = append(lines, m.styles.CPSelected.Render(prefix+label)+"  "+keyStr)
			} else {
				lines = append(lines, m.styles.CPItem.Render(prefix+label)+"  "+m.styles.InfoText.Render(keyStr))
			}
			globalIndex++
		}
	}

	listContent := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if listContent == "" {
		listContent = m.styles.InfoText.Render("  No matching commands")
	}

	input := m.commandPaletteInput.View()

	modalContent := lipgloss.JoinVertical(lipgloss.Left,
		input,
		"",
		listContent,
	)

	maxW := m.width - 12
	if maxW < 40 {
		maxW = 40
	}
	maxH := m.height - 8
	if maxH < 10 {
		maxH = 10
	}

	return m.styles.Modal.
		Width(maxW).
		MaxHeight(maxH).
		Render(modalContent)
}
