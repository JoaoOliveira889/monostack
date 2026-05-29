package domain

import "strings"

type AWSConfig struct {
	ServiceName     string             `json:"service_name"`
	EndpointURL     string             `json:"endpoint_url"`
	Region          string             `json:"region"`
	AccessKeyID     string             `json:"access_key_id"`
	SecretAccessKey string             `json:"secret_access_key"`
	UseMock         bool               `json:"use_mock"`
	LeftPanelRatio  float64            `json:"left_panel_ratio"`
	PanelRatios     map[string]float64 `json:"panel_ratios,omitempty"`
	EnabledServices []string           `json:"enabled_services,omitempty"`
	SnapshotPath    string             `json:"snapshot_path,omitempty"`
}

const (
	ServiceS3      = "s3"
	ServiceSQS     = "sqs"
	ServiceSNS     = "sns"
	ServiceSecrets = "secrets"
)

func DefaultEnabledServices() []string {
	return []string{ServiceS3, ServiceSQS, ServiceSNS, ServiceSecrets}
}

func NormalizeEnabledServices(values []string) []string {
	allowed := map[string]struct{}{
		ServiceS3:      {},
		ServiceSQS:     {},
		ServiceSNS:     {},
		ServiceSecrets: {},
	}

	ordered := make([]string, 0, len(allowed))
	seen := map[string]struct{}{}
	for _, candidate := range values {
		value := strings.ToLower(strings.TrimSpace(candidate))
		if value == "" {
			continue
		}
		if _, ok := allowed[value]; !ok {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ordered = append(ordered, value)
	}

	if len(ordered) == 0 {
		return DefaultEnabledServices()
	}

	return ordered
}

func ServiceEnabled(values []string, service string) bool {
	service = strings.ToLower(strings.TrimSpace(service))
	for _, value := range values {
		if strings.EqualFold(value, service) {
			return true
		}
	}
	return false
}

type ConfigStore interface {
	Load() (*AWSConfig, error)
	Save(cfg *AWSConfig) error
}

type ManagedSubscription struct {
	Name            string   `json:"name"`
	TopicARN        string   `json:"topic_arn"`
	DestinationARN  string   `json:"destination_arn"`
	DestinationType string   `json:"destination_type"`
	EventTypes      []string `json:"event_types"`
	FilterScope     string   `json:"filter_scope,omitempty"`
	SubscriptionARN string   `json:"subscription_arn,omitempty"`
}

type SubscriptionStore interface {
	LoadAll() ([]ManagedSubscription, error)
	SaveAll(subs []ManagedSubscription) error
}

type SubscriptionScript struct {
	Version            int                       `yaml:"version"`
	DefaultQueue       string                    `yaml:"default_queue,omitempty"`
	DefaultFilterScope string                    `yaml:"default_filter_scope,omitempty"`
	Subscriptions      []SubscriptionScriptEntry `yaml:"subscriptions"`
}

type SubscriptionScriptEntry struct {
	Name        string   `yaml:"name"`
	Topic       string   `yaml:"topic,omitempty"`
	Queue       string   `yaml:"queue,omitempty"`
	EventType   []string `yaml:"event_type"`
	FilterScope string   `yaml:"filter_scope,omitempty"`
}

const (
	SubscriptionFilterScopeMessageAttributes = "message_attributes"
	SubscriptionFilterScopeMessageBody       = "message_body"
)
