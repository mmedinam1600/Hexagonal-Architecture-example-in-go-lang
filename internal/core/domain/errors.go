package domain

import "errors"

// Domain-level errors (business significance).
// These are mapped at adapters-in (HTTP) to proper status codes.
var (
	ErrInvalidAmount    = errors.New("invalid amount: must be > 0")
	ErrInsufficientFund = errors.New("insufficient funds")
	ErrInvalidCLABE     = errors.New("invalid CLABE: must be 18 digits")
	ErrEmptyHolder      = errors.New("holder name cannot be empty")
)