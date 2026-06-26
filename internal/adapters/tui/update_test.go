package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"monostack/internal/domain"
	"monostack/internal/usecase"
)

func TestUpdate_WindowSize(t *testing.T) {
	m := mkModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.width != 100 {
		t.Errorf("expected width 100, got %d", model.width)
	}
	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestUpdate_ConfigLoaded(t *testing.T) {
	m := mkModel()
	cfg := &domain.AWSConfig{
		ServiceName:    "test",
		EndpointURL:    "http://localhost:4566",
		Region:         "us-east-1",
		UseMock:        true,
		LeftPanelRatio: 0.3,
		PanelRatios: map[string]float64{
			domain.ServiceS3:      0.6,
			domain.ServiceSecrets: 0.4,
		},
	}
	msg := configLoadedMsg{Config: cfg}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.config == nil {
		t.Fatal("expected config to be non-nil")
	}
	if model.config.ServiceName != "test" {
		t.Errorf("expected ServiceName 'test', got %q", model.config.ServiceName)
	}
	if model.leftPanelRatio != 0.6 {
		t.Errorf("expected leftPanelRatio to sync from active panel, got %f", model.leftPanelRatio)
	}
	if model.loading {
		t.Error("expected loading to be false after config loaded")
	}
}

func TestUpdate_ResizePersistsActivePanelRatio(t *testing.T) {
	m := mkModel()
	m.config = &domain.AWSConfig{
		LeftPanelRatio: 0.5,
		PanelRatios: map[string]float64{
			domain.ServiceS3:      0.4,
			domain.ServiceSecrets: 0.6,
		},
	}
	m.activeTab = panelSecrets
	m.leftPanelRatio = 0.6

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if cmd == nil {
		t.Fatal("expected resize to persist config")
	}
	if model.leftPanelRatio != 0.65 {
		t.Fatalf("expected secrets ratio to increase to 0.65, got %f", model.leftPanelRatio)
	}
	if model.config.PanelRatios[domain.ServiceSecrets] != 0.65 {
		t.Fatalf("expected secrets panel ratio to be persisted, got %f", model.config.PanelRatios[domain.ServiceSecrets])
	}
	if model.config.PanelRatios[domain.ServiceS3] != 0.4 {
		t.Fatalf("expected S3 panel ratio to remain unchanged, got %f", model.config.PanelRatios[domain.ServiceS3])
	}

	result, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model, ok = result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.activeTab != panelS3 {
		t.Fatalf("expected tab switch to S3, got %v", model.activeTab)
	}
	if model.leftPanelRatio != 0.4 {
		t.Fatalf("expected S3 ratio to restore to 0.4, got %f", model.leftPanelRatio)
	}
}

func TestConfigFromSettingsInputsPreservesPanelRatios(t *testing.T) {
	m := mkModel()
	m.config = &domain.AWSConfig{
		LeftPanelRatio: 0.6,
		PanelRatios: map[string]float64{
			domain.ServiceSecrets: 0.6,
		},
	}
	m.leftPanelRatio = 0.6
	m.settingsInputs = make([]textinput.Model, 8)
	for i := range m.settingsInputs {
		m.settingsInputs[i] = textinput.New()
	}
	m.settingsInputs[0].SetValue("MiniStack")
	m.settingsInputs[1].SetValue("http://localhost:4566")
	m.settingsInputs[2].SetValue("us-east-1")
	m.settingsInputs[3].SetValue("key")
	m.settingsInputs[4].SetValue("secret")
	m.settingsInputs[5].SetValue("true")
	m.settingsInputs[6].SetValue("/tmp/monostack")
	m.settingsInputs[7].SetValue("s3,sqs,sns,secrets")

	cfg := m.configFromSettingsInputs()
	if cfg.LeftPanelRatio != 0.6 {
		t.Fatalf("expected left panel ratio to be preserved, got %f", cfg.LeftPanelRatio)
	}
	if cfg.PanelRatios[domain.ServiceSecrets] != 0.6 {
		t.Fatalf("expected panel ratios to be preserved, got %f", cfg.PanelRatios[domain.ServiceSecrets])
	}
}

func TestUpdate_StatusMsg(t *testing.T) {
	m := mkModel()
	msg := statusMsg{Message: "all good"}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.statusMsg != "all good" {
		t.Errorf("expected statusMsg 'all good', got %q", model.statusMsg)
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	m := mkModel()
	msg := errMsg{Error: testError{msg: "something failed"}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.errorMsg != "something failed" {
		t.Errorf("expected errorMsg 'something failed', got %q", model.errorMsg)
	}
	if model.statusMsg != "" {
		t.Errorf("expected statusMsg to be empty, got %q", model.statusMsg)
	}
	if model.loading {
		t.Error("expected loading to be false after error")
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := mkModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	result, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit")
	}

	if _, ok := result.(Model); !ok {
		t.Fatalf("expected Model, got %T", result)
	}
}

func TestUpdate_HelpToggle(t *testing.T) {
	m := mkModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showHelpModal {
		t.Error("expected showHelpModal to be true after ? key")
	}
}

func TestUpdate_S3BucketsLoaded(t *testing.T) {
	m := mkModel()
	buckets := []domain.S3Bucket{{Name: "bucket1"}, {Name: "bucket2"}}
	msg := s3BucketsLoadedMsg{Buckets: buckets}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if len(model.buckets) != 2 {
		t.Errorf("expected 2 buckets, got %d", len(model.buckets))
	}
	if model.loading {
		t.Error("expected loading to be false after buckets loaded")
	}
}

func TestUpdate_MoveSelection(t *testing.T) {
	m := mkModel()
	m.activeTab = panelSQS
	m.queues = []domain.SQSQueue{
		{Name: "queue1"},
		{Name: "queue2"},
		{Name: "queue3"},
	}
	m.selectedQueueIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.selectedQueueIndex != 1 {
		t.Errorf("expected selectedQueueIndex 1 after down, got %d", model.selectedQueueIndex)
	}
}

func TestUpdate_TabSwitch(t *testing.T) {
	m := mkModel()
	m.activeTab = panelS3

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.activeTab != panelSNS {
		t.Errorf("expected activeTab panelSNS, got %v", model.activeTab)
	}
}

func TestUpdate_OpenSecretValueModalConfiguresViewport(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.activeTab = panelSecrets
	m.secrets = []domain.Secret{{Name: "secret-dev-webapi-tenants", ARN: "arn:aws:secretsmanager:us-east-1:000000000000:secret:secret-dev-webapi-tenants"}}
	m.selectedSecretIndex = 0
	m.secretValueDisplay = `{
  "placeholder": "replace-me",
  "ConnectionStrings__DatabaseConnection": "Host=aurora-postgres-dev.cluster-cjwkqcmaeru1.sa-east-1.rds.amazonaws.com;Port=5432;Database=db_tenants;Username=;Password=;"
}`

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSecretValueModal {
		t.Fatal("expected secret value modal to open")
	}
	if model.secretValueViewport.Width == 0 || model.secretValueViewport.Height == 0 {
		t.Fatalf("expected secret value viewport to be configured, got %dx%d", model.secretValueViewport.Width, model.secretValueViewport.Height)
	}
	if !strings.Contains(model.secretValueViewport.View(), "ConnectionStrings__DatabaseConnection") {
		t.Fatalf("expected viewport to render the secret value content, got %q", model.secretValueViewport.View())
	}
}

func TestUpdate_SecretUpdateCtrlLClearsOnlyValue(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.activeTab = panelSecrets
	m.secrets = []domain.Secret{{Name: "secret-dev-webapi-tenants", ARN: "arn:aws:secretsmanager:us-east-1:000000000000:secret:secret-dev-webapi-tenants", Description: "CorpX-Tenant-Secret"}}
	m.selectedSecretIndex = 0
	m.secretValueDisplay = "{\n  \"foo\": \"bar\"\n}"
	m.showSecretUpdateModal = true
	m.secretUpdateStep = 0
	m.secretUpdateValueInput.SetValue(m.secretValueDisplay)

	msg := tea.KeyMsg{Type: tea.KeyCtrlL}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.secretUpdateValueInput.Value() != "" {
		t.Fatalf("expected value editor to be cleared, got %q", model.secretUpdateValueInput.Value())
	}
	if !model.showSecretUpdateModal {
		t.Fatal("expected update modal to remain open")
	}
}

func TestOrderedSecretVersions_CurrentFirst(t *testing.T) {
	versions := orderedSecretVersions([]domain.SecretVersion{
		{VersionID: "v2", Stages: []string{"AWSPREVIOUS"}},
		{VersionID: "v1", Stages: []string{"AWSCURRENT"}},
		{VersionID: "v3", Stages: nil},
	})

	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].VersionID != "v1" {
		t.Fatalf("expected current version first, got %q", versions[0].VersionID)
	}
	if label := secretVersionVisualLabel(0, versions[0]); label != "v1 (current)" {
		t.Fatalf("expected current label v1 (current), got %q", label)
	}
	if label := secretVersionVisualLabel(1, versions[1]); label != "v2" {
		t.Fatalf("expected second label v2, got %q", label)
	}
}

func TestRenderBrandWordmark_CompactOmitsBracketMark(t *testing.T) {
	m := mkModel()
	wordmark := m.renderBrandWordmark(true)
	if strings.Contains(wordmark, "[M]") {
		t.Fatalf("expected compact brand wordmark to omit [M], got %q", wordmark)
	}
	if !strings.Contains(wordmark, "MonoStack") {
		t.Fatalf("expected compact brand wordmark to contain MonoStack, got %q", wordmark)
	}
}

func TestSecretValueCopyTextUsesFormattedValue(t *testing.T) {
	m := mkModel()
	m.secretValueDisplay = "{\n  \"foo\": \"bar\"\n}"
	if got := m.secretValueCopyText(); got != "{\n  \"foo\": \"bar\"\n}" {
		t.Fatalf("expected formatted value to be copied, got %q", got)
	}
}

func TestFormatSecretValueDisplay_NormalizesUnicodeQuoteEscapes(t *testing.T) {
	value := `{"Local_baas":"{\u0022apiKey\u0022:\u0022pk_test\u0022,\u0022apiSecret\u0022:\u0022sk_test\u0022}"}`

	got := formatSecretValueDisplay(value)

	if strings.Contains(got, `\u0022`) {
		t.Fatalf("expected display to normalize unicode quote escapes, got %q", got)
	}
	if !strings.Contains(got, `\"apiKey\":\"pk_test\"`) {
		t.Fatalf("expected nested json string to use standard escaped quotes, got %q", got)
	}
}

func TestUpdate_SecretListUsesExactName(t *testing.T) {
	m := mkModel()
	m.width = 140
	m.height = 48
	m.activeTab = panelSecrets
	m.secrets = []domain.Secret{{
		Name:        "secret-dev-webapi-tenants",
		ARN:         "arn:aws:secretsmanager:us-east-1:000000000000:secret:secret-dev-webapi-tenants-78422e",
		Description: "CorpX-Tenant-Secret",
	}}
	m.selectedSecretIndex = 0
	m.secretVersions = []domain.SecretVersion{{VersionID: "v1", Stages: []string{"AWSCURRENT"}}}

	view := m.renderSecretsPanel()
	if !strings.Contains(view, "secret-dev-webapi-tenants") {
		t.Fatalf("expected exact secret name to appear in view, got %q", view)
	}
	if strings.Contains(view, "> secret-dev-webapi-tenants-78422e") {
		t.Fatalf("expected ARN suffix not to be duplicated in list line, got %q", view)
	}
	if !strings.Contains(view, "Description: CorpX-Tenant-Secret") {
		t.Fatalf("expected description metadata to be visible, got %q", view)
	}
	if strings.Contains(view, "Secret Value") {
		t.Fatalf("expected main details panel not to render secret value preview, got %q", view)
	}
}

func TestUpdate_SecretsNavigationSwitchesBetweenLists(t *testing.T) {
	m := mkModel()
	m.width = 140
	m.height = 48
	m.activeTab = panelSecrets
	m.secrets = []domain.Secret{{
		Name: "secret-dev-webapi-tenants",
		ARN:  "arn:aws:secretsmanager:us-east-1:000000000000:secret:secret-dev-webapi-tenants",
	}}
	m.secretVersions = []domain.SecretVersion{
		{VersionID: "v1", Stages: []string{"AWSPREVIOUS"}},
		{VersionID: "v2", Stages: []string{"AWSCURRENT"}},
	}
	m.selectedSecretIndex = 0
	m.selectedSecretVersionIndex = 1
	m.secretsFocus = focusSecrets

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.secretsFocus != focusSecretVersions {
		t.Fatalf("expected focus to move to versions, got %v", model.secretsFocus)
	}

	result, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model, ok = result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.secretsFocus != focusSecrets {
		t.Fatalf("expected focus to move back to secrets, got %v", model.secretsFocus)
	}
}

func TestPromoteSecretVersionCmd_UsesCurrentStage(t *testing.T) {
	mockSecrets := &mockSecretsManager{}
	mockSecrets.UpdateSecretVersionStageFunc = func(ctx context.Context, cfg *domain.AWSConfig, secretID, versionStage, moveToVersionID, removeFromVersionID string) error {
		if secretID != "arn:secret" || versionStage != "AWSCURRENT" || moveToVersionID != "v1" || removeFromVersionID != "v2" {
			t.Fatalf("unexpected stage update args: secretID=%q stage=%q moveTo=%q removeFrom=%q", secretID, versionStage, moveToVersionID, removeFromVersionID)
		}
		return nil
	}

	m := mkModel()
	m.awsUseCase = usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, &mockSNSManager{}, mockSecrets)
	m.config = &domain.AWSConfig{}
	m.secretVersions = []domain.SecretVersion{
		{VersionID: "v1", Stages: []string{"AWSPREVIOUS"}},
		{VersionID: "v2", Stages: []string{"AWSCURRENT"}},
	}

	msg := m.promoteSecretVersionCmd("arn:secret", "v1")()
	updated, ok := msg.(secretStageUpdatedMsg)
	if !ok {
		t.Fatalf("expected secretStageUpdatedMsg, got %T", msg)
	}
	if updated.VersionID != "v1" {
		t.Fatalf("expected promoted version id v1, got %q", updated.VersionID)
	}
}

func TestNewModelWithDI(t *testing.T) {
	s3 := &mockS3Manager{}
	sqs := &mockSQSManager{}
	sns := &mockSNSManager{}
	cfg := &mockConfigStore{}

	awsUC := usecase.NewAWSUseCase(s3, sqs, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCase(cfg)
	snapshotUC := usecase.NewSnapshotUseCase(awsUC, cfgUC)
	m := NewModel(awsUC, cfgUC, snapshotUC)

	if m.awsUseCase == nil {
		t.Error("expected non-nil awsUseCase")
	}
	if m.configUseCase == nil {
		t.Error("expected non-nil configUseCase")
	}
}

func TestUpdate_ConnectionRefusedFriendlyError(t *testing.T) {
	m := mkModel()
	m.config = &domain.AWSConfig{
		ServiceName: "MiniStack",
		EndpointURL: "http://localhost:4566",
	}
	msg := errMsg{Error: testError{msg: "dial tcp [::1]:4566: connect: connection refused"}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	expected := "Service unreachable at http://localhost:4566. Is MiniStack/LocalStack running? (Press [4] for Settings to configure or enable Mock Mode)"
	if model.errorMsg != expected {
		t.Errorf("expected errorMsg %q, got %q", expected, model.errorMsg)
	}
}

func TestUpdate_TabSwitchClearsError(t *testing.T) {
	m := mkModel()
	m.errorMsg = "some error"
	m.statusMsg = "some status"
	m.activeTab = panelS3

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.errorMsg != "" {
		t.Errorf("expected errorMsg to be cleared, got %q", model.errorMsg)
	}
	if model.statusMsg != "" {
		t.Errorf("expected statusMsg to be cleared, got %q", model.statusMsg)
	}
}

func TestUpdate_DeleteKeyTabHandling(t *testing.T) {

	m := mkModel()
	m.activeTab = panelSQS
	m.queues = []domain.SQSQueue{
		{Name: "test-queue", URL: "http://sqs/test-queue"},
	}
	m.selectedQueueIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSqsConfirmDelete {
		t.Error("expected showSqsConfirmDelete to be true when 'd' pressed on SQS tab")
	}

	m2 := mkModel()
	m2.activeTab = panelSNS
	m2.topics = []domain.SNSTopic{
		{Name: "test-topic", ARN: "arn:sns:test-topic"},
	}
	m2.selectedTopicIndex = 0
	m2.snsSubFocus = focusTopics

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	result, _ = m2.Update(msg)
	model, ok = result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSnsConfirmDelete {
		t.Error("expected showSnsConfirmDelete to be true when 'd' pressed on SNS tab")
	}
}

func TestUpdate_BatchSubscribeKeyTabHandling(t *testing.T) {

	m := mkModel()
	m.activeTab = panelSQS
	m.queues = []domain.SQSQueue{
		{Name: "test-queue", URL: "http://sqs/test-queue"},
	}
	m.topics = []domain.SNSTopic{
		{Name: "test-topic", ARN: "arn:sns:test-topic"},
	}
	m.selectedQueueIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSqsBatchSubModal {
		t.Error("expected showSqsBatchSubModal to be true when 'b' pressed on SQS tab")
	}
}

func TestUpdate_S3DeleteHandling(t *testing.T) {

	m := mkModel()
	m.activeTab = panelS3
	m.s3Focus = focusBuckets
	m.buckets = []domain.S3Bucket{{Name: "test-bucket"}}
	m.selectedBucketIndex = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showS3ConfirmDelete {
		t.Error("expected showS3ConfirmDelete to be true when 'd' pressed on S3 bucket list")
	}
	if !model.s3DeleteIsBucket {
		t.Error("expected s3DeleteIsBucket to be true")
	}
	if model.s3DeleteBucket != "test-bucket" {
		t.Errorf("expected s3DeleteBucket 'test-bucket', got %q", model.s3DeleteBucket)
	}

	m2 := mkModel()
	m2.activeTab = panelS3
	m2.s3Focus = focusObjects
	m2.buckets = []domain.S3Bucket{{Name: "test-bucket"}}
	m2.selectedBucketIndex = 0
	m2.objects = []domain.S3Object{{Key: "test-key", Size: 123}}
	m2.selectedObjectIndex = 0

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	result, _ = m2.Update(msg)
	model2, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model2.showS3ConfirmDelete {
		t.Error("expected showS3ConfirmDelete to be true when 'd' pressed on S3 object list")
	}
	if model2.s3DeleteIsBucket {
		t.Error("expected s3DeleteIsBucket to be false for object deletion")
	}
	if model2.s3DeleteBucket != "test-bucket" || model2.s3DeleteKey != "test-key" {
		t.Errorf("expected bucket 'test-bucket' and key 'test-key', got bucket=%q key=%q", model2.s3DeleteBucket, model2.s3DeleteKey)
	}

	m3 := mkModel()
	m3.activeTab = panelS3
	m3.showS3ConfirmDelete = true
	m3.s3DeleteIsBucket = true
	m3.s3DeleteBucket = "test-bucket"

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	result, cmd := m3.Update(msg)
	model3, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model3.showS3ConfirmDelete {
		t.Error("expected showS3ConfirmDelete to be false after confirmation")
	}
	if !model3.loading {
		t.Error("expected loading to be true after confirmation")
	}
	if cmd == nil {
		t.Error("expected a command to be returned for bucket deletion")
	}

	m4 := mkModel()
	m4.activeTab = panelS3
	m4.selectedBucketIndex = 5
	m4.selectedObjectIndex = 5
	m4.s3Focus = focusObjects

	msgDeleted := s3BucketDeletedMsg{Bucket: "test-bucket"}
	result, cmd = m4.Update(msgDeleted)
	model4, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model4.selectedBucketIndex != 0 {
		t.Errorf("expected selectedBucketIndex to reset to 0, got %d", model4.selectedBucketIndex)
	}
	if model4.selectedObjectIndex != 0 {
		t.Errorf("expected selectedObjectIndex to reset to 0, got %d", model4.selectedObjectIndex)
	}
	if model4.s3Focus != focusBuckets {
		t.Errorf("expected s3Focus to reset to focusBuckets, got %v", model4.s3Focus)
	}
	if cmd == nil {
		t.Error("expected a reload command for buckets")
	}
}

func TestUpdate_S3CreateHandling(t *testing.T) {

	m := mkModel()
	m.activeTab = panelS3
	m.s3Focus = focusBuckets

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showS3CreateModal {
		t.Error("expected showS3CreateModal to be true")
	}

	model.s3CreateInput.SetValue("my-bucket")
	msgEsc := tea.KeyMsg{Type: tea.KeyEsc}
	resultEsc, _ := model.Update(msgEsc)
	modelEsc, ok := resultEsc.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEsc)
	}
	if modelEsc.showS3CreateModal {
		t.Error("expected showS3CreateModal to be false")
	}
	if modelEsc.s3CreateInput.Value() != "" {
		t.Errorf("expected input to be reset, got %q", modelEsc.s3CreateInput.Value())
	}

	model.s3CreateInput.SetValue("new-bucket")
	msgEnter := tea.KeyMsg{Type: tea.KeyEnter}
	resultEnter, cmd := model.Update(msgEnter)
	modelEnter, ok := resultEnter.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEnter)
	}
	if modelEnter.showS3CreateModal {
		t.Error("expected showS3CreateModal to be false after submit")
	}
	if !modelEnter.loading {
		t.Error("expected loading to be true after submit")
	}
	if cmd == nil {
		t.Fatal("expected creation cmd to be returned")
	}

	mCreated := mkModel()
	mCreated.loading = true
	msgCreated := s3BucketCreatedMsg{Bucket: "new-bucket"}
	resultCreated, cmdReload := mCreated.Update(msgCreated)
	modelCreated, ok := resultCreated.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultCreated)
	}
	if modelCreated.loading {
		t.Error("expected loading to be false after s3BucketCreatedMsg")
	}
	if modelCreated.statusMsg != `Bucket "new-bucket" created successfully!` {
		t.Errorf("unexpected statusMsg: %q", modelCreated.statusMsg)
	}
	if cmdReload == nil {
		t.Error("expected reload command")
	}
}

