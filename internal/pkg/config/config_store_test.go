package config

import (
	"os"
	"path/filepath"
	"testing"

	"monostack/internal/domain"
)

func TestFileConfigStore_LoadSave(t *testing.T) {
	dir, err := os.MkdirTemp("", "monostack-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "config.json")
	store := &FileConfigStore{filePath: filePath}

	original := &domain.AWSConfig{
		ServiceName:    "test-profile",
		EndpointURL:    "http://localhost:4566",
		Region:         "us-east-1",
		AccessKeyID:    "test-key",
		SecretAccessKey: "test-secret",
		UseMock:        true,
		LeftPanelRatio: 0.4,
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ServiceName != original.ServiceName {
		t.Errorf("expected ServiceName %q, got %q", original.ServiceName, loaded.ServiceName)
	}
	if loaded.EndpointURL != original.EndpointURL {
		t.Errorf("expected EndpointURL %q, got %q", original.EndpointURL, loaded.EndpointURL)
	}
	if loaded.Region != original.Region {
		t.Errorf("expected Region %q, got %q", original.Region, loaded.Region)
	}
	if loaded.UseMock != original.UseMock {
		t.Errorf("expected UseMock %v, got %v", original.UseMock, loaded.UseMock)
	}
}

func TestFileConfigStore_FilePermissions(t *testing.T) {
	dir, err := os.MkdirTemp("", "monostack-perm-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "config.json")
	store := &FileConfigStore{filePath: filePath}

	if err := store.Save(&domain.AWSConfig{}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}

func TestFileConfigStore_LoadDefaultOnMissing(t *testing.T) {
	dir, err := os.MkdirTemp("", "monostack-default-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "nonexistent.json")
	store := &FileConfigStore{filePath: filePath}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load default failed: %v", err)
	}

	if cfg.ServiceName == "" {
		t.Error("expected non-empty ServiceName in default config")
	}
}

func TestFileConfigStore_LeftPanelRatioClamp(t *testing.T) {
	dir, err := os.MkdirTemp("", "monostack-clamp-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "config.json")
	store := &FileConfigStore{filePath: filePath}

	original := &domain.AWSConfig{LeftPanelRatio: 0.99}
	if err := store.Save(original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LeftPanelRatio != 0.3 {
		t.Errorf("expected clamped LeftPanelRatio 0.3, got %f", loaded.LeftPanelRatio)
	}
}

func TestNewFileConfigStore(t *testing.T) {
	store := NewFileConfigStore("test.json")
	if store.filePath == "" {
		t.Error("expected non-empty filePath")
	}
}
