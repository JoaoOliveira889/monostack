package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
	"monostack/internal/pkg/ui"
	"monostack/internal/usecase"
)

func mkModel() Model {
	s3 := &mockS3Manager{}
	sqs := &mockSQSManager{}
	sns := &mockSNSManager{}
	cfg := &mockConfigStore{}

	awsUC := usecase.NewAWSUseCase(s3, sqs, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCase(cfg)
	snapshotUC := usecase.NewSnapshotUseCase(awsUC, cfgUC)
	return NewModel(awsUC, cfgUC, snapshotUC)
}

func TestNewModel_InitialState(t *testing.T) {
	m := mkModel()

	if m.activeTab != panelS3 {
		t.Errorf("expected activeTab panelS3, got %v", m.activeTab)
	}
	if m.leftPanelRatio != 0.3 {
		t.Errorf("expected leftPanelRatio 0.3, got %f", m.leftPanelRatio)
	}
	if m.loading != true {
		t.Error("expected loading to be true initially")
	}
	if m.showHelpModal {
		t.Error("expected showHelpModal to be false")
	}
	if m.showPeekModal {
		t.Error("expected showPeekModal to be false")
	}
	if m.showSqsSendModal {
		t.Error("expected showSqsSendModal to be false")
	}
	if m.showSnsPublishModal {
		t.Error("expected showSnsPublishModal to be false")
	}
}

func TestModel_Init(t *testing.T) {
	m := mkModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestModel_VersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("expected Version to be non-empty")
	}
}

func TestModel_StylesNotPanic(t *testing.T) {
	m := mkModel()

	_ = m.styles.Header.Render("test")
	_ = m.styles.Footer.Render("test")
	_ = m.styles.ActiveTab.Render("test")
	_ = m.styles.InactiveTab.Render("test")
	_ = m.styles.BorderPanel.Render("test")
	_ = m.styles.FocusedPanel.Render("test")
	_ = m.styles.Title.Render("test")
	_ = m.styles.Highlight.Render("test")
	_ = m.styles.ListItem.Render("test")
	_ = m.styles.SelectedListItem.Render("test")
	_ = m.styles.InputLabel.Render("test")
	_ = m.styles.InputFocused.Render("test")
	_ = m.styles.InputUnfocused.Render("test")
	_ = m.styles.Modal.Render("test")
	_ = m.styles.SuccessBadge.Render("test")
	_ = m.styles.ErrorBadge.Render("test")
	_ = m.styles.InfoText.Render("test")
}

func TestModel_StylesAddFormatting(t *testing.T) {
	m := mkModel()

	plain := lipgloss.NewStyle().Render("test")
	if m.styles.Header.Render("test") == plain {
		t.Error("expected Header style to add formatting")
	}
	if m.styles.SuccessBadge.Render("test") == plain {
		t.Error("expected SuccessBadge style to add formatting")
	}
	if m.styles.ErrorBadge.Render("test") == plain {
		t.Error("expected ErrorBadge style to add formatting")
	}
	if m.styles.ActiveTab.Render("test") == plain {
		t.Error("expected ActiveTab style to add formatting")
	}
}

func TestRenderFooter_ShowsVersionAndHelpShortcut(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.activeTab = panelS3
	footer := m.renderFooter()
	if !strings.Contains(footer, "MonoStack") {
		t.Fatalf("footer should show version: %q", footer)
	}
	if !strings.Contains(footer, "? help") {
		t.Fatalf("footer should advertise help shortcut: %q", footer)
	}
}

func TestRenderFooter_DoesNotWrapOnNarrowWidth(t *testing.T) {
	m := mkModel()
	m.width = 44
	m.activeTab = panelSNS
	m.snsSubFocus = focusTopics
	footer := m.renderFooter()
	if strings.Contains(footer, "\n") {
		t.Fatalf("footer wrapped unexpectedly: %q", footer)
	}
	if !strings.Contains(footer, "? help") {
		t.Fatalf("footer should preserve help shortcut in narrow width: %q", footer)
	}
	if !strings.Contains(footer, "MonoStack "+Version) {
		t.Fatalf("footer should preserve version in narrow width: %q", footer)
	}
}

