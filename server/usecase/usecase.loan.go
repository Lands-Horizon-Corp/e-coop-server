package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
)

// LoanChargesRateComputation calculates the loan charges based on the rate scheme and loan transaction
func (t *TransactionService) LoanChargesRateComputation(_ context.Context, crs core.ChargesRateScheme, ald core.LoanTransaction) float64 {

	result := 0.0

	termHeaders := []int{
		crs.ByTermHeader1,
		crs.ByTermHeader2,
		crs.ByTermHeader3,
		crs.ByTermHeader4,
		crs.ByTermHeader5,
		crs.ByTermHeader6,
		crs.ByTermHeader7,
		crs.ByTermHeader8,
		crs.ByTermHeader9,
		crs.ByTermHeader10,
		crs.ByTermHeader11,
		crs.ByTermHeader12,
		crs.ByTermHeader13,
		crs.ByTermHeader14,
		crs.ByTermHeader15,
		crs.ByTermHeader16,
		crs.ByTermHeader17,
		crs.ByTermHeader18,
		crs.ByTermHeader19,
		crs.ByTermHeader20,
		crs.ByTermHeader21,
		crs.ByTermHeader22,
	}
	modeOfPaymentHeaders := []int{
		crs.ModeOfPaymentHeader1,
		crs.ModeOfPaymentHeader2,
		crs.ModeOfPaymentHeader3,
		crs.ModeOfPaymentHeader4,
		crs.ModeOfPaymentHeader5,
		crs.ModeOfPaymentHeader6,
		crs.ModeOfPaymentHeader7,
		crs.ModeOfPaymentHeader8,
		crs.ModeOfPaymentHeader9,
		crs.ModeOfPaymentHeader10,
		crs.ModeOfPaymentHeader11,
		crs.ModeOfPaymentHeader12,
		crs.ModeOfPaymentHeader13,
		crs.ModeOfPaymentHeader14,
		crs.ModeOfPaymentHeader15,
		crs.ModeOfPaymentHeader16,
		crs.ModeOfPaymentHeader17,
		crs.ModeOfPaymentHeader18,
		crs.ModeOfPaymentHeader19,
		crs.ModeOfPaymentHeader20,
		crs.ModeOfPaymentHeader21,
		crs.ModeOfPaymentHeader22,
	}

	findLastApplicableRate := func(rates []float64, headers []int, terms int) float64 {
		lastRate := 0.0
		minLen := min(len(rates), len(headers))
		for i := range minLen {
			rate := rates[i]
			term := headers[i]
			if term > terms || rate <= 0 {
				break
			}
			lastRate = rate
		}
		return lastRate
	}

	computeCharge := func(applied, rate float64, mode core.LoanModeOfPayment) float64 {
		if rate <= 0 {
			return 0.0
		}
		// Use precise decimal arithmetic for percentage calculation
		base := t.provider.Service.Decimal.MultiplyByPercentage(applied, rate)

		switch mode {
		case core.LoanModeOfPaymentDaily:
			return t.provider.Service.Decimal.Divide(base, 30.0)
		case core.LoanModeOfPaymentWeekly:
			weeklyBase := t.provider.Service.Decimal.Multiply(base, 7.0)
			return t.provider.Service.Decimal.Divide(weeklyBase, 30.0)
		case core.LoanModeOfPaymentSemiMonthly:
			semiMonthlyBase := t.provider.Service.Decimal.Multiply(base, 15.0)
			return t.provider.Service.Decimal.Divide(semiMonthlyBase, 30.0)
		case core.LoanModeOfPaymentMonthly:
			return base
		case core.LoanModeOfPaymentQuarterly:
			return t.provider.Service.Decimal.Multiply(base, 3.0)
		case core.LoanModeOfPaymentSemiAnnual:
			return t.provider.Service.Decimal.Multiply(base, 6.0)
		default:
			return 0.0
		}
	}

	switch crs.Type {
	case core.ChargesRateSchemeTypeByRange:
		for _, data := range crs.ChargesRateByRangeOrMinimumAmounts {
			if ald.Applied1 < data.From || ald.Applied1 > data.To {
				continue
			}
			charge := 0.0
			if data.Charge > 0 {
				// Use precise decimal arithmetic for percentage calculation
				charge = t.provider.Service.Decimal.MultiplyByPercentage(ald.Applied1, data.Charge)
			} else if data.Amount > 0 {
				charge = data.Amount
			}
			if charge > 0 {
				result = charge

				if result >= data.MinimumAmount && data.MinimumAmount > 0 {

					result = data.MinimumAmount
				}
				return result
			}
		}
	case core.ChargesRateSchemeTypeByType:

		if crs.MemberType != nil && ald.MemberProfile.MemberTypeID != &crs.MemberType.ID {
			return 0.0
		}
		if crs.ModeOfPayment != nil && ald.ModeOfPayment != *crs.ModeOfPayment {
			return 0.0
		}
		for _, data := range crs.ChargesRateSchemeModeOfPayments {
			if ald.Applied1 < data.From || ald.Applied1 > data.To {
				continue
			}
			chargesTerms := []float64{
				data.Column1,
				data.Column2,
				data.Column3,
				data.Column4,
				data.Column5,
				data.Column6,
				data.Column7,
				data.Column8,
				data.Column9,
				data.Column10,
				data.Column11,
				data.Column12,
				data.Column13,
				data.Column14,
				data.Column15,
				data.Column16,
				data.Column17,
				data.Column18,
				data.Column19,
				data.Column20,
				data.Column21,
				data.Column22,
			}
			lastRate := findLastApplicableRate(chargesTerms, modeOfPaymentHeaders, ald.Terms)

			if lastRate == 0.0 {
				continue
			}
			result = computeCharge(ald.Applied1, lastRate, ald.ModeOfPayment)
			if result > 0 {
				return result
			}
		}
	case core.ChargesRateSchemeTypeByTerm:
		if ald.Terms < 1 {
			return 0.0
		}
		for _, data := range crs.ChargesRateByTerms {
			if data.ModeOfPayment != ald.ModeOfPayment {
				continue
			}
			chargesTerms := []float64{
				data.Rate1,
				data.Rate2,
				data.Rate3,
				data.Rate4,
				data.Rate5,
				data.Rate6,
				data.Rate7,
				data.Rate8,
				data.Rate9,
				data.Rate10,
				data.Rate11,
				data.Rate12,
				data.Rate13,
				data.Rate14,
				data.Rate15,
				data.Rate16,
				data.Rate17,
				data.Rate18,
				data.Rate19,
				data.Rate20,
				data.Rate21,
				data.Rate22,
			}
			lastRate := findLastApplicableRate(chargesTerms, termHeaders, ald.Terms)
			if lastRate == 0.0 {
				continue
			}
			result = computeCharge(ald.Applied1, lastRate, ald.ModeOfPayment)
			if result > 0 {
				return result
			}
		}
	}
	return result
}

