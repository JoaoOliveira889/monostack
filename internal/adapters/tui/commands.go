package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"monostack/internal/domain"
)

const transientMessageDuration = 3 * time.Second

func (m *Model) loadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := m.configUseCase.GetConfig()
		if err != nil {
			return errMsg{Error: err}
		}
		return configLoadedMsg{Config: cfg}
	}
}

func (m *Model) saveConfigCmd(cfg *domain.AWSConfig) tea.Cmd {
	return func() tea.Msg {
		err := m.configUseCase.SaveConfig(cfg)
		if err != nil {
			return errMsg{Error: err}
		}
		return configSavedMsg{Config: cfg}
	}
}

func splitCSVList(value string) []string {
	var items []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func normalizeSubscriptionScope(scope string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "", domain.SubscriptionFilterScopeMessageBody:
		return domain.SubscriptionFilterScopeMessageBody, nil
	case domain.SubscriptionFilterScopeMessageAttributes:
		return domain.SubscriptionFilterScopeMessageAttributes, nil
	default:
		return "", fmt.Errorf("invalid filter_scope %q", scope)
	}
}

func normalizeSubscriptionScopeWithDefault(scope, defaultScope string) (string, error) {
	defaultScope = strings.TrimSpace(defaultScope)
	if defaultScope == "" {
		defaultScope = domain.SubscriptionFilterScopeMessageBody
	}
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "":
		return normalizeSubscriptionScope(defaultScope)
	case domain.SubscriptionFilterScopeMessageAttributes, domain.SubscriptionFilterScopeMessageBody:
		return normalizeSubscriptionScope(scope)
	default:
		return "", fmt.Errorf("invalid filter_scope %q", scope)
	}
}

func snsSubscriptionKey(topicARN, protocol, endpoint string) string {
	return strings.Join([]string{topicARN, protocol, endpoint}, "|")
}

func managedSubscriptionKey(topicARN, destinationARN, destinationType string, queueMap map[string]string) string {
	return strings.Join([]string{topicARN, destinationType, normalizeSubscriptionEndpoint(destinationType, destinationARN, queueMap)}, "|")
}

func normalizeSubscriptionEndpoint(destinationType, endpoint string, queueMap map[string]string) string {
	endpoint = strings.TrimSpace(endpoint)
	if strings.EqualFold(strings.TrimSpace(destinationType), "sqs") {
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
			if arn := domain.QueueARNFromURL(endpoint, ""); arn != "" {
				return arn
			}
		}
		if queueMap != nil {
			if arn, ok := queueMap[endpoint]; ok && arn != "" {
				return arn
			}
			for _, arn := range queueMap {
				if arn == endpoint {
					return arn
				}
			}
		}
	}
	return endpoint
}

func topicScriptFileName(topicARN string) (string, error) {
	trimmed := strings.TrimSpace(topicARN)
	if trimmed == "" {
		return "", fmt.Errorf("topic arn is required")
	}
	safe := strings.NewReplacer(":", "_", "/", "_").Replace(trimmed)
	return safe + ".yaml", nil
}

func yamlScriptPathForTopic(topicARN string) (string, error) {
	fileName, err := topicScriptFileName(topicARN)
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home dir: %w", err)
	}
	return filepath.Join(home, ".config", "monostack", "subscription_scripts", fileName), nil
}

func inferredQueueRefFromTopic(topicName string) string {
	trimmed := strings.TrimSpace(topicName)
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, "-sns") {
		return strings.TrimSuffix(trimmed, "-sns") + "-sqs"
	}
	return strings.TrimSuffix(trimmed, "-topics") + "-sqs"
}

func resolveYamlSubscriptionTopic(entryTopic, currentTopicARN string, topicMap map[string]string) (string, error) {
	currentTopicARN = strings.TrimSpace(currentTopicARN)
	if entryTopic != "" {
		srcARN, ok := topicMap[strings.TrimSpace(entryTopic)]
		if !ok {
			return "", fmt.Errorf("topic %q not found in LocalStack", entryTopic)
		}
		return srcARN, nil
	}
	if currentTopicARN == "" {
		return "", fmt.Errorf("topic is required when the subscription script is not tied to an active SNS topic")
	}
	return currentTopicARN, nil
}

