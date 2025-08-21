package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/robertguss/rss-agent-cli/pkg/errs"
)

// Config defines retry behavior with exponential backoff parameters.
type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	MaxElapsed time.Duration
}

// DefaultConfig returns a retry configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries: 3,
		BaseDelay:  250 * time.Millisecond,
		MaxDelay:   2 * time.Second,
		Multiplier: 2.0,
		MaxElapsed: 30 * time.Second,
	}
}

// Do executes a function with exponential backoff retry logic.
func Do(ctx context.Context, config Config, fn func() error) error {
	if config.MaxRetries <= 0 {
		return fn()
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = config.BaseDelay
	b.MaxInterval = config.MaxDelay
	b.Multiplier = config.Multiplier
	b.MaxElapsedTime = config.MaxElapsed

	backoffWithContext := backoff.WithContext(b, ctx)

	var lastErr error
	attempt := 0

	operation := func() error {
		attempt++
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt >= config.MaxRetries {
			return backoff.Permanent(err)
		}

		if !errs.IsRetryable(err) {
			return backoff.Permanent(err)
		}

		return err
	}

	err := backoff.Retry(operation, backoffWithContext)
	if err != nil {
		return lastErr
	}

	return nil
}

func DoWithCallback(ctx context.Context, config Config, fn func() error, onRetry func(attempt int, err error)) error {
	if config.MaxRetries <= 0 {
		return fn()
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = config.BaseDelay
	b.MaxInterval = config.MaxDelay
	b.Multiplier = config.Multiplier
	b.MaxElapsedTime = config.MaxElapsed

	backoffWithContext := backoff.WithContext(b, ctx)

	var lastErr error
	attempt := 0

	operation := func() error {
		attempt++
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt > 1 && onRetry != nil {
			onRetry(attempt-1, err)
		}

		if attempt >= config.MaxRetries {
			return backoff.Permanent(err)
		}

		if !errs.IsRetryable(err) {
			return backoff.Permanent(err)
		}

		return err
	}

	err := backoff.Retry(operation, backoffWithContext)
	if err != nil {
		return lastErr
	}

	return nil
}

func DoSimple(ctx context.Context, fn func() error) error {
	return Do(ctx, DefaultConfig(), fn)
}
