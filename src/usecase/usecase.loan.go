package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func LoanChargesRateComputation(
	crs types.ChargesRateScheme,
	ald types.LoanTransaction,
) float64 {

	var result decimal.Decimal
	applied := decimal.NewFromFloat(ald.Applied1)

	termHeaders := []int{
		crs.ByTermHeader1, crs.ByTermHeader2, crs.ByTermHeader3,
		crs.ByTermHeader4, crs.ByTermHeader5, crs.ByTermHeader6,
		crs.ByTermHeader7, crs.ByTermHeader8, crs.ByTermHeader9,
		crs.ByTermHeader10, crs.ByTermHeader11, crs.ByTermHeader12,
		crs.ByTermHeader13, crs.ByTermHeader14, crs.ByTermHeader15,
		crs.ByTermHeader16, crs.ByTermHeader17, crs.ByTermHeader18,
		crs.ByTermHeader19, crs.ByTermHeader20, crs.ByTermHeader21,
		crs.ByTermHeader22,
	}

	modeHeaders := []int{
		crs.ModeOfPaymentHeader1, crs.ModeOfPaymentHeader2, crs.ModeOfPaymentHeader3,
		crs.ModeOfPaymentHeader4, crs.ModeOfPaymentHeader5, crs.ModeOfPaymentHeader6,
		crs.ModeOfPaymentHeader7, crs.ModeOfPaymentHeader8, crs.ModeOfPaymentHeader9,
		crs.ModeOfPaymentHeader10, crs.ModeOfPaymentHeader11, crs.ModeOfPaymentHeader12,
		crs.ModeOfPaymentHeader13, crs.ModeOfPaymentHeader14, crs.ModeOfPaymentHeader15,
		crs.ModeOfPaymentHeader16, crs.ModeOfPaymentHeader17, crs.ModeOfPaymentHeader18,
		crs.ModeOfPaymentHeader19, crs.ModeOfPaymentHeader20, crs.ModeOfPaymentHeader21,
		crs.ModeOfPaymentHeader22,
	}

	findLastRate := func(rates []decimal.Decimal, headers []int, terms int) decimal.Decimal {
		last := decimal.Zero
		limit := min(len(rates), len(headers))
		for i := range limit {
			if headers[i] > terms || rates[i].LessThanOrEqual(decimal.Zero) {
				break
			}
			last = rates[i]
		}
		return last
	}

	computeCharge := func(rate decimal.Decimal) decimal.Decimal {
		if rate.LessThanOrEqual(decimal.Zero) {
			return decimal.Zero
		}

		base := applied.Mul(rate).Div(decimal.NewFromInt(100))

		switch ald.ModeOfPayment {
		case types.LoanModeOfPaymentDaily:
			return base.Div(decimal.NewFromInt(30))

		case types.LoanModeOfPaymentWeekly:
			return base.Mul(decimal.NewFromInt(7)).Div(decimal.NewFromInt(30))

		case types.LoanModeOfPaymentSemiMonthly:
			return base.Mul(decimal.NewFromInt(15)).Div(decimal.NewFromInt(30))

		case types.LoanModeOfPaymentMonthly:
			return base

		case types.LoanModeOfPaymentQuarterly:
			return base.Mul(decimal.NewFromInt(3))

		case types.LoanModeOfPaymentSemiAnnual:
			return base.Mul(decimal.NewFromInt(6))

		default:
			return decimal.Zero
		}
	}

	switch crs.Type {

	/* ---------------- BY RANGE ---------------- */

	case types.ChargesRateSchemeTypeByRange:
		for _, r := range crs.ChargesRateByRangeOrMinimumAmounts {
			if ald.Applied1 < r.From || ald.Applied1 > r.To {
				continue
			}

			var charge decimal.Decimal
			if r.Charge > 0 {
				charge = applied.Mul(decimal.NewFromFloat(r.Charge)).Div(decimal.NewFromInt(100))
			} else if r.Amount > 0 {
				charge = decimal.NewFromFloat(r.Amount)
			}

			if charge.GreaterThan(decimal.Zero) {
				min := decimal.NewFromFloat(r.MinimumAmount)
				if min.GreaterThan(decimal.Zero) && charge.GreaterThanOrEqual(min) {
					return min.InexactFloat64()
				}
				return charge.InexactFloat64()
			}
		}

	/* ---------------- BY TYPE ---------------- */

	case types.ChargesRateSchemeTypeByType:

		if crs.MemberType != nil &&
			ald.MemberProfile.MemberTypeID != &crs.MemberType.ID {
			return 0
		}

		if crs.ModeOfPayment != nil &&
			ald.ModeOfPayment != *crs.ModeOfPayment {
			return 0
		}

		for _, m := range crs.ChargesRateSchemeModeOfPayments {
			if ald.Applied1 < m.From || ald.Applied1 > m.To {
				continue
			}

			rates := []decimal.Decimal{
				decimal.NewFromFloat(m.Column1), decimal.NewFromFloat(m.Column2),
				decimal.NewFromFloat(m.Column3), decimal.NewFromFloat(m.Column4),
				decimal.NewFromFloat(m.Column5), decimal.NewFromFloat(m.Column6),
				decimal.NewFromFloat(m.Column7), decimal.NewFromFloat(m.Column8),
				decimal.NewFromFloat(m.Column9), decimal.NewFromFloat(m.Column10),
				decimal.NewFromFloat(m.Column11), decimal.NewFromFloat(m.Column12),
				decimal.NewFromFloat(m.Column13), decimal.NewFromFloat(m.Column14),
				decimal.NewFromFloat(m.Column15), decimal.NewFromFloat(m.Column16),
				decimal.NewFromFloat(m.Column17), decimal.NewFromFloat(m.Column18),
				decimal.NewFromFloat(m.Column19), decimal.NewFromFloat(m.Column20),
				decimal.NewFromFloat(m.Column21), decimal.NewFromFloat(m.Column22),
			}

			rate := findLastRate(rates, modeHeaders, ald.Terms)
			result = computeCharge(rate)
			if result.GreaterThan(decimal.Zero) {
				return result.InexactFloat64()
			}
		}

	/* ---------------- BY TERM ---------------- */

	case types.ChargesRateSchemeTypeByTerm:
		if ald.Terms < 1 {
			return 0
		}

		for _, t := range crs.ChargesRateByTerms {
			if t.ModeOfPayment != ald.ModeOfPayment {
				continue
			}

			rates := []decimal.Decimal{
				decimal.NewFromFloat(t.Rate1), decimal.NewFromFloat(t.Rate2),
				decimal.NewFromFloat(t.Rate3), decimal.NewFromFloat(t.Rate4),
				decimal.NewFromFloat(t.Rate5), decimal.NewFromFloat(t.Rate6),
				decimal.NewFromFloat(t.Rate7), decimal.NewFromFloat(t.Rate8),
				decimal.NewFromFloat(t.Rate9), decimal.NewFromFloat(t.Rate10),
				decimal.NewFromFloat(t.Rate11), decimal.NewFromFloat(t.Rate12),
				decimal.NewFromFloat(t.Rate13), decimal.NewFromFloat(t.Rate14),
				decimal.NewFromFloat(t.Rate15), decimal.NewFromFloat(t.Rate16),
				decimal.NewFromFloat(t.Rate17), decimal.NewFromFloat(t.Rate18),
				decimal.NewFromFloat(t.Rate19), decimal.NewFromFloat(t.Rate20),
				decimal.NewFromFloat(t.Rate21), decimal.NewFromFloat(t.Rate22),
			}

			rate := findLastRate(rates, termHeaders, ald.Terms)
			result = computeCharge(rate)
			if result.GreaterThan(decimal.Zero) {
				return result.InexactFloat64()
			}
		}
	}

	return 0
}

