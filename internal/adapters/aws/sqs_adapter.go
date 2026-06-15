package aws

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"monostack/internal/domain"
	"monostack/internal/pkg/retry"
)

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

type SQSAdapter struct{ cache *ClientCache }

var _ domain.SQSManager = (*SQSAdapter)(nil)

func NewSQSAdapter(cache *ClientCache) *SQSAdapter {
	return &SQSAdapter{cache: cache}
}

func queueNameFromRef(queueRef string) string {
	trimmed := strings.TrimSpace(queueRef)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "arn:aws:sqs:") {
		parts := strings.Split(trimmed, ":")
		if len(parts) >= 6 {
			return parts[5]
		}
	}
	if strings.Contains(trimmed, "://") {
		if parsed, err := url.Parse(trimmed); err == nil {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return trimmed
}

func (a *SQSAdapter) ListQueues(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SQSQueue, error) {
	if cfg.UseMock {
		return []domain.SQSQueue{
			{
				URL:                "https://sqs.us-east-1.amazonaws.com/123456789012/mock-orders-queue",
				ARN:                "arn:aws:sqs:us-east-1:123456789012:mock-orders-queue",
				Name:               "mock-orders-queue",
				MessagesAvailable:  15,
				MessagesDelayed:    2,
				MessagesNotVisible: 1,
			},
			{
				URL:                "https://sqs.us-east-1.amazonaws.com/123456789012/mock-dead-letter-queue",
				ARN:                "arn:aws:sqs:us-east-1:123456789012:mock-dead-letter-queue",
				Name:               "mock-dead-letter-queue",
				MessagesAvailable:  3,
				MessagesDelayed:    0,
				MessagesNotVisible: 0,
			},
		}, nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQS client: %w", err)
	}
	var queues []domain.SQSQueue
	var nextToken *string

	var out *sqs.ListQueuesOutput
	for {
		err = retry.Do(ctx, retry.DefaultConfig, func() error {
			var innerErr error
			out, innerErr = client.ListQueues(ctx, &sqs.ListQueuesInput{
				NextToken: nextToken,
			})
			return innerErr
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list queues: %w", err)
		}

		for _, qURL := range out.QueueUrls {
			name := qURL[strings.LastIndex(qURL, "/")+1:]

			var attrOut *sqs.GetQueueAttributesOutput
			attrErr := retry.Do(ctx, retry.DefaultConfig, func() error {
				var innerErr error
				attrOut, innerErr = client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
					QueueUrl: aws.String(qURL),
					AttributeNames: []types.QueueAttributeName{
						types.QueueAttributeNameApproximateNumberOfMessages,
						types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
						types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
					},
				})
				return innerErr
			})

			var avail, delayed, notVis int
			if attrErr == nil {
				avail, _ = strconv.Atoi(attrOut.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)])
				delayed, _ = strconv.Atoi(attrOut.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesDelayed)])
				notVis, _ = strconv.Atoi(attrOut.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesNotVisible)])
			}

			queues = append(queues, domain.SQSQueue{
				URL:                qURL,
				ARN:                domain.QueueARNFromURL(qURL, cfg.Region),
				Name:               name,
				MessagesAvailable:  avail,
				MessagesDelayed:    delayed,
				MessagesNotVisible: notVis,
			})
		}

		if out.NextToken == nil || *out.NextToken == "" {
			break
		}
		nextToken = out.NextToken
	}

	return queues, nil
}

func (a *SQSAdapter) SendMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL string, body string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SQS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(body),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to send SQS message: %w", err)
	}
	return nil
}

func (a *SQSAdapter) ReceiveMessages(ctx context.Context, cfg *domain.AWSConfig, queueURL string, maxMessages int) ([]domain.SQSMessage, error) {
	if cfg.UseMock {
		return []domain.SQSMessage{
			{ID: "msg-1111", Body: `{"order_id": 9942, "item": "Sleek mechanical keyboard", "amount": 129.99}`},
			{ID: "msg-2222", Body: `{"order_id": 9943, "item": "Vibrant custom HSL cables", "amount": 25.50}`},
		}, nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQS client: %w", err)
	}
	var out *sqs.ReceiveMessageOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: int32(clamp(maxMessages, 1, 10)),
			VisibilityTimeout:   5,
			WaitTimeSeconds:     1,
		})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	var messages []domain.SQSMessage
	for _, m := range out.Messages {
		messages = append(messages, domain.SQSMessage{
			ID:            aws.ToString(m.MessageId),
			Body:          aws.ToString(m.Body),
			ReceiptHandle: aws.ToString(m.ReceiptHandle),
		})
	}
	return messages, nil
}

