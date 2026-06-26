package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
	"monostack/internal/pkg/ui"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing Monostack..."
	}

	if m.showSplash {
		return m.renderSplash()
	}

	header := m.renderHeader()
	tabBar := m.renderTabBar()
	body := m.renderBody()
	footer := m.renderFooter()

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		body,
		footer,
	)

	if toastView := m.renderToasts(); toastView != "" {
		startRow := lipgloss.Height(header) + lipgloss.Height(tabBar)
		toastWidth := lipgloss.Width(toastView)
		startCol := m.width - toastWidth - 2
		if startCol < 0 {
			startCol = 0
		}
		view = overlayString(view, toastView, startRow, startCol)
	}

	if m.showProgress && m.width > 0 {
		progressView := m.progress.view()
		view += "\n" + lipgloss.NewStyle().
			Width(m.width).
			Padding(0, 1).
			Render(progressView)
	}

	var modalView string
	switch {
	case m.showS3ConfirmDelete:
		modalView = m.styles.Modal.Render(m.renderS3ConfirmDeleteModal())
	case m.showSqsConfirmDelete:
		modalView = m.styles.Modal.Render(m.renderSqsConfirmDeleteModal())
	case m.showSqsPurgeAllConfirm:
		modalView = m.styles.Modal.Render(m.renderSqsPurgeAllConfirmModal())
	case m.showSqsSubDeleteConfirm:
		modalView = m.styles.Modal.Render(m.renderSqsSubDeleteConfirmModal())
	case m.showSnsPublishModal:
		modalView = m.styles.Modal.Render(m.renderSnsPublishModal())
	case m.showS3CreateModal:
		modalView = m.styles.Modal.Render(m.renderS3CreateBucketModal())
	case m.showS3CreateFolderModal:
		modalView = m.styles.Modal.Render(m.renderS3CreateFolderModal())
	case m.showS3UploadModal:
		modalView = m.styles.Modal.Render(m.renderS3UploadModal())
	case m.showS3PreviewModal:
		modalView = m.styles.Modal.Render(m.renderS3PreviewModal())
	case m.showVersionsModal:
		modalView = m.styles.Modal.Render(m.renderS3VersionsModal())
	case m.showSqsCreateModal:
		modalView = m.styles.Modal.Render(m.renderSqsCreateQueueModal())
	case m.showSnsCreateTopicModal:
		modalView = m.styles.Modal.Render(m.renderSnsCreateTopicModal())
	case m.showSnsSimpleSubModal:
		modalView = m.styles.Modal.Render(m.renderSnsSimpleSubModal())
	case m.showSnsBatchSubModal:
		modalView = m.styles.Modal.Render(m.renderSnsBatchSubModal())
	case m.showSnsYamlApplyConfirm:
		modalView = m.styles.Modal.Render(m.renderSnsYamlApplyConfirmModal())
	case m.showSnsYamlImportModal:
		modalView = m.styles.Modal.Render(m.renderSnsYamlImportModal())
	case m.showSqsBatchSubModal:
		modalView = m.styles.Modal.Render(m.renderSqsBatchSubModal())
	case m.showSnsSubEditModal:
		modalView = m.styles.Modal.Render(m.renderSnsEditSubModal())
	case m.showSnsConfirmDelete:
		modalView = m.styles.Modal.Render(m.renderSnsConfirmDeleteModal())
	case m.showSecretDeleteConfirm:
		modalView = m.styles.Modal.Render(m.renderSecretDeleteConfirmModal())
	case m.showSecretRestoreConfirm:
		modalView = m.styles.Modal.Render(m.renderSecretRestoreConfirmModal())
	case m.showSecretPromoteConfirm:
		modalView = m.styles.Modal.Render(m.renderSecretPromoteConfirmModal())
	case m.showSecretClipboardConfirm:
		modalView = m.styles.Modal.Render(m.renderSecretClipboardConfirmModal())
	case m.showSecretCreateModal:
		modalView = m.styles.Modal.Render(m.renderSecretCreateModal())
	case m.showSecretUpdateModal:
		modalView = m.styles.Modal.Render(m.renderSecretUpdateModal())
	case m.showSecretValueModal:
		modalView = m.styles.Modal.Render(m.renderSecretValueModal())
	case m.showExportModal:
		modalView = m.styles.Modal.Render(m.renderExportProfileModal())
	case m.showImportModal:
		modalView = m.styles.Modal.Render(m.renderImportProfileModal())
	case m.showSingleExportModal:
		modalView = m.styles.Modal.Render(m.renderSingleExportModal())
	case m.showPeekModal:
		modalView = m.styles.Modal.Render(m.renderPeekModal())
	case m.showHelpModal:
		modalView = m.renderHelpModal()
	case m.showLogsModal:
		modalView = m.styles.Modal.Render(m.renderLogsModal())
	case m.showInspectionModal:
		modalView = m.styles.Modal.Render(m.renderInspectionModal())
	case m.showProfileModal:
		modalView = m.styles.Modal.Render(m.renderProfileModal())
	case m.showProfileDeleteConfirm:
		modalView = m.styles.Modal.Render(m.renderProfileDeleteConfirmModal())
	case m.showMultiDeleteConfirm:
		modalView = m.styles.Modal.Render(m.renderMultiDeleteConfirmModal())
	case m.showCommandPalette:
		modalView = m.renderCommandPalette()
	}

	if modalView != "" {
		modalWidth := lipgloss.Width(modalView)
		modalHeight := lipgloss.Height(modalView)
		startRow := (m.height - modalHeight) / 2
		startCol := (m.width - modalWidth) / 2
		if startRow < 0 {
			startRow = 0
		}
		if startCol < 0 {
			startCol = 0
		}
		view = overlayString(view, modalView, startRow, startCol)
	}

	return lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		Render(view)
}

