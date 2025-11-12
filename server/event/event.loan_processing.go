package event

import (
	"context"

	"github.com/google/uuid"
)

func (e *Event) LoanProcessing(ctx context.Context, loanTransactionID *uuid.UUID)
