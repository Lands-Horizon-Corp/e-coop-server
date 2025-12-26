package event

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanAdjustment(
	context context.Context,
	userOrg core.UserOrganization,
	la core.LoanTransactionAdjustmentRequest,
) error {

	fmt.Println("[LoanAdjustment] START")
	fmt.Printf("[LoanAdjustment] UserOrgID=%v LoanAccountID=%v AdjustmentType=%v Amount=%v\n",
		userOrg.ID, la.LoanAccountID, la.AdjustmentType, la.Amount,
	)

	tx, endTx := e.provider.Service.Database.StartTransaction(context)
	fmt.Println("[LoanAdjustment] Transaction started")

	loanAccount, err := e.core.LoanAccountManager().GetByIDLock(context, tx, la.LoanAccountID, "Account")
	if err != nil {
		fmt.Printf("[LoanAdjustment] ERROR fetching LoanAccount: %v\n", err)
		return endTx(eris.Wrap(err, "Account not found for adjustment"))
	}

	fmt.Printf(
		"[LoanAdjustment] LoanAccount loaded: ID=%v Amount=%v TotalAdd=%v TotalDeduction=%v\n",
		loanAccount.ID,
		loanAccount.Amount,
		loanAccount.TotalAdd,
		loanAccount.TotalDeduction,
	)

	loanTransaction, err := e.core.LoanTransactionManager().GetByIDLock(
		context,
		tx,
		loanAccount.LoanTransactionID,
	)
	if err != nil {
		fmt.Printf("[LoanAdjustment] ERROR fetching LoanTransaction: %v\n", err)
		return endTx(eris.Wrap(err, "Loan transaction not found for adjustment"))
	}

	fmt.Printf(
		"[LoanAdjustment] LoanTransaction loaded: ID=%v ReleasedDate=%v\n",
		loanTransaction.ID,
		loanTransaction.ReleasedDate,
	)

	if loanTransaction.ReleasedDate == nil {
		fmt.Println("[LoanAdjustment] ERROR: Loan transaction not released")
		return endTx(eris.New("Cannot adjust loan account for unreleased loan transaction"))
	}

	amount := 0.0

	switch la.AdjustmentType {
	case core.LoanAdjustmentTypeAdd:
		fmt.Println("[LoanAdjustment] AdjustmentType = ADD")
		loanAccount.TotalAdd += la.Amount
		loanAccount.TotalAddCount += 1
		amount = la.Amount

	case core.LoanAdjustmentTypeDeduct:
		fmt.Println("[LoanAdjustment] AdjustmentType = DEDUCT")
		loanAccount.TotalDeduction += la.Amount
		loanAccount.TotalDeductionCount += 1
		amount = -la.Amount

	default:
		fmt.Printf("[LoanAdjustment] ERROR: Invalid AdjustmentType=%v\n", la.AdjustmentType)
		return endTx(eris.New("Invalid adjustment type specified"))
	}

	fmt.Printf(
		"[LoanAdjustment] Calculated adjustment amount=%v (before balance=%v)\n",
		amount,
		loanAccount.Amount,
	)

	loanAccount.Amount += amount

	fmt.Printf(
		"[LoanAdjustment] Updated LoanAccount Amount=%v TotalAdd=%v TotalDeduction=%v\n",
		loanAccount.Amount,
		loanAccount.TotalAdd,
		loanAccount.TotalDeduction,
	)

	if err := e.core.LoanAccountManager().UpdateByIDWithTx(
		context,
		tx,
		loanAccount.ID,
		loanAccount,
	); err != nil {
		fmt.Printf("[LoanAdjustment] ERROR updating LoanAccount: %v\n", err)
		return endTx(eris.Wrap(err, "Failed to update loan account balance"))
	}

	fmt.Println("[LoanAdjustment] LoanAccount updated successfully")

	if err := endTx(nil); err != nil {
		fmt.Printf("[LoanAdjustment] ERROR committing transaction: %v\n", err)
		return eris.Wrap(err, "Failed to save loan adjustment changes")
	}

	fmt.Println("[LoanAdjustment] SUCCESS - transaction committed")
	return nil
}
