package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
	"monostack/internal/pkg/ui"
)

func visibleRange(total, cursor, maxVisible int) (int, int) {
	if maxVisible < 1 {
		maxVisible = 1
	}
	start := 0
	if cursor >= maxVisible {
		start = cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func humanBytes(size int64) string {
	switch {
	case size >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	case size >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func shortResourceName(value string) string {
	if idx := strings.LastIndex(value, ":"); idx >= 0 && idx < len(value)-1 {
		return value[idx+1:]
	}
	if idx := strings.LastIndex(value, "/"); idx >= 0 && idx < len(value)-1 {
		return value[idx+1:]
	}
	return value
}

func isLikelyImageKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp", ".tiff"} {
		if strings.HasSuffix(key, ext) {
			return true
		}
	}
	return false
}

func truncateToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}

	ellipsis := "…"
	if maxWidth <= lipgloss.Width(ellipsis) {
		return strings.Repeat(".", maxWidth)
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes)
		if lipgloss.Width(candidate)+lipgloss.Width(ellipsis) <= maxWidth {
			return candidate + ellipsis
		}
		runes = runes[:len(runes)-1]
	}

	return ellipsis
}

func (m Model) renderMetricPill(label string, value string, accent lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBg)).
		Background(accent).
		Padding(0, 1).
		Bold(true).
		Render(label + " " + value)
}

func (m Model) renderMutedPill(label string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorFg)).
		Background(lipgloss.Color(ui.ColorBorder)).
		Padding(0, 1).
		Render(label)
}

func (m Model) renderSectionCaption(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorAccent)).
		Bold(true).
		Render("  " + text)
}

func (m Model) renderDetailLine(label, value string, width int) string {
	text := fmt.Sprintf("  %-10s %s", label+":", value)
	if width > 0 {
		text = truncateToWidth(text, width)
	}
	return m.styles.InfoText.Render(text)
}

func (m Model) renderPillRows(maxWidth int, pills ...string) string {
	if maxWidth < 1 {
		maxWidth = 1
	}
	var rows []string
	var current []string
	currentWidth := 0
	flush := func() {
		if len(current) == 0 {
			return
		}
		row := current[0]
		for i := 1; i < len(current); i++ {
			row += " " + current[i]
		}
		rows = append(rows, row)
		current = nil
		currentWidth = 0
	}
	for _, pill := range pills {
		pillWidth := lipgloss.Width(pill)
		if len(current) == 0 {
			current = append(current, pill)
			currentWidth = pillWidth
			continue
		}
		if currentWidth+1+pillWidth <= maxWidth {
			current = append(current, pill)
			currentWidth += 1 + pillWidth
			continue
		}
		flush()
		current = append(current, pill)
		currentWidth = pillWidth
	}
	flush()
	if len(rows) <= 1 {
		return strings.Join(rows, "\n  ")
	}
	return strings.Join(rows, "\n\n  ")
}

func renderCounterSummary(queue domain.SQSQueue) string {
	return fmt.Sprintf("r:%d f:%d d:%d  %s", queue.MessagesAvailable, queue.MessagesNotVisible, queue.MessagesDelayed, queue.Name)
}

func renderRouteSummary(sub domain.SNSSubscription, managed []domain.ManagedSubscription, direction string) string {
	name := shortResourceName(sub.Endpoint)
	for _, item := range managed {
		if item.SubscriptionARN == sub.ARN {
			name = item.Name
			break
		}
	}
	filterLabel := "all"
	if len(sub.FilterPolicy) > 0 {
		filterLabel = "filtered"
	}
	return fmt.Sprintf("%s %-4s %s %s", direction, strings.ToUpper(sub.Protocol), filterLabel, name)
}

func (m Model) routeCountForTopic(topicARN string) int {
	if topicARN == "" {
		return 0
	}
	count := 0
	for _, sub := range m.allSubscriptions {
		if sub.TopicARN == topicARN {
			count++
		}
	}
	return count
}

func (m Model) routeCountForQueue(queue domain.SQSQueue) int {
	if len(m.allSubscriptions) == 0 {
		return 0
	}
	count := 0
	for _, sub := range m.allSubscriptions {
		if sub.Endpoint == queue.URL || sub.Endpoint == queue.ARN {
			count++
		}
	}
	return count
}

func (m Model) renderTitledPanel(width, height int, title string, content string, active bool, accent lipgloss.Color) string {
	if width < 12 {
		width = 12
	}
	if height < 5 {
		height = 5
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 0).
		BorderTop(false)

	border := borderStyle.GetBorderStyle()
	maxTitleWidth := width - 4
	if maxTitleWidth < 5 {
		maxTitleWidth = 5
	}
	truncatedTitle := truncateToWidth(title, maxTitleWidth)
	titleText := fmt.Sprintf("─[%s]─", truncatedTitle)
	titleWidth := lipgloss.Width(titleText)

	repeatCount := width - titleWidth - 2
	if repeatCount < 0 {
		repeatCount = 0
	}
	topLine := border.TopLeft + titleText + strings.Repeat(border.Top, repeatCount) + border.TopRight

	var styledTopLine string
	topLineStyle := lipgloss.NewStyle().Foreground(accent)
	if active {
		topLineStyle = topLineStyle.Bold(true)
	}
	styledTopLine = topLineStyle.Render(topLine)

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	lines := strings.Split(content, "\n")
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}
	clippedContent := strings.Join(lines, "\n")

	panel := borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(clippedContent)

	return lipgloss.JoinVertical(lipgloss.Left, styledTopLine, panel)
}

func (m Model) renderLine(ctx selectionContext, index int, cursor int, text string, width int, active bool) string {
	selected := index == cursor && active
	rangeSelected := m.isIndexSelected(ctx, index)

	bg := lipgloss.Color(ui.ColorBg)
	fg := lipgloss.Color(ui.ColorFg)
	if rangeSelected {
		bg = lipgloss.Color(ui.ColorSelected)
	}
	if selected {
		bg = lipgloss.Color(ui.ColorHighlight)
		fg = lipgloss.Color(ui.ColorBg)
	}

	lineStyle := lipgloss.NewStyle().Background(bg).Foreground(fg)
	if selected || rangeSelected {
		lineStyle = lineStyle.Bold(true)
	}

	prefix := "  "
	if selected {
		prefix = "> "
	} else if rangeSelected {
		prefix = "* "
	}

	maxTextWidth := width - lipgloss.Width(prefix)
	if maxTextWidth < 0 {
		maxTextWidth = 0
	}
	text = truncateToWidth(text, maxTextWidth)

	line := prefix + text

	padLen := width - lipgloss.Width(line)
	if padLen > 0 {
		line += strings.Repeat(" ", padLen)
	}

	return lineStyle.Render(line)
}

