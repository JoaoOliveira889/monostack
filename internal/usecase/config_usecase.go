package usecase

import "monostack/internal/domain"

type ConfigUseCase struct {
	store             domain.ConfigStore
	subscriptionStore domain.SubscriptionStore
}

func NewConfigUseCase(store domain.ConfigStore) *ConfigUseCase {
	return &ConfigUseCase{store: store}
}

func NewConfigUseCaseWithSubscriptions(store domain.ConfigStore, subStore domain.SubscriptionStore) *ConfigUseCase {
	return &ConfigUseCase{store: store, subscriptionStore: subStore}
}

func (uc *ConfigUseCase) GetConfig() (*domain.AWSConfig, error) {
	return uc.store.Load()
}

func (uc *ConfigUseCase) SaveConfig(cfg *domain.AWSConfig) error {
	return uc.store.Save(cfg)
}

func (uc *ConfigUseCase) ListProfiles() ([]string, error) {
	return uc.store.ListProfiles()
}

func (uc *ConfigUseCase) SwitchProfile(name string) error {
	return uc.store.SwitchProfile(name)
}

func (uc *ConfigUseCase) SaveProfile(name string, cfg *domain.AWSConfig) error {
	return uc.store.SaveProfile(name, cfg)
}

func (uc *ConfigUseCase) DeleteProfile(name string) error {
	return uc.store.DeleteProfile(name)
}

func (uc *ConfigUseCase) LoadSubscriptions() ([]domain.ManagedSubscription, error) {
	if uc.subscriptionStore == nil {
		return nil, nil
	}
	return uc.subscriptionStore.LoadAll()
}

func (uc *ConfigUseCase) SaveSubscriptions(subs []domain.ManagedSubscription) error {
	if uc.subscriptionStore == nil {
		return nil
	}
	return uc.subscriptionStore.SaveAll(subs)
}
