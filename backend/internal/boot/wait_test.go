package boot

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestUntilReadyRetriesThenSucceeds(t *testing.T) {
	calls := 0
	err := UntilReady(context.Background(), "postgres", 5, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("connection refused")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestUntilReadyGivesUpAfterTheBudget(t *testing.T) {
	err := UntilReady(context.Background(), "redis", 2, time.Millisecond, func(context.Context) error {
		return errors.New("no such host")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "connect redis after 2 attempts: no such host" {
		t.Fatalf("error = %q", got)
	}
}

func TestUntilReadyStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := UntilReady(ctx, "postgres", 5, time.Second, func(context.Context) error {
		called = true
		return errors.New("should not be called")
	})
	if err == nil {
		t.Fatal("expected cancellation")
	}
	if called {
		t.Fatal("ping ran after cancel")
	}
}
