package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"monostack/internal/domain"
)

func TestAWSUseCase_ListS3Buckets(t *testing.T) {
	expected := []domain.S3Bucket{{Name: "bucket1"}}
	s3 := &mockS3Manager{
		ListBucketsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
			return expected, nil
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	buckets, err := uc.ListS3Buckets(context.Background(), &domain.AWSConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Name != "bucket1" {
		t.Errorf("expected [bucket1], got %v", buckets)
	}
}

func TestAWSUseCase_ListS3Buckets_Error(t *testing.T) {
	s3 := &mockS3Manager{
		ListBucketsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
			return nil, errors.New("aws error")
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	_, err := uc.ListS3Buckets(context.Background(), &domain.AWSConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAWSUseCase_ListSQSQueues(t *testing.T) {
	var called bool
	sqs := &mockSQSManager{
		ListQueuesFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
			called = true
			return []domain.SQSQueue{{Name: "queue1"}}, nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, sqs, &mockSNSManager{}, &mockSecretsManager{})
	queues, err := uc.ListSQSQueues(context.Background(), &domain.AWSConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected ListQueues to be called")
	}
	if len(queues) != 1 || queues[0].Name != "queue1" {
		t.Errorf("expected [queue1], got %v", queues)
	}
}

func TestAWSUseCase_ListSNSTopics(t *testing.T) {
	expected := []domain.SNSTopic{{ARN: "arn:aws:sns:test", Name: "test-topic"}}
	sns := &mockSNSManager{
		ListTopicsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
			return expected, nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	topics, err := uc.ListSNSTopics(context.Background(), &domain.AWSConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) != 1 || topics[0].Name != "test-topic" {
		t.Errorf("expected [test-topic], got %v", topics)
	}
}

func TestAWSUseCase_PublishSNSMessage(t *testing.T) {
	var called bool
	sns := &mockSNSManager{
		PublishMessageFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, body, subject string, attrs map[string]string) error {
			called = true
			if topicARN != "arn:test" {
				t.Errorf("expected topicARN 'arn:test', got %q", topicARN)
			}
			return nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	err := uc.PublishSNSMessage(context.Background(), &domain.AWSConfig{}, "arn:test", "body", "subject", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected PublishMessage to be called")
	}
}

func TestAWSUseCase_SendSQSMessage(t *testing.T) {
	var called bool
	sqs := &mockSQSManager{
		SendMessageFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL, body string) error {
			called = true
			return nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, sqs, &mockSNSManager{}, &mockSecretsManager{})
	err := uc.SendSQSMessage(context.Background(), &domain.AWSConfig{}, "url", "body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected SendMessage to be called")
	}
}

func TestAWSUseCase_DeleteS3Object(t *testing.T) {
	var called bool
	s3 := &mockS3Manager{
		DeleteObjectFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) error {
			called = true
			return nil
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	err := uc.DeleteS3Object(context.Background(), &domain.AWSConfig{}, "bucket", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DeleteObject to be called")
	}
}

func TestAWSUseCase_GetS3PresignedURL(t *testing.T) {
	s3 := &mockS3Manager{
		GetPresignedURLFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) (string, error) {
			return "https://presigned.url/test", nil
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	url, err := uc.GetS3PresignedURL(context.Background(), &domain.AWSConfig{}, "bucket", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://presigned.url/test" {
		t.Errorf("expected 'https://presigned.url/test', got %q", url)
	}
}

func TestAWSUseCase_DeleteS3Bucket(t *testing.T) {
	var called bool
	s3 := &mockS3Manager{
		DeleteBucketFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket string) error {
			called = true
			if bucket != "test-bucket" {
				t.Errorf("expected bucket 'test-bucket', got %q", bucket)
			}
			return nil
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	err := uc.DeleteS3Bucket(context.Background(), &domain.AWSConfig{}, "test-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DeleteBucket to be called")
	}
}

func TestAWSUseCase_CreateS3Bucket(t *testing.T) {
	var called bool
	s3 := &mockS3Manager{
		CreateBucketFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket string) error {
			called = true
			if bucket != "new-bucket" {
				t.Errorf("expected bucket 'new-bucket', got %q", bucket)
			}
			return nil
		},
	}

	uc := NewAWSUseCase(s3, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})
	err := uc.CreateS3Bucket(context.Background(), &domain.AWSConfig{}, "new-bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected CreateBucket to be called")
	}
}

func TestAWSUseCase_CreateSQSQueue(t *testing.T) {
	var called bool
	sqs := &mockSQSManager{
		CreateQueueFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueName string) (string, error) {
			called = true
			if queueName != "new-queue" {
				t.Errorf("expected queue 'new-queue', got %q", queueName)
			}
			return "http://sqs/new-queue", nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, sqs, &mockSNSManager{}, &mockSecretsManager{})
	url, err := uc.CreateSQSQueue(context.Background(), &domain.AWSConfig{}, "new-queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "http://sqs/new-queue" {
		t.Errorf("expected url 'http://sqs/new-queue', got %q", url)
	}
	if !called {
		t.Error("expected CreateQueue to be called")
	}
}

func TestAWSUseCase_CreateSNSSubscription_ResolvesQueueARN(t *testing.T) {
	var resolvedARN bool
	var resolvedURL bool
	var policySet string
	sqs := &mockSQSManager{
		ResolveQueueURLFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueRef string) (string, error) {
			resolvedURL = true
			if queueRef != "http://localhost:4566/000000000000/orders-queue" {
				t.Fatalf("unexpected queue ref: %s", queueRef)
			}
			return queueRef, nil
		},
		ResolveQueueARNFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string) (string, error) {
			resolvedARN = true
			if queueURL != "http://localhost:4566/000000000000/orders-queue" {
				t.Fatalf("unexpected queue url: %s", queueURL)
			}
			return "arn:aws:sqs:us-east-1:000000000000:orders-queue", nil
		},
		GetQueueAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
			if queueURL != "http://localhost:4566/000000000000/orders-queue" {
				t.Fatalf("unexpected queue url for policy lookup: %s", queueURL)
			}
			return map[string]string{}, nil
		},
		SetQueueAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error {
			if queueURL != "http://localhost:4566/000000000000/orders-queue" {
				t.Fatalf("unexpected queue url for policy update: %s", queueURL)
			}
			policySet = attributes["Policy"]
			return nil
		},
	}
	var receivedEndpoint string
	sns := &mockSNSManager{
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			receivedEndpoint = endpoint
			return domain.SNSSubscription{ARN: "sub-1", Endpoint: endpoint, Protocol: protocol, TopicARN: topicARN}, nil
		},
	}

	uc := NewAWSUseCase(&mockS3Manager{}, sqs, sns, &mockSecretsManager{})
	_, err := uc.CreateSNSSubscription(context.Background(), &domain.AWSConfig{}, "arn:aws:sns:us-east-1:000000000000:orders", "sqs", "http://localhost:4566/000000000000/orders-queue", nil, domain.SubscriptionFilterScopeMessageAttributes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resolvedARN || !resolvedURL {
		t.Fatal("expected queue ARN and URL resolution to be called")
	}
	if receivedEndpoint != "arn:aws:sqs:us-east-1:000000000000:orders-queue" {
		t.Fatalf("expected resolved ARN endpoint, got %s", receivedEndpoint)
	}
	if !strings.Contains(policySet, "sns.amazonaws.com") {
		t.Fatalf("expected queue policy to be applied, got %s", policySet)
	}
}

func TestAWSUseCase_CreateSNSSubscription_RejectsTopicToTopic(t *testing.T) {
	uc := NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, &mockSNSManager{}, &mockSecretsManager{})

	_, err := uc.CreateSNSSubscription(
		context.Background(),
		&domain.AWSConfig{},
		"arn:aws:sns:us-east-1:000000000000:accounts",
		"sns",
		"arn:aws:sns:us-east-1:000000000000:billings",
		nil,
		domain.SubscriptionFilterScopeMessageAttributes,
	)
	if err == nil {
		t.Fatal("expected error for sns topic-to-topic subscription")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}