// LoanNumberOfPayments calculates the total number of payments for a loan based on terms and payment mode
func (t *TransactionService) LoanNumberOfPayments(mp core.LoanModeOfPayment, terms int) (int, error) {
	switch mp {
	case core.LoanModeOfPaymentDaily:
		return terms * 30, nil
	case core.LoanModeOfPaymentWeekly:
		return terms * 4, nil
	case core.LoanModeOfPaymentSemiMonthly:
		return terms * 2, nil
	case core.LoanModeOfPaymentMonthly:
		return terms, nil
	case core.LoanModeOfPaymentQuarterly:
		return terms / 3, nil
	case core.LoanModeOfPaymentSemiAnnual:
		return terms / 6, nil
	case core.LoanModeOfPaymentLumpsum:
		return 1, nil
	case core.LoanModeOfPaymentFixedDays:
		if terms <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		return terms, nil
	}
	return 0, eris.New("not implemented yet")
}

// LoanComputation calculates the loan amount after applying automatic loan deduction rules using precise decimal arithmetic
func (t *TransactionService) LoanComputation(ald core.AutomaticLoanDeduction, lt core.LoanTransaction) float64 {
	result := lt.Applied1

	// --- Min/Max check ---
	if ald.MinAmount > 0 && result < ald.MinAmount {
		return 0.0
	}

	if ald.MaxAmount > 0 && result > ald.MaxAmount {
		return 0.0
	}

	// --- Percentage application using decimal arithmetic ---
	if ald.ChargesPercentage1 > 0 || ald.ChargesPercentage2 > 0 {
		switch {
		case ald.ChargesPercentage1 > 0 && ald.ChargesPercentage2 > 0:
			if ald.AddOn {
				result = t.provider.Service.Decimal.MultiplyByPercentage(result, ald.ChargesPercentage2)
			} else {
				result = t.provider.Service.Decimal.MultiplyByPercentage(result, ald.ChargesPercentage1)
			}
		case ald.ChargesPercentage1 > 0:
			result = t.provider.Service.Decimal.MultiplyByPercentage(result, ald.ChargesPercentage1)
		default:
			result = t.provider.Service.Decimal.MultiplyByPercentage(result, ald.ChargesPercentage2)
		}
	}

	// --- Divisor application using decimal arithmetic ---
	if ald.ChargesDivisor > 0 && result > 0 {
		// result = (result / ald.ChargesDivisor) * ald.ChargesAmount
		dividedResult := t.provider.Service.Decimal.Divide(result, ald.ChargesDivisor)
		result = t.provider.Service.Decimal.Multiply(dividedResult, ald.ChargesAmount)
	}

	// --- Annum adjustments (when months = 0) using decimal arithmetic ---
	if ald.NumberOfMonths == 0 {
		if ald.Anum == 1 {
			result = t.provider.Service.Decimal.Divide(result, 12)
		}
	}

	// --- Number of months adjustments using decimal arithmetic ---
	if ald.NumberOfMonths == -1 {
		// result = (result * float64(lt.Terms)) / 12
		multipliedResult := t.provider.Service.Decimal.Multiply(result, float64(lt.Terms))
		result = t.provider.Service.Decimal.Divide(multipliedResult, 12)
	} else if ald.NumberOfMonths > 0 {
		// result = (result * float64(lt.Terms)) / float64(ald.NumberOfMonths)
		multipliedResult := t.provider.Service.Decimal.Multiply(result, float64(lt.Terms))
		result = t.provider.Service.Decimal.Divide(multipliedResult, float64(ald.NumberOfMonths))
	}

	if result == lt.Applied1 {
		return ald.ChargesAmount
	}

	// Round to 2 decimal places for financial precision
	return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2)
}

