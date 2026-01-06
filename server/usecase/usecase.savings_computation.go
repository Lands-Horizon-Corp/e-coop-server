package usecase

import "github.com/shopspring/decimal"

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

func GetSavingsEndingBalance(data SavingsBalanceComputation) SavingsBalanceResult {
	result := SavingsBalanceResult{
		Balance:        0.0,
		InterestAmount: data.InterestAmount,
		InterestTax:    data.InterestTax,
	}

	if len(data.DailyBalance) == 0 {
		return result
	}

	// Convert daily balances to decimal
	dailyBalances := make([]decimal.Decimal, len(data.DailyBalance))
	for i, b := range data.DailyBalance {
		dailyBalances[i] = decimal.NewFromFloat(b)
	}

	var balanceForCalculation decimal.Decimal

	switch data.SavingsType {
	case SavingsTypeLowest:
		balanceForCalculation = dailyBalances[0]
		for _, b := range dailyBalances {
			if b.LessThan(balanceForCalculation) {
				balanceForCalculation = b
			}
		}

	case SavingsTypeHighest:
		balanceForCalculation = dailyBalances[0]
		for _, b := range dailyBalances {
			if b.GreaterThan(balanceForCalculation) {
				balanceForCalculation = b
			}
		}

	case SavingsTypeAverage:
		sum := decimal.Zero
		for _, b := range dailyBalances {
			sum = sum.Add(b)
		}
		balanceForCalculation = sum.Div(decimal.NewFromInt(int64(len(dailyBalances))))

	case SavingsTypeStart:
		balanceForCalculation = dailyBalances[0]

	case SavingsTypeEnd:
		balanceForCalculation = dailyBalances[len(dailyBalances)-1]

	default:
		// Default to lowest balance
		balanceForCalculation = dailyBalances[0]
		for _, b := range dailyBalances {
			if b.LessThan(balanceForCalculation) {
				balanceForCalculation = b
			}
		}
	}

	// Add interest amount
	finalBalance := balanceForCalculation.Add(decimal.NewFromFloat(data.InterestAmount))
	result.Balance = finalBalance.InexactFloat64()

	return result
}