func LoanNumberOfPayments(mp types.LoanModeOfPayment, terms int) (int, error) {
	switch mp {
	case types.LoanModeOfPaymentDaily:
		return terms * 30, nil
	case types.LoanModeOfPaymentWeekly:
		return terms * 4, nil
	case types.LoanModeOfPaymentSemiMonthly:
		return terms * 2, nil
	case types.LoanModeOfPaymentMonthly:
		return terms, nil
	case types.LoanModeOfPaymentQuarterly:
		return terms / 3, nil
	case types.LoanModeOfPaymentSemiAnnual:
		return terms / 6, nil
	case types.LoanModeOfPaymentLumpsum:
		return 1, nil
	case types.LoanModeOfPaymentFixedDays:
		if terms <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		return terms, nil
	}
	return 0, eris.New("not implemented yet")
}

func LoanComputation(
	ald types.AutomaticLoanDeduction,
	lt types.LoanTransaction,
) float64 {

	result := decimal.NewFromFloat(lt.Applied1)
	if ald.MinAmount > 0 &&
		result.LessThan(decimal.NewFromFloat(ald.MinAmount)) {
		return 0
	}

	if ald.MaxAmount > 0 &&
		result.GreaterThan(decimal.NewFromFloat(ald.MaxAmount)) {
		return 0
	}
	if ald.ChargesPercentage1 > 0 || ald.ChargesPercentage2 > 0 {
		switch {
		case ald.ChargesPercentage1 > 0 && ald.ChargesPercentage2 > 0:
			if ald.AddOn {
				result = result.
					Mul(decimal.NewFromFloat(ald.ChargesPercentage2)).
					Div(decimal.NewFromInt(100))
			} else {
				result = result.
					Mul(decimal.NewFromFloat(ald.ChargesPercentage1)).
					Div(decimal.NewFromInt(100))
			}

		case ald.ChargesPercentage1 > 0:
			result = result.
				Mul(decimal.NewFromFloat(ald.ChargesPercentage1)).
				Div(decimal.NewFromInt(100))

		default:
			result = result.
				Mul(decimal.NewFromFloat(ald.ChargesPercentage2)).
				Div(decimal.NewFromInt(100))
		}
	}
	if ald.ChargesDivisor > 0 && result.GreaterThan(decimal.Zero) {
		result = result.
			Div(decimal.NewFromFloat(ald.ChargesDivisor)).
			Mul(decimal.NewFromFloat(ald.ChargesAmount))
	}
	switch {
	case ald.NumberOfMonths == 0 && ald.Anum == 1:
		result = result.Div(decimal.NewFromInt(12))

	case ald.NumberOfMonths == -1:
		result = result.
			Mul(decimal.NewFromInt(int64(lt.Terms))).
			Div(decimal.NewFromInt(12))

	case ald.NumberOfMonths > 0:
		result = result.
			Mul(decimal.NewFromInt(int64(lt.Terms))).
			Div(decimal.NewFromInt(int64(ald.NumberOfMonths)))
	}
	if result.Equal(decimal.NewFromFloat(lt.Applied1)) {
		return decimal.NewFromFloat(ald.ChargesAmount).InexactFloat64()
	}
	return result.Round(2).InexactFloat64()
}

