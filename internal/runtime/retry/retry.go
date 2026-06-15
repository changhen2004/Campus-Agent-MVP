package retry

import (
	"context"
	"time"
)

type Options struct {
	Attempts int
	Delay    time.Duration
	Sleep    func(context.Context, time.Duration) error
}

func Do(ctx context.Context, opts Options, operation func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	attempts := opts.Attempts
	if attempts <= 0 {
		attempts = 1
	}

	sleep := opts.Sleep
	if sleep == nil {
		sleep = sleepWithContext
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := operation(ctx); err != nil {
			lastErr = err
		} else {
			return nil
		}

		if attempt == attempts {
			break
		}

		if err := sleep(ctx, opts.Delay); err != nil {
			return err
		}
	}

	return lastErr
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
