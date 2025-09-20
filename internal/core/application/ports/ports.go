package ports

import (
	"context"
	"hexagonal-bank/internal/core/domain"
)

// Readers/Writers are split (ISP) to keep interfaces small.
type AccountReader interface {
	ByID(ctx context.Context, id string) (*domain.Account, error)
}

type AccountWriter interface {
	Save(ctx context.Context, account *domain.Account) error
	Create(ctx context.Context, account *domain.Account) error
}

// PaymentGateway abstracts an external payment rail (here: STP).
type PaymentGateway interface {
	SendTransfer(ctx context.Context, fromID, toID string, cents int64) (string, error)
}

// EventPublisher is a simplified event bus.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}