func (m Model) renderS3Panel() string {
	leftWidth := int(float64(m.width-4) * m.panelRatioFor(panelS3))
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := m.width - 4 - leftWidth
	height := m.mainPanelHeight()
	innerHeight := height - 2

	var leftBuilder strings.Builder
	leftBuilder.WriteString("\n")
	if len(m.buckets) == 0 {
		leftBuilder.WriteString(m.styles.InfoText.Render("  No buckets found"))
	} else {
		totalObjects := len(m.objects)
		leftBuilder.WriteString("  " + m.renderPillRows(leftWidth-4,
			m.renderMetricPill("Buckets", fmt.Sprintf("%d", len(m.buckets)), lipgloss.Color(ui.ColorMono)),
			m.renderMetricPill("Objects", fmt.Sprintf("%d", totalObjects), lipgloss.Color(ui.ColorStack)),
		) + "\n\n")

		maxVisibleBuckets := innerHeight - 7
		startB, endB := visibleRange(len(m.buckets), m.selectedBucketIndex, maxVisibleBuckets)

		for i := startB; i < endB; i++ {
			bucketName := m.buckets[i].Name
			objectCountLabel := ""
			if m.s3ObjectsLoadedFor == bucketName {
				objectCountLabel = fmt.Sprintf(" %d obj", len(m.s3ObjectsCache[bucketName]))
			}
			leftBuilder.WriteString(m.renderLine(selectionS3Buckets, i, m.selectedBucketIndex, bucketName+objectCountLabel, leftWidth-4, m.s3Focus == focusBuckets) + "\n")
		}
		if len(m.buckets) > 0 && m.selectedBucketIndex < len(m.buckets) {
			selectedBucket := m.buckets[m.selectedBucketIndex]
			leftBuilder.WriteString("\n")
			leftBuilder.WriteString(m.renderSectionCaption("Selected Bucket") + "\n")
			leftBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth("  Name: "+selectedBucket.Name, leftWidth-4)) + "\n")
			if cached, ok := m.s3ObjectsCache[selectedBucket.Name]; ok {
				leftBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth(fmt.Sprintf("  Cached objects: %d", len(cached)), leftWidth-4)) + "\n")
			}
		}
	}
	leftPanel := m.renderTitledPanel(leftWidth, height, "Buckets", leftBuilder.String(), m.s3Focus == focusBuckets, lipgloss.Color(ui.ColorMono))

	var rightBuilder strings.Builder
	rightBuilder.WriteString("\n")
	bucketName := "Select a Bucket"
	if len(m.buckets) > 0 && m.selectedBucketIndex < len(m.buckets) {
		bucketName = m.buckets[m.selectedBucketIndex].Name
	}

	if len(m.objects) == 0 {
		rightBuilder.WriteString(m.styles.InfoText.Render("  This bucket has no files/keys."))
		rightBuilder.WriteString("\n")
		rightBuilder.WriteString(m.styles.InfoText.Render("  Press [u] to upload a file into this bucket."))
	} else {
		keyWidth := rightWidth - 45
		if keyWidth < 15 {
			keyWidth = 15
		}
		header := fmt.Sprintf("  %-*s %-12s %-25s", keyWidth, "File Key", "Size", "Last Modified")
		rightBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth(header, rightWidth-4)) + "\n")
		rightBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth(strings.Repeat("—", rightWidth-8), rightWidth-4)) + "\n")

		maxVisibleObjs := innerHeight - 5
		if maxVisibleObjs < 1 {
			maxVisibleObjs = 1
		}

		startO := 0
		if m.selectedObjectIndex >= maxVisibleObjs {
			startO = m.selectedObjectIndex - maxVisibleObjs + 1
		}
		endO := startO + maxVisibleObjs
		if endO > len(m.objects) {
			endO = len(m.objects)
			startO = endO - maxVisibleObjs
			if startO < 0 {
				startO = 0
			}
		}

		for i := startO; i < endO; i++ {
			o := m.objects[i]
			sizeStr := humanBytes(o.Size)

			keyStr := o.Key
			if len(keyStr) > keyWidth {
				keyStr = "…" + keyStr[len(keyStr)-(keyWidth-1):]
			}
			if isLikelyImageKey(o.Key) {
				keyStr = "img " + keyStr
			}

			row := fmt.Sprintf("%-*s %-12s %-25s", keyWidth, keyStr, sizeStr, o.LastModified)
			rightBuilder.WriteString(m.renderLine(selectionS3Objects, i, m.selectedObjectIndex, row, rightWidth-4, m.s3Focus == focusObjects) + "\n")
		}
	}
	if len(m.objects) > 0 && m.selectedObjectIndex < len(m.objects) {
		selected := m.objects[m.selectedObjectIndex]
		rightBuilder.WriteString("\n")
		rightBuilder.WriteString(m.renderSectionCaption("Selected Object") + "\n")
		rightBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth("  Key: "+selected.Key, rightWidth-4)) + "\n")
		rightBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth("  Size: "+humanBytes(selected.Size)+"  Last modified: "+selected.LastModified, rightWidth-4)) + "\n")
		if isLikelyImageKey(selected.Key) {
			rightBuilder.WriteString(m.styles.InfoText.Render(truncateToWidth("  This looks like an image. Press [b] to open it in the browser.", rightWidth-4)) + "\n")
		}
	}
	rightPanel := m.renderTitledPanel(rightWidth, height, fmt.Sprintf("Objects: %s", bucketName), rightBuilder.String(), m.s3Focus == focusObjects, lipgloss.Color(ui.ColorStack))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m Model) renderS3UploadModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Upload S3 Object") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Upload a local file into the selected bucket.") + "\n\n")

	pathLabel := "Local file path:"
	keyLabel := "Object key:"
	if m.s3UploadFocus == 0 {
		pathLabel = "> Local file path:"
	} else {
		keyLabel = "> Object key:"
	}

	builder.WriteString(m.styles.InfoText.Render(pathLabel) + "\n")
	builder.WriteString(m.s3UploadPathInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(keyLabel) + "\n")
	builder.WriteString(m.s3UploadKeyInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("[Enter] Upload  |  [Tab] Switch field  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderS3PreviewModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("S3 Object Preview") + "\n\n")
	if m.selectedObjectIndex >= len(m.objects) {
		builder.WriteString(m.styles.InfoText.Render("No object selected."))
		return builder.String()
	}

	obj := m.objects[m.selectedObjectIndex]
	builder.WriteString(m.styles.InfoText.Render("Bucket: "+m.selectedS3BucketName()) + "\n")
	builder.WriteString(m.styles.InfoText.Render("Key: "+obj.Key) + "\n")
	builder.WriteString(m.styles.InfoText.Render("Size: "+humanBytes(obj.Size)) + "\n")
	builder.WriteString(m.styles.InfoText.Render("Last modified: "+obj.LastModified) + "\n\n")

	if isLikelyImageKey(obj.Key) {
		builder.WriteString(m.renderMetricPill("Type", "Image asset", lipgloss.Color(ui.ColorEmerald)) + "\n\n")
		builder.WriteString(m.styles.InfoText.Render("Open in browser to preview the rendered image quickly.") + "\n")
	} else {
		builder.WriteString(m.renderMutedPill("Binary or document object") + "\n\n")
		builder.WriteString(m.styles.InfoText.Render("Use browser open for remote preview or download it locally.") + "\n")
	}

	builder.WriteString("\n")
	builder.WriteString(m.styles.InfoText.Render("[b] Open in browser  |  [w] Download  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderSQSPanel() string {
	leftWidth := int(float64(m.width-4) * m.panelRatioFor(panelSQS))
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := m.width - 4 - leftWidth
	height := m.mainPanelHeight()
	innerHeight := height - 2

	var leftBuilder strings.Builder
	leftBuilder.WriteString("\n")
	if len(m.queues) == 0 {
		leftBuilder.WriteString(m.styles.InfoText.Render("  No SQS queues discovered."))
	} else {
		totalAvailable := 0
		totalInflight := 0
		totalDelayed := 0
		for _, q := range m.queues {
			totalAvailable += q.MessagesAvailable
			totalInflight += q.MessagesNotVisible
			totalDelayed += q.MessagesDelayed
		}
		leftBuilder.WriteString("  " + m.renderPillRows(leftWidth-4,
			m.renderMetricPill("Queues", fmt.Sprintf("%d", len(m.queues)), lipgloss.Color(ui.ColorMono)),
			m.renderMetricPill("Routes", fmt.Sprintf("%d", len(m.allSubscriptions)), lipgloss.Color(ui.ColorStack)),
			m.renderMetricPill("Ready", fmt.Sprintf("%d", totalAvailable), lipgloss.Color(ui.ColorEmerald)),
			m.renderMetricPill("In-flight", fmt.Sprintf("%d", totalInflight), lipgloss.Color(ui.ColorAmber)),
		) + "\n")
		if totalDelayed > 0 {
			leftBuilder.WriteString("  " + m.renderPillRows(leftWidth-4, m.renderMutedPill(fmt.Sprintf("Delayed %d", totalDelayed))) + "\n")
		}
		leftBuilder.WriteString("\n")

		maxVisibleQueues := innerHeight - 7
		startQ, endQ := visibleRange(len(m.queues), m.selectedQueueIndex, maxVisibleQueues)
		for i := startQ; i < endQ; i++ {
			q := m.queues[i]
			line := fmt.Sprintf("%2d routes  %s", m.routeCountForQueue(q), q.Name)
			if q.MessagesAvailable > 0 || q.MessagesNotVisible > 0 || q.MessagesDelayed > 0 {
				line += fmt.Sprintf("  r:%d f:%d d:%d", q.MessagesAvailable, q.MessagesNotVisible, q.MessagesDelayed)
			}
			leftBuilder.WriteString(m.renderLine(selectionSQSQueues, i, m.selectedQueueIndex, line, leftWidth-4, m.sqsFocus == focusQueues) + "\n")
		}
		if len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
			selected := m.queues[m.selectedQueueIndex]
			leftBuilder.WriteString("\n")
			leftBuilder.WriteString(m.renderSectionCaption("Selected Queue") + "\n")
			leftBuilder.WriteString(m.renderDetailLine("Ready", fmt.Sprintf("%d", selected.MessagesAvailable), leftWidth-4) + "\n")
			leftBuilder.WriteString(m.renderDetailLine("In-flight", fmt.Sprintf("%d", selected.MessagesNotVisible), leftWidth-4) + "\n")
			leftBuilder.WriteString(m.renderDetailLine("Delayed", fmt.Sprintf("%d", selected.MessagesDelayed), leftWidth-4) + "\n")
			leftBuilder.WriteString(m.renderDetailLine("URL", selected.URL, leftWidth-4) + "\n")
			leftBuilder.WriteString(m.renderDetailLine("ARN", selected.ARN, leftWidth-4) + "\n")
		}
	}
	leftPanel := m.renderTitledPanel(leftWidth, height, "SQS Queues", leftBuilder.String(), m.sqsFocus == focusQueues, lipgloss.Color(ui.ColorMono))

	var rightBuilder strings.Builder
	rightBuilder.WriteString("\n")
	if len(m.queues) == 0 {
		rightBuilder.WriteString(m.styles.InfoText.Render("  No queue selected"))
	} else if m.selectedQueueIndex >= len(m.queues) {
		rightBuilder.WriteString(m.styles.InfoText.Render("  Select a queue to view subscriptions"))
	} else {
		q := m.queues[m.selectedQueueIndex]
		rightBuilder.WriteString("  " + m.renderPillRows(rightWidth-4,
			m.renderMetricPill("Target", q.Name, lipgloss.Color(ui.ColorStack)),
			m.renderMutedPill(fmt.Sprintf("%d routes", len(m.queueSubscriptions))),
			m.renderMutedPill(fmt.Sprintf("%d ready", q.MessagesAvailable)),
		) + "\n\n")

		if len(m.queueSubscriptions) == 0 {
			rightBuilder.WriteString(m.styles.InfoText.Render("  No SNS subscriptions") + "\n")
			rightBuilder.WriteString(m.styles.InfoText.Render("  Press [b] to connect one or more SNS topics."))
		} else {
			rightBuilder.WriteString(m.renderSectionCaption("Incoming Topic Routes") + "\n")
			maxVisibleSubs := innerHeight - 8
			startS, endS := visibleRange(len(m.queueSubscriptions), m.selectedQueueSubIndex, maxVisibleSubs)
			for i := startS; i < endS; i++ {
				sub := m.queueSubscriptions[i]
				filterTag := "all"
				if len(sub.FilterPolicy) > 0 {
					filterTag = "filtered"
				}
				line := fmt.Sprintf("%-20s %-4s %s", shortResourceName(sub.TopicARN), strings.ToUpper(sub.Protocol), filterTag)
				rightBuilder.WriteString(m.renderLine(selectionSQSSubs, i, m.selectedQueueSubIndex, line, rightWidth-4, m.sqsFocus == focusQueueSubs) + "\n")
			}
			if m.selectedQueueSubIndex < len(m.queueSubscriptions) {
				selected := m.queueSubscriptions[m.selectedQueueSubIndex]
				rightBuilder.WriteString("\n")
				rightBuilder.WriteString(m.renderSectionCaption("Selected Link") + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Topic", selected.TopicARN, rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Protocol", strings.ToUpper(selected.Protocol), rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Endpoint", shortResourceName(selected.Endpoint), rightWidth-4) + "\n")
				if len(selected.FilterPolicy) > 0 {
					rightBuilder.WriteString(m.renderDetailLine("Filters", formatFilterPolicy(selected.FilterPolicy), rightWidth-4) + "\n")
					rightBuilder.WriteString(m.renderDetailLine("Scope", formatFilterScope(selected.FilterScope), rightWidth-4) + "\n")
				}
			}
		}
	}
	rightPanel := m.renderTitledPanel(rightWidth, height, "Subscriptions (SNS→SQS)", rightBuilder.String(), m.sqsFocus == focusQueueSubs, lipgloss.Color(ui.ColorStack))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m Model) renderSNSPanel() string {
	leftWidth := int(float64(m.width-4) * m.panelRatioFor(panelSNS))
	if leftWidth < 15 {
		leftWidth = 15
	}
	rightWidth := m.width - 4 - leftWidth
	height := m.mainPanelHeight()
	innerHeight := height - 2

	var leftBuilder strings.Builder
	leftBuilder.WriteString("\n")
	if len(m.topics) == 0 {
		leftBuilder.WriteString(m.styles.InfoText.Render("  No SNS topics found"))
	} else {
		leftBuilder.WriteString("  " + m.renderPillRows(leftWidth-4,
			m.renderMetricPill("Topics", fmt.Sprintf("%d", len(m.topics)), lipgloss.Color(ui.ColorMono)),
			m.renderMetricPill("Routes", fmt.Sprintf("%d", len(m.allSubscriptions)), lipgloss.Color(ui.ColorStack)),
		) + "\n\n")
		maxVisibleTopics := innerHeight - 7
		startT, endT := visibleRange(len(m.topics), m.selectedTopicIndex, maxVisibleTopics)
		for i := startT; i < endT; i++ {
			topic := m.topics[i]
			line := fmt.Sprintf("%2d routes  %s", m.routeCountForTopic(topic.ARN), topic.Name)
			leftBuilder.WriteString(m.renderLine(selectionSNSTopics, i, m.selectedTopicIndex, line, leftWidth-4, m.snsSubFocus == focusTopics) + "\n")
		}
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			selected := m.topics[m.selectedTopicIndex]
			leftBuilder.WriteString("\n")
			leftBuilder.WriteString(m.renderSectionCaption("Selected Topic") + "\n")
			leftBuilder.WriteString(m.renderDetailLine("ARN", selected.ARN, leftWidth-4) + "\n")
		}
	}
	leftPanel := m.renderTitledPanel(leftWidth, height, "Topics", leftBuilder.String(), m.snsSubFocus == focusTopics, lipgloss.Color(ui.ColorMono))

	var rightBuilder strings.Builder
	rightBuilder.WriteString("\n")
	if len(m.topics) == 0 {
		rightBuilder.WriteString(m.styles.InfoText.Render("  No topic selected"))
	} else if m.selectedTopicIndex >= len(m.topics) {
		rightBuilder.WriteString(m.styles.InfoText.Render("  Select a topic to view subscriptions"))
	} else {
		topicARN := m.topics[m.selectedTopicIndex].ARN
		rightBuilder.WriteString("  " + m.renderPillRows(rightWidth-4,
			m.renderMetricPill("Topic", shortResourceName(topicARN), lipgloss.Color(ui.ColorStack)),
			m.renderMutedPill(fmt.Sprintf("%d outgoing", m.snsOutgoingCount)),
			m.renderMutedPill(fmt.Sprintf("%d incoming", max(0, len(m.subscriptions)-m.snsOutgoingCount))),
		) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("ARN", topicARN, rightWidth-4) + "\n\n")

		if len(m.subscriptions) == 0 {
			rightBuilder.WriteString(m.styles.InfoText.Render("  No active subscriptions") + "\n")
			rightBuilder.WriteString(m.styles.InfoText.Render("  Press [c] to create one"))
		} else {
			rightBuilder.WriteString(m.renderSectionCaption("Routing Links") + "\n")
			maxVisibleSubs := innerHeight - 9
			startS, endS := visibleRange(len(m.subscriptions), m.selectedSubIndex, maxVisibleSubs)
			for i := startS; i < endS; i++ {
				sub := m.subscriptions[i]
				direction := "OUT"
				if i >= m.snsOutgoingCount {
					direction = "IN "
				}
				line := renderRouteSummary(sub, m.managedSubs, direction)
				if len(sub.FilterPolicy) > 0 {
					scopeTag := "attr"
					if sub.FilterScope == domain.SubscriptionFilterScopeMessageBody {
						scopeTag = "body"
					}
					line += " " + scopeTag
				}
				rightBuilder.WriteString(m.renderLine(selectionSNSSubs, i, m.selectedSubIndex, line, rightWidth-4, m.snsSubFocus == focusSubs) + "\n")
			}
			if m.selectedSubIndex < len(m.subscriptions) {
				sub := m.subscriptions[m.selectedSubIndex]
				rightBuilder.WriteString("\n")
				rightBuilder.WriteString(m.renderSectionCaption("Selected Route") + "\n")
				direction := "outgoing"
				if m.selectedSubIndex >= m.snsOutgoingCount {
					direction = "incoming"
				}
				rightBuilder.WriteString(m.renderDetailLine("Direction", direction, rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Protocol", strings.ToUpper(sub.Protocol), rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Source", shortResourceName(sub.TopicARN), rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Endpoint", sub.Endpoint, rightWidth-4) + "\n")
				if len(sub.FilterPolicy) > 0 {
					rightBuilder.WriteString(m.renderDetailLine("Filter", formatFilterPolicy(sub.FilterPolicy), rightWidth-4) + "\n")
					rightBuilder.WriteString(m.renderDetailLine("Scope", formatFilterScope(sub.FilterScope), rightWidth-4) + "\n")
				}
			}
		}
	}
	rightPanel := m.renderTitledPanel(rightWidth, height, "Subscriptions", rightBuilder.String(), m.snsSubFocus == focusSubs, lipgloss.Color(ui.ColorStack))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func maskSecretValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	if len(value) <= 8 {
		return strings.Repeat("•", len(value))
	}
	return value[:2] + strings.Repeat("•", len(value)-6) + value[len(value)-2:]
}

func formatSecretValueDisplay(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "(empty)"
	}

	normalized := normalizeSecretValueEscapes(trimmed)

	if json.Valid([]byte(normalized)) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(normalized), "", "  "); err == nil {
			return pretty.String()
		}
	}

	return value
}

func normalizeSecretValueEscapes(value string) string {
	return strings.NewReplacer(`\u0022`, `\"`, `\u0026`, `&`, `\u003c`, `<`, `\u003e`, `>`).Replace(value)
}

func renderSecretValuePreview(value string, width, maxLines int) string {
	formatted := wrapSecretValueDisplay(value, width)
	if width < 1 {
		width = 1
	}
	lines := strings.Split(formatted, "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = append(lines[:maxLines], "…")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderSecretsPanel() string {
	leftWidth := int(float64(m.width-4) * m.panelRatioFor(panelSecrets))
	if leftWidth < 18 {
		leftWidth = 18
	}
	rightWidth := m.width - 4 - leftWidth
	height := m.mainPanelHeight()
	innerHeight := height - 2

	var leftBuilder strings.Builder
	leftBuilder.WriteString("\n")
	if len(m.secrets) == 0 {
		leftBuilder.WriteString(m.styles.InfoText.Render("  No secrets found"))
	} else {
		leftBuilder.WriteString("  " + m.renderPillRows(leftWidth-4,
			m.renderMetricPill("Secrets", fmt.Sprintf("%d", len(m.secrets)), lipgloss.Color(ui.ColorMono)),
			m.renderMetricPill("Versions", fmt.Sprintf("%d", len(m.secretVersions)), lipgloss.Color(ui.ColorStack)),
		) + "\n\n")

		maxVisible := innerHeight - 7
		start, end := visibleRange(len(m.secrets), m.selectedSecretIndex, maxVisible)
		for i := start; i < end; i++ {
			secret := m.secrets[i]
			leftBuilder.WriteString(m.renderLine(selectionSecrets, i, m.selectedSecretIndex, secret.Name, leftWidth-4, m.secretsFocus == focusSecrets) + "\n")
		}
	}
	leftPanel := m.renderTitledPanel(leftWidth, height, "Secrets Manager", leftBuilder.String(), m.secretsFocus == focusSecrets, lipgloss.Color(ui.ColorMono))

	var rightBuilder strings.Builder
	rightBuilder.WriteString("\n")
	if len(m.secrets) == 0 {
		rightBuilder.WriteString(m.styles.InfoText.Render("  No secret selected"))
	} else if m.selectedSecretIndex >= len(m.secrets) {
		rightBuilder.WriteString(m.styles.InfoText.Render("  Select a secret to inspect"))
	} else {
		selected := m.secrets[m.selectedSecretIndex]
		rightBuilder.WriteString("  " + m.renderPillRows(rightWidth-4,
			m.renderMetricPill("Secret", selected.Name, lipgloss.Color(ui.ColorStack)),
			m.renderMutedPill(fmt.Sprintf("%d versions", len(m.secretVersions))),
		) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("Description", selected.Description, rightWidth-4) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("Last changed", selected.LastChangedDate, rightWidth-4) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("Primary region", selected.PrimaryRegion, rightWidth-4) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("KMS key", selected.KMSKeyID, rightWidth-4) + "\n")
		rightBuilder.WriteString(m.renderDetailLine("Rotation", fmt.Sprintf("%t", selected.RotationEnabled), rightWidth-4) + "\n\n")

		if len(m.secretVersions) == 0 {
			rightBuilder.WriteString(m.styles.InfoText.Render("  No version metadata available") + "\n")
		} else {
			rightBuilder.WriteString(m.renderSectionCaption("Versions") + "\n")
			maxVisible := innerHeight - 11
			start, end := visibleRange(len(m.secretVersions), m.selectedSecretVersionIndex, maxVisible)
			for i := start; i < end; i++ {
				version := m.secretVersions[i]
				line := secretVersionVisualLabel(i, version)
				rightBuilder.WriteString(m.renderLine(selectionSecretVersions, i, m.selectedSecretVersionIndex, line, rightWidth-4, m.secretsFocus == focusSecretVersions) + "\n")
			}
			if m.selectedSecretVersionIndex < len(m.secretVersions) {
				version := m.secretVersions[m.selectedSecretVersionIndex]
				rightBuilder.WriteString("\n")
				rightBuilder.WriteString(m.renderSectionCaption("Version Details") + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Label", secretVersionVisualLabel(m.selectedSecretVersionIndex, version), rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Version ID", version.VersionID, rightWidth-4) + "\n")
				rightBuilder.WriteString(m.renderDetailLine("Created", version.CreatedDate, rightWidth-4) + "\n")
				if m.secretVersionIsCurrent(version) {
					rightBuilder.WriteString(m.styles.InfoText.Render("  This version is current.") + "\n")
				} else {
					rightBuilder.WriteString(m.styles.InfoText.Render("  [r] Make selected version current") + "\n")
				}
			}
		}
	}
	rightPanel := m.renderTitledPanel(rightWidth, height, "Secrets Details", rightBuilder.String(), m.secretsFocus == focusSecretVersions, lipgloss.Color(ui.ColorStack))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m Model) renderSecretCreateModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Create Secret") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Secret name:") + "\n")
	builder.WriteString(m.secretCreateNameInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Secret value:") + "\n")
	builder.WriteString(m.secretCreateValueInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("[Ctrl+S] Create  |  [Tab] Switch field  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderSecretUpdateModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Update Secret Value") + "\n\n")
	if len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
		builder.WriteString(m.styles.InfoText.Render("Secret: "+m.secrets[m.selectedSecretIndex].Name) + "\n")
		builder.WriteString(m.styles.InfoText.Render("Name stays unchanged in AWS.") + "\n\n")
	}
	builder.WriteString(m.styles.InfoText.Render("New secret value:") + "\n")
	builder.WriteString(m.secretUpdateValueInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("[Ctrl+S] Save  |  [Ctrl+L] Clear JSON  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderSecretDeleteConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Confirm Delete Secret") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Delete secret %q?", m.secretDeleteName)) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSecretRestoreConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Restore Secret") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Restore secret %q from recovery window?", m.secretDeleteName)) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSecretPromoteConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Promote Secret Version") + "\n\n")
	label := m.secretPromoteVersionLabel
	if label == "" {
		label = m.secretPromoteVersionID
	}
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Make version %q current?", label)) + "\n")
	if m.secretPromoteCurrentID != "" {
		current := m.secretPromoteCurrentLabel
		if current == "" {
			current = m.secretPromoteCurrentID
		}
		builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Current version: %s", current)) + "\n\n")
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSecretClipboardConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Copy Secret to Clipboard") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("The secret value will be placed on your system clipboard.") + "\n")
	builder.WriteString(m.styles.InfoText.Render("Other applications may be able to read it from there.") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSecretValueModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Secret Value") + "\n\n")
	if m.secretValueViewport.Width > 0 && m.secretValueViewport.Height > 0 {
		builder.WriteString(m.secretValueViewport.View())
		builder.WriteString("\n\n")
	} else {
		content := m.secretValue.SecretString
		if strings.TrimSpace(content) == "" {
			content = m.secretValue.SecretBinaryBase64
		}
		builder.WriteString(m.styles.InfoText.Render(renderSecretValuePreview(content, 78, 18)) + "\n\n")
	}
	builder.WriteString(m.styles.InfoText.Render("[e] Edit  |  [c] Copy  |  [j/k] Scroll  |  [PgUp/PgDn] Page  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderSnsCreateTopicModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Create SNS Topic") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Enter topic name:") + "\n")
	builder.WriteString(m.snsCreateTopicInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Create | [Esc] to Close"))
	return builder.String()
}

func (m Model) renderS3CreateBucketModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Create S3 Bucket") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Enter bucket name:") + "\n")
	builder.WriteString(m.s3CreateInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Create | [Esc] to Close"))
	return builder.String()
}

func (m Model) renderS3CreateFolderModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Create S3 Folder") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Enter folder prefix:") + "\n")
	builder.WriteString(m.s3FolderInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Create | [Esc] to Close"))
	return builder.String()
}

func (m Model) renderSqsCreateQueueModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Create SQS Queue") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Enter queue name:") + "\n")
	builder.WriteString(m.sqsCreateInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Create | [Esc] to Close"))
	return builder.String()
}

func (m Model) renderSnsSimpleSubModal() string {
	var builder strings.Builder

	switch m.snsSimpleSubStep {
	case 0:
		builder.WriteString(m.styles.Title.Render("Subscribe — Select Source Topic") + "\n\n")
		builder.WriteString(m.styles.InfoText.Render("Select source SNS topic to subscribe from:") + "\n")
		if len(m.snsSimpleSubSources) == 0 {
			builder.WriteString(m.styles.InfoText.Render("  No other topics available"))
		} else {
			maxVisible := len(m.snsSimpleSubSources)
			if maxVisible > 8 {
				maxVisible = 8
			}
			for i := 0; i < maxVisible && i < len(m.snsSimpleSubSources); i++ {
				t := m.snsSimpleSubSources[i]
				if i == m.snsSimpleSubCursor {
					builder.WriteString(m.styles.SelectedListItem.Render("> "+t.Name) + "\n")
				} else {
					builder.WriteString(m.styles.ListItem.Render("  "+t.Name) + "\n")
				}
			}
		}
		builder.WriteString("\n")
		builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Select | [Up/Down] Navigate | [Esc] Close"))

	case 1:
		builder.WriteString(m.styles.Title.Render("Subscribe — Event Types (optional)") + "\n\n")
		builder.WriteString(m.styles.InfoText.Render("Filter by event types (comma-separated, or leave empty for all):") + "\n")
		builder.WriteString(m.snsSimpleSubEventInput.View() + "\n\n")
		builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Subscribe | [Esc] Back"))
	}

	return builder.String()
}