func LoanModeOfPayment(lt *types.LoanTransaction) (float64, error) {
	applied := decimal.NewFromFloat(lt.Applied1)

	switch lt.ModeOfPayment {
	case types.LoanModeOfPaymentDaily:
		// Applied1 / Terms / 30
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		termsDiv := applied.Div(decimal.NewFromInt(int64(lt.Terms)))
		result := termsDiv.Div(decimal.NewFromInt(30))
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentWeekly:
		// Applied1 / Terms / 4
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		termsDiv := applied.Div(decimal.NewFromInt(int64(lt.Terms)))
		result := termsDiv.Div(decimal.NewFromInt(4))
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentSemiMonthly:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		termsDiv := applied.Div(decimal.NewFromInt(int64(lt.Terms)))
		result := termsDiv.Div(decimal.NewFromInt(2))
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentMonthly:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		result := applied.Div(decimal.NewFromInt(int64(lt.Terms)))
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentQuarterly:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		termsDiv := decimal.NewFromInt(int64(lt.Terms)).Div(decimal.NewFromInt(3))
		result := applied.Div(termsDiv)
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentSemiAnnual:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		termsDiv := decimal.NewFromInt(int64(lt.Terms)).Div(decimal.NewFromInt(6))
		result := applied.Div(termsDiv)
		return result.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentLumpsum:
		return applied.Round(2).InexactFloat64(), nil

	case types.LoanModeOfPaymentFixedDays:
		if lt.Terms <= 0 {
			return 0, eris.New("invalid terms: must be greater than 0")
		}
		if lt.ModeOfPaymentFixedDays <= 0 {
			return 0, eris.New("invalid fixed days: must be greater than 0")
		}
		result := applied.Div(decimal.NewFromInt(int64(lt.Terms)))
		return result.Round(2).InexactFloat64(), nil
	}

	return 0, eris.New("loan mode of payment not implemented yet")
}

