package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"monostack/internal/domain"
	"monostack/internal/usecase"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshLogViewport()
		m.configureSecretsLayout()
		if m.showInspectionModal {
			m.configureInspectionViewport()
		}
		return m, nil

	case splashTickMsg:
		if m.showSplash {
			m.splashFrame++
			if m.splashFrame < splashFrameLimit {
				return m, m.splashTickCmd()
			}
			m.showSplash = false
		}
		return m, nil

	case configLoadedMsg:
		m.config = msg.Config
		m.loading = false
		m.leftPanelRatio = msg.Config.LeftPanelRatio
		m.appendCommandLog("config", msg.Config.ServiceName, "configuration loaded", nil)
		if len(m.settingsInputs) == 0 {
			m.settingsInputs = make([]textinput.Model, 8)
			for i := range m.settingsInputs {
				m.settingsInputs[i] = textinput.New()
				m.settingsInputs[i].Width = 40
			}
		}
		m.syncSettingsInputsFromConfig()
		m.settingsEditMode = false
		m.ensureActiveTabVisible()
		return m, m.reloadTabCmd()

	case configSavedMsg:
		m.config = msg.Config
		m.leftPanelRatio = msg.Config.LeftPanelRatio
		m.loading = false
		m.syncSettingsInputsFromConfig()
		m.appendCommandLog("config", msg.Config.ServiceName, "configuration saved", nil)
		m.ensureActiveTabVisible()
		cmds = []tea.Cmd{m.setStatusMessage("Configuration saved successfully!")}
		if reload := m.reloadTabCmd(); reload != nil {
			cmds = append(cmds, reload)
		}
		return m, tea.Batch(cmds...)

	case s3BucketsLoadedMsg:
		m.buckets = msg.Buckets
		m.loading = false
		m.errorMsg = ""
		if len(m.buckets) > 0 && m.selectedBucketIndex < len(m.buckets) {
			bucket := m.buckets[m.selectedBucketIndex].Name
			return m, m.loadS3ObjectsCmd(bucket)
		}
		return m, nil

	case s3ObjectsLoadedMsg:
		m.s3ObjectsCache[msg.Bucket] = msg.Objects
		m.s3ObjectsLoadedFor = msg.Bucket
		m.objects = msg.Objects
		m.selectedObjectIndex = 0
		m.loading = false
		m.errorMsg = ""
		return m, nil

	case s3ObjectDeletedMsg:
		m.loading = false
		delete(m.s3ObjectsCache, msg.Bucket)
		if m.s3ObjectsLoadedFor == msg.Bucket {
			m.s3ObjectsLoadedFor = ""
		}
		if m.selectedBucketIndex < len(m.buckets) {
			bucket := m.buckets[m.selectedBucketIndex].Name
			return m, tea.Batch(m.setStatusMessage("Object deleted successfully!"), m.loadS3ObjectsCmd(bucket))
		}
		return m, m.setStatusMessage("Object deleted successfully!")

	case s3BucketDeletedMsg:
		m.loading = false
		m.appendCommandLog("delete bucket", msg.Bucket, "bucket deleted", nil)
		delete(m.s3ObjectsCache, msg.Bucket)
		m.s3ObjectsLoadedFor = ""
		m.buckets = nil
		m.selectedBucketIndex = 0
		m.selectedObjectIndex = 0
		m.s3Focus = focusBuckets
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Bucket \"%s\" deleted successfully!", msg.Bucket)), m.loadS3BucketsCmd())

	case s3BucketCreatedMsg:
		m.loading = false
		m.appendCommandLog("create bucket", msg.Bucket, "bucket created", nil)
		m.buckets = nil
		m.selectedBucketIndex = 0
		m.selectedObjectIndex = 0
		m.s3Focus = focusBuckets
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Bucket \"%s\" created successfully!", msg.Bucket)), m.loadS3BucketsCmd())

	case s3FolderCreatedMsg:
		m.loading = false
		m.appendCommandLog("create folder", msg.Bucket+"/"+msg.Key, "folder created", nil)
		delete(m.s3ObjectsCache, msg.Bucket)
		m.s3ObjectsLoadedFor = ""
		if msg.Bucket != "" {
			return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Folder \"%s\" created successfully!", msg.Key)), m.loadS3ObjectsCmd(msg.Bucket))
		}
		return m, m.setStatusMessage(fmt.Sprintf("Folder \"%s\" created successfully!", msg.Key))

	case s3ObjectDownloadedMsg:
		m.loading = false
		m.appendCommandLog("download object", msg.DestPath, "object downloaded", nil)
		return m, m.setStatusMessage("File downloaded to: " + msg.DestPath)

	case s3ObjectUploadedMsg:
		m.loading = false
		m.appendCommandLog("upload object", msg.Bucket+"/"+msg.Key, "object uploaded", nil)
		delete(m.s3ObjectsCache, msg.Bucket)
		m.s3ObjectsLoadedFor = ""
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Object %q uploaded to bucket %q", msg.Key, msg.Bucket)), m.loadS3ObjectsCmd(msg.Bucket))

	case sqsQueuesLoadedMsg:
		m.queues = msg.Queues
		m.allSubscriptions = msg.AllSubscriptions
		m.loading = false
		m.errorMsg = ""
		m.queueSubscriptions = nil
		if m.activeTab == panelSQS && len(m.queues) > 0 {
			q := m.queues[m.selectedQueueIndex]
			return m, m.loadSQSQueueSubscriptionsCmd(q.URL, q.ARN)
		}
		return m, nil

	case sqsQueuePurgedMsg:
		m.loading = false
		m.appendCommandLog("purge queue", msg.QueueURL, "queue purged", nil)
		return m, tea.Batch(m.setStatusMessage("SQS Queue purged successfully!"), m.loadSQSQueuesCmd())

	case sqsQueuesPurgedMsg:
		m.loading = false
		m.appendCommandLog("purge all queues", "SQS", fmt.Sprintf("%d queues purged", msg.Count), nil)
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Purged %d SQS queues successfully!", msg.Count)), m.loadSQSQueuesCmd())

	case sqsQueueDeletedMsg:
		m.loading = false
		m.appendCommandLog("delete queue", msg.QueueURL, "queue deleted", nil)
		if m.selectedQueueIndex >= len(m.queues)-1 && m.selectedQueueIndex > 0 {
			m.selectedQueueIndex--
		}
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Queue \"%s\" deleted successfully!", msg.Name)), m.loadSQSQueuesCmd())

	case sqsQueueCreatedMsg:
		m.loading = false
		m.appendCommandLog("create queue", msg.Name, "queue created", nil)
		m.selectedQueueIndex = 0
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Queue \"%s\" created successfully!", msg.Name)), m.loadSQSQueuesCmd())

	case sqsMessageSentMsg:
		m.loading = false
		m.appendCommandLog("send message", msg.QueueURL, "message sent", nil)
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("SQS message published to %s successfully!", msg.QueueURL)), m.loadSQSQueuesCmd())

	case sqsMessagesLoadedMsg:
		m.peekMessages = msg.Messages
		m.showPeekModal = true
		m.loading = false
		m.errorMsg = ""
		return m, nil

	case sqsSubscriptionsLoadedMsg:
		m.queueSubscriptions = msg.Subscriptions
		m.loading = false
		m.errorMsg = ""
		return m, nil

	case snsTopicsLoadedMsg:
		m.topics = msg.Topics
		m.allSubscriptions = msg.AllSubscriptions
		m.loading = false
		m.errorMsg = ""
		if len(m.topics) > 0 && m.selectedTopicIndex >= len(m.topics) {
			m.selectedTopicIndex = 0
		}
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(
				m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN),
				m.loadManagedSubscriptionsCmd(),
			)
		}
		return m, m.loadManagedSubscriptionsCmd()

	case snsTopicCreatedMsg:
		m.loading = false
		m.appendCommandLog("create topic", msg.Topic.ARN, "topic created", nil)
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("SNS topic %s created successfully!", msg.Topic.Name)), m.loadSNSTopicsCmd())

	case snsTopicDeletedMsg:
		m.loading = false
		m.appendCommandLog("delete topic", msg.ARN, "topic deleted", nil)
		if m.selectedTopicIndex >= len(m.topics) {
			m.selectedTopicIndex = 0
		}
		return m, tea.Batch(m.setStatusMessage("SNS topic deleted successfully!"), m.loadSNSTopicsCmd())

	case secretsLoadedMsg:
		m.secrets = msg.Secrets
		m.loading = false
		m.errorMsg = ""
		if len(m.secrets) > 0 && m.selectedSecretIndex >= len(m.secrets) {
			m.selectedSecretIndex = 0
		}
		if len(m.secrets) > 0 {
			secretID := m.secrets[m.selectedSecretIndex].ARN
			return m, m.loadSecretDetailsCmd(secretID)
		}
		return m, nil

	case secretDetailsLoadedMsg:
		m.secretDetailsLoadedFor = msg.Secret.ARN
		m.secretValueLoadedFor = msg.Secret.ARN
		m.secretVersions = orderedSecretVersions(msg.Versions)
		m.secretValue = msg.Value
		m.secretValueDisplay = formatSecretValueDisplay(msg.Value.SecretString)
		if strings.TrimSpace(m.secretValueDisplay) == "(empty)" && strings.TrimSpace(msg.Value.SecretBinaryBase64) != "" {
			m.secretValueDisplay = formatSecretValueDisplay(msg.Value.SecretBinaryBase64)
		}
		m.selectedSecretVersionIndex = 0
		if idx := m.currentSecretCurrentVersionIndex(); idx >= 0 {
			m.selectedSecretVersionIndex = idx
		}
		m.refreshSecretValueViewport()
		m.loading = false
		m.errorMsg = ""
		if len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
			m.secrets[m.selectedSecretIndex] = msg.Secret
		}
		return m, nil

	case secretCreatedMsg:
		m.loading = false
		m.appendCommandLog("create secret", msg.Secret.Name, "secret created", nil)
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Secret %q created successfully!", msg.Secret.Name)), m.loadSecretsCmd())

	case secretValueUpdatedMsg:
		m.loading = false
		m.appendCommandLog("update secret", msg.SecretID, "secret value updated", nil)
		return m, tea.Batch(m.setStatusMessage("Secret value updated successfully!"), m.loadSecretsCmd())

	case secretDeletedMsg:
		m.loading = false
		m.appendCommandLog("delete secret", msg.Name, "secret deleted", nil)
		if m.selectedSecretIndex >= len(m.secrets)-1 && m.selectedSecretIndex > 0 {
			m.selectedSecretIndex--
		}
		return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("Secret %q deleted successfully!", msg.Name)), m.loadSecretsCmd())

	case secretRestoredMsg:
		m.loading = false
		m.appendCommandLog("restore secret", msg.SecretID, "secret restored", nil)
		return m, tea.Batch(m.setStatusMessage("Secret restored successfully!"), m.loadSecretsCmd())

	case secretStageUpdatedMsg:
		m.loading = false
		m.appendCommandLog("promote version", msg.SecretID, msg.VersionID, nil)
		return m, tea.Batch(m.setStatusMessage("Secret version promoted successfully!"), m.loadSecretsCmd())

	case snsSubscriptionsLoadedMsg:
		m.subscriptions = append(msg.Subscriptions, msg.IncomingSubscriptions...)
		if msg.AllSubscriptions != nil {
			m.allSubscriptions = msg.AllSubscriptions
		}
		m.snsOutgoingCount = len(msg.Subscriptions)
		if m.selectedSubIndex >= len(m.subscriptions) {
			m.selectedSubIndex = 0
		}
		m.loading = false
		m.errorMsg = ""
		return m, nil

	case snsSubscriptionCreatedMsg:
		m.loading = false
		m.appendCommandLog("create subscription", msg.Subscription.Endpoint, "subscription created", nil)
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(m.setStatusMessage("SNS subscription created successfully!"), m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN))
		}
		return m, m.setStatusMessage("SNS subscription created successfully!")

	case snsMessagePublishedMsg:
		m.loading = false
		m.appendCommandLog("publish message", msg.TopicARN, "message published", nil)
		return m, m.setStatusMessage(fmt.Sprintf("SNS message published to topic %s successfully!", msg.TopicARN))

	case snsSubscriptionDeletedMsg:
		m.loading = false
		m.appendCommandLog("delete subscription", msg.ARN, "subscription deleted", nil)
		if m.activeTab == panelSQS && len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
			q := m.queues[m.selectedQueueIndex]
			return m, tea.Batch(m.setStatusMessage("SNS subscription deleted successfully!"), m.loadSQSQueueSubscriptionsCmd(q.URL, q.ARN), m.loadSQSQueuesCmd())
		}
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(m.setStatusMessage("SNS subscription deleted successfully!"), m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN))
		}
		return m, m.setStatusMessage("SNS subscription deleted successfully!")

	case snsSubscriptionUpdatedMsg:
		m.loading = false
		m.appendCommandLog("update subscription", msg.ARN, "filter updated", nil)
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(m.setStatusMessage("SNS subscription filter updated successfully!"), m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN))
		}
		return m, m.setStatusMessage("SNS subscription filter updated successfully!")

	case managedSubscriptionsLoadedMsg:
		m.managedSubs = msg.Subscriptions
		return m, nil

	case managedSubscriptionsUpdatedMsg:
		return m, m.loadManagedSubscriptionsCmd()

	case snsBatchSubscriptionsCreatedMsg:
		m.loading = false
		m.appendCommandLog("batch subscribe", "SNS", fmt.Sprintf("%d subscriptions created", msg.Count), nil)
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("%d SNS subscriptions created successfully!", msg.Count)), m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN))
		}
		return m, m.setStatusMessage(fmt.Sprintf("%d SNS subscriptions created successfully!", msg.Count))

	case snsYamlImportAppliedMsg:
		m.loading = false
		m.appendCommandLog("yaml import", "SNS", fmt.Sprintf("%d created, %d repaired, %d unchanged", msg.Created, msg.Repaired, msg.Unchanged), nil)
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("%d created, %d repaired from YAML!", msg.Created, msg.Repaired)), m.loadSNSSubscriptionsCmd(m.topics[m.selectedTopicIndex].ARN))
		}
		return m, m.setStatusMessage(fmt.Sprintf("%d created, %d repaired from YAML!", msg.Created, msg.Repaired))

	case sqsBatchSubscriptionsCreatedMsg:
		m.loading = false
		m.appendCommandLog("batch subscribe", "SQS", fmt.Sprintf("%d subscriptions created", msg.Count), nil)
		if len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
			q := m.queues[m.selectedQueueIndex]
			return m, tea.Batch(m.setStatusMessage(fmt.Sprintf("%d SQS subscriptions created successfully!", msg.Count)), m.loadSQSQueueSubscriptionsCmd(q.URL, q.ARN), m.loadSQSQueuesCmd())
		}
		return m, m.setStatusMessage(fmt.Sprintf("%d SQS subscriptions created successfully!", msg.Count))

	case yamlScriptLoadedMsg:
		if m.snsYamlSavedContent == nil {
			m.snsYamlSavedContent = make(map[string]string)
		}
		m.snsYamlSavedContent[msg.TopicARN] = msg.Content
		if m.snsYamlCurrentTopicARN == msg.TopicARN && msg.Content != "" {
			currentValue := m.snsYamlImportTextarea.Value()
			if strings.TrimSpace(currentValue) == "" || currentValue == m.snsYamlSavedContent[msg.TopicARN] {
				m.snsYamlImportTextarea.SetValue(msg.Content)
			}
		}
		return m, nil

	case yamlScriptSavedMsg:
		if m.snsYamlSavedContent == nil {
			m.snsYamlSavedContent = make(map[string]string)
		}
		m.snsYamlSavedContent[msg.TopicARN] = m.snsYamlImportTextarea.Value()
		return m, nil

	case inspectionLoadedMsg:
		m.loading = false
		m.errorMsg = ""
		m.inspectionTitle = msg.Title
		m.inspectionSubtitle = msg.Subtitle
		m.inspectionContent = msg.Content
		m.showInspectionModal = true
		m.configureInspectionViewport()
		return m, nil

	case profileExportedMsg:
		m.loading = false
		m.appendCommandLog("export snapshot", msg.Path, "snapshot exported", nil)
		if m.config != nil {
			m.config.SnapshotPath = msg.Path
		}
		return m, m.setStatusMessage(fmt.Sprintf("Snapshot exported to %s", msg.Path))

	case profileImportedMsg:
		if cfg := msg.Config; cfg != nil {
			m.config = cfg
			m.leftPanelRatio = cfg.LeftPanelRatio
			m.syncSettingsInputsFromConfig()
			m.ensureActiveTabVisible()
		}
		m.loading = false
		m.appendCommandLog("import snapshot", msg.Path, fmt.Sprintf("%d subscriptions restored", msg.SubsCount), nil)
		if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			cmds = []tea.Cmd{m.loadYamlScriptCmd(m.topics[m.selectedTopicIndex].ARN)}
		}
		if reload := m.reloadTabCmd(); reload != nil {
			cmds = append(cmds, reload)
		}
		return m, tea.Batch(append([]tea.Cmd{m.setStatusMessage(fmt.Sprintf("Snapshot loaded: %d subscriptions restored from %s", msg.SubsCount, msg.Path))}, cmds...)...)

	case statusMsg:
		m.loading = false
		return m, m.setStatusMessage(msg.Message)

	case errMsg:
		errStr := msg.Error.Error()
		m.loading = false
		m.appendCommandLog("error", m.currentResourceTarget(), errStr, msg.Error)
		if strings.Contains(errStr, "connect: connection refused") || strings.Contains(errStr, "request send failed") || strings.Contains(errStr, "dial tcp") {
			endpoint := "localhost:4566"
			if m.config != nil && m.config.EndpointURL != "" {
				endpoint = m.config.EndpointURL
			}
			return m, m.setErrorMessage(fmt.Sprintf("Service unreachable at %s. Is MiniStack/LocalStack running? (Press [4] for Settings to configure or enable Mock Mode)", endpoint))
		} else {
			return m, m.setErrorMessage(errStr)
		}

	case clearStatusMsg:
		if m.statusMsgID == msg.id {
			m.statusMsg = ""
			m.errorMsg = ""
		}
		return m, nil

	case tea.KeyMsg:

		if m.showS3CreateModal {
			switch msg.String() {
			case "esc":
				m.showS3CreateModal = false
				m.s3CreateInput.Blur()
				m.s3CreateInput.SetValue("")
				return m, nil
			case "enter":
				name := m.s3CreateInput.Value()
				m.showS3CreateModal = false
				m.s3CreateInput.Blur()
				m.s3CreateInput.SetValue("")
				if name != "" {
					m.loading = true
					return m, m.createS3BucketCmd(name)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.s3CreateInput, cmd = m.s3CreateInput.Update(msg)
				return m, cmd
			}
		}

		if m.showS3CreateFolderModal {
			switch msg.String() {
			case "esc":
				m.showS3CreateFolderModal = false
				m.s3FolderInput.Blur()
				m.s3FolderInput.SetValue("")
				return m, nil
			case "enter":
				prefix := strings.TrimSpace(m.s3FolderInput.Value())
				bucket := m.selectedS3BucketName()
				m.showS3CreateFolderModal = false
				m.s3FolderInput.Blur()
				m.s3FolderInput.SetValue("")
				if bucket != "" && prefix != "" {
					m.loading = true
					delete(m.s3ObjectsCache, bucket)
					m.s3ObjectsLoadedFor = ""
					return m, m.createS3FolderCmd(bucket, prefix)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.s3FolderInput, cmd = m.s3FolderInput.Update(msg)
				return m, cmd
			}
		}

		if m.showS3UploadModal {
			switch msg.String() {
			case "esc":
				m.showS3UploadModal = false
				m.s3UploadFocus = 0
				m.s3UploadPathInput.Blur()
				m.s3UploadKeyInput.Blur()
				m.s3UploadPathInput.SetValue("")
				m.s3UploadKeyInput.SetValue("")
				return m, nil
			case "tab":
				if m.s3UploadFocus == 0 {
					m.s3UploadFocus = 1
					m.s3UploadPathInput.Blur()
					m.s3UploadKeyInput.Focus()
				} else {
					m.s3UploadFocus = 0
					m.s3UploadKeyInput.Blur()
					m.s3UploadPathInput.Focus()
				}
				return m, nil
			case "enter":
				filePath := expandPath(strings.TrimSpace(m.s3UploadPathInput.Value()))
				key := strings.TrimSpace(m.s3UploadKeyInput.Value())
				bucket := m.selectedS3BucketName()
				m.showS3UploadModal = false
				m.s3UploadFocus = 0
				m.s3UploadPathInput.Blur()
				m.s3UploadKeyInput.Blur()
				m.s3UploadPathInput.SetValue("")
				m.s3UploadKeyInput.SetValue("")
				if bucket != "" && filePath != "" {
					if key == "" {
						key = defaultS3ObjectKey(filePath)
					}
					m.loading = true
					delete(m.s3ObjectsCache, bucket)
					m.s3ObjectsLoadedFor = ""
					return m, m.uploadS3ObjectCmd(bucket, filePath, key)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				if m.s3UploadFocus == 0 {
					before := m.s3UploadPathInput.Value()
					m.s3UploadPathInput, cmd = m.s3UploadPathInput.Update(msg)
					after := m.s3UploadPathInput.Value()
					if after != before && strings.TrimSpace(m.s3UploadKeyInput.Value()) == "" {
						m.s3UploadKeyInput.SetValue(defaultS3ObjectKey(after))
					}
				} else {
					m.s3UploadKeyInput, cmd = m.s3UploadKeyInput.Update(msg)
				}
				return m, cmd
			}
		}

		if m.showS3PreviewModal {
			if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "v" {
				m.showS3PreviewModal = false
				return m, nil
			}
			if msg.String() == "b" {
				if bucket := m.selectedS3BucketName(); bucket != "" && m.selectedObjectIndex < len(m.objects) {
					return m, m.openPresignedURLCmd(bucket, m.objects[m.selectedObjectIndex].Key)
				}
			}
			if msg.String() == "w" {
				if bucket := m.selectedS3BucketName(); bucket != "" && m.selectedObjectIndex < len(m.objects) {
					return m, m.downloadS3ObjectCmd(bucket, m.objects[m.selectedObjectIndex].Key)
				}
			}
			return m, nil
		}

		if m.showSqsCreateModal {
			switch msg.String() {
			case "esc":
				m.showSqsCreateModal = false
				m.sqsCreateInput.Blur()
				m.sqsCreateInput.SetValue("")
				return m, nil
			case "enter":
				name := m.sqsCreateInput.Value()
				m.showSqsCreateModal = false
				m.sqsCreateInput.Blur()
				m.sqsCreateInput.SetValue("")
				if name != "" {
					m.loading = true
					return m, m.createSQSQueueCmd(name)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.sqsCreateInput, cmd = m.sqsCreateInput.Update(msg)
				return m, cmd
			}
		}

		if m.showSqsPurgeAllConfirm {
			switch msg.String() {
			case "y", "Y":
				queueURLs := make([]string, 0, len(m.queues))
				for _, queue := range m.queues {
					queueURLs = append(queueURLs, queue.URL)
				}
				m.showSqsPurgeAllConfirm = false
				m.loading = true
				return m, m.purgeSQSQueuesCmd(queueURLs)
			default:
				m.showSqsPurgeAllConfirm = false
				return m, nil
			}
		}

		if m.showSqsSendModal {
			switch msg.String() {
			case "esc":
				m.showSqsSendModal = false
				m.sqsSendMessageInput.Blur()
				return m, nil
			case "enter":
				body := m.sqsSendMessageInput.Value()
				m.showSqsSendModal = false
				m.sqsSendMessageInput.Blur()
				m.sqsSendMessageInput.SetValue("")
				if len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
					qURL := m.queues[m.selectedQueueIndex].URL
					return m, tea.Batch(
						m.sendSQSMessageCmd(qURL, body),
						m.loadSQSQueuesCmd(),
					)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.sqsSendMessageInput, cmd = m.sqsSendMessageInput.Update(msg)
				return m, cmd
			}
		}

		if m.showS3ConfirmDelete {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showS3ConfirmDelete = false
				m.loading = true
				if m.s3DeleteIsBucket {
					return m, m.deleteS3BucketCmd(m.s3DeleteBucket)
				}
				return m, m.deleteS3ObjectCmd(m.s3DeleteBucket, m.s3DeleteKey)
			}
			m.showS3ConfirmDelete = false
			return m, nil
		}

		if m.showSqsConfirmDelete {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSqsConfirmDelete = false
				m.loading = true
				return m, m.deleteSQSQueueCmd(m.sqsDeleteQueueURL, m.sqsDeleteQueueName)
			}
			m.showSqsConfirmDelete = false
			return m, nil
		}

		if m.showSqsSubDeleteConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSqsSubDeleteConfirm = false
				m.loading = true
				subARN := m.sqsDeleteSubARN
				return m, tea.Batch(
					m.deleteSNSSubscriptionCmd(subARN),
					m.removeManagedSubByARN(subARN),
				)
			}
			m.showSqsSubDeleteConfirm = false
			return m, nil
		}

		if m.showSnsPublishModal {
			switch msg.String() {
			case "esc":
				m.showSnsPublishModal = false
				m.snsPublishInput.Blur()
				m.snsPublishAttrsInput.Blur()
				return m, nil
			case "tab":
				if m.snsPublishInput.Focused() {
					m.snsPublishInput.Blur()
					m.snsPublishAttrsInput.Focus()
				} else {
					m.snsPublishAttrsInput.Blur()
					m.snsPublishInput.Focus()
				}
				return m, nil
			case "enter":
				body := m.snsPublishInput.Value()
				attrStr := m.snsPublishAttrsInput.Value()
				m.showSnsPublishModal = false
				m.snsPublishInput.Blur()
				m.snsPublishInput.SetValue("")
				m.snsPublishAttrsInput.Blur()
				m.snsPublishAttrsInput.SetValue("")
				if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
					topicARN := m.topics[m.selectedTopicIndex].ARN

					attrs := parseMessageAttributes(attrStr)
					return m, m.publishSNSMessageCmd(topicARN, body, attrs)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				if m.snsPublishInput.Focused() {
					m.snsPublishInput, cmd = m.snsPublishInput.Update(msg)
				} else {
					m.snsPublishAttrsInput, cmd = m.snsPublishAttrsInput.Update(msg)
				}
				return m, cmd
			}
		}

		if m.showSnsCreateTopicModal {
			switch msg.String() {
			case "esc":
				m.showSnsCreateTopicModal = false
				m.snsCreateTopicInput.Blur()
				m.snsCreateTopicInput.SetValue("")
				return m, nil
			case "enter":
				name := m.snsCreateTopicInput.Value()
				m.showSnsCreateTopicModal = false
				m.snsCreateTopicInput.Blur()
				m.snsCreateTopicInput.SetValue("")
				if name != "" {
					m.loading = true
					return m, m.createSNSTopicCmd(name)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.snsCreateTopicInput, cmd = m.snsCreateTopicInput.Update(msg)
				return m, cmd
			}
		}

		if m.showSnsSimpleSubModal {
			switch m.snsSimpleSubStep {
			case 0:
				switch msg.String() {
				case "esc":
					m.showSnsSimpleSubModal = false
					m.snsSimpleSubStep = 0
					return m, nil
				case "up", "k":
					if m.snsSimpleSubCursor > 0 {
						m.snsSimpleSubCursor--
					}
					return m, nil
				case "down", "j":
					if m.snsSimpleSubCursor < len(m.snsSimpleSubSources)-1 {
						m.snsSimpleSubCursor++
					}
					return m, nil
				case "enter":
					if m.snsSimpleSubCursor < len(m.snsSimpleSubSources) {
						m.snsSimpleSubStep = 1
						m.snsSimpleSubEventInput.Focus()
					}
					return m, nil
				}
				return m, nil
			case 1:
				switch msg.String() {
				case "esc":
					m.snsSimpleSubStep = 0
					m.snsSimpleSubEventInput.Blur()
					m.snsSimpleSubEventInput.SetValue("")
					return m, nil
				case "enter":
					m.showSnsSimpleSubModal = false
					m.snsSimpleSubEventInput.Blur()
					m.snsSimpleSubEventInput.SetValue("")
					m.snsSimpleSubStep = 0

					if m.snsSimpleSubCursor >= len(m.snsSimpleSubSources) {
						return m, nil
					}

					m.showSnsSimpleSubModal = false
					return m, m.setStatusMessage("SNS topic-to-topic subscriptions are not supported; subscribe SQS queues directly")
				default:
					var cmd tea.Cmd
					m.snsSimpleSubEventInput, cmd = m.snsSimpleSubEventInput.Update(msg)
					return m, cmd
				}
			}
		}

		if m.showSnsBatchSubModal {
			switch msg.String() {
			case "esc":
				m.showSnsBatchSubModal = false
				return m, nil
			case "up", "k":
				if m.snsBatchSubCursor > 0 {
					m.snsBatchSubCursor--
				}
				return m, nil
			case "down", "j":
				if m.snsBatchSubCursor < len(m.snsBatchSubList)-1 {
					m.snsBatchSubCursor++
				}
				return m, nil
			case " ":
				if m.snsBatchSubCursor < len(m.snsBatchSubList) {
					m.snsBatchSubList[m.snsBatchSubCursor].checked =
						!m.snsBatchSubList[m.snsBatchSubCursor].checked
				}
				return m, nil
			case "enter":
				var selected []toggleOption
				for _, opt := range m.snsBatchSubList {
					if opt.checked {
						selected = append(selected, opt)
					}
				}
				m.showSnsBatchSubModal = false
				if len(selected) > 0 && len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
					m.showSnsBatchSubModal = false
					return m, m.setStatusMessage("SNS topic-to-topic subscriptions are not supported; subscribe SQS queues directly")
				}
				return m, nil
			}
			return m, nil
		}

		if m.showSnsYamlApplyConfirm {
			switch msg.String() {
			case "esc", "n", "N":
				m.showSnsYamlApplyConfirm = false
				m.snsYamlPendingContent = ""
				return m, nil
			case "y", "Y":
				content := m.snsYamlPendingContent
				m.showSnsYamlApplyConfirm = false
				m.showSnsYamlImportModal = false
				m.snsYamlImportTextarea.Blur()
				m.snsYamlPendingContent = ""
				topicARN := m.snsYamlCurrentTopicARN
				if content != "" {
					m.loading = true
					return m, tea.Batch(
						m.saveYamlScriptCmd(topicARN, content),
						m.importSubscriptionsYamlContentCmd(content, topicARN, m.topics, m.queues),
					)
				}
				return m, m.saveYamlScriptCmd(topicARN, content)
			}
			return m, nil
		}

		if m.showSnsYamlImportModal {
			switch msg.String() {
			case "esc":

				content := m.snsYamlImportTextarea.Value()
				m.showSnsYamlImportModal = false
				m.snsYamlImportTextarea.Blur()
				return m, m.saveYamlScriptCmd(m.snsYamlCurrentTopicARN, content)
			case "ctrl+s":
				m.snsYamlPendingContent = m.snsYamlImportTextarea.Value()
				m.showSnsYamlApplyConfirm = true
				return m, nil
			case "ctrl+k":
				m.snsYamlImportTextarea.SetValue("")
				m.snsYamlImportTextarea.Blur()
				m.snsYamlImportTextarea.Focus()
				if m.snsYamlSavedContent != nil {
					m.snsYamlSavedContent[m.snsYamlCurrentTopicARN] = ""
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.snsYamlImportTextarea, cmd = m.snsYamlImportTextarea.Update(msg)
				return m, cmd
			}
		}

		if m.showSnsSubEditModal {
			switch msg.String() {
			case "esc":
				m.showSnsSubEditModal = false
				m.snsSubEditInput.Blur()
				m.snsSubEditInput.SetValue("")
				return m, nil
			case "enter":
				eventTypesStr := m.snsSubEditInput.Value()
				m.showSnsSubEditModal = false
				m.snsSubEditInput.Blur()
				m.snsSubEditInput.SetValue("")

				eventTypes := splitCSVList(eventTypesStr)

				var filterPolicy map[string][]string
				if len(eventTypes) > 0 {
					filterPolicy = map[string][]string{"event_type": eventTypes}
				}

				if m.selectedSubIndex < len(m.subscriptions) {
					subARN := m.subscriptions[m.selectedSubIndex].ARN
					filterScope := m.subscriptions[m.selectedSubIndex].FilterScope
					m.loading = true
					return m, m.updateSNSSubscriptionCmd(subARN, filterPolicy, filterScope)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.snsSubEditInput, cmd = m.snsSubEditInput.Update(msg)
				return m, cmd
			}
		}

		if m.showSnsConfirmDelete {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSnsConfirmDelete = false
				m.loading = true
				if m.snsDeleteIsTopic {
					return m, m.deleteSNSTopicCmd(m.snsDeleteARN)
				}
				return m, tea.Batch(
					m.deleteSNSSubscriptionCmd(m.snsDeleteARN),
					m.removeManagedSubByARN(m.snsDeleteARN),
				)
			}
			m.showSnsConfirmDelete = false
			return m, nil
		}

		if m.showSecretCreateModal {
			switch msg.String() {
			case "esc":
				m.showSecretCreateModal = false
				m.secretCreateStep = 0
				m.secretCreateNameInput.Blur()
				m.secretCreateValueInput.Blur()
				m.secretCreateNameInput.SetValue("")
				m.secretCreateValueInput.SetValue("")
				return m, nil
			case "tab":
				m.secretCreateStep = (m.secretCreateStep + 1) % 2
				m.secretCreateNameInput.Blur()
				m.secretCreateValueInput.Blur()
				switch m.secretCreateStep {
				case 0:
					m.secretCreateNameInput.Focus()
				case 1:
					m.secretCreateValueInput.Focus()
				}
				return m, nil
			case "ctrl+s":
				name := strings.TrimSpace(m.secretCreateNameInput.Value())
				value := m.secretCreateValueInput.Value()
				m.showSecretCreateModal = false
				m.secretCreateStep = 0
				m.secretCreateNameInput.Blur()
				m.secretCreateValueInput.Blur()
				m.secretCreateNameInput.SetValue("")
				m.secretCreateValueInput.SetValue("")
				if name != "" {
					m.loading = true
					return m, m.createSecretCmd(name, value, "")
				}
				return m, nil
			case "enter":
				if m.secretCreateStep == 0 {
					m.secretCreateStep = 1
					m.secretCreateNameInput.Blur()
					m.secretCreateValueInput.Focus()
					return m, nil
				}
				var cmd tea.Cmd
				m.secretCreateValueInput, cmd = m.secretCreateValueInput.Update(msg)
				return m, cmd
			default:
				var cmd tea.Cmd
				switch m.secretCreateStep {
				case 0:
					m.secretCreateNameInput, cmd = m.secretCreateNameInput.Update(msg)
				case 1:
					m.secretCreateValueInput, cmd = m.secretCreateValueInput.Update(msg)
				}
				return m, cmd
			}
		}

		if m.showSecretUpdateModal {
			switch msg.String() {
			case "esc":
				m.showSecretUpdateModal = false
				m.secretUpdateStep = 0
				m.secretUpdateValueInput.Blur()
				m.secretUpdateValueInput.SetValue("")
				return m, nil
			case "ctrl+l":
				m.secretUpdateValueInput.SetValue("")
				m.secretUpdateValueInput.CursorEnd()
				return m, nil
			case "ctrl+s":
				value := m.secretUpdateValueInput.Value()
				secretID := ""
				if m.selectedSecretIndex < len(m.secrets) {
					secretID = m.secrets[m.selectedSecretIndex].ARN
				}
				m.showSecretUpdateModal = false
				m.secretUpdateStep = 0
				m.secretUpdateValueInput.Blur()
				m.secretUpdateValueInput.SetValue("")
				if secretID != "" {
					m.loading = true
					return m, m.updateSecretValueCmd(secretID, value, "")
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.secretUpdateValueInput, cmd = m.secretUpdateValueInput.Update(msg)
				return m, cmd
			}
		}

		if m.showSecretDeleteConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSecretDeleteConfirm = false
				m.loading = true
				return m, m.deleteSecretCmd(m.secretDeleteID, m.secretDeleteName)
			}
			m.showSecretDeleteConfirm = false
			return m, nil
		}

		if m.showSecretRestoreConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSecretRestoreConfirm = false
				m.loading = true
				return m, m.restoreSecretCmd(m.secretDeleteID)
			}
			m.showSecretRestoreConfirm = false
			return m, nil
		}

		if m.showSecretPromoteConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showSecretPromoteConfirm = false
				m.loading = true
				return m, m.promoteSecretVersionCmd(m.secretPromoteSecretID, m.secretPromoteVersionID)
			}
			m.showSecretPromoteConfirm = false
			m.secretPromoteSecretID = ""
			m.secretPromoteVersionID = ""
			m.secretPromoteVersionLabel = ""
			m.secretPromoteCurrentID = ""
			m.secretPromoteCurrentLabel = ""
			return m, nil
		}

		if m.showSecretValueModal {
			switch msg.String() {
			case "esc", "enter":
				m.showSecretValueModal = false
				return m, nil
			case "c":
				text := strings.TrimSpace(m.secretValueCopyText())
				if text == "" {
					return m, m.setStatusMessage("Nothing to copy")
				}
				m.secretClipboardText = text
				m.showSecretClipboardConfirm = true
				return m, nil
			case "e", "u":
				m.showSecretValueModal = false
				m.showSecretUpdateModal = true
				m.secretUpdateStep = 0
				m.secretUpdateValueInput.SetValue(m.secretValueDisplay)
				m.secretUpdateValueInput.CursorEnd()
				m.secretUpdateValueInput.Focus()
				m.configureSecretsLayout()
				return m, nil
			case "j", "down":
				m.secretValueViewport.LineDown(1)
				return m, nil
			case "k", "up":
				m.secretValueViewport.LineUp(1)
				return m, nil
			case "pgdown":
				m.secretValueViewport.HalfPageDown()
				return m, nil
			case "pgup":
				m.secretValueViewport.HalfPageUp()
				return m, nil
			}
			var cmd tea.Cmd
			m.secretValueViewport, cmd = m.secretValueViewport.Update(msg)
			return m, cmd
		}

		if m.showSecretClipboardConfirm {
			switch msg.String() {
			case "y", "Y":
				m.showSecretClipboardConfirm = false
				return m, m.copySecretValueCmd()
			default:
				m.showSecretClipboardConfirm = false
				return m, nil
			}
		}

		if m.showExportModal {
			switch msg.String() {
			case "esc":
				m.showExportModal = false
				m.exportPathInput.Blur()
				m.exportPathInput.SetValue("")
				return m, nil
			case "enter":
				path := m.exportPathInput.Value()
				m.showExportModal = false
				m.exportPathInput.Blur()
				m.exportPathInput.SetValue("")
				if path != "" {
					m.loading = true
					return m, m.exportProfileCmd(path)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.exportPathInput, cmd = m.exportPathInput.Update(msg)
				return m, cmd
			}
		}

		if m.showImportModal {
			switch msg.String() {
			case "esc":
				m.showImportModal = false
				m.importPathInput.Blur()
				m.importPathInput.SetValue("")
				return m, nil
			case "enter":
				path := m.importPathInput.Value()
				m.showImportModal = false
				m.importPathInput.Blur()
				m.importPathInput.SetValue("")
				if path != "" {
					m.loading = true
					return m, m.importProfileCmd(path)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.importPathInput, cmd = m.importPathInput.Update(msg)
				return m, cmd
			}
		}

		if m.showSqsBatchSubModal {
			switch msg.String() {
			case "esc":
				m.showSqsBatchSubModal = false
				return m, nil
			case "up", "k":
				if m.sqsBatchSubCursor > 0 {
					m.sqsBatchSubCursor--
				}
				return m, nil
			case "down", "j":
				if m.sqsBatchSubCursor < len(m.sqsBatchSubList)-1 {
					m.sqsBatchSubCursor++
				}
				return m, nil
			case " ":
				if m.sqsBatchSubCursor < len(m.sqsBatchSubList) {
					m.sqsBatchSubList[m.sqsBatchSubCursor].checked =
						!m.sqsBatchSubList[m.sqsBatchSubCursor].checked
				}
				return m, nil
			case "enter":
				var selected []toggleOption
				for _, opt := range m.sqsBatchSubList {
					if opt.checked {
						selected = append(selected, opt)
					}
				}
				m.showSqsBatchSubModal = false
				if len(selected) > 0 && len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
					qARN := m.queues[m.selectedQueueIndex].ARN
					m.loading = true
					return m, m.batchSubscribeSQSCmd(selected, qARN)
				}
				return m, nil
			}
			return m, nil
		}

		if m.showPeekModal {
			if msg.String() == "esc" || msg.String() == "enter" {
				m.showPeekModal = false
				m.peekMessages = nil
				return m, nil
			}
			return m, nil
		}

		if m.showHelpModal {
			switch msg.String() {
			case "esc":
				m.showHelpModal = false
				return m, nil
			case "up", "k":
				m.helpViewport.LineUp(3)
				return m, nil
			case "down", "j":
				m.helpViewport.LineDown(3)
				return m, nil
			case "ctrl+p", "?":
				m.showHelpModal = false
				return m, nil
			}
			return m, nil
		}

		if m.showLogsModal {
			switch msg.String() {
			case "esc", "o":
				m.showLogsModal = false
				m.clearSelection()
				return m, nil
			case "up", "k":
				if m.commandLogCursor > 0 {
					m.commandLogCursor--
					m.refreshLogViewport()
				}
				return m, nil
			case "down", "j":
				if m.commandLogCursor < len(m.commandLogs)-1 {
					m.commandLogCursor++
					m.refreshLogViewport()
				}
				return m, nil
			case "space":
				return m, m.toggleSelection()
			case "y":
				return m, m.copySelectedTextCmd()
			}
			return m, nil
		}

		if m.showInspectionModal {
			switch msg.String() {
			case "esc", "enter":
				m.showInspectionModal = false
				return m, nil
			case "up", "k":
				m.inspectionViewport.LineUp(1)
				return m, nil
			case "down", "j":
				m.inspectionViewport.LineDown(1)
				return m, nil
			case "pgup", "b":
				m.inspectionViewport.HalfViewUp()
				return m, nil
			case "pgdown", "f":
				m.inspectionViewport.HalfViewDown()
				return m, nil
			case "o":
				m.showInspectionModal = false
				m.showLogsModal = true
				m.refreshLogViewport()
				return m, nil
			}
			return m, nil
		}

		if m.activeTab == panelConfig {
			if m.settingsEditMode {

				switch msg.String() {
				case "esc":
					m.settingsInputs[m.focusedInput].Blur()
					m.settingsEditMode = false
					m.syncSettingsInputsFromConfig()
					return m, nil
				case "enter":
					m.settingsInputs[m.focusedInput].Blur()
					m.settingsEditMode = false
					newCfg := m.configFromSettingsInputs()
					m.loading = true
					return m, m.saveConfigCmd(newCfg)
				case "tab":
					m.settingsInputs[m.focusedInput].Blur()
					m.focusedInput = (m.focusedInput + 1) % len(m.settingsInputs)
					m.settingsInputs[m.focusedInput].Focus()
					return m, nil
				case "shift+tab":
					m.settingsInputs[m.focusedInput].Blur()
					m.focusedInput = (m.focusedInput - 1 + len(m.settingsInputs)) % len(m.settingsInputs)
					m.settingsInputs[m.focusedInput].Focus()
					return m, nil
				default:
					var cmd tea.Cmd
					m.settingsInputs[m.focusedInput], cmd = m.settingsInputs[m.focusedInput].Update(msg)
					return m, cmd
				}
			} else {

				switch msg.String() {
				case "enter":
					m.settingsEditMode = true
					m.settingsInputs[m.focusedInput].Focus()
					return m, nil
				case "down", "j", "tab":
					m.focusedInput = (m.focusedInput + 1) % len(m.settingsInputs)
					return m, nil
				case "up", "k", "shift+tab":
					m.focusedInput = (m.focusedInput - 1 + len(m.settingsInputs)) % len(m.settingsInputs)
					return m, nil
				case "s":
					newCfg := m.configFromSettingsInputs()
					m.loading = true
					return m, m.saveConfigCmd(newCfg)
				}

				if key.Matches(msg, keys.ProfileExport) {
					m.exportPathInput.SetValue(usecase.DefaultSnapshotPath())
					m.showExportModal = true
					m.exportPathInput.Focus()
					return m, nil
				}
				if key.Matches(msg, keys.ProfileImport) {
					m.importPathInput.SetValue(usecase.DefaultSnapshotPath())
					m.showImportModal = true
					m.importPathInput.Focus()
					return m, nil
				}
			}
		}

		if m.activeTab == panelSecrets && !m.showSecretCreateModal && !m.showSecretUpdateModal && !m.showSecretDeleteConfirm && !m.showSecretRestoreConfirm && !m.showSecretPromoteConfirm && !m.showSecretClipboardConfirm && !m.showSecretValueModal {
			switch msg.String() {
			case "up", "k":
				m.moveSelectionUp()
				if m.secretsFocus == focusSecrets {
					return m, m.triggerSubpanelLoadCmd()
				}
				return m, nil
			case "down", "j":
				m.moveSelectionDown()
				if m.secretsFocus == focusSecrets {
					return m, m.triggerSubpanelLoadCmd()
				}
				return m, nil
			case "right", "l":
				if m.secretsFocus == focusSecrets && len(m.secretVersions) > 0 {
					m.secretsFocus = focusSecretVersions
					if m.selectedSecretVersionIndex >= len(m.secretVersions) {
						m.selectedSecretVersionIndex = 0
					}
					return m, nil
				}
			case "left", "h", "esc":
				if m.secretsFocus == focusSecretVersions {
					m.secretsFocus = focusSecrets
					return m, nil
				}
			case "enter", "v":
				if m.secretsFocus == focusSecrets && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
					m.configureSecretsLayout()
					m.showSecretValueModal = true
					return m, nil
				}
				if m.secretsFocus == focusSecretVersions && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) && len(m.secretVersions) > 0 && m.selectedSecretVersionIndex < len(m.secretVersions) {
					version := m.secretVersions[m.selectedSecretVersionIndex]
					if m.secretVersionIsCurrent(version) {
						return m, m.setStatusMessage("Version is already current")
					}
					m.secretPromoteSecretID = m.secrets[m.selectedSecretIndex].ARN
					m.secretPromoteVersionID = version.VersionID
					m.secretPromoteVersionLabel = secretVersionVisualLabel(m.selectedSecretVersionIndex, version)
					m.secretPromoteCurrentID = m.currentSecretVersionID()
					if idx := m.currentSecretCurrentVersionIndex(); idx >= 0 {
						m.secretPromoteCurrentLabel = secretVersionVisualLabel(idx, m.secretVersions[idx])
					} else {
						m.secretPromoteCurrentLabel = ""
					}
					m.showSecretPromoteConfirm = true
					return m, nil
				}
			case "c":
				m.configureSecretsLayout()
				m.showSecretCreateModal = true
				m.secretCreateStep = 0
				m.secretCreateNameInput.SetValue("")
				m.secretCreateValueInput.SetValue("")
				m.secretCreateNameInput.Focus()
				m.secretCreateValueInput.Blur()
				return m, nil
			case "u":
				if m.secretsFocus == focusSecrets && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
					m.configureSecretsLayout()
					m.showSecretUpdateModal = true
					m.secretUpdateStep = 0
					m.secretUpdateValueInput.SetValue(m.secretValueDisplay)
					m.secretUpdateValueInput.CursorEnd()
					m.secretUpdateValueInput.Focus()
					return m, nil
				}
			case "d":
				if m.secretsFocus == focusSecrets && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
					m.secretDeleteID = m.secrets[m.selectedSecretIndex].ARN
					m.secretDeleteName = m.secrets[m.selectedSecretIndex].Name
					m.showSecretDeleteConfirm = true
					return m, nil
				}
			case "r":
				if m.secretsFocus == focusSecrets && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
					m.secretDeleteID = m.secrets[m.selectedSecretIndex].ARN
					m.secretDeleteName = m.secrets[m.selectedSecretIndex].Name
					m.showSecretRestoreConfirm = true
					return m, nil
				}
				if m.secretsFocus == focusSecretVersions && len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) && len(m.secretVersions) > 0 && m.selectedSecretVersionIndex < len(m.secretVersions) {
					version := m.secretVersions[m.selectedSecretVersionIndex]
					if m.secretVersionIsCurrent(version) {
						return m, m.setStatusMessage("Version is already current")
					}
					m.secretPromoteSecretID = m.secrets[m.selectedSecretIndex].ARN
					m.secretPromoteVersionID = version.VersionID
					m.secretPromoteVersionLabel = secretVersionVisualLabel(m.selectedSecretVersionIndex, version)
					m.secretPromoteCurrentID = m.currentSecretVersionID()
					if idx := m.currentSecretCurrentVersionIndex(); idx >= 0 {
						m.secretPromoteCurrentLabel = secretVersionVisualLabel(idx, m.secretVersions[idx])
					} else {
						m.secretPromoteCurrentLabel = ""
					}
					m.showSecretPromoteConfirm = true
					return m, nil
				}
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case msg.String() == "ctrl+p" || msg.String() == "?":
			m.showHelpModal = !m.showHelpModal
			return m, nil

		case key.Matches(msg, keys.CommandLog):
			m.showLogsModal = true
			m.refreshLogViewport()
			return m, nil

		case msg.String() == "1" || msg.String() == "2" || msg.String() == "3" || msg.String() == "4" || msg.String() == "5":
			if tabIndex := int(msg.String()[0] - '0'); tabIndex >= 1 && tabIndex <= 5 {
				if panel, ok := m.panelForTabIndex(tabIndex); ok {
					return m, m.activatePanel(panel)
				}
			}

		case msg.String() == "space":
			return m, m.toggleSelection()

		case msg.String() == "y":
			return m, m.copySelectedTextCmd()

		case msg.String() == "esc" && m.selectionActive:
			m.clearSelection()
			return m, nil

		case msg.String() == "<" || msg.String() == ",":
			if m.activeTab != panelConfig {
				m.setActivePanelRatio(m.leftPanelRatio - 0.05)
				if m.config != nil {
					m.config.LeftPanelRatio = m.leftPanelRatio
					m.loading = true
					return m, m.saveConfigCmd(m.config)
				}
			}
			return m, nil

		case msg.String() == ">" || msg.String() == ".":
			if m.activeTab != panelConfig {
				m.setActivePanelRatio(m.leftPanelRatio + 0.05)
				if m.config != nil {
					m.config.LeftPanelRatio = m.leftPanelRatio
					m.loading = true
					return m, m.saveConfigCmd(m.config)
				}
			}
			return m, nil

		case key.Matches(msg, keys.TabConfig):
			m.activeTab = panelConfig
			m.errorMsg = ""
			m.statusMsg = ""
			m.clearSelection()
			return m, nil

		case key.Matches(msg, keys.Up):
			m.moveSelectionUp()
			return m, m.triggerSubpanelLoadCmd()

		case key.Matches(msg, keys.Down):
			m.moveSelectionDown()
			return m, m.triggerSubpanelLoadCmd()

		case key.Matches(msg, keys.Right):
			switch m.activeTab {
			case panelS3:
				if m.s3Focus == focusBuckets {
					m.s3Focus = focusObjects
					m.selectedObjectIndex = 0
				}
			case panelSQS:
				if m.sqsFocus == focusQueues && len(m.queueSubscriptions) > 0 {
					m.sqsFocus = focusQueueSubs
					m.selectedQueueSubIndex = 0
				}
			case panelSNS:
				if m.snsSubFocus == focusTopics && len(m.topics) > 0 {
					m.snsSubFocus = focusSubs
					m.selectedSubIndex = 0
				}
			}

		case key.Matches(msg, keys.Enter):
			switch m.activeTab {
			case panelSQS, panelSNS:
				m.loading = true
				return m, m.openInspectionCmd()
			case panelS3:
				if m.s3Focus == focusBuckets {
					m.s3Focus = focusObjects
					m.selectedObjectIndex = 0
				}
			}

		case key.Matches(msg, keys.Left) || key.Matches(msg, keys.Esc):
			switch m.activeTab {
			case panelS3:
				if m.s3Focus == focusObjects {
					m.s3Focus = focusBuckets
				}
			case panelSQS:
				if m.sqsFocus == focusQueueSubs {
					m.sqsFocus = focusQueues
				}
			case panelSNS:
				if m.snsSubFocus == focusSubs {
					m.snsSubFocus = focusTopics
				}
			}

		case key.Matches(msg, keys.S3Delete) || key.Matches(msg, keys.SQSDelete) || key.Matches(msg, keys.SNSDelete) || key.Matches(msg, keys.SQSSubDelete):
			switch m.activeTab {
			case panelS3:
				if m.s3Focus == focusObjects && len(m.objects) > 0 {
					m.s3DeleteBucket = m.buckets[m.selectedBucketIndex].Name
					m.s3DeleteKey = m.objects[m.selectedObjectIndex].Key
					m.s3DeleteIsBucket = false
					m.showS3ConfirmDelete = true
					return m, nil
				} else if m.s3Focus == focusBuckets && len(m.buckets) > 0 {
					m.s3DeleteBucket = m.buckets[m.selectedBucketIndex].Name
					m.s3DeleteKey = ""
					m.s3DeleteIsBucket = true
					m.showS3ConfirmDelete = true
					return m, nil
				}
			case panelSQS:
				if m.sqsFocus == focusQueues && len(m.queues) > 0 {
					q := m.queues[m.selectedQueueIndex]
					m.sqsDeleteQueueURL = q.URL
					m.sqsDeleteQueueName = q.Name
					m.showSqsConfirmDelete = true
					return m, nil
				} else if m.sqsFocus == focusQueueSubs && len(m.queueSubscriptions) > 0 && m.selectedQueueSubIndex < len(m.queueSubscriptions) {
					sub := m.queueSubscriptions[m.selectedQueueSubIndex]
					m.sqsDeleteSubARN = sub.ARN
					m.sqsDeleteSubLabel = sub.TopicARN
					m.showSqsSubDeleteConfirm = true
					return m, nil
				}
			case panelSNS:
				if m.snsSubFocus == focusTopics && len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
					m.snsDeleteARN = m.topics[m.selectedTopicIndex].ARN
					m.snsDeleteLabel = m.topics[m.selectedTopicIndex].Name
					m.snsDeleteIsTopic = true
					m.showSnsConfirmDelete = true
					return m, nil
				} else if m.snsSubFocus == focusSubs && len(m.subscriptions) > 0 && m.selectedSubIndex < len(m.subscriptions) {
					m.snsDeleteARN = m.subscriptions[m.selectedSubIndex].ARN
					m.snsDeleteLabel = m.subscriptions[m.selectedSubIndex].Endpoint
					m.snsDeleteIsTopic = false
					m.showSnsConfirmDelete = true
					return m, nil
				}
			}

		case key.Matches(msg, keys.S3Presign) && m.activeTab == panelS3 && m.s3Focus == focusObjects && len(m.objects) > 0:
			bucket := m.buckets[m.selectedBucketIndex].Name
			keyStr := m.objects[m.selectedObjectIndex].Key

			return m, m.openPresignedURLCmd(bucket, keyStr)

		case key.Matches(msg, keys.S3Download):
			if m.activeTab == panelS3 && m.s3Focus == focusObjects && len(m.objects) > 0 {
				bucket := m.buckets[m.selectedBucketIndex].Name
				keyStr := m.objects[m.selectedObjectIndex].Key
				return m, tea.Batch(m.setStatusMessage("Starting S3 download..."), m.downloadS3ObjectCmd(bucket, keyStr))
			}

		case key.Matches(msg, keys.S3Upload):
			if m.activeTab == panelS3 && len(m.buckets) > 0 {
				m.showS3UploadModal = true
				m.s3UploadFocus = 0
				m.s3UploadPathInput.Focus()
				m.s3UploadKeyInput.Blur()
				if m.selectedObjectIndex < len(m.objects) && m.s3Focus == focusObjects {
					m.s3UploadKeyInput.SetValue(m.objects[m.selectedObjectIndex].Key)
				} else {
					m.s3UploadKeyInput.SetValue("")
				}
				return m, nil
			}

		case key.Matches(msg, keys.S3Preview):
			if m.activeTab == panelS3 && m.s3Focus == focusObjects && len(m.objects) > 0 {
				m.showS3PreviewModal = true
				return m, nil
			}

		case key.Matches(msg, keys.S3Folder):
			if m.activeTab == panelS3 && m.s3Focus == focusObjects && len(m.buckets) > 0 {
				m.showS3CreateFolderModal = true
				if m.selectedObjectIndex < len(m.objects) {
					keyStr := m.objects[m.selectedObjectIndex].Key
					if idx := strings.LastIndex(keyStr, "/"); idx >= 0 {
						m.s3FolderInput.SetValue(keyStr[:idx+1])
					} else {
						m.s3FolderInput.SetValue("")
					}
				} else {
					m.s3FolderInput.SetValue("")
				}
				m.s3FolderInput.Focus()
				return m, nil
			}

		case key.Matches(msg, keys.SQSPurge):
			if m.activeTab == panelSQS && len(m.queues) > 0 {
				m.loading = true
				return m, m.purgeSQSQueueCmd(m.queues[m.selectedQueueIndex].URL)
			}

		case key.Matches(msg, keys.SQSPurgeAll):
			if m.activeTab == panelSQS && len(m.queues) > 0 {
				m.showSqsPurgeAllConfirm = true
				return m, nil
			}

		case key.Matches(msg, keys.SQSView):
			if m.activeTab == panelSQS && len(m.queues) > 0 {
				qURL := m.queues[m.selectedQueueIndex].URL
				m.loading = true
				return m, m.peekSQSMessagesCmd(qURL)
			}

		case key.Matches(msg, keys.SQSSend):
			if m.activeTab == panelSQS && len(m.queues) > 0 {
				m.showSqsSendModal = true
				m.sqsSendMessageInput.Focus()
				return m, nil
			} else if m.activeTab == panelSNS && len(m.topics) > 0 {
				m.showSnsPublishModal = true
				m.snsPublishInput.Focus()

				var attrHints []string
				for _, sub := range m.subscriptions {
					for attr, vals := range sub.FilterPolicy {
						attrHints = append(attrHints, attr+"="+strings.Join(vals, ","))
					}
				}
				if len(attrHints) > 0 {
					m.snsPublishAttrsInput.SetValue(strings.Join(attrHints, ", "))
				} else {
					m.snsPublishAttrsInput.SetValue("")
				}
				return m, nil
			}

		case key.Matches(msg, keys.S3Create) || key.Matches(msg, keys.SQSCreate) || key.Matches(msg, keys.SNSCreate):
			switch m.activeTab {
			case panelS3:
				if m.s3Focus == focusBuckets {
					m.showS3CreateModal = true
					m.s3CreateInput.Focus()
					return m, nil
				}
			case panelSQS:
				m.showSqsCreateModal = true
				m.sqsCreateInput.Focus()
				return m, nil
			case panelSNS:
				if m.snsSubFocus == focusTopics {
					m.showSnsCreateTopicModal = true
					m.snsCreateTopicInput.Focus()
					return m, nil
				} else if m.snsSubFocus == focusSubs && len(m.topics) > 0 {
					return m, m.setErrorMessage("Direct SNS topic-to-topic subscriptions are not supported. Subscribe the destination SQS queue from the SQS panel.")
				}
			}

		case key.Matches(msg, keys.SNSEdit):
			if m.activeTab == panelSNS && m.snsSubFocus == focusSubs && len(m.subscriptions) > 0 && m.selectedSubIndex < len(m.subscriptions) {
				sub := m.subscriptions[m.selectedSubIndex]
				currentVal := ""
				if et, ok := sub.FilterPolicy["event_type"]; ok {
					currentVal = strings.Join(et, ", ")
				}
				m.snsSubEditInput.SetValue(currentVal)
				m.showSnsSubEditModal = true
				m.snsSubEditInput.Focus()
				return m, nil
			}

		case key.Matches(msg, keys.SNSBatch) || key.Matches(msg, keys.SQSBatchSubscribe):
			switch m.activeTab {
			case panelSNS:
				if m.snsSubFocus == focusSubs && len(m.topics) > 0 {
					return m, m.setErrorMessage("Direct SNS topic-to-topic subscriptions are not supported. Subscribe the destination SQS queue from the SQS panel.")
				}
			case panelSQS:
				if len(m.queues) > 0 && len(m.topics) > 0 {
					var opts []toggleOption
					for _, t := range m.topics {
						opts = append(opts, toggleOption{
							label: t.Name,
							arn:   t.ARN,
						})
					}
					m.sqsBatchSubList = opts
					m.sqsBatchSubCursor = 0
					m.showSqsBatchSubModal = true
					return m, nil
				}
			}

		case key.Matches(msg, keys.SNSImportYaml):
			if m.activeTab == panelSNS && m.snsSubFocus == focusSubs {
				if len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
					m.snsYamlCurrentTopicARN = m.topics[m.selectedTopicIndex].ARN
				}
				content := ""
				if m.snsYamlSavedContent != nil {
					content = m.snsYamlSavedContent[m.snsYamlCurrentTopicARN]
				}
				m.snsYamlImportTextarea.SetValue(content)

				taWidth := m.width - 14
				if taWidth < 40 {
					taWidth = 40
				}
				if taWidth > 100 {
					taWidth = 100
				}
				m.snsYamlImportTextarea.SetWidth(taWidth)
				taHeight := m.height - 14
				if taHeight < 8 {
					taHeight = 8
				}
				if taHeight > 24 {
					taHeight = 24
				}
				m.snsYamlImportTextarea.SetHeight(taHeight)

				m.showSnsYamlImportModal = true
				m.snsYamlImportTextarea.Focus()
				return m, m.loadYamlScriptCmd(m.snsYamlCurrentTopicARN)
			}

		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) reloadTabCmd() tea.Cmd {
	switch m.activeTab {
	case panelS3:
		if len(m.buckets) > 0 {
			return nil
		}
		return m.loadS3BucketsCmd()
	case panelSQS:
		return m.loadSQSQueuesCmd()
	case panelSNS:
		return m.loadSNSTopicsCmd()
	case panelSecrets:
		return m.loadSecretsCmd()
	}
	return nil
}

