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

	endingBalance := decimal.NewFromFloat(
		data.DailyBalance[len(data.DailyBalance)-1],
	)

	interest := decimal.NewFromFloat(data.InterestAmount)
	tax := decimal.NewFromFloat(data.InterestTax)

	netInterest := interest.Sub(tax)

	result.Balance = endingBalance.
		Add(netInterest).
		InexactFloat64()

	return result
}