func (m Model) renderSnsBatchSubModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Batch Subscribe — Select Source Topics") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Toggle with [Space], confirm with [Enter]:") + "\n")

	if len(m.snsBatchSubList) == 0 {
		builder.WriteString(m.styles.InfoText.Render("  No topics available"))
	} else {
		maxVisible := len(m.snsBatchSubList)
		if maxVisible > 10 {
			maxVisible = 10
		}
		startB := 0
		if m.snsBatchSubCursor >= maxVisible {
			startB = m.snsBatchSubCursor - maxVisible + 1
		}
		endB := startB + maxVisible
		if endB > len(m.snsBatchSubList) {
			endB = len(m.snsBatchSubList)
			startB = endB - maxVisible
			if startB < 0 {
				startB = 0
			}
		}
		for i := startB; i < endB; i++ {
			opt := m.snsBatchSubList[i]
			prefix := "[ ]"
			if opt.checked {
				prefix = "[x]"
			}
			line := prefix + " " + opt.label
			if i == m.snsBatchSubCursor {
				builder.WriteString(m.styles.SelectedListItem.Render("> "+line) + "\n")
			} else {
				builder.WriteString(m.styles.ListItem.Render("  "+line) + "\n")
			}
		}
	}

	builder.WriteString("\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Confirm | [Space] Toggle | [Esc] Close"))
	return builder.String()
}

