package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

const maxCommandLogs = 120

func (m *Model) splashTickCmd() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return splashTickMsg{}
	})
}

func (m *Model) appendCommandLog(action, target, output string, err error) {
	entry := CommandLogEntry{
		Time:   time.Now(),
		Action: action,
		Target: target,
		Output: output,
		Error:  err,
	}
	m.commandLogs = append(m.commandLogs, entry)
	if len(m.commandLogs) > maxCommandLogs {
		m.commandLogs = append([]CommandLogEntry(nil), m.commandLogs[len(m.commandLogs)-maxCommandLogs:]...)
	}
	m.refreshLogViewport()
}

func (m Model) currentResourceTarget() string {
	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusObjects && m.selectedObjectIndex < len(m.objects) && m.selectedBucketIndex < len(m.buckets) {
			return m.buckets[m.selectedBucketIndex].Name + "/" + m.objects[m.selectedObjectIndex].Key
		}
		if m.selectedBucketIndex < len(m.buckets) {
			return m.buckets[m.selectedBucketIndex].Name
		}
	case panelSQS:
		if m.sqsFocus == focusQueueSubs && m.selectedQueueSubIndex < len(m.queueSubscriptions) {
			return shortResourceName(m.queueSubscriptions[m.selectedQueueSubIndex].TopicARN)
		}
		if m.selectedQueueIndex < len(m.queues) {
			return m.queues[m.selectedQueueIndex].Name
		}
	case panelSNS:
		if m.snsSubFocus == focusSubs && m.selectedSubIndex < len(m.subscriptions) {
			return shortResourceName(m.subscriptions[m.selectedSubIndex].Endpoint)
		}
		if m.selectedTopicIndex < len(m.topics) {
			return m.topics[m.selectedTopicIndex].Name
		}
	case panelSecrets:
		if m.secretsFocus == focusSecretVersions && m.selectedSecretVersionIndex < len(m.secretVersions) {
			return m.secretVersions[m.selectedSecretVersionIndex].VersionID
		}
		if m.selectedSecretIndex < len(m.secrets) {
			return m.secrets[m.selectedSecretIndex].Name
		}
	}
	return ""
}

func (m *Model) logSelectionContext() (selectionContext, int, bool) {
	if m.showLogsModal {
		return selectionCommandLogs, m.commandLogCursor, len(m.commandLogs) > 0
	}

	switch m.activeTab {
	case panelS3:
		switch m.s3Focus {
		case focusBuckets:
			return selectionS3Buckets, m.selectedBucketIndex, len(m.buckets) > 0
		case focusObjects:
			return selectionS3Objects, m.selectedObjectIndex, len(m.objects) > 0
		}
	case panelSQS:
		switch m.sqsFocus {
		case focusQueues:
			return selectionSQSQueues, m.selectedQueueIndex, len(m.queues) > 0
		case focusQueueSubs:
			return selectionSQSSubs, m.selectedQueueSubIndex, len(m.queueSubscriptions) > 0
		}
	case panelSNS:
		switch m.snsSubFocus {
		case focusTopics:
			return selectionSNSTopics, m.selectedTopicIndex, len(m.topics) > 0
		case focusSubs:
			return selectionSNSSubs, m.selectedSubIndex, len(m.subscriptions) > 0
		}
	case panelSecrets:
		if m.secretsFocus == focusSecretVersions {
			return selectionSecretVersions, m.selectedSecretVersionIndex, len(m.secretVersions) > 0
		}
		return selectionSecrets, m.selectedSecretIndex, len(m.secrets) > 0
	case panelConfig:
		return selectionSettings, m.focusedInput, len(m.settingsInputs) > 0
	}

	return selectionNone, 0, false
}

func (m *Model) clearSelection() {
	m.selectionActive = false
	m.selectionContext = selectionNone
	m.selectionStart = 0
	m.selectionEnd = 0
}

func (m *Model) toggleSelection() tea.Cmd {
	ctx, index, ok := m.logSelectionContext()
	if !ok {
		return nil
	}

	if !m.selectionActive || m.selectionContext != ctx {
		m.selectionActive = true
		m.selectionContext = ctx
		m.selectionStart = index
		m.selectionEnd = index
		return m.setStatusMessage("Selection started")
	}

	if m.selectionStart == index && m.selectionEnd == index {
		m.clearSelection()
		return m.setStatusMessage("Selection cleared")
	}

	m.selectionEnd = index
	return m.setStatusMessage("Selection updated")
}

func (m Model) selectionBounds() (selectionContext, int, int, bool) {
	if !m.selectionActive {
		return selectionNone, 0, 0, false
	}
	start := m.selectionStart
	end := m.selectionEnd
	if start > end {
		start, end = end, start
	}
	return m.selectionContext, start, end, true
}

func (m Model) isIndexSelected(ctx selectionContext, index int) bool {
	selectedCtx, start, end, ok := m.selectionBounds()
	if !ok || selectedCtx != ctx {
		return false
	}
	return index >= start && index <= end
}

