package domain

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type S3Bucket struct {
	Name string
}

type S3Object struct {
	Key          string
	Size         int64
	LastModified string
	ContentType  string
	Metadata     map[string]string
}

type S3ObjectVersion struct {
	Key          string
	VersionID    string
	IsLatest     bool
	Size         int64
	LastModified string
}

type S3Manager interface {
	ListBuckets(ctx context.Context, cfg *AWSConfig) ([]S3Bucket, error)
	ListObjects(ctx context.Context, cfg *AWSConfig, bucket string, prefix string) ([]S3Object, error)
	DeleteObject(ctx context.Context, cfg *AWSConfig, bucket string, key string) error
	DeleteBucket(ctx context.Context, cfg *AWSConfig, bucket string) error
	CreateBucket(ctx context.Context, cfg *AWSConfig, name string) error
	CreateFolder(ctx context.Context, cfg *AWSConfig, bucket string, key string) error
	UploadObject(ctx context.Context, cfg *AWSConfig, bucket string, key string, filePath string) error
	UploadObjectMultipart(ctx context.Context, cfg *AWSConfig, bucket string, key string, filePath string) error
	UploadObjectWithMetadata(ctx context.Context, cfg *AWSConfig, bucket string, key string, filePath string, metadata map[string]string) error
	GetPresignedURL(ctx context.Context, cfg *AWSConfig, bucket string, key string) (string, error)
	DownloadObject(ctx context.Context, cfg *AWSConfig, bucket string, key string, destPath string) error
	HeadObject(ctx context.Context, cfg *AWSConfig, bucket string, key string) (contentType string, metadata map[string]string, err error)
	ListObjectVersions(ctx context.Context, cfg *AWSConfig, bucket string, key string) ([]S3ObjectVersion, error)
	DeleteObjectVersion(ctx context.Context, cfg *AWSConfig, bucket string, key string, versionID string) error
}

type SQSQueue struct {
	URL                string
	ARN                string
	Name               string
	MessagesAvailable  int
	MessagesDelayed    int
	MessagesNotVisible int
}

type SQSMessage struct {
	ID            string
	Body          string
	ReceiptHandle string
}

type SQSManager interface {
	ListQueues(ctx context.Context, cfg *AWSConfig) ([]SQSQueue, error)
	SendMessage(ctx context.Context, cfg *AWSConfig, queueURL string, body string) error
	ReceiveMessages(ctx context.Context, cfg *AWSConfig, queueURL string, maxMessages int) ([]SQSMessage, error)
	DeleteMessage(ctx context.Context, cfg *AWSConfig, queueURL string, receiptHandle string) error
	PurgeQueue(ctx context.Context, cfg *AWSConfig, queueURL string) error
	DeleteQueue(ctx context.Context, cfg *AWSConfig, queueURL string) error
	CreateQueue(ctx context.Context, cfg *AWSConfig, name string) (string, error)
	GetQueueAttributes(ctx context.Context, cfg *AWSConfig, queueURL string, attributeNames []string) (map[string]string, error)
	SetQueueAttributes(ctx context.Context, cfg *AWSConfig, queueURL string, attributes map[string]string) error
	ResolveQueueURL(ctx context.Context, cfg *AWSConfig, queueRef string) (string, error)
	ResolveQueueARN(ctx context.Context, cfg *AWSConfig, queueURL string) (string, error)
}

type SNSTopic struct {
	ARN  string
	Name string
}

type Secret struct {
	ARN               string
	Name              string
	Description       string
	CreatedDate       string
	LastChangedDate   string
	DeletedDate       string
	PrimaryRegion     string
	KMSKeyID          string
	RotationEnabled   bool
	RotationLambdaARN string
	VersionCount      int
}

type SecretVersion struct {
	VersionID   string
	Stages      []string
	CreatedDate string
}

type SecretValue struct {
	VersionID          string
	SecretString       string
	SecretBinaryBase64 string
}

type SNSSubscription struct {
	ARN          string
	TopicARN     string
	Protocol     string
	Endpoint     string
	FilterPolicy map[string][]string
	FilterScope  string
}

type SNSManager interface {
	ListTopics(ctx context.Context, cfg *AWSConfig) ([]SNSTopic, error)
	PublishMessage(ctx context.Context, cfg *AWSConfig, topicARN string, body string, subject string, messageAttributes map[string]string) error
	ListSubscriptions(ctx context.Context, cfg *AWSConfig, topicARN string) ([]SNSSubscription, error)
	ListAllSubscriptions(ctx context.Context, cfg *AWSConfig) ([]SNSSubscription, error)
	CreateTopic(ctx context.Context, cfg *AWSConfig, name string) (SNSTopic, error)
	DeleteTopic(ctx context.Context, cfg *AWSConfig, topicARN string) error
	CreateSubscription(ctx context.Context, cfg *AWSConfig, topicARN string, protocol string, endpoint string, filterPolicy map[string][]string, filterScope string) (SNSSubscription, error)
	DeleteSubscription(ctx context.Context, cfg *AWSConfig, subscriptionARN string) error
	GetSubscriptionAttributes(ctx context.Context, cfg *AWSConfig, subscriptionARN string) (map[string]string, error)
	SetSubscriptionAttributes(ctx context.Context, cfg *AWSConfig, subscriptionARN string, attributeName string, attributeValue string) error
}

type SecretsManager interface {
	ListSecrets(ctx context.Context, cfg *AWSConfig) ([]Secret, error)
	DescribeSecret(ctx context.Context, cfg *AWSConfig, secretID string) (Secret, error)
	GetSecretValue(ctx context.Context, cfg *AWSConfig, secretID string, versionID string, versionStage string) (SecretValue, error)
	ListSecretVersions(ctx context.Context, cfg *AWSConfig, secretID string) ([]SecretVersion, error)
	CreateSecret(ctx context.Context, cfg *AWSConfig, name string, value string, description string) (Secret, error)
	UpdateSecretValue(ctx context.Context, cfg *AWSConfig, secretID string, value string, description string) (SecretValue, error)
	UpdateSecretVersionStage(ctx context.Context, cfg *AWSConfig, secretID string, versionStage string, moveToVersionID string, removeFromVersionID string) error
	DeleteSecret(ctx context.Context, cfg *AWSConfig, secretID string, recoveryWindowDays int, forceDeleteWithoutRecovery bool) error
	RestoreSecret(ctx context.Context, cfg *AWSConfig, secretID string) error
}

func QueueARNFromURL(queueURL, regionHint string) string {
	parsed, err := url.Parse(queueURL)
	if err != nil {
		return ""
	}

	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathParts) < 2 {
		return ""
	}

	hostParts := strings.Split(parsed.Hostname(), ".")
	region := strings.TrimSpace(regionHint)
	if len(hostParts) >= 3 && hostParts[0] == "sqs" {
		region = hostParts[1]
	}
	accountID := pathParts[0]
	queueName := pathParts[1]
	if queueName == "" || region == "" || accountID == "" {
		return ""
	}

	return fmt.Sprintf("arn:aws:sqs:%s:%s:%s", region, accountID, queueName)
}

func SNSSubscriptionKey(topicARN, protocol, endpoint string) string {
	return strings.Join([]string{topicARN, protocol, endpoint}, "|")
}
