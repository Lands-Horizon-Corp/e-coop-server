package service

import (
	"context"
	"errors"
	"math"

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

func (t *TransactionService) LoanModeOfPayment(ctx context.Context, lt *model.LoanTransaction) (float64, error) {
	switch lt.ModeOfPayment {
	case model.LoanModeOfPaymentDaily:
		return lt.Applied1 / float64(lt.Terms) / 30, nil
	case model.LoanModeOfPaymentWeekly:
		return lt.Applied1 / float64(lt.Terms) / 4, nil
	case model.LoanModeOfPaymentSemiMonthly:
		return lt.Applied1 / float64(lt.Terms) / 2, nil
	case model.LoanModeOfPaymentMonthly:
		return lt.Applied1 / float64(lt.Terms), nil
	case model.LoanModeOfPaymentQuarterly:
		return lt.Applied1 / (float64(lt.Terms) / 3), nil
	case model.LoanModeOfPaymentSemiAnnual:
		return lt.Applied1 / (float64(lt.Terms) / 6), nil
	case model.LoanModeOfPaymentLumpsum:
		return lt.Applied1, nil
	case model.LoanModeOfPaymentFixedDays:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		if lt.ModeOfPaymentFixedDays <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		return lt.Applied1 / float64(lt.Terms), nil
	}
	return 0, eris.New("not implemented yet")
}

func (t *TransactionService) SuggestedNumberOfTerms(
	ctx context.Context,
	suggestedAmount float64,
	principal float64,
	modeOfPayment model.LoanModeOfPayment,
	fixedDays int,
) (int, error) {
	if suggestedAmount <= 0 {
		return 0, errors.New("suggested amount must be greater than zero")
	}
	if principal <= 0 {
		return 0, errors.New("invalid total loan amount")
	}

	var terms float64

	switch modeOfPayment {
	case model.LoanModeOfPaymentDaily:
		// daily = total / (payment * 30)
		terms = (principal / suggestedAmount) / 30
	case model.LoanModeOfPaymentWeekly:
		// weekly = total / (payment * 4)
		terms = (principal / suggestedAmount) / 4
	case model.LoanModeOfPaymentSemiMonthly:
		// semi-monthly = total / (payment * 2)
		terms = (principal / suggestedAmount) / 2
	case model.LoanModeOfPaymentMonthly:
		// monthly = total / payment
		terms = principal / suggestedAmount
	case model.LoanModeOfPaymentQuarterly:
		// quarterly = total / (payment / 3)
		terms = (principal / suggestedAmount) * 3
	case model.LoanModeOfPaymentSemiAnnual:
		// semi-annual = total / (payment / 6)
		terms = (principal / suggestedAmount) * 6
	case model.LoanModeOfPaymentLumpsum:
		terms = 1
	case model.LoanModeOfPaymentFixedDays:
		if fixedDays <= 0 {
			return 0, errors.New("invalid fixed days: must be greater than 0")
		}
		terms = principal / suggestedAmount
	default:
		return 0, errors.New("unsupported mode of payment")
	}

	numberOfTerms := int(math.Ceil(terms))
	if numberOfTerms < 1 {
		numberOfTerms = 1
	}
	return numberOfTerms, nil
}
