package domain

import "strings"

// CLABE is a simple Value Object enforcing the 18-digit rule.
type CLABE struct {
	value string
}

// NewCLABE validates and normalizes a CLABE string.
func NewCLABE(raw string) (CLABE, error) {
	normalized := strings.ReplaceAll(raw, " ", "")
	if len(normalized) != 18 {
		return CLABE{}, ErrInvalidCLABE
	}
	for _, r := range normalized {
		if r < '0' || r > '9' {
			return CLABE{}, ErrInvalidCLABE
		}
	}
	return CLABE{value: normalized}, nil
}

func (c CLABE) String() string { return c.value }
