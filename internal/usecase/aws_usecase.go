package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"monostack/internal/domain"
)

type AWSUseCase struct {
	s3      domain.S3Manager
	sqs     domain.SQSManager
	sns     domain.SNSManager
	secrets domain.SecretsManager
}

func NewAWSUseCase(s3 domain.S3Manager, sqs domain.SQSManager, sns domain.SNSManager, secrets domain.SecretsManager) *AWSUseCase {
	return &AWSUseCase{
		s3:      s3,
		sqs:     sqs,
		sns:     sns,
		secrets: secrets,
	}
}

func (uc *AWSUseCase) ListS3Buckets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
	return uc.s3.ListBuckets(ctx, cfg)
}

func (uc *AWSUseCase) ListS3Objects(ctx context.Context, cfg *domain.AWSConfig, bucket string, prefix string) ([]domain.S3Object, error) {
	return uc.s3.ListObjects(ctx, cfg, bucket, prefix)
}

func (uc *AWSUseCase) DeleteS3Object(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) error {
	return uc.s3.DeleteObject(ctx, cfg, bucket, key)
}

func (uc *AWSUseCase) DeleteS3Bucket(ctx context.Context, cfg *domain.AWSConfig, bucket string) error {
	return uc.s3.DeleteBucket(ctx, cfg, bucket)
}

func (uc *AWSUseCase) CreateS3Bucket(ctx context.Context, cfg *domain.AWSConfig, name string) error {
	return uc.s3.CreateBucket(ctx, cfg, name)
}

func (uc *AWSUseCase) CreateS3Folder(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) error {
	return uc.s3.CreateFolder(ctx, cfg, bucket, key)
}

func (uc *AWSUseCase) UploadS3Object(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, filePath string) error {
	return uc.s3.UploadObject(ctx, cfg, bucket, key, filePath)
}

func (uc *AWSUseCase) GetS3PresignedURL(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) (string, error) {
	return uc.s3.GetPresignedURL(ctx, cfg, bucket, key)
}

func (uc *AWSUseCase) DownloadS3Object(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, destPath string) error {
	return uc.s3.DownloadObject(ctx, cfg, bucket, key, destPath)
}

func (uc *AWSUseCase) ListSQSQueues(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
	return uc.sqs.ListQueues(ctx, cfg)
}

func (uc *AWSUseCase) SendSQSMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL string, body string) error {
	return uc.sqs.SendMessage(ctx, cfg, queueURL, body)
}

func (uc *AWSUseCase) ReceiveSQSMessages(ctx context.Context, cfg *domain.AWSConfig, queueURL string, maxMessages int) ([]domain.SQSMessage, error) {
	return uc.sqs.ReceiveMessages(ctx, cfg, queueURL, maxMessages)
}

func (uc *AWSUseCase) PurgeSQSQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	return uc.sqs.PurgeQueue(ctx, cfg, queueURL)
}

func (uc *AWSUseCase) PurgeSQSQueues(ctx context.Context, cfg *domain.AWSConfig, queueURLs []string) error {
	for _, queueURL := range queueURLs {
		if err := uc.sqs.PurgeQueue(ctx, cfg, queueURL); err != nil {
			return err
		}
	}
	return nil
}

func (uc *AWSUseCase) DeleteSQSQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	return uc.sqs.DeleteQueue(ctx, cfg, queueURL)
}

func (uc *AWSUseCase) CreateSQSQueue(ctx context.Context, cfg *domain.AWSConfig, name string) (string, error) {
	return uc.sqs.CreateQueue(ctx, cfg, name)
}

func (uc *AWSUseCase) GetSQSQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
	return uc.sqs.GetQueueAttributes(ctx, cfg, queueURL, attributeNames)
}

func (uc *AWSUseCase) SetSQSQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error {
	return uc.sqs.SetQueueAttributes(ctx, cfg, queueURL, attributes)
}

func (uc *AWSUseCase) ListSNSTopics(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
	return uc.sns.ListTopics(ctx, cfg)
}

func (uc *AWSUseCase) PublishSNSMessage(ctx context.Context, cfg *domain.AWSConfig, topicARN string, body string, subject string, attrs map[string]string) error {
	return uc.sns.PublishMessage(ctx, cfg, topicARN, body, subject, attrs)
}

func (uc *AWSUseCase) ListSNSSubscriptions(ctx context.Context, cfg *domain.AWSConfig, topicARN string) ([]domain.SNSSubscription, error) {
	return uc.sns.ListSubscriptions(ctx, cfg, topicARN)
}

func (uc *AWSUseCase) ListAllSNSSubscriptions(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error) {
	return uc.sns.ListAllSubscriptions(ctx, cfg)
}

func (uc *AWSUseCase) CreateSNSTopic(ctx context.Context, cfg *domain.AWSConfig, name string) (domain.SNSTopic, error) {
	return uc.sns.CreateTopic(ctx, cfg, name)
}

func (uc *AWSUseCase) DeleteSNSTopic(ctx context.Context, cfg *domain.AWSConfig, topicARN string) error {
	return uc.sns.DeleteTopic(ctx, cfg, topicARN)
}