func (m Model) selectedText() string {
	ctx, start, end, ok := m.selectionBounds()
	if ok {
		switch ctx {
		case selectionS3Buckets:
			return m.s3BucketSelectionText(start, end)
		case selectionS3Objects:
			return m.s3ObjectSelectionText(start, end)
		case selectionSQSQueues:
			return m.sqsQueueSelectionText(start, end)
		case selectionSQSSubs:
			return m.sqsSubscriptionSelectionText(start, end)
		case selectionSNSTopics:
			return m.snsTopicSelectionText(start, end)
		case selectionSNSSubs:
			return m.snsSubscriptionSelectionText(start, end)
		case selectionSettings:
			return m.settingsSelectionText(start, end)
		case selectionCommandLogs:
			return m.commandLogSelectionText(start, end)
		}
	}

	switch {
	case m.showLogsModal:
		return m.commandLogSelectionText(m.commandLogCursor, m.commandLogCursor)
	case m.activeTab == panelS3 && m.s3Focus == focusBuckets:
		return m.s3BucketSelectionText(m.selectedBucketIndex, m.selectedBucketIndex)
	case m.activeTab == panelS3 && m.s3Focus == focusObjects:
		return m.s3ObjectSelectionText(m.selectedObjectIndex, m.selectedObjectIndex)
	case m.activeTab == panelSQS && m.sqsFocus == focusQueues:
		return m.sqsQueueSelectionText(m.selectedQueueIndex, m.selectedQueueIndex)
	case m.activeTab == panelSQS && m.sqsFocus == focusQueueSubs:
		return m.sqsSubscriptionSelectionText(m.selectedQueueSubIndex, m.selectedQueueSubIndex)
	case m.activeTab == panelSNS && m.snsSubFocus == focusTopics:
		return m.snsTopicSelectionText(m.selectedTopicIndex, m.selectedTopicIndex)
	case m.activeTab == panelSNS && m.snsSubFocus == focusSubs:
		return m.snsSubscriptionSelectionText(m.selectedSubIndex, m.selectedSubIndex)
	case m.activeTab == panelSecrets:
		if m.secretsFocus == focusSecretVersions {
			return m.secretVersionSelectionText(m.selectedSecretVersionIndex, m.selectedSecretVersionIndex)
		}
		return m.secretSelectionText(m.selectedSecretIndex, m.selectedSecretIndex)
	case m.activeTab == panelConfig:
		return m.settingsSelectionText(m.focusedInput, m.focusedInput)
	default:
		return ""
	}
}

func (m *Model) copySelectedTextCmd() tea.Cmd {
	text := strings.TrimSpace(m.selectedText())
	if text == "" {
		return m.setStatusMessage("Nothing to copy")
	}

	return func() tea.Msg {
		if err := clipboard.WriteAll(text); err != nil {
			return errMsg{Error: fmt.Errorf("clipboard unavailable: %w", err)}
		}
		return statusMsg{Message: "Copied to clipboard"}
	}
}

func (m *Model) copySecretValueCmd() tea.Cmd {
	text := strings.TrimSpace(m.secretValueCopyText())
	if text == "" {
		return m.setStatusMessage("Nothing to copy")
	}

	return func() tea.Msg {
		if err := clipboard.WriteAll(text); err != nil {
			return errMsg{Error: fmt.Errorf("clipboard unavailable: %w", err)}
		}
		return statusMsg{Message: "Secret value copied to clipboard"}
	}
}

func (m Model) s3BucketSelectionText(start, end int) string {
	if len(m.buckets) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.buckets) {
		end = len(m.buckets) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		lines = append(lines, m.buckets[i].Name)
	}
	return strings.Join(lines, "\n")
}

func (m Model) s3ObjectSelectionText(start, end int) string {
	if len(m.objects) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.objects) {
		end = len(m.objects) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		o := m.objects[i]
		lines = append(lines, fmt.Sprintf("%s | %s | %s", o.Key, humanBytes(o.Size), o.LastModified))
	}
	return strings.Join(lines, "\n")
}

func (m Model) sqsQueueSelectionText(start, end int) string {
	if len(m.queues) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.queues) {
		end = len(m.queues) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		q := m.queues[i]
		lines = append(lines, fmt.Sprintf("%s | %s | ready:%d in-flight:%d delayed:%d", q.Name, q.URL, q.MessagesAvailable, q.MessagesNotVisible, q.MessagesDelayed))
	}
	return strings.Join(lines, "\n")
}

func (m Model) sqsSubscriptionSelectionText(start, end int) string {
	if len(m.queueSubscriptions) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.queueSubscriptions) {
		end = len(m.queueSubscriptions) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		sub := m.queueSubscriptions[i]
		lines = append(lines, fmt.Sprintf("%s | %s | %s", sub.TopicARN, sub.Protocol, sub.Endpoint))
	}
	return strings.Join(lines, "\n")
}

