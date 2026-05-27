package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"monostack/internal/domain"
)

type FileSubscriptionStore struct {
	filePath string
}

func NewFileSubscriptionStore(fileName string) *FileSubscriptionStore {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "monostack")
	return &FileSubscriptionStore{
		filePath: filepath.Join(dir, fileName),
	}
}

func (s *FileSubscriptionStore) LoadAll() ([]domain.ManagedSubscription, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var subs []domain.ManagedSubscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	if subs == nil {
		subs = []domain.ManagedSubscription{}
	}
	return subs, nil
}

func (s *FileSubscriptionStore) SaveAll(subs []domain.ManagedSubscription) error {
	if subs == nil {
		subs = []domain.ManagedSubscription{}
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

var _ domain.SubscriptionStore = (*FileSubscriptionStore)(nil)
