package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanProcessing(context context.Context, ctx echo.Context, loanTransactionID *uuid.UUID) (*core.LoanTransaction, error) {
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, eris.Wrap(err, "loan processing: failed to get loan transaction by id")
	}
	return loanTransaction, nil
}
