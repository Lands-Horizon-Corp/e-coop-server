package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func LoanTotalMemberProfile(context context.Context, service *horizon.HorizonService, memberProfileID uuid.UUID) (*float64, error) {
	memberProfile, err := core.MemberProfileManager(service).GetByID(context, memberProfileID)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get member profile by id: %s", memberProfileID)
	}

	loanTransactions, err := core.LoanTransactionManager(service).Find(context, &core.LoanTransaction{
		MemberProfileID: &memberProfile.ID,
		OrganizationID:  memberProfile.OrganizationID,
		BranchID:        memberProfile.BranchID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to find loan transactions for member profile id: %s", memberProfileID)
	}

	totalDec := decimal.Zero

	for _, loanTransaction := range loanTransactions {
		generalLedgers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
			AccountID:      loanTransaction.AccountID,
			OrganizationID: memberProfile.OrganizationID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to get latest general ledger for loan account id: %s", *loanTransaction.AccountID)
		}

		balance, err := usecase.CalculateBalance(usecase.Balance{
			GeneralLedgers: generalLedgers,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for loan account id: %s", *loanTransaction.AccountID)
		}

		loanAccounts, err := core.LoanAccountManager(service).Find(context, &core.LoanAccount{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    memberProfile.OrganizationID,
			BranchID:          memberProfile.BranchID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to find loan accounts for loan transaction id: %s", loanTransaction.ID)
		}

		for _, loanAccount := range loanAccounts {
			amountDec := decimal.NewFromFloat(loanAccount.Amount)
			totalPaymentDec := decimal.NewFromFloat(loanAccount.TotalPayment)

			balanceDec := amountDec.Add(totalPaymentDec)
			if balanceDec.LessThan(decimal.Zero) {
				balanceDec = decimal.Zero
			}

			totalDec = totalDec.Add(balanceDec)
		}

		// Add computed balance from CalculateBalance
		balanceDec := decimal.NewFromFloat(balance.Balance)
		totalDec = totalDec.Add(balanceDec)
	}

	totalFloat := totalDec.InexactFloat64()
	return &totalFloat, nil
}
