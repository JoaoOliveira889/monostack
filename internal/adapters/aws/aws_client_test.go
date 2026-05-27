package aws

import (
	"testing"

	"monostack/internal/domain"
)

func TestQueueARNFromURL(t *testing.T) {
	got := domain.QueueARNFromURL("https://sqs.us-east-1.amazonaws.com/123456789012/dev-webapi-accounts-sqs", "")
	want := "arn:aws:sqs:us-east-1:123456789012:dev-webapi-accounts-sqs"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestQueueARNFromURL_LocalhostFallback(t *testing.T) {
	got := domain.QueueARNFromURL("http://localhost:4566/000000000000/dev-webapi-accounts-sqs", "us-east-1")
	want := "arn:aws:sqs:us-east-1:000000000000:dev-webapi-accounts-sqs"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
