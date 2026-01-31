package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type LoanScheduleStatus string

const (
	LoanScheduleStatusPaid    LoanScheduleStatus = "paid"
	LoanScheduleStatusDue     LoanScheduleStatus = "due"
	LoanScheduleStatusOverdue LoanScheduleStatus = "overdue"
	LoanScheduleStatusSkipped LoanScheduleStatus = "skipped"
	LoanScheduleStatusAdvance LoanScheduleStatus = "advance"
	LoanScheduleStatusDefault LoanScheduleStatus = "default"
)

type LoanPayments struct {
	Amount  float64   `json:"amount"`
	PayDate time.Time `json:"pay_date"`

	GeneralLedger *types.GeneralLedger `json:"general_ledger"`
}
type LoanPaymentSchedule struct {
	LoanPayments []*LoanPayments `json:"loan_payments"`

	PaymentDate   time.Time `json:"payment_date"`
	ScheduledDate time.Time `json:"scheduled_date"`
	ActualDate    time.Time `json:"actual_date"`
	DaysSkipped   int       `json:"days_skipped"`

	AmountDue  float64 `json:"amount_due" validate:"required,gte=0"`  // due or overdue
	AmountPaid float64 `json:"amount_paid" validate:"required,gte=0"` // paid or advance

	Balance         float64 `json:"balance" validate:"required,gte=0"`          // Principal amount + interest amount + Fines
	PrincipalAmount float64 `json:"principal_amount" validate:"required,gte=0"` // the amount the user will pay

	InterestAmount float64 `json:"interest_amount" validate:"required,gte=0"`
	FinesAmount    float64 `json:"fines_amount" validate:"required,gte=0"`

	Type LoanScheduleStatus `json:"type" validate:"required,oneof=paid due overdue skipped advance"`
}
type LoanAccountSummary struct {
	LoanAccount      *types.LoanAccountResponse `json:"loan_account"`
	PaymentSchedules []*LoanPaymentSchedule     `json:"payment_schedules"`
	TotalAmountDue   float64                    `json:"total_amount_due"`
	TotalAmountPaid  float64                    `json:"total_amount_paid"`
	CurrentBalance   float64                    `json:"current_balance"`
	NextDueDate      *time.Time                 `json:"next_due_date,omitempty"`
	DaysOverdue      int                        `json:"days_overdue"`
	OverdueAmount    float64                    `json:"overdue_amount"`
	CompletionStatus string                     `json:"completion_status"`
}

type LoanGuideResponse struct {
	LoanAccounts     []*LoanAccountSummary `json:"loan_accounts"`
	TotalLoans       int                   `json:"total_loans"`
	ActiveLoans      int                   `json:"active_loans"`
	CompletedLoans   int                   `json:"completed_loans"`
	DefaultedLoans   int                   `json:"defaulted_loans"`
	TotalOutstanding float64               `json:"total_outstanding"`
	TotalOverdue     float64               `json:"total_overdue"`
}

