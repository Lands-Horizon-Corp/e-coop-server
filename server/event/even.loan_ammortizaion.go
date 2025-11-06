package event

// func (e Event) LoanAmortizationSchedule(ctx context.Context, loanTransactionID uuid.UUID) error {
// 	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID)
// 	if err != nil {
// 		return err
// 	}
// 	holidays, err := e.core.HolidayManager.Find(ctx, &core.Holiday{
// 		OrganizationID: loanTransaction.OrganizationID,
// 		BranchID:       loanTransaction.BranchID,
// 	})

// 	for {

// 	}

// 	// return err
// 	return nil
// }
// func (e Event) isHoliday(date time.Time, currency *core.Currency, holidays []*core.Holiday) (bool, error) {

// 	loc, err := time.LoadLocation(currency.Timezone)
// 	if err != nil {
// 		return false, err
// 	}
// 	localDate := date.In(loc).Truncate(24 * time.Hour)
// 	for _, holiday := range holidays {
// 		holidayDate := holiday.EntryDate.In(loc).Truncate(24 * time.Hour)
// 		if localDate.Equal(holidayDate) {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

// func (e Event) isWeekend(date time.Time, currency *core.Currency) (bool, error) {
// 	loc, err := time.LoadLocation(currency.Timezone)
// 	if err != nil {
// 		return false, err
// 	}
// 	localDate := date.In(loc)
// 	weekday := localDate.Weekday()
// 	if weekday == time.Saturday || weekday == time.Sunday {
// 		return true, nil
// 	}
// 	return false, nil
// }
