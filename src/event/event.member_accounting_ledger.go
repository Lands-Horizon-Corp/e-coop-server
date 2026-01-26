package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
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

func MemberAccountingLedger(
	context context.Context, service *horizon.HorizonService,
	ctx echo.Context,
	memberProfileID *uuid.UUID,
) (*MemberAccountingLedgerSummary, error) {
	userOrg, err := CurrentUserOrganization(context, service, ctx)
	if err != nil {
		return nil, eris.Wrap(err, "user authentication failed or organization not found")
	}
	if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
		return nil, eris.New("user is not authorized to view member general ledger totals")
	}

	if userOrg.Branch.BranchSetting.CashOnHandAccountID == nil {
		return nil, eris.New("cash on hand account not set for branch")
	}
	if userOrg.Branch.BranchSetting.PaidUpSharedCapitalAccountID == nil {
		return nil, eris.New("paid-up shared capital account not set for branch")
	}

	entries, err := core.MemberAccountingLedgerMemberProfileEntries(
		context, service,
		*memberProfileID,
		userOrg.OrganizationID,
		*userOrg.BranchID,
		*userOrg.Branch.BranchSetting.CashOnHandAccountID,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve member accounting ledger entries")
	}

	paidUpShareCapital, err := core.MemberAccountingLedgerManager(service).Find(context, &types.MemberAccountingLedger{
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

	totalLoans, err := LoanTotalMemberProfile(context, service, *memberProfileID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to calculate total loans")
	}

	return &MemberAccountingLedgerSummary{
		TotalDeposits:                     totalDepositsDec.InexactFloat64(),
		TotalShareCapitalPlusFixedSavings: totalShareCapitalDec.InexactFloat64(),
		TotalLoans:                        *totalLoans,
	}, nil
}