func (m *Model) moveSelectionUp() {
	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusBuckets {
			if m.selectedBucketIndex > 0 {
				m.selectedBucketIndex--
			}
		} else {
			if m.selectedObjectIndex > 0 {
				m.selectedObjectIndex--
			}
		}
	case panelSQS:
		if m.sqsFocus == focusQueues {
			if m.selectedQueueIndex > 0 {
				m.selectedQueueIndex--
			}
		} else if m.sqsFocus == focusQueueSubs {
			if m.selectedQueueSubIndex > 0 {
				m.selectedQueueSubIndex--
			}
		}
	case panelSNS:
		if m.snsSubFocus == focusTopics {
			if m.selectedTopicIndex > 0 {
				m.selectedTopicIndex--
			}
		} else {
			if m.selectedSubIndex > 0 {
				m.selectedSubIndex--
			}
		}
	case panelSecrets:
		if m.secretsFocus == focusSecretVersions {
			if m.selectedSecretVersionIndex > 0 {
				m.selectedSecretVersionIndex--
			}
		} else {
			if m.selectedSecretIndex > 0 {
				m.selectedSecretIndex--
			}
		}
	}
}

func (m *Model) moveSelectionDown() {
	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusBuckets {
			if m.selectedBucketIndex < len(m.buckets)-1 {
				m.selectedBucketIndex++
			}
		} else {
			if m.selectedObjectIndex < len(m.objects)-1 {
				m.selectedObjectIndex++
			}
		}
	case panelSQS:
		if m.sqsFocus == focusQueues {
			if m.selectedQueueIndex < len(m.queues)-1 {
				m.selectedQueueIndex++
			}
		} else if m.sqsFocus == focusQueueSubs {
			if m.selectedQueueSubIndex < len(m.queueSubscriptions)-1 {
				m.selectedQueueSubIndex++
			}
		}
	case panelSNS:
		if m.snsSubFocus == focusTopics {
			if m.selectedTopicIndex < len(m.topics)-1 {
				m.selectedTopicIndex++
			}
		} else {
			if m.selectedSubIndex < len(m.subscriptions)-1 {
				m.selectedSubIndex++
			}
		}
	case panelSecrets:
		if m.secretsFocus == focusSecretVersions {
			if m.selectedSecretVersionIndex < len(m.secretVersions)-1 {
				m.selectedSecretVersionIndex++
			}
		} else {
			if m.selectedSecretIndex < len(m.secrets)-1 {
				m.selectedSecretIndex++
			}
		}
	}
}

