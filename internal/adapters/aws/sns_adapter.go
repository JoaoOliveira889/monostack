package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"monostack/internal/domain"
	"monostack/internal/pkg/retry"
)

type SNSAdapter struct{ cache *ClientCache }

var _ domain.SNSManager = (*SNSAdapter)(nil)

const filterPolicyScopeAttribute = "FilterPolicyScope"

func NewSNSAdapter(cache *ClientCache) *SNSAdapter {
	return &SNSAdapter{cache: cache}
}

func parseSubscriptionAttributes(sub domain.SNSSubscription, attrs map[string]string) domain.SNSSubscription {
	if attrs == nil {
		return sub
	}

	if fpJSON := attrs["FilterPolicy"]; fpJSON != "" {
		var filterPolicy map[string][]string
		if err := json.Unmarshal([]byte(fpJSON), &filterPolicy); err == nil {
			sub.FilterPolicy = filterPolicy
			if sub.FilterScope == "" {
				sub.FilterScope = domain.SubscriptionFilterScopeMessageAttributes
			}
		}
	}

	if scope := attrs[filterPolicyScopeAttribute]; scope != "" {
		switch scope {
		case "MessageBody":
			sub.FilterScope = domain.SubscriptionFilterScopeMessageBody
		default:
			sub.FilterScope = domain.SubscriptionFilterScopeMessageAttributes
		}
	}

	return sub
}

func (a *SNSAdapter) ListTopics(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSTopic, error) {
	if cfg.UseMock {
		return []domain.SNSTopic{
			{ARN: "arn:aws:sns:us-east-1:123456789012:mock-user-registrations", Name: "mock-user-registrations"},
			{ARN: "arn:aws:sns:us-east-1:123456789012:mock-payment-events", Name: "mock-payment-events"},
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SNS client: %w", err)
	}
	var out *sns.ListTopicsOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ListTopics(ctx, &sns.ListTopicsInput{})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list SNS topics: %w", err)
	}

	var topics []domain.SNSTopic
	for _, t := range out.Topics {
		arn := aws.ToString(t.TopicArn)
		name := arn[strings.LastIndex(arn, ":")+1:]
		topics = append(topics, domain.SNSTopic{
			ARN:  arn,
			Name: name,
		})
	}
	return topics, nil
}

func (a *SNSAdapter) PublishMessage(ctx context.Context, cfg *domain.AWSConfig, topicARN string, body string, subject string, messageAttributes map[string]string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SNS client: %w", err)
	}
	input := &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(body),
	}
	if subject != "" {
		input.Subject = aws.String(subject)
	}

	for k, v := range messageAttributes {
		if input.MessageAttributes == nil {
			input.MessageAttributes = make(map[string]types.MessageAttributeValue)
		}
		input.MessageAttributes[k] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}

	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.Publish(ctx, input)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to publish to SNS topic: %w", err)
	}
	return nil
}

func (a *SNSAdapter) ListSubscriptions(ctx context.Context, cfg *domain.AWSConfig, topicARN string) ([]domain.SNSSubscription, error) {
	if cfg.UseMock {
		return []domain.SNSSubscription{
			{
				ARN:          "arn:aws:sns:us-east-1:123456789012:mock-user-registrations:sub-1",
				Protocol:     "sqs",
				Endpoint:     "arn:aws:sqs:us-east-1:123456789012:mock-orders-queue",
				FilterPolicy: map[string][]string{"event_type": []string{"pix_received", "pix_sent"}},
				FilterScope:  domain.SubscriptionFilterScopeMessageBody,
			},
			{
				ARN:      "arn:aws:sns:us-east-1:123456789012:mock-user-registrations:sub-2",
				Protocol: "email",
				Endpoint: "admin@mock-company.com",
			},
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SNS client: %w", err)
	}
	var out *sns.ListSubscriptionsByTopicOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
			TopicArn: aws.String(topicARN),
		})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions for topic %s: %w", topicARN, err)
	}

	subs := make([]domain.SNSSubscription, len(out.Subscriptions))
	for i, s := range out.Subscriptions {
		subs[i] = domain.SNSSubscription{
			ARN:      aws.ToString(s.SubscriptionArn),
			TopicARN: aws.ToString(s.TopicArn),
			Protocol: aws.ToString(s.Protocol),
			Endpoint: aws.ToString(s.Endpoint),
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := range subs {
		if subs[i].ARN == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var attrs *sns.GetSubscriptionAttributesOutput
			attrErr := retry.Do(ctx, retry.DefaultConfig, func() error {
				var innerErr error
				attrs, innerErr = client.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
					SubscriptionArn: aws.String(subs[idx].ARN),
				})
				return innerErr
			})
			if attrErr == nil {
				subs[idx] = parseSubscriptionAttributes(subs[idx], attrs.Attributes)
			}
		}(i)
	}
	wg.Wait()

	return subs, nil
}