func TestRenderHeaderUsesTextHeader(t *testing.T) {
	m := mkModel()
	m.width = 100
	m.config = nil

	header := m.renderHeader()
	if !strings.Contains(header, "MonoStack") {
		t.Fatalf("header should render the brand title: %q", header)
	}
}

func TestMainPanelHeightRespondsToHeaderAndFooter(t *testing.T) {
	m := mkModel()
	m.width = 100
	m.height = 40
	m.statusMsg = "saved"

	height := m.mainPanelHeight()
	if height <= 0 {
		t.Fatalf("expected positive panel height, got %d", height)
	}
	if height >= m.height {
		t.Fatalf("expected panel height to account for header/footer, got %d", height)
	}
}

func TestRenderSecretValuePreview_WrapsAndPreservesContent(t *testing.T) {
	value := `{"placeholder":"replace-me","ConnectionStrings__DatabaseConnection":"Host=aurora-postgres-dev.cluster-cjwkqcmaeru1.sa-east-1.rds.amazonaws.com;Port=5432;Database=db_tenants;Username=;Password=;"}`
	preview := renderSecretValuePreview(value, 40, 12)

	if !strings.Contains(preview, "ConnectionStrings__DatabaseConnection") {
		t.Fatalf("expected preview to include the secret content, got %q", preview)
	}
	if !strings.Contains(preview, "\n") {
		t.Fatalf("expected preview to span multiple lines, got %q", preview)
	}
}

func TestRouteCountsUseAllSubscriptions(t *testing.T) {
	m := mkModel()
	m.allSubscriptions = []domain.SNSSubscription{
		{TopicARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns", Endpoint: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs"},
		{TopicARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns", Endpoint: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs"},
		{TopicARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-notifications-sns", Endpoint: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-notifications-sqs"},
	}

	if got := m.routeCountForTopic("arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns"); got != 2 {
		t.Fatalf("expected 2 topic routes, got %d", got)
	}
	if got := m.routeCountForQueue(domain.SQSQueue{URL: "https://sqs.us-east-1.amazonaws.com/000000000000/dev-webapi-pix-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs"}); got != 2 {
		t.Fatalf("expected 2 queue routes, got %d", got)
	}
}

func TestRenderHelpModalFitsNarrowWidth(t *testing.T) {
	m := mkModel()
	m.width = 72
	m.height = 32

	help := m.renderHelpModal()
	for _, line := range strings.Split(help, "\n") {
		if lipgloss.Width(line) > m.width {
			t.Fatalf("expected help line to fit within width %d, got %d for %q", m.width, lipgloss.Width(line), line)
		}
	}
	if !strings.Contains(help, "MonoStack SHORTCUTS") {
		t.Fatalf("expected help modal to include branded title, got %q", help)
	}
	if !strings.Contains(help, "ESC or ctrl+p to close") {
		t.Fatalf("expected help modal to include close hint, got %q", help)
	}
	topBorderFound := false
	for _, line := range strings.Split(help, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "╔") && strings.HasSuffix(trimmed, "╗") {
			topBorderFound = true
			break
		}
	}
	if !topBorderFound {
		t.Fatalf("expected help modal top border to be closed, got %q", help)
	}
}

func TestRenderTitledPanelActiveDoesNotFillBodyBackground(t *testing.T) {
	m := mkModel()

	panel := m.renderTitledPanel(40, 10, "Title", "body", true, lipgloss.Color("#7aa2f7"))
	if strings.Contains(panel, ui.ColorSelected) {
		t.Fatalf("expected active panel not to use selected background fill, got %q", panel)
	}
}
