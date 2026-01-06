package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

type MemberAccountingLedgerSummary struct {
	TotalDeposits                     float64 `json:"total_deposits"`
	TotalShareCapitalPlusFixedSavings float64 `json:"total_share_capital_plus_fixed_savings"`
	TotalLoans                        float64 `json:"total_loans"`
}

func (e *Event) MemberAccountingLedgerSummary(
	context context.Context,
	ctx echo.Context,
	memberProfileID *uuid.UUID,
) (*MemberAccountingLedgerSummary, error) {
	userOrg, err := e.token.CurrentUserOrganization(context, ctx)
	if err != nil {
		return nil, eris.Wrap(err, "user authentication failed or organization not found")
	}
	if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
		return nil, eris.New("user is not authorized to view member general ledger totals")
	}

	if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
		return nil, eris.New("cash on hand account not set for branch")
	}
	if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
		return nil, eris.New("paid-up shared capital account not set for branch")
	}

	entries, err := e.core.MemberAccountingLedgerMemberProfileEntries(
		context,
		*memberProfileID,
		userOrg.OrganizationID,
		*userOrg.BranchID,
		*userOrg.Branch.BranchSetting.CashOnHandAccountID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve member accounting ledger entries")
	}

	paidUpShareCapital, err := e.core.MemberAccountingLedgerManager().Find(context, &core.MemberAccountingLedger{
		MemberProfileID: *memberProfileID,
		OrganizationID:  userOrg.OrganizationID,
		BranchID:        *userOrg.BranchID,
		AccountID:       *userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve paid-up share capital entries")
	}

	// Use shopspring decimal for precise accumulation
	totalShareCapitalDec := decimal.Zero
	for _, entry := range paidUpShareCapital {
		balDec := decimal.NewFromFloat(entry.Balance)
		totalShareCapitalDec = totalShareCapitalDec.Add(balDec)
	}

	totalDepositsDec := decimal.Zero
	for _, entry := range entries {
		balDec := decimal.NewFromFloat(entry.Balance)
		totalDepositsDec = totalDepositsDec.Add(balDec)
	}

	totalLoans, err := e.LoanTotalMemberProfile(context, *memberProfileID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to calculate total loans")
	}

	return &MemberAccountingLedgerSummary{
		TotalDeposits:                     totalDepositsDec.InexactFloat64(),
		TotalShareCapitalPlusFixedSavings: totalShareCapitalDec.InexactFloat64(),
		TotalLoans:                        *totalLoans,
	}, nil
}
