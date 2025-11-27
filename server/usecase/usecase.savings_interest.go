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

func (t *TransactionService) SavingsInterestComputation(data SavingsInterestComputation) SavingsInterestComputationResult {

	result := SavingsInterestComputationResult{
		Interest:      0.0,
		InterestTax:   0.0,
		EndingBalance: 0.0,
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
		balanceForCalculation = t.provider.Service.Decimal.AddMultiple(data.DailyBalance...) / float64(daysInPeriod)

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

	// Skip if balance is 0 or negative using precise decimal comparison
	if t.provider.Service.Decimal.IsLessThan(balanceForCalculation, 0) || t.provider.Service.Decimal.IsEqual(balanceForCalculation, 0) {
		return result
	}

	// Calculate interest using precise decimal arithmetic: Interest = Balance × Interest_Rate × (Days_in_Period ÷ Annual_Divisor)
	daysPeriodRatio := t.provider.Service.Decimal.Divide(float64(daysInPeriod), float64(data.AnnualDivisor))
	totalInterest := t.provider.Service.Decimal.MultiplyMultiple(balanceForCalculation, data.InterestRate, daysPeriodRatio)
	totalTax := t.provider.Service.Decimal.Multiply(totalInterest, data.InterestTaxRate)

	result.Interest = totalInterest
	result.InterestTax = totalTax
	result.EndingBalance = balanceForCalculation
	return result
}

type SavingsBalanceComputation struct {
	DailyBalance   []float64
	SavingsType    SavingsType
	InterestAmount float64
	InterestTax    float64
}

type SavingsBalanceResult struct {
	Balance        float64
	InterestAmount float64
	InterestTax    float64
}

func (t *TransactionService) GetSavingsEndingBalance(data SavingsBalanceComputation) SavingsBalanceResult {
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

	// Subtract interest amount and interest tax from balance
	balanceAfterInterest := t.provider.Service.Decimal.Subtract(balanceForCalculation, data.InterestAmount)
	finalBalance := t.provider.Service.Decimal.Subtract(balanceAfterInterest, data.InterestTax)

	result.Balance = finalBalance
	return result
}