func (m Model) renderBody() string {
	mainPanel := ""
	if m.loading {
		mainPanel = lipgloss.NewStyle().
			Height(m.mainPanelHeight()).
			Width(m.width-2).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.styles.InfoText.Render("* Fetching data from AWS services..."))
		return mainPanel
	}

	switch m.activeTab {
	case panelS3:
		mainPanel = m.renderS3Panel()
	case panelSQS:
		mainPanel = m.renderSQSPanel()
	case panelSNS:
		mainPanel = m.renderSNSPanel()
	case panelSecrets:
		mainPanel = m.renderSecretsPanel()
	case panelConfig:
		mainPanel = m.renderConfigPanel()
	}

	return "\n\n" + mainPanel
}

func (m Model) renderHeader() string {
	header := m.renderHeaderBrand()
	profileInfo := "Offline"
	if m.config != nil {
		mockTag := ""
		if m.config.UseMock {
			mockTag = " (Mock)"
		}
		endpoint := m.config.EndpointURL
		if endpoint == "" {
			endpoint = "Real AWS"
		} else {

			endpoint = strings.TrimPrefix(endpoint, "http://")
			endpoint = strings.TrimPrefix(endpoint, "https://")
		}

		prefix := ""
		if m.config.ActiveProfile != "" {
			prefix = "[" + m.config.ActiveProfile + "] "
		}

		healthIcons := m.renderHealthDots()
		profileInfo = fmt.Sprintf("%s%s %s%s %s", prefix, m.config.ServiceName, m.config.Region, mockTag, healthIcons)
	}

	statsStr := m.styles.InfoText.Render(profileInfo)

	spacerLen := m.width - lipgloss.Width(header) - lipgloss.Width(statsStr)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	headerLine := lipgloss.JoinHorizontal(lipgloss.Top,
		header,
		spacer,
		statsStr,
	)

	return headerLine
}

