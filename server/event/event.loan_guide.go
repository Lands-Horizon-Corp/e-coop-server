package event

import (
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
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

	Value   float64            `json:"value" validate:"required,gte=0"`
	Balance float64            `json:"balance" validate:"required,gte=0"`
	Type    LoanScheduleStatus `json:"type" validate:"required,oneof=paid due overdue skipped advance"`
}
type LoanAccountSummary struct {
	LoanAccount      core.LoanAccountResponse `json:"loan_account"`
	PaymentSchedules []LoanPaymentSchedule    `json:"payment_schedules"`
}

type LoanGuideResponse struct {
	LoanAccounts []LoanAccountSummary `json:"loan_accounts"`
}

// func (e *Event) LoanGuide(
// 	context context.Context, useOrg core.UserOrganization, loanTransactionID uuid.UUID,
// ) (*LoanGuideResponse, error) {

// }
