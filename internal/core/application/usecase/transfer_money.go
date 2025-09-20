package usecase

import (
	"context"
	"fmt"
	"hexagonal-bank/internal/core/application/ports"
)

type TransferInput struct {
	FromID string
	ToID   string
	Cents  int64
}

type TransferOutput struct {
	FromBalance int64 `json:"from_balance_cents"`
	ToBalance   int64 `json:"to_balance_cents"`
}

type TransferMoneyUseCase struct {
	accountReader  ports.AccountReader
	accountWriter  ports.AccountWriter
	paymentGateway ports.PaymentGateway
	eventPublisher ports.EventPublisher
}

func NewTransferMoneyUseCase(
	accountReader ports.AccountReader,
	accountWriter ports.AccountWriter,
	paymentGateway ports.PaymentGateway,
	eventPublisher ports.EventPublisher,
) *TransferMoneyUseCase {
	return &TransferMoneyUseCase{
		accountReader:  accountReader,
		accountWriter:  accountWriter,
		paymentGateway: paymentGateway,
		eventPublisher: eventPublisher,
	}
}

func (useCase *TransferMoneyUseCase) Execute(ctx context.Context, input TransferInput) (TransferOutput, error) {
	fromAccount, err := useCase.accountReader.ByID(ctx, input.FromID)
	if err != nil {
		return TransferOutput{}, err
	}

	toAccount, err := useCase.accountReader.ByID(ctx, input.ToID)
	if err != nil {
		return TransferOutput{}, err
	}

	// Domain rules first
	if err := fromAccount.Debit(input.Cents); err != nil {
		return TransferOutput{}, err
	}
	if err := toAccount.Credit(input.Cents); err != nil {
		return TransferOutput{}, err
	}

	// External side-effect (STP) via port
	status, err := useCase.paymentGateway.SendTransfer(ctx, input.FromID, input.ToID, input.Cents)
	if err != nil {
		return TransferOutput{}, err
	}
	if status != "OK" {
		return TransferOutput{}, fmt.Errorf("stp not ok: %s", status)
	}

	// Persist
	if err := useCase.accountWriter.Save(ctx, fromAccount); err != nil {
		return TransferOutput{}, err
	}
	if err := useCase.accountWriter.Save(ctx, toAccount); err != nil {
		return TransferOutput{}, err
	}

	// Publish integration event (fire-and-forget)
	_ = useCase.eventPublisher.Publish(ctx, "transfer.completed", map[string]any{
		"from_id": input.FromID, "to_id": input.ToID, "cents": input.Cents,
	})

	return TransferOutput{FromBalance: fromAccount.Balance(), ToBalance: toAccount.Balance()}, nil
}
