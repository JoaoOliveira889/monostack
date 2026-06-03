package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"monostack/internal/usecase"

	"monostack/internal/domain"
)

func testCmd(t *testing.T, cmd tea.Cmd, expectedMsgType string) {
	t.Helper()
	if cmd == nil {
		t.Fatalf("expected non-nil command, got nil")
	}
	msg := cmd()
	if msg == nil {
		t.Fatalf("expected non-nil message from command")
	}
	switch msg.(type) {
	case configLoadedMsg:
		if expectedMsgType != "configLoadedMsg" {
			t.Errorf("expected %s, got configLoadedMsg", expectedMsgType)
		}
	case configSavedMsg:
		if expectedMsgType != "configSavedMsg" {
			t.Errorf("expected %s, got configSavedMsg", expectedMsgType)
		}
	case s3BucketsLoadedMsg:
		if expectedMsgType != "s3BucketsLoadedMsg" {
			t.Errorf("expected %s, got s3BucketsLoadedMsg", expectedMsgType)
		}
	case s3ObjectsLoadedMsg:
		if expectedMsgType != "s3ObjectsLoadedMsg" {
			t.Errorf("expected %s, got s3ObjectsLoadedMsg", expectedMsgType)
		}
	case s3ObjectDeletedMsg:
		if expectedMsgType != "s3ObjectDeletedMsg" {
			t.Errorf("expected %s, got s3ObjectDeletedMsg", expectedMsgType)
		}
	case s3ObjectDownloadedMsg:
		if expectedMsgType != "s3ObjectDownloadedMsg" {
			t.Errorf("expected %s, got s3ObjectDownloadedMsg", expectedMsgType)
		}
	case s3FolderCreatedMsg:
		if expectedMsgType != "s3FolderCreatedMsg" {
			t.Errorf("expected %s, got s3FolderCreatedMsg", expectedMsgType)
		}
	case sqsQueuesLoadedMsg:
		if expectedMsgType != "sqsQueuesLoadedMsg" {
			t.Errorf("expected %s, got sqsQueuesLoadedMsg", expectedMsgType)
		}
	case sqsQueuePurgedMsg:
		if expectedMsgType != "sqsQueuePurgedMsg" {
			t.Errorf("expected %s, got sqsQueuePurgedMsg", expectedMsgType)
		}
	case sqsQueuesPurgedMsg:
		if expectedMsgType != "sqsQueuesPurgedMsg" {
			t.Errorf("expected %s, got sqsQueuesPurgedMsg", expectedMsgType)
		}
	case sqsMessagesLoadedMsg:
		if expectedMsgType != "sqsMessagesLoadedMsg" {
			t.Errorf("expected %s, got sqsMessagesLoadedMsg", expectedMsgType)
		}
	case snsTopicsLoadedMsg:
		if expectedMsgType != "snsTopicsLoadedMsg" {
			t.Errorf("expected %s, got snsTopicsLoadedMsg", expectedMsgType)
		}
	case statusMsg:
		if expectedMsgType != "statusMsg" {
			t.Errorf("expected %s, got statusMsg", expectedMsgType)
		}
	case errMsg:
		if expectedMsgType != "errMsg" {
			t.Errorf("expected %s, got errMsg", expectedMsgType)
		}
	default:
		t.Errorf("unexpected message type: %T", msg)
	}
}

func TestSplitCSVList(t *testing.T) {
	got := splitCSVList("pix_received, pix_sent, , pix_key_validate")
	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got))
	}
	if got[0] != "pix_received" || got[2] != "pix_key_validate" {
		t.Fatalf("unexpected split result: %#v", got)
	}
}

