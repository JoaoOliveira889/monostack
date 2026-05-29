package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"monostack/internal/domain"
	"monostack/internal/usecase"
)

const (
	panelRatioDefault = 0.5
	panelRatioMin     = 0.1
	panelRatioMax     = 0.9
)

func serviceKeyForPanel(panel activePanel) string {
	switch panel {
	case panelS3:
		return domain.ServiceS3
	case panelSQS:
		return domain.ServiceSQS
	case panelSNS:
		return domain.ServiceSNS
	case panelSecrets:
		return domain.ServiceSecrets
	default:
		return ""
	}
}

func normalizePanelRatio(value float64) float64 {
	if value < panelRatioMin || value > panelRatioMax {
		return panelRatioDefault
	}
	return value
}

func clonePanelRatios(values map[string]float64) map[string]float64 {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]float64, len(values))
	for key, value := range values {
		cloned[key] = normalizePanelRatio(value)
	}
	return cloned
}

func (m Model) panelRatioFor(panel activePanel) float64 {
	if panel == panelConfig {
		return panelRatioDefault
	}
	if m.config == nil {
		return panelRatioDefault
	}

	key := serviceKeyForPanel(panel)
	if key == "" {
		return panelRatioDefault
	}

	if m.config.PanelRatios != nil {
		if value, ok := m.config.PanelRatios[key]; ok {
			return normalizePanelRatio(value)
		}
	}

	return panelRatioDefault
}

func (m *Model) syncActivePanelRatio() {
	if m.activeTab == panelConfig {
		return
	}
	m.leftPanelRatio = m.panelRatioFor(m.activeTab)
}

func (m *Model) setActivePanelRatio(ratio float64) {
	ratio = normalizePanelRatio(ratio)
	m.leftPanelRatio = ratio
	if m.config == nil {
		return
	}

	key := serviceKeyForPanel(m.activeTab)
	if key == "" {
		return
	}
	if m.config.PanelRatios == nil {
		m.config.PanelRatios = make(map[string]float64)
	}
	m.config.PanelRatios[key] = ratio
	m.config.LeftPanelRatio = ratio
}

func (m Model) panelEnabled(panel activePanel) bool {
	if panel == panelConfig {
		return true
	}
	if m.config == nil {
		return true
	}
	services := domain.NormalizeEnabledServices(m.config.EnabledServices)
	switch panel {
	case panelS3:
		return domain.ServiceEnabled(services, domain.ServiceS3)
	case panelSQS:
		return domain.ServiceEnabled(services, domain.ServiceSQS)
	case panelSNS:
		return domain.ServiceEnabled(services, domain.ServiceSNS)
	case panelSecrets:
		return domain.ServiceEnabled(services, domain.ServiceSecrets)
	default:
		return false
	}
}

func (m Model) visibleServicePanels() []activePanel {
	panels := make([]activePanel, 0, 4)
	if m.panelEnabled(panelS3) {
		panels = append(panels, panelS3)
	}
	if m.panelEnabled(panelSQS) {
		panels = append(panels, panelSQS)
	}
	if m.panelEnabled(panelSNS) {
		panels = append(panels, panelSNS)
	}
	if m.panelEnabled(panelSecrets) {
		panels = append(panels, panelSecrets)
	}
	return panels
}

func (m Model) visiblePanels() []activePanel {
	panels := append([]activePanel(nil), m.visibleServicePanels()...)
	panels = append(panels, panelConfig)
	return panels
}

func (m Model) firstVisibleServicePanel() activePanel {
	for _, panel := range []activePanel{panelS3, panelSQS, panelSNS, panelSecrets} {
		if m.panelEnabled(panel) {
			return panel
		}
	}
	return panelConfig
}

func (m *Model) ensureActiveTabVisible() {
	if m.activeTab == panelConfig || m.panelEnabled(m.activeTab) {
		m.syncActivePanelRatio()
		return
	}
	m.activeTab = m.firstVisibleServicePanel()
	m.syncActivePanelRatio()
}

func (m Model) tabLabel(panel activePanel, index int) string {
	switch panel {
	case panelS3:
		return fmt.Sprintf("[%d] S3 Explorer", index)
	case panelSQS:
		return fmt.Sprintf("[%d] SQS Queues", index)
	case panelSNS:
		return fmt.Sprintf("[%d] SNS Topics", index)
	case panelSecrets:
		return fmt.Sprintf("[%d] Secrets", index)
	case panelConfig:
		return "[5] Settings"
	default:
		return ""
	}
}

