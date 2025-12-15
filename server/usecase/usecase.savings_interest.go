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

func (t *UsecaseService) SavingsInterestComputation(data SavingsInterestComputation) SavingsInterestComputationResult {

	result := SavingsInterestComputationResult{
		Interest:      0.0,
		InterestTax:   0.0,
		EndingBalance: 0.0,
	}

	daysInPeriod := len(data.DailyBalance)
	if daysInPeriod < 30 {
		return result
	}

	if len(data.DailyBalance) == 0 {
		return result
	}

	var balanceForCalculation float64
	actualEndingBalance := data.DailyBalance[len(data.DailyBalance)-1]

	switch data.SavingsType {
	case SavingsTypeLowest:
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if t.provider.Service.Decimal.IsLessThan(dailyBalance, lowestBalance) {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance

	case SavingsTypeHighest:
		highestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if t.provider.Service.Decimal.IsGreaterThan(dailyBalance, highestBalance) {
				highestBalance = dailyBalance
			}
		}
		balanceForCalculation = highestBalance

	case SavingsTypeAverage:
		balanceForCalculation = t.provider.Service.Decimal.AddMultiple(data.DailyBalance...) / float64(daysInPeriod)

	case SavingsTypeStart:
		balanceForCalculation = data.DailyBalance[0]

	case SavingsTypeEnd:
		balanceForCalculation = data.DailyBalance[len(data.DailyBalance)-1]

	default:
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if t.provider.Service.Decimal.IsLessThan(dailyBalance, lowestBalance) {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance
	}

	if t.provider.Service.Decimal.IsLessThan(balanceForCalculation, 0) || t.provider.Service.Decimal.IsEqual(balanceForCalculation, 0) {
		return result
	}

	daysPeriodRatio := t.provider.Service.Decimal.Divide(float64(daysInPeriod), float64(data.AnnualDivisor))
	grossInterest := t.provider.Service.Decimal.MultiplyMultiple(balanceForCalculation, data.InterestRate, daysPeriodRatio)
	totalTax := t.provider.Service.Decimal.Multiply(grossInterest, data.InterestTaxRate)
	totalInterest := t.provider.Service.Decimal.Subtract(grossInterest, totalTax)

	result.Interest = totalInterest
	result.InterestTax = totalTax
	result.EndingBalance = t.provider.Service.Decimal.Add(actualEndingBalance, totalInterest)
	return result
}
