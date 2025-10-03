package service

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/rotisserie/eris"
)

func (t *TransactionService) LoanComputation(ctx context.Context, ald model.AutomaticLoanDeduction, lt model.LoanTransaction) float64 {
	result := lt.Applied1
	// --- Min/Max check ---
	if ald.MinAmount > 0 && result < ald.MinAmount {
		return 0.0
	}

	if ald.MaxAmount > 0 && result > ald.MaxAmount {
		return 0.0
	}
	// --- Percentage application ---
	if ald.ChargesPercentage1 > 0 || ald.ChargesPercentage2 > 0 {
		if ald.ChargesPercentage1 > 0 && ald.ChargesPercentage2 > 0 {
			if ald.AddOn {
				result *= ald.ChargesPercentage2 / 100
			} else {
				result *= ald.ChargesPercentage1 / 100
			}
		} else if ald.ChargesPercentage1 > 0 {
			result *= ald.ChargesPercentage1 / 100
		} else {
			result *= ald.ChargesPercentage2 / 100
		}
	}

	// --- Divisor application ---
	if ald.ChargesDivisor > 0 && result > 0 {
		result = (result / ald.ChargesDivisor) * ald.ChargesAmount
	}

	// --- Annum adjustments (when months = 0) ---
	if ald.NumberOfMonths == 0 {
		if ald.Anum == 1 {
			result /= 12
		}
	}

	// --- Number of months adjustments ---
	if ald.NumberOfMonths == -1 {
		result = (result * float64(lt.Terms)) / 12
	} else if ald.NumberOfMonths > 0 {
		result = (result * float64(lt.Terms)) / float64(ald.NumberOfMonths)
	}

	if result == lt.Applied1 {
		return ald.ChargesAmount
	}

	return result
}

func (t *TransactionService) LoanBalancing(ctx context.Context, entries []*model.LoanTransactionEntry) ([]*model.LoanTransactionEntry, error) {
	static, addOn, deduction := []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}, []*model.LoanTransactionEntry{}
	for _, entries := range entries {
		if entries.Type == model.LoanTransactionStatic {
			static = append(static, entries)
		}
		if entries.Type == model.LoanTransactionAddOn {
			addOn = append(addOn, entries)
		}
		if entries.Type == model.LoanTransactionDeduction {
			deduction = append(deduction, entries)
		}
	}
	if len(static) <= 1 {
		return nil, eris.New("at least 2 static entries are required")
	}

	return nil, nil
}
