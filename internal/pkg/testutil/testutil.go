package testutil

import (
	"context"

	"monostack/internal/domain"
)

type MockConfigStore struct {
	LoadFunc          func() (*domain.AWSConfig, error)
	SaveFunc          func(cfg *domain.AWSConfig) error
	ListProfilesFunc  func() ([]string, error)
	SwitchProfileFunc func(name string) error
	SaveProfileFunc   func(name string, cfg *domain.AWSConfig) error
	DeleteProfileFunc func(name string) error
}

func (m *MockConfigStore) Load() (*domain.AWSConfig, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc()
	}
	return &domain.AWSConfig{}, nil
}

func (m *MockConfigStore) Save(cfg *domain.AWSConfig) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(cfg)
	}
	return nil
}

func (m *MockConfigStore) ListProfiles() ([]string, error) {
	if m.ListProfilesFunc != nil {
		return m.ListProfilesFunc()
	}
	return nil, nil
}

func (m *MockConfigStore) SwitchProfile(name string) error {
	if m.SwitchProfileFunc != nil {
		return m.SwitchProfileFunc(name)
	}
	return nil
}

func (m *MockConfigStore) SaveProfile(name string, cfg *domain.AWSConfig) error {
	if m.SaveProfileFunc != nil {
		return m.SaveProfileFunc(name, cfg)
	}
	return nil
}

func (m *MockConfigStore) DeleteProfile(name string) error {
	if m.DeleteProfileFunc != nil {
		return m.DeleteProfileFunc(name)
	}
	return nil
}

type MockS3Manager struct {
	ListBucketsFunc             func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error)
	ListObjectsFunc             func(ctx context.Context, cfg *domain.AWSConfig, bucket, prefix string) ([]domain.S3Object, error)
	DeleteObjectFunc            func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) error
	DeleteBucketFunc            func(ctx context.Context, cfg *domain.AWSConfig, bucket string) error
	CreateBucketFunc            func(ctx context.Context, cfg *domain.AWSConfig, name string) error
	CreateFolderFunc            func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) error
	UploadObjectFunc            func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string) error
	UploadObjectMultipartFunc   func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string) error
	UploadObjectWithMetadataFunc func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string, metadata map[string]string) error
	GetPresignedURLFunc         func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) (string, error)
	DownloadObjectFunc          func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, destPath string) error
	HeadObjectFunc              func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) (string, map[string]string, error)
	ListObjectVersionsFunc      func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) ([]domain.S3ObjectVersion, error)
	DeleteObjectVersionFunc     func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, versionID string) error
}

func (m *MockS3Manager) ListBuckets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
	if m.ListBucketsFunc != nil {
		return m.ListBucketsFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockS3Manager) ListObjects(ctx context.Context, cfg *domain.AWSConfig, bucket, prefix string) ([]domain.S3Object, error) {
	if m.ListObjectsFunc != nil {
		return m.ListObjectsFunc(ctx, cfg, bucket, prefix)
	}
	return nil, nil
}

func (m *MockS3Manager) DeleteObject(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) error {
	if m.DeleteObjectFunc != nil {
		return m.DeleteObjectFunc(ctx, cfg, bucket, key)
	}
	return nil
}

func (m *MockS3Manager) DeleteBucket(ctx context.Context, cfg *domain.AWSConfig, bucket string) error {
	if m.DeleteBucketFunc != nil {
		return m.DeleteBucketFunc(ctx, cfg, bucket)
	}
	return nil
}

func (m *MockS3Manager) CreateBucket(ctx context.Context, cfg *domain.AWSConfig, name string) error {
	if m.CreateBucketFunc != nil {
		return m.CreateBucketFunc(ctx, cfg, name)
	}
	return nil
}

func (m *MockS3Manager) CreateFolder(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) error {
	if m.CreateFolderFunc != nil {
		return m.CreateFolderFunc(ctx, cfg, bucket, key)
	}
	return nil
}

func (m *MockS3Manager) UploadObject(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string) error {
	if m.UploadObjectFunc != nil {
		return m.UploadObjectFunc(ctx, cfg, bucket, key, filePath)
	}
	return nil
}

