package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanAdjustment(
	context context.Context,
	userOrg core.UserOrganization,
	la core.LoanTransactionAdjustmentRequest,
) error {

	tx, endTx := e.provider.Service.Database.StartTransaction(context)

	loanAccount, err := e.core.LoanAccountManager.GetByIDLock(context, tx, la.LoanAccount, "Account")
	if err != nil {
		return endTx(eris.Wrap(err, "Account not found for adjustment"))
	}

	loanTransaction, err := e.core.LoanTransactionManager.GetByIDLock(context, tx, loanAccount.LoanTransactionID)
	if err != nil {
		return endTx(eris.Wrap(err, "Loan transaction not found for adjustment"))
	}
	if loanTransaction.ReleasedDate == nil {
		return endTx(eris.New("Cannot adjust loan account for unreleased loan transaction"))
	}

	amount := 0.0
	switch la.AdjustmentType {
	case core.LoanAdjustmentTypeAdd:
		loanAccount.TotalAdd += la.Amount
		loanAccount.TotalAddCount += 1
		amount = la.Amount
	case core.LoanAdjustmentTypeDeduct:
		loanAccount.TotalDeduction += la.Amount
		loanAccount.TotalDeductionCount += 1
		amount = -la.Amount
	default:
		return endTx(eris.New("Invalid adjustment type specified"))
	}

	loanAccount.Amount += amount
	if err := e.core.LoanAccountManager.UpdateByIDWithTx(context, tx, loanAccount.ID, loanAccount); err != nil {
		return endTx(eris.Wrap(err, "Failed to update loan account balance"))
	}

	if err := endTx(nil); err != nil {
		return eris.Wrap(err, "Failed to save loan adjustment changes")
	}

	return nil
}
