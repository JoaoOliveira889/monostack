package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"monostack/internal/domain"
)

const snapshotFileName = "monostack-snapshot.yaml"

const maxSnapshotSizeWarningBytes = 10 * 1024 * 1024

type SnapshotUseCase struct {
	aws    *AWSUseCase
	config *ConfigUseCase
}

func NewSnapshotUseCase(aws *AWSUseCase, config *ConfigUseCase) *SnapshotUseCase {
	return &SnapshotUseCase{aws: aws, config: config}
}

func (uc *SnapshotUseCase) EstimateExportSize(ctx context.Context, cfg *domain.AWSConfig) (int64, error) {
	buckets, err := uc.aws.ListS3Buckets(ctx, cfg)
	if err != nil {
		return 0, err
	}
	var totalSize int64
	for _, bucket := range buckets {
		objects, err := uc.aws.ListS3Objects(ctx, cfg, bucket.Name, "")
		if err != nil {
			continue
		}
		for _, object := range objects {
			totalSize += (object.Size * 4) / 3
			totalSize += 64
		}
	}
	return totalSize, nil
}

func (uc *SnapshotUseCase) ShouldWarnSize(estimatedBytes int64) bool {
	return estimatedBytes > maxSnapshotSizeWarningBytes
}

func (uc *SnapshotUseCase) Export(ctx context.Context, rawPath string) (string, error) {
	cfg, err := uc.config.GetConfig()
	if err != nil {
		return "", err
	}

	snapshot, err := uc.collectSnapshot(ctx, cfg)
	if err != nil {
		return "", err
	}

	return uc.writeSnapshot(ctx, rawPath, cfg, snapshot)
}

func (uc *SnapshotUseCase) ExportS3Bucket(ctx context.Context, cfg *domain.AWSConfig, bucketName string) (*domain.AppProfile, error) {
	snapshot := &domain.AppProfile{
		Version: 2,
		Config:  redactedConfig(cfg),
	}

	subs, _ := uc.config.LoadSubscriptions()
	snapshot.Subscriptions = subs

	objects, err := uc.aws.ListS3Objects(ctx, cfg, bucketName, "")
	if err != nil {
		return nil, err
	}

	bucketSnapshot := domain.S3BucketSnapshot{Name: bucketName}
	for _, object := range objects {
		content, err := uc.downloadS3ObjectBytes(ctx, cfg, bucketName, object.Key)
		if err != nil {
			return nil, fmt.Errorf("bucket %s object %s: %w", bucketName, object.Key, err)
		}
		contentType, metadata, _ := uc.aws.HeadS3Object(ctx, cfg, bucketName, object.Key)
		bucketSnapshot.Objects = append(bucketSnapshot.Objects, domain.S3ObjectSnapshot{
			Key:           object.Key,
			Size:          object.Size,
			LastModified:  object.LastModified,
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			ContentType:   contentType,
			Metadata:      metadata,
		})
	}
	snapshot.S3 = append(snapshot.S3, bucketSnapshot)
	return snapshot, nil
}

func (uc *SnapshotUseCase) ExportSQSQueue(ctx context.Context, cfg *domain.AWSConfig, queueName string) (*domain.AppProfile, error) {
	snapshot := &domain.AppProfile{
		Version: 2,
		Config:  redactedConfig(cfg),
	}

	subs, _ := uc.config.LoadSubscriptions()
	snapshot.Subscriptions = subs

	queues, err := uc.aws.ListSQSQueues(ctx, cfg)
	if err != nil {
		return nil, err
	}

	for _, queue := range queues {
		if queue.Name != queueName {
			continue
		}
		attrs, err := uc.aws.GetSQSQueueAttributes(ctx, cfg, queue.URL, []string{"All"})
		if err != nil {
			attrs = map[string]string{}
		}
		snapshot.SQS = append(snapshot.SQS, domain.SQSQueueSnapshot{
			Name:       queue.Name,
			URL:        queue.URL,
			ARN:        queue.ARN,
			Attributes: attrs,
		})
		break
	}

	if len(snapshot.SQS) == 0 {
		return nil, fmt.Errorf("queue %q not found", queueName)
	}

	return snapshot, nil
}