// LoanModeOfPayment calculates the payment amount per period based on loan terms and mode of payment using precise decimal arithmetic
func (t *TransactionService) LoanModeOfPayment(_ context.Context, lt *core.LoanTransaction) (float64, error) {
	switch lt.ModeOfPayment {
	case core.LoanModeOfPaymentDaily:
		// lt.Applied1 / float64(lt.Terms) / 30
		termsDivision := t.provider.Service.Decimal.Divide(lt.Applied1, float64(lt.Terms))
		result := t.provider.Service.Decimal.Divide(termsDivision, 30)
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentWeekly:
		// lt.Applied1 / float64(lt.Terms) / 4
		termsDivision := t.provider.Service.Decimal.Divide(lt.Applied1, float64(lt.Terms))
		result := t.provider.Service.Decimal.Divide(termsDivision, 4)
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentSemiMonthly:
		// lt.Applied1 / float64(lt.Terms) / 2
		termsDivision := t.provider.Service.Decimal.Divide(lt.Applied1, float64(lt.Terms))
		result := t.provider.Service.Decimal.Divide(termsDivision, 2)
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentMonthly:
		// lt.Applied1 / float64(lt.Terms)
		result := t.provider.Service.Decimal.Divide(lt.Applied1, float64(lt.Terms))
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentQuarterly:
		// lt.Applied1 / (float64(lt.Terms) / 3)
		termsDivision := t.provider.Service.Decimal.Divide(float64(lt.Terms), 3)
		result := t.provider.Service.Decimal.Divide(lt.Applied1, termsDivision)
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentSemiAnnual:
		// lt.Applied1 / (float64(lt.Terms) / 6)
		termsDivision := t.provider.Service.Decimal.Divide(float64(lt.Terms), 6)
		result := t.provider.Service.Decimal.Divide(lt.Applied1, termsDivision)
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	case core.LoanModeOfPaymentLumpsum:
		return t.provider.Service.Decimal.RoundToDecimalPlaces(lt.Applied1, 2), nil
	case core.LoanModeOfPaymentFixedDays:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		if lt.ModeOfPaymentFixedDays <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		result := t.provider.Service.Decimal.Divide(lt.Applied1, float64(lt.Terms))
		return t.provider.Service.Decimal.RoundToDecimalPlaces(result, 2), nil
	}
	return 0, eris.New("not implemented yet")
}