func (m Model) renderSnsYamlImportModal() string {
	var builder strings.Builder
	topicLabel := "(no topic selected)"
	if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
		topicLabel = m.topics[m.selectedTopicIndex].Name
	} else if m.snsYamlCurrentTopicARN != "" {
		topicLabel = shortResourceName(m.snsYamlCurrentTopicARN)
	}
	enterHint := m.styles.InfoText.Render("Enter → new line")
	applyHint := m.styles.InfoText.Render("[Ctrl+S] Apply subscriptions")
	escHint := m.styles.InfoText.Render("[Esc] Save & close")

	builder.WriteString(m.styles.Title.Render("Subscription YAML — "+topicLabel) + "\n")
	builder.WriteString(m.styles.InfoText.Render("  • Each topic keeps its own YAML file") + "\n")
	builder.WriteString(m.styles.InfoText.Render("  • `topic` can be omitted here; it defaults to the open topic") + "\n")
	builder.WriteString(m.styles.InfoText.Render("  • `queue` can be omitted; Monostack tries `default_queue` then the sibling `-sqs` queue") + "\n")
	builder.WriteString(m.styles.InfoText.Render("  • `default_filter_scope` defaults to message_body") + "\n")
	builder.WriteString(m.styles.InfoText.Render("  • Fields: version, default_queue?, default_filter_scope?, name, topic?, event_type[], filter_scope?, queue?") + "\n\n")

	builder.WriteString(m.snsYamlImportTextarea.View() + "\n\n")

	builder.WriteString(applyHint + "  │  " + enterHint + "  │  " + escHint + "  │  " + m.styles.InfoText.Render("[Ctrl+K] Discard"))
	return builder.String()
}