func resolveYamlSubscriptionQueue(entryQueue, defaultQueue, inferredQueue string, queueMap map[string]string) (string, string, error) {
	queueRef := strings.TrimSpace(entryQueue)
	if queueRef == "" {
		queueRef = strings.TrimSpace(defaultQueue)
	}
	if queueRef == "" {
		queueRef = strings.TrimSpace(inferredQueue)
	}
	if queueRef == "" {
		return "", "", fmt.Errorf("queue is required (set queue, default_queue, or use a topic name that maps to a queue)")
	}
	if strings.HasPrefix(queueRef, "arn:aws:sqs:") {
		return queueRef, queueRef, nil
	}
	if endpoint, ok := queueMap[queueRef]; ok && endpoint != "" {
		return queueRef, endpoint, nil
	}
	return "", "", fmt.Errorf("queue %q not found in the loaded SQS queues", queueRef)
}

func clearStatusCmd(id int) tea.Cmd {
	return tea.Tick(transientMessageDuration, func(time.Time) tea.Msg {
		return clearStatusMsg{id: id}
	})
}

func (m *Model) setStatusMessage(message string) tea.Cmd {
	m.statusMsg = message
	m.errorMsg = ""
	m.statusMsgID++
	return clearStatusCmd(m.statusMsgID)
}

func (m *Model) setErrorMessage(message string) tea.Cmd {
	m.errorMsg = message
	m.statusMsg = ""
	m.statusMsgID++
	return clearStatusCmd(m.statusMsgID)
}

func (m *Model) loadS3BucketsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		buckets, err := m.awsUseCase.ListS3Buckets(ctx, m.config)
		if err != nil {
			return errMsg{Error: err}
		}
		return s3BucketsLoadedMsg{Buckets: buckets}
	}
}

func (m *Model) loadS3ObjectsCmd(bucket string) tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.s3ObjectsCache[bucket]; ok {
			return s3ObjectsLoadedMsg{Bucket: bucket, Objects: cached}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		objects, err := m.awsUseCase.ListS3Objects(ctx, m.config, bucket, "")
		if err != nil {
			return errMsg{Error: err}
		}
		return s3ObjectsLoadedMsg{Bucket: bucket, Objects: objects}
	}
}

func (m *Model) deleteS3ObjectCmd(bucket, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteS3Object(ctx, m.config, bucket, key)
		if err != nil {
			return errMsg{Error: err}
		}
		return s3ObjectDeletedMsg{Bucket: bucket, Key: key}
	}
}

func (m *Model) deleteS3BucketCmd(bucket string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteS3Bucket(ctx, m.config, bucket)
		if err != nil {
			return errMsg{Error: err}
		}
		return s3BucketDeletedMsg{Bucket: bucket}
	}
}

func (m *Model) createS3BucketCmd(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.CreateS3Bucket(ctx, m.config, name)
		if err != nil {
			return errMsg{Error: err}
		}
		return s3BucketCreatedMsg{Bucket: name}
	}
}

func (m *Model) downloadS3ObjectCmd(bucket, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		destDir := filepath.Join(home, "Downloads", "monostack")
		destPath := filepath.Join(destDir, filepath.Base(key))

		err = m.awsUseCase.DownloadS3Object(ctx, m.config, bucket, key, destPath)
		if err != nil {
			return errMsg{Error: err}
		}
		return s3ObjectDownloadedMsg{DestPath: destPath}
	}
}

func (m *Model) uploadS3ObjectCmd(bucket, filePath, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := m.awsUseCase.UploadS3Object(ctx, m.config, bucket, key, filePath); err != nil {
			return errMsg{Error: err}
		}
		return s3ObjectUploadedMsg{Bucket: bucket, Key: key}
	}
}

func (m *Model) loadSQSQueuesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		queues, err := m.awsUseCase.ListSQSQueues(ctx, m.config)
		if err != nil {
			return errMsg{Error: err}
		}
		allSubs, err := m.awsUseCase.ListAllSNSSubscriptions(ctx, m.config)
		if err != nil {
			allSubs = nil
		} else if allSubs == nil {
			allSubs = []domain.SNSSubscription{}
		}
		return sqsQueuesLoadedMsg{Queues: queues, AllSubscriptions: allSubs}
	}
}

func (m *Model) purgeSQSQueueCmd(queueURL string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.awsUseCase.PurgeSQSQueue(ctx, m.config, queueURL)
		if err != nil {
			return errMsg{Error: err}
		}
		return sqsQueuePurgedMsg{QueueURL: queueURL}
	}
}

func (m *Model) deleteSQSQueueCmd(queueURL, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteSQSQueue(ctx, m.config, queueURL)
		if err != nil {
			return errMsg{Error: err}
		}
		return sqsQueueDeletedMsg{QueueURL: queueURL, Name: name}
	}
}

