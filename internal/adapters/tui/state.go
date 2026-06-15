package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	"monostack/internal/domain"
)

type s3PanelState struct {
	buckets             []domain.S3Bucket
	selectedBucketIndex int
	objects             []domain.S3Object
	selectedObjectIndex int
	s3ObjectsCache      map[string][]domain.S3Object
	s3ObjectsLoadedFor  string
	s3Focus             focusArea

	currentPrefix   string   // current folder prefix, e.g. "images/2024/"
	prefixStack     []string // breadcrumb stack
	breadcrumbPositions []int // X positions of breadcrumb segment starts

	filterInput  textinput.Model
	filterActive bool
	filterQuery  string

	sortField     int // 0=name, 1=size, 2=date
	sortAscending bool

	showS3ConfirmDelete bool
	s3DeleteBucket      string
	s3DeleteKey         string
	s3DeleteIsBucket    bool
	showS3CreateModal   bool
	s3CreateInput       textinput.Model
	showS3CreateFolderModal bool
	s3FolderInput           textinput.Model
	showS3UploadModal   bool
	s3UploadPathInput   textinput.Model
	s3UploadKeyInput    textinput.Model
	s3UploadFocus       int
	showS3PreviewModal  bool

	showVersionsModal    bool
	objectVersions       []domain.S3ObjectVersion
	selectedVersionIndex int
	versionObjectKey     string
}

type sqsPanelState struct {
	queues                []domain.SQSQueue
	selectedQueueIndex    int
	queueSubscriptions    []domain.SNSSubscription
	selectedQueueSubIndex int
	peekMessages          []domain.SQSMessage
	selectedPeekIndex     int
	sqsFocus              focusArea

	filterInput  textinput.Model
	filterActive bool
	filterQuery  string

	sortField     int // 0=name, 1=available, 2=delayed, 3=in-flight
	sortAscending bool

	showPeekModal           bool
	showSqsSendModal        bool
	sqsSendMessageInput     textinput.Model
	showSqsConfirmDelete    bool
	sqsDeleteQueueURL       string
	sqsDeleteQueueName      string
	showSqsPurgeAllConfirm  bool
	showSqsCreateModal      bool
	sqsCreateInput          textinput.Model
	showSqsSubDeleteConfirm bool
	sqsDeleteSubARN         string
	sqsDeleteSubLabel       string
}

type snsPanelState struct {
	topics             []domain.SNSTopic
	selectedTopicIndex int
	subscriptions      []domain.SNSSubscription
	allSubscriptions   []domain.SNSSubscription
	selectedSubIndex   int
	managedSubs        []domain.ManagedSubscription
	snsSubFocus        focusArea
	snsOutgoingCount   int

	filterInput  textinput.Model
	filterActive bool
	filterQuery  string

	sortField     int // 0=name
	sortAscending bool

	showSnsPublishModal     bool
	snsPublishInput         textinput.Model
	snsPublishAttrsInput    textinput.Model
	showSnsCreateTopicModal bool
	snsCreateTopicInput     textinput.Model
	showSnsSimpleSubModal   bool
	snsSimpleSubStep        int
	snsSimpleSubCursor      int
	snsSimpleSubSources     []domain.SNSTopic
	snsSimpleSubEventInput  textinput.Model
	showSnsBatchSubModal    bool
	snsBatchSubCursor       int
	snsBatchSubList         []toggleOption
	showSnsYamlImportModal  bool
	showSnsYamlApplyConfirm bool
	snsYamlImportTextarea   textarea.Model
	snsYamlSavedContent     map[string]string
	snsYamlCurrentTopicARN  string
	snsYamlPendingContent   string
	showSqsBatchSubModal    bool
	sqsBatchSubCursor       int
	sqsBatchSubList         []toggleOption
	showSnsSubEditModal     bool
	snsSubEditInput         textinput.Model
	showSnsConfirmDelete    bool
	snsDeleteARN            string
	snsDeleteLabel          string
	snsDeleteIsTopic        bool
}

type secretsPanelState struct {
	secrets                    []domain.Secret
	selectedSecretIndex        int
	secretVersions             []domain.SecretVersion
	selectedSecretVersionIndex int
	secretValue                domain.SecretValue
	secretValueDisplay         string
	secretValueLoadedFor       string
	secretDetailsLoadedFor     string
	secretValueViewport        viewport.Model
	secretsFocus               focusArea

	filterInput  textinput.Model
	filterActive bool
	filterQuery  string

	sortField     int // 0=name, 1=date
	sortAscending bool

	showSecretCreateModal      bool
	showSecretUpdateModal      bool
	showSecretDeleteConfirm    bool
	showSecretRestoreConfirm   bool
	showSecretPromoteConfirm   bool
	showSecretClipboardConfirm bool
	secretClipboardText        string
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
}
