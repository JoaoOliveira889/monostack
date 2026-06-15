package aws

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"monostack/internal/domain"
)

type ClientCache struct {
	mu         sync.RWMutex
	s3Client   *s3.Client
	sqsClient  *sqs.Client
	snsClient  *sns.Client
	secretsClient *secretsmanager.Client
	configHash string
}

func NewClientCache() *ClientCache {
	return &ClientCache{}
}

func (c *ClientCache) configKey(cfg *domain.AWSConfig) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s", cfg.EndpointURL, cfg.Region, cfg.AccessKeyID, cfg.SecretAccessKey)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *ClientCache) S3(ctx context.Context, cfg *domain.AWSConfig) (*s3.Client, error) {
	key := c.configKey(cfg)

	c.mu.RLock()
	if c.s3Client != nil && c.configHash == key {
		client := c.s3Client
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	c.mu.Lock()
	c.s3Client = client
	c.configHash = key
	c.mu.Unlock()

	return client, nil
}

func (c *ClientCache) SQS(ctx context.Context, cfg *domain.AWSConfig) (*sqs.Client, error) {
	key := c.configKey(cfg)

	c.mu.RLock()
	if c.sqsClient != nil && c.configHash == key {
		client := c.sqsClient
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(awsCfg)

	c.mu.Lock()
	c.sqsClient = client
	c.configHash = key
	c.mu.Unlock()

	return client, nil
}

func (c *ClientCache) SNS(ctx context.Context, cfg *domain.AWSConfig) (*sns.Client, error) {
	key := c.configKey(cfg)

	c.mu.RLock()
	if c.snsClient != nil && c.configHash == key {
		client := c.snsClient
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := sns.NewFromConfig(awsCfg)

	c.mu.Lock()
	c.snsClient = client
	c.configHash = key
	c.mu.Unlock()

	return client, nil
}

func (c *ClientCache) Secrets(ctx context.Context, cfg *domain.AWSConfig) (*secretsmanager.Client, error) {
	key := c.configKey(cfg)

	c.mu.RLock()
	if c.secretsClient != nil && c.configHash == key {
		client := c.secretsClient
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(awsCfg)

	c.mu.Lock()
	c.secretsClient = client
	c.configHash = key
	c.mu.Unlock()

	return client, nil
}
