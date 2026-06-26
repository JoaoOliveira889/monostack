package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"monostack/internal/domain"
	"monostack/internal/pkg/retry"
)

type SecretsAdapter struct{ cache *ClientCache }

var _ domain.SecretsManager = (*SecretsAdapter)(nil)

func NewSecretsAdapter(cache *ClientCache) *SecretsAdapter {
	return &SecretsAdapter{cache: cache}
}

func (a *SecretsAdapter) ListSecrets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.Secret, error) {
	if cfg.UseMock {
		return []domain.Secret{
			{ARN: "arn:aws:secretsmanager:us-east-1:123456789012:secret:mock/payment-api", Name: "payment-api", Description: "Mock payment API secret", CreatedDate: time.Now().Add(-48 * time.Hour).Format(time.RFC3339), LastChangedDate: time.Now().Add(-4 * time.Hour).Format(time.RFC3339), KMSKeyID: "alias/aws/secrets", VersionCount: 2},
			{ARN: "arn:aws:secretsmanager:us-east-1:123456789012:secret:mock/sqs-credentials", Name: "sqs-credentials", Description: "Mock queue credentials", CreatedDate: time.Now().Add(-72 * time.Hour).Format(time.RFC3339), LastChangedDate: time.Now().Add(-24 * time.Hour).Format(time.RFC3339), KMSKeyID: "alias/aws/secrets", VersionCount: 1},
		}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	var out *secretsmanager.ListSecretsOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var secrets []domain.Secret
	for _, item := range out.SecretList {
		secret := domain.Secret{
			ARN:               aws.ToString(item.ARN),
			Name:              aws.ToString(item.Name),
			Description:       aws.ToString(item.Description),
			CreatedDate:       formatTime(item.CreatedDate),
			LastChangedDate:   formatTime(item.LastChangedDate),
			DeletedDate:       formatTime(item.DeletedDate),
			PrimaryRegion:     aws.ToString(item.PrimaryRegion),
			KMSKeyID:          aws.ToString(item.KmsKeyId),
			RotationEnabled:   aws.ToBool(item.RotationEnabled),
			RotationLambdaARN: aws.ToString(item.RotationLambdaARN),
			VersionCount:      1,
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

func (a *SecretsAdapter) DescribeSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) (domain.Secret, error) {
	if cfg.UseMock {
		return domain.Secret{
			ARN:             "arn:aws:secretsmanager:us-east-1:123456789012:secret:mock/payment-api",
			Name:            "payment-api",
			Description:     "Mock payment API secret",
			CreatedDate:     time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			LastChangedDate: time.Now().Add(-4 * time.Hour).Format(time.RFC3339),
			KMSKeyID:        "alias/aws/secrets",
			VersionCount:    2,
		}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return domain.Secret{}, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	var out *secretsmanager.DescribeSecretOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{SecretId: aws.String(secretID)})
		return innerErr
	})
	if err != nil {
		return domain.Secret{}, fmt.Errorf("failed to describe secret %s: %w", secretID, err)
	}

	return domain.Secret{
		ARN:               aws.ToString(out.ARN),
		Name:              aws.ToString(out.Name),
		Description:       aws.ToString(out.Description),
		CreatedDate:       formatTime(out.CreatedDate),
		LastChangedDate:   formatTime(out.LastChangedDate),
		DeletedDate:       formatTime(out.DeletedDate),
		PrimaryRegion:     aws.ToString(out.PrimaryRegion),
		KMSKeyID:          aws.ToString(out.KmsKeyId),
		RotationEnabled:   aws.ToBool(out.RotationEnabled),
		RotationLambdaARN: aws.ToString(out.RotationLambdaARN),
		VersionCount:      len(out.VersionIdsToStages),
	}, nil
}

func (a *SecretsAdapter) GetSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID string, versionID string, versionStage string) (domain.SecretValue, error) {
	if cfg.UseMock {
		return domain.SecretValue{
			VersionID:    "mock-version-1",
			SecretString: `{"token":"mock-secret-value"}`,
		}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return domain.SecretValue{}, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	input := &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretID)}
	if strings.TrimSpace(versionID) != "" {
		input.VersionId = aws.String(versionID)
	}
	if strings.TrimSpace(versionStage) != "" {
		input.VersionStage = aws.String(versionStage)
	}
	var out *secretsmanager.GetSecretValueOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.GetSecretValue(ctx, input)
		return innerErr
	})
	if err != nil {
		return domain.SecretValue{}, fmt.Errorf("failed to get secret value for %s: %w", secretID, err)
	}

	value := domain.SecretValue{VersionID: aws.ToString(out.VersionId), SecretString: aws.ToString(out.SecretString)}
	if len(out.SecretBinary) > 0 {
		value.SecretBinaryBase64 = base64.StdEncoding.EncodeToString(out.SecretBinary)
	}
	return value, nil
}