func (m *Model) triggerSubpanelLoadCmd() tea.Cmd {
	switch m.activeTab {
	case panelS3:
		if m.s3Focus == focusBuckets && len(m.buckets) > 0 && m.selectedBucketIndex < len(m.buckets) {
			bucket := m.buckets[m.selectedBucketIndex].Name
			if m.s3ObjectsLoadedFor == bucket {
				m.objects = m.s3ObjectsCache[bucket]
				return nil
			}
			return m.loadS3ObjectsCmd(bucket)
		}
	case panelSQS:
		if m.sqsFocus == focusQueues && len(m.queues) > 0 && m.selectedQueueIndex < len(m.queues) {
			q := m.queues[m.selectedQueueIndex]
			return m.loadSQSQueueSubscriptionsCmd(q.URL, q.ARN)
		}
	case panelSNS:
		if m.snsSubFocus == focusTopics && len(m.topics) > 0 && m.selectedTopicIndex < len(m.topics) {
			topicARN := m.topics[m.selectedTopicIndex].ARN
			return tea.Batch(
				m.loadSNSSubscriptionsCmd(topicARN),
				m.loadManagedSubscriptionsCmd(),
			)
		}
	case panelSecrets:
		if len(m.secrets) > 0 && m.selectedSecretIndex < len(m.secrets) {
			secretID := m.secrets[m.selectedSecretIndex].ARN
			if secretID != m.secretDetailsLoadedFor {
				return m.loadSecretDetailsCmd(secretID)
			}
		}
	}
	return nil
}