func (a *SNSAdapter) ListAllSubscriptions(ctx context.Context, cfg *domain.AWSConfig) ([]domain.SNSSubscription, error) {
	if cfg.UseMock {
		return []domain.SNSSubscription{
			{ARN: "arn:aws:sns:us-east-1:123456789012:mock-user-registrations:sub-1", TopicARN: "arn:aws:sns:us-east-1:123456789012:mock-user-registrations", Protocol: "sqs", Endpoint: "arn:aws:sqs:us-east-1:123456789012:mock-orders-queue", FilterPolicy: map[string][]string{"event_type": []string{"pix_received", "pix_sent"}}, FilterScope: domain.SubscriptionFilterScopeMessageBody},
			{ARN: "arn:aws:sns:us-east-1:123456789012:mock-user-registrations:sub-2", TopicARN: "arn:aws:sns:us-east-1:123456789012:mock-user-registrations", Protocol: "email", Endpoint: "admin@mock-company.com"},
			{ARN: "arn:aws:sns:us-east-1:123456789012:mock-payment-events:sub-3", TopicARN: "arn:aws:sns:us-east-1:123456789012:mock-payment-events", Protocol: "sqs", Endpoint: "arn:aws:sqs:us-east-1:123456789012:mock-orders-queue"},
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SNS client: %w", err)
	}
	var allSubs []domain.SNSSubscription
	var nextToken *string

	var out *sns.ListSubscriptionsOutput
	for {
		err = retry.Do(ctx, retry.DefaultConfig, func() error {
			var innerErr error
			out, innerErr = client.ListSubscriptions(ctx, &sns.ListSubscriptionsInput{
				NextToken: nextToken,
			})
			return innerErr
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list all SNS subscriptions: %w", err)
		}

		for _, s := range out.Subscriptions {
			allSubs = append(allSubs, domain.SNSSubscription{
				ARN:      aws.ToString(s.SubscriptionArn),
				TopicARN: aws.ToString(s.TopicArn),
				Protocol: aws.ToString(s.Protocol),
				Endpoint: aws.ToString(s.Endpoint),
			})
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := range allSubs {
		if allSubs[i].ARN == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var attrs *sns.GetSubscriptionAttributesOutput
			attrErr := retry.Do(ctx, retry.DefaultConfig, func() error {
				var innerErr error
				attrs, innerErr = client.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
					SubscriptionArn: aws.String(allSubs[idx].ARN),
				})
				return innerErr
			})
			if attrErr == nil {
				allSubs[idx] = parseSubscriptionAttributes(allSubs[idx], attrs.Attributes)
			}
		}(i)
	}
	wg.Wait()

	return allSubs, nil
}

func (a *SNSAdapter) CreateTopic(ctx context.Context, cfg *domain.AWSConfig, name string) (domain.SNSTopic, error) {
	if cfg.UseMock {
		return domain.SNSTopic{
			ARN:  fmt.Sprintf("arn:aws:sns:us-east-1:123456789012:%s", name),
			Name: name,
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return domain.SNSTopic{}, fmt.Errorf("failed to get SNS client: %w", err)
	}
	var out *sns.CreateTopicOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.CreateTopic(ctx, &sns.CreateTopicInput{
			Name: aws.String(name),
		})
		return innerErr
	})
	if err != nil {
		return domain.SNSTopic{}, fmt.Errorf("failed to create SNS topic %s: %w", name, err)
	}

	return domain.SNSTopic{
		ARN:  aws.ToString(out.TopicArn),
		Name: name,
	}, nil
}

