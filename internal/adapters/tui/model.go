package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"monostack/internal/domain"
	"monostack/internal/pkg/ui"
	"monostack/internal/usecase"
)

type activePanel int

const (
	panelS3 activePanel = iota
	panelSQS
	panelSNS
	panelSecrets
	panelConfig
)

type focusArea int

const (
	focusBuckets focusArea = iota
	focusObjects
	focusTopics
	focusSubs
	focusQueues
	focusQueueSubs
	focusSecrets
	focusSecretVersions
)

type selectionContext string

const (
	selectionNone           selectionContext = ""
	selectionS3Buckets      selectionContext = "s3_buckets"
	selectionS3Objects      selectionContext = "s3_objects"
	selectionSQSQueues      selectionContext = "sqs_queues"
	selectionSQSSubs        selectionContext = "sqs_subscriptions"
	selectionSNSTopics      selectionContext = "sns_topics"
	selectionSNSSubs        selectionContext = "sns_subscriptions"
	selectionSecrets        selectionContext = "secrets"
	selectionSecretVersions selectionContext = "secret_versions"
	selectionSettings       selectionContext = "settings"
	selectionCommandLogs    selectionContext = "command_logs"
)

type toggleOption struct {
	label   string
	arn     string
	checked bool
}

type Model struct {
	awsUseCase      *usecase.AWSUseCase
	configUseCase   *usecase.ConfigUseCase
	snapshotUseCase *usecase.SnapshotUseCase
	config          *domain.AWSConfig

	activeTab   activePanel
	s3Focus     focusArea
	width       int
	height      int
	loading     bool
	statusMsg   string
	errorMsg    string
	statusMsgID int

	buckets             []domain.S3Bucket
	selectedBucketIndex int
	objects             []domain.S3Object
	selectedObjectIndex int
	s3ObjectsCache      map[string][]domain.S3Object
	s3ObjectsLoadedFor  string

	showS3ConfirmDelete bool
	s3DeleteBucket      string
	s3DeleteKey         string
	s3DeleteIsBucket    bool

	showS3CreateModal bool
	s3CreateInput     textinput.Model

	showS3CreateFolderModal bool
	s3FolderInput           textinput.Model

	showS3UploadModal bool
	s3UploadPathInput textinput.Model
	s3UploadKeyInput  textinput.Model
	s3UploadFocus     int

	showS3PreviewModal bool

	queues              []domain.SQSQueue
	selectedQueueIndex  int
	peekMessages        []domain.SQSMessage
	showPeekModal       bool
	showSqsSendModal    bool
	sqsSendMessageInput textinput.Model

	showSqsConfirmDelete bool
	sqsDeleteQueueURL    string
	sqsDeleteQueueName   string

	showSqsPurgeAllConfirm bool
	sqsPurgeAllInput       textinput.Model

	showSqsCreateModal bool
	sqsCreateInput     textinput.Model

	queueSubscriptions    []domain.SNSSubscription
	selectedQueueSubIndex int
	sqsFocus              focusArea

	showSqsSubDeleteConfirm bool
	sqsDeleteSubARN         string
	sqsDeleteSubLabel       string

	topics               []domain.SNSTopic
	selectedTopicIndex   int
	showSnsPublishModal  bool
	snsPublishInput      textinput.Model
	snsPublishAttrsInput textinput.Model

	subscriptions    []domain.SNSSubscription
	allSubscriptions []domain.SNSSubscription
	selectedSubIndex int
	managedSubs      []domain.ManagedSubscription
	snsSubFocus      focusArea
	snsOutgoingCount int
	secretsFocus     focusArea

	secrets                    []domain.Secret
	selectedSecretIndex        int
	secretVersions             []domain.SecretVersion
	selectedSecretVersionIndex int
	secretValue                domain.SecretValue
	secretValueDisplay         string
	secretValueLoadedFor       string
	secretDetailsLoadedFor     string
	secretValueViewport        viewport.Model
	showSecretCreateModal      bool
	showSecretUpdateModal      bool
	showSecretDeleteConfirm    bool
	showSecretRestoreConfirm   bool
	showSecretPromoteConfirm   bool
	showSecretValueModal       bool
	secretCreateStep           int
	secretUpdateStep           int
	secretCreateNameInput      textinput.Model
	secretCreateValueInput     textarea.Model
	secretUpdateValueInput     textarea.Model
	secretDeleteID             string
	secretDeleteName           string
	secretPromoteSecretID      string
	secretPromoteVersionID     string
	secretPromoteVersionLabel  string
	secretPromoteCurrentID     string
	secretPromoteCurrentLabel  string

	showSnsCreateTopicModal bool
	snsCreateTopicInput     textinput.Model

	showSnsSimpleSubModal  bool
	snsSimpleSubStep       int
	snsSimpleSubCursor     int
	snsSimpleSubSources    []domain.SNSTopic
	snsSimpleSubEventInput textinput.Model

	showSnsBatchSubModal bool
	snsBatchSubCursor    int
	snsBatchSubList      []toggleOption

	showSnsYamlImportModal  bool
	showSnsYamlApplyConfirm bool
	snsYamlImportTextarea   textarea.Model
	snsYamlSavedContent     map[string]string
	snsYamlCurrentTopicARN  string
	snsYamlPendingContent   string

	showSqsBatchSubModal bool
	sqsBatchSubCursor    int
	sqsBatchSubList      []toggleOption

	showSnsSubEditModal bool
	snsSubEditInput     textinput.Model

	showSnsConfirmDelete bool
	snsDeleteARN         string
	snsDeleteLabel       string
	snsDeleteIsTopic     bool

	showExportModal bool
	showImportModal bool
	exportPathInput textinput.Model
	importPathInput textinput.Model

	settingsInputs   []textinput.Model
	focusedInput     int
	settingsEditMode bool

	leftPanelRatio      float64
	showHelpModal       bool
	showLogsModal       bool
	showInspectionModal bool

	showSplash  bool
	splashFrame int

	commandLogs        []CommandLogEntry
	commandLogCursor   int
	logViewport        viewport.Model
	helpViewport       viewport.Model
	inspectionTitle    string
	inspectionSubtitle string
	inspectionContent  string
	inspectionViewport viewport.Model
	selectionActive    bool
	selectionContext   selectionContext
	selectionStart     int
	selectionEnd       int

	styles styles
}