func (m Model) renderSnsYamlApplyConfirmModal() string {
	var builder strings.Builder
	topicLabel := shortResourceName(m.snsYamlCurrentTopicARN)
	if topicLabel == "" {
		topicLabel = "selected topic"
	}
	builder.WriteString(m.styles.Title.Render("Apply YAML Subscriptions?") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("This will save the YAML for "+topicLabel+" and create/update SNS -> SQS routes.") + "\n")
	builder.WriteString(m.styles.InfoText.Render("Subscriptions with stale filter scope will be repaired automatically.") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("[Y] Apply  |  [N] Cancel  |  [Esc] Cancel"))
	return builder.String()
}

func (m Model) renderSnsEditSubModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Edit Subscription Filter") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Update event types (comma-separated):") + "\n")
	builder.WriteString(m.snsSubEditInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Save | [Esc] to Close"))
	return builder.String()
}

func (m Model) renderSqsBatchSubModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Subscribe SNS Topics → This Queue") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Toggle with [Space], confirm with [Enter]:") + "\n")

	if len(m.sqsBatchSubList) == 0 {
		builder.WriteString(m.styles.InfoText.Render("  No topics available"))
	} else {
		maxVisible := len(m.sqsBatchSubList)
		if maxVisible > 10 {
			maxVisible = 10
		}
		startB := 0
		if m.sqsBatchSubCursor >= maxVisible {
			startB = m.sqsBatchSubCursor - maxVisible + 1
		}
		endB := startB + maxVisible
		if endB > len(m.sqsBatchSubList) {
			endB = len(m.sqsBatchSubList)
			startB = endB - maxVisible
			if startB < 0 {
				startB = 0
			}
		}
		for i := startB; i < endB; i++ {
			opt := m.sqsBatchSubList[i]
			prefix := "[ ]"
			if opt.checked {
				prefix = "[x]"
			}
			line := prefix + " " + opt.label
			if i == m.sqsBatchSubCursor {
				builder.WriteString(m.styles.SelectedListItem.Render("> "+line) + "\n")
			} else {
				builder.WriteString(m.styles.ListItem.Render("  "+line) + "\n")
			}
		}
	}

	builder.WriteString("\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Confirm | [Space] Toggle | [Esc] Close"))
	return builder.String()
}

