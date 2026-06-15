package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoSucceedsAfterTransientFailures(t *testing.T) {
	t.Parallel()

	attempts := 0
	err := Do(context.Background(), Options{
		Attempts: 3,
		Delay:    time.Second,
		Sleep:    func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("not ready")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("attempt count mismatch: got %d want %d", attempts, 3)
	}
}

func TestDoReturnsLastErrorAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("still down")
	attempts := 0
	err := Do(context.Background(), Options{
		Attempts: 2,
		Sleep:    func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		attempts++
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got %v want %v", err, wantErr)
	}

	if attempts != 2 {
		t.Fatalf("attempt count mismatch: got %d want %d", attempts, 2)
	}
}

func TestDoStopsWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	err := Do(ctx, Options{
		Attempts: 3,
		Sleep:    func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		attempts++
		return errors.New("not ready")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error mismatch: got %v want %v", err, context.Canceled)
	}

	if attempts != 0 {
		t.Fatalf("attempt count mismatch: got %d want %d", attempts, 0)
	}
}
