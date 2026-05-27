package domain

type AWSConfig struct {
	ServiceName     string  `json:"service_name"`
	EndpointURL     string  `json:"endpoint_url"`
	Region          string  `json:"region"`
	AccessKeyID     string  `json:"access_key_id"`
	SecretAccessKey string  `json:"secret_access_key"`
	UseMock         bool    `json:"use_mock"`
	LeftPanelRatio  float64 `json:"left_panel_ratio"`
	SnapshotPath    string  `json:"snapshot_path,omitempty"`
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