func (m *MockS3Manager) UploadObjectMultipart(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string) error {
	if m.UploadObjectMultipartFunc != nil {
		return m.UploadObjectMultipartFunc(ctx, cfg, bucket, key, filePath)
	}
	return nil
}

func (m *MockS3Manager) GetPresignedURL(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) (string, error) {
	if m.GetPresignedURLFunc != nil {
		return m.GetPresignedURLFunc(ctx, cfg, bucket, key)
	}
	return "", nil
}

func (m *MockS3Manager) UploadObjectWithMetadata(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string, metadata map[string]string) error {
	if m.UploadObjectWithMetadataFunc != nil {
		return m.UploadObjectWithMetadataFunc(ctx, cfg, bucket, key, filePath, metadata)
	}
	return nil
}

func (m *MockS3Manager) HeadObject(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) (string, map[string]string, error) {
	if m.HeadObjectFunc != nil {
		return m.HeadObjectFunc(ctx, cfg, bucket, key)
	}
	return "", nil, nil
}

func (m *MockS3Manager) DownloadObject(ctx context.Context, cfg *domain.AWSConfig, bucket, key, destPath string) error {
	if m.DownloadObjectFunc != nil {
		return m.DownloadObjectFunc(ctx, cfg, bucket, key, destPath)
	}
	return nil
}

func (m *MockS3Manager) ListObjectVersions(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) ([]domain.S3ObjectVersion, error) {
	if m.ListObjectVersionsFunc != nil {
		return m.ListObjectVersionsFunc(ctx, cfg, bucket, key)
	}
	return nil, nil
}

func (m *MockS3Manager) DeleteObjectVersion(ctx context.Context, cfg *domain.AWSConfig, bucket, key, versionID string) error {
	if m.DeleteObjectVersionFunc != nil {
		return m.DeleteObjectVersionFunc(ctx, cfg, bucket, key, versionID)
	}
	return nil
}

type MockSQSManager struct {
	ListQueuesFunc         func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error)
	SendMessageFunc        func(ctx context.Context, cfg *domain.AWSConfig, queueURL, body string) error
	ReceiveMessagesFunc    func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, maxMessages int) ([]domain.SQSMessage, error)
	DeleteMessageFunc      func(ctx context.Context, cfg *domain.AWSConfig, queueURL, receiptHandle string) error
	PurgeQueueFunc         func(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error
	DeleteQueueFunc        func(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error
	CreateQueueFunc        func(ctx context.Context, cfg *domain.AWSConfig, name string) (string, error)
	GetQueueAttributesFunc func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error)
	SetQueueAttributesFunc func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error
	ResolveQueueURLFunc    func(ctx context.Context, cfg *domain.AWSConfig, queueRef string) (string, error)
	ResolveQueueARNFunc    func(ctx context.Context, cfg *domain.AWSConfig, queueURL string) (string, error)
}

func (m *MockSQSManager) ListQueues(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
	if m.ListQueuesFunc != nil {
		return m.ListQueuesFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockSQSManager) SendMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL, body string) error {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, cfg, queueURL, body)
	}
	return nil
}

func (m *MockSQSManager) ReceiveMessages(ctx context.Context, cfg *domain.AWSConfig, queueURL string, maxMessages int) ([]domain.SQSMessage, error) {
	if m.ReceiveMessagesFunc != nil {
		return m.ReceiveMessagesFunc(ctx, cfg, queueURL, maxMessages)
	}
	return nil, nil
}

func (m *MockSQSManager) DeleteMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL, receiptHandle string) error {
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, cfg, queueURL, receiptHandle)
	}
	return nil
}

func (m *MockSQSManager) PurgeQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	if m.PurgeQueueFunc != nil {
		return m.PurgeQueueFunc(ctx, cfg, queueURL)
	}
	return nil
}

func (m *MockSQSManager) DeleteQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	if m.DeleteQueueFunc != nil {
		return m.DeleteQueueFunc(ctx, cfg, queueURL)
	}
	return nil
}

func (m *MockSQSManager) CreateQueue(ctx context.Context, cfg *domain.AWSConfig, name string) (string, error) {
	if m.CreateQueueFunc != nil {
		return m.CreateQueueFunc(ctx, cfg, name)
	}
	return "", nil
}