func (m Model) renderSnsConfirmDeleteModal() string {
	var builder strings.Builder
	label := m.snsDeleteLabel
	builder.WriteString(m.styles.Title.Render("Confirm Delete") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Delete %s?", label)) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderS3ConfirmDeleteModal() string {
	var builder strings.Builder
	if m.s3DeleteIsBucket {
		builder.WriteString(m.styles.Title.Render("Confirm Delete S3 Bucket") + "\n\n")
		builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Delete bucket \"%s\"?", m.s3DeleteBucket)) + "\n\n")
	} else {
		builder.WriteString(m.styles.Title.Render("Confirm Delete S3 Object") + "\n\n")
		builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Delete %s / %s?", m.s3DeleteBucket, m.s3DeleteKey)) + "\n\n")
	}
	builder.WriteString(m.styles.InfoText.Render("⚠  This cannot be undone.") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSqsConfirmDeleteModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Confirm Delete SQS Queue") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Delete queue  \"%s\"?", m.sqsDeleteQueueName)) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("⚠  All messages will be lost and this cannot be undone.") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSqsPurgeAllConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Confirm Purge All SQS Queues") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Purge %d loaded queues?", len(m.queues))) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderSqsSubDeleteConfirmModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Confirm Unsubscribe") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(fmt.Sprintf("Remove subscription %s?", m.sqsDeleteSubLabel)) + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Y] to Confirm | Any other key to Cancel"))
	return builder.String()
}

func (m Model) renderExportProfileModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Export Snapshot") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Save config, S3, SQS, SNS and Secrets state to a snapshot file.") + "\n")
	builder.WriteString(m.styles.InfoText.Render("Enter destination folder or file:") + "\n")
	builder.WriteString(m.exportPathInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Export | [Esc] to Cancel"))
	return builder.String()
}

func (m Model) renderImportProfileModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Load Snapshot") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Restore and apply a previously exported snapshot.") + "\n")
	builder.WriteString(m.styles.InfoText.Render("Enter snapshot path:") + "\n")
	builder.WriteString(m.importPathInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Load | [Esc] to Cancel"))
	return builder.String()
}

func (m Model) renderConfigPanel() string {
	var builder strings.Builder
	width := m.width - 4
	height := m.mainPanelHeight()
	innerHeight := height - 2

	labels := []string{
		"1. Profile Name:     ",
		"2. Endpoint URL:     ",
		"3. Region:           ",
		"4. Access Key ID:    ",
		"5. Secret Key:       ",
		"6. Mock Mode:        ",
		"7. Snapshot Path:    ",
		"8. Enabled Services: ",
	}

	maxVisibleFields := (innerHeight - 1) / 5
	if maxVisibleFields < 1 {
		maxVisibleFields = 1
	}
	if maxVisibleFields > len(m.settingsInputs) {
		maxVisibleFields = len(m.settingsInputs)
	}

	start := 0
	if m.focusedInput >= maxVisibleFields {
		start = m.focusedInput - maxVisibleFields + 1
	}
	end := start + maxVisibleFields
	if end > len(m.settingsInputs) {
		end = len(m.settingsInputs)
		start = end - maxVisibleFields
		if start < 0 {
			start = 0
		}
	}

	if start > 0 {
		builder.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSubtle)).Render("  ^ More settings above...") + "\n")
	} else {
		builder.WriteString("\n")
	}

	for i := start; i < end; i++ {
		var label string
		var field string

		if i == m.focusedInput {
			if m.settingsEditMode {
				label = m.styles.InputLabel.Render(labels[i]) + lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSuccess)).Bold(true).Render(" [EDITING]")
				field = m.styles.InputFocused.Render(m.settingsInputs[i].View())
			} else {
				label = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorHighlight)).Bold(true).Render("> " + labels[i][3:])
				field = m.styles.InputUnfocused.Copy().BorderForeground(lipgloss.Color(ui.ColorHighlight)).Render(m.settingsInputs[i].View())
			}
		} else {
			label = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSubtle)).Render(labels[i])
			field = m.styles.InputUnfocused.Render(m.settingsInputs[i].View())
		}
		builder.WriteString("  " + label + "\n" + field + "\n")
		if i < end-1 {
			builder.WriteString("\n")
		}
	}

	if end < len(m.settingsInputs) {
		builder.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSubtle)).Render("  v More settings below...") + "\n")
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorAccent)).Bold(true)
	builder.WriteString("\n")

	if m.config != nil && strings.HasPrefix(strings.ToLower(m.config.EndpointURL), "http://") {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorError)).Bold(true)
		builder.WriteString(warnStyle.Render("  WARNING: Endpoint uses HTTP - credentials sent in plaintext") + "\n\n")
	}

	builder.WriteString(hintStyle.Render("  [E] Export Snapshot") + "  " + hintStyle.Render("[L] Load Snapshot") + "  " + hintStyle.Render("[S] Save"))

	return m.renderTitledPanel(width, height, "Connection Settings", builder.String(), true, lipgloss.Color(ui.ColorStack))
}

func (m Model) renderSqsSendModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Publish SQS Message Payload") + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Type message body payload:") + "\n")
	builder.WriteString(m.sqsSendMessageInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render("Press [Enter] to Send | [Esc] to Close"))

	return builder.String()
}

func (m Model) renderSnsPublishModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Publish SNS Event Message") + "\n\n")

	bodyLabel := "Body:"
	attrsLabel := "Attrs:"
	if m.snsPublishInput.Focused() {
		bodyLabel = "> Body:"
	} else {
		attrsLabel = "> Attrs:"
	}
	if m.snsPublishAttrsInput.Value() != "" {
		attrsLabel += " (tab to switch)"
	}

	builder.WriteString(m.styles.InfoText.Render(bodyLabel) + "\n")
	builder.WriteString(m.snsPublishInput.View() + "\n\n")
	builder.WriteString(m.styles.InfoText.Render(attrsLabel) + "\n")
	builder.WriteString(m.snsPublishAttrsInput.View() + "\n\n")

	builder.WriteString(m.styles.InfoText.Render("[Enter] Publish  |  [Tab] Switch field  |  [Esc] Close"))
	return builder.String()
}

func (m Model) renderPeekModal() string {
	var builder strings.Builder
	builder.WriteString(m.styles.Title.Render("Peek Queue Messages (max 5)") + "\n\n")

	if len(m.peekMessages) == 0 {
		builder.WriteString(m.styles.InfoText.Render("No messages visible in the queue.") + "\n")
	} else {
		for i, msg := range m.peekMessages {
			builder.WriteString(m.styles.Highlight.Render(fmt.Sprintf("[%d] ID: %s", i+1, msg.ID)) + "\n")
			builder.WriteString(m.styles.ListItem.Render(msg.Body) + "\n\n")
		}
	}

	builder.WriteString(m.styles.InfoText.Render("Press [Esc / Enter] to Close"))
	return builder.String()
}