func TestNormalizeFilterScopeStrict(t *testing.T) {
	scope, err := domain.NormalizeFilterScopeStrict("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scope != domain.SubscriptionFilterScopeMessageAttributes {
		t.Fatalf("expected default scope %q, got %q", domain.SubscriptionFilterScopeMessageAttributes, scope)
	}

	scope, err = domain.NormalizeFilterScopeStrict(domain.SubscriptionFilterScopeMessageAttributes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scope != domain.SubscriptionFilterScopeMessageAttributes {
		t.Fatalf("expected scope %q, got %q", domain.SubscriptionFilterScopeMessageAttributes, scope)
	}
}

func TestYamlScriptCommandsArePerTopic(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	model := mkModel()

	firstSave := model.saveYamlScriptCmd("arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns", "pix-yaml")
	if msg := firstSave(); msg == nil {
		t.Fatal("expected save message for first topic")
	}
	secondSave := model.saveYamlScriptCmd("arn:aws:sns:us-east-1:000000000000:dev-webapi-notifications-sns", "notifications-yaml")
	if msg := secondSave(); msg == nil {
		t.Fatal("expected save message for second topic")
	}

	firstPath, err := yamlScriptPathForTopic("arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns")
	if err != nil {
		t.Fatalf("unexpected path error: %v", err)
	}
	secondPath, err := yamlScriptPathForTopic("arn:aws:sns:us-east-1:000000000000:dev-webapi-notifications-sns")
	if err != nil {
		t.Fatalf("unexpected path error: %v", err)
	}
	if firstPath == secondPath {
		t.Fatalf("expected distinct paths per topic, got %q", firstPath)
	}

	if data, err := os.ReadFile(firstPath); err != nil || string(data) != "pix-yaml" {
		t.Fatalf("unexpected first topic content: %q, %v", string(data), err)
	}
	if info, err := os.Stat(filepath.Dir(firstPath)); err != nil {
		t.Fatalf("expected yaml directory to exist: %v", err)
	} else if perm := info.Mode().Perm(); perm != 0o750 {
		t.Fatalf("expected yaml directory permissions 0750, got %o", perm)
	}
	if info, err := os.Stat(firstPath); err != nil {
		t.Fatalf("expected yaml file to exist: %v", err)
	} else if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected yaml file permissions 0600, got %o", perm)
	}
	if data, err := os.ReadFile(secondPath); err != nil || string(data) != "notifications-yaml" {
		t.Fatalf("unexpected second topic content: %q, %v", string(data), err)
	}

	load := model.loadYamlScriptCmd("arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns")
	msg := load()
	loaded, ok := msg.(yamlScriptLoadedMsg)
	if !ok {
		t.Fatalf("expected yamlScriptLoadedMsg, got %T", msg)
	}
	if loaded.Content != "pix-yaml" {
		t.Fatalf("unexpected loaded content: %q", loaded.Content)
	}

	if base := filepath.Base(firstPath); base == "" {
		t.Fatal("expected non-empty yaml path base")
	}
}