func (m *MockSQSManager) GetQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
	if m.GetQueueAttributesFunc != nil {
		return m.GetQueueAttributesFunc(ctx, cfg, queueURL, attributeNames)
	}
	return nil, nil
}

func (m *MockSQSManager) SetQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error {
	if m.SetQueueAttributesFunc != nil {
		return m.SetQueueAttributesFunc(ctx, cfg, queueURL, attributes)
	}
	return nil
}

func (m *MockSQSManager) ResolveQueueURL(ctx context.Context, cfg *domain.AWSConfig, queueRef string) (string, error) {
	if m.ResolveQueueURLFunc != nil {
		return m.ResolveQueueURLFunc(ctx, cfg, queueRef)
	}
	return "", nil
}

func (m *MockSQSManager) ResolveQueueARN(ctx context.Context, cfg *domain.AWSConfig, queueURL string) (string, error) {
	if m.ResolveQueueARNFunc != nil {
		return m.ResolveQueueARNFunc(ctx, cfg, queueURL)
	}
	return "", nil
}

type MockSNSManager struct {
	ListTopicsFunc                func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error)
	PublishMessageFunc            func(ctx context.Context, cfg *domain.AWSConfig, topicARN, body, subject string, attrs map[string]string) error
	ListSubscriptionsFunc         func(ctx context.Context, cfg *domain.AWSConfig, topicARN string) ([]domain.SNSSubscription, error)
	ListAllSubscriptionsFunc      func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error)
	CreateTopicFunc               func(ctx context.Context, cfg *domain.AWSConfig, name string) (domain.SNSTopic, error)
	DeleteTopicFunc               func(ctx context.Context, cfg *domain.AWSConfig, topicARN string) error
	CreateSubscriptionFunc        func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error)
	DeleteSubscriptionFunc        func(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) error
	GetSubscriptionAttributesFunc func(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) (map[string]string, error)
	SetSubscriptionAttributesFunc func(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN, attributeName, attributeValue string) error
}

func (m *MockSNSManager) ListTopics(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
	if m.ListTopicsFunc != nil {
		return m.ListTopicsFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockSNSManager) PublishMessage(ctx context.Context, cfg *domain.AWSConfig, topicARN, body, subject string, attrs map[string]string) error {
	if m.PublishMessageFunc != nil {
		return m.PublishMessageFunc(ctx, cfg, topicARN, body, subject, attrs)
	}
	return nil
}

func (m *MockSNSManager) ListSubscriptions(ctx context.Context, cfg *domain.AWSConfig, topicARN string) ([]domain.SNSSubscription, error) {
	if m.ListSubscriptionsFunc != nil {
		return m.ListSubscriptionsFunc(ctx, cfg, topicARN)
	}
	return nil, nil
}

func (m *MockSNSManager) ListAllSubscriptions(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error) {
	if m.ListAllSubscriptionsFunc != nil {
		return m.ListAllSubscriptionsFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockSNSManager) CreateTopic(ctx context.Context, cfg *domain.AWSConfig, name string) (domain.SNSTopic, error) {
	if m.CreateTopicFunc != nil {
		return m.CreateTopicFunc(ctx, cfg, name)
	}
	return domain.SNSTopic{}, nil
}

func (m *MockSNSManager) DeleteTopic(ctx context.Context, cfg *domain.AWSConfig, topicARN string) error {
	if m.DeleteTopicFunc != nil {
		return m.DeleteTopicFunc(ctx, cfg, topicARN)
	}
	return nil
}

func (m *MockSNSManager) CreateSubscription(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
	if m.CreateSubscriptionFunc != nil {
		return m.CreateSubscriptionFunc(ctx, cfg, topicARN, protocol, endpoint, filterPolicy, filterScope)
	}
	return domain.SNSSubscription{}, nil
}

func (m *MockSNSManager) DeleteSubscription(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) error {
	if m.DeleteSubscriptionFunc != nil {
		return m.DeleteSubscriptionFunc(ctx, cfg, subscriptionARN)
	}
	return nil
}

func (m *MockSNSManager) GetSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) (map[string]string, error) {
	if m.GetSubscriptionAttributesFunc != nil {
		return m.GetSubscriptionAttributesFunc(ctx, cfg, subscriptionARN)
	}
	return nil, nil
}

func (m *MockSNSManager) SetSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN, attributeName, attributeValue string) error {
	if m.SetSubscriptionAttributesFunc != nil {
		return m.SetSubscriptionAttributesFunc(ctx, cfg, subscriptionARN, attributeName, attributeValue)
	}
	return nil
}

