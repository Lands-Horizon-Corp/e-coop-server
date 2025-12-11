package usecase

type SavingsBalanceComputation struct {
	DailyBalance   []float64   `json:"daily_balance"`
	SavingsType    SavingsType `json:"savings_type"`
	InterestAmount float64     `json:"interest_amount"`
	InterestTax    float64     `json:"interest_tax"`
}

type SavingsBalanceResult struct {
	Balance        float64 `json:"balance"`
	InterestAmount float64 `json:"interest_amount"`
	InterestTax    float64 `json:"interest_tax"`
}

func (t *UsecaseService) GetSavingsEndingBalance(data SavingsBalanceComputation) SavingsBalanceResult {
	result := SavingsBalanceResult{
		Balance:        0.0,
		InterestAmount: data.InterestAmount,
		InterestTax:    data.InterestTax,
	}

	// Handle empty slice
	if len(data.DailyBalance) == 0 {
		return result
	}

	var balanceForCalculation float64

	// Determine which balance to use based on SavingsType
	switch data.SavingsType {
	case SavingsTypeLowest:
		// Find the lowest balance in the period using precise decimal comparison
		lowestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if t.provider.Service.Decimal.IsLessThan(dailyBalance, lowestBalance) {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance

	case SavingsTypeHighest:
		// Find the highest balance in the period using precise decimal comparison
		highestBalance := data.DailyBalance[0]
		for _, dailyBalance := range data.DailyBalance {
			if t.provider.Service.Decimal.IsGreaterThan(dailyBalance, highestBalance) {
				highestBalance = dailyBalance
			}
		}
		balanceForCalculation = highestBalance

	case SavingsTypeAverage:
		// Calculate average daily balance using precise decimal arithmetic
		balanceForCalculation = t.provider.Service.Decimal.AddMultiple(data.DailyBalance...) / float64(len(data.DailyBalance))

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
			if t.provider.Service.Decimal.IsLessThan(dailyBalance, lowestBalance) {
				lowestBalance = dailyBalance
			}
		}
		balanceForCalculation = lowestBalance
	}

	// Add interest amount to balance (net interest after tax has already been deducted)
	finalBalance := t.provider.Service.Decimal.Add(balanceForCalculation, data.InterestAmount)

	result.Balance = finalBalance
	return result
}
