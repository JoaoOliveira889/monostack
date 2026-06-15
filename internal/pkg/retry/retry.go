package retry

import (
	"context"
	"math"
	"math/rand/v2"
	"net"
	"strings"
	"time"
)

type Config struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

var DefaultConfig = Config{
	MaxAttempts: 3,
	InitialWait: 500 * time.Millisecond,
	MaxWait:     5 * time.Second,
	Multiplier:  2.0,
}

func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultConfig.MaxAttempts
	}
	if cfg.InitialWait <= 0 {
		cfg.InitialWait = DefaultConfig.InitialWait
	}
	if cfg.MaxWait <= 0 {
		cfg.MaxWait = DefaultConfig.MaxWait
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = DefaultConfig.Multiplier
	}

	var lastErr error
	wait := cfg.InitialWait

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			if !isRetryable(err) {
				return err
			}
			lastErr = err
			if attempt == cfg.MaxAttempts-1 {
				return lastErr
			}

			jitter := time.Duration(float64(wait) * 0.2 * (rand.Float64() - 0.5))
			sleepDuration := wait + jitter
			if sleepDuration < 0 {
				sleepDuration = wait
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
			}

			wait = time.Duration(math.Min(float64(wait)*cfg.Multiplier, float64(cfg.MaxWait)))
		} else {
			return nil
		}
	}

	return lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if ctxError(err) {
		return false
	}

	msg := err.Error()
	if isTimeout(msg) || isConnectionRefused(msg) || isThrottling(msg) || isServerError(msg) {
		return true
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}

func ctxError(err error) bool {
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	return false
}

func isTimeout(msg string) bool {
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "Timeout") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "TLS handshake timeout")
}

func isConnectionRefused(msg string) bool {
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connect: connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe")
}

func isThrottling(msg string) bool {
	return strings.Contains(msg, "Throttling") ||
		strings.Contains(msg, "ThrottlingException") ||
		strings.Contains(msg, "RequestLimitExceeded") ||
		strings.Contains(msg, "TooManyRequestsException") ||
		strings.Contains(msg, "ProvisionedThroughputExceededException") ||
		strings.Contains(msg, "SlowDown") ||
		strings.Contains(msg, "PriorRequestNotComplete")
}

func isServerError(msg string) bool {
	return strings.Contains(msg, "InternalError") ||
		strings.Contains(msg, "InternalFailure") ||
		strings.Contains(msg, "ServiceUnavailable") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "504")
}
