package usecase

type SavingsType string

const (
	SavingsTypeAverage SavingsType = "average"
	SavingsTypeLowest  SavingsType = "lowest"
	SavingsTypeHighest SavingsType = "highest"
	SavingsTypeEnd     SavingsType = "end"
	SavingsTypeStart   SavingsType = "start"
)

type SavingsInterestComputation struct {
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

func (t *UsecaseService) SavingsInterestComputation(
	data SavingsInterestComputation,
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

	interestRate := data.InterestRate
	interestTaxRate := data.InterestTaxRate

	if interestRate > 1 {
		interestRate = t.provider.Service.Decimal.Divide(interestRate, 100)
	}

	if interestTaxRate > 1 {
		interestTaxRate = t.provider.Service.Decimal.Divide(interestTaxRate, 100)
	}

	// Extra safety
	if interestRate <= 0 || interestTaxRate < 0 || interestTaxRate >= 1 {
		return result
	}

	var balanceForCalculation float64
	actualEndingBalance := data.DailyBalance[len(data.DailyBalance)-1]

	switch data.SavingsType {
	case SavingsTypeLowest:
		lowest := data.DailyBalance[0]
		for _, v := range data.DailyBalance {
			if t.provider.Service.Decimal.IsLessThan(v, lowest) {
				lowest = v
			}
		}
		balanceForCalculation = lowest

	case SavingsTypeHighest:
		highest := data.DailyBalance[0]
		for _, v := range data.DailyBalance {
			if t.provider.Service.Decimal.IsGreaterThan(v, highest) {
				highest = v
			}
		}
		balanceForCalculation = highest

	case SavingsTypeAverage:
		balanceForCalculation =
			t.provider.Service.Decimal.AddMultiple(data.DailyBalance...) /
				float64(daysInPeriod)

	case SavingsTypeStart:
		balanceForCalculation = data.DailyBalance[0]

	case SavingsTypeEnd:
		balanceForCalculation = data.DailyBalance[daysInPeriod-1]

	default:
		lowest := data.DailyBalance[0]
		for _, v := range data.DailyBalance {
			if t.provider.Service.Decimal.IsLessThan(v, lowest) {
				lowest = v
			}
		}
		balanceForCalculation = lowest
	}

	if balanceForCalculation <= 0 {
		return result
	}

	daysPeriodRatio :=
		t.provider.Service.Decimal.Divide(
			float64(daysInPeriod),
			float64(data.AnnualDivisor),
		)

	grossInterest :=
		t.provider.Service.Decimal.MultiplyMultiple(
			balanceForCalculation,
			interestRate,
			daysPeriodRatio,
		)

	if grossInterest <= 0 {
		return result
	}

	totalTax :=
		t.provider.Service.Decimal.Multiply(grossInterest, interestTaxRate)

	totalInterest :=
		t.provider.Service.Decimal.Subtract(grossInterest, totalTax)

	result.Interest = totalInterest
	result.InterestTax = totalTax
	result.EndingBalance =
		t.provider.Service.Decimal.Add(actualEndingBalance, totalInterest)

	return result
}
