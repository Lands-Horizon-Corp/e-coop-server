package event

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
)

func (e *Event) AccountTransactionProcess(
	context context.Context,
	userOrg core.UserOrganization,
	data core.AccountTransactionProcessGLRequest,
) error {
	// data.StartDate data.EndDate
	// All ledgers group them by account

	// find all ledgers
	// Collection

	/*
		1. Collection (Payment & Deposit)
		2. Disbursement (Withdrawal, Loan Releases, Cash & Check Voucher)
		3..Journal (Journal Voucher & Adjustment)
	*/
	// e.core.DailyBookingCollection()
	// e.core.DailyDisbursementCollection()
	// e.core.DailyJournalCollection()
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
