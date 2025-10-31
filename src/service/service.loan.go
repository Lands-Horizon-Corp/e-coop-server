package service

import (
	"context"
	"errors"
	"math"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/rotisserie/eris"
)

func (t *TransactionService) LoanChargesRateComputation(ctx context.Context, crs model_core.ChargesRateScheme, ald model_core.LoanTransaction) float64 {

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
		for i := 0; i < minLen; i++ {
			rate := rates[i]
			term := headers[i]
			if term > terms || rate <= 0 {
				break
			}
			lastRate = rate
		}
		return lastRate
	}

	computeCharge := func(applied, rate float64, mode model_core.LoanModeOfPayment) float64 {
		if rate <= 0 {
			return 0.0
		}
		base := applied * rate / 100.0
		switch mode {
		case model_core.LoanModeOfPaymentDaily:
			return base / 30.0
		case model_core.LoanModeOfPaymentWeekly:
			return base * 7.0 / 30.0
		case model_core.LoanModeOfPaymentSemiMonthly:
			return base * 15.0 / 30.0
		case model_core.LoanModeOfPaymentMonthly:
			return base
		case model_core.LoanModeOfPaymentQuarterly:
			return base * 3.0
		case model_core.LoanModeOfPaymentSemiAnnual:
			return base * 6.0
		default:
			return 0.0
		}
	}

	switch crs.Type {
	case model_core.ChargesRateSchemeTypeByRange:
		for _, data := range crs.ChargesRateByRangeOrMinimumAmounts {
			if ald.Applied1 < data.From || ald.Applied1 > data.To {
				continue
			}
			charge := 0.0
			if data.Charge > 0 {
				charge = ald.Applied1 * (data.Charge / 100.0)

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
	case model_core.ChargesRateSchemeTypeByType:

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
	case model_core.ChargesRateSchemeTypeByTerm:
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

func (t *TransactionService) LoanNumberOfPayments(ctx context.Context, lt *model_core.LoanTransaction) (int, error) {
	switch lt.ModeOfPayment {
	case model_core.LoanModeOfPaymentDaily:
		return lt.Terms * 30, nil
	case model_core.LoanModeOfPaymentWeekly:
		return lt.Terms * 4, nil
	case model_core.LoanModeOfPaymentSemiMonthly:
		return lt.Terms * 2, nil
	case model_core.LoanModeOfPaymentMonthly:
		return lt.Terms, nil
	case model_core.LoanModeOfPaymentQuarterly:
		return lt.Terms / 3, nil
	case model_core.LoanModeOfPaymentSemiAnnual:
		return lt.Terms / 6, nil
	case model_core.LoanModeOfPaymentLumpsum:
		return 1, nil
	case model_core.LoanModeOfPaymentFixedDays:
		if lt.ModeOfPaymentFixedDays <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		return lt.Terms, nil
	}
	return 0, eris.New("not implemented yet")
}

func (t *TransactionService) LoanComputation(ctx context.Context, ald model_core.AutomaticLoanDeduction, lt model_core.LoanTransaction) float64 {
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

func (t *TransactionService) LoanModeOfPayment(ctx context.Context, lt *model_core.LoanTransaction) (float64, error) {
	switch lt.ModeOfPayment {
	case model_core.LoanModeOfPaymentDaily:
		return lt.Applied1 / float64(lt.Terms) / 30, nil
	case model_core.LoanModeOfPaymentWeekly:
		return lt.Applied1 / float64(lt.Terms) / 4, nil
	case model_core.LoanModeOfPaymentSemiMonthly:
		return lt.Applied1 / float64(lt.Terms) / 2, nil
	case model_core.LoanModeOfPaymentMonthly:
		return lt.Applied1 / float64(lt.Terms), nil
	case model_core.LoanModeOfPaymentQuarterly:
		return lt.Applied1 / (float64(lt.Terms) / 3), nil
	case model_core.LoanModeOfPaymentSemiAnnual:
		return lt.Applied1 / (float64(lt.Terms) / 6), nil
	case model_core.LoanModeOfPaymentLumpsum:
		return lt.Applied1, nil
	case model_core.LoanModeOfPaymentFixedDays:
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
	modeOfPayment model_core.LoanModeOfPayment,
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
	case model_core.LoanModeOfPaymentDaily:
		// daily = total / (payment * 30)
		terms = (principal / suggestedAmount) / 30
	case model_core.LoanModeOfPaymentWeekly:
		// weekly = total / (payment * 4)
		terms = (principal / suggestedAmount) / 4
	case model_core.LoanModeOfPaymentSemiMonthly:
		// semi-monthly = total / (payment * 2)
		terms = (principal / suggestedAmount) / 2
	case model_core.LoanModeOfPaymentMonthly:
		// monthly = total / payment
		terms = principal / suggestedAmount
	case model_core.LoanModeOfPaymentQuarterly:
		// quarterly = total / (payment / 3)
		terms = (principal / suggestedAmount) * 3
	case model_core.LoanModeOfPaymentSemiAnnual:
		// semi-annual = total / (payment / 6)
		terms = (principal / suggestedAmount) * 6
	case model_core.LoanModeOfPaymentLumpsum:
		terms = 1
	case model_core.LoanModeOfPaymentFixedDays:
		if fixedDays <= 0 {
			return 0, errors.New("invalid fixed days: must be greater than 0")
		}
		terms = principal / suggestedAmount
	default:
		return 0, errors.New("unsupported mode of payment")
	}

	numberOfTerms := max(1, int(math.Ceil(terms)))
	return numberOfTerms, nil
}
