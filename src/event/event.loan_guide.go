package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

type LoanScheduleStatus string

const (
	LoanScheduleStatusPaid    LoanScheduleStatus = "paid"
	LoanScheduleStatusDue     LoanScheduleStatus = "due"
	LoanScheduleStatusOverdue LoanScheduleStatus = "overdue"
	LoanScheduleStatusSkipped LoanScheduleStatus = "skipped"
	LoanScheduleStatusAdvance LoanScheduleStatus = "advance"
)

type LoanPayments struct {
	Amount  float64   `json:"amount"`
	PayDate time.Time `json:"pay_date"`

	GeneralLedger *types.GeneralLedger `json:"general_ledger"`
}
type LoanPaymentSchedule struct {
	LoanPayments LoanPayments `json:"loan_payments"`

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
	// loanTransaction, err := core.LoanTransactionManager(service).GetByID(ctx, loanTransactionID, "Account")
	// if err != nil {
	// 	return nil, eris.Wrap(err, "LoanGuide: failed to get loan transaction")
	// }
	// numberOfPayments, err := usecase.LoanNumberOfPayments(loanTransaction.ModeOfPayment, loanTransaction.Terms)
	// if err != nil {
	// 	return nil, eris.Wrapf(err, "failed to calculate number of payments for loan with mode: %s and terms: %d",
	// 		loanTransaction.ModeOfPayment, loanTransaction.Terms)
	// }

	// loanAccounts, err := core.LoanAccountManager(service).Find(ctx, &types.LoanAccount{
	// 	LoanTransactionID: loanTransaction.ID,
	// 	OrganizationID:    userOrg.OrganizationID,
	// 	BranchID:          *userOrg.BranchID,
	// })
	// if err != nil {
	// 	return nil, eris.Wrap(err, "LoanGuide: failed to find loan accounts")
	// }
	// amortization, err := LoanAmortization(ctx, service, loanTransactionID, userOrg)
	// if err != nil {
	// 	return nil, eris.Wrap(err, "GenerateLoanSchedule: failed to generate amortization")
	// }

	// currentDate := userOrg.TimeMachine()
	// for _, entry := range amortization.Schedule {
	// 	accountSummary := &LoanAccountSummary{
	// 		LoanAccount:      core.LoanAccountManager(service).ToModel(acc),
	// 		PaymentSchedules: []*LoanPaymentSchedule{},
	// 		TotalAmountDue:   0,
	// 		TotalAmountPaid:  0,
	// 		CurrentBalance:   0,
	// 		NextDueDate:      nil,
	// 		DaysOverdue:      0,
	// 		OverdueAmount:    0,
	// 		CompletionStatus: "active",
	// 	}
	// }

	// for _, acc := range loanAccounts {
	// 	// generalLedgers, err := core.GeneralLedgerManager(service).ArrFind(ctx, []query.ArrFilterSQL{
	// 	// 	{Field: "account_id", Op: query.ModeEqual, Value: acc.AccountID},
	// 	// 	{Field: "organization_id", Op: query.ModeEqual, Value: userOrg.OrganizationID},
	// 	// 	{Field: "branch_id", Op: query.ModeEqual, Value: userOrg.BranchID},
	// 	// }, []query.ArrFilterSortSQL{
	// 	// 	{Field: "entry_date", Order: query.SortOrderAsc},
	// 	// })
	// 	// for _, ledger := range generalLedgers {
	// 	// 	//
	// 	// }

	// 	if err != nil {
	// 		return nil, eris.Wrap(err, "LoanGuide: failed to fetch general ledgers")
	// 	}
	// 	accountSummary := &LoanAccountSummary{
	// 		LoanAccount:      core.LoanAccountManager(service).ToModel(acc),
	// 		PaymentSchedules: []*LoanPaymentSchedule{},ss
	// 		TotalAmountDue:   0,
	// 		TotalAmountPaid:  0,
	// 		CurrentBalance:   0,
	// 		NextDueDate:      nil,
	// 		DaysOverdue:      0,
	// 		OverdueAmount:    0,
	// 		CompletionStatus: "active",
	// 	}
	// 	response.LoanAccounts = append(response.LoanAccounts, accountSummary)
	// 	response.TotalLoans++
	// }

	return response, nil
}