func (m Model) renderHelpModal() string {
	type helpEntry struct {
		key    string
		action string
	}
	type helpSection struct {
		title   string
		entries []helpEntry
	}

	tabKeys := []string{}
	for i := range m.visibleServicePanels() {
		tabKeys = append(tabKeys, fmt.Sprintf("%d", i+1))
	}
	tabKeys = append(tabKeys, "5")

	sections := []helpSection{
		{
			title: "NAVIGATION",
			entries: []helpEntry{
				{key: "jk | arrows", action: "Move selection"},
				{key: "h | l | arrows", action: "Switch panels"},
				{key: "< | >", action: "Resize panels"},
				{key: strings.Join(tabKeys, " | "), action: "Jump to tab"},
				{key: "tab", action: "Cycle focus"},
			},
		},
		{
			title: "GENERAL",
			entries: []helpEntry{
				{key: "? | ctrl+p", action: "Toggle help"},
				{key: "o", action: "Command logs"},
				{key: "space", action: "Start or extend selection"},
				{key: "y", action: "Copy selected text"},
				{key: "esc", action: "Back or cancel"},
				{key: "q", action: "Quit"},
			},
		},
	}

	if m.panelEnabled(panelS3) {
		sections = append(sections, helpSection{
			title: "AWS S3",
			entries: []helpEntry{
				{key: "enter | →", action: "Select bucket"},
				{key: "esc | ←", action: "Back to buckets"},
				{key: "c", action: "Create bucket"},
				{key: "f", action: "Create folder"},
				{key: "u", action: "Upload object"},
				{key: "v", action: "Preview object"},
				{key: "w", action: "Download object"},
				{key: "d", action: "Delete object"},
			},
		})
	}

	if m.panelEnabled(panelSQS) {
		sections = append(sections, helpSection{
			title: "AWS SQS",
			entries: []helpEntry{
				{key: "enter", action: "Inspect queue"},
				{key: "l | →", action: "Open routes"},
				{key: "v", action: "Peek messages"},
				{key: "s", action: "Send message"},
				{key: "p", action: "Purge queue"},
				{key: "P", action: "Purge all loaded queues"},
				{key: "b", action: "Subscribe topics"},
				{key: "d", action: "Delete queue"},
			},
		})
	}

	if m.panelEnabled(panelSNS) {
		sections = append(sections, helpSection{
			title: "AWS SNS",
			entries: []helpEntry{
				{key: "enter", action: "Inspect topic"},
				{key: "l | →", action: "Open subscriptions"},
				{key: "c", action: "Create topic"},
				{key: "b", action: "Batch subscribe"},
				{key: "i", action: "Import YAML"},
				{key: "e", action: "Edit filter"},
				{key: "s", action: "Publish message"},
				{key: "d", action: "Delete topic"},
			},
		})
	}

	if m.panelEnabled(panelSecrets) {
		sections = append(sections, helpSection{
			title: "SECRETS",
			entries: []helpEntry{
				{key: "enter", action: "Inspect secret"},
				{key: "l | h", action: "Switch list or versions"},
				{key: "r", action: "Promote version"},
				{key: "c", action: "Create secret"},
				{key: "u", action: "Update secret"},
				{key: "v", action: "Reveal value"},
				{key: "R", action: "Recover secret"},
				{key: "d", action: "Delete secret"},
			},
		})
	}

	sections = append(sections, helpSection{
		title: "SETTINGS",
		entries: []helpEntry{
			{key: "enter", action: "Edit field"},
			{key: "tab | s-tab", action: "Cycle inputs"},
			{key: "esc", action: "Stop editing"},
		},
	})

	renderSection := func(section helpSection) string {
		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorMono)).Bold(true)
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorFg))
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorStack)).Bold(true)

		var lines []string
		lines = append(lines, titleStyle.Render(" "+section.title+" "))
		for _, entry := range section.entries {
			lines = append(lines, fmt.Sprintf("  %s %s", keyStyle.Render(entry.key+":"), descStyle.Render(entry.action)))
		}
		return strings.Join(lines, "\n")
	}

	panelWidth, panelHeight := clampModalSize(m.width, m.height, 6, 72, 4, 20)

	innerWidth := panelWidth - 6
	if innerWidth < 60 {
		innerWidth = 60
	}

	columnCount := 1
	contentWidth := innerWidth
	switch {
	case contentWidth >= 140:
		columnCount = 3
	case contentWidth >= 100:
		columnCount = 2
	}
	if panelHeight >= 28 && columnCount < 3 && contentWidth >= 110 {
		columnCount++
	}
	if columnCount > len(sections) {
		columnCount = len(sections)
	}

	for columnCount > 1 {
		chunkSize := (len(sections) + columnCount - 1) / columnCount
		var columnWidths []int
		for i := 0; i < columnCount; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > len(sections) {
				end = len(sections)
			}
			if start >= len(sections) {
				break
			}
			width := 0
			for _, section := range sections[start:end] {
				for _, line := range strings.Split(renderSection(section), "\n") {
					if w := lipgloss.Width(line); w > width {
						width = w
					}
				}
			}
			columnWidths = append(columnWidths, width)
		}
		totalWidth := 0
		for _, width := range columnWidths {
			totalWidth += width
		}
		totalWidth += 4 * (len(columnWidths) - 1)
		if totalWidth <= contentWidth {
			break
		}
		columnCount--
	}

	chunkSize := (len(sections) + columnCount - 1) / columnCount
	var columns []string
	for i := 0; i < columnCount; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(sections) {
			end = len(sections)
		}
		if start >= len(sections) {
			break
		}
		var columnSections []string
		for _, section := range sections[start:end] {
			columnSections = append(columnSections, renderSection(section))
		}
		columns = append(columns, lipgloss.JoinVertical(lipgloss.Left, columnSections...))
	}

	body := columns[0]
	for i := 1; i < len(columns); i++ {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, "    ", columns[i])
	}

	innerHeight := panelHeight - 6
	if innerHeight < 12 {
		innerHeight = 12
	}
	vpHeight := innerHeight - 3
	if vpHeight < 5 {
		vpHeight = 5
	}
	if m.helpViewport.Width != innerWidth || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(innerWidth, vpHeight)
	} else {
		m.helpViewport.Width = innerWidth
		m.helpViewport.Height = vpHeight
	}

	m.helpViewport.SetContent(body)

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandStackStyle.Render("SHORTCUTS"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title),
		"",
		m.helpViewport.View(),
		"",
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSubtle)).Render("Press ESC or ctrl+p to close"),
		),
	)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(panelWidth).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(content),
	)
}

func clampModalSize(termWidth, termHeight, widthOffset, minWidth, heightOffset, minHeight int) (int, int) {
	w := termWidth - widthOffset
	if w < minWidth {
		w = minWidth
	}
	if w > termWidth-2 {
		w = termWidth - 2
	}
	h := termHeight - heightOffset
	if h < minHeight {
		h = minHeight
	}
	if h > termHeight {
		h = termHeight
	}
	return w, h
}
