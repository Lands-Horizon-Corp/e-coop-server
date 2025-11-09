package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
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
		ledger, err := e.core.GeneralLedgerLatestLoanMemberAccount(
			context,
			memberProfile.ID,
			*loanTransaction.AccountID,
			memberProfile.OrganizationID,
			memberProfile.BranchID,
		)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to get latest general ledger for loan account id: %s", *loanTransaction.AccountID)
		}
		total = e.provider.Service.Decimal.Add(total, ledger.Balance)
	}
	return &total, nil
}
