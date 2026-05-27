package tui

import (

	"monostack/internal/domain"
	"monostack/internal/pkg/testutil"
)

type mockConfigStore = testutil.MockConfigStore

type mockSubscriptionStore struct {
	loadAllFunc func() ([]domain.ManagedSubscription, error)
	saveAllFunc func([]domain.ManagedSubscription) error
	savedSubs   []domain.ManagedSubscription
}

func (m *mockSubscriptionStore) LoadAll() ([]domain.ManagedSubscription, error) {
	if m.loadAllFunc != nil {
		return m.loadAllFunc()
	}
	return append([]domain.ManagedSubscription(nil), m.savedSubs...), nil
}

func (m *mockSubscriptionStore) SaveAll(subs []domain.ManagedSubscription) error {
	if m.saveAllFunc != nil {
		return m.saveAllFunc(subs)
	}
	m.savedSubs = append([]domain.ManagedSubscription(nil), subs...)
	return nil
}

type mockS3Manager = testutil.MockS3Manager

type mockSQSManager = testutil.MockSQSManager

type mockSNSManager = testutil.MockSNSManager

type mockSecretsManager = testutil.MockSecretsManager

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}
