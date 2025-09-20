package memory

import (
	"context"
	"errors"
	"hexagonal-bank/internal/core/application/ports"
	"hexagonal-bank/internal/core/domain"
	"sync"
)

// AccountRepository is an in-memory repository simulating a database (thread-safe).
type AccountRepository struct {
	mutex sync.RWMutex
	data  map[string]*domain.Account
}

func NewAccountRepo() *AccountRepository {
	return &AccountRepository{data: make(map[string]*domain.Account)}
}

func (repository *AccountRepository) ByID(ctx context.Context, id string) (*domain.Account, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	account, exists := repository.data[id]
	if !exists {
		return nil, errors.New("not found")
	}
	return cloneAccount(account), nil
}

func (repository *AccountRepository) Save(ctx context.Context, account *domain.Account) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	repository.data[account.ID] = cloneAccount(account)
	return nil
}

func (repository *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	if _, exists := repository.data[account.ID]; exists {
		return errors.New("already exists")
	}
	repository.data[account.ID] = cloneAccount(account)
	return nil
}

func cloneAccount(account *domain.Account) *domain.Account {
	copy := *account
	return &copy
}

// Ensure interface compliance (at compile-time).
var _ ports.AccountReader = (*AccountRepository)(nil)
var _ ports.AccountWriter = (*AccountRepository)(nil)
