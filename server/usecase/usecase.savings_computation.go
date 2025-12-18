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

	if len(data.DailyBalance) == 0 {
		return result
	}

	var balanceForCalculation float64

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
		balanceForCalculation = t.provider.Service.Decimal.AddMultiple(data.DailyBalance...) / float64(len(data.DailyBalance))

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

	finalBalance := t.provider.Service.Decimal.Add(balanceForCalculation, data.InterestAmount)

	result.Balance = finalBalance
	return result
}