func SuggestedNumberOfTerms(
	_ context.Context,
	suggestedAmount float64,
	principal float64,
	modeOfPayment types.LoanModeOfPayment,
	fixedDays int,
) (int, error) {

	if suggestedAmount <= 0 {
		return 0, errors.New("suggested amount must be greater than zero")
	}
	if principal <= 0 {
		return 0, errors.New("invalid total loan amount")
	}

	baseTerms := decimal.NewFromFloat(principal).Div(decimal.NewFromFloat(suggestedAmount))

	var terms decimal.Decimal

	switch modeOfPayment {
	case types.LoanModeOfPaymentDaily:
		terms = baseTerms.Div(decimal.NewFromInt(30))
	case types.LoanModeOfPaymentWeekly:
		terms = baseTerms.Div(decimal.NewFromInt(4))
	case types.LoanModeOfPaymentSemiMonthly:
		terms = baseTerms.Div(decimal.NewFromInt(2))
	case types.LoanModeOfPaymentMonthly:
		terms = baseTerms
	case types.LoanModeOfPaymentQuarterly:
		terms = baseTerms.Mul(decimal.NewFromInt(3))
	case types.LoanModeOfPaymentSemiAnnual:
		terms = baseTerms.Mul(decimal.NewFromInt(6))
	case types.LoanModeOfPaymentLumpsum:
		terms = decimal.NewFromInt(1)
	case types.LoanModeOfPaymentFixedDays:
		if fixedDays <= 0 {
			return 0, errors.New("invalid fixed days: must be greater than 0")
		}
		terms = baseTerms
	default:
		return 0, errors.New("unsupported mode of payment")
	}

	numberOfTerms := max(int(math.Ceil(terms.InexactFloat64())), 1)

	return numberOfTerms, nil
}

