package main

// AccountType represents the type of account in the system
type AccountType string

// Account type constants
const (
	AccountTypeFines     AccountType = "Fines"
	AccountTypeInterest  AccountType = "Interest"
	AccountTypeSVFLedger AccountType = "SVF-Ledger"
)

type Account struct {
	Name        string
	AccountType AccountType
}

func main() {
	// payments []
}
