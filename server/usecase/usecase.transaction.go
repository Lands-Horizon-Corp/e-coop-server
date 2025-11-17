package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

// TransactionService provides methods for handling financial transactions
// and balance calculations in the cooperative system.
type TransactionService struct {
	model    *core.Core
	provider *server.Provider
}

// NewTransactionService creates a new instance of TransactionService
// with the provided model core for database operations.
func NewTransactionService(
	model *core.Core,
	provider *server.Provider,
) (*TransactionService, error) {
	return &TransactionService{
		model:    model,
		provider: provider,
	}, nil
}
