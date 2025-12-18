package report

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Reports) loanTransactionReport(ctx context.Context, data ReportData) (result []byte, err error) {
	result, err = data.generated.PDF("/api/v1/loan-transaction/:loan_transaction_id", func(params ...string) ([]byte, error) {
		if len(params) == 0 || params[0] == "" {
			return nil, eris.New("missing loan transaction id in route params")
		}
		loanTransactionID, err := uuid.Parse(params[0])
		if err != nil {
			return nil, eris.Wrapf(err, "invalid loan transaction ID: %s", params[0])
		}

		loanTransaction, getErr := r.core.LoanTransactionManager.GetByID(
			ctx, loanTransactionID,
			"ReleasedBy", "MemberProfile",
			"Account.Currency",
		)
		if getErr != nil {
			return nil, eris.Wrapf(getErr, "Failed to get loan transaction by ID: %s", loanTransactionID)
		}

		if loanTransaction == nil {
			return nil, eris.Wrapf(nil, "loan transaction %s not found", loanTransactionID)
		}
		if loanTransaction.MemberProfileID == nil {
			return nil, eris.Wrapf(nil, "loan transaction %s has no member profile id", loanTransactionID)
		}

		branch, err := r.core.BranchManager.GetByID(ctx, loanTransaction.BranchID)
		if err != nil {
			return nil, eris.Wrapf(err, "Failed to get branch by ID: %s", loanTransaction.BranchID)
		}
		memberProfile, err := r.core.MemberProfileManager.GetByID(ctx, *loanTransaction.MemberProfileID)
		if err != nil {
			return nil, eris.Wrapf(err, "Failed to get member profile by ID: %s", loanTransaction.MemberProfileID)
		}
		loanTransactionEntries, err := r.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
			BranchID:          loanTransaction.BranchID,
			OrganizationID:    loanTransaction.OrganizationID,
			LoanTransactionID: loanTransaction.ID,
		}, "Account.Currency")
		if err != nil {
			return nil, eris.Wrapf(err, "Failed to find loan transaction entries for loan transaction ID: %s", loanTransaction.ID)
		}

		loan_transaction_entries := make([]map[string]any, 0)
		var total_debit float64
		var total_credit float64
		for _, entry := range loanTransactionEntries {
			var accountName string
			var debitFmt, creditFmt string
			if entry.Account != nil {
				accountName = entry.Account.Name
				if entry.Account.Currency != nil {
					debitFmt = entry.Account.Currency.ToFormat(entry.Debit)
					creditFmt = entry.Account.Currency.ToFormat(entry.Credit)
				}
			}
			loan_transaction_entries = append(loan_transaction_entries, map[string]any{
				"account_title": accountName,
				"debit":         debitFmt,
				"credit":        creditFmt,
			})
			total_debit += entry.Debit
			total_credit += entry.Credit
		}

		amount := total_credit - total_debit
		var processorName string
		if loanTransaction.ReleasedBy != nil {
			processorName = loanTransaction.ReleasedBy.FullName
		}
		var currencyAmountWords string
		currencyAmountWords = loanTransaction.Account.Currency.ToWords(amount)

		loanReleaseVoucher := map[string]any{
			"header_title":   branch.Name,
			"header_address": branch.Address,
			"tax_number":     branch.TaxIdentificationNumber,
			"report_title":   "Loan Release Voucher",

			"pay_to":          memberProfile.FullName,
			"address":         memberProfile.Address(),
			"contact":         memberProfile.ContactNumber,
			"voucher_no":      loanTransaction.Voucher,
			"date_release":    loanTransaction.ReadableReleaseDate(),
			"terms":           loanTransaction.Terms,
			"mode_of_payment": loanTransaction.ModeOfPayment,
			"processor":       processorName,
			"due_date":        loanTransaction.ReadableDueDate(),

			"loan_transaction_entries": loan_transaction_entries,

			"total_debit":           total_debit,
			"total_credit":          total_credit,
			"total_amount_in_words": currencyAmountWords,

			"prepared_by":          loanTransaction.PreparedByName,
			"payeee":               memberProfile.FullName,
			"cetified_correct":     loanTransaction.CertifiedByName,
			"paid_by":              "",
			"approved_for_payment": loanTransaction.ApprovedByName,
		}

		res, genErr := data.report.Generate(ctx, loanReleaseVoucher)
		if genErr != nil {
			return nil, genErr
		}
		return res, nil
	})

	return result, err
}
