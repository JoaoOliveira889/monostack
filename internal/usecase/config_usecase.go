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
