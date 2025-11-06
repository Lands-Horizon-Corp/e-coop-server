package event

// func (e Event) LoanAmortizationSchedule(ctx context.Context, loanTransactionID uuid.UUID) error {
// 	loanTransaction, err := e.core.LoanTransactionManager.GetByID(ctx, loanTransactionID, "Account.Currency")
// 	if err != nil {
// 		return err
// 	}
// 	holidays, err := e.core.HolidayManager.Find(ctx, &core.Holiday{
// 		OrganizationID: loanTransaction.OrganizationID,
// 		BranchID:       loanTransaction.BranchID,
// 	})

// 	numberOfPayments, err := e.usecase.LoanNumberOfPayments(loanTransaction)
// 	if err != nil {
// 		return err
// 	}

// 	currency := loanTransaction.Account.Currency

// 	// Excluding
// 	excludeSaturday := loanTransaction.ExcludeSaturday
// 	excludeSunday := loanTransaction.ExcludeSunday
// 	excludeHolidays := loanTransaction.ExcludeHoliday

// 	// Monthly Exact Day
// 	isMonthlyExactDay := loanTransaction.ModeOfPaymentMonthlyExactDay

// 	// Weekly Exact Day monday - tuesday, wednesday, thursday, friday, saturday, sunday
// 	weeklyExactDay := loanTransaction.ModeOfPaymentWeekly

// 	// For Semi-Monthly Exact Days
// 	semiMonthlyExactDay1 := loanTransaction.ModeOfPaymentSemiMonthlyPay1
// 	semiMonthlyExactDay2 := loanTransaction.ModeOfPaymentSemiMonthlyPay2

// 	paymentDate := time.Now().UTC()

// 	for i := 0; i < numberOfPayments; i++ {
// 		for {
// 			var skip bool
// 			if excludeSaturday {
// 				if sat, _ := e.isSaturday(paymentDate, currency); sat {
// 					skip = true
// 				}
// 			}
// 			if excludeSunday {
// 				if sun, _ := e.isSunday(paymentDate, currency); sun {
// 					skip = true
// 				}
// 			}
// 			if excludeHolidays {
// 				if hol, _ := e.isHoliday(paymentDate, currency, holidays); hol {
// 					skip = true
// 				}
// 			}
// 			if !skip {
// 				break
// 			}
// 			paymentDate = paymentDate.AddDate(0, 0, 1)
// 		}
// 		switch loanTransaction.ModeOfPayment {
// 		case core.LoanModeOfPaymentDaily:
// 		case core.LoanModeOfPaymentWeekly:
// 		case core.LoanModeOfPaymentSemiMonthly:
// 		case core.LoanModeOfPaymentMonthly:
// 		case core.LoanModeOfPaymentQuarterly:
// 		case core.LoanModeOfPaymentSemiAnnual:
// 		case core.LoanModeOfPaymentLumpsum:
// 		case core.LoanModeOfPaymentFixedDays:
// 		}
// 	}

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

// func (e Event) isSunday(date time.Time, currency *core.Currency) (bool, error) {
// 	loc, err := time.LoadLocation(currency.Timezone)
// 	if err != nil {
// 		return false, err
// 	}
// 	localDate := date.In(loc)
// 	weekday := localDate.Weekday()
// 	if weekday == time.Sunday {
// 		return true, nil
// 	}
// 	return false, nil
// }

// func (e Event) isSaturday(date time.Time, currency *core.Currency) (bool, error) {
// 	loc, err := time.LoadLocation(currency.Timezone)
// 	if err != nil {
// 		return false, err
// 	}
// 	localDate := date.In(loc)
// 	weekday := localDate.Weekday()
// 	if weekday == time.Saturday {
// 		return true, nil
// 	}
// 	return false, nil
// }
