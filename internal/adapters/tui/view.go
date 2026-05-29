package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

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

	if m.showS3ConfirmDelete {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderS3ConfirmDeleteModal()),
		)
	}

	if m.showSqsConfirmDelete {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSqsConfirmDeleteModal()),
		)
	}

	if m.showSqsPurgeAllConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSqsPurgeAllConfirmModal()),
		)
	}

	if m.showSqsSubDeleteConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSqsSubDeleteConfirmModal()),
		)
	}

	if m.showSnsPublishModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsPublishModal()),
		)
	}

	if m.showS3CreateModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderS3CreateBucketModal()),
		)
	}

	if m.showS3CreateFolderModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderS3CreateFolderModal()),
		)
	}

	if m.showS3UploadModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderS3UploadModal()),
		)
	}

	if m.showS3PreviewModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderS3PreviewModal()),
		)
	}

	if m.showSqsCreateModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSqsCreateQueueModal()),
		)
	}

	if m.showSnsCreateTopicModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsCreateTopicModal()),
		)
	}

	if m.showSnsSimpleSubModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsSimpleSubModal()),
		)
	}

	if m.showSnsBatchSubModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsBatchSubModal()),
		)
	}

	if m.showSnsYamlApplyConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsYamlApplyConfirmModal()),
		)
	}

	if m.showSnsYamlImportModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsYamlImportModal()),
		)
	}

	if m.showSqsBatchSubModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSqsBatchSubModal()),
		)
	}

	if m.showSnsSubEditModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsEditSubModal()),
		)
	}

	if m.showSnsConfirmDelete {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSnsConfirmDeleteModal()),
		)
	}

	if m.showSecretDeleteConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretDeleteConfirmModal()),
		)
	}

	if m.showSecretRestoreConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretRestoreConfirmModal()),
		)
	}

	if m.showSecretPromoteConfirm {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretPromoteConfirmModal()),
		)
	}

	if m.showSecretCreateModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretCreateModal()),
		)
	}

	if m.showSecretUpdateModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretUpdateModal()),
		)
	}

	if m.showSecretValueModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderSecretValueModal()),
		)
	}

	if m.showExportModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderExportProfileModal()),
		)
	}

	if m.showImportModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderImportProfileModal()),
		)
	}

	if m.showPeekModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderPeekModal()),
		)
	}

	if m.showHelpModal {
		return m.renderHelpModal()
	}

	if m.showLogsModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderLogsModal()),
		)
	}

	if m.showInspectionModal {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.Modal.Render(m.renderInspectionModal()),
		)
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
		profileInfo = fmt.Sprintf("[%s] %s%s ", m.config.ServiceName, endpoint, mockTag)
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

	if m.statusMsg != "" {
		status := lipgloss.NewStyle().
			Foreground(m.styles.InfoText.GetForeground()).
			Width(m.width).
			Render("  SUCCESS: " + m.statusMsg)
		return headerLine + "\n" + status
	}
	if m.errorMsg != "" {
		status := lipgloss.NewStyle().
			Foreground(m.styles.ErrorBadge.GetBackground()).
			Width(m.width).
			Render("  ERROR: " + m.errorMsg)
		return headerLine + "\n" + status
	}

	return headerLine
}

func (m Model) renderHeaderBrand() string {
	return m.styles.Header.Render(" " + renderBrandWordmark(true) + " ")
}

func (m Model) renderTabBar() string {
	visible := m.visiblePanels()
	renderedTabs := make([]string, 0, len(visible))
	for i, panel := range visible {
		tab := m.tabLabel(panel, i+1)
		if m.activeTab == panel {
			renderedTabs = append(renderedTabs, m.styles.ActiveTab.Render(tab))
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
				m.footerItem("p", "purge"),
				m.footerItem("P", "purge all"),
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

	primary = append(primary, m.footerItem("o", "logs"))
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
	footerKey := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorAccent)).
		Bold(true).
		Render(k)
	footerAction := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorFg)).
		Render(action)
	return footerKey + " " + footerAction
}

func (m Model) mainPanelHeight() int {
	height := m.height - lipgloss.Height(m.renderHeader()) - lipgloss.Height(m.renderTabBar()) - lipgloss.Height(m.renderFooter()) - 2
	if height < 5 {
		height = 5
	}
	return height
}