func (m Model) renderHealthDots() string {
	if m.config == nil || m.config.UseMock {
		return ""
	}
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	dot := func(up bool) string {
		if up {
			return okStyle.Render("●")
		}
		return errStyle.Render("●")
	}
	services := m.config.EnabledServices
	if len(services) == 0 {
		services = domain.DefaultEnabledServices()
	}
	var parts []string
	for _, svc := range services {
		switch svc {
		case domain.ServiceS3:
			parts = append(parts, "S3:"+dot(m.serviceHealth.S3))
		case domain.ServiceSQS:
			parts = append(parts, "SQS:"+dot(m.serviceHealth.SQS))
		case domain.ServiceSNS:
			parts = append(parts, "SNS:"+dot(m.serviceHealth.SNS))
		case domain.ServiceSecrets:
			parts = append(parts, "SEC:"+dot(m.serviceHealth.Secrets))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func (m Model) renderHeaderBrand() string {
	return m.styles.Header.Render(" " + m.renderBrandWordmark(true) + " ")
}

func (m Model) renderTabBar() string {
	visible := m.visiblePanels()
	renderedTabs := make([]string, 0, len(visible))
	for i, panel := range visible {
		tab := m.tabLabel(panel, i+1)
		if m.activeTab == panel {
			renderedTabs = append(renderedTabs, m.activeTabStyle(panel).Render(tab))
		} else {
			renderedTabs = append(renderedTabs, m.styles.InactiveTab.Render(tab))
		}
	}

	return m.renderPillRows(m.width-4, renderedTabs...)
}

func (m Model) renderFooter() string {
	sep := m.styles.InfoText.Render(" • ")
	var primary []string

	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusBuckets {
			primary = []string{
				m.footerItem("jk", "nav buckets"),
				m.footerItem("enter", "open"),
				m.footerItem("c", "create"),
				m.footerItem("f", "folder"),
				m.footerItem("u", "upload"),
			}
		} else {
			primary = []string{
				m.footerItem("jk", "nav files"),
				m.footerItem("esc", "back"),
				m.footerItem("f", "folder"),
				m.footerItem("v", "preview"),
				m.footerItem("V", "versions"),
				m.footerItem("w", "download"),
				m.footerItem("d", "delete"),
			}
		}
	case panelSQS:
		if m.sqsFocus == focusQueues {
			primary = []string{
				m.footerItem("jk", "nav queues"),
				m.footerItem("enter", "inspect"),
				m.footerItem("l", "routes"),
				m.footerItem("c", "create"),
				m.footerItem("s", "send"),
				m.footerItem("m", "purge"),
				m.footerItem("M", "purge all"),
			}
		} else {
			primary = []string{
				m.footerItem("jk", "nav routes"),
				m.footerItem("esc", "back"),
				m.footerItem("b", "add route"),
				m.footerItem("d", "unlink"),
			}
		}
	case panelSNS:
		if m.snsSubFocus == focusTopics {
			primary = []string{
				m.footerItem("jk", "nav topics"),
				m.footerItem("enter", "inspect"),
				m.footerItem("l", "routes"),
				m.footerItem("c", "create"),
				m.footerItem("s", "publish"),
				m.footerItem("d", "delete"),
			}
		} else {
			primary = []string{
				m.footerItem("jk", "nav routes"),
				m.footerItem("enter", "inspect"),
				m.footerItem("esc", "back"),
				m.footerItem("c", "link"),
				m.footerItem("b", "batch"),
				m.footerItem("d", "delete"),
			}
		}
	case panelSecrets:
		if m.secretsFocus == focusSecrets {
			primary = []string{
				m.footerItem("jk", "nav secrets"),
				m.footerItem("l", "versions"),
				m.footerItem("enter", "inspect"),
				m.footerItem("c", "create"),
				m.footerItem("u", "update"),
				m.footerItem("v", "value"),
				m.footerItem("R", "recover"),
				m.footerItem("d", "delete"),
			}
		} else {
			primary = []string{
				m.footerItem("jk", "nav versions"),
				m.footerItem("h", "back"),
				m.footerItem("r", "make current"),
				m.footerItem("esc", "back"),
			}
		}
	case panelConfig:
		if m.settingsEditMode {
			primary = []string{
				m.footerItem("tab", "next field"),
				m.footerItem("enter", "done"),
				m.footerItem("esc", "cancel"),
			}
		} else {
			primary = []string{
				m.footerItem("jk", "nav"),
				m.footerItem("enter", "edit"),
				m.footerItem("s", "save"),
				m.footerItem("E", "export"),
				m.footerItem("L", "load"),
			}
		}
	}

	if m.multiSelectActive {
		count := m.multiSelectCountForPanel()
		primary = []string{
			m.footerItem("space", "toggle item"),
			m.footerItem("d", fmt.Sprintf("delete %d selected", count)),
			m.footerItem("esc", "clear"),
		}
	}

	primary = append(primary, m.footerItem("o", "logs"))
	primary = append(primary, m.footerItem("ctrl+p", "commands"))
	primary = append(primary, m.footerItem("T", "theme"))
	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorSubtle)).
		Render(fmt.Sprintf("MonoStack %s", Version))

	contentWidth := m.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	help := m.footerItem("?", "help")
	rendered := strings.Join(primary, sep)
	maxLeftWidth := contentWidth - lipgloss.Width(version) - 1
	if maxLeftWidth < lipgloss.Width(help) {
		maxLeftWidth = lipgloss.Width(help)
	}
	for len(primary) > 0 && lipgloss.Width(rendered)+lipgloss.Width(sep)+lipgloss.Width(help) > maxLeftWidth {
		primary = primary[:len(primary)-1]
		rendered = strings.Join(primary, sep)
	}

	left := help
	if rendered != "" {
		left = rendered + sep + help
	}

	spacerLen := contentWidth - lipgloss.Width(left) - lipgloss.Width(version)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + left + spacer + version
	if footerWidth := lipgloss.Width(footerText); footerWidth < contentWidth+1 {
		footerText += strings.Repeat(" ", contentWidth+1-footerWidth)
	}

	return m.styles.Footer.Padding(0, 0).Render(footerText)
}

func (m Model) footerItem(k, action string) string {
	footerKey := m.styles.FooterKey.Render(k)
	footerAction := m.styles.FooterAction.Render(action)
	return footerKey + " " + footerAction
}

func (m Model) mainPanelHeight() int {
	height := m.height - lipgloss.Height(m.renderHeader()) - lipgloss.Height(m.renderTabBar()) - lipgloss.Height(m.renderFooter()) - 2
	if height < 5 {
		height = 5
	}
	return height
}

func (m Model) activeTabStyle(panel activePanel) lipgloss.Style {
	tc := m.themeColors()
	var color string
	switch panel {
	case panelS3:
		color = tc.Amber
	case panelSQS:
		color = tc.Indigo
	case panelSNS:
		color = tc.Rose
	case panelSecrets:
		color = tc.Emerald
	case panelConfig:
		color = tc.Accent
	default:
		color = tc.Stack
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color(tc.Bg)).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)
}
