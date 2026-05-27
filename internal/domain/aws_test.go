package domain

import (
	"testing"
)

func TestS3Bucket_Fields(t *testing.T) {
	b := S3Bucket{Name: "my-bucket"}
	if b.Name != "my-bucket" {
		t.Errorf("expected Name 'my-bucket', got %q", b.Name)
	}
}

func TestS3Object_Fields(t *testing.T) {
	o := S3Object{Key: "path/to/file.txt", Size: 1024, LastModified: "2026-01-01T00:00:00Z"}
	if o.Key != "path/to/file.txt" {
		t.Errorf("expected Key 'path/to/file.txt', got %q", o.Key)
	}
	if o.Size != 1024 {
		t.Errorf("expected Size 1024, got %d", o.Size)
	}
}

func TestSQSQueue_Fields(t *testing.T) {
	q := SQSQueue{
		URL:                "https://sqs.us-east-1.amazonaws.com/123/test",
		Name:               "test-queue",
		MessagesAvailable:  10,
		MessagesDelayed:    2,
		MessagesNotVisible: 1,
	}
	if q.Name != "test-queue" {
		t.Errorf("expected Name 'test-queue', got %q", q.Name)
	}
	if q.MessagesAvailable != 10 {
		t.Errorf("expected MessagesAvailable 10, got %d", q.MessagesAvailable)
	}
}

func TestSNSSubscription_Fields(t *testing.T) {
	s := SNSSubscription{
		ARN:         "arn:aws:sns:us-east-1:123:topic:sub",
		Protocol:    "sqs",
		Endpoint:    "arn:aws:sqs:us-east-1:123:queue",
		FilterScope: SubscriptionFilterScopeMessageBody,
	}
	if s.Protocol != "sqs" {
		t.Errorf("expected Protocol 'sqs', got %q", s.Protocol)
	}
	if s.FilterScope != SubscriptionFilterScopeMessageBody {
		t.Errorf("expected FilterScope %q, got %q", SubscriptionFilterScopeMessageBody, s.FilterScope)
	}
}

func TestS3ManagerInterface(t *testing.T) {
	m := struct{ S3Manager }{}
	if m.S3Manager != nil {
		t.Error("expected nil interface")
	}
}

func TestSQSManagerInterface(t *testing.T) {
	m := struct{ SQSManager }{}
	if m.SQSManager != nil {
		t.Error("expected nil interface")
	}
}

func TestSNSManagerInterface(t *testing.T) {
	m := struct{ SNSManager }{}
	if m.SNSManager != nil {
		t.Error("expected nil interface")
	}
}
