package domain

import "testing"

func TestAccountInvariants(t *testing.T) {
	clabe, err := NewCLABE("032180000118359719")
	if err != nil { t.Fatalf("clabe: %v", err) }

	account, err := NewAccount("acc-1", "Alice", clabe)
	if err != nil { t.Fatalf("new account: %v", err) }

	if err := account.Credit(1000); err != nil { t.Fatalf("credit: %v", err) }
	if got := account.Balance(); got != 1000 { t.Fatalf("want=1000 got=%d", got) }

	if err := account.Debit(500); err != nil { t.Fatalf("debit: %v", err) }
	if got := account.Balance(); got != 500 { t.Fatalf("want=500 got=%d", got) }

	if err := account.Debit(600); err == nil { t.Fatalf("expected insufficient funds") }
}
