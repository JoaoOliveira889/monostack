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

type SQSAdapter struct{}

var _ domain.SQSManager = (*SQSAdapter)(nil)

func NewSQSAdapter() *SQSAdapter {
	return &SQSAdapter{}
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	out, err := client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	var queues []domain.SQSQueue
	for _, qURL := range out.QueueUrls {
		name := qURL[strings.LastIndex(qURL, "/")+1:]

		attrOut, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(qURL),
			AttributeNames: []types.QueueAttributeName{
				types.QueueAttributeNameApproximateNumberOfMessages,
				types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
				types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
			},
		})

		var avail, delayed, notVis int
		if err == nil {
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

	return queues, nil
}

func (a *SQSAdapter) SendMessage(ctx context.Context, cfg *domain.AWSConfig, queueURL string, body string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(body),
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	out, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: int32(clamp(maxMessages, 1, 10)),
		VisibilityTimeout:   5,
		WaitTimeSeconds:     1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	var messages []domain.SQSMessage
	for _, m := range out.Messages {
		messages = append(messages, domain.SQSMessage{
			ID:   aws.ToString(m.MessageId),
			Body: aws.ToString(m.Body),
		})
	}
	return messages, nil
}

func (a *SQSAdapter) PurgeQueue(ctx context.Context, cfg *domain.AWSConfig, queueURL string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	_, err = client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: aws.String(queueURL),
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	_, err = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
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

	out, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: names,
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	_, err = client.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(queueURL),
		Attributes: attributes,
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

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsCfg)
	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
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
