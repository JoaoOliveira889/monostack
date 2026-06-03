package config

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"

	"monostack/internal/domain"
)

const DefaultPanelRatio = 0.5

func NormalizePanelRatio(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return DefaultPanelRatio
	}
	if value <= 0.05 || value >= 0.95 {
		return DefaultPanelRatio
	}
	return value
}

func ClonePanelRatios(values map[string]float64) map[string]float64 {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]float64, len(values))
	for key, value := range values {
		cloned[key] = NormalizePanelRatio(value)
	}
	return cloned
}

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
			ServiceName:     "MiniStack",
			EndpointURL:     "http://localhost:4566",
			Region:          "us-east-1",
			UseMock:         false,
			LeftPanelRatio:  DefaultPanelRatio,
			EnabledServices: domain.DefaultEnabledServices(),
			SnapshotPath:    "",
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

	cfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
	cfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	cfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)

	if cfg.SnapshotPath == "" {
		cfg.SnapshotPath = ""
	}

	return &cfg, nil
}

func (s *FileConfigStore) Save(cfg *domain.AWSConfig) error {
	if cfg != nil {
		cfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)
		cfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
		cfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	}
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