func (m *Model) createSQSQueueCmd(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := m.awsUseCase.CreateSQSQueue(ctx, m.config, name)
		if err != nil {
			return errMsg{Error: err}
		}
		return sqsQueueCreatedMsg{Name: name}
	}
}

func (m *Model) peekSQSMessagesCmd(queueURL string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		messages, err := m.awsUseCase.ReceiveSQSMessages(ctx, m.config, queueURL, 5)
		if err != nil {
			return errMsg{Error: err}
		}
		return sqsMessagesLoadedMsg{QueueURL: queueURL, Messages: messages}
	}
}

func (m *Model) sendSQSMessageCmd(queueURL, body string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.SendSQSMessage(ctx, m.config, queueURL, body)
		if err != nil {
			return errMsg{Error: err}
		}
		return sqsMessageSentMsg{QueueURL: queueURL}
	}
}

func (m *Model) loadSNSTopicsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		topics, err := m.awsUseCase.ListSNSTopics(ctx, m.config)
		if err != nil {
			return errMsg{Error: err}
		}
		allSubs, err := m.awsUseCase.ListAllSNSSubscriptions(ctx, m.config)
		if err != nil {
			allSubs = nil
		} else if allSubs == nil {
			allSubs = []domain.SNSSubscription{}
		}
		return snsTopicsLoadedMsg{Topics: topics, AllSubscriptions: allSubs}
	}
}

func (m *Model) loadSecretsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		secrets, err := m.awsUseCase.ListSecrets(ctx, m.config)
		if err != nil {
			return errMsg{Error: err}
		}
		return secretsLoadedMsg{Secrets: secrets}
	}
}

func (m *Model) loadSecretDetailsCmd(secretID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		secret, err := m.awsUseCase.DescribeSecret(ctx, m.config, secretID)
		if err != nil {
			return errMsg{Error: err}
		}
		versions, err := m.awsUseCase.ListSecretVersions(ctx, m.config, secretID)
		if err != nil {
			versions = nil
		}
		value, err := m.awsUseCase.GetSecretValue(ctx, m.config, secretID, "", "AWSCURRENT")
		if err != nil {
			value = domain.SecretValue{}
		}
		return secretDetailsLoadedMsg{Secret: secret, Versions: versions, Value: value}
	}
}

func (m *Model) createSecretCmd(name, value, description string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		secret, err := m.awsUseCase.CreateSecret(ctx, m.config, name, value, description)
		if err != nil {
			return errMsg{Error: err}
		}
		return secretCreatedMsg{Secret: secret}
	}
}

func (m *Model) updateSecretValueCmd(secretID, value, description string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		secretValue, err := m.awsUseCase.UpdateSecretValue(ctx, m.config, secretID, value, description)
		if err != nil {
			return errMsg{Error: err}
		}
		return secretValueUpdatedMsg{SecretID: secretID, Value: secretValue}
	}
}

func (m Model) currentSecretVersionID() string {
	for _, version := range m.secretVersions {
		for _, stage := range version.Stages {
			if stage == "AWSCURRENT" {
				return version.VersionID
			}
		}
	}
	return ""
}

func (m *Model) promoteSecretVersionCmd(secretID, versionID string) tea.Cmd {
	currentVersionID := m.currentSecretVersionID()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		if err := m.awsUseCase.UpdateSecretVersionStage(ctx, m.config, secretID, "AWSCURRENT", versionID, currentVersionID); err != nil {
			return errMsg{Error: err}
		}
		return secretStageUpdatedMsg{SecretID: secretID, VersionID: versionID}
	}
}

func (m *Model) deleteSecretCmd(secretID, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteSecret(ctx, m.config, secretID, 7, false)
		if err != nil {
			return errMsg{Error: err}
		}
		return secretDeletedMsg{SecretID: secretID, Name: name}
	}
}

func (m *Model) restoreSecretCmd(secretID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.RestoreSecret(ctx, m.config, secretID)
		if err != nil {
			return errMsg{Error: err}
		}
		return secretRestoredMsg{SecretID: secretID}
	}
}

func (m *Model) publishSNSMessageCmd(topicARN, body string, attrs map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.PublishSNSMessage(ctx, m.config, topicARN, body, "Published from Monostack TUI", attrs)
		if err != nil {
			return errMsg{Error: err}
		}
		return snsMessagePublishedMsg{TopicARN: topicARN}
	}
}

