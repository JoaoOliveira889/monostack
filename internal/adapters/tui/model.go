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
	selectionSQSTopics      selectionContext = "sqs_topics"
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
	program         *tea.Program
	logCh           chan string

	s3PanelState
	sqsPanelState
	snsPanelState
	secretsPanelState

	activeTab   activePanel
	width       int
	height      int
	loading     bool
	statusMsg   string
	errorMsg    string
	statusMsgID int

	leftPanelRatio float64
	showHelpModal  bool
	showLogsModal  bool
	showInspectionModal bool
	showSplash     bool
	splashFrame    int

	commandLogs      []CommandLogEntry
	commandLogCursor int
	logViewport      viewport.Model
	helpViewport     viewport.Model
	inspectionTitle    string
	inspectionSubtitle string
	inspectionContent  string
	inspectionViewport viewport.Model
	selectionActive    bool
	selectionContext   selectionContext
	selectionStart     int
	selectionEnd       int

	showExportModal       bool
	showImportModal       bool
	exportPathInput       textinput.Model
	importPathInput       textinput.Model
	showSingleExportModal bool
	singleExportPathInput textinput.Model
	showProgress          bool
	progress              progressBar
	progressTracker       *progressTracker

	settingsInputs   []textinput.Model
	focusedInput     int
	settingsEditMode bool

	showProfileModal    bool
	profileList         []string
	profileCursor       int
	profileCreateInput  textinput.Model
	showProfileDeleteConfirm bool
	profileDeleteName   string

	toasts                  []Toast
	toastTimerActive        bool

	showCommandPalette      bool
	commandPaletteInput     textinput.Model
	commandPaletteItems     []CommandPaletteItem
	commandPaletteFiltered  []CommandPaletteItem
	commandPaletteCursor    int

	multiSelectActive       bool
	s3MultiSelected         multiSelectSet
	s3ObjectsMultiSelected  multiSelectSet
	sqsMultiSelected        multiSelectSet
	sqsTopMultiSelected     multiSelectSet
	snsMultiSelected        multiSelectSet
	snsSubsMultiSelected    multiSelectSet
	secretsMultiSelected    multiSelectSet

	showMultiDeleteConfirm bool
	multiDeleteLabel       string
	multiDeleteKind        multiDeleteKind

	styles styles
}

type multiDeleteKind int

const (
	multiDelS3Buckets multiDeleteKind = iota
	multiDelS3Objects
	multiDelSQSQueues
	multiDelSNSTopics
	multiDelSecrets
)

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
	m.showSingleExportModal = false
	m.showSecretClipboardConfirm = false
	m.showProfileModal = false
	m.showProfileDeleteConfirm = false
	m.showCommandPalette = false
	m.clearSelection()
	m.clearMultiSelect()
}

func (m Model) anyModalOpen() bool {
	return m.showS3ConfirmDelete ||
		m.showS3CreateModal ||
		m.showS3CreateFolderModal ||
		m.showS3UploadModal ||
		m.showS3PreviewModal ||
		m.showVersionsModal ||
		m.showSqsConfirmDelete ||
		m.showSqsPurgeAllConfirm ||
		m.showSqsCreateModal ||
		m.showSqsSendModal ||
		m.showSqsBatchSubModal ||
		m.showSqsSubDeleteConfirm ||
		m.showPeekModal ||
		m.showSnsPublishModal ||
		m.showSnsCreateTopicModal ||
		m.showSnsSimpleSubModal ||
		m.showSnsBatchSubModal ||
		m.showSnsYamlImportModal ||
		m.showSnsYamlApplyConfirm ||
		m.showSnsConfirmDelete ||
		m.showSnsSubEditModal ||
		m.showSecretCreateModal ||
		m.showSecretUpdateModal ||
		m.showSecretDeleteConfirm ||
		m.showSecretRestoreConfirm ||
		m.showSecretPromoteConfirm ||
		m.showSecretClipboardConfirm ||
		m.showSecretValueModal ||
		m.showExportModal ||
		m.showImportModal ||
		m.showSingleExportModal ||
		m.showHelpModal ||
		m.showLogsModal ||
		m.showInspectionModal ||
		m.showProfileModal ||
		m.showProfileDeleteConfirm ||
		m.settingsEditMode ||
		m.showCommandPalette ||
		m.showMultiDeleteConfirm
}

type styles struct {
	Header           lipgloss.Style
	Subtitle         lipgloss.Style
	Footer           lipgloss.Style
	ActiveTab        lipgloss.Style
	InactiveTab      lipgloss.Style
	BorderPanel      lipgloss.Style
	FocusedPanel     lipgloss.Style
	Card             lipgloss.Style
	Cursor           lipgloss.Style
	ToastSuccess     lipgloss.Style
	ToastError       lipgloss.Style
	ToastInfo        lipgloss.Style
	Title            lipgloss.Style
	Highlight        lipgloss.Style
	ListItem         lipgloss.Style
	SelectedListItem lipgloss.Style
	MultiSelectedMarker lipgloss.Style
	CPCategory       lipgloss.Style
	CPItem           lipgloss.Style
	CPSelected       lipgloss.Style
	CPKey            lipgloss.Style
	InputLabel       lipgloss.Style
	InputFocused     lipgloss.Style
	InputUnfocused   lipgloss.Style
	Modal            lipgloss.Style
	SuccessBadge     lipgloss.Style
	WarningBadge     lipgloss.Style
	ErrorBadge       lipgloss.Style
	InfoText         lipgloss.Style
}

var Version = "0.0.4"

