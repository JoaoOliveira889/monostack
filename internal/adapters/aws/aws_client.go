package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"monostack/internal/domain"
)

func GetSDKConfig(ctx context.Context, cfg *domain.AWSConfig) (aws.Config, error) {

	if cfg.EndpointURL != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.EndpointURL,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		})

		keyID := cfg.AccessKeyID
		secretKey := cfg.SecretAccessKey
		if keyID == "" {
			keyID = "test"
		}
		if secretKey == "" {
			secretKey = "test"
		}
		if keyID == "test" && secretKey == "test" && !cfg.UseMock {
			fmt.Fprintln(os.Stderr, "monostack: WARNING: using default 'test'/'test' credentials for "+cfg.EndpointURL+"; set Access Key ID and Secret Access Key in Settings")
		}
		creds := credentials.NewStaticCredentialsProvider(keyID, secretKey, "")

		return config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(creds),
			config.WithEndpointResolverWithOptions(customResolver),
		)
	}

	if cfg.Region != "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	}
	return config.LoadDefaultConfig(ctx)
}