func (a *SecretsAdapter) ListSecretVersions(ctx context.Context, cfg *domain.AWSConfig, secretID string) ([]domain.SecretVersion, error) {
	if cfg.UseMock {
		return []domain.SecretVersion{
			{VersionID: "mock-version-1", Stages: []string{"AWSCURRENT"}, CreatedDate: time.Now().Add(-4 * time.Hour).Format(time.RFC3339)},
			{VersionID: "mock-version-0", Stages: []string{"AWSPREVIOUS"}, CreatedDate: time.Now().Add(-48 * time.Hour).Format(time.RFC3339)},
		}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	var out *secretsmanager.ListSecretVersionIdsOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ListSecretVersionIds(ctx, &secretsmanager.ListSecretVersionIdsInput{SecretId: aws.String(secretID), IncludeDeprecated: aws.Bool(true)})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for %s: %w", secretID, err)
	}

	versions := make([]domain.SecretVersion, 0, len(out.Versions))
	for _, version := range out.Versions {
		versions = append(versions, domain.SecretVersion{
			VersionID:   aws.ToString(version.VersionId),
			Stages:      append([]string(nil), version.VersionStages...),
			CreatedDate: formatTime(version.CreatedDate),
		})
	}
	return versions, nil
}

func (a *SecretsAdapter) CreateSecret(ctx context.Context, cfg *domain.AWSConfig, name string, value string, description string) (domain.Secret, error) {
	if cfg.UseMock {
		return domain.Secret{
			ARN:             fmt.Sprintf("arn:aws:secretsmanager:us-east-1:123456789012:secret:mock/%s", name),
			Name:            name,
			Description:     description,
			CreatedDate:     time.Now().Format(time.RFC3339),
			LastChangedDate: time.Now().Format(time.RFC3339),
			KMSKeyID:        "alias/aws/secrets",
			VersionCount:    1,
		}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return domain.Secret{}, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(value),
	}
	if strings.TrimSpace(description) != "" {
		input.Description = aws.String(description)
	}
	var out *secretsmanager.CreateSecretOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.CreateSecret(ctx, input)
		return innerErr
	})
	if err != nil {
		return domain.Secret{}, fmt.Errorf("failed to create secret %s: %w", name, err)
	}
	return domain.Secret{
		ARN:             aws.ToString(out.ARN),
		Name:            name,
		Description:     description,
		CreatedDate:     time.Now().Format(time.RFC3339),
		LastChangedDate: time.Now().Format(time.RFC3339),
		VersionCount:    1,
	}, nil
}

func (a *SecretsAdapter) UpdateSecretValue(ctx context.Context, cfg *domain.AWSConfig, secretID string, value string, description string) (domain.SecretValue, error) {
	if cfg.UseMock {
		return domain.SecretValue{VersionID: "mock-version-2", SecretString: value}, nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return domain.SecretValue{}, fmt.Errorf("failed to get Secrets client: %w", err)
	}
	if strings.TrimSpace(description) != "" {
		if err := retry.Do(ctx, retry.DefaultConfig, func() error {
			_, innerErr := client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
				SecretId:    aws.String(secretID),
				Description: aws.String(description),
			})
			return innerErr
		}); err != nil {
			return domain.SecretValue{}, fmt.Errorf("failed to update secret metadata for %s: %w", secretID, err)
		}
	}

	var out *secretsmanager.PutSecretValueOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretID),
			SecretString: aws.String(value),
		})
		return innerErr
	})
	if err != nil {
		return domain.SecretValue{}, fmt.Errorf("failed to update secret value for %s: %w", secretID, err)
	}
	return domain.SecretValue{VersionID: aws.ToString(out.VersionId), SecretString: value}, nil
}

func (a *SecretsAdapter) UpdateSecretVersionStage(ctx context.Context, cfg *domain.AWSConfig, secretID string, versionStage string, moveToVersionID string, removeFromVersionID string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get Secrets client: %w", err)
	}
	input := &secretsmanager.UpdateSecretVersionStageInput{
		SecretId:        aws.String(secretID),
		VersionStage:    aws.String(versionStage),
		MoveToVersionId: aws.String(moveToVersionID),
	}
	if strings.TrimSpace(removeFromVersionID) != "" {
		input.RemoveFromVersionId = aws.String(removeFromVersionID)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.UpdateSecretVersionStage(ctx, input)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to update secret version stage for %s: %w", secretID, err)
	}
	return nil
}

func (a *SecretsAdapter) DeleteSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string, recoveryWindowDays int, forceDeleteWithoutRecovery bool) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get Secrets client: %w", err)
	}
	input := &secretsmanager.DeleteSecretInput{SecretId: aws.String(secretID)}
	if forceDeleteWithoutRecovery {
		input.ForceDeleteWithoutRecovery = aws.Bool(true)
	} else if recoveryWindowDays > 0 {
		input.RecoveryWindowInDays = aws.Int64(int64(recoveryWindowDays))
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteSecret(ctx, input)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", secretID, err)
	}
	return nil
}

func (a *SecretsAdapter) RestoreSecret(ctx context.Context, cfg *domain.AWSConfig, secretID string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.Secrets(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get Secrets client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.RestoreSecret(ctx, &secretsmanager.RestoreSecretInput{SecretId: aws.String(secretID)})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to restore secret %s: %w", secretID, err)
	}
	return nil
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