func (m *Model) createSNSTopicCmd(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		topic, err := m.awsUseCase.CreateSNSTopic(ctx, m.config, name)
		if err != nil {
			return errMsg{Error: err}
		}
		return snsTopicCreatedMsg{Topic: topic}
	}
}

func (m *Model) deleteSNSTopicCmd(topicARN string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteSNSTopic(ctx, m.config, topicARN)
		if err != nil {
			return errMsg{Error: err}
		}
		return snsTopicDeletedMsg{ARN: topicARN}
	}
}

func (m *Model) loadSNSSubscriptionsCmd(topicARN string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		outgoing, err := m.awsUseCase.ListSNSSubscriptions(ctx, m.config, topicARN)
		if err != nil {
			return errMsg{Error: err}
		}

		allSubs, err := m.awsUseCase.ListAllSNSSubscriptions(ctx, m.config)
		if err != nil {
			return snsSubscriptionsLoadedMsg{Subscriptions: outgoing, IncomingSubscriptions: nil}
		}
		if allSubs == nil {
			allSubs = []domain.SNSSubscription{}
		}

		var incoming []domain.SNSSubscription
		for _, sub := range allSubs {
			if sub.Endpoint == topicARN && sub.TopicARN != topicARN {
				incoming = append(incoming, sub)
			}
		}

		return snsSubscriptionsLoadedMsg{Subscriptions: outgoing, IncomingSubscriptions: incoming, AllSubscriptions: allSubs}
	}
}

func (m *Model) loadManagedSubscriptionsCmd() tea.Cmd {
	return func() tea.Msg {
		subs, err := m.configUseCase.LoadSubscriptions()
		if err != nil {
			return errMsg{Error: err}
		}
		if subs == nil {
			subs = []domain.ManagedSubscription{}
		}

		needsCleanup := false
		for i := range subs {
			if subs[i].Name == "sqs-batch" || subs[i].Name == "batch" {
				if idx := strings.LastIndex(subs[i].TopicARN, ":"); idx >= 0 {
					subs[i].Name = subs[i].TopicARN[idx+1:]
					needsCleanup = true
				}
			}
		}
		if needsCleanup {
			_ = m.configUseCase.SaveSubscriptions(subs)
		}
		return managedSubscriptionsLoadedMsg{Subscriptions: subs}
	}
}

func (m *Model) saveManagedSubscriptionsCmd(subs []domain.ManagedSubscription) tea.Cmd {
	return func() tea.Msg {
		err := m.configUseCase.SaveSubscriptions(subs)
		if err != nil {
			return errMsg{Error: err}
		}
		return managedSubscriptionsUpdatedMsg{}
	}
}

func (m *Model) createSNSSubscriptionCmd(topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		sub, err := m.awsUseCase.CreateSNSSubscription(ctx, m.config, topicARN, protocol, endpoint, filterPolicy, filterScope)
		if err != nil {
			return errMsg{Error: err}
		}
		return snsSubscriptionCreatedMsg{Subscription: sub}
	}
}

func (m *Model) deleteSNSSubscriptionCmd(subARN string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		err := m.awsUseCase.DeleteSNSSubscription(ctx, m.config, subARN)
		if err != nil {
			return errMsg{Error: err}
		}
		return snsSubscriptionDeletedMsg{ARN: subARN}
	}
}

func (m *Model) updateSNSSubscriptionCmd(subARN string, filterPolicy map[string][]string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		fpJSON, err := json.Marshal(filterPolicy)
		if err != nil {
			return errMsg{Error: err}
		}

		err = m.awsUseCase.SetSNSSubscriptionAttributes(ctx, m.config, subARN, "FilterPolicy", string(fpJSON))
		if err != nil {
			return errMsg{Error: err}
		}
		return snsSubscriptionUpdatedMsg{ARN: subARN}
	}
}

func (m *Model) batchSubscribeSNSCmd(topics []toggleOption, destARN string, eventTypes []string, filterScope string) tea.Cmd {
	return func() tea.Msg {
		_ = topics
		_ = destARN
		_ = eventTypes
		_ = filterScope
		return errMsg{Error: fmt.Errorf("sns topic-to-topic subscriptions are not supported; subscribe the billings SQS directly to the source SNS topics")}
	}
}

