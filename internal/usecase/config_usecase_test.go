package usecase

import (
	"errors"
	"testing"

	"monostack/internal/domain"
)

func TestConfigUseCase_GetConfig(t *testing.T) {
	expectedCfg := &domain.AWSConfig{
		ServiceName: "test",
		Region:      "us-east-1",
	}

	store := &mockConfigStore{
		LoadFunc: func() (*domain.AWSConfig, error) {
			return expectedCfg, nil
		},
	}

	uc := NewConfigUseCase(store)
	cfg, err := uc.GetConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServiceName != expectedCfg.ServiceName {
		t.Errorf("expected ServiceName %q, got %q", expectedCfg.ServiceName, cfg.ServiceName)
	}
}

func TestConfigUseCase_GetConfig_Error(t *testing.T) {
	store := &mockConfigStore{
		LoadFunc: func() (*domain.AWSConfig, error) {
			return nil, errors.New("load error")
		},
	}

	uc := NewConfigUseCase(store)
	_, err := uc.GetConfig()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConfigUseCase_SaveConfig(t *testing.T) {
	var called bool
	store := &mockConfigStore{
		SaveFunc: func(cfg *domain.AWSConfig) error {
			called = true
			if cfg.ServiceName != "saved" {
				t.Errorf("expected ServiceName 'saved', got %q", cfg.ServiceName)
			}
			return nil
		},
	}

	uc := NewConfigUseCase(store)
	err := uc.SaveConfig(&domain.AWSConfig{ServiceName: "saved"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Save to be called")
	}
}

func TestConfigUseCase_SaveConfig_Error(t *testing.T) {
	store := &mockConfigStore{
		SaveFunc: func(cfg *domain.AWSConfig) error {
			return errors.New("save error")
		},
	}

	uc := NewConfigUseCase(store)
	err := uc.SaveConfig(&domain.AWSConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
