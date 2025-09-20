package domain

import (
	"fmt"
	"strings"
)

// Account is an Entity responsible for protecting its invariants.
// Invariants:
//  - balance must never go below 0
//  - holderName must not be empty
type Account struct {
	ID         string
	holderName string
	clabe      CLABE
	balance    int64 // stored in cents
}

// NewAccount constructs a valid Account with zero balance.
func NewAccount(id string, holderName string, clabe CLABE) (*Account, error) {
	if strings.TrimSpace(holderName) == "" {
		return nil, ErrEmptyHolder
	}
	return &Account{
		ID:         id,
		holderName: holderName,
		clabe:      clabe,
		balance:    0,
	}, nil
}

// Debit decreases the balance while ensuring it never goes negative.
func (a *Account) Debit(cents int64) error {
	if cents <= 0 {
		return ErrInvalidAmount
	}
	if a.balance-cents < 0 {
		return ErrInsufficientFund
	}
	a.balance -= cents
	return nil
}

// Credit increases the balance, validating amount.
func (a *Account) Credit(cents int64) error {
	if cents <= 0 {
		return ErrInvalidAmount
	}
	a.balance += cents
	return nil
}

func (a *Account) HolderName() string { return a.holderName }
func (a *Account) CLABE() string      { return a.clabe.String() }
func (a *Account) Balance() int64     { return a.balance }

func (a *Account) String() string {
	return fmt.Sprintf(
		"Account{ID=%s, holder=%s, clabe=%s, balance=%d}",
		a.ID, a.holderName, a.clabe.String(), a.balance,
	)
}
