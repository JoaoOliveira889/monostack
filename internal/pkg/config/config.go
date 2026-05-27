package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"monostack/internal/domain"
)

type FileConfigStore struct {
	filePath string
}

func NewFileConfigStore(fileName string) *FileConfigStore {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "monostack")
	return &FileConfigStore{
		filePath: filepath.Join(dir, fileName),
	}
}

func (s *FileConfigStore) Load() (*domain.AWSConfig, error) {
	if _, err := os.Stat(s.filePath); errors.Is(err, os.ErrNotExist) {
		defaultConfig := &domain.AWSConfig{
			ServiceName:    "MiniStack",
			EndpointURL:    "http://localhost:4566",
			Region:         "us-east-1",
			UseMock:        false,
			LeftPanelRatio: 0.3,
			SnapshotPath:   "",
		}
		_ = s.Save(defaultConfig)
		return defaultConfig, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var cfg domain.AWSConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.LeftPanelRatio <= 0.05 || cfg.LeftPanelRatio >= 0.95 {
		cfg.LeftPanelRatio = 0.3
	}

	if cfg.SnapshotPath == "" {
		cfg.SnapshotPath = ""
	}

	return &cfg, nil
}

func (s *FileConfigStore) Save(cfg *domain.AWSConfig) error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

var _ domain.ConfigStore = (*FileConfigStore)(nil)
