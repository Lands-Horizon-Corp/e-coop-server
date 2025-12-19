package event

import (
	"context"
	"fmt"
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
	fmt.Printf("[DEBUG] GenerateMutualFundEntries called for MutualFund ID=%v, Amount=%v, ComputationType=%v\n",
		mutualFund.ID, mutualFund.Amount, mutualFund.ComputationType)

	result := []*core.MutualFundEntry{}
	if mutualFund == nil {
		fmt.Printf("[DEBUG] mutualFund is nil, returning empty result\n")
		return result, nil
	}

	now := time.Now().UTC()
	fmt.Printf("[DEBUG] Current time (UTC): %v\n", now)

	memberProfile, err := e.core.MemberProfileManager.Find(context, &core.MemberProfile{
		BranchID:       *userOrg.BranchID,
		OrganizationID: userOrg.OrganizationID,
	})
	if err != nil {
		fmt.Printf("[DEBUG] ERROR finding member profiles: %v\n", err)
		return nil, eris.Wrap(err, "failed to find member profiles")
	}
	fmt.Printf("[DEBUG] Found %d member profiles\n", len(memberProfile))

	for _, profile := range memberProfile {
		fmt.Printf("[DEBUG] Processing profile ID=%v, CreatedAt=%v\n", profile.ID, profile.CreatedAt)

		if handlers.UUIDPtrEqual(&profile.ID, &mutualFund.MemberProfileID) {
			fmt.Printf("[DEBUG] Skipping own profile (ID matches MemberProfileID)\n")
			continue
		}
		if mutualFund.MemberType != nil && !handlers.UUIDPtrEqual(profile.MemberTypeID, mutualFund.MemberTypeID) {
			fmt.Printf("[DEBUG] Skipping profile due to MemberType mismatch\n")
			continue
		}

		amount := 0.0
		fmt.Printf("[DEBUG] Starting amount calculation for profile ID=%v, ComputationType=%v\n", profile.ID, mutualFund.ComputationType)

		switch mutualFund.ComputationType {
		case core.ComputationTypeContinuous:
			amount = mutualFund.Amount
			fmt.Printf("[DEBUG] ComputationTypeContinuous: amount set to %v\n", amount)

		case core.ComputationTypeUpToZero:
			fmt.Printf("[DEBUG] Entering ComputationTypeUpToZero branch\n")
			memberAccuntingLedger, err := e.core.MemberAccountingLedgerManager.FindOne(context, &core.MemberAccountingLedger{
				MemberProfileID: profile.ID,
				AccountID:       *mutualFund.AccountID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			})
			if err != nil {
				if !eris.Is(err, gorm.ErrRecordNotFound) {
					fmt.Printf("[DEBUG] ERROR finding ledger (UpToZero): %v\n", err)
					return nil, eris.Wrap(err, "failed to find member accounting ledger")
				}
				fmt.Printf("[DEBUG] Ledger not found (UpToZero), treating balance as 0\n")
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
				fmt.Printf("[DEBUG] UpToZero: balance=%v, benefit=%v, deduction=%v → final amount=%v\n",
					currentBalance, benefitAmount, deduction, amount)
			}

		case core.ComputationTypeSufficient:
			fmt.Printf("[DEBUG] Entering ComputationTypeSufficient branch\n")
			memberAccuntingLedger, err := e.core.MemberAccountingLedgerManager.FindOne(context, &core.MemberAccountingLedger{
				MemberProfileID: profile.ID,
				AccountID:       *mutualFund.AccountID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			})
			if err != nil {
				if !eris.Is(err, gorm.ErrRecordNotFound) {
					fmt.Printf("[DEBUG] ERROR finding ledger (Sufficient): %v\n", err)
					return nil, eris.Wrap(err, "failed to find member accounting ledger")
				}
				fmt.Printf("[DEBUG] Ledger not found (Sufficient), amount=0\n")
				amount = 0
			} else {
				currentBalance := memberAccuntingLedger.Balance
				benefitAmount := mutualFund.Amount
				if currentBalance >= benefitAmount {
					amount = benefitAmount
					fmt.Printf("[DEBUG] Sufficient: balance=%v >= benefit=%v → amount=%v\n", currentBalance, benefitAmount, amount)
				} else {
					amount = 0
					fmt.Printf("[DEBUG] Sufficient: balance=%v < benefit=%v → amount=0\n", currentBalance, benefitAmount)
				}
			}

		case core.ComputationTypeByMembershipYear:
			fmt.Printf("[DEBUG] Entering ComputationTypeByMembershipYear branch\n")
			monthsOfMembership := int(time.Since(profile.CreatedAt).Hours() / 24 / 30)
			fmt.Printf("[DEBUG] Months of membership calculated: %d\n", monthsOfMembership)
			for _, tier := range mutualFund.MutualFundTables {
				if monthsOfMembership >= tier.MonthFrom && monthsOfMembership <= tier.MonthTo {
					amount = tier.Amount
					fmt.Printf("[DEBUG] Tier matched: MonthFrom=%d, MonthTo=%d → amount=%v\n", tier.MonthFrom, tier.MonthTo, amount)
					break
				}
			}
			if amount == 0 {
				amount = mutualFund.TotalAmount
				fmt.Printf("[DEBUG] No tier matched, falling back to TotalAmount=%v\n", amount)
			}

		default:
			fmt.Printf("[DEBUG] Unknown ComputationType: %v, amount remains 0\n", mutualFund.ComputationType)
		}

		fmt.Printf("[DEBUG] Final amount for profile ID=%v: %v\n", profile.ID, amount)

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
			fmt.Printf("[DEBUG] Added MutualFundEntry for profile ID=%v with amount=%v (total entries now: %d)\n",
				profile.ID, amount, len(result))
		} else {
			fmt.Printf("[DEBUG] Amount is 0, skipping entry for profile ID=%v\n", profile.ID)
		}
	}

	fmt.Printf("[DEBUG] GenerateMutualFundEntries completed: generated %d entries\n", len(result))
	return result, nil
}
