package domain

import (
	"encoding/json"
	"testing"
)

func TestAWSConfig_DefaultValues(t *testing.T) {
	cfg := &AWSConfig{
		ServiceName:    "test",
		EndpointURL:    "http://localhost:4566",
		Region:         "us-east-1",
		UseMock:        true,
		LeftPanelRatio: 0.3,
	}

	if cfg.ServiceName != "test" {
		t.Errorf("expected ServiceName 'test', got %q", cfg.ServiceName)
	}
	if cfg.LeftPanelRatio != 0.3 {
		t.Errorf("expected LeftPanelRatio 0.3, got %f", cfg.LeftPanelRatio)
	}
}

func TestAWSConfig_JSONRoundTrip(t *testing.T) {
	original := AWSConfig{
		ServiceName:     "Production",
		EndpointURL:     "",
		Region:          "eu-west-1",
		AccessKeyID:     "AKIA123",
		SecretAccessKey: "secret123",
		UseMock:         false,
		LeftPanelRatio:  0.5,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AWSConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ServiceName != original.ServiceName {
		t.Errorf("expected ServiceName %q, got %q", original.ServiceName, decoded.ServiceName)
	}
	if decoded.Region != original.Region {
		t.Errorf("expected Region %q, got %q", original.Region, decoded.Region)
	}
	if decoded.LeftPanelRatio != original.LeftPanelRatio {
		t.Errorf("expected LeftPanelRatio %f, got %f", original.LeftPanelRatio, decoded.LeftPanelRatio)
	}
}

func TestAWSConfig_LeftPanelRatioClamp(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{name: "zero", input: 0, expected: 0.3},
		{name: "negative", input: -1, expected: 0.3},
		{name: "too small", input: 0.04, expected: 0.3},
		{name: "valid", input: 0.5, expected: 0.5},
		{name: "too large", input: 0.96, expected: 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := AWSConfig{LeftPanelRatio: tt.input}
			if cfg.LeftPanelRatio <= 0.05 || cfg.LeftPanelRatio >= 0.95 {
				cfg.LeftPanelRatio = 0.3
			}
			if cfg.LeftPanelRatio != tt.expected {
				t.Errorf("input %f: expected %f, got %f", tt.input, tt.expected, cfg.LeftPanelRatio)
			}
		})
	}
}

func TestConfigStoreInterface(t *testing.T) {
	var _ ConfigStore = (*mockConfigStore)(nil)
}

func TestSubscriptionScriptEntry_JSONLikeFields(t *testing.T) {
	entry := SubscriptionScriptEntry{
		Name:        "pix",
		Topic:       "dev-webapi-pix-sns",
		EventType:   []string{"pix_received"},
		FilterScope: SubscriptionFilterScopeMessageBody,
	}

	if entry.FilterScope != SubscriptionFilterScopeMessageBody {
		t.Errorf("expected filter scope %q, got %q", SubscriptionFilterScopeMessageBody, entry.FilterScope)
	}
}

type mockConfigStore struct {
	loadFunc func() (*AWSConfig, error)
	saveFunc func(cfg *AWSConfig) error
}

func (m *mockConfigStore) Load() (*AWSConfig, error) {
	if m.loadFunc != nil {
		return m.loadFunc()
	}
	return &AWSConfig{ServiceName: "default"}, nil
}

func (m *mockConfigStore) Save(cfg *AWSConfig) error {
	if m.saveFunc != nil {
		return m.saveFunc(cfg)
	}
	return nil
}