func LoanGuide(
	ctx context.Context,
	service *horizon.HorizonService,
	userOrg *types.UserOrganization,
	loanTransactionID uuid.UUID,
) (*LoanGuideResponse, error) {
	response := &LoanGuideResponse{
		LoanAccounts:     []*LoanAccountSummary{},
		TotalLoans:       0,
		ActiveLoans:      0,
		CompletedLoans:   0,
		DefaultedLoans:   0,
		TotalOutstanding: 0,
		TotalOverdue:     0,
	}
	loanTransaction, err := core.LoanTransactionManager(service).GetByID(ctx, loanTransactionID, "Account")
	if err != nil {
		return nil, eris.Wrap(err, "LoanGuide: failed to get loan transaction")
	}
	loanAccounts, err := core.LoanAccountManager(service).Find(ctx, &types.LoanAccount{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    userOrg.OrganizationID,
		BranchID:          *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "LoanGuide: failed to find loan accounts")
	}
	amortization, err := LoanAmortization(ctx, service, loanTransactionID, userOrg)
	if err != nil {
		return nil, eris.Wrap(err, "LoanGuide: GenerateLoanSchedule: failed to generate amortization")
	}
	// currentDate := userOrg.TimeMachine()

	for _, acc := range loanAccounts {
		schedule := []*LoanPaymentSchedule{}

		// LoanScheduleStatusPaid
		// LoanScheduleStatusDue
		// LoanScheduleStatusOverdue
		// LoanScheduleStatusSkipped
		// LoanScheduleStatusAdvance
		// LoanScheduleStatusDefault

		// LoanScheduleStatusOverdue
		generalLedgers, err := core.GeneralLedgerManager(service).ArrFind(ctx, []query.ArrFilterSQL{
			{Field: "account_id", Op: query.ModeEqual, Value: acc.AccountID},
			{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
			{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
		}, []query.ArrFilterSortSQL{
			{Field: "entry_date", Order: query.SortOrderAsc},
		})
		if err != nil {
			return nil, eris.Wrap(err, "LoanGuide: error getting general ledger")
		}
		fmt.Println(generalLedgers)
		filterSchedule(amortization.Schedule, acc.AccountID, func(entry *LoanAmortizationSchedule, schedAcc *AccountValue) {
			payments := []*LoanPayments{}
			for len(generalLedgers) > 0 {
				gl := generalLedgers[0]
				if gl.EntryDate.After(entry.ScheduledDate) {
					break
				}
				payments = append(payments, &LoanPayments{
					Amount:        gl.Credit,
					PayDate:       gl.EntryDate,
					GeneralLedger: gl,
				})
				generalLedgers = generalLedgers[1:]
			}
			// for each general ledger (payments)
			// 		if payment date <= entry.ScheduledDate add it to payments and make sure if its payed then remove it from stacks the general ledger
			// if  entry.ScheduledDate is behind current date   just put due
			// if  entry.ScheduledDate is behind 1 month   then overdue

			scheduleType := LoanScheduleStatusDefault
			// daysPast := int(currentDate.Sub(entry.ScheduledDate).Hours() / 24)
			// if daysPast < 0 {
			// 	// Future: treat as due (upcoming payment).
			// 	scheduleType = LoanScheduleStatusDue
			// } else if daysPast <= 60 { // Approx 2 months (current or recent past due).
			// 	scheduleType = LoanScheduleStatusDue
			// } else {
			// 	// More than 2 months past: overdue.
			// 	scheduleType = LoanScheduleStatusOverdue
			// }

			schedule = append(schedule, &LoanPaymentSchedule{
				LoanPayments: payments,
				PaymentDate:  entry.ScheduledDate,
				ActualDate:   entry.ActualDate,
				DaysSkipped:  entry.DaysSkipped,
				Balance:      schedAcc.Value,
				Type:         scheduleType,
			})
		})

		accountSummary := &LoanAccountSummary{
			LoanAccount:      core.LoanAccountManager(service).ToModel(acc),
			PaymentSchedules: schedule,
			TotalAmountDue:   0,
			TotalAmountPaid:  0,
			CurrentBalance:   0,
			NextDueDate:      nil,
			DaysOverdue:      0,
			OverdueAmount:    0,
			CompletionStatus: "active",
		}
		response.LoanAccounts = append(response.LoanAccounts, accountSummary)
	}

	return response, nil
}

func filterSchedule(
	schedule []*LoanAmortizationSchedule,
	accID *uuid.UUID,
	callback func(*LoanAmortizationSchedule, *AccountValue),
) {
	for _, entry := range schedule {
		for _, schedAcc := range entry.Accounts {
			if helpers.UUIDPtrEqual(&schedAcc.Account.ID, accID) {
				callback(entry, schedAcc)
				break
			}
		}
	}
}
