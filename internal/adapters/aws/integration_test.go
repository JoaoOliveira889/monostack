//go:build integration
// +build integration

package aws

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"monostack/internal/domain"
)

func integrationConfig() *domain.AWSConfig {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "test"
	}
	if secretKey == "" {
		secretKey = "test"
	}
	return &domain.AWSConfig{
		EndpointURL:     endpoint,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		ServiceName:     "integration-test",
	}
}

func TestIntegration_S3Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := integrationConfig()
	adapter := NewS3Adapter()

	bucketName := "monostack-integration-bucket-" + time.Now().Format("20060102150405")

	if err := adapter.CreateBucket(ctx, cfg, bucketName); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = adapter.DeleteBucket(ctx, cfg, bucketName)
	})

	buckets, err := adapter.ListBuckets(ctx, cfg)
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}
	found := false
	for _, b := range buckets {
		if b.Name == bucketName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bucket %q not found in list", bucketName)
	}

	tmpFile, err := os.CreateTemp("", "monostack-integration-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString("integration test content"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	objectKey := "test/integration.txt"
	if err := adapter.UploadObject(ctx, cfg, bucketName, objectKey, tmpFile.Name()); err != nil {
		t.Fatalf("UploadObject failed: %v", err)
	}

	objects, err := adapter.ListObjects(ctx, cfg, bucketName, "")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}
	found = false
	for _, o := range objects {
		if o.Key == objectKey {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("object %q not found in bucket %q", objectKey, bucketName)
	}

	destDir := os.TempDir()
	destPath := destDir + "/monostack-integration-download.txt"
	defer os.Remove(destPath)
	if err := adapter.DownloadObject(ctx, cfg, bucketName, objectKey, destPath); err != nil {
		t.Fatalf("DownloadObject failed: %v", err)
	}
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != "integration test content" {
		t.Fatalf("unexpected content: %q", string(data))
	}

	presignedURL, err := adapter.GetPresignedURL(ctx, cfg, bucketName, objectKey)
	if err != nil {
		t.Fatalf("GetPresignedURL failed: %v", err)
	}
	if !strings.Contains(presignedURL, bucketName) {
		t.Fatalf("unexpected presigned URL: %s", presignedURL)
	}

	if err := adapter.CreateFolder(ctx, cfg, bucketName, "reports/"); err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	if err := adapter.DeleteObject(ctx, cfg, bucketName, objectKey); err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}
}

func TestIntegration_SQSLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := integrationConfig()
	adapter := NewSQSAdapter()

	queueName := "monostack-integration-queue-" + time.Now().Format("20060102150405")

	queueURL, err := adapter.CreateQueue(ctx, cfg, queueName)
	if err != nil {
		t.Fatalf("CreateQueue failed: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = adapter.DeleteQueue(ctx, cfg, queueURL)
	})

	queues, err := adapter.ListQueues(ctx, cfg)
	if err != nil {
		t.Fatalf("ListQueues failed: %v", err)
	}
	found := false
	for _, q := range queues {
		if q.Name == queueName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("queue %q not found in list", queueName)
	}

	if err := adapter.SendMessage(ctx, cfg, queueURL, `{"event":"integration_test"}`); err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	messages, err := adapter.ReceiveMessages(ctx, cfg, queueURL, 5)
	if err != nil {
		t.Fatalf("ReceiveMessages failed: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}

	if err := adapter.PurgeQueue(ctx, cfg, queueURL); err != nil {
		t.Fatalf("PurgeQueue failed: %v", err)
	}

	attrs, err := adapter.GetQueueAttributes(ctx, cfg, queueURL, []string{"QueueArn"})
	if err != nil {
		t.Fatalf("GetQueueAttributes failed: %v", err)
	}
	if attrs["QueueArn"] == "" {
		t.Fatal("expected QueueArn attribute")
	}
}

func TestIntegration_SNSLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := integrationConfig()
	adapter := NewSNSAdapter()

	topicName := "monostack-integration-topic-" + time.Now().Format("20060102150405")

	topic, err := adapter.CreateTopic(ctx, cfg, topicName)
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = adapter.DeleteTopic(ctx, cfg, topic.ARN)
	})

	topics, err := adapter.ListTopics(ctx, cfg)
	if err != nil {
		t.Fatalf("ListTopics failed: %v", err)
	}
	found := false
	for _, t := range topics {
		if t.Name == topicName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("topic %q not found in list", topicName)
	}

	if err := adapter.PublishMessage(ctx, cfg, topic.ARN, "integration test message", "TestSubject", map[string]string{"event_type": "test"}); err != nil {
		t.Fatalf("PublishMessage failed: %v", err)
	}

	subs, err := adapter.ListSubscriptions(ctx, cfg, topic.ARN)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	_ = subs
}

func TestIntegration_SecretsLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := integrationConfig()
	adapter := NewSecretsAdapter()

	secretName := "monostack-integration-secret-" + time.Now().Format("20060102150405")

	secret, err := adapter.CreateSecret(ctx, cfg, secretName, `{"api_key":"test_value"}`, "integration test secret")
	if err != nil {
		t.Fatalf("CreateSecret failed: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = adapter.DeleteSecret(ctx, cfg, secret.ARN, 0, true)
	})

	secrets, err := adapter.ListSecrets(ctx, cfg)
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}
	found := false
	for _, s := range secrets {
		if s.Name == secretName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("secret %q not found in list", secretName)
	}

	described, err := adapter.DescribeSecret(ctx, cfg, secret.ARN)
	if err != nil {
		t.Fatalf("DescribeSecret failed: %v", err)
	}
	if described.Description != "integration test secret" {
		t.Fatalf("unexpected description: %q", described.Description)
	}

	value, err := adapter.GetSecretValue(ctx, cfg, secret.ARN, "", "AWSCURRENT")
	if err != nil {
		t.Fatalf("GetSecretValue failed: %v", err)
	}
	if value.SecretString != `{"api_key":"test_value"}` {
		t.Fatalf("unexpected secret value: %q", value.SecretString)
	}

	updated, err := adapter.UpdateSecretValue(ctx, cfg, secret.ARN, `{"api_key":"updated_value"}`, "updated description")
	if err != nil {
		t.Fatalf("UpdateSecretValue failed: %v", err)
	}
	_ = updated

	versions, err := adapter.ListSecretVersions(ctx, cfg, secret.ARN)
	if err != nil {
		t.Fatalf("ListSecretVersions failed: %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("expected at least 2 versions, got %d", len(versions))
	}

	if err := adapter.DeleteSecret(ctx, cfg, secret.ARN, 0, true); err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}
}