func (uc *AWSUseCase) CreateSNSSubscription(ctx context.Context, cfg *domain.AWSConfig, topicARN string, protocol string, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
	if strings.EqualFold(protocol, "sns") {
		return domain.SNSSubscription{}, fmt.Errorf("sns topic-to-topic subscriptions are not supported; subscribe the destination SQS queue directly or republish from the consumer service")
	}
	if strings.EqualFold(protocol, "sqs") {
		resolvedEndpoint := endpoint
		if !strings.HasPrefix(endpoint, "arn:aws:sqs:") {
			var err error
			resolvedEndpoint, err = uc.sqs.ResolveQueueARN(ctx, cfg, endpoint)
			if err != nil {
				return domain.SNSSubscription{}, fmt.Errorf("failed to resolve SQS endpoint ARN from %q: %w", endpoint, err)
			}
		}

		queueURL, err := uc.sqs.ResolveQueueURL(ctx, cfg, endpoint)
		if err != nil {
			return domain.SNSSubscription{}, fmt.Errorf("failed to resolve SQS queue URL from %q: %w", endpoint, err)
		}
		if err := uc.ensureSQSQueuePolicy(ctx, cfg, queueURL, resolvedEndpoint, topicARN); err != nil {
			return domain.SNSSubscription{}, err
		}

		endpoint = resolvedEndpoint
	}
	return uc.sns.CreateSubscription(ctx, cfg, topicARN, protocol, endpoint, filterPolicy, filterScope)
}

func (uc *AWSUseCase) DeleteSNSSubscription(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) error {
	return uc.sns.DeleteSubscription(ctx, cfg, subscriptionARN)
}

func (uc *AWSUseCase) GetSNSSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) (map[string]string, error) {
	return uc.sns.GetSubscriptionAttributes(ctx, cfg, subscriptionARN)
}

func (uc *AWSUseCase) SetSNSSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string, attributeName string, attributeValue string) error {
	return uc.sns.SetSubscriptionAttributes(ctx, cfg, subscriptionARN, attributeName, attributeValue)
}

func (uc *AWSUseCase) ListSecrets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.Secret, error) {
	return uc.secrets.ListSecrets(ctx, cfg)
}

func (uc *AWSUseCase) DescribeSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) (domain.Secret, error) {
	return uc.secrets.DescribeSecret(ctx, cfg, secretID)
}

func (uc *AWSUseCase) GetSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID string, versionID string, versionStage string) (domain.SecretValue, error) {
	return uc.secrets.GetSecretValue(ctx, cfg, secretID, versionID, versionStage)
}

func (uc *AWSUseCase) ListSecretVersions(ctx context.Context, cfg *domain.AWSConfig, secretID string) ([]domain.SecretVersion, error) {
	return uc.secrets.ListSecretVersions(ctx, cfg, secretID)
}

func (uc *AWSUseCase) CreateSecret(ctx context.Context, cfg *domain.AWSConfig, name string, value string, description string) (domain.Secret, error) {
	return uc.secrets.CreateSecret(ctx, cfg, name, value, description)
}

func (uc *AWSUseCase) UpdateSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID string, value string, description string) (domain.SecretValue, error) {
	return uc.secrets.UpdateSecretValue(ctx, cfg, secretID, value, description)
}

func (uc *AWSUseCase) UpdateSecretVersionStage(ctx context.Context, cfg *domain.AWSConfig, secretID string, versionStage string, moveToVersionID string, removeFromVersionID string) error {
	return uc.secrets.UpdateSecretVersionStage(ctx, cfg, secretID, versionStage, moveToVersionID, removeFromVersionID)
}

func (uc *AWSUseCase) DeleteSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string, recoveryWindowDays int, forceDeleteWithoutRecovery bool) error {
	return uc.secrets.DeleteSecret(ctx, cfg, secretID, recoveryWindowDays, forceDeleteWithoutRecovery)
}

func (uc *AWSUseCase) RestoreSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) error {
	return uc.secrets.RestoreSecret(ctx, cfg, secretID)
}

func (uc *AWSUseCase) ensureSQSQueuePolicy(ctx context.Context, cfg *domain.AWSConfig, queueURL, queueARN, topicARN string) error {
	attrs, err := uc.sqs.GetQueueAttributes(ctx, cfg, queueURL, []string{"Policy"})
	if err != nil {
		attrs = map[string]string{}
	}

	policy := sqsQueuePolicy{Version: "2012-10-17"}
	if raw := attrs["Policy"]; raw != "" {
		_ = json.Unmarshal([]byte(raw), &policy)
	}

	if hasTopicPolicy(policy, topicARN) {
		return nil
	}

	policy.Statement = append(policy.Statement, sqsQueuePolicyStatement{
		Sid:       sanitizePolicySID(topicARN),
		Effect:    "Allow",
		Principal: map[string]string{"Service": "sns.amazonaws.com"},
		Action:    "SQS:SendMessage",
		Resource:  queueARN,
		Condition: map[string]map[string]string{
			"ArnEquals": {
				"aws:SourceArn": topicARN,
			},
		},
	})

	payload, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	return uc.sqs.SetQueueAttributes(ctx, cfg, queueURL, map[string]string{
		"Policy": string(payload),
	})
}

type sqsQueuePolicy struct {
	Version   string                    `json:"Version"`
	Statement []sqsQueuePolicyStatement `json:"Statement"`
}

type sqsQueuePolicyStatement struct {
	Sid       string                       `json:"Sid"`
	Effect    string                       `json:"Effect"`
	Principal map[string]string            `json:"Principal"`
	Action    string                       `json:"Action"`
	Resource  string                       `json:"Resource"`
	Condition map[string]map[string]string `json:"Condition,omitempty"`
}

func hasTopicPolicy(policy sqsQueuePolicy, topicARN string) bool {
	for _, stmt := range policy.Statement {
		if stmt.Condition == nil {
			continue
		}
		if stmt.Condition["ArnEquals"]["aws:SourceArn"] == topicARN {
			return true
		}
	}
	return false
}

func sanitizePolicySID(value string) string {
	replacer := strings.NewReplacer(":", "-", "/", "-", ".", "-", "_", "-")
	clean := replacer.Replace(value)
	if len(clean) > 64 {
		clean = clean[:64]
	}
	return "Allow" + clean
}
