package eventbus

import (
	"context"
	"encoding/json"
	"hexagonal-bank/internal/core/application/ports"
	"hexagonal-bank/internal/platform/logging"
)

// LocalBus is a simple logger-based publisher for demo purposes.
type LocalBus struct {
	logger logging.Logger
}

func NewLocalBus(logger logging.Logger) *LocalBus { return &LocalBus{logger: logger} }

func (bus *LocalBus) Publish(ctx context.Context, topic string, payload any) error {
	encoded, _ := json.Marshal(payload)
	bus.logger.Info("event published", "topic", topic, "payload", string(encoded))
	return nil
}

// Ensure interface compliance
var _ ports.EventPublisher = (*LocalBus)(nil)
