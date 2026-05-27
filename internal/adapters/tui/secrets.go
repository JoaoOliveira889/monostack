package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
)

func leadingWhitespacePrefix(line string) string {
	i := 0
	for i < len(line) {
		switch line[i] {
		case ' ', '\t':
			i++
		default:
			return line[:i]
		}
	}
	return line
}

func wrapSecretLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	if line == "" {
		return []string{""}
	}

	indent := leadingWhitespacePrefix(line)
	continuationPrefix := indent + "  "
	remaining := line
	var lines []string
	first := true

	for len(remaining) > 0 {
		prefix := ""
		limit := width
		if !first {
			prefix = continuationPrefix
			limit = width - lipgloss.Width(prefix)
			if limit < 1 {
				limit = 1
			}
		}

		runes := []rune(remaining)
		if lipgloss.Width(remaining) <= limit {
			lines = append(lines, prefix+remaining)
			break
		}

		cut := limit
		if cut > len(runes) {
			cut = len(runes)
		}
		segment := strings.TrimRight(string(runes[:cut]), " ")
		if segment == "" {
			segment = string(runes[:cut])
		}
		lines = append(lines, prefix+segment)
		remaining = strings.TrimLeft(string(runes[cut:]), " ")
		first = false
	}

	return lines
}

func wrapSecretValueDisplay(value string, width int) string {
	display := formatSecretValueDisplay(value)
	if strings.TrimSpace(display) == "" {
		return "(empty)"
	}

	var lines []string
	for _, rawLine := range strings.Split(display, "\n") {
		lines = append(lines, wrapSecretLine(rawLine, width)...)
	}
	return strings.Join(lines, "\n")
}

func orderedSecretVersions(versions []domain.SecretVersion) []domain.SecretVersion {
	if len(versions) == 0 {
		return nil
	}

	ordered := make([]domain.SecretVersion, 0, len(versions))
	for _, version := range versions {
		if containsSecretStage(version.Stages, "AWSCURRENT") {
			ordered = append(ordered, version)
		}
	}
	for _, version := range versions {
		if !containsSecretStage(version.Stages, "AWSCURRENT") {
			ordered = append(ordered, version)
		}
	}
	return ordered
}

func containsSecretStage(stages []string, target string) bool {
	for _, stage := range stages {
		if stage == target {
			return true
		}
	}
	return false
}

func (m Model) currentSecretCurrentVersionIndex() int {
	for i, version := range m.secretVersions {
		if containsSecretStage(version.Stages, "AWSCURRENT") {
			return i
		}
	}
	return -1
}

func (m Model) secretVersionIsCurrent(version domain.SecretVersion) bool {
	return containsSecretStage(version.Stages, "AWSCURRENT")
}

func secretVersionVisualLabel(index int, version domain.SecretVersion) string {
	label := fmt.Sprintf("v%d", index+1)
	if containsSecretStage(version.Stages, "AWSCURRENT") {
		label += " (current)"
	}
	return label
}

func (m Model) secretValueCopyText() string {
	text := strings.TrimSpace(m.secretValueDisplay)
	if text != "" && text != "(empty)" {
		return text
	}
	if value := strings.TrimSpace(m.secretValue.SecretString); value != "" {
		return value
	}
	if value := strings.TrimSpace(m.secretValue.SecretBinaryBase64); value != "" {
		return value
	}
	return ""
}

func (m *Model) refreshSecretValueViewport() {
	if m.secretValueViewport.Width == 0 || m.secretValueViewport.Height == 0 {
		return
	}

	content := m.secretValueDisplay
	if strings.TrimSpace(content) == "" || content == "(empty)" {
		content = m.secretValue.SecretString
		if strings.TrimSpace(content) == "" {
			content = m.secretValue.SecretBinaryBase64
		}
	}

	if strings.TrimSpace(content) == "" {
		content = "(empty)"
	}

	m.secretValueViewport.SetContent(wrapSecretValueDisplay(content, m.secretValueViewport.Width))
	m.secretValueViewport.GotoTop()
}

func (m *Model) configureSecretsLayout() {
	editorWidth := m.width - 22
	if editorWidth < 42 {
		editorWidth = 42
	}
	if editorWidth > 90 {
		editorWidth = 90
	}

	editorHeight := m.height / 4
	if editorHeight < 8 {
		editorHeight = 8
	}
	if editorHeight > 14 {
		editorHeight = 14
	}

	m.secretCreateNameInput.Width = editorWidth
	m.secretCreateValueInput.SetWidth(editorWidth)
	m.secretCreateValueInput.SetHeight(editorHeight)

	m.secretUpdateValueInput.SetWidth(editorWidth)
	m.secretUpdateValueInput.SetHeight(editorHeight)

	modalWidth := m.width - 18
	if modalWidth < 60 {
		modalWidth = 60
	}
	if modalWidth > 96 {
		modalWidth = 96
	}

	modalHeight := m.height - 14
	if modalHeight < 12 {
		modalHeight = 12
	}
	if modalHeight > 28 {
		modalHeight = 28
	}

	m.secretValueViewport = viewport.New(modalWidth-4, modalHeight-6)
	m.refreshSecretValueViewport()
}
