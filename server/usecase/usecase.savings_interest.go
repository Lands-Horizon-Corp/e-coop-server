package usecase

import "github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"

type SavingsInterestComputation struct {
	DailyBalance    []float64
	InterestRate    float64
	InterestTaxRate float64
	ComputationType core.SavingsComputationType
	AnnualDivisor   int
}

type SavingsInterestComputationResult struct {
	Interest    float64
	InterestTax float64
}

func (t *TransactionService) SavingsInterestComputation(data SavingsInterestComputation) SavingsInterestComputationResult {

	result := SavingsInterestComputationResult{
		Interest:    0.0,
		InterestTax: 0.0,
	}
	switch data.ComputationType {
	case core.SavingsComputationTypeDailyLowestBalance:
		daysInPeriod := len(data.DailyBalance)
		if daysInPeriod < 30 {
			return result
		}
		if len(data.DailyBalance) == 0 {
			return result
		}
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if dailyBalance < lowestBalance {
				lowestBalance = dailyBalance
			}
		}
		if lowestBalance <= 0 {
			return result
		}
		totalInterest := lowestBalance * data.InterestRate * (float64(daysInPeriod) / float64(data.AnnualDivisor))
		totalTax := totalInterest * data.InterestTaxRate
		result.Interest = totalInterest
		result.InterestTax = totalTax

	case core.SavingsComputationTypeAverageDailyBalance:
		daysInPeriod := len(data.DailyBalance)
		if daysInPeriod < 30 {
			return result
		}
		if len(data.DailyBalance) == 0 {
			return result
		}
		totalDailyBalance := 0.0
		for _, dailyBalance := range data.DailyBalance {
			totalDailyBalance += dailyBalance
		}
		averageDailyBalance := totalDailyBalance / float64(daysInPeriod)

		if averageDailyBalance <= 0 {
			return result
		}
		totalInterest := averageDailyBalance * data.InterestRate * (float64(daysInPeriod) / float64(data.AnnualDivisor))
		totalTax := totalInterest * data.InterestTaxRate

		result.Interest = totalInterest
		result.InterestTax = totalTax

	case core.SavingsComputationTypeMonthlyEndLowestBalance:
	case core.SavingsComputationTypeADBEndBalance:
	case core.SavingsComputationTypeMonthlyLowestBalanceAverage:
	case core.SavingsComputationTypeMonthlyEndBalanceAverage:
	case core.SavingsComputationTypeMonthlyEndBalanceTotal:
	}

	return result
}
