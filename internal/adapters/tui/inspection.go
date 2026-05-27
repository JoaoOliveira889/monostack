package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
	"monostack/internal/pkg/ui"
)

func (m *Model) refreshInspectionViewport() {
	if m.inspectionViewport.Width == 0 || m.inspectionViewport.Height == 0 {
		return
	}
	m.inspectionViewport.SetContent(m.inspectionContent)
}

func (m *Model) configureInspectionViewport() {
	width := m.width - 16
	if width < 52 {
		width = 52
	}
	if width > 110 {
		width = 110
	}

	height := m.height - 12
	if height < 12 {
		height = 12
	}
	if height > 28 {
		height = 28
	}

	m.inspectionViewport = viewport.New(width, height)
	m.refreshInspectionViewport()
}

func (m Model) openInspectionCmd() tea.Cmd {
	switch m.activeTab {
	case panelSQS:
		if len(m.queues) == 0 || m.selectedQueueIndex >= len(m.queues) {
			return m.setStatusMessage("No queue selected")
		}
		queue := m.queues[m.selectedQueueIndex]
		return m.inspectQueueCmd(queue, append([]domain.SNSSubscription(nil), m.queueSubscriptions...))
	case panelSNS:
		if len(m.topics) == 0 || m.selectedTopicIndex >= len(m.topics) {
			return m.setStatusMessage("No topic selected")
		}
		topic := m.topics[m.selectedTopicIndex]
		subs := append([]domain.SNSSubscription(nil), m.subscriptions...)
		managed := append([]domain.ManagedSubscription(nil), m.managedSubs...)
		if m.snsSubFocus == focusSubs && m.selectedSubIndex < len(m.subscriptions) {
			return m.inspectSubscriptionCmd(topic, m.subscriptions[m.selectedSubIndex], managed, subs)
		}
		return m.inspectTopicCmd(topic, managed, subs)
	default:
		return nil
	}
}

func (m Model) inspectQueueCmd(queue domain.SQSQueue, queueSubs []domain.SNSSubscription) tea.Cmd {
	logs := append([]CommandLogEntry(nil), m.commandLogs...)
	cfg := m.config
	awsUC := m.awsUseCase
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		messages, err := awsUC.ReceiveSQSMessages(ctx, cfg, queue.URL, 5)
		if err != nil {
			return errMsg{Error: fmt.Errorf("queue inspection for %s: %w", queue.Name, err)}
		}

		content := renderQueueInspection(queue, queueSubs, messages, filterLogsForQueue(logs, queue))
		return inspectionLoadedMsg{
			Title:    fmt.Sprintf("Queue Inspection — %s", queue.Name),
			Subtitle: "Recent messages, routing, metadata, and related errors",
			Content:  content,
		}
	}
}

func (m Model) inspectTopicCmd(topic domain.SNSTopic, managed []domain.ManagedSubscription, subs []domain.SNSSubscription) tea.Cmd {
	logs := append([]CommandLogEntry(nil), m.commandLogs...)
	return func() tea.Msg {
		content := renderTopicInspection(topic, managed, subs, filterLogsForTopic(logs, topic))
		return inspectionLoadedMsg{
			Title:    fmt.Sprintf("Topic Inspection — %s", topic.Name),
			Subtitle: "Routing summary, filters, subscriptions, and related errors",
			Content:  content,
		}
	}
}

func (m Model) inspectSubscriptionCmd(topic domain.SNSTopic, sub domain.SNSSubscription, managed []domain.ManagedSubscription, subs []domain.SNSSubscription) tea.Cmd {
	logs := append([]CommandLogEntry(nil), m.commandLogs...)
	return func() tea.Msg {
		content := renderSubscriptionInspection(topic, sub, managed, subs, filterLogsForSubscription(logs, topic, sub))
		return inspectionLoadedMsg{
			Title:    fmt.Sprintf("Route Inspection — %s", shortResourceName(sub.Endpoint)),
			Subtitle: "Selected route details, filters, direction, and related errors",
			Content:  content,
		}
	}
}