const splashFrameLimit = 20
const autoRefreshInterval = 5 * time.Second

func NewModel(awsUseCase *usecase.AWSUseCase, configUseCase *usecase.ConfigUseCase, snapshotUseCase *usecase.SnapshotUseCase) Model {
	s := makeDefaultStyles()

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

	s3FilterInput := textinput.New()
	s3FilterInput.Placeholder = "Filter..."
	s3FilterInput.Width = 30

	sqsFilterInput := textinput.New()
	sqsFilterInput.Placeholder = "Filter..."
	sqsFilterInput.Width = 30

	snsFilterInput := textinput.New()
	snsFilterInput.Placeholder = "Filter..."
	snsFilterInput.Width = 30

	secretsFilterInput := textinput.New()
	secretsFilterInput.Placeholder = "Filter..."
	secretsFilterInput.Width = 30

	exportInput := textinput.New()
	exportInput.Placeholder = usecase.DefaultSnapshotPath()
	exportInput.Width = 50

	importInput := textinput.New()
	importInput.Placeholder = usecase.DefaultSnapshotPath()
	importInput.Width = 50

	singleExportInput := textinput.New()
	singleExportInput.Placeholder = usecase.DefaultSnapshotPath()
	singleExportInput.Width = 50

	settingsInputs := make([]textinput.Model, 8)
	for i := range settingsInputs {
		settingsInputs[i] = textinput.New()
		settingsInputs[i].Width = 40
	}

	profileCreateInput := textinput.New()
	profileCreateInput.Placeholder = "profile-name"
	profileCreateInput.Width = 40

	cpInput := textinput.New()
	cpInput.Placeholder = "Type a command..."
	cpInput.Width = 40
	cpInput.Prompt = "> "
	cpInput.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorAccent)).
		Bold(true)

	return Model{
		awsUseCase:              awsUseCase,
		configUseCase:           configUseCase,
		snapshotUseCase:         snapshotUseCase,
		logCh:                   make(chan string, 256),
		activeTab:               panelS3,
		s3PanelState: s3PanelState{
			s3Focus:       focusBuckets,
			s3ObjectsCache: make(map[string][]domain.S3Object),
			s3CreateInput:  s3CInput,
			s3UploadPathInput: s3UploadPath,
			s3UploadKeyInput:  s3UploadKey,
			s3FolderInput:     s3FolderInput,
			filterInput:  s3FilterInput,
		},
		sqsPanelState: sqsPanelState{
			sqsSendMessageInput: sqsInput,
			sqsCreateInput:      sqsCInput,
			sqsFocus:            focusQueues,
			filterInput:  sqsFilterInput,
		},
		snsPanelState: snsPanelState{
			snsPublishInput:      snsInput,
			snsPublishAttrsInput: snsAttrsInput,
			snsCreateTopicInput:  createTopicInput,
			snsSimpleSubEventInput: subEventInput,
			snsSubEditInput:      subEditInput,
			snsYamlImportTextarea: yamlTA,
			snsYamlSavedContent:  make(map[string]string),
			snsSubFocus:          focusTopics,
			showSnsYamlApplyConfirm: false,
			filterInput:  snsFilterInput,
		},
		secretsPanelState: secretsPanelState{
			secretsFocus:          focusSecrets,
			secretCreateNameInput:  secretNameInput,
			secretCreateValueInput: secretValueTA,
			secretUpdateValueInput: secretUpdateValueTA,
			secretValueViewport:    viewport.New(0, 0),
			filterInput:  secretsFilterInput,
		},
		styles:           s,
		exportPathInput:       exportInput,
		importPathInput:       importInput,
		singleExportPathInput: singleExportInput,
		settingsInputs:   settingsInputs,
		loading:          true,
		leftPanelRatio:   0.5,
		showHelpModal:    false,
		showLogsModal:    false,
		showInspectionModal: false,
		showSplash:       true,
		logViewport:      viewport.New(0, 0),
		inspectionViewport: viewport.New(0, 0),
		profileCreateInput:  profileCreateInput,
		commandPaletteInput: cpInput,
		s3MultiSelected:        newMultiSelectSet(),
		s3ObjectsMultiSelected: newMultiSelectSet(),
		sqsMultiSelected:       newMultiSelectSet(),
		sqsTopMultiSelected:    newMultiSelectSet(),
		snsMultiSelected:       newMultiSelectSet(),
		snsSubsMultiSelected:   newMultiSelectSet(),
		secretsMultiSelected:   newMultiSelectSet(),
	}
}

func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

func (m *Model) LogCh() chan<- string {
	return m.logCh
}

func (m Model) spinnerView() string {
	return ui.SpinnerFrames[m.splashFrame%len(ui.SpinnerFrames)]
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadConfigCmd(),
		m.splashTickCmd(),
		m.autoRefreshTickCmd(),
		m.logCaptureCmd(),
	)
}

func makeDefaultStyles() styles {
	return styles{
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

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.ColorBorder)).
			Background(lipgloss.Color(ui.ColorBg)).
			Padding(1, 2),

		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorHighlight)).
			Bold(true),

		ToastSuccess: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorSuccess)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 2),

		ToastError: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorError)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 2),

		ToastInfo: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorCyan)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true).
			Padding(0, 2),

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

		MultiSelectedMarker: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSuccess)).
			Bold(true),

		CPCategory: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorAccent)).
			Bold(true),

		CPItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorFg)),

		CPSelected: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorHighlight)).
			Foreground(lipgloss.Color(ui.ColorBg)).
			Bold(true),

		CPKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorStack)),

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

		WarningBadge: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.ColorWarning)).
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
}