// SuggestedNumberOfTerms calculates the suggested number of terms for a loan based on payment amount and other factors
func (t *TransactionService) SuggestedNumberOfTerms(
	_ context.Context,
	suggestedAmount float64,
	principal float64,
	modeOfPayment core.LoanModeOfPayment,
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
	case core.LoanModeOfPaymentDaily:
		// daily = total / (payment * 30) using decimal arithmetic
		principalDivision := t.provider.Service.Decimal.Divide(principal, suggestedAmount)
		terms = t.provider.Service.Decimal.Divide(principalDivision, 30)
	case core.LoanModeOfPaymentWeekly:
		// weekly = total / (payment * 4) using decimal arithmetic
		principalDivision := t.provider.Service.Decimal.Divide(principal, suggestedAmount)
		terms = t.provider.Service.Decimal.Divide(principalDivision, 4)
	case core.LoanModeOfPaymentSemiMonthly:
		// semi-monthly = total / (payment * 2) using decimal arithmetic
		principalDivision := t.provider.Service.Decimal.Divide(principal, suggestedAmount)
		terms = t.provider.Service.Decimal.Divide(principalDivision, 2)
	case core.LoanModeOfPaymentMonthly:
		// monthly = total / payment using decimal arithmetic
		terms = t.provider.Service.Decimal.Divide(principal, suggestedAmount)
	case core.LoanModeOfPaymentQuarterly:
		// quarterly = total / (payment / 3) using decimal arithmetic
		principalDivision := t.provider.Service.Decimal.Divide(principal, suggestedAmount)
		terms = t.provider.Service.Decimal.Multiply(principalDivision, 3)
	case core.LoanModeOfPaymentSemiAnnual:
		// semi-annual = total / (payment / 6) using decimal arithmetic
		principalDivision := t.provider.Service.Decimal.Divide(principal, suggestedAmount)
		terms = t.provider.Service.Decimal.Multiply(principalDivision, 6)
	case core.LoanModeOfPaymentLumpsum:
		terms = 1
	case core.LoanModeOfPaymentFixedDays:
		if fixedDays <= 0 {
			return 0, errors.New("invalid fixed days: must be greater than 0")
		}
		terms = t.provider.Service.Decimal.Divide(principal, suggestedAmount)
	default:
		return 0, errors.New("unsupported mode of payment")
	}

	numberOfTerms := max(1, int(math.Ceil(terms)))
	return numberOfTerms, nil
}