func (m Model) snsTopicSelectionText(start, end int) string {
	if len(m.topics) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.topics) {
		end = len(m.topics) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		t := m.topics[i]
		lines = append(lines, fmt.Sprintf("%s | %s", t.Name, t.ARN))
	}
	return strings.Join(lines, "\n")
}

func (m Model) snsSubscriptionSelectionText(start, end int) string {
	if len(m.subscriptions) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.subscriptions) {
		end = len(m.subscriptions) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		sub := m.subscriptions[i]
		lines = append(lines, fmt.Sprintf("%s | %s | %s", sub.TopicARN, sub.Protocol, sub.Endpoint))
	}
	return strings.Join(lines, "\n")
}

func (m Model) settingsSelectionText(start, end int) string {
	if len(m.settingsInputs) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.settingsInputs) {
		end = len(m.settingsInputs) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		lines = append(lines, m.settingsInputs[i].Value())
	}
	return strings.Join(lines, "\n")
}

func (m Model) secretSelectionText(start, end int) string {
	if len(m.secrets) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.secrets) {
		end = len(m.secrets) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		lines = append(lines, fmt.Sprintf("%s | %s", m.secrets[i].Name, m.secrets[i].ARN))
	}
	return strings.Join(lines, "\n")
}

func (m Model) secretVersionSelectionText(start, end int) string {
	if len(m.secretVersions) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.secretVersions) {
		end = len(m.secretVersions) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		version := m.secretVersions[i]
		lines = append(lines, fmt.Sprintf("%s | %s | %s", secretVersionVisualLabel(i, version), version.VersionID, version.CreatedDate))
	}
	return strings.Join(lines, "\n")
}

func (m Model) currentResourceARN() string {
	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusObjects && len(m.objects) > 0 && m.selectedObjectIndex < len(m.objects) {
			return m.objects[m.selectedObjectIndex].Key
		}
		if m.s3Focus == focusBuckets && len(m.buckets) > 0 && m.selectedBucketIndex < len(m.buckets) {
			return m.buckets[m.selectedBucketIndex].Name
		}
	case panelSQS:
		if m.sqsFocus == focusQueueSubs && len(m.queueSubscriptions) > 0 && m.selectedQueueSubIndex < len(m.queueSubscriptions) {
			return m.queueSubscriptions[m.selectedQueueSubIndex].ARN
		}
		if len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
			return m.queues[m.selectedQueueIndex].ARN
		}
	case panelSNS:
		if m.snsSubFocus == focusSubs && len(m.subscriptions) > 0 && m.selectedSubIndex < len(m.subscriptions) {
			return m.subscriptions[m.selectedSubIndex].ARN
		}
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m.topics[m.selectedTopicIndex].ARN
		}
	case panelSecrets:
		if m.secretsFocus == focusSecretVersions && len(m.secretVersions) > 0 && m.selectedSecretVersionIndex < len(m.secretVersions) {
			return m.secretVersions[m.selectedSecretVersionIndex].VersionID
		}
		if len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
			return m.secrets[m.selectedSecretIndex].ARN
		}
	}
	return ""
}

func (m *Model) copyResourceARNCmd() tea.Cmd {
	arn := strings.TrimSpace(m.currentResourceARN())
	if arn == "" {
		return m.setStatusMessage("Nothing to copy")
	}
	return func() tea.Msg {
		if err := clipboard.WriteAll(arn); err != nil {
			return errMsg{Error: fmt.Errorf("clipboard unavailable: %w", err)}
		}
		return statusMsg{Message: "ARN copied to clipboard"}
	}
}

func (m Model) commandLogSelectionText(start, end int) string {
	if len(m.commandLogs) == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end >= len(m.commandLogs) {
		end = len(m.commandLogs) - 1
	}
	if start > end {
		start, end = end, start
	}
	var lines []string
	for i := start; i <= end; i++ {
		entry := m.commandLogs[i]
		status := "success"
		if entry.Error != nil {
			status = "error=" + entry.Error.Error()
		}
		lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s", entry.Time.Format(time.RFC3339), entry.Action, entry.Target, status, strings.TrimSpace(entry.Output)))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) copyS3PathCmd() tea.Cmd {
	var path string
	bucket := m.selectedS3BucketName()
	if bucket == "" {
		return m.setStatusMessage("Nothing to copy")
	}

	if m.s3Focus == focusObjects && len(m.objects) > 0 && m.selectedObjectIndex < len(m.objects) {
		path = fmt.Sprintf("s3://%s/%s", bucket, m.objects[m.selectedObjectIndex].Key)
	} else {
		path = fmt.Sprintf("s3://%s", bucket)
	}

	return func() tea.Msg {
		err := clipboard.WriteAll(path)
		if err != nil {
			return statusMsg{Message: "S3 path: " + path + " (clipboard unavailable)"}
		}
		return statusMsg{Message: "S3 path copied: " + path}
	}
}