type MockSecretsManager struct {
	ListSecretsFunc              func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.Secret, error)
	DescribeSecretFunc           func(ctx context.Context, cfg *domain.AWSConfig, secretID string) (domain.Secret, error)
	GetSecretValueFunc           func(ctx context.Context, cfg *domain.AWSConfig, secretID, versionID, versionStage string) (domain.SecretValue, error)
	ListSecretVersionsFunc       func(ctx context.Context, cfg *domain.AWSConfig, secretID string) ([]domain.SecretVersion, error)
	CreateSecretFunc             func(ctx context.Context, cfg *domain.AWSConfig, name, value, description string) (domain.Secret, error)
	UpdateSecretValueFunc        func(ctx context.Context, cfg *domain.AWSConfig, secretID, value, description string) (domain.SecretValue, error)
	UpdateSecretVersionStageFunc func(ctx context.Context, cfg *domain.AWSConfig, secretID, versionStage, moveToVersionID, removeFromVersionID string) error
	DeleteSecretFunc             func(ctx context.Context, cfg *domain.AWSConfig, secretID string, recoveryWindowDays int, forceDeleteWithoutRecovery bool) error
	RestoreSecretFunc            func(ctx context.Context, cfg *domain.AWSConfig, secretID string) error
}

func (m *MockSecretsManager) ListSecrets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.Secret, error) {
	if m.ListSecretsFunc != nil {
		return m.ListSecretsFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockSecretsManager) DescribeSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) (domain.Secret, error) {
	if m.DescribeSecretFunc != nil {
		return m.DescribeSecretFunc(ctx, cfg, secretID)
	}
	return domain.Secret{}, nil
}

func (m *MockSecretsManager) GetSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID, versionID, versionStage string) (domain.SecretValue, error) {
	if m.GetSecretValueFunc != nil {
		return m.GetSecretValueFunc(ctx, cfg, secretID, versionID, versionStage)
	}
	return domain.SecretValue{}, nil
}

func (m *MockSecretsManager) ListSecretVersions(ctx context.Context, cfg *domain.AWSConfig, secretID string) ([]domain.SecretVersion, error) {
	if m.ListSecretVersionsFunc != nil {
		return m.ListSecretVersionsFunc(ctx, cfg, secretID)
	}
	return nil, nil
}

func (m *MockSecretsManager) CreateSecret(ctx context.Context, cfg *domain.AWSConfig, name, value, description string) (domain.Secret, error) {
	if m.CreateSecretFunc != nil {
		return m.CreateSecretFunc(ctx, cfg, name, value, description)
	}
	return domain.Secret{}, nil
}

func (m *MockSecretsManager) UpdateSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID, value, description string) (domain.SecretValue, error) {
	if m.UpdateSecretValueFunc != nil {
		return m.UpdateSecretValueFunc(ctx, cfg, secretID, value, description)
	}
	return domain.SecretValue{}, nil
}

func (m *MockSecretsManager) UpdateSecretVersionStage(ctx context.Context, cfg *domain.AWSConfig, secretID, versionStage, moveToVersionID, removeFromVersionID string) error {
	if m.UpdateSecretVersionStageFunc != nil {
		return m.UpdateSecretVersionStageFunc(ctx, cfg, secretID, versionStage, moveToVersionID, removeFromVersionID)
	}
	return nil
}

func (m *MockSecretsManager) DeleteSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string, recoveryWindowDays int, forceDeleteWithoutRecovery bool) error {
	if m.DeleteSecretFunc != nil {
		return m.DeleteSecretFunc(ctx, cfg, secretID, recoveryWindowDays, forceDeleteWithoutRecovery)
	}
	return nil
}

func (m *MockSecretsManager) RestoreSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) error {
	if m.RestoreSecretFunc != nil {
		return m.RestoreSecretFunc(ctx, cfg, secretID)
	}
	return nil
}
