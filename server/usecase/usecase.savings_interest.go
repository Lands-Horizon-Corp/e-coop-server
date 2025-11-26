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
	Interest    float64
	InterestTax float64
}

func (t *TransactionService) SavingsInterestComputation(data SavingsInterestComputation) SavingsInterestComputationResult {

	result := SavingsInterestComputationResult{
		Interest:    0.0,
		InterestTax: 0.0,
	}

	// Check minimum period requirement
	daysInPeriod := len(data.DailyBalance)
	if daysInPeriod < 30 {
		return result
	}

	// Handle empty slice
	if len(data.DailyBalance) == 0 {
		return result
	}

	var balanceForCalculation float64

	// Determine which balance to use based on SavingsType
	switch data.SavingsType {
	case SavingsTypeLowest:
		// Find the lowest balance in the period
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if dailyBalance < lowestBalance {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance

	case SavingsTypeHighest:
		// Find the highest balance in the period
		highestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if dailyBalance > highestBalance {
				highestBalance = dailyBalance
			}
		}
		balanceForCalculation = highestBalance

	case SavingsTypeAverage:
		// Calculate average daily balance
		totalDailyBalance := 0.0
		for _, dailyBalance := range data.DailyBalance {
			totalDailyBalance += dailyBalance
		}
		balanceForCalculation = totalDailyBalance / float64(daysInPeriod)

	case SavingsTypeStart:
		// Use the first day's balance
		balanceForCalculation = data.DailyBalance[0]

	case SavingsTypeEnd:
		// Use the last day's balance
		balanceForCalculation = data.DailyBalance[len(data.DailyBalance)-1]

	default:
		// Default to lowest balance if SavingsType is not recognized
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if dailyBalance < lowestBalance {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance
	}

	// Skip if balance is 0 or negative
	if balanceForCalculation <= 0 {
		return result
	}

	// Calculate interest: Interest = Balance × Interest_Rate × (Days_in_Period ÷ Annual_Divisor)
	totalInterest := balanceForCalculation * data.InterestRate * (float64(daysInPeriod) / float64(data.AnnualDivisor))
	totalTax := totalInterest * data.InterestTaxRate

	result.Interest = totalInterest
	result.InterestTax = totalTax

	return result
}
