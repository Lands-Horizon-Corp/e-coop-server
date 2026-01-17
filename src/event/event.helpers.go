package event

import (
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func nextWeekday(from time.Time, weekday time.Weekday) time.Time {
	d := from.AddDate(0, 0, 1)
	for d.Weekday() != weekday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

func skippedDaysCount(
	startDate time.Time, currency *types.Currency, excludeSaturday, excludeSunday, excludeHolidays bool, holidays []*types.Holiday) (int, error) {
	skippedDays := 0
	currentDate := startDate
	for {
		skip, err := skippedDate(currentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
		if err != nil {
			return 0, err
		}
		if !skip {
			return skippedDays, nil
		}
		currentDate = currentDate.AddDate(0, 0, 1)
		skippedDays++
	}
}

func skippedDate(date time.Time, currency *types.Currency, excludeSaturday, excludeSunday, excludeHolidays bool, holidays []*types.Holiday) (bool, error) {
	if excludeSaturday {
		isSat, err := isSaturday(date, currency)
		if err != nil {
			return false, err
		}
		if isSat {
			return true, nil
		}
	}
	if excludeSunday {
		isSun, err := isSunday(date, currency)
		if err != nil {
			return false, err
		}
		if isSun {
			return true, nil
		}
	}
	if excludeHolidays {
		isHol, err := isHoliday(date, currency, holidays)
		if err != nil {
			return false, err
		}
		if isHol {
			return true, nil
		}
	}
	return false, nil
}

func isHoliday(date time.Time, currency *types.Currency, holidays []*types.Holiday) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc)
	year, month, day := localDate.Date()
	for _, holiday := range holidays {
		hYear, hMonth, hDay := holiday.EntryDate.Date()
		if year == hYear && month == hMonth && day == hDay {
			return true, nil
		}
	}
	return false, nil
}

func isSunday(date time.Time, currency *types.Currency) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc)
	weekday := localDate.Weekday()
	if weekday == time.Sunday {
		return true, nil
	}
	return false, nil
}

func isSaturday(date time.Time, currency *types.Currency) (bool, error) {
	loc, err := time.LoadLocation(currency.Timezone)
	if err != nil {
		return false, err
	}
	localDate := date.In(loc)
	weekday := localDate.Weekday()
	if weekday == time.Saturday {
		return true, nil
	}
	return false, nil
}
