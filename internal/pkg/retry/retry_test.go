package retry

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestDo_SuccessFirstAttempt(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialWait: 10 * time.Millisecond, MaxWait: 100 * time.Millisecond, Multiplier: 2.0}
	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetryOnNetworkError(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialWait: 5 * time.Millisecond, MaxWait: 50 * time.Millisecond, Multiplier: 2.0}
	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return &net.DNSError{Err: "no such host", Name: "localhost"}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsRetries(t *testing.T) {
	cfg := Config{MaxAttempts: 2, InitialWait: 5 * time.Millisecond, MaxWait: 50 * time.Millisecond, Multiplier: 2.0}
	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return errors.New("ThrottlingException: Rate exceeded")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	cfg := Config{MaxAttempts: 3, InitialWait: 10 * time.Millisecond, MaxWait: 100 * time.Millisecond, Multiplier: 2.0}
	calls := 0
	err := Do(context.Background(), cfg, func() error {
		calls++
		return errors.New("AccessDenied: not authorized")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for non-retryable error, got %d", calls)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	cfg := Config{MaxAttempts: 5, InitialWait: 500 * time.Millisecond, MaxWait: 2 * time.Second, Multiplier: 2.0}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		return errors.New("connection refused")
	})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestDo_DefaultsApplied(t *testing.T) {
	err := Do(context.Background(), Config{}, func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsRetryable_TimeoutNetError(t *testing.T) {
	timeoutErr := &netTimeoutError{msg: "i/o timeout"}
	if !isRetryable(timeoutErr) {
		t.Fatal("expected timeout net.Error to be retryable")
	}
}

func TestIsRetryable_NonRetryable(t *testing.T) {
	if isRetryable(nil) {
		t.Fatal("nil should not be retryable")
	}
	if isRetryable(errors.New("AccessDeniedException")) {
		t.Fatal("AccessDenied should not be retryable")
	}
	if isRetryable(errors.New("ValidationError")) {
		t.Fatal("ValidationError should not be retryable")
	}
}

func TestIsRetryable_RetryableErrors(t *testing.T) {
	for _, msg := range []string{
		"RequestLimitExceeded: too many requests",
		"ThrottlingException: rate limit",
		"ServiceUnavailable: try again",
		"ProvisionedThroughputExceededException",
		"SlowDown: reduce request rate",
		"InternalError: something went wrong",
	} {
		if !isRetryable(errors.New(msg)) {
			t.Fatalf("expected %q to be retryable", msg)
		}
	}
}

type netTimeoutError struct {
	msg string
}

func (e *netTimeoutError) Error() string   { return e.msg }
func (e *netTimeoutError) Timeout() bool   { return true }
func (e *netTimeoutError) Temporary() bool { return true }
