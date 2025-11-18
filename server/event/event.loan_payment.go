package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type LoanPaymentSchedule struct {
	Date          string  `json:"date"`
	Amount        float64 `json:"amount"`
	Paid          bool    `json:"paid"`
	Due           bool    `json:"due"`
	IsAdvance     bool    `json:"is_advance"`     // Payment made before due date
	IsFuture      bool    `json:"is_future"`      // Future scheduled payment
	DaysEarly     int     `json:"days_early"`     // Days paid before schedule (if advance)
	DaysOverdue   int     `json:"days_overdue"`   // Days overdue (if unpaid and past due)
	PaymentStatus string  `json:"payment_status"` // "paid", "overdue", "upcoming", "advance"
}

// LoanPaymentPerAccount represents payment details for a single loan account
type LoanPaymentPerAccount struct {
	Account                core.AccountResponse  `json:"account"`
	LoanPaymentSchedule    []LoanPaymentSchedule `json:"loan_payment_schedule"`
	TotalPrincipal         float64               `json:"total_principal"` // Total principal amount (sum of all scheduled payments for this account)
	TotalPaidAmount        float64               `json:"total_paid_amount"`
	TotalRemainingBalance  float64               `json:"total_remaining_balance"` // Remaining balance to be paid (TotalPrincipal - TotalPaidAmount)
	TotalDueAmount         float64               `json:"total_due_amount"`
	TotalAdvancePayment    float64               `json:"total_advance_payment"`    // Total amount paid in advance
	SuggestedPaymentAmount float64               `json:"suggested_payment_amount"` // Recommended payment (includes overdue + next upcoming)
	NextPaymentDate        string                `json:"next_payment_date"`
	LastPaymentDate        string                `json:"last_payment_date"`     // Date of the last payment made (day only)
	LastPaymentAmount      float64               `json:"last_payment_amount"`   // Sum of all payments made on the last payment date
	AdvancePaymentCount    int                   `json:"advance_payment_count"` // Number of advance payments made
	OverduePaymentCount    int                   `json:"overdue_payment_count"` // Number of overdue payments
	IsLoanFullyPaid        bool                  `json:"is_loan_fully_paid"`    // True if all scheduled payments for this account are completed
}

// LoanPaymentSummary represents aggregated summary across all accounts
type LoanPaymentSummary struct {
	TotalAccounts           int     `json:"total_accounts"`             // Total number of loan accounts
	TotalPrincipal          float64 `json:"total_principal"`            // Total principal amount across all accounts (sum of all scheduled payments)
	TotalPaidAmount         float64 `json:"total_paid_amount"`          // Sum of all paid amounts
	TotalRemainingBalance   float64 `json:"total_remaining_balance"`    // Total remaining balance across all accounts (TotalPrincipal - TotalPaidAmount)
	TotalDueAmount          float64 `json:"total_due_amount"`           // Sum of all due amounts
	TotalAdvancePayment     float64 `json:"total_advance_payment"`      // Sum of all advance payments
	TotalSuggestedPayment   float64 `json:"total_suggested_payment"`    // Sum of all suggested payments across accounts
	TotalScheduledPayments  int     `json:"total_scheduled_payments"`   // Total number of scheduled payments
	TotalPaidPayments       int     `json:"total_paid_payments"`        // Total number of paid payments
	TotalOverduePayments    int     `json:"total_overdue_payments"`     // Total number of overdue payments
	TotalAdvancePayments    int     `json:"total_advance_payments"`     // Total number of advance payments
	TotalUpcomingPayments   int     `json:"total_upcoming_payments"`    // Total number of upcoming payments
	EarliestNextPaymentDate string  `json:"earliest_next_payment_date"` // Earliest next payment date across all accounts
	LastPaymentDate         string  `json:"last_payment_date"`          // Most recent payment date across all accounts (day only)
	LastPaymentAmount       float64 `json:"last_payment_amount"`        // Sum of all payments made on the last payment date across all accounts
	AccountsWithOverdue     int     `json:"accounts_with_overdue"`      // Number of accounts with overdue payments
	AccountsFullyPaid       int     `json:"accounts_fully_paid"`        // Number of accounts fully paid
	AccountsWithAdvance     int     `json:"accounts_with_advance"`      // Number of accounts with advance payments
	OverallPaymentStatus    string  `json:"overall_payment_status"`     // "current", "overdue", "advance", "mixed"
	IsLoanFullyPaid         bool    `json:"is_loan_fully_paid"`         // True if all scheduled payments across all accounts are completed
}

