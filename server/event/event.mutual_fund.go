package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func (e *Event) GenerateMutualFundEntries(
	context context.Context,
	userOrg *core.UserOrganization,
	mutualFund *core.MutualFund,
) ([]*core.MutualFundEntry, error) {
	result := []*core.MutualFundEntry{}
	if mutualFund == nil {
		return result, nil
	}
	now := time.Now().UTC()
	memberProfile, err := e.core.MemberProfileManager.Find(context, &core.MemberProfile{
		BranchID:       *userOrg.BranchID,
		OrganizationID: userOrg.OrganizationID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member profiles")
	}

	for _, profile := range memberProfile {
		if handlers.UUIDPtrEqual(&profile.ID, &mutualFund.MemberProfileID) {
			continue
		}
		if mutualFund.MemberType != nil && !handlers.UUIDPtrEqual(profile.MemberTypeID, mutualFund.MemberTypeID) {
			continue
		}

		amount := 0.0
		switch mutualFund.ComputationType {
		case core.ComputationTypeContinuous:
			amount = mutualFund.Amount
		case core.ComputationTypeUpToZero:
			memberAccuntingLedger, err := e.core.MemberAccountingLedgerManager.FindOne(context, &core.MemberAccountingLedger{
				MemberProfileID: profile.ID,
				AccountID:       *mutualFund.AccountID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			})
			if err != nil {
				if !eris.Is(err, gorm.ErrRecordNotFound) {
					return nil, eris.Wrap(err, "failed to find member accounting ledger")
				}
				amount = 0
			} else {
				currentBalance := memberAccuntingLedger.Balance
				benefitAmount := mutualFund.Amount
				var deduction float64
				if currentBalance >= benefitAmount {
					deduction = benefitAmount
				} else {
					deduction = currentBalance
				}
				amount = currentBalance - deduction
			}

		case core.ComputationTypeSufficient:

			memberAccuntingLedger, err := e.core.MemberAccountingLedgerManager.FindOne(context, &core.MemberAccountingLedger{
				MemberProfileID: profile.ID,
				AccountID:       *mutualFund.AccountID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			})
			if err != nil {
				if !eris.Is(err, gorm.ErrRecordNotFound) {
					return nil, eris.Wrap(err, "failed to find member accounting ledger")
				}
				amount = 0
			} else {
				currentBalance := memberAccuntingLedger.Balance
				benefitAmount := mutualFund.Amount
				if currentBalance >= benefitAmount {
					amount = benefitAmount
				} else {
					amount = 0
				}
			}

		case core.ComputationTypeByMembershipYear:
			monthsOfMembership := int(time.Since(profile.CreatedAt).Hours() / 24 / 30)
			for _, tier := range mutualFund.MutualFundTables {
				if monthsOfMembership >= tier.MonthFrom && monthsOfMembership <= tier.MonthTo {
					amount = tier.Amount
					break
				}
			}
			if amount == 0 {
				amount = mutualFund.TotalAmount
			}
		}

		if amount != 0 {
			entry := &core.MutualFundEntry{
				CreatedAt:       now,
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       now,
				UpdatedByID:     userOrg.UserID,
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				MemberProfileID: profile.ID,
				MemberProfile:   profile,
				Amount:          amount,
				Account:         mutualFund.Account,
				AccountID:       *mutualFund.AccountID,
				MutualFundID:    mutualFund.ID,
			}
			result = append(result, entry)
		}
	}
	return result, nil
}
