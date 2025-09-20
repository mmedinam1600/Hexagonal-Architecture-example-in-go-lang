package usecase

import (
	"context"
	"hexagonal-bank/internal/core/application/ports"
)

type DepositInput struct {
	AccountID string
	Cents     int64
}

type DepositOutput struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance_cents"`
}

type DepositMoneyUseCase struct {
	accountReader ports.AccountReader
	accountWriter ports.AccountWriter
}

func NewDepositMoneyUseCase(accountReader ports.AccountReader, accountWriter ports.AccountWriter) *DepositMoneyUseCase {
	return &DepositMoneyUseCase{accountReader: accountReader, accountWriter: accountWriter}
}

func (useCase *DepositMoneyUseCase) Execute(ctx context.Context, input DepositInput) (DepositOutput, error) {
	account, err := useCase.accountReader.ByID(ctx, input.AccountID)
	if err != nil {
		return DepositOutput{}, err
	}
	if err := account.Credit(input.Cents); err != nil {
		return DepositOutput{}, err
	}
	if err := useCase.accountWriter.Save(ctx, account); err != nil {
		return DepositOutput{}, err
	}
	return DepositOutput{ID: account.ID, Balance: account.Balance()}, nil
}

// ReaderPort exposes the reader (used by HTTP adapter for GET /accounts/:id)
func (useCase *DepositMoneyUseCase) ReaderPort() ports.AccountReader { return useCase.accountReader }