func (a *SNSAdapter) DeleteTopic(ctx context.Context, cfg *domain.AWSConfig, topicARN string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SNS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteTopic(ctx, &sns.DeleteTopicInput{
			TopicArn: aws.String(topicARN),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete SNS topic %s: %w", topicARN, err)
	}
	return nil
}

func (a *SNSAdapter) CreateSubscription(ctx context.Context, cfg *domain.AWSConfig, topicARN string, protocol string, endpoint string, filterPolicy map[string][]string, filterScope string) (domain.SNSSubscription, error) {
	if cfg.UseMock {
		subARN := fmt.Sprintf("%s:sub-mock-%d", topicARN, len(filterPolicy))
		normalizedScope, err := domain.NormalizeFilterScopeStrict(filterScope)
		if err != nil {
			return domain.SNSSubscription{}, err
		}
		return domain.SNSSubscription{
			ARN:          subARN,
			Protocol:     protocol,
			Endpoint:     endpoint,
			FilterPolicy: filterPolicy,
			FilterScope:  normalizedScope,
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return domain.SNSSubscription{}, fmt.Errorf("failed to get SNS client: %w", err)
	}
	normalizedScope, err := domain.NormalizeFilterScopeStrict(filterScope)
	if err != nil {
		return domain.SNSSubscription{}, err
	}
	if protocol == "sqs" {
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
			endpoint = domain.QueueARNFromURL(endpoint, cfg.Region)
		}
		if endpoint == "" {
			return domain.SNSSubscription{}, fmt.Errorf("failed to resolve SQS endpoint ARN")
		}
	}

	attributes := map[string]string{}
	if len(filterPolicy) > 0 {
		fpJSON, err := json.Marshal(filterPolicy)
		if err != nil {
			return domain.SNSSubscription{}, fmt.Errorf("failed to marshal filter policy: %w", err)
		}
		attributes["FilterPolicy"] = string(fpJSON)
		if normalizedScope == domain.SubscriptionFilterScopeMessageBody {
			attributes[filterPolicyScopeAttribute] = "MessageBody"
		}
	}

	var out *sns.SubscribeOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.Subscribe(ctx, &sns.SubscribeInput{
			TopicArn:              aws.String(topicARN),
			Protocol:              aws.String(protocol),
			Endpoint:              aws.String(endpoint),
			ReturnSubscriptionArn: true,
			Attributes:            attributes,
		})
		return innerErr
	})
	if err != nil {
		return domain.SNSSubscription{}, fmt.Errorf("failed to subscribe to SNS topic %s: %w", topicARN, err)
	}

	return domain.SNSSubscription{
		ARN:          aws.ToString(out.SubscriptionArn),
		Protocol:     protocol,
		Endpoint:     endpoint,
		FilterPolicy: filterPolicy,
		FilterScope:  normalizedScope,
	}, nil
}

func (a *SNSAdapter) DeleteSubscription(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SNS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.Unsubscribe(ctx, &sns.UnsubscribeInput{
			SubscriptionArn: aws.String(subscriptionARN),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete SNS subscription %s: %w", subscriptionARN, err)
	}
	return nil
}

func (a *SNSAdapter) GetSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string) (map[string]string, error) {
	if cfg.UseMock {
		return map[string]string{
			"FilterPolicy":      `{"event_type":["pix_received","pix_sent"]}`,
			"FilterPolicyScope": "MessageBody",
		}, nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get SNS client: %w", err)
	}
	var out *sns.GetSubscriptionAttributesOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(subscriptionARN),
		})
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription attributes for %s: %w", subscriptionARN, err)
	}

	return out.Attributes, nil
}

func (a *SNSAdapter) SetSubscriptionAttributes(ctx context.Context, cfg *domain.AWSConfig, subscriptionARN string, attributeName string, attributeValue string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.SNS(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get SNS client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.SetSubscriptionAttributes(ctx, &sns.SetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(subscriptionARN),
			AttributeName:   aws.String(attributeName),
			AttributeValue:  aws.String(attributeValue),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to set subscription attribute %s for %s: %w", attributeName, subscriptionARN, err)
	}
	return nil
}
