package usecase

import (
	"monostack/internal/domain"
	"monostack/internal/pkg/testutil"
)

type mockConfigStore = testutil.MockConfigStore

type mockS3Manager = testutil.MockS3Manager

var _ domain.S3Manager = (*mockS3Manager)(nil)

type mockSQSManager = testutil.MockSQSManager

var _ domain.SQSManager = (*mockSQSManager)(nil)

type mockSNSManager = testutil.MockSNSManager

var _ domain.SNSManager = (*mockSNSManager)(nil)

type mockSecretsManager = testutil.MockSecretsManager

var _ domain.SecretsManager = (*mockSecretsManager)(nil)