func (m *Model) importSubscriptionsYamlContentCmd(yamlContent string, currentTopicARN string, topics []domain.SNSTopic, queues []domain.SQSQueue) tea.Cmd {
	return func() tea.Msg {
		var script domain.SubscriptionScript
		if err := yaml.Unmarshal([]byte(yamlContent), &script); err != nil {
			return errMsg{Error: fmt.Errorf("failed to parse YAML: %w", err)}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		topicMap := make(map[string]string)
		for _, t := range topics {
			topicMap[t.Name] = t.ARN
		}
		queueMap := make(map[string]string)
		for _, q := range queues {
			queueMap[q.Name] = q.ARN
		}

		existingSubs, err := m.awsUseCase.ListAllSNSSubscriptions(ctx, m.config)
		if err != nil {
			return errMsg{Error: fmt.Errorf("failed to inspect existing subscriptions: %w", err)}
		}
		existingIndex := make(map[string]domain.SNSSubscription, len(existingSubs))
		for _, sub := range existingSubs {
			key := snsSubscriptionKey(sub.TopicARN, sub.Protocol, normalizeSubscriptionEndpoint(sub.Protocol, sub.Endpoint, queueMap))
			existingIndex[key] = sub
		}

		currentSubs, _ := m.configUseCase.LoadSubscriptions()
		managedByKey := make(map[string]int, len(currentSubs))
		for i, sub := range currentSubs {
			key := managedSubscriptionKey(sub.TopicARN, sub.DestinationARN, sub.DestinationType, queueMap)
			managedByKey[key] = i
		}

		var createdCount int
		var repairedCount int
		var unchangedCount int

		for _, entry := range script.Subscriptions {
			srcARN, err := resolveYamlSubscriptionTopic(entry.Topic, currentTopicARN, topicMap)
			if err != nil {
				return errMsg{Error: err}
			}
			srcTopicName := shortResourceName(srcARN)

			var filterPolicy map[string][]string
			if len(entry.EventType) > 0 {
				filterPolicy = map[string][]string{"event_type": entry.EventType}
			}
			normalizedScope, err := normalizeSubscriptionScopeWithDefault(entry.FilterScope, script.DefaultFilterScope)
			if err != nil {
				return errMsg{Error: fmt.Errorf("subscription %q: %w", entry.Name, err)}
			}

			_, endpoint, err := resolveYamlSubscriptionQueue(entry.Queue, script.DefaultQueue, inferredQueueRefFromTopic(srcTopicName), queueMap)
			if err != nil {
				return errMsg{Error: fmt.Errorf("subscription %q: %w", entry.Name, err)}
			}

			protocol := "sqs"
			key := snsSubscriptionKey(srcARN, protocol, endpoint)
			desiredManaged := domain.ManagedSubscription{
				Name:            entry.Name,
				TopicARN:        srcARN,
				DestinationARN:  endpoint,
				DestinationType: "sqs",
				EventTypes:      entry.EventType,
				FilterScope:     normalizedScope,
			}

			if existing, ok := existingIndex[key]; ok && existing.ARN != "" {
				routeRepaired := false
				if !filterPolicyEqual(existing.FilterPolicy, filterPolicy) {
					if filterPolicy == nil {

					} else {
						fpJSON, marshalErr := json.Marshal(filterPolicy)
						if marshalErr != nil {
							return errMsg{Error: marshalErr}
						}
						if err := m.awsUseCase.SetSNSSubscriptionAttributes(ctx, m.config, existing.ARN, "FilterPolicy", string(fpJSON)); err != nil {
							return errMsg{Error: fmt.Errorf("failed to update filter policy for %s: %w", entry.Topic, err)}
						}
						routeRepaired = true
					}
				}

				if !strings.EqualFold(existing.FilterScope, normalizedScope) {
					scopeValue := "MessageAttributes"
					if normalizedScope == domain.SubscriptionFilterScopeMessageBody {
						scopeValue = "MessageBody"
					}
					if err := m.awsUseCase.SetSNSSubscriptionAttributes(ctx, m.config, existing.ARN, "FilterPolicyScope", scopeValue); err != nil {
						return errMsg{Error: fmt.Errorf("failed to update filter scope for %s: %w", srcTopicName, err)}
					}
					routeRepaired = true
				}

				if routeRepaired {
					repairedCount++
				} else {
					unchangedCount++
				}
				desiredManaged.SubscriptionARN = existing.ARN
			} else {
				sub, err := m.awsUseCase.CreateSNSSubscription(ctx, m.config, srcARN, protocol, endpoint, filterPolicy, normalizedScope)
				if err != nil {
					return errMsg{Error: fmt.Errorf("failed to subscribe to %q: %w", srcTopicName, err)}
				}
				desiredManaged.SubscriptionARN = sub.ARN
				createdCount++
			}

			managedKey := managedSubscriptionKey(desiredManaged.TopicARN, desiredManaged.DestinationARN, desiredManaged.DestinationType, nil)
			if idx, ok := managedByKey[managedKey]; ok {
				currentSubs[idx] = desiredManaged
			} else {
				managedByKey[managedKey] = len(currentSubs)
				currentSubs = append(currentSubs, desiredManaged)
			}
		}

		_ = m.configUseCase.SaveSubscriptions(currentSubs)

		return snsYamlImportAppliedMsg{Created: createdCount, Repaired: repairedCount, Unchanged: unchangedCount}
	}
}

func filterPolicyEqual(a, b map[string][]string) bool {
	return reflect.DeepEqual(a, b)
}

func (m *Model) saveYamlScriptCmd(topicARN string, content string) tea.Cmd {
	return func() tea.Msg {
		path, err := yamlScriptPathForTopic(topicARN)
		if err != nil {
			return errMsg{Error: err}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return errMsg{Error: err}
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return errMsg{Error: err}
		}
		return yamlScriptSavedMsg{TopicARN: topicARN}
	}
}

func (m *Model) loadYamlScriptCmd(topicARN string) tea.Cmd {
	return func() tea.Msg {
		path, err := yamlScriptPathForTopic(topicARN)
		if err != nil {
			return yamlScriptLoadedMsg{TopicARN: topicARN, Content: ""}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return yamlScriptLoadedMsg{TopicARN: topicARN, Content: ""}
		}
		return yamlScriptLoadedMsg{TopicARN: topicARN, Content: string(data)}
	}
}

func (m *Model) exportProfileCmd(rawPath string) tea.Cmd {
	return func() tea.Msg {
		if m.snapshotUseCase == nil {
			return errMsg{Error: fmt.Errorf("snapshot usecase is not configured")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		path, err := m.snapshotUseCase.Export(ctx, expandPath(rawPath))
		if err != nil {
			return errMsg{Error: err}
		}
		return profileExportedMsg{Path: path}
	}
}

func (m *Model) importProfileCmd(rawPath string) tea.Cmd {
	return func() tea.Msg {
		if m.snapshotUseCase == nil {
			return errMsg{Error: fmt.Errorf("snapshot usecase is not configured")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		profile, err := m.snapshotUseCase.Import(ctx, expandPath(rawPath))
		if err != nil {
			return errMsg{Error: err}
		}

		return profileImportedMsg{
			Config:    profile.Config,
			SubsCount: len(profile.Subscriptions),
			Path:      expandPath(rawPath),
		}
	}
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func defaultS3ObjectKey(filePath string) string {
	trimmed := strings.TrimSpace(filePath)
	if trimmed == "" {
		return ""
	}
	return path.Base(expandPath(trimmed))
}

func openURL(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func (m *Model) loadSQSQueueSubscriptionsCmd(queueURL string, queueARN string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		allSubs, err := m.awsUseCase.ListAllSNSSubscriptions(ctx, m.config)
		if err != nil {
			return errMsg{Error: fmt.Errorf("failed to list subscriptions: %w", err)}
		}

		var filtered []domain.SNSSubscription
		for _, sub := range allSubs {
			if sub.Endpoint == queueURL || sub.Endpoint == queueARN {
				filtered = append(filtered, sub)
			}
		}
		return sqsSubscriptionsLoadedMsg{Subscriptions: filtered}
	}
}

func (m *Model) batchSubscribeSQSCmd(topics []toggleOption, queueARN string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var createdSubs []domain.SNSSubscription
		var managedSubs []domain.ManagedSubscription

		for _, t := range topics {
			sub, err := m.awsUseCase.CreateSNSSubscription(ctx, m.config, t.arn, "sqs", queueARN, nil, domain.SubscriptionFilterScopeMessageAttributes)
			if err != nil {
				return errMsg{Error: fmt.Errorf("failed to subscribe %s: %w", t.arn, err)}
			}
			createdSubs = append(createdSubs, sub)
			managedSubs = append(managedSubs, domain.ManagedSubscription{
				Name:            t.label,
				TopicARN:        t.arn,
				DestinationARN:  queueARN,
				DestinationType: "sqs",
				SubscriptionARN: sub.ARN,
			})
		}

		currentSubs, _ := m.configUseCase.LoadSubscriptions()
		currentSubs = append(currentSubs, managedSubs...)
		_ = m.configUseCase.SaveSubscriptions(currentSubs)

		return sqsBatchSubscriptionsCreatedMsg{Count: len(createdSubs)}
	}
}
