package event

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

/*
1. Collection (Payment & Deposit)
2. Disbursement (Withdrawal, Loan Releases, Cash & Check Voucher)
3..Journal (Journal Voucher & Adjustment)
*/
const (
	DailyCollectionBook       = 1
	CashCheckDisbursementBook = 2
	GeneralJournalBook        = 3
)

func BuildJVNumberSimple(date time.Time, bookType int) string {
	monthDay := date.Format("0102")
	var bookCode string
	switch bookType {
	case DailyCollectionBook:
		bookCode = "01"
	case CashCheckDisbursementBook:
		bookCode = "02"
	case GeneralJournalBook:
		bookCode = "03"
	default:
		bookCode = "00"
	}
	return fmt.Sprintf("%s%s", monthDay, bookCode)
}

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
	now := time.Now()
	tx, endTx := e.provider.Service.Database.StartTransaction(context)

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		// Destroy previous entries
		if err := e.core.AccountTransactionDestroyer(context, tx, currentDate, userOrg.OrganizationID, *userOrg.BranchID); err != nil {
			return endTx(err)
		}

		// ----- DAILY BOOKING -----
		booking, err := e.core.DailyBookingCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return endTx(err)
		}
		totalDebit := decimal.Zero
		totalCredit := decimal.Zero
		accountTransactionCollection := &core.AccountTransaction{
			CreatedAt:      now,
			CreatedByID:    userOrg.ID,
			UpdatedAt:      now,
			UpdatedByID:    userOrg.ID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			JVNumber:       BuildJVNumberSimple(currentDate, GeneralJournalBook),
			Description:    "TOTAL DAILY COLLECTION",
			Date:           currentDate,
			Debit:          0,
			Credit:         0,
			Source:         core.AccountTransactionSourceDailyCollectionBook, // <-- here
		}
		if err := e.core.AccountTransactionManager().CreateWithTx(context, tx, accountTransactionCollection); err != nil {
			return endTx(err)
		}
		for _, summary := range usecase.SumGeneralLedgerByAccount(booking) {
			entry := &core.AccountTransactionEntry{
				CreatedAt:            now,
				CreatedByID:          userOrg.ID,
				UpdatedAt:            now,
				UpdatedByID:          userOrg.ID,
				OrganizationID:       userOrg.OrganizationID,
				BranchID:             *userOrg.BranchID,
				AccountTransactionID: accountTransactionCollection.ID,
				AccountID:            summary.AccountID,
				Debit:                summary.Debit,
				Credit:               summary.Credit,
			}
			if err := e.core.AccountTransactionEntryManager().CreateWithTx(context, tx, entry); err != nil {
				return endTx(err)
			}
			totalDebit = totalDebit.Add(decimal.NewFromFloat(summary.Debit))
			totalCredit = totalCredit.Add(decimal.NewFromFloat(summary.Credit))
		}
		accountTransactionCollection.Debit = totalDebit.InexactFloat64()
		accountTransactionCollection.Credit = totalCredit.InexactFloat64() // Fixed
		if err := e.core.AccountTransactionManager().UpdateByIDWithTx(context, tx, accountTransactionCollection.ID, accountTransactionCollection); err != nil {
			return endTx(err)
		}

		// ----- DAILY DISBURSEMENT -----
		disbursement, err := e.core.DailyDisbursementCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return endTx(err)
		}
		totalDebit, totalCredit = decimal.Zero, decimal.Zero
		disbursementTransaction := &core.AccountTransaction{
			CreatedAt:      now,
			CreatedByID:    userOrg.ID,
			UpdatedAt:      now,
			UpdatedByID:    userOrg.ID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			JVNumber:       BuildJVNumberSimple(currentDate, CashCheckDisbursementBook),
			Description:    "TOTAL CASH/CHECK DISBURSEMENT",
			Date:           currentDate,
			Debit:          0,
			Credit:         0,
			Source:         core.AccountTransactionSourceCashCheckDisbursementBook,
		}
		if err := e.core.AccountTransactionManager().CreateWithTx(context, tx, disbursementTransaction); err != nil {
			return endTx(err)
		}
		for _, summary := range usecase.SumGeneralLedgerByAccount(disbursement) {
			entry := &core.AccountTransactionEntry{
				CreatedAt:            now,
				CreatedByID:          userOrg.ID,
				UpdatedAt:            now,
				UpdatedByID:          userOrg.ID,
				OrganizationID:       userOrg.OrganizationID,
				BranchID:             *userOrg.BranchID,
				AccountTransactionID: disbursementTransaction.ID,
				AccountID:            summary.AccountID,
				Debit:                summary.Debit,
				Credit:               summary.Credit,
			}
			if err := e.core.AccountTransactionEntryManager().CreateWithTx(context, tx, entry); err != nil {
				return endTx(err)
			}
			totalDebit = totalDebit.Add(decimal.NewFromFloat(summary.Debit))
			totalCredit = totalCredit.Add(decimal.NewFromFloat(summary.Credit))
		}
		disbursementTransaction.Debit = totalDebit.InexactFloat64()
		disbursementTransaction.Credit = totalCredit.InexactFloat64()
		if err := e.core.AccountTransactionManager().UpdateByIDWithTx(context, tx, disbursementTransaction.ID, disbursementTransaction); err != nil {
			return endTx(err)
		}

		// ----- DAILY JOURNAL -----
		journal, err := e.core.DailyJournalCollection(context, currentDate, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return endTx(err)
		}
		totalDebit, totalCredit = decimal.Zero, decimal.Zero
		journalTransaction := &core.AccountTransaction{
			CreatedAt:      now,
			CreatedByID:    userOrg.ID,
			UpdatedAt:      now,
			UpdatedByID:    userOrg.ID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			JVNumber:       BuildJVNumberSimple(currentDate, GeneralJournalBook),
			Description:    "TOTAL GENERAL JOURNAL",
			Date:           currentDate,
			Debit:          0,
			Credit:         0,
			Source:         core.AccountTransactionSourceGeneralJournal,
		}
		if err := e.core.AccountTransactionManager().CreateWithTx(context, tx, journalTransaction); err != nil {
			return endTx(err)
		}
		for _, summary := range usecase.SumGeneralLedgerByAccount(journal) {
			entry := &core.AccountTransactionEntry{
				CreatedAt:            now,
				CreatedByID:          userOrg.ID,
				UpdatedAt:            now,
				UpdatedByID:          userOrg.ID,
				OrganizationID:       userOrg.OrganizationID,
				BranchID:             *userOrg.BranchID,
				AccountTransactionID: journalTransaction.ID,
				AccountID:            summary.AccountID,
				Debit:                summary.Debit,
				Credit:               summary.Credit,
			}
			if err := e.core.AccountTransactionEntryManager().CreateWithTx(context, tx, entry); err != nil {
				return endTx(err)
			}
			totalDebit = totalDebit.Add(decimal.NewFromFloat(summary.Debit))
			totalCredit = totalCredit.Add(decimal.NewFromFloat(summary.Credit))
		}
		journalTransaction.Debit = totalDebit.InexactFloat64()
		journalTransaction.Credit = totalCredit.InexactFloat64()
		if err := e.core.AccountTransactionManager().UpdateByIDWithTx(context, tx, journalTransaction.ID, journalTransaction); err != nil {
			return endTx(err)
		}
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