func (uc *SnapshotUseCase) ExportSNSTopic(ctx context.Context, cfg *domain.AWSConfig, topicARN string) (*domain.AppProfile, error) {
	snapshot := &domain.AppProfile{
		Version: 2,
		Config:  redactedConfig(cfg),
	}

	subs, _ := uc.config.LoadSubscriptions()
	snapshot.Subscriptions = subs

	topics, err := uc.aws.ListSNSTopics(ctx, cfg)
	if err != nil {
		return nil, err
	}

	for _, topic := range topics {
		if topic.ARN != topicARN {
			continue
		}
		topicSubs, err := uc.aws.ListSNSSubscriptions(ctx, cfg, topic.ARN)
		if err != nil {
			topicSubs = nil
		}
		snapshot.SNS = append(snapshot.SNS, domain.SNSTopicSnapshot{
			Name:          topic.Name,
			ARN:           topic.ARN,
			Subscriptions: topicSubs,
		})
		break
	}

	if len(snapshot.SNS) == 0 {
		return nil, fmt.Errorf("topic %q not found", topicARN)
	}

	return snapshot, nil
}

func (uc *SnapshotUseCase) ExportSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) (*domain.AppProfile, error) {
	snapshot := &domain.AppProfile{
		Version: 2,
		Config:  redactedConfig(cfg),
	}

	subs, _ := uc.config.LoadSubscriptions()
	snapshot.Subscriptions = subs

	secrets, err := uc.aws.ListSecrets(ctx, cfg)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.ARN != secretID {
			continue
		}
		secretSnapshot := domain.SecretSnapshot{
			Name:        secret.Name,
			Description: secret.Description,
			KMSKeyID:    secret.KMSKeyID,
		}
		val, valErr := uc.aws.GetSecretValue(ctx, cfg, secret.ARN, "", "")
		if valErr == nil {
			secretSnapshot.SecretString = val.SecretString
			secretSnapshot.SecretBinaryB64 = val.SecretBinaryBase64
		}
		snapshot.Secrets = append(snapshot.Secrets, secretSnapshot)
		break
	}

	if len(snapshot.Secrets) == 0 {
		return nil, fmt.Errorf("secret %q not found", secretID)
	}

	return snapshot, nil
}

func (uc *SnapshotUseCase) ExportSingleResourceToPath(ctx context.Context, rawPath string, cfg *domain.AWSConfig, snapshot *domain.AppProfile) (string, error) {
	return uc.writeSnapshot(ctx, rawPath, cfg, snapshot)
}

func (uc *SnapshotUseCase) writeSnapshot(ctx context.Context, rawPath string, cfg *domain.AWSConfig, snapshot *domain.AppProfile) (string, error) {
	targetPath, err := resolveSnapshotTarget(rawPath, cfg)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
		return "", fmt.Errorf("cannot create snapshot folder: %w", err)
	}

	data, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		return "", fmt.Errorf("failed to write snapshot: %w", err)
	}

	return targetPath, nil
}

func (uc *SnapshotUseCase) Import(ctx context.Context, rawPath string) (*domain.AppProfile, error) {
	targetPath, err := resolveSnapshotPathForRead(rawPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snapshot domain.AppProfile
	if err := yaml.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	cfg := snapshot.Config
	if cfg == nil {
		cfg, err = uc.config.GetConfig()
		if err != nil {
			return nil, err
		}
	}

	if snapshot.Version == 0 {
		snapshot.Version = 2
	}

	if err := uc.applySnapshot(ctx, cfg, &snapshot); err != nil {
		return nil, err
	}

	if err := uc.config.SaveConfig(cfg); err != nil {
		return nil, err
	}

	snapshot.Config = cfg
	return &snapshot, nil
}

func redactedConfig(cfg *domain.AWSConfig) *domain.AWSConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	cloned.SecretAccessKey = ""
	cloned.AccessKeyID = ""
	cloned.PanelRatios = nil
	cloned.SnapshotPath = ""
	return &cloned
}