func TestUpdate_S3FolderCreateHandling(t *testing.T) {
	m := mkModel()
	m.activeTab = panelS3
	m.s3Focus = focusObjects
	m.buckets = []domain.S3Bucket{{Name: "assets"}}
	m.objects = []domain.S3Object{{Key: "reports/2026/invoice.json"}}
	m.selectedBucketIndex = 0
	m.selectedObjectIndex = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showS3CreateFolderModal {
		t.Fatal("expected showS3CreateFolderModal to be true")
	}
	if model.s3FolderInput.Value() != "reports/2026/" {
		t.Fatalf("expected folder input to be prefilled, got %q", model.s3FolderInput.Value())
	}

	model.s3FolderInput.SetValue("reports/2026")
	resultEnter, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	modelEnter, ok := resultEnter.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEnter)
	}
	if modelEnter.showS3CreateFolderModal {
		t.Fatal("expected modal to close after submit")
	}
	if !modelEnter.loading {
		t.Fatal("expected loading to be true after folder submit")
	}
	if cmd == nil {
		t.Fatal("expected create folder command")
	}
}

func TestUpdate_SQSCreateHandling(t *testing.T) {

	m := mkModel()
	m.activeTab = panelSQS

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	result, _ := m.Update(msg)
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSqsCreateModal {
		t.Error("expected showSqsCreateModal to be true")
	}

	model.sqsCreateInput.SetValue("my-queue")
	msgEsc := tea.KeyMsg{Type: tea.KeyEsc}
	resultEsc, _ := model.Update(msgEsc)
	modelEsc, ok := resultEsc.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEsc)
	}
	if modelEsc.showSqsCreateModal {
		t.Error("expected showSqsCreateModal to be false")
	}
	if modelEsc.sqsCreateInput.Value() != "" {
		t.Errorf("expected input to be reset, got %q", modelEsc.sqsCreateInput.Value())
	}

	model.sqsCreateInput.SetValue("new-queue")
	msgEnter := tea.KeyMsg{Type: tea.KeyEnter}
	resultEnter, cmd := model.Update(msgEnter)
	modelEnter, ok := resultEnter.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEnter)
	}
	if modelEnter.showSqsCreateModal {
		t.Error("expected showSqsCreateModal to be false after submit")
	}
	if !modelEnter.loading {
		t.Error("expected loading to be true after submit")
	}
	if cmd == nil {
		t.Fatal("expected creation cmd to be returned")
	}

	mCreated := mkModel()
	mCreated.loading = true
	msgCreated := sqsQueueCreatedMsg{Name: "new-queue"}
	resultCreated, cmdReload := mCreated.Update(msgCreated)
	modelCreated, ok := resultCreated.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultCreated)
	}
	if modelCreated.loading {
		t.Error("expected loading to be false after sqsQueueCreatedMsg")
	}
	if modelCreated.statusMsg != `Queue "new-queue" created successfully!` {
		t.Errorf("unexpected statusMsg: %q", modelCreated.statusMsg)
	}
	if cmdReload == nil {
		t.Error("expected reload command")
	}
}