func TestImportSubscriptionsYamlContentCmd_QueuePresentTargetsSQS(t *testing.T) {
	subStore := &mockSubscriptionStore{}
	sns := &mockSNSManager{
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			return domain.SNSSubscription{ARN: "sub-2", TopicARN: topicARN, Protocol: protocol, Endpoint: endpoint}, nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCaseWithSubscriptions(&mockConfigStore{}, subStore)
	model := NewModel(awsUC, cfgUC, usecase.NewSnapshotUseCase(awsUC, cfgUC))

	yamlContent := "version: 1\nsubscriptions:\n  - name: notifications\n    topic: dev-webapi-notifications-sns\n    queue: dev-webapi-billings-sqs\n"
	cmd := model.importSubscriptionsYamlContentCmd(
		yamlContent,
		"arn:aws:sns:us-east-1:000000000000:dev-webapi-notifications-sns",
		[]domain.SNSTopic{{Name: "dev-webapi-notifications-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-notifications-sns"}},
		[]domain.SQSQueue{{Name: "dev-webapi-billings-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs"}},
	)

	msg := cmd()
	applied, ok := msg.(snsYamlImportAppliedMsg)
	if !ok {
		t.Fatalf("expected snsYamlImportAppliedMsg, got %T", msg)
	}
	if applied.Created != 1 || applied.Repaired != 0 {
		t.Fatalf("unexpected import counts: %#v", applied)
	}
	if len(subStore.savedSubs) != 1 {
		t.Fatalf("expected managed subscription saved, got %#v", subStore.savedSubs)
	}
	if subStore.savedSubs[0].DestinationType != "sqs" {
		t.Fatalf("expected destination type sqs, got %q", subStore.savedSubs[0].DestinationType)
	}
}

func TestImportSubscriptionsYamlContentCmd_DefaultQueueTargetsSQS(t *testing.T) {
	subStore := &mockSubscriptionStore{}
	sns := &mockSNSManager{
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			return domain.SNSSubscription{ARN: "sub-3", TopicARN: topicARN, Protocol: protocol, Endpoint: endpoint}, nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCaseWithSubscriptions(&mockConfigStore{}, subStore)
	model := NewModel(awsUC, cfgUC, usecase.NewSnapshotUseCase(awsUC, cfgUC))

	yamlContent := "version: 1\ndefault_queue: dev-webapi-billings-sqs\nsubscriptions:\n  - name: account\n    topic: dev-webapi-accounts-sns\n"
	cmd := model.importSubscriptionsYamlContentCmd(
		yamlContent,
		"arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns",
		[]domain.SNSTopic{{Name: "dev-webapi-accounts-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns"}},
		[]domain.SQSQueue{{Name: "dev-webapi-billings-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs"}},
	)

	msg := cmd()
	applied, ok := msg.(snsYamlImportAppliedMsg)
	if !ok {
		t.Fatalf("expected snsYamlImportAppliedMsg, got %T", msg)
	}
	if applied.Created != 1 {
		t.Fatalf("unexpected created count: %#v", applied)
	}
	if len(subStore.savedSubs) != 1 {
		t.Fatalf("expected managed subscription saved, got %#v", subStore.savedSubs)
	}
	if subStore.savedSubs[0].DestinationARN != "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs" {
		t.Fatalf("unexpected destination ARN: %q", subStore.savedSubs[0].DestinationARN)
	}
}

func TestImportSubscriptionsYamlContentCmd_WithoutQueueInfersSiblingQueue(t *testing.T) {
	subStore := &mockSubscriptionStore{}
	sns := &mockSNSManager{
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			return domain.SNSSubscription{ARN: "sub-4", TopicARN: topicARN, Protocol: protocol, Endpoint: endpoint}, nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCaseWithSubscriptions(&mockConfigStore{}, subStore)
	model := NewModel(awsUC, cfgUC, usecase.NewSnapshotUseCase(awsUC, cfgUC))

	yamlContent := "version: 1\nsubscriptions:\n  - name: pix\n    topic: dev-webapi-pix-sns\n"
	cmd := model.importSubscriptionsYamlContentCmd(
		yamlContent,
		"arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns",
		[]domain.SNSTopic{{Name: "dev-webapi-pix-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns"}},
		[]domain.SQSQueue{{Name: "dev-webapi-pix-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs"}},
	)

	msg := cmd()
	applied, ok := msg.(snsYamlImportAppliedMsg)
	if !ok {
		t.Fatalf("expected snsYamlImportAppliedMsg, got %T", msg)
	}
	if applied.Created != 1 {
		t.Fatalf("expected one inferred route, got %#v", applied)
	}
	if len(subStore.savedSubs) != 1 {
		t.Fatalf("expected managed subscription saved, got %#v", subStore.savedSubs)
	}
	if subStore.savedSubs[0].DestinationARN != "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs" {
		t.Fatalf("expected inferred sibling queue ARN, got %q", subStore.savedSubs[0].DestinationARN)
	}
}

func TestImportSubscriptionsYamlContentCmd_AllowsExplicitTopicDifferentFromActiveFile(t *testing.T) {
	subStore := &mockSubscriptionStore{}
	var createdTopicARN string
	sns := &mockSNSManager{
		CreateSubscriptionFunc: func(ctx context.Context, cfg *domain.AWSConfig, topicARN, protocol, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
			createdTopicARN = topicARN
			return domain.SNSSubscription{ARN: "sub-5", TopicARN: topicARN, Protocol: protocol, Endpoint: endpoint}, nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCaseWithSubscriptions(&mockConfigStore{}, subStore)
	model := NewModel(awsUC, cfgUC, usecase.NewSnapshotUseCase(awsUC, cfgUC))

	yamlContent := "version: 1\nsubscriptions:\n  - name: pix\n    topic: dev-webapi-pix-sns\n    event_type:\n      - pix_received\n"
	cmd := model.importSubscriptionsYamlContentCmd(
		yamlContent,
		"arn:aws:sns:us-east-1:000000000000:dev-webapi-statements-sns",
		[]domain.SNSTopic{
			{Name: "dev-webapi-statements-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-statements-sns"},
			{Name: "dev-webapi-pix-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns"},
		},
		[]domain.SQSQueue{{Name: "dev-webapi-pix-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-pix-sqs"}},
	)

	msg := cmd()
	applied, ok := msg.(snsYamlImportAppliedMsg)
	if !ok {
		t.Fatalf("expected snsYamlImportAppliedMsg, got %T", msg)
	}
	if applied.Created != 1 {
		t.Fatalf("expected one created subscription, got %#v", applied)
	}
	if createdTopicARN != "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns" {
		t.Fatalf("expected explicit topic to be honored, got %q", createdTopicARN)
	}
	if len(subStore.savedSubs) != 1 {
		t.Fatalf("expected managed subscription saved, got %#v", subStore.savedSubs)
	}
	if subStore.savedSubs[0].TopicARN != "arn:aws:sns:us-east-1:000000000000:dev-webapi-pix-sns" {
		t.Fatalf("expected saved managed subscription to use explicit topic, got %q", subStore.savedSubs[0].TopicARN)
	}
}

func TestImportSubscriptionsYamlContentCmd_RepaintsExistingSubscriptionScope(t *testing.T) {
	subStore := &mockSubscriptionStore{
		savedSubs: []domain.ManagedSubscription{
			{
				Name:            "account",
				TopicARN:        "arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns",
				DestinationARN:  "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs",
				DestinationType: "sqs",
				EventTypes:      []string{"account_created_event"},
				FilterScope:     domain.SubscriptionFilterScopeMessageAttributes,
				SubscriptionARN: "sub-existing",
			},
		},
	}
	var scopeUpdated, policyUpdated bool
	sns := &mockSNSManager{
		ListAllSubscriptionsFunc: func(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error) {
			return []domain.SNSSubscription{
				{
					ARN:          "sub-existing",
					TopicARN:     "arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns",
					Protocol:     "sqs",
					Endpoint:     "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs",
					FilterPolicy: map[string][]string{"event_type": []string{"account_created_event"}},
					FilterScope:  domain.SubscriptionFilterScopeMessageAttributes,
				},
			}, nil
		},
		SetSubscriptionAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN, attributeName, attributeValue string) error {
			if subscriptionARN != "sub-existing" {
				t.Fatalf("unexpected subscription arn: %s", subscriptionARN)
			}
			switch attributeName {
			case "FilterPolicyScope":
				scopeUpdated = true
				if attributeValue != "MessageBody" {
					t.Fatalf("expected message body scope, got %s", attributeValue)
				}
			case "FilterPolicy":
				policyUpdated = true
			default:
				t.Fatalf("unexpected attribute update: %s", attributeName)
			}
			return nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	cfgUC := usecase.NewConfigUseCaseWithSubscriptions(&mockConfigStore{}, subStore)
	model := NewModel(awsUC, cfgUC, usecase.NewSnapshotUseCase(awsUC, cfgUC))

	yamlContent := "version: 1\ndefault_queue: dev-webapi-billings-sqs\ndefault_filter_scope: message_body\nsubscriptions:\n  - name: account\n    topic: dev-webapi-accounts-sns\n    event_type:\n      - account_created_event\n"
	cmd := model.importSubscriptionsYamlContentCmd(
		yamlContent,
		"arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns",
		[]domain.SNSTopic{{Name: "dev-webapi-accounts-sns", ARN: "arn:aws:sns:us-east-1:000000000000:dev-webapi-accounts-sns"}},
		[]domain.SQSQueue{{Name: "dev-webapi-billings-sqs", ARN: "arn:aws:sqs:us-east-1:000000000000:dev-webapi-billings-sqs"}},
	)

	msg := cmd()
	applied, ok := msg.(snsYamlImportAppliedMsg)
	if !ok {
		t.Fatalf("expected snsYamlImportAppliedMsg, got %T", msg)
	}
	if applied.Repaired != 1 {
		t.Fatalf("expected 1 repaired subscription, got %#v", applied)
	}
	if !scopeUpdated {
		t.Fatal("expected filter scope update")
	}
	if policyUpdated {
		t.Fatal("did not expect policy update when policy was already correct")
	}
}

func TestUpdateSNSSubscriptionCmd_UpdatesPolicyScope(t *testing.T) {
	sns := &mockSNSManager{
		SetSubscriptionAttributesFunc: func(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN, attributeName, attributeValue string) error {
			switch attributeName {
			case "FilterPolicy":
				if attributeValue != `{"event_type":["pix_received"]}` {
					t.Fatalf("unexpected filter policy payload: %s", attributeValue)
				}
			case "FilterPolicyScope":
				if attributeValue != "MessageBody" {
					t.Fatalf("unexpected filter scope payload: %s", attributeValue)
				}
			default:
				t.Fatalf("unexpected attribute: %s", attributeName)
			}
			return nil
		},
	}
	awsUC := usecase.NewAWSUseCase(&mockS3Manager{}, &mockSQSManager{}, sns, &mockSecretsManager{})
	model := mkModel()
	model.awsUseCase = awsUC

	msg := model.updateSNSSubscriptionCmd("sub-1", map[string][]string{"event_type": []string{"pix_received"}}, domain.SubscriptionFilterScopeMessageBody)()
	if _, ok := msg.(snsSubscriptionUpdatedMsg); !ok {
		t.Fatalf("expected snsSubscriptionUpdatedMsg, got %T", msg)
	}
}