func ComputeFines(
	balance float64,
	finesAmortRate float64,
	finesMaturityRate float64,
	daysSkipped int,
	mode types.LoanModeOfPayment,
	noGracePeriodDaily bool,
	account types.Account,
) float64 {
	if daysSkipped <= 0 {
		return 0.0
	}

	// Determine the applicable fines rate
	finesRate := decimal.NewFromFloat(finesAmortRate)
	if daysSkipped > 30 {
		finesRate = decimal.NewFromFloat(finesMaturityRate)
	}
	if finesRate.Cmp(decimal.Zero) <= 0 {
		return 0.0
	}

	// Apply grace period if applicable
	if !noGracePeriodDaily {
		gracePercentage := decimal.Zero
		switch mode {
		case types.LoanModeOfPaymentDaily, types.LoanModeOfPaymentFixedDays:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntryDailyAmortization)
		case types.LoanModeOfPaymentWeekly:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntryWeeklyAmortization)
		case types.LoanModeOfPaymentMonthly:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntryMonthlyAmortization)
		case types.LoanModeOfPaymentSemiMonthly:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntrySemiMonthlyAmortization)
		case types.LoanModeOfPaymentQuarterly:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntryQuarterlyAmortization)
		case types.LoanModeOfPaymentSemiAnnual:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntrySemiAnnualAmortization)
		case types.LoanModeOfPaymentLumpsum:
			gracePercentage = decimal.NewFromFloat(account.CohCibFinesGracePeriodEntryLumpsumAmortization)
		}

		if gracePercentage.Cmp(decimal.NewFromFloat(100)) >= 0 {
			return 0.0
		}

		if gracePercentage.Cmp(decimal.Zero) > 0 {
			factor := decimal.NewFromFloat(1).Sub(gracePercentage.Div(decimal.NewFromFloat(100)))
			finesRate = finesRate.Mul(factor)
		}
	}

	balanceDec := decimal.NewFromFloat(balance)

	// Helper to multiply finesRate by number of periods
	calc := func(periods float64) float64 {
		result := balanceDec.Mul(finesRate.Div(decimal.NewFromFloat(100))).Mul(decimal.NewFromFloat(periods))
		rounded, _ := result.Round(2).Float64()
		return rounded
	}

	switch mode {
	case types.LoanModeOfPaymentDaily, types.LoanModeOfPaymentFixedDays:
		return calc(float64(daysSkipped))
	case types.LoanModeOfPaymentWeekly:
		return calc(float64(daysSkipped) / 7.0)
	case types.LoanModeOfPaymentSemiMonthly:
		return calc(float64(daysSkipped) / 15.0)
	case types.LoanModeOfPaymentMonthly:
		return calc(float64(daysSkipped) / 30.0)
	case types.LoanModeOfPaymentQuarterly:
		return calc(float64(daysSkipped) / 90.0)
	case types.LoanModeOfPaymentSemiAnnual:
		return calc(float64(daysSkipped) / 180.0)
	case types.LoanModeOfPaymentLumpsum:
		finalRate := decimal.NewFromFloat(finesMaturityRate)
		if finalRate.Cmp(decimal.Zero) <= 0 {
			finalRate = decimal.NewFromFloat(finesAmortRate)
		}
		result := balanceDec.Mul(finalRate.Div(decimal.NewFromFloat(100)))
		rounded, _ := result.Round(2).Float64()
		return rounded
	default:
		return 0.0
	}
}

func ComputeInterest(
	balance float64,
	rate float64,
	mode types.LoanModeOfPayment,
) float64 {
	b := decimal.NewFromFloat(balance)
	r := decimal.NewFromFloat(rate).Div(decimal.NewFromInt(100))
	calc := func(multiplier decimal.Decimal) float64 {
		return b.Mul(multiplier).Round(2).InexactFloat64()
	}
	switch mode {
	case types.LoanModeOfPaymentMonthly, types.LoanModeOfPaymentLumpsum:
		return calc(r)
	case types.LoanModeOfPaymentDaily, types.LoanModeOfPaymentFixedDays:
		dailyRate := r.Div(decimal.NewFromInt(30))
		return calc(dailyRate)
	case types.LoanModeOfPaymentSemiMonthly:
		dailyRate := r.Div(decimal.NewFromInt(30))
		return calc(dailyRate.Mul(decimal.NewFromInt(15)))
	case types.LoanModeOfPaymentWeekly:
		dailyRate := r.Div(decimal.NewFromInt(30))
		return calc(dailyRate.Mul(decimal.NewFromInt(7)))
	case types.LoanModeOfPaymentQuarterly:
		return calc(r.Mul(decimal.NewFromInt(3)))

	case types.LoanModeOfPaymentSemiAnnual:
		return calc(r.Mul(decimal.NewFromInt(6)))
	default:
		return 0.0
	}
}

func ComputeInterestStraight(balance float64, rate float64, terms int) float64 {
	if rate <= 0 || balance <= 0 {
		return 0
	}
	balanceDec := decimal.NewFromFloat(balance)
	rateDec := decimal.NewFromFloat(rate)
	straightInterest := balanceDec.Mul(rateDec).Div(decimal.NewFromInt(100))
	if terms > 1 {
		straightInterest = straightInterest.Mul(decimal.NewFromInt(int64(terms)))
	}
	return straightInterest.Round(2).InexactFloat64()
}
