package usecase

import "github.com/shopspring/decimal"

type SavingsType string

const (
	SavingsTypeAverage SavingsType = "average"
	SavingsTypeLowest  SavingsType = "lowest"
	SavingsTypeHighest SavingsType = "highest"
	SavingsTypeEnd     SavingsType = "end"
	SavingsTypeStart   SavingsType = "start"
)

type SavingsInterest struct {
	DailyBalance    []float64
	InterestRate    float64
	InterestTaxRate float64
	SavingsType     SavingsType
	AnnualDivisor   int
}

type SavingsInterestComputationResult struct {
	Interest      float64
	InterestTax   float64
	EndingBalance float64
}

func SavingsInterestComputation(
	data SavingsInterest,
) SavingsInterestComputationResult {

	result := SavingsInterestComputationResult{
		Interest:      0.0,
		InterestTax:   0.0,
		EndingBalance: 0.0,
	}

	daysInPeriod := len(data.DailyBalance)
	if daysInPeriod < 30 || daysInPeriod == 0 {
		return result
	}

	interestRate := decimal.NewFromFloat(data.InterestRate)
	interestTaxRate := decimal.NewFromFloat(data.InterestTaxRate)

	// Convert rates from percentage if > 1
	if interestRate.GreaterThan(decimal.NewFromInt(1)) {
		interestRate = interestRate.Div(decimal.NewFromInt(100))
	}
	if interestTaxRate.GreaterThan(decimal.NewFromInt(1)) {
		interestTaxRate = interestTaxRate.Div(decimal.NewFromInt(100))
	}

	// Safety check
	if interestRate.LessThanOrEqual(decimal.Zero) || interestTaxRate.LessThan(decimal.Zero) || interestTaxRate.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return result
	}

	var balanceForCalculation decimal.Decimal
	actualEndingBalance := decimal.NewFromFloat(data.DailyBalance[daysInPeriod-1])

	// Compute balance based on savings type
	switch data.SavingsType {
	case SavingsTypeLowest:
		lowest := decimal.NewFromFloat(data.DailyBalance[0])
		for _, v := range data.DailyBalance {
			if decimal.NewFromFloat(v).LessThan(lowest) {
				lowest = decimal.NewFromFloat(v)
			}
		}
		balanceForCalculation = lowest

	case SavingsTypeHighest:
		highest := decimal.NewFromFloat(data.DailyBalance[0])
		for _, v := range data.DailyBalance {
			if decimal.NewFromFloat(v).GreaterThan(highest) {
				highest = decimal.NewFromFloat(v)
			}
		}
		balanceForCalculation = highest

	case SavingsTypeAverage:
		sum := decimal.Zero
		for _, v := range data.DailyBalance {
			sum = sum.Add(decimal.NewFromFloat(v))
		}
		balanceForCalculation = sum.Div(decimal.NewFromInt(int64(daysInPeriod)))

	case SavingsTypeStart:
		balanceForCalculation = decimal.NewFromFloat(data.DailyBalance[0])

	case SavingsTypeEnd:
		balanceForCalculation = actualEndingBalance

	default:
		lowest := decimal.NewFromFloat(data.DailyBalance[0])
		for _, v := range data.DailyBalance {
			if decimal.NewFromFloat(v).LessThan(lowest) {
				lowest = decimal.NewFromFloat(v)
			}
		}
		balanceForCalculation = lowest
	}

	if balanceForCalculation.LessThanOrEqual(decimal.Zero) {
		return result
	}

	// Compute days period ratio
	daysPeriodRatio := decimal.NewFromInt(int64(daysInPeriod)).Div(decimal.NewFromInt(int64(data.AnnualDivisor)))

	// Compute gross interest
	grossInterest := balanceForCalculation.Mul(interestRate).Mul(daysPeriodRatio)

	if grossInterest.LessThanOrEqual(decimal.Zero) {
		return result
	}

	// Compute tax and net interest
	totalTax := grossInterest.Mul(interestTaxRate)
	totalInterest := grossInterest.Sub(totalTax)

	result.Interest = totalInterest.InexactFloat64()
	result.InterestTax = totalTax.InexactFloat64()
	result.EndingBalance = actualEndingBalance.Add(totalInterest).InexactFloat64()

	return result
}
