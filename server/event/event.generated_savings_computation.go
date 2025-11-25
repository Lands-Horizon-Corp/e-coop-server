package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

func (e *Event) GenerateSavingsInterestComputation(context context.Context, useOrg core.UserOrganization, savings *core.GeneratedSavingsInterest) (
	[]*core.GeneratedSavingsInterestEntry, error) {
	if savings == nil {
		return nil, eris.New("savings is nil")
	}

	// memberAccountingLedger, err := e.core.MemberAccountingLedgerFilterByCriteria(
	// 	context,
	// 	useOrg.OrganizationID,
	// 	*useOrg.BranchID,
	// 	// Criteria
	// 	savings.AccountID,
	// 	savings.MemberTypeID,
	// )
	// if err != nil {
	// 	return nil, eris.Wrap(err, "failed to filter member accounting ledger")
	// }

	// get general ledger

	return []*core.GeneratedSavingsInterestEntry{}, nil
}

// Average Daily Balance (ADB)
