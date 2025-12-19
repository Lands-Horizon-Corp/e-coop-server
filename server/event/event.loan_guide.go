package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
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

type LoanPaymentSchedule struct {
	PaymentDate   time.Time `json:"payment_date"`
	ScheduledDate time.Time `json:"scheduled_date"`
	ActualDate    time.Time `json:"actual_date"`
	DaysSkipped   int       `json:"days_skipped"`

	AmountDue       float64            `json:"amount_due" validate:"required,gte=0"`
	AmountPaid      float64            `json:"amount_paid" validate:"required,gte=0"`
	Balance         float64            `json:"balance" validate:"required,gte=0"`
	PrincipalAmount float64            `json:"principal_amount" validate:"required,gte=0"`
	InterestAmount  float64            `json:"interest_amount" validate:"required,gte=0"`
	FinesAmount     float64            `json:"fines_amount" validate:"required,gte=0"`
	Type            LoanScheduleStatus `json:"type" validate:"required,oneof=paid due overdue skipped advance"`
}
type LoanAccountSummary struct {
	LoanAccount      core.LoanAccountResponse `json:"loan_account"`
	PaymentSchedules []LoanPaymentSchedule    `json:"payment_schedules"`
	TotalAmountDue   float64                  `json:"total_amount_due"`
	TotalAmountPaid  float64                  `json:"total_amount_paid"`
	CurrentBalance   float64                  `json:"current_balance"`
	NextDueDate      *time.Time               `json:"next_due_date,omitempty"`
	DaysOverdue      int                      `json:"days_overdue"`
	OverdueAmount    float64                  `json:"overdue_amount"`
	CompletionStatus string                   `json:"completion_status"`
}

type LoanGuideResponse struct {
	LoanAccounts     []LoanAccountSummary `json:"loan_accounts"`
	TotalLoans       int                  `json:"total_loans"`
	ActiveLoans      int                  `json:"active_loans"`
	CompletedLoans   int                  `json:"completed_loans"`
	DefaultedLoans   int                  `json:"defaulted_loans"`
	TotalOutstanding float64              `json:"total_outstanding"`
	TotalOverdue     float64              `json:"total_overdue"`
}

func (e *Event) LoanGuide(
	context context.Context, userOrg *core.UserOrganization, loanTransactionID uuid.UUID,
) (*LoanGuideResponse, error) {
	return nil, nil
}