// LoanPaymentResponse is the main response structure containing both details and summary
type LoanPaymentResponse struct {
	AccountPayments []LoanPaymentPerAccount `json:"account_payments"` // Payment details per account
	Summary         LoanPaymentSummary      `json:"summary"`          // Aggregated summary across all accounts
}

func (e *Event) LoanPaymenSummary(
	context context.Context,
	loanTransactionID *uuid.UUID,
	userOrg *core.UserOrganization,
) (*LoanPaymentResponse, error) {
	// ===============================================================================================
	// STEP 1: VALIDATE INPUT PARAMETERS
	// ===============================================================================================

	if loanTransactionID == nil {
		return nil, eris.New("loan transaction ID is required")
	}
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve loan transaction with id: %s", *loanTransactionID)
	}

	accounts, err := e.core.AccountManager.Find(context, &core.Account{
		BranchID:       *userOrg.BranchID,
		OrganizationID: userOrg.OrganizationID,
		LoanAccountID:  loanTransaction.AccountID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction id: %s", *loanTransactionID)
	}
	accounts = append(accounts, loanTransaction.Account)

	// ===============================================================================================
	// STEP 2: INITIALIZE RESULT CONTAINERS AND SUMMARY TRACKERS
	// ===============================================================================================
	accountPayments := []LoanPaymentPerAccount{}

	// Summary aggregation variables
	summaryTotalPrincipal := 0.0
	summaryTotalPaidAmount := 0.0
	summaryTotalDueAmount := 0.0
	summaryTotalAdvancePayment := 0.0
	summaryTotalSuggestedPayment := 0.0
	summaryTotalScheduledPayments := 0
	summaryTotalPaidPayments := 0
	summaryTotalOverduePayments := 0
	summaryTotalAdvancePayments := 0
	summaryTotalUpcomingPayments := 0
	summaryAccountsWithOverdue := 0
	summaryAccountsFullyPaid := 0
	summaryAccountsWithAdvance := 0
	var earliestNextPaymentDate *time.Time
	var overallLastPaymentDate *time.Time
	overallLastPaymentAmount := 0.0

	// ===============================================================================================
	// STEP 3: PROCESS EACH ACCOUNT
	// ===============================================================================================
	for _, account := range accounts {

		// -------------------------------------------------------------------------------------------
		// 3.3: Get Amortization Schedule
		// -------------------------------------------------------------------------------------------
		amortizationSchedule, err := e.LoanAmortizationSchedule(context, loanTransaction.ID, userOrg)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to retrieve amortization schedule for loan transaction id: %s", loanTransaction.ID)
		}

		// -------------------------------------------------------------------------------------------
		// 3.4: Get General Ledger Entries for Payment History
		// -------------------------------------------------------------------------------------------
		entries, err := e.core.GeneralLedgerByLoanTransaction(
			context,
			loanTransaction.ID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to retrieve general ledger entries for loan transaction id: %s", loanTransaction.ID)
		}

		// Calculate balance to track payments
		balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: entries,
			AccountID:      &account.ID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for account id: %s", &account.ID)
		}

		// -------------------------------------------------------------------------------------------
		// 3.5: Build Payment Schedule from Amortization
		// -------------------------------------------------------------------------------------------
		paymentSchedule := []LoanPaymentSchedule{}
		totalPrincipal := 0.0 // Total principal for this account (sum of all scheduled payments)
		totalPaidAmount := 0.0
		totalDueAmount := 0.0
		totalAdvancePayment := 0.0
		advancePaymentCount := 0
		overduePaymentCount := 0
		paidPaymentCount := 0
		upcomingPaymentCount := 0
		suggestedPaymentAmount := 0.0
		nextUpcomingAmount := 0.0
		var nextUpcomingDate *time.Time // Track the date of next upcoming payment
		var nextPaymentDate *time.Time
		var lastPaymentDate *time.Time
		lastPaymentAmount := 0.0 // Sum of all payments on the last payment date
		now := userOrg.UserOrgTime()

		if amortizationSchedule != nil && amortizationSchedule.Schedule != nil {
			cumulativeExpected := 0.0
			cumulativePaid := balance.Credit

			for _, schedule := range amortizationSchedule.Schedule {
				// Find the payment amount for this account in the schedule
				paymentAmount := 0.0
				for _, accountValue := range schedule.Accounts {
					if accountValue.Account != nil && handlers.UUIDPtrEqual(
						&accountValue.Account.ID, &account.ID) {
						paymentAmount = accountValue.Value
						break
					}
				}

				// Skip if no payment amount found for this account in this schedule
				if paymentAmount == 0.0 {
					continue
				}

				// Add to total principal (all scheduled payments)
				totalPrincipal = e.provider.Service.Decimal.Add(totalPrincipal, paymentAmount)

				// Count scheduled payments for this account
				summaryTotalScheduledPayments++

				// Determine if this payment has been made using decimal-safe comparison
				cumulativeExpected = e.provider.Service.Decimal.Add(cumulativeExpected, paymentAmount)
				// Use a small epsilon for float comparison to handle precision issues
				const epsilon = 0.000001
				isPaid := (cumulativePaid - cumulativeExpected) >= -epsilon

				// Check if payment is due (scheduled date is today or in the past)
				// Truncate to start of day for accurate date-only comparison
				scheduledDateStart := time.Date(schedule.ScheduledDate.Year(), schedule.ScheduledDate.Month(), schedule.ScheduledDate.Day(), 0, 0, 0, 0, schedule.ScheduledDate.Location())
				nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				isDue := scheduledDateStart.Before(nowStart) || scheduledDateStart.Equal(nowStart)

				// Determine if this is an advance payment (paid before due date)
				isAdvance := isPaid && !isDue
				// isFuture: payment is scheduled in the future and not yet paid
				isFuture := !isDue && !isPaid

				// Calculate days early or overdue
				daysEarly := 0
				daysOverdue := 0
				paymentStatus := "upcoming"

				if isPaid {
					paidPaymentCount++
					// Add to total paid amount (all paid payments)
					totalPaidAmount = e.provider.Service.Decimal.Add(totalPaidAmount, paymentAmount)

					// Track last payment date and sum all payments on that date (date-only comparison)
					scheduledDateOnly := time.Date(schedule.ScheduledDate.Year(), schedule.ScheduledDate.Month(), schedule.ScheduledDate.Day(), 0, 0, 0, 0, schedule.ScheduledDate.Location())
					if lastPaymentDate == nil {
						lastPaymentDate = &scheduledDateOnly
						lastPaymentAmount = paymentAmount
					} else {
						lastPaymentDateOnly := time.Date(lastPaymentDate.Year(), lastPaymentDate.Month(), lastPaymentDate.Day(), 0, 0, 0, 0, lastPaymentDate.Location())
						if scheduledDateOnly.After(lastPaymentDateOnly) {
							// New later date found, reset amount
							lastPaymentDate = &scheduledDateOnly
							lastPaymentAmount = paymentAmount
						} else if scheduledDateOnly.Equal(lastPaymentDateOnly) {
							// Same date, accumulate the amount
							lastPaymentAmount = e.provider.Service.Decimal.Add(lastPaymentAmount, paymentAmount)
						}
					}

					if isAdvance {
						// Calculate days paid in advance (scheduled date - current date)
						// Use start of day for accurate date difference
						daysDiff := scheduledDateStart.Sub(nowStart).Hours() / 24
						daysEarly = max(int(daysDiff), 0)
						paymentStatus = "advance"
						advancePaymentCount++
						// Track advance payment separately (subset of paid)
						totalAdvancePayment = e.provider.Service.Decimal.Add(totalAdvancePayment, paymentAmount)
					} else {
						paymentStatus = "paid"
					}
				} else if isDue {
					// Calculate days overdue (current date - scheduled date)
					// Use start of day for accurate date difference
					daysDiff := nowStart.Sub(scheduledDateStart).Hours() / 24
					daysOverdue = max(int(daysDiff), 0)
					paymentStatus = "overdue"
					totalDueAmount = e.provider.Service.Decimal.Add(totalDueAmount, paymentAmount)
					overduePaymentCount++
				} else {
					upcomingPaymentCount++

					// Capture the next upcoming payment amount (chronologically nearest unpaid future payment)
					if nextUpcomingDate == nil || schedule.ScheduledDate.Before(*nextUpcomingDate) {
						nextUpcomingDate = &schedule.ScheduledDate
						nextUpcomingAmount = paymentAmount
					}
				}

				// Capture next payment date (first unpaid scheduled date, either overdue or upcoming)
				if !isPaid {
					if nextPaymentDate == nil || schedule.ScheduledDate.Before(*nextPaymentDate) {
						nextPaymentDate = &schedule.ScheduledDate
					}
				}

				// Add to schedule
				paymentSchedule = append(paymentSchedule, LoanPaymentSchedule{
					Date:          schedule.ScheduledDate.Format("2006-01-02"),
					Amount:        paymentAmount,
					Paid:          isPaid,
					Due:           isDue && !isPaid,
					IsAdvance:     isAdvance,
					IsFuture:      isFuture,
					DaysEarly:     daysEarly,
					DaysOverdue:   daysOverdue,
					PaymentStatus: paymentStatus,
				})
			}
		}

		// -------------------------------------------------------------------------------------------
		// 3.6: Calculate Suggested Payment Amount
		// -------------------------------------------------------------------------------------------
		// Suggested payment = All overdue amounts + Next upcoming payment
		// Edge cases:
		// - If no overdue and no upcoming: suggestedPaymentAmount = 0 (fully paid or advance paid)
		// - If all payments overdue: suggestedPaymentAmount = totalDueAmount (no upcoming to add)
		// - If some overdue + upcoming: suggestedPaymentAmount = overdue + next scheduled
		suggestedPaymentAmount = e.provider.Service.Decimal.Add(totalDueAmount, nextUpcomingAmount)

		// -------------------------------------------------------------------------------------------
		// 3.7: Format Next Payment Date and Last Payment Date
		// -------------------------------------------------------------------------------------------
		nextPaymentDateStr := ""
		if nextPaymentDate != nil {
			nextPaymentDateStr = nextPaymentDate.Format("2006-01-02")

			// Track earliest next payment date across all accounts
			if earliestNextPaymentDate == nil || nextPaymentDate.Before(*earliestNextPaymentDate) {
				earliestNextPaymentDate = nextPaymentDate
			}
		}

		lastPaymentDateStr := ""
		if lastPaymentDate != nil {
			lastPaymentDateStr = lastPaymentDate.Format("2006-01-02")

			// Track most recent payment date across all accounts (date-only comparison)
			lastPaymentDateOnly := time.Date(lastPaymentDate.Year(), lastPaymentDate.Month(), lastPaymentDate.Day(), 0, 0, 0, 0, lastPaymentDate.Location())
			if overallLastPaymentDate == nil {
				overallLastPaymentDate = &lastPaymentDateOnly
				overallLastPaymentAmount = lastPaymentAmount
			} else {
				overallLastPaymentDateOnly := time.Date(overallLastPaymentDate.Year(), overallLastPaymentDate.Month(), overallLastPaymentDate.Day(), 0, 0, 0, 0, overallLastPaymentDate.Location())
				if lastPaymentDateOnly.After(overallLastPaymentDateOnly) {
					// New later date found, reset amount
					overallLastPaymentDate = &lastPaymentDateOnly
					overallLastPaymentAmount = lastPaymentAmount
				} else if lastPaymentDateOnly.Equal(overallLastPaymentDateOnly) {
					// Same date, accumulate the amount from this account
					overallLastPaymentAmount = e.provider.Service.Decimal.Add(overallLastPaymentAmount, lastPaymentAmount)
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 3.8: Update Summary Aggregations
		// -------------------------------------------------------------------------------------------
		summaryTotalPrincipal = e.provider.Service.Decimal.Add(summaryTotalPrincipal, totalPrincipal)
		summaryTotalPaidAmount = e.provider.Service.Decimal.Add(summaryTotalPaidAmount, totalPaidAmount)
		summaryTotalDueAmount = e.provider.Service.Decimal.Add(summaryTotalDueAmount, totalDueAmount)
		summaryTotalAdvancePayment = e.provider.Service.Decimal.Add(summaryTotalAdvancePayment, totalAdvancePayment)
		summaryTotalSuggestedPayment = e.provider.Service.Decimal.Add(summaryTotalSuggestedPayment, suggestedPaymentAmount)
		summaryTotalPaidPayments += paidPaymentCount
		summaryTotalOverduePayments += overduePaymentCount
		summaryTotalAdvancePayments += advancePaymentCount
		summaryTotalUpcomingPayments += upcomingPaymentCount

		// Track account-level statuses
		if overduePaymentCount > 0 {
			summaryAccountsWithOverdue++
		}
		if advancePaymentCount > 0 {
			summaryAccountsWithAdvance++
		}
		// Account is fully paid if all scheduled payments are paid (no overdue, no upcoming unpaid)
		totalScheduledForAccount := len(paymentSchedule)
		isAccountFullyPaid := totalScheduledForAccount > 0 && paidPaymentCount == totalScheduledForAccount
		if isAccountFullyPaid {
			summaryAccountsFullyPaid++
		}

		// -------------------------------------------------------------------------------------------
		// 3.9: Build and Append Account Payment
		// -------------------------------------------------------------------------------------------
		totalRemainingBalance := e.provider.Service.Decimal.Subtract(totalPrincipal, totalPaidAmount)
		accountPayments = append(accountPayments, LoanPaymentPerAccount{
			Account:                *e.core.AccountManager.ToModel(account),
			LoanPaymentSchedule:    paymentSchedule,
			TotalPrincipal:         totalPrincipal,
			TotalPaidAmount:        totalPaidAmount,
			TotalRemainingBalance:  totalRemainingBalance,
			TotalDueAmount:         totalDueAmount,
			TotalAdvancePayment:    totalAdvancePayment,
			SuggestedPaymentAmount: suggestedPaymentAmount,
			NextPaymentDate:        nextPaymentDateStr,
			LastPaymentDate:        lastPaymentDateStr,
			LastPaymentAmount:      lastPaymentAmount,
			AdvancePaymentCount:    advancePaymentCount,
			OverduePaymentCount:    overduePaymentCount,
			IsLoanFullyPaid:        isAccountFullyPaid,
		})
	}

	// ===============================================================================================
	// STEP 4: BUILD OVERALL SUMMARY
	// ===============================================================================================
	earliestNextPaymentDateStr := ""
	if earliestNextPaymentDate != nil {
		earliestNextPaymentDateStr = earliestNextPaymentDate.Format("2006-01-02")
	}

	overallLastPaymentDateStr := ""
	if overallLastPaymentDate != nil {
		overallLastPaymentDateStr = overallLastPaymentDate.Format("2006-01-02")
	}

	// Determine overall payment status
	overallStatus := "current"
	if summaryAccountsWithOverdue > 0 {
		overallStatus = "overdue"
	} else if summaryAccountsWithAdvance > 0 && summaryAccountsWithOverdue == 0 {
		overallStatus = "advance"
	} else if summaryAccountsWithAdvance > 0 && summaryAccountsWithOverdue > 0 {
		overallStatus = "mixed"
	}

	// Determine if the entire loan is fully paid (all accounts fully paid)
	isOverallLoanFullyPaid := len(accountPayments) > 0 && summaryAccountsFullyPaid == len(accountPayments)

	// Calculate overall remaining balance
	summaryTotalRemainingBalance := e.provider.Service.Decimal.Subtract(summaryTotalPrincipal, summaryTotalPaidAmount)

	summary := LoanPaymentSummary{
		TotalAccounts:           len(accountPayments),
		TotalPrincipal:          summaryTotalPrincipal,
		TotalPaidAmount:         summaryTotalPaidAmount,
		TotalRemainingBalance:   summaryTotalRemainingBalance,
		TotalDueAmount:          summaryTotalDueAmount,
		TotalAdvancePayment:     summaryTotalAdvancePayment,
		TotalSuggestedPayment:   summaryTotalSuggestedPayment,
		TotalScheduledPayments:  summaryTotalScheduledPayments,
		TotalPaidPayments:       summaryTotalPaidPayments,
		TotalOverduePayments:    summaryTotalOverduePayments,
		TotalAdvancePayments:    summaryTotalAdvancePayments,
		TotalUpcomingPayments:   summaryTotalUpcomingPayments,
		EarliestNextPaymentDate: earliestNextPaymentDateStr,
		LastPaymentDate:         overallLastPaymentDateStr,
		LastPaymentAmount:       overallLastPaymentAmount,
		AccountsWithOverdue:     summaryAccountsWithOverdue,
		AccountsFullyPaid:       summaryAccountsFullyPaid,
		AccountsWithAdvance:     summaryAccountsWithAdvance,
		OverallPaymentStatus:    overallStatus,
		IsLoanFullyPaid:         isOverallLoanFullyPaid,
	}

	// ===============================================================================================
	// STEP 5: RETURN COMPREHENSIVE RESPONSE
	// ===============================================================================================
	return &LoanPaymentResponse{
		AccountPayments: accountPayments,
		Summary:         summary,
	}, nil
}
