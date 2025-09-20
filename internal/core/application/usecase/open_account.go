package usecase

import (
	"context"
	"hexagonal-bank/internal/core/application/ports"
	"hexagonal-bank/internal/core/domain"
	"hexagonal-bank/internal/shared/id"
)

// OpenAccountInput is the request DTO for opening an account.
type OpenAccountInput struct {
	HolderName string
	CLABE      string
}

// OpenAccountOutput is the response DTO.
type OpenAccountOutput struct {
	ID         string `json:"id"`
	HolderName string `json:"holder_name"`
	CLABE      string `json:"clabe"`
	Balance    int64  `json:"balance_cents"`
}

// OpenAccountUseCase orchestrates account creation.
type OpenAccountUseCase struct {
	accountWriter ports.AccountWriter
}

func NewOpenAccountUseCase(accountWriter ports.AccountWriter) *OpenAccountUseCase {
	return &OpenAccountUseCase{accountWriter: accountWriter}
}

func (useCase *OpenAccountUseCase) Execute(ctx context.Context, input OpenAccountInput) (OpenAccountOutput, error) {
	clabe, err := domain.NewCLABE(input.CLABE)
	if err != nil {
		return OpenAccountOutput{}, err
	}
	newID := id.New()
	account, err := domain.NewAccount(newID, input.HolderName, clabe)
	if err != nil {
		return OpenAccountOutput{}, err
	}
	if err := useCase.accountWriter.Create(ctx, account); err != nil {
		return OpenAccountOutput{}, err
	}
	return OpenAccountOutput{
		ID:         account.ID,
		HolderName: account.HolderName(),
		CLABE:      account.CLABE(),
		Balance:    account.Balance(),
	}, nil
}
