package event

import (
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

// helper: returns the next date after 'from' that falls on 'weekday'
func (e Event) nextWeekday(from time.Time, weekday time.Weekday) time.Time {
	// Move to the next day to avoid returning the current day if matches
	d := from.AddDate(0, 0, 1)
	for d.Weekday() != weekday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

func (e Event) skippedDaysCount(
	startDate time.Time, currency *core.Currency, excludeSaturday, excludeSunday, excludeHolidays bool, holidays []*core.Holiday) (int, error) {
	skippedDays := 0
	currentDate := startDate
	for {
		skip, err := e.skippedDate(currentDate, currency, excludeSaturday, excludeSunday, excludeHolidays, holidays)
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

func (e Event) skippedDate(date time.Time, currency *core.Currency, excludeSaturday, excludeSunday, excludeHolidays bool, holidays []*core.Holiday) (bool, error) {
	if excludeSaturday {
		isSat, err := e.isSaturday(date, currency)
		if err != nil {
			return false, err
		}
		if isSat {
			return true, nil
		}
	}
	if excludeSunday {
		isSun, err := e.isSunday(date, currency)
		if err != nil {
			return false, err
		}
		if isSun {
			return true, nil
		}
	}
	if excludeHolidays {
		isHol, err := e.isHoliday(date, currency, holidays)
		if err != nil {
			return false, err
		}
		if isHol {
			return true, nil
		}
	}
	return false, nil
}

func (e Event) isHoliday(date time.Time, currency *core.Currency, holidays []*core.Holiday) (bool, error) {
	// Convert to the currency's timezone
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

func (e Event) isSunday(date time.Time, currency *core.Currency) (bool, error) {
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

func (e Event) isSaturday(date time.Time, currency *core.Currency) (bool, error) {
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