func TestUpdate_SQSPurgeAllHandling(t *testing.T) {
	m := mkModel()
	m.activeTab = panelSQS
	m.queues = []domain.SQSQueue{{Name: "q1", URL: "url-1"}, {Name: "q2", URL: "url-2"}}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showSqsPurgeAllConfirm {
		t.Fatal("expected purge-all modal to open")
	}

	resultEnter, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	modelEnter, ok := resultEnter.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", resultEnter)
	}
	if modelEnter.showSqsPurgeAllConfirm {
		t.Fatal("expected purge-all modal to close")
	}
	if !modelEnter.loading {
		t.Fatal("expected loading after purge-all submit")
	}
	if cmd == nil {
		t.Fatal("expected purge-all command")
	}
}

func TestUpdate_EnterOpensSQSInspection(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.activeTab = panelSQS
	m.queues = []domain.SQSQueue{{Name: "error", URL: "http://localhost:4566/000000000000/error", ARN: "arn:aws:sqs:us-east-1:000000000000:error"}}
	m.commandLogs = []CommandLogEntry{{Action: "error", Target: "error", Output: "failed", Error: testError{msg: "boom"}}}
	m.awsUseCase = usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{
		ReceiveMessagesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, maxMessages int) ([]domain.SQSMessage, error) {
			return []domain.SQSMessage{{ID: "m-1", Body: `{"kind":"test"}`}}, nil
		},
	}, &mockSNSManager{}, &mockSecretsManager{})

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.loading {
		t.Fatal("expected loading to start")
	}
	if cmd == nil {
		t.Fatal("expected inspection command")
	}

	msg := cmd()
	loaded, ok := msg.(inspectionLoadedMsg)
	if !ok {
		t.Fatalf("expected inspectionLoadedMsg, got %T", msg)
	}
	if !strings.Contains(loaded.Content, `real queue named "error"`) {
		t.Fatalf("expected queue clarification in inspection content: %q", loaded.Content)
	}

	result, _ = model.Update(loaded)
	model, ok = result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.showInspectionModal {
		t.Fatal("expected inspection modal to open")
	}

	result, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model, ok = result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if model.showInspectionModal {
		t.Fatal("expected inspection modal to close on esc")
	}
}

func TestUpdate_EnterOpensSNSInspection(t *testing.T) {
	m := mkModel()
	m.width = 120
	m.height = 40
	m.activeTab = panelSNS
	m.topics = []domain.SNSTopic{{Name: "dev-webapi-billings-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-billings-sns"}}
	m.subscriptions = []domain.SNSSubscription{{
		ARN:          "sub-1",
		TopicARN:     "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns",
		Protocol:     "sns",
		Endpoint:     "arn:aws:sns:us-east-1:000000000000:dev-webapi-billings-sns",
		FilterPolicy: map[string][]string{"event_type": []string{"pix_received"}},
		FilterScope:  domain.SubscriptionFilterScopeMessageBody,
	}}
	m.snsOutgoingCount = 0
	m.snsSubFocus = focusSubs

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	if !model.loading {
		t.Fatal("expected loading to start")
	}
	if cmd == nil {
		t.Fatal("expected inspection command")
	}

	msg := cmd()
	loaded, ok := msg.(inspectionLoadedMsg)
	if !ok {
		t.Fatalf("expected inspectionLoadedMsg, got %T", msg)
	}
	if !strings.Contains(loaded.Content, "message body") {
		t.Fatalf("expected filter scope in inspection content: %q", loaded.Content)
	}
}