type CommandLogEntry struct {
	Time   time.Time
	Action string
	Target string
	Output string
	Error  error
}

func (m *Model) cancelSpecialModes() {
	m.showHelpModal = false
	m.showLogsModal = false
	m.showInspectionModal = false
	m.showPeekModal = false
	m.showExportModal = false
	m.showImportModal = false
	m.clearSelection()
}

type styles struct {
	Header           lipgloss.Style
	Subtitle         lipgloss.Style
	Footer           lipgloss.Style
	ActiveTab        lipgloss.Style
	InactiveTab      lipgloss.Style
	BorderPanel      lipgloss.Style
	FocusedPanel     lipgloss.Style
	Title            lipgloss.Style
	Highlight        lipgloss.Style
	ListItem         lipgloss.Style
	SelectedListItem lipgloss.Style
	InputLabel       lipgloss.Style
	InputFocused     lipgloss.Style
	InputUnfocused   lipgloss.Style
	Modal            lipgloss.Style
	SuccessBadge     lipgloss.Style
	ErrorBadge       lipgloss.Style
	InfoText         lipgloss.Style
}

var Version = "0.0.4"

const splashFrameLimit = 20

func NewModel(awsUseCase *usecase.AWSUseCase, configUseCase *usecase.ConfigUseCase, snapshotUseCase *usecase.SnapshotUseCase) Model {

	s := styles{
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorMono)).
			Bold(true).
			Padding(0, 1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSubtle)).
			PaddingLeft(1),

		Footer: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorBg)).
			Foreground(lipgloss.Color(ui.ColorSubtle)).
			Padding(0, 1),

		ActiveTab: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorStack)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 1).
			MarginRight(1),

		InactiveTab: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorBg)).
			Foreground(lipgloss.Color(ui.ColorSubtle)).
			Padding(0, 1).
			MarginRight(1),

		BorderPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorMono)).
			Padding(0, 0),

		FocusedPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorStack)).
			Padding(0, 0),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorStack)).
			Bold(true).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorHighlight)),

		ListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorFg)),

		SelectedListItem: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorHighlight)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true),

		InputLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorAccent)).
			Bold(true),

		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ui.ColorStack)).
			Padding(0, 1).
			Width(44).
			MarginLeft(2),

		InputUnfocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ui.ColorBorder)).
			Padding(0, 1).
			Width(44).
			MarginLeft(2),

		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorStack)).
			Background(lipgloss.Color(ui.ColorBg)).
			Padding(1, 2),

		SuccessBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorSuccess)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 1),

		ErrorBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorError)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 1),

		InfoText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSubtle)),
	}

	sqsInput := textinput.New()
	sqsInput.Placeholder = `{"body": "payload"}`
	sqsInput.Width = 50

	snsInput := textinput.New()
	snsInput.Placeholder = "Message body to publish"
	snsInput.Width = 50

	snsAttrsInput := textinput.New()
	snsAttrsInput.Placeholder = "event_type=pix_received"
	snsAttrsInput.Width = 50

	createTopicInput := textinput.New()
	createTopicInput.Placeholder = "topic-name"
	createTopicInput.Width = 50

	s3CInput := textinput.New()
	s3CInput.Placeholder = "bucket-name"
	s3CInput.Width = 50

	s3UploadPath := textinput.New()
	s3UploadPath.Placeholder = "/absolute/path/to/file.png"
	s3UploadPath.Width = 56

	s3UploadKey := textinput.New()
	s3UploadKey.Placeholder = "folder/file.png"
	s3UploadKey.Width = 56

	s3FolderInput := textinput.New()
	s3FolderInput.Placeholder = "reports/2026/"
	s3FolderInput.Width = 56

	sqsCInput := textinput.New()
	sqsCInput.Placeholder = "queue-name"
	sqsCInput.Width = 50

	sqsPurgeAllInput := textinput.New()
	sqsPurgeAllInput.Placeholder = "purge all"
	sqsPurgeAllInput.Width = 24

	subEventInput := textinput.New()
	subEventInput.Placeholder = "pix_received, pix_sent (or leave empty for all)"
	subEventInput.Width = 50

	subEditInput := textinput.New()
	subEditInput.Placeholder = "pix_received, pix_sent"
	subEditInput.Width = 50

	secretNameInput := textinput.New()
	secretNameInput.Placeholder = "secret-name"
	secretNameInput.Width = 54

	secretValueTA := textarea.New()
	secretValueTA.Placeholder = "{\n  \"token\": \"...\"\n}"
	secretValueTA.SetWidth(60)
	secretValueTA.SetHeight(10)
	secretValueTA.ShowLineNumbers = false
	secretValueTA.KeyMap.InsertNewline.SetKeys("enter")
	secretValueTA.FocusedStyle.CursorLine = lipgloss.NewStyle()
	secretValueTA.FocusedStyle.Base = lipgloss.NewStyle()

	secretUpdateValueTA := textarea.New()
	secretUpdateValueTA.Placeholder = "{\n  \"token\": \"new-value\"\n}"
	secretUpdateValueTA.SetWidth(60)
	secretUpdateValueTA.SetHeight(10)
	secretUpdateValueTA.ShowLineNumbers = false
	secretUpdateValueTA.KeyMap.InsertNewline.SetKeys("enter")
	secretUpdateValueTA.FocusedStyle.CursorLine = lipgloss.NewStyle()
	secretUpdateValueTA.FocusedStyle.Base = lipgloss.NewStyle()

	yamlTA := textarea.New()
	yamlTA.Placeholder = "version: 1\n\nsubscriptions:\n  - name: my-service\n    topic: my-topic-sns\n    event_type:\n      - event_created"
	yamlTA.SetWidth(60)
	yamlTA.SetHeight(18)
	yamlTA.ShowLineNumbers = false
	yamlTA.KeyMap.InsertNewline.SetKeys("enter")
	yamlTA.FocusedStyle.CursorLine = lipgloss.NewStyle()
	yamlTA.FocusedStyle.Base = lipgloss.NewStyle()

	exportInput := textinput.New()
	exportInput.Placeholder = usecase.DefaultSnapshotPath()
	exportInput.Width = 50

	importInput := textinput.New()
	importInput.Placeholder = usecase.DefaultSnapshotPath()
	importInput.Width = 50

	settingsInputs := make([]textinput.Model, 8)
	for i := range settingsInputs {
		settingsInputs[i] = textinput.New()
		settingsInputs[i].Width = 40
	}

	return Model{
		awsUseCase:              awsUseCase,
		configUseCase:           configUseCase,
		snapshotUseCase:         snapshotUseCase,
		activeTab:               panelS3,
		s3Focus:                 focusBuckets,
		styles:                  s,
		s3ObjectsCache:          make(map[string][]domain.S3Object),
		sqsSendMessageInput:     sqsInput,
		snsPublishInput:         snsInput,
		snsPublishAttrsInput:    snsAttrsInput,
		snsCreateTopicInput:     createTopicInput,
		s3CreateInput:           s3CInput,
		s3UploadPathInput:       s3UploadPath,
		s3UploadKeyInput:        s3UploadKey,
		s3FolderInput:           s3FolderInput,
		sqsCreateInput:          sqsCInput,
		sqsPurgeAllInput:        sqsPurgeAllInput,
		snsSimpleSubEventInput:  subEventInput,
		snsSubEditInput:         subEditInput,
		snsYamlImportTextarea:   yamlTA,
		snsYamlSavedContent:     make(map[string]string),
		secretCreateNameInput:   secretNameInput,
		secretCreateValueInput:  secretValueTA,
		secretUpdateValueInput:  secretUpdateValueTA,
		exportPathInput:         exportInput,
		importPathInput:         importInput,
		settingsInputs:          settingsInputs,
		snsSubFocus:             focusTopics,
		sqsFocus:                focusQueues,
		secretsFocus:            focusSecrets,
		loading:                 true,
		leftPanelRatio:          0.5,
		showHelpModal:           false,
		showLogsModal:           false,
		showInspectionModal:     false,
		showSnsYamlApplyConfirm: false,
		secretValueDisplay:      "",
		showSplash:              true,
		logViewport:             viewport.New(0, 0),
		inspectionViewport:      viewport.New(0, 0),
		secretValueViewport:     viewport.New(0, 0),
	}
}

func (m Model) spinnerView() string {
	return ui.SpinnerFrames[m.splashFrame%len(ui.SpinnerFrames)]
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadConfigCmd(),
		m.splashTickCmd(),
	)
}