func (uc *SnapshotUseCase) collectSnapshot(ctx context.Context, cfg *domain.AWSConfig) (*domain.AppProfile, error) {
	snapshot := &domain.AppProfile{
		Version: 2,
		Config:  redactedConfig(cfg),
	}

	subs, err := uc.config.LoadSubscriptions()
	if err != nil {
		return nil, err
	}
	snapshot.Subscriptions = subs

	if err := uc.collectS3Snapshot(ctx, cfg, snapshot); err != nil {
		return nil, err
	}
	if err := uc.collectSQSSnapshot(ctx, cfg, snapshot); err != nil {
		return nil, err
	}
	if err := uc.collectSNSSnapshot(ctx, cfg, snapshot); err != nil {
		return nil, err
	}
	if err := uc.collectSecretsSnapshot(ctx, cfg, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (uc *SnapshotUseCase) collectS3Snapshot(ctx context.Context, cfg *domain.AWSConfig, snapshot *domain.AppProfile) error {
	buckets, err := uc.aws.ListS3Buckets(ctx, cfg)
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		objects, err := uc.aws.ListS3Objects(ctx, cfg, bucket.Name, "")
		if err != nil {
			return err
		}

		bucketSnapshot := domain.S3BucketSnapshot{Name: bucket.Name}
		for _, object := range objects {
			content, err := uc.downloadS3ObjectBytes(ctx, cfg, bucket.Name, object.Key)
			if err != nil {
				return fmt.Errorf("bucket %s object %s: %w", bucket.Name, object.Key, err)
			}

			contentType, metadata, _ := uc.aws.HeadS3Object(ctx, cfg, bucket.Name, object.Key)

			bucketSnapshot.Objects = append(bucketSnapshot.Objects, domain.S3ObjectSnapshot{
				Key:           object.Key,
				Size:          object.Size,
				LastModified:  object.LastModified,
				ContentBase64: base64.StdEncoding.EncodeToString(content),
				ContentType:   contentType,
				Metadata:      metadata,
			})
		}
		snapshot.S3 = append(snapshot.S3, bucketSnapshot)
	}

	return nil
}

func (uc *SnapshotUseCase) collectSQSSnapshot(ctx context.Context, cfg *domain.AWSConfig, snapshot *domain.AppProfile) error {
	queues, err := uc.aws.ListSQSQueues(ctx, cfg)
	if err != nil {
		return err
	}

	for _, queue := range queues {
		attrs, err := uc.aws.GetSQSQueueAttributes(ctx, cfg, queue.URL, []string{"All"})
		if err != nil {
			attrs = map[string]string{}
		}
		snapshot.SQS = append(snapshot.SQS, domain.SQSQueueSnapshot{
			Name:       queue.Name,
			URL:        queue.URL,
			ARN:        queue.ARN,
			Attributes: attrs,
		})
	}

	return nil
}

func (uc *SnapshotUseCase) collectSNSSnapshot(ctx context.Context, cfg *domain.AWSConfig, snapshot *domain.AppProfile) error {
	topics, err := uc.aws.ListSNSTopics(ctx, cfg)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		subs, err := uc.aws.ListSNSSubscriptions(ctx, cfg, topic.ARN)
		if err != nil {
			return err
		}
		snapshot.SNS = append(snapshot.SNS, domain.SNSTopicSnapshot{
			Name:          topic.Name,
			ARN:           topic.ARN,
			Subscriptions: subs,
		})
	}

	return nil
}

func (uc *SnapshotUseCase) collectSecretsSnapshot(ctx context.Context, cfg *domain.AWSConfig, snapshot *domain.AppProfile) error {
	secrets, err := uc.aws.ListSecrets(ctx, cfg)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		secretSnapshot := domain.SecretSnapshot{
			Name:        secret.Name,
			Description: secret.Description,
			KMSKeyID:    secret.KMSKeyID,
		}

		val, valErr := uc.aws.GetSecretValue(ctx, cfg, secret.ARN, "", "")
		if valErr == nil {
			secretSnapshot.SecretString = val.SecretString
			secretSnapshot.SecretBinaryB64 = val.SecretBinaryBase64
		}

		snapshot.Secrets = append(snapshot.Secrets, secretSnapshot)
	}

	return nil
}

