package stp

import (
	"context"
	"errors"
	"fmt"
	"hexagonal-bank/internal/core/application/ports"
	"hexagonal-bank/internal/platform/backoff"
	"hexagonal-bank/internal/platform/logging"
	"math/rand"
	"time"
)

// FakeSTP simulates a flaky external API and uses retry + backoff with jitter.
type FakeSTP struct {
	logger logging.Logger
}

func NewFakeSTP(logger logging.Logger) *FakeSTP { return &FakeSTP{logger: logger} }

func (client *FakeSTP) SendTransfer(ctx context.Context, fromID, toID string, cents int64) (string, error) {
	const (
		maxRetries = 4
		baseDelay  = 200 * time.Millisecond
		multiplier = 2.0
		maxDelay   = 3 * time.Second
	)
	for attemptIndex := 0; attemptIndex <= maxRetries; attemptIndex++ {
		// 70% success simulation; 30% transient failure
		if rand.Float64() < 0.7 {
			return "OK", nil
		}
		// transient error
		transientErr := errors.New("temporary STP outage")
		if attemptIndex == maxRetries {
			client.logger.Error("STP failed after retries", "from", fromID, "to", toID, "err", transientErr)
			return "FAILED", transientErr
		}
		sleepDuration := backoff.FullJitter(attemptIndex, baseDelay, multiplier, maxDelay)
		client.logger.Warn("STP transient error, retrying", "attempt", attemptIndex, "sleep", sleepDuration)
		select {
		case <-ctx.Done():
			return "CANCELLED", ctx.Err()
		case <-time.After(sleepDuration):
		}
	}
	return "FAILED", fmt.Errorf("unreachable") // defensive
}

// Ensure interface compliance
var _ ports.PaymentGateway = (*FakeSTP)(nil)
