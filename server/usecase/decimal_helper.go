package usecase

import (
	"github.com/shopspring/decimal"
)

// DecimalOperations provides utility functions for precise decimal arithmetic operations
// This helper ensures financial calculations are performed with high precision to avoid
// floating-point errors common in monetary computations.
type DecimalOperations struct{}

// NewDecimalHelper creates a new instance of DecimalOperations
func NewDecimalHelper() *DecimalOperations {
	return &DecimalOperations{}
}

// NewDecimal converts float64 to decimal.Decimal for precise calculations
func (d *DecimalOperations) NewDecimal(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

// NewDecimalFromString creates a decimal from string representation
func (d *DecimalOperations) NewDecimalFromString(value string) (decimal.Decimal, error) {
	return decimal.NewFromString(value)
}

// Add performs precise decimal addition (a + b)
func (d *DecimalOperations) Add(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Add(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// AddMultiple adds multiple values with precision
func (d *DecimalOperations) AddMultiple(values ...float64) float64 {
	result := decimal.NewFromFloat(0)
	for _, value := range values {
		result = result.Add(decimal.NewFromFloat(value))
	}
	resultFloat, _ := result.Float64()
	return resultFloat
}

// Subtract performs precise decimal subtraction (a - b)
func (d *DecimalOperations) Subtract(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Sub(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// Multiply performs precise decimal multiplication (a * b)
func (d *DecimalOperations) Multiply(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Mul(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// MultiplyMultiple multiplies multiple values with precision
func (d *DecimalOperations) MultiplyMultiple(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := decimal.NewFromFloat(values[0])
	for i := 1; i < len(values); i++ {
		result = result.Mul(decimal.NewFromFloat(values[i]))
	}
	resultFloat, _ := result.Float64()
	return resultFloat
}

// Divide performs precise decimal division (a / b)
func (d *DecimalOperations) Divide(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Div(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// DivideWithPrecision performs precise decimal division with specified decimal places
func (d *DecimalOperations) DivideWithPrecision(a, b float64, precision int32) float64 {
	if b == 0 {
		return 0
	}
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.DivRound(decB, precision)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// MultiplyByPercentage calculates percentage of a value using precise decimal arithmetic
// Example: MultiplyByPercentage(1000, 15.5) = 1000 * 15.5% = 155
func (d *DecimalOperations) MultiplyByPercentage(amount, percentage float64) float64 {
	decAmount := decimal.NewFromFloat(amount)
	decPercentage := decimal.NewFromFloat(percentage).Div(decimal.NewFromInt(100))
	result := decAmount.Mul(decPercentage)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// CalculatePercentage calculates what percentage one value is of another
// Example: CalculatePercentage(250, 1000) = 25%
func (d *DecimalOperations) CalculatePercentage(part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	decPart := decimal.NewFromFloat(part)
	decWhole := decimal.NewFromFloat(whole)
	decHundred := decimal.NewFromInt(100)
	result := decPart.Div(decWhole).Mul(decHundred)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// AddPercentage adds a percentage to a base amount
// Example: AddPercentage(1000, 10) = 1000 + (1000 * 10%) = 1100
func (d *DecimalOperations) AddPercentage(baseAmount, percentage float64) float64 {
	percentageAmount := d.MultiplyByPercentage(baseAmount, percentage)
	return d.Add(baseAmount, percentageAmount)
}

// SubtractPercentage subtracts a percentage from a base amount
// Example: SubtractPercentage(1000, 10) = 1000 - (1000 * 10%) = 900
func (d *DecimalOperations) SubtractPercentage(baseAmount, percentage float64) float64 {
	percentageAmount := d.MultiplyByPercentage(baseAmount, percentage)
	return d.Subtract(baseAmount, percentageAmount)
}

// RoundToDecimalPlaces rounds a float64 to specified decimal places
func (d *DecimalOperations) RoundToDecimalPlaces(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.Round(places)
	result, _ := rounded.Float64()
	return result
}

// RoundUp rounds up to specified decimal places
func (d *DecimalOperations) RoundUp(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundUp(places)
	result, _ := rounded.Float64()
	return result
}

// RoundDown rounds down to specified decimal places
func (d *DecimalOperations) RoundDown(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundDown(places)
	result, _ := rounded.Float64()
	return result
}

// RoundBank performs banker's rounding to specified decimal places
func (d *DecimalOperations) RoundBank(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundBank(places)
	result, _ := rounded.Float64()
	return result
}

// IsEqual compares two float64 values for equality with precision tolerance
func (d *DecimalOperations) IsEqual(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.Equal(decB)
}

// IsGreaterThan checks if a > b with decimal precision
func (d *DecimalOperations) IsGreaterThan(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.GreaterThan(decB)
}

// IsLessThan checks if a < b with decimal precision
func (d *DecimalOperations) IsLessThan(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.LessThan(decB)
}

// Abs returns the absolute value
func (d *DecimalOperations) Abs(value float64) float64 {
	dec := decimal.NewFromFloat(value)
	result := dec.Abs()
	resultFloat, _ := result.Float64()
	return resultFloat
}

// Min returns the minimum of two values
func (d *DecimalOperations) Min(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	if decA.LessThan(decB) {
		return a
	}
	return b
}

// Max returns the maximum of two values
func (d *DecimalOperations) Max(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	if decA.GreaterThan(decB) {
		return a
	}
	return b
}

// CompoundInterest calculates compound interest: P(1 + r)^t
func (d *DecimalOperations) CompoundInterest(principal, rate float64, periods int) float64 {
	decPrincipal := decimal.NewFromFloat(principal)
	decRate := decimal.NewFromFloat(rate).Div(decimal.NewFromInt(100)) // Convert percentage to decimal
	decOne := decimal.NewFromInt(1)

	// Calculate (1 + rate)
	base := decOne.Add(decRate)

	// Calculate (1 + rate)^periods
	result := decPrincipal
	for range periods {
		result = result.Mul(base)
	}

	resultFloat, _ := result.Float64()
	return resultFloat
}

// SimpleInterest calculates simple interest: P * r * t
func (d *DecimalOperations) SimpleInterest(principal, rate, time float64) float64 {
	decPrincipal := decimal.NewFromFloat(principal)
	decRate := decimal.NewFromFloat(rate).Div(decimal.NewFromInt(100)) // Convert percentage to decimal
	decTime := decimal.NewFromFloat(time)

	result := decPrincipal.Mul(decRate).Mul(decTime)
	resultFloat, _ := result.Float64()
	return resultFloat
}

// Global instance for backward compatibility
var GlobalDecimalHelper = NewDecimalHelper()
