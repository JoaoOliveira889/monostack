package config

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"

	"monostack/internal/domain"
	"monostack/internal/pkg/crypto"
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

	if cfg.Profiles == nil || len(cfg.Profiles) == 0 {
		cfg.Profiles = map[string]*domain.AWSConfig{
			"default": {
				ServiceName:     cfg.ServiceName,
				EndpointURL:     cfg.EndpointURL,
				Region:          cfg.Region,
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
				AccessKeyIDEnc:     cfg.AccessKeyIDEnc,
				SecretAccessKeyEnc: cfg.SecretAccessKeyEnc,
				UseMock:         cfg.UseMock,
				LeftPanelRatio:  cfg.LeftPanelRatio,
				PanelRatios:     ClonePanelRatios(cfg.PanelRatios),
				EnabledServices: append([]string(nil), cfg.EnabledServices...),
				SnapshotPath:    cfg.SnapshotPath,
			},
		}
		cfg.ActiveProfile = "default"
		_ = s.writeRaw(&cfg)
	}

	if cfg.ActiveProfile != "" {
		if profile, ok := cfg.Profiles[cfg.ActiveProfile]; ok && profile != nil {
			s.overlayProfile(&cfg, profile)
		}
	}

	if cfg.AccessKeyIDEnc != "" {
		dec, err := crypto.Decrypt(cfg.AccessKeyIDEnc)
		if err != nil {
			log.Printf("monostack: WARNING: failed to decrypt stored AccessKeyID (machine key may have changed). Please re-enter credentials in Settings.")
			cfg.AccessKeyIDEnc = ""
		} else {
			cfg.AccessKeyID = string(dec)
		}
	}
	if cfg.SecretAccessKeyEnc != "" {
		dec, err := crypto.Decrypt(cfg.SecretAccessKeyEnc)
		if err != nil {
			log.Printf("monostack: WARNING: failed to decrypt stored SecretAccessKey (machine key may have changed). Please re-enter credentials in Settings.")
			cfg.SecretAccessKeyEnc = ""
		} else {
			cfg.SecretAccessKey = string(dec)
		}
	}

	if cfg.SnapshotPath == "" {
		cfg.SnapshotPath = ""
	}

	return &cfg, nil
}

func (s *FileConfigStore) overlayProfile(cfg *domain.AWSConfig, profile *domain.AWSConfig) {
	if profile.ServiceName != "" {
		cfg.ServiceName = profile.ServiceName
	}
	if profile.EndpointURL != "" {
		cfg.EndpointURL = profile.EndpointURL
	}
	if profile.Region != "" {
		cfg.Region = profile.Region
	}
	if profile.AccessKeyID != "" {
		cfg.AccessKeyID = profile.AccessKeyID
	}
	if profile.SecretAccessKey != "" {
		cfg.SecretAccessKey = profile.SecretAccessKey
	}
	if profile.AccessKeyIDEnc != "" {
		cfg.AccessKeyIDEnc = profile.AccessKeyIDEnc
	}
	if profile.SecretAccessKeyEnc != "" {
		cfg.SecretAccessKeyEnc = profile.SecretAccessKeyEnc
	}
	cfg.UseMock = profile.UseMock
	if profile.LeftPanelRatio > 0 {
		cfg.LeftPanelRatio = profile.LeftPanelRatio
	}
	if profile.PanelRatios != nil {
		cfg.PanelRatios = ClonePanelRatios(profile.PanelRatios)
	}
	if len(profile.EnabledServices) > 0 {
		cfg.EnabledServices = append([]string(nil), profile.EnabledServices...)
	}
	if profile.SnapshotPath != "" {
		cfg.SnapshotPath = profile.SnapshotPath
	}
}