func renderQueueInspection(queue domain.SQSQueue, queueSubs []domain.SNSSubscription, messages []domain.SQSMessage, logs []CommandLogEntry) string {
	var b strings.Builder

	b.WriteString(sectionHeader("Queue Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Name:      %s\n", queue.Name))
	if strings.EqualFold(strings.TrimSpace(queue.Name), "error") {
		b.WriteString("Note:      This is a real queue named \"error\".\n")
	}
	b.WriteString(fmt.Sprintf("Ready:     %d\n", queue.MessagesAvailable))
	b.WriteString(fmt.Sprintf("In-flight: %d\n", queue.MessagesNotVisible))
	b.WriteString(fmt.Sprintf("Delayed:   %d\n", queue.MessagesDelayed))
	b.WriteString(fmt.Sprintf("URL:       %s\n", queue.URL))
	b.WriteString(fmt.Sprintf("ARN:       %s\n\n", queue.ARN))

	b.WriteString(sectionHeader("Incoming SNS Routes"))
	b.WriteString("\n")
	if len(queueSubs) == 0 {
		b.WriteString("No SNS topics are currently linked to this queue.\n\n")
	} else {
		for _, sub := range queueSubs {
			filterLabel := "all events"
			if len(sub.FilterPolicy) > 0 {
				filterLabel = fmt.Sprintf("%s via %s", formatFilterPolicy(sub.FilterPolicy), formatFilterScope(sub.FilterScope))
			}
			b.WriteString(fmt.Sprintf("- %s (%s)\n", shortResourceName(sub.TopicARN), filterLabel))
		}
		b.WriteString("\n")
	}

	b.WriteString(sectionHeader("Recent Messages"))
	b.WriteString("\n")
	if len(messages) == 0 {
		b.WriteString("No messages available in the latest peek.\n\n")
	} else {
		for i, msg := range messages {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, msg.ID))
			body := strings.TrimSpace(msg.Body)
			if body == "" {
				body = "(empty body)"
			}
			for _, line := range wrapLines(body, 76) {
				b.WriteString("   " + line + "\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(sectionHeader("Related Command Logs"))
	b.WriteString("\n")
	b.WriteString(renderInspectionLogs(logs))

	return strings.TrimRight(b.String(), "\n")
}

func renderTopicInspection(topic domain.SNSTopic, managed []domain.ManagedSubscription, subs []domain.SNSSubscription, logs []CommandLogEntry) string {
	var b strings.Builder

	outgoing := 0
	incoming := 0
	for i := range subs {
		if subs[i].TopicARN == topic.ARN {
			outgoing++
		} else {
			incoming++
		}
	}

	b.WriteString(sectionHeader("Topic Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Name:      %s\n", topic.Name))
	b.WriteString(fmt.Sprintf("Outgoing:  %d\n", outgoing))
	b.WriteString(fmt.Sprintf("Incoming:  %d\n", incoming))
	b.WriteString(fmt.Sprintf("ARN:       %s\n\n", topic.ARN))

	b.WriteString(sectionHeader("Routing Overview"))
	b.WriteString("\n")
	if len(subs) == 0 {
		b.WriteString("No subscriptions loaded for this topic.\n\n")
	} else {
		for _, sub := range subs {
			direction := "outgoing"
			if sub.TopicARN != topic.ARN {
				direction = "incoming"
			}
			b.WriteString(fmt.Sprintf("- %s [%s -> %s]\n", resolveManagedName(sub, managed), direction, strings.ToUpper(sub.Protocol)))
			b.WriteString(fmt.Sprintf("  Endpoint: %s\n", sub.Endpoint))
			if len(sub.FilterPolicy) == 0 {
				b.WriteString("  Filter:   all events\n")
			} else {
				b.WriteString(fmt.Sprintf("  Filter:   %s (%s)\n", formatFilterPolicy(sub.FilterPolicy), formatFilterScope(sub.FilterScope)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(sectionHeader("Related Command Logs"))
	b.WriteString("\n")
	b.WriteString(renderInspectionLogs(logs))

	return strings.TrimRight(b.String(), "\n")
}

func renderSubscriptionInspection(topic domain.SNSTopic, sub domain.SNSSubscription, managed []domain.ManagedSubscription, subs []domain.SNSSubscription, logs []CommandLogEntry) string {
	var b strings.Builder

	direction := "outgoing"
	if sub.TopicARN != topic.ARN {
		direction = "incoming"
	}

	b.WriteString(sectionHeader("Route Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Name:      %s\n", resolveManagedName(sub, managed)))
	b.WriteString(fmt.Sprintf("Direction: %s\n", direction))
	b.WriteString(fmt.Sprintf("Protocol:  %s\n", strings.ToUpper(sub.Protocol)))
	b.WriteString(fmt.Sprintf("Source:    %s\n", sub.TopicARN))
	b.WriteString(fmt.Sprintf("Endpoint:  %s\n", sub.Endpoint))
	if len(sub.FilterPolicy) == 0 {
		b.WriteString("Filter:    all events\n")
	} else {
		b.WriteString(fmt.Sprintf("Filter:    %s\n", formatFilterPolicy(sub.FilterPolicy)))
		b.WriteString(fmt.Sprintf("Scope:     %s\n", formatFilterScope(sub.FilterScope)))
	}
	b.WriteString("\n")

	if direction == "incoming" {
		b.WriteString(sectionHeader("Current Target Topic"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Selected topic %s is receiving from %s.\n\n", topic.Name, shortResourceName(sub.TopicARN)))
	} else {
		b.WriteString(sectionHeader("Current Source Topic"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Selected topic %s is publishing to %s.\n\n", topic.Name, shortResourceName(sub.Endpoint)))
	}

	b.WriteString(sectionHeader("Sibling Routes"))
	b.WriteString("\n")
	siblings := 0
	for _, candidate := range subs {
		if candidate.ARN == sub.ARN {
			continue
		}
		if candidate.TopicARN == sub.TopicARN {
			siblings++
			b.WriteString(fmt.Sprintf("- %s [%s]\n", shortResourceName(candidate.Endpoint), strings.ToUpper(candidate.Protocol)))
		}
	}
	if siblings == 0 {
		b.WriteString("No additional routes share this source topic.\n")
	}
	b.WriteString("\n")

	b.WriteString(sectionHeader("Related Command Logs"))
	b.WriteString("\n")
	b.WriteString(renderInspectionLogs(logs))

	return strings.TrimRight(b.String(), "\n")
}

func renderInspectionLogs(logs []CommandLogEntry) string {
	if len(logs) == 0 {
		return "No related command logs recorded yet."
	}

	var b strings.Builder
	for _, entry := range logs {
		stamp := entry.Time.Format("15:04:05")
		target := entry.Target
		if strings.TrimSpace(target) == "" {
			target = "app"
		}
		category := "resource"
		if strings.EqualFold(entry.Action, "error") && strings.TrimSpace(entry.Target) == "" {
			category = "app"
		}
		b.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", stamp, entry.Action, category))
		b.WriteString(fmt.Sprintf("  Target: %s\n", target))
		if entry.Error != nil {
			b.WriteString(fmt.Sprintf("  Error:  %s\n", entry.Error.Error()))
		} else if strings.TrimSpace(entry.Output) != "" {
			b.WriteString(fmt.Sprintf("  Note:   %s\n", entry.Output))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func filterLogsForQueue(logs []CommandLogEntry, queue domain.SQSQueue) []CommandLogEntry {
	var filtered []CommandLogEntry
	for _, entry := range logs {
		if containsAny(entry.Target, queue.Name, queue.URL, queue.ARN) {
			filtered = append(filtered, entry)
			continue
		}
		if strings.EqualFold(entry.Action, "error") && entry.Target == queue.Name {
			filtered = append(filtered, entry)
		}
	}
	return lastLogs(filtered, 6)
}

func filterLogsForTopic(logs []CommandLogEntry, topic domain.SNSTopic) []CommandLogEntry {
	var filtered []CommandLogEntry
	for _, entry := range logs {
		if containsAny(entry.Target, topic.Name, topic.ARN) {
			filtered = append(filtered, entry)
		}
	}
	return lastLogs(filtered, 6)
}

func filterLogsForSubscription(logs []CommandLogEntry, topic domain.SNSTopic, sub domain.SNSSubscription) []CommandLogEntry {
	var filtered []CommandLogEntry
	for _, entry := range logs {
		if containsAny(entry.Target, topic.Name, topic.ARN, sub.Endpoint, sub.ARN, shortResourceName(sub.Endpoint)) {
			filtered = append(filtered, entry)
		}
	}
	return lastLogs(filtered, 6)
}

func lastLogs(logs []CommandLogEntry, limit int) []CommandLogEntry {
	if len(logs) <= limit {
		return logs
	}
	return append([]CommandLogEntry(nil), logs[len(logs)-limit:]...)
}

func containsAny(value string, candidates ...string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" && strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func formatFilterPolicy(filterPolicy map[string][]string) string {
	if len(filterPolicy) == 0 {
		return "all events"
	}
	keys := make([]string, 0, len(filterPolicy))
	for key := range filterPolicy {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, strings.Join(filterPolicy[key], ",")))
	}
	return strings.Join(parts, " | ")
}

func formatFilterScope(scope string) string {
	if scope == domain.SubscriptionFilterScopeMessageBody {
		return "message body"
	}
	return "message attributes"
}

func resolveManagedName(sub domain.SNSSubscription, managed []domain.ManagedSubscription) string {
	for _, item := range managed {
		if item.SubscriptionARN == sub.ARN {
			return item.Name
		}
		if item.DestinationARN == sub.Endpoint && item.TopicARN == sub.TopicARN {
			return item.Name
		}
	}
	return shortResourceName(sub.Endpoint)
}

func wrapLines(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	for _, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			lines = append(lines, "")
			continue
		}
		for lipgloss.Width(line) > width {
			cut := width
			if cut > len(line) {
				cut = len(line)
			}
			lines = append(lines, strings.TrimSpace(line[:cut]))
			line = strings.TrimSpace(line[cut:])
		}
		lines = append(lines, line)
	}
	return lines
}

func sectionHeader(title string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorAccent)).
		Bold(true).
		Render(title)
}

func (m Model) renderInspectionModal() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render(m.inspectionTitle) + "\n")
	if strings.TrimSpace(m.inspectionSubtitle) != "" {
		b.WriteString(m.styles.InfoText.Render(m.inspectionSubtitle) + "\n\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString(m.inspectionViewport.View())
	b.WriteString("\n\n")
	b.WriteString(m.styles.InfoText.Render("[esc] Close  |  [jk] Scroll  |  [o] Global logs"))

	return b.String()
}