func (uc *SnapshotUseCase) applySnapshot(ctx context.Context, cfg *domain.AWSConfig, snapshot *domain.AppProfile) error {
	if err := uc.applyS3Snapshot(ctx, cfg, snapshot.S3); err != nil {
		return err
	}

	queueMap, err := uc.applySQSSnapshot(ctx, cfg, snapshot.SQS)
	if err != nil {
		return err
	}

	topicMap, err := uc.applySNSTopics(ctx, cfg, snapshot.SNS)
	if err != nil {
		return err
	}

	existingSubs, _ := uc.aws.ListAllSNSSubscriptions(ctx, cfg)
	existingIndex := make(map[string]domain.SNSSubscription)
	for _, sub := range existingSubs {
		existingIndex[snsSubscriptionKey(sub.TopicARN, sub.Protocol, normalizeSubscriptionEndpoint(sub.Protocol, sub.Endpoint, queueMap))] = sub
	}

	createdIndex := make(map[string]string)
	for _, topicSnapshot := range snapshot.SNS {
		snapshotTopicARN := topicSnapshot.ARN
		sourceARN := topicMap[snapshotTopicARN]
		if sourceARN == "" {
			sourceARN = snapshotTopicARN
		}
		for _, sub := range topicSnapshot.Subscriptions {
			destEndpoint := normalizeSubscriptionEndpoint(sub.Protocol, sub.Endpoint, queueMap)
			actualKey := snsSubscriptionKey(sourceARN, sub.Protocol, destEndpoint)
			snapshotKey := snsSubscriptionKey(snapshotTopicARN, sub.Protocol, destEndpoint)
			if existing, ok := existingIndex[actualKey]; ok && existing.ARN != "" {
				createdIndex[snapshotKey] = existing.ARN
				continue
			}

			created, err := uc.aws.CreateSNSSubscription(ctx, cfg, sourceARN, sub.Protocol, destEndpoint, sub.FilterPolicy, domain.NormalizeFilterScope(sub.FilterScope))
			if err != nil {
				return fmt.Errorf("failed to recreate subscription for %s: %w", sourceARN, err)
			}
			createdIndex[snapshotKey] = created.ARN
		}
	}

	importedSubs := append([]domain.ManagedSubscription(nil), snapshot.Subscriptions...)
	for i := range importedSubs {
		normalizedDestination := normalizeManagedDestinationARN(importedSubs[i].DestinationARN, queueMap)
		key := managedSubscriptionKey(importedSubs[i].TopicARN, normalizedDestination, importedSubs[i].DestinationType, queueMap)
		if arn := createdIndex[key]; arn != "" {
			importedSubs[i].SubscriptionARN = arn
			continue
		}
		if existing, ok := existingIndex[snsSubscriptionKey(importedSubs[i].TopicARN, destinationTypeToProtocol(importedSubs[i].DestinationType), normalizeSubscriptionEndpoint(destinationTypeToProtocol(importedSubs[i].DestinationType), normalizedDestination, queueMap))]; ok {
			importedSubs[i].SubscriptionARN = existing.ARN
		}
	}

	if err := uc.config.SaveSubscriptions(importedSubs); err != nil {
		return err
	}

	if err := uc.applySecretsSnapshot(ctx, cfg, snapshot.Secrets); err != nil {
		return err
	}

	return nil
}

