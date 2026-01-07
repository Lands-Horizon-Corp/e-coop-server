package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

/*
1. Collection (Payment & Deposit)
2. Disbursement (Withdrawal, Loan Releases, Cash & Check Voucher)
3..Journal (Journal Voucher & Adjustment)
*/
func (e *Event) AccountTransactionProcess(
	context context.Context,
	userOrg core.UserOrganization,
	data core.AccountTransactionProcessGLRequest,
) error {
	startDate := time.Date(
		data.StartDate.Year(),
		data.StartDate.Month(),
		data.StartDate.Day(),
		0, 0, 0, 0,
		data.StartDate.Location(),
	)

	endDate := time.Date(
		data.EndDate.Year(),
		data.EndDate.Month(),
		data.EndDate.Day(),
		0, 0, 0, 0,
		data.EndDate.Location(),
	)

	if endDate.Before(startDate) {
		return eris.New("end date cannot be before start date")
	}

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		booking, err := e.core.DailyBookingCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return err
		}
		usecase.SumGeneralLedgerByAccount(booking)

		disbursement, err := e.core.DailyDisbursementCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return err
		}
		usecase.SumGeneralLedgerByAccount(disbursement)

		journal, err := e.core.DailyJournalCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return err
		}
		usecase.SumGeneralLedgerByAccount(journal)
	}

	return nil
}

func (e *Event) AccountTransactionLedgers(
	context context.Context,
	userOrg core.UserOrganization,
	year int,
	accountId *uuid.UUID,
) ([]*core.AccountTransactionLedgerResponse, error) {
	return nil, nil
}
