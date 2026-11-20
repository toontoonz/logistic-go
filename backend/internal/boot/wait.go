// Package boot retries slow dependencies while the process is starting.
package boot

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// UntilReady calls ping until it returns nil, ctx is cancelled, or
// attempts is exhausted.
//
// One refused dial must not kill the process. On Docker Desktop a
// database container can be "up" for a few seconds before its hostname
// answers, and exiting on the first error looks like a crash instead
// of a wait.
func UntilReady(ctx context.Context, name string, attempts int, pause time.Duration, ping func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("connect %s: %w", name, err)
		}
		// The timeout covers a single dial. It is not the lifetime of
		// the retry loop, so a 2s deadline must not be reported as
		// "the process gave up after one try".
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		last = ping(pingCtx)
		cancel()
		if last == nil {
			if attempt > 1 {
				slog.Info("connected", "service", name, "attempt", attempt)
			}
			return nil
		}
		slog.Warn("not ready", "service", name, "attempt", attempt, "max", attempts, "err", last)
		timer := time.NewTimer(pause)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("connect %s: %w", name, ctx.Err())
		case <-timer.C:
		}
	}
	return fmt.Errorf("connect %s after %d attempts: %w", name, attempts, last)
}