func (uc *SnapshotUseCase) applyS3Snapshot(ctx context.Context, cfg *domain.AWSConfig, buckets []domain.S3BucketSnapshot) error {
	existingBuckets, err := uc.aws.ListS3Buckets(ctx, cfg)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{}, len(existingBuckets))
	for _, bucket := range existingBuckets {
		existing[bucket.Name] = struct{}{}
	}

	for _, bucket := range buckets {
		if _, ok := existing[bucket.Name]; !ok {
			if err := uc.aws.CreateS3Bucket(ctx, cfg, bucket.Name); err != nil {
				return err
			}
		}

		for _, object := range bucket.Objects {
			content, err := base64.StdEncoding.DecodeString(object.ContentBase64)
			if err != nil {
				return fmt.Errorf("bucket %s object %s: invalid base64: %w", bucket.Name, object.Key, err)
			}

			tmpFile, err := os.CreateTemp("", "monostack-s3-*")
			if err != nil {
				return err
			}
			tmpPath := tmpFile.Name()
			tmpFile.Close()

			if err := os.WriteFile(tmpPath, content, 0o600); err != nil {
				_ = os.Remove(tmpPath)
				return err
			}
			if len(object.Metadata) > 0 {
				err = uc.aws.UploadS3ObjectWithMetadata(ctx, cfg, bucket.Name, object.Key, tmpPath, object.Metadata)
			} else {
				err = uc.aws.UploadS3Object(ctx, cfg, bucket.Name, object.Key, tmpPath)
			}
			if err != nil {
				_ = os.Remove(tmpPath)
				return err
			}
			_ = os.Remove(tmpPath)
		}
	}

	return nil
}

func (uc *SnapshotUseCase) applySecretsSnapshot(ctx context.Context, cfg *domain.AWSConfig, secrets []domain.SecretSnapshot) error {
	existingSecrets, err := uc.aws.ListSecrets(ctx, cfg)
	if err != nil {
		return err
	}

	existingByName := make(map[string]struct{}, len(existingSecrets))
	for _, s := range existingSecrets {
		existingByName[s.Name] = struct{}{}
	}

	for _, s := range secrets {
		if _, ok := existingByName[s.Name]; ok {
			continue
		}
		secretValue := s.SecretString
		if secretValue == "" && s.SecretBinaryB64 != "" {
			secretValue = s.SecretBinaryB64
		}
		if secretValue == "" {
			secretValue = "placeholder-value"
		}
		if _, err := uc.aws.CreateSecret(ctx, cfg, s.Name, secretValue, s.Description); err != nil {
			return fmt.Errorf("failed to recreate secret %q: %w", s.Name, err)
		}
	}

	return nil
}

func (uc *SnapshotUseCase) applySQSSnapshot(ctx context.Context, cfg *domain.AWSConfig, queues []domain.SQSQueueSnapshot) (map[string]domain.SQSQueueSnapshot, error) {
	existingQueues, err := uc.aws.ListSQSQueues(ctx, cfg)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]domain.SQSQueue, len(existingQueues))
	for _, queue := range existingQueues {
		existingByName[queue.Name] = queue
	}

	resolved := make(map[string]domain.SQSQueueSnapshot)
	for _, queue := range queues {
		if existing, ok := existingByName[queue.Name]; ok {
			resolved[queue.ARN] = domain.SQSQueueSnapshot{
				Name:       existing.Name,
				URL:        existing.URL,
				ARN:        existing.ARN,
				Attributes: queue.Attributes,
			}
			if len(queue.Attributes) > 0 {
				if err := uc.aws.SetSQSQueueAttributes(ctx, cfg, existing.URL, queue.Attributes); err != nil {
					return nil, err
				}
			}
			continue
		}

		queueURL, err := uc.aws.CreateSQSQueue(ctx, cfg, queue.Name)
		if err != nil {
			return nil, err
		}
		queueARN := queue.ARN
		if queueARN == "" {
			queueARN = domain.QueueARNFromURL(queueURL, cfg.Region)
		}
		resolved[queueARN] = domain.SQSQueueSnapshot{
			Name:       queue.Name,
			URL:        queueURL,
			ARN:        queueARN,
			Attributes: queue.Attributes,
		}

		if len(queue.Attributes) > 0 {
			if err := uc.aws.SetSQSQueueAttributes(ctx, cfg, queueURL, queue.Attributes); err != nil {
				return nil, err
			}
		}
	}

	return resolved, nil
}

