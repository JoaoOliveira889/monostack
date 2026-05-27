package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"monostack/internal/domain"
)

type mockSubscriptionStore struct {
	LoadFunc func() ([]domain.ManagedSubscription, error)
	SaveFunc func([]domain.ManagedSubscription) error
}

func (m *mockSubscriptionStore) LoadAll() ([]domain.ManagedSubscription, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc()
	}
	return nil, nil
}

func (m *mockSubscriptionStore) SaveAll(subs []domain.ManagedSubscription) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(subs)
	}
	return nil
}

func TestSnapshotUseCase_Export(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &domain.AWSConfig{
		ServiceName:    "MiniStack",
		EndpointURL:    "http://localhost:4566",
		Region:         "us-east-1",
		SnapshotPath:   tempDir,
		LeftPanelRatio: 0.3,
	}

	cfgStore := &mockConfigStore{
		LoadFunc: func() (*domain.AWSConfig, error) {
			return cfg, nil
		},
	}

	subStore := &mockSubscriptionStore{
		LoadFunc: func() ([]domain.ManagedSubscription, error) {
			return []domain.ManagedSubscription{
				{
					Name:            "orders",
					TopicARN:        "arn:aws:sns:us-east-1:123456789012:orders",
					DestinationARN:  "arn:aws:sqs:us-east-1:123456789012:orders-queue",
					DestinationType: "sqs",
				},
			}, nil
		},
	}

	s3 := &mockS3Manager{
		ListBucketsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
			return []domain.S3Bucket{{Name: "assets"}}, nil
		},
		ListObjectsFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket, prefix string) ([]domain.S3Object, error) {
			return []domain.S3Object{{Key: "images/logo.png", Size: 5, LastModified: "2025-01-01T00:00:00Z"}}, nil
		},
		DownloadObjectFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, destPath string) error {
			return os.WriteFile(destPath, []byte("hello"), 0o600)
		},
	}

	sqs := &mockSQSManager{
		ListQueuesFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
			return []domain.SQSQueue{{Name: "orders-queue", URL: "http://sqs.us-east-1.amazonaws.com/123456789012/orders-queue", ARN: "arn:aws:sqs:us-east-1:123456789012:orders-queue"}}, nil
		},
		GetQueueAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
			return map[string]string{"VisibilityTimeout": "30"}, nil
		},
	}

	sns := &mockSNSManager{
		ListTopicsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
			return []domain.SNSTopic{{ARN: "arn:aws:sns:us-east-1:123456789012:orders", Name: "orders"}}, nil
		},
		ListSubscriptionsFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN string) ([]domain.SNSSubscription, error) {
			return []domain.SNSSubscription{
				{
					ARN:      "arn:aws:sns:us-east-1:123456789012:orders:sub-1",
					TopicARN: topicARN,
					Protocol: "sqs",
					Endpoint: "arn:aws:sqs:us-east-1:123456789012:orders-queue",
				},
			}, nil
		},
	}

	awsUC := NewAWSUseCase(s3, sqs, sns, &mockSecretsManager{})
	cfgUC := NewConfigUseCaseWithSubscriptions(cfgStore, subStore)
	uc := NewSnapshotUseCase(awsUC, cfgUC)

	path, err := uc.Export(context.Background(), "")
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if filepath.Dir(path) != tempDir {
		t.Fatalf("expected snapshot in %s, got %s", tempDir, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	var exported domain.AppProfile
	if err := yaml.Unmarshal(data, &exported); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(exported.S3) != 1 || len(exported.S3[0].Objects) != 1 {
		t.Fatalf("expected one exported S3 bucket/object, got %#v", exported.S3)
	}
	if exported.S3[0].Objects[0].ContentBase64 == "" {
		t.Fatal("expected object content to be exported")
	}
	if len(exported.SQS) != 1 || len(exported.SNS) != 1 {
		t.Fatalf("expected SQS/SNS data in snapshot, got %#v %#v", exported.SQS, exported.SNS)
	}
}

func TestSnapshotUseCase_Import(t *testing.T) {
	tempDir := t.TempDir()
	snapshotPath := filepath.Join(tempDir, "snapshot.yaml")

	profile := domain.AppProfile{
		Version: 2,
		Config: &domain.AWSConfig{
			ServiceName:    "MiniStack",
			EndpointURL:    "http://localhost:4566",
			Region:         "us-east-1",
			SnapshotPath:   tempDir,
			LeftPanelRatio: 0.4,
		},
		Subscriptions: []domain.ManagedSubscription{
			{
				Name:            "orders",
				TopicARN:        "arn:aws:sns:us-east-1:123456789012:orders",
				DestinationARN:  "arn:aws:sqs:us-east-1:123456789012:orders-queue",
				DestinationType: "sqs",
			},
		},
		S3: []domain.S3BucketSnapshot{
			{
				Name: "assets",
				Objects: []domain.S3ObjectSnapshot{
					{Key: "images/logo.png", ContentBase64: "aGVsbG8="},
				},
			},
		},
		SQS: []domain.SQSQueueSnapshot{
			{
				Name:       "orders-queue",
				URL:        "http://sqs.us-east-1.amazonaws.com/123456789012/orders-queue",
				ARN:        "arn:aws:sqs:us-east-1:123456789012:orders-queue",
				Attributes: map[string]string{"VisibilityTimeout": "45"},
			},
		},
		SNS: []domain.SNSTopicSnapshot{
			{
				Name: "orders",
				ARN:  "arn:aws:sns:us-east-1:123456789012:orders",
				Subscriptions: []domain.SNSSubscription{
					{
						TopicARN:    "arn:aws:sns:us-east-1:123456789012:orders",
						Protocol:    "sqs",
						Endpoint:    "arn:aws:sqs:us-east-1:123456789012:orders-queue",
						FilterScope: domain.SubscriptionFilterScopeMessageAttributes,
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if err := os.WriteFile(snapshotPath, data, 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	var savedCfg *domain.AWSConfig
	var savedSubs []domain.ManagedSubscription
	var queuePolicy string
	createBucketCalled := false
	createQueueCalled := false
	createTopicCalled := false
	createSubscriptionCalled := false

	cfgStore := &mockConfigStore{
		LoadFunc: func() (*domain.AWSConfig, error) {
			return profile.Config, nil
		},
		SaveFunc: func(cfg *domain.AWSConfig) error {
			savedCfg = cfg
			return nil
		},
	}

	subStore := &mockSubscriptionStore{
		SaveFunc: func(subs []domain.ManagedSubscription) error {
			savedSubs = append([]domain.ManagedSubscription(nil), subs...)
			return nil
		},
	}

	s3 := &mockS3Manager{
		ListBucketsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
			return nil, nil
		},
		CreateBucketFunc: func(ctx context.Context, cfg *domain.AWSConfig, name string) error {
			createBucketCalled = true
			if name != "assets" {
				t.Fatalf("unexpected bucket %q", name)
			}
			return nil
		},
		UploadObjectFunc: func(ctx context.Context, cfg *domain.AWSConfig, bucket, key, filePath string) error {
			if bucket != "assets" || key != "images/logo.png" {
				t.Fatalf("unexpected upload target %s/%s", bucket, key)
			}
			content, readErr := os.ReadFile(filePath)
			if readErr != nil {
				return readErr
			}
			if string(content) != "hello" {
				t.Fatalf("unexpected upload content %q", string(content))
			}
			return nil
		},
	}

	sqs := &mockSQSManager{
		ListQueuesFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
			return nil, nil
		},
		CreateQueueFunc: func(ctx context.Context, cfg *domain.AWSConfig, name string) (string, error) {
			createQueueCalled = true
			if name != "orders-queue" {
				t.Fatalf("unexpected queue %q", name)
			}
			return "http://sqs.us-east-1.amazonaws.com/123456789012/orders-queue", nil
		},
		ResolveQueueURLFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueRef string) (string, error) {
			if queueRef != "arn:aws:sqs:us-east-1:123456789012:orders-queue" {
				t.Fatalf("unexpected queue ref for url resolution %q", queueRef)
			}
			return "http://sqs.us-east-1.amazonaws.com/123456789012/orders-queue", nil
		},
		SetQueueAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error {
			if queueURL != "http://sqs.us-east-1.amazonaws.com/123456789012/orders-queue" {
				t.Fatalf("unexpected queue URL %q", queueURL)
			}
			if policy, ok := attributes["Policy"]; ok {
				queuePolicy = policy
			}
			return nil
		},
		GetQueueAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
			return map[string]string{}, nil
		},
	}

	sns := &mockSNSManager{
		ListTopicsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
			return nil, nil
		},
		CreateTopicFunc: func(ctx context.Context, cfg *domain.AWSConfig, name string) (domain.SNSTopic, error) {
			createTopicCalled = true
			if name != "orders" {
				t.Fatalf("unexpected topic %q", name)
			}
			return domain.SNSTopic{
				ARN:  "arn:aws:sns:us-east-1:123456789012:orders",
				Name: name,
			}, nil
		},
		ListAllSubscriptionsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error) {
			return nil, nil
		},
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			createSubscriptionCalled = true
			if topicARN != "arn:aws:sns:us-east-1:123456789012:orders" {
				t.Fatalf("unexpected topicARN %q", topicARN)
			}
			if protocol != "sqs" || endpoint != "arn:aws:sqs:us-east-1:123456789012:orders-queue" {
				t.Fatalf("unexpected subscription endpoint %s %s", protocol, endpoint)
			}
			return domain.SNSSubscription{
				ARN:      "arn:aws:sns:us-east-1:123456789012:orders:sub-1",
				TopicARN: topicARN,
				Protocol: protocol,
				Endpoint: endpoint,
			}, nil
		},
	}

	awsUC := NewAWSUseCase(s3, sqs, sns, &mockSecretsManager{})
	cfgUC := NewConfigUseCaseWithSubscriptions(cfgStore, subStore)
	uc := NewSnapshotUseCase(awsUC, cfgUC)

	imported, err := uc.Import(context.Background(), snapshotPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if imported.Config == nil || imported.Config.ServiceName != "MiniStack" {
		t.Fatalf("unexpected imported config %#v", imported.Config)
	}
	if !createBucketCalled || !createQueueCalled || !createTopicCalled || !createSubscriptionCalled {
		t.Fatal("expected all provisioning calls to be made")
	}
	if savedCfg == nil || savedCfg.EndpointURL != profile.Config.EndpointURL {
		t.Fatalf("expected config to be saved, got %#v", savedCfg)
	}
	if len(savedSubs) != 1 || savedSubs[0].SubscriptionARN == "" {
		t.Fatalf("expected managed subscriptions to be saved with ARNs, got %#v", savedSubs)
	}
	if !strings.Contains(queuePolicy, "sns.amazonaws.com") {
		t.Fatalf("expected queue policy to allow SNS publish, got %s", queuePolicy)
	}
}
