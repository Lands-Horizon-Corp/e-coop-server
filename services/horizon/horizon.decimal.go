package horizon

import (
	"github.com/shopspring/decimal"
)

type DecimalOperations struct{}

func NewDecimalHelper() *DecimalOperations {
	return &DecimalOperations{}
}

func (d *DecimalOperations) NewFromFloat(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

func (d *DecimalOperations) NewDecimal(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

func (d *DecimalOperations) NewDecimalFromString(value string) (decimal.Decimal, error) {
	return decimal.NewFromString(value)
}

func (d *DecimalOperations) Add(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Add(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) AddMultiple(values ...float64) float64 {
	result := decimal.NewFromFloat(0)
	for _, value := range values {
		result = result.Add(decimal.NewFromFloat(value))
	}
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) Subtract(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Sub(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) Multiply(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	result := decA.Mul(decB)
	resultFloat, _ := result.Float64()
	return resultFloat
}

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

func (d *DecimalOperations) MultiplyByPercentage(amount, percentage float64) float64 {
	decAmount := decimal.NewFromFloat(amount)
	decPercentage := decimal.NewFromFloat(percentage).Div(decimal.NewFromInt(100))
	result := decAmount.Mul(decPercentage)
	resultFloat, _ := result.Float64()
	return resultFloat
}

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

func (d *DecimalOperations) AddPercentage(baseAmount, percentage float64) float64 {
	percentageAmount := d.MultiplyByPercentage(baseAmount, percentage)
	return d.Add(baseAmount, percentageAmount)
}

func (d *DecimalOperations) SubtractPercentage(baseAmount, percentage float64) float64 {
	percentageAmount := d.MultiplyByPercentage(baseAmount, percentage)
	return d.Subtract(baseAmount, percentageAmount)
}

func (d *DecimalOperations) RoundToDecimalPlaces(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.Round(places)
	result, _ := rounded.Float64()
	return result
}

func (d *DecimalOperations) RoundUp(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundUp(places)
	result, _ := rounded.Float64()
	return result
}

func (d *DecimalOperations) RoundDown(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundDown(places)
	result, _ := rounded.Float64()
	return result
}

func (d *DecimalOperations) RoundBank(value float64, places int32) float64 {
	dec := decimal.NewFromFloat(value)
	rounded := dec.RoundBank(places)
	result, _ := rounded.Float64()
	return result
}

func (d *DecimalOperations) IsEqual(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.Equal(decB)
}

func (d *DecimalOperations) IsGreaterThan(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.GreaterThan(decB)
}

func (d *DecimalOperations) IsLessThan(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.LessThan(decB)
}

func (d *DecimalOperations) IsGreaterThanOrEqual(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.GreaterThanOrEqual(decB)
}

func (d *DecimalOperations) IsLessThanOrEqual(a, b float64) bool {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	return decA.LessThanOrEqual(decB)
}

func (d *DecimalOperations) Abs(value float64) float64 {
	dec := decimal.NewFromFloat(value)
	result := dec.Abs()
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) Min(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	if decA.LessThan(decB) {
		return a
	}
	return b
}

func (d *DecimalOperations) Max(a, b float64) float64 {
	decA := decimal.NewFromFloat(a)
	decB := decimal.NewFromFloat(b)
	if decA.GreaterThan(decB) {
		return a
	}
	return b
}

func (d *DecimalOperations) CompoundInterest(principal, rate float64, periods int) float64 {
	decPrincipal := decimal.NewFromFloat(principal)
	decRate := decimal.NewFromFloat(rate).Div(decimal.NewFromInt(100)) // Convert percentage to decimal
	decOne := decimal.NewFromInt(1)

	base := decOne.Add(decRate)

	result := decPrincipal
	for range periods {
		result = result.Mul(base)
	}

	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) SimpleInterest(principal, rate, time float64) float64 {
	decPrincipal := decimal.NewFromFloat(principal)
	decRate := decimal.NewFromFloat(rate).Div(decimal.NewFromInt(100))
	decTime := decimal.NewFromFloat(time)

	result := decPrincipal.Mul(decRate).Mul(decTime)
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) Clamp(value, min, max float64) float64 {
	if d.IsLessThan(value, min) {
		return min
	}
	if d.IsGreaterThan(value, max) {
		return max
	}
	return value
}

func (d *DecimalOperations) ClampMin(value, min float64) float64 {
	if d.IsLessThan(value, min) {
		return min
	}
	return value
}

func (d *DecimalOperations) ClampMax(value, max float64) float64 {
	if d.IsGreaterThan(value, max) {
		return max
	}
	return value
}

func (d *DecimalOperations) Negate(value float64) float64 {
	dec := decimal.NewFromFloat(value)
	result := dec.Neg()
	resultFloat, _ := result.Float64()
	return resultFloat
}

func (d *DecimalOperations) NegateInt(value int) float64 {
	dec := decimal.NewFromInt(int64(value))
	result := dec.Neg()
	resultFloat, _ := result.Float64()
	return resultFloat
}