func (s *FileConfigStore) syncToProfile(cfg *domain.AWSConfig) {
	if cfg.ActiveProfile == "" || cfg.Profiles == nil {
		return
	}
	profile, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok || profile == nil {
		profile = &domain.AWSConfig{}
		cfg.Profiles[cfg.ActiveProfile] = profile
	}
	profile.ServiceName = cfg.ServiceName
	profile.EndpointURL = cfg.EndpointURL
	profile.Region = cfg.Region
	profile.AccessKeyID = cfg.AccessKeyID
	profile.SecretAccessKey = cfg.SecretAccessKey
	profile.AccessKeyIDEnc = cfg.AccessKeyIDEnc
	profile.SecretAccessKeyEnc = cfg.SecretAccessKeyEnc
	profile.UseMock = cfg.UseMock
	profile.LeftPanelRatio = cfg.LeftPanelRatio
	profile.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	profile.EnabledServices = append([]string(nil), cfg.EnabledServices...)
	profile.SnapshotPath = cfg.SnapshotPath
}

func (s *FileConfigStore) writeRaw(cfg *domain.AWSConfig) error {
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

func (s *FileConfigStore) Save(cfg *domain.AWSConfig) error {
	persistCfg := &domain.AWSConfig{}
	if cfg != nil {
		*persistCfg = *cfg

		persistCfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)
		persistCfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
		persistCfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)

		if persistCfg.AccessKeyID != "" {
			enc, err := crypto.Encrypt([]byte(persistCfg.AccessKeyID))
			if err != nil {
				return err
			}
			persistCfg.AccessKeyIDEnc = enc
			persistCfg.AccessKeyID = ""
		}
		if persistCfg.SecretAccessKey != "" {
			enc, err := crypto.Encrypt([]byte(persistCfg.SecretAccessKey))
			if err != nil {
				return err
			}
			persistCfg.SecretAccessKeyEnc = enc
			persistCfg.SecretAccessKey = ""
		}

		s.syncToProfile(persistCfg)

		for _, profile := range persistCfg.Profiles {
			if profile.AccessKeyID != "" {
				enc, err := crypto.Encrypt([]byte(profile.AccessKeyID))
				if err != nil {
					return err
				}
				profile.AccessKeyIDEnc = enc
				profile.AccessKeyID = ""
			}
			if profile.SecretAccessKey != "" {
				enc, err := crypto.Encrypt([]byte(profile.SecretAccessKey))
				if err != nil {
					return err
				}
				profile.SecretAccessKeyEnc = enc
				profile.SecretAccessKey = ""
			}
		}
	}

	return s.writeRaw(persistCfg)
}

var _ domain.ConfigStore = (*FileConfigStore)(nil)

func (s *FileConfigStore) ListProfiles() ([]string, error) {
	cfg, err := s.Load()
	if err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		return nil, nil
	}
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	return names, nil
}

func (s *FileConfigStore) SwitchProfile(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		return nil
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return nil
	}
	cfg.ActiveProfile = name
	if profile := cfg.Profiles[name]; profile != nil {
		s.overlayProfile(cfg, profile)
	}
	cfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
	cfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	cfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)
	return s.writeRaw(cfg)
}

func (s *FileConfigStore) SaveProfile(name string, profileCfg *domain.AWSConfig) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*domain.AWSConfig)
	}
	entry := &domain.AWSConfig{}
	if profileCfg != nil {
		*entry = *profileCfg
	}
	cfg.Profiles[name] = entry
	cfg.ActiveProfile = name
	if profileCfg != nil {
		s.overlayProfile(cfg, profileCfg)
	}
	cfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
	cfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	cfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)
	return s.Save(cfg)
}

func (s *FileConfigStore) DeleteProfile(name string) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		return nil
	}
	delete(cfg.Profiles, name)
	if cfg.ActiveProfile == name {
		for n := range cfg.Profiles {
			cfg.ActiveProfile = n
			if profile := cfg.Profiles[n]; profile != nil {
				s.overlayProfile(cfg, profile)
			}
			break
		}
	}
	cfg.LeftPanelRatio = NormalizePanelRatio(cfg.LeftPanelRatio)
	cfg.PanelRatios = ClonePanelRatios(cfg.PanelRatios)
	cfg.EnabledServices = domain.NormalizeEnabledServices(cfg.EnabledServices)
	return s.writeRaw(cfg)
}