func (m Model) selectedS3BucketName() string {
	if m.selectedBucketIndex >= 0 && m.selectedBucketIndex < len(m.buckets) {
		return m.buckets[m.selectedBucketIndex].Name
	}
	return ""
}

func (m *Model) removeManagedSubByARN(arn string) tea.Cmd {
	return func() tea.Msg {
		var updated []domain.ManagedSubscription
		for _, s := range m.managedSubs {
			if s.SubscriptionARN != arn && s.DestinationARN != arn {
				updated = append(updated, s)
			}
		}
		err := m.configUseCase.SaveSubscriptions(updated)
		if err != nil {
			return errMsg{Error: err}
		}
		return managedSubscriptionsUpdatedMsg{}
	}
}

func (m *Model) openPresignedURLCmd(bucket, keyStr string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		url, err := m.awsUseCase.GetS3PresignedURL(ctx, m.config, bucket, keyStr)
		if err != nil {
			return errMsg{Error: err}
		}

		if err := openURL(url); err != nil {
			return errMsg{Error: err}
		}

		return statusMsg{Message: "Object link opened in your default browser!"}
	}
}

func parseMessageAttributes(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	attrs := make(map[string]string)
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(p[:eq])
		val := strings.TrimSpace(p[eq+1:])
		if key != "" {
			attrs[key] = val
		}
	}
	return attrs
}