func (m Model) panelForTabIndex(index int) (activePanel, bool) {
	if index == 5 {
		return panelConfig, true
	}
	servicePanels := m.visibleServicePanels()
	if index < 1 || index > len(servicePanels) {
		return 0, false
	}
	return servicePanels[index-1], true
}

func (m *Model) activatePanel(panel activePanel) tea.Cmd {
	if panel != panelConfig && !m.panelEnabled(panel) {
		return nil
	}
	m.activeTab = panel
	m.syncActivePanelRatio()
	m.errorMsg = ""
	m.statusMsg = ""
	m.clearSelection()

	switch panel {
	case panelS3:
		m.s3Focus = focusBuckets
		if len(m.buckets) > 0 {
			m.loading = false
			return nil
		}
		m.loading = true
		return m.loadS3BucketsCmd()
	case panelSQS:
		m.sqsFocus = focusQueues
		m.loading = true
		return tea.Batch(m.loadSQSQueuesCmd(), m.loadSNSTopicsCmd())
	case panelSNS:
		m.loading = true
		return m.loadSNSTopicsCmd()
	case panelSecrets:
		m.secretsFocus = focusSecrets
		m.loading = true
		return m.loadSecretsCmd()
	default:
		return nil
	}
}

func (m *Model) syncSettingsInputsFromConfig() {
	if len(m.settingsInputs) < 8 || m.config == nil {
		return
	}

	m.settingsInputs[0].Placeholder = "Service Name (e.g. MiniStack)"
	m.settingsInputs[0].SetValue(m.config.ServiceName)

	m.settingsInputs[1].Placeholder = "Endpoint URL (e.g. http://localhost:4566)"
	m.settingsInputs[1].SetValue(m.config.EndpointURL)

	m.settingsInputs[2].Placeholder = "Region (e.g. us-east-1)"
	m.settingsInputs[2].SetValue(m.config.Region)

	m.settingsInputs[3].Placeholder = "Access Key ID"
	m.settingsInputs[3].SetValue(m.config.AccessKeyID)

	m.settingsInputs[4].Placeholder = "Secret Access Key"
	m.settingsInputs[4].SetValue(m.config.SecretAccessKey)
	m.settingsInputs[4].EchoMode = textinput.EchoPassword
	m.settingsInputs[4].EchoCharacter = '•'

	m.settingsInputs[5].Placeholder = "Mock Mode (true/false)"
	if m.config.UseMock {
		m.settingsInputs[5].SetValue("true")
	} else {
		m.settingsInputs[5].SetValue("false")
	}

	m.settingsInputs[6].Placeholder = "Snapshot directory or file path"
	if m.config.SnapshotPath != "" {
		m.settingsInputs[6].SetValue(m.config.SnapshotPath)
	} else {
		m.settingsInputs[6].SetValue(usecase.DefaultSnapshotPath())
	}

	m.settingsInputs[7].Placeholder = "Enabled services (s3,sqs,sns,secrets)"
	services := domain.NormalizeEnabledServices(m.config.EnabledServices)
	m.settingsInputs[7].SetValue(strings.Join(services, ","))

	for i := range m.settingsInputs {
		m.settingsInputs[i].Blur()
	}
}

func (m Model) configFromSettingsInputs() *domain.AWSConfig {
	mockVal := strings.ToLower(strings.TrimSpace(m.settingsInputs[5].Value()))
	useMock := mockVal == "true" || mockVal == "1" || mockVal == "yes"

	cfg := &domain.AWSConfig{
		ServiceName:     m.settingsInputs[0].Value(),
		EndpointURL:     m.settingsInputs[1].Value(),
		Region:          m.settingsInputs[2].Value(),
		AccessKeyID:     m.settingsInputs[3].Value(),
		SecretAccessKey: m.settingsInputs[4].Value(),
		UseMock:         useMock,
		LeftPanelRatio:  m.leftPanelRatio,
		SnapshotPath:    strings.TrimSpace(m.settingsInputs[6].Value()),
		EnabledServices: domain.NormalizeEnabledServices(splitCSVList(m.settingsInputs[7].Value())),
	}

	if m.config != nil {
		cfg.PanelRatios = clonePanelRatios(m.config.PanelRatios)
	}

	return cfg
}
