package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (e *Event) LoanTotalMemberProfile(context context.Context, memberProfileID uuid.UUID) (*float64, error) {
	memberProfile, err := e.core.MemberProfileManager.GetByID(context, memberProfileID)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get member profile by id: %s", memberProfileID)
	}
	loanTransactions, err := e.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
		MemberProfileID: &memberProfile.ID,
		OrganizationID:  memberProfile.OrganizationID,
		BranchID:        memberProfile.BranchID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to find loan transactions for member profile id: %s", memberProfileID)
	}

	total := 0.0
	for _, loanTransaction := range loanTransactions {
		generalLedgers, err := e.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
			AccountID:      loanTransaction.AccountID,
			OrganizationID: memberProfile.OrganizationID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to get latest general ledger for loan account id: %s", *loanTransaction.AccountID)
		}
		balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: generalLedgers,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for loan account id: %s", *loanTransaction.AccountID)
		}

		loanAccounts, err := e.core.LoanAccountManager.Find(context, &core.LoanAccount{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    memberProfile.OrganizationID,
			BranchID:          memberProfile.BranchID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to find loan accounts for loan transaction id: %s", loanTransaction.ID)
		}
		for _, loanAccount := range loanAccounts {
			balance := e.provider.Service.Decimal.ClampMin(
				e.provider.Service.Decimal.Add(loanAccount.Amount, loanAccount.TotalPayment), 0)
			total = e.provider.Service.Decimal.Add(total, balance)
		}

		total = e.provider.Service.Decimal.Add(total, balance.Balance)
	}

	return &total, nil
}