func (a *SQSAdapter) DeleteMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL, receiptHandle string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SQS client: %w", err)
	}

	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: aws.String(receiptHandle),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete SQS message: %w", err)
	}
	return nil
}

func (a *SQSAdapter) PurgeQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SQS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
			QueueUrl: aws.String(queueURL),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to purge SQS queue: %w", err)
	}
	return nil
}

func (a *SQSAdapter) DeleteQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SQS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: aws.String(queueURL),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete SQS queue: %w", err)
	}
	return nil
}

func (a *SQSAdapter) CreateQueue(ctx context.Context, cfg *domain.AWSConfig, name string) (string, error) {
	if cfg.UseMock {
		return "https://sqs.us-east-1.amazonaws.com/123456789012/" + name, nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get SQS client: %w", err)
	}
	var out *sqs.CreateQueueOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.CreateQueue(ctx, &sqs.CreateQueueInput{
			QueueName: aws.String(name),
		})
		return innerErr
	})
	if err != nil {
		return "", fmt.Errorf("failed to create SQS queue: %w", err)
	}
	return *out.QueueUrl, nil
}

func (a *SQSAdapter) GetQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributeNames []string) (map[string]string, error) {
	if cfg.UseMock {
		return map[string]string{
			"VisibilityTimeout": "30",
		}, nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQS client: %w", err)
	}
	names := make([]types.QueueAttributeName, 0, len(attributeNames))
	for _, name := range attributeNames {
		if strings.TrimSpace(name) == "" {
			continue
		}
		names = append(names, types.QueueAttributeName(name))
	}
	if len(names) == 0 {
		names = []types.QueueAttributeName{types.QueueAttributeName("All")}
	}

	var out *sqs.GetQueueAttributesOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(queueURL),
			AttributeNames: names,
		})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	return out.Attributes, nil
}

func (a *SQSAdapter) SetQueueAttributes(ctx context.Context, cfg *domain.AWSConfig, queueURL string, attributes map[string]string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SQS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
			QueueUrl:   aws.String(queueURL),
			Attributes: attributes,
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to set queue attributes: %w", err)
	}
	return nil
}

func (a *SQSAdapter) ResolveQueueURL(ctx context.Context, cfg *domain.AWSConfig, queueRef string) (string, error) {
	trimmed := strings.TrimSpace(queueRef)
	if trimmed == "" {
		return "", fmt.Errorf("queue reference is required")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}

	queueName := queueNameFromRef(trimmed)
	if queueName == "" {
		return "", fmt.Errorf("failed to resolve queue url for %q", queueRef)
	}
	if cfg.UseMock {
		return "https://sqs.us-east-1.amazonaws.com/123456789012/" + queueName, nil
	}

	client, err := a.cache.SQS(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get SQS client: %w", err)
	}
	var out *sqs.GetQueueUrlOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: aws.String(queueName),
		})
		return innerErr
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve queue url for %q: %w", queueRef, err)
	}
	return aws.ToString(out.QueueUrl), nil
}

func (a *SQSAdapter) ResolveQueueARN(ctx context.Context, cfg *domain.AWSConfig, queueURL string) (string, error) {
	if queueURL == "" {
		return "", fmt.Errorf("queue url is required")
	}

	if strings.HasPrefix(queueURL, "arn:aws:sqs:") {
		return queueURL, nil
	}

	queueRef := queueURL
	if !strings.HasPrefix(queueRef, "http://") && !strings.HasPrefix(queueRef, "https://") {
		if resolvedURL, err := a.ResolveQueueURL(ctx, cfg, queueRef); err == nil {
			queueRef = resolvedURL
		}
	}

	if parsed := domain.QueueARNFromURL(queueRef, cfg.Region); parsed != "" {
		return parsed, nil
	}

	if cfg.UseMock {
		if queueName := queueNameFromRef(queueURL); queueName != "" {
			return fmt.Sprintf("arn:aws:sqs:us-east-1:123456789012:%s", queueName), nil
		}
		return "", fmt.Errorf("failed to resolve queue arn for %s", queueURL)
	}

	attrs, err := a.GetQueueAttributes(ctx, cfg, queueURL, []string{"QueueArn"})
	if err == nil {
		if arn := strings.TrimSpace(attrs["QueueArn"]); arn != "" {
			return arn, nil
		}
	}

	return "", fmt.Errorf("failed to resolve queue arn for %s", queueURL)
}
