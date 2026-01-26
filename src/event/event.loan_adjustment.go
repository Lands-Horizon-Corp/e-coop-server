package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func LoanAdjustment(
	context context.Context, service *horizon.HorizonService,
	userOrg types.UserOrganization,
	la types.LoanTransactionAdjustmentRequest,
) error {
	tx, endTx := service.Database.StartTransaction(context)
	loanAccount, err := core.LoanAccountManager(service).GetByIDLock(context, tx, la.LoanAccountID, "Account")
	if err != nil {
		return endTx(eris.Wrap(err, "Account not found for adjustment"))
	}
	loanTransaction, err := core.LoanTransactionManager(service).GetByIDLock(
		context,
		tx,
		loanAccount.LoanTransactionID,
	)
	if err != nil {
		return endTx(eris.Wrap(err, "Loan transaction not found for adjustment"))
	}

	if loanTransaction.ReleasedDate == nil {
		return endTx(eris.New("Cannot adjust loan account for unreleased loan transaction"))
	}

	amount := 0.0

	switch la.AdjustmentType {
	case types.LoanAdjustmentTypeAdd:
		loanAccount.TotalAdd += la.Amount
		loanAccount.TotalAddCount += 1
		amount = la.Amount

	case types.LoanAdjustmentTypeDeduct:
		loanAccount.TotalDeduction += la.Amount
		loanAccount.TotalDeductionCount += 1
		amount = -la.Amount

	default:
		return endTx(eris.New("Invalid adjustment type specified"))
	}

	loanAccount.Amount += amount
	if err := core.LoanAccountManager(service).UpdateByIDWithTx(
		context,
		tx,
		loanAccount.ID,
		loanAccount,
	); err != nil {
		return endTx(eris.Wrap(err, "Failed to update loan account balance"))
	}
	if err := endTx(nil); err != nil {

		return eris.Wrap(err, "Failed to save loan adjustment changes")
	}
	return nil
}
