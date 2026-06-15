package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"monostack/internal/domain"
)

type healthCheckMsg struct {
	OK  bool
	Err string
}

type configLoadedMsg struct {
	Config *domain.AWSConfig
}

type configSavedMsg struct {
	Config *domain.AWSConfig
}

type s3BucketsLoadedMsg struct {
	Buckets []domain.S3Bucket
}

type s3ObjectsLoadedMsg struct {
	Bucket  string
	Objects []domain.S3Object
}

type s3ObjectDeletedMsg struct {
	Bucket string
	Key    string
}

type s3BucketDeletedMsg struct {
	Bucket string
}

type s3BucketCreatedMsg struct {
	Bucket string
}

type s3FolderCreatedMsg struct {
	Bucket string
	Key    string
}

type s3ObjectDownloadedMsg struct {
	DestPath string
}

type s3ObjectUploadedMsg struct {
	Bucket string
	Key    string
}

type sqsQueuesLoadedMsg struct {
	Queues           []domain.SQSQueue
	AllSubscriptions []domain.SNSSubscription
}

type sqsQueuePurgedMsg struct {
	QueueURL string
}

type sqsQueuesPurgedMsg struct {
	Count int
}

type sqsQueueDeletedMsg struct {
	QueueURL string
	Name     string
}

type sqsQueueCreatedMsg struct {
	Name string
}

type sqsMessageSentMsg struct {
	QueueURL string
}

type sqsMessagesLoadedMsg struct {
	QueueURL string
	Messages []domain.SQSMessage
}

type sqsMessagesDeletedMsg struct {
	QueueURL string
	Count    int
}

type snsTopicsLoadedMsg struct {
	Topics           []domain.SNSTopic
	AllSubscriptions []domain.SNSSubscription
}

type snsTopicCreatedMsg struct {
	Topic domain.SNSTopic
}

type snsTopicDeletedMsg struct {
	ARN string
}

type secretsLoadedMsg struct {
	Secrets []domain.Secret
}

type secretDetailsLoadedMsg struct {
	Secret   domain.Secret
	Versions []domain.SecretVersion
	Value    domain.SecretValue
}

type secretCreatedMsg struct {
	Secret domain.Secret
}

type secretValueUpdatedMsg struct {
	SecretID string
	Value    domain.SecretValue
}

type secretDeletedMsg struct {
	SecretID string
	Name     string
}

type secretRestoredMsg struct {
	SecretID string
}

type secretStageUpdatedMsg struct {
	SecretID  string
	VersionID string
}

type snsMessagePublishedMsg struct {
	TopicARN string
}

type snsSubscriptionsLoadedMsg struct {
	Subscriptions         []domain.SNSSubscription
	IncomingSubscriptions []domain.SNSSubscription
	AllSubscriptions      []domain.SNSSubscription
}

type snsSubscriptionCreatedMsg struct {
	Subscription domain.SNSSubscription
}

type snsSubscriptionDeletedMsg struct {
	ARN string
}

type snsSubscriptionUpdatedMsg struct {
	ARN string
}

type managedSubscriptionsLoadedMsg struct {
	Subscriptions []domain.ManagedSubscription
}

type managedSubscriptionsUpdatedMsg struct{}

type snsBatchSubscriptionsCreatedMsg struct {
	Count int
}

type snsYamlImportAppliedMsg struct {
	Created   int
	Repaired  int
	Unchanged int
}

type sqsSubscriptionsLoadedMsg struct {
	Subscriptions []domain.SNSSubscription
}

type sqsBatchSubscriptionsCreatedMsg struct {
	Count int
}

type yamlScriptLoadedMsg struct {
	TopicARN string
	Content  string
}

type yamlScriptSavedMsg struct {
	TopicARN string
}

type profileExportedMsg struct {
	Path string
}

type profileImportedMsg struct {
	Config    *domain.AWSConfig
	SubsCount int
	Path      string
}

type statusMsg struct {
	Message string
}

type errMsg struct {
	Error error
}

type clearStatusMsg struct {
	id int
}

type splashTickMsg struct{}

type autoRefreshTickMsg struct{}

type progressMsg struct {
	Operation string
	Percent   float64
}

type progressDoneMsg struct {
	Operation string
	Result    tea.Msg
}

type sdkLogMsg struct {
	Text string
}

type logTickMsg struct{}

type singleResourceExportedMsg struct {
	Path string
}

type s3VersionsLoadedMsg struct {
	Bucket   string
	Key      string
	Versions []domain.S3ObjectVersion
}

type s3VersionDeletedMsg struct {
	Bucket    string
	Key       string
	VersionID string
}

type inspectionLoadedMsg struct {
	Title    string
	Subtitle string
	Content  string
}

type profilesLoadedMsg struct {
	Profiles []string
}

type profileSwitchedMsg struct {
	Name   string
	Config *domain.AWSConfig
}

type profileSavedMsg struct {
	Name   string
	Config *domain.AWSConfig
}

type profileDeletedMsg struct {
	Name   string
	Config *domain.AWSConfig
}