func (uc *SnapshotUseCase) applySNSTopics(ctx context.Context, cfg *domain.AWSConfig, topics []domain.SNSTopicSnapshot) (map[string]string, error) {
	existingTopics, err := uc.aws.ListSNSTopics(ctx, cfg)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]domain.SNSTopic, len(existingTopics))
	for _, topic := range existingTopics {
		existingByName[topic.Name] = topic
	}

	resolved := make(map[string]string)
	for _, topic := range topics {
		if existing, ok := existingByName[topic.Name]; ok {
			resolved[topic.ARN] = existing.ARN
			continue
		}

		created, err := uc.aws.CreateSNSTopic(ctx, cfg, topic.Name)
		if err != nil {
			return nil, err
		}
		resolved[topic.ARN] = created.ARN
	}

	return resolved, nil
}

func (uc *SnapshotUseCase) downloadS3ObjectBytes(ctx context.Context, cfg *domain.AWSConfig, bucket, key string) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "monostack-s3-export-*")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := uc.aws.DownloadS3Object(ctx, cfg, bucket, key, tmpPath); err != nil {
		return nil, err
	}
	return os.ReadFile(tmpPath)
}

func resolveSnapshotTarget(rawPath string, cfg *domain.AWSConfig) (string, error) {
	path := strings.TrimSpace(rawPath)
	if path == "" && cfg != nil {
		path = strings.TrimSpace(cfg.SnapshotPath)
	}
	if path == "" {
		path = defaultSnapshotPath()
	}

	return normalizeSnapshotPath(path), nil
}

func resolveSnapshotPathForRead(rawPath string) (string, error) {
	path := strings.TrimSpace(rawPath)
	if path == "" {
		return "", fmt.Errorf("snapshot path is required")
	}
	return normalizeSnapshotPath(path), nil
}

func defaultSnapshotPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "monostack", snapshotFileName)
}

func DefaultSnapshotPath() string {
	return defaultSnapshotPath()
}

func normalizeSnapshotPath(path string) string {
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return filepath.Join(path, snapshotFileName)
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, snapshotFileName)
	}
	if filepath.Ext(path) == "" && !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		return filepath.Join(path, snapshotFileName)
	}
	return path
}

func snsSubscriptionKey(topicARN, protocol, endpoint string) string {
	return strings.Join([]string{topicARN, protocol, endpoint}, "|")
}

func managedSubscriptionKey(topicARN, destinationARN, destinationType string, queueMap map[string]domain.SQSQueueSnapshot) string {
	return strings.Join([]string{topicARN, destinationTypeToProtocol(destinationType), normalizeSubscriptionEndpoint(destinationTypeToProtocol(destinationType), destinationARN, queueMap)}, "|")
}

func normalizeSubscriptionEndpoint(protocol, endpoint string, queueMap map[string]domain.SQSQueueSnapshot) string {
	endpoint = strings.TrimSpace(endpoint)
	if protocol == "sqs" {
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
			if arn := domain.QueueARNFromURL(endpoint, ""); arn != "" {
				return arn
			}
		}
		if queueMap != nil {
			if snapshot, ok := queueMap[endpoint]; ok && snapshot.ARN != "" {
				return snapshot.ARN
			}
			for _, snapshot := range queueMap {
				if snapshot.URL == endpoint && snapshot.ARN != "" {
					return snapshot.ARN
				}
			}
		}
	}
	return endpoint
}

func normalizeManagedDestinationARN(destinationARN string, queueMap map[string]domain.SQSQueueSnapshot) string {
	return normalizeSubscriptionEndpoint("sqs", destinationARN, queueMap)
}

func destinationTypeToProtocol(destinationType string) string {
	switch strings.ToLower(strings.TrimSpace(destinationType)) {
	case "sqs", "sns":
		return strings.ToLower(strings.TrimSpace(destinationType))
	default:
		return "sns"
	}
}
