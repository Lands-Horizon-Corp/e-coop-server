package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func GenerateMutualFundEntries(
	context context.Context, service *horizon.HorizonService,
	userOrg *types.UserOrganization,
	mutualFund *types.MutualFund,
) ([]*types.MutualFundEntry, error) {
	result := []*types.MutualFundEntry{}
	if mutualFund == nil {
		return result, nil
	}
	now := time.Now().UTC()
	memberProfile, err := core.MemberProfileManager(service).Find(context, &types.MemberProfile{
		BranchID:       *userOrg.BranchID,
		OrganizationID: userOrg.OrganizationID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to find member profiles")
	}

	for _, profile := range memberProfile {
		if !profile.IsMutualFundMember {
			continue
		}
		if helpers.UUIDPtrEqual(&profile.ID, &mutualFund.MemberProfileID) {
			continue
		}
		if mutualFund.MemberType != nil && !helpers.UUIDPtrEqual(profile.MemberTypeID, mutualFund.MemberTypeID) {
			continue
		}

		amount := 0.0
		switch mutualFund.ComputationType {
		case types.ComputationTypeContinuous:
			amount = mutualFund.Amount
		case types.ComputationTypeUpToZero:
			memberAccuntingLedger, err := core.MemberAccountingLedgerManager(service).FindOne(context, &types.MemberAccountingLedger{
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

		case types.ComputationTypeSufficient:

			memberAccuntingLedger, err := core.MemberAccountingLedgerManager(service).FindOne(context, &types.MemberAccountingLedger{
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

		case types.ComputationTypeByMembershipYear:
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
			entry := &types.MutualFundEntry{
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
