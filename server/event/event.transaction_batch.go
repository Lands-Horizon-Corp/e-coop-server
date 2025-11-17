package event

import (
	"context"

	"github.com/google/uuid"
)

func (m *Event) TransactionBatchBalancing(context context.Context, transactionBatchID uuid.UUID) error {

	// transactionBatch, err := m.core.TransactionBatchManager.GetByID(context, transactionBatchID)
	// if err != nil {
	// 	return eris.Wrap(err, "failed to get transaction batch by ID")
	// }

	// Adjustment entry
	// Loan release
	// Cash Check voucher
	return nil
}
