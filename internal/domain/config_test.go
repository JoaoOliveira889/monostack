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
		LeftPanelRatio: 0.5,
		PanelRatios: map[string]float64{
			ServiceS3: 0.6,
		},
	}

	if cfg.ServiceName != "test" {
		t.Errorf("expected ServiceName 'test', got %q", cfg.ServiceName)
	}
	if cfg.LeftPanelRatio != 0.5 {
		t.Errorf("expected LeftPanelRatio 0.5, got %f", cfg.LeftPanelRatio)
	}
	if cfg.PanelRatios[ServiceS3] != 0.6 {
		t.Fatalf("expected PanelRatios for %q to be 0.6, got %f", ServiceS3, cfg.PanelRatios[ServiceS3])
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
		PanelRatios: map[string]float64{
			ServiceS3:      0.6,
			ServiceSecrets: 0.4,
		},
		EnabledServices: []string{ServiceS3, ServiceSNS},
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
	if len(decoded.PanelRatios) != len(original.PanelRatios) {
		t.Fatalf("expected PanelRatios %v, got %v", original.PanelRatios, decoded.PanelRatios)
	}
	if len(decoded.EnabledServices) != len(original.EnabledServices) {
		t.Fatalf("expected EnabledServices %v, got %v", original.EnabledServices, decoded.EnabledServices)
	}
}

func TestNormalizeEnabledServices(t *testing.T) {
	got := NormalizeEnabledServices([]string{"S3", "sns", "invalid", "sqs", "sns"})
	expected := []string{ServiceS3, ServiceSNS, ServiceSQS}

	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, got)
		}
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

func (m *mockConfigStore) ListProfiles() ([]string, error) {
	return nil, nil
}

func (m *mockConfigStore) SwitchProfile(name string) error {
	return nil
}

func (m *mockConfigStore) SaveProfile(name string, cfg *AWSConfig) error {
	return nil
}

func (m *mockConfigStore) DeleteProfile(name string) error {
	return nil
}
