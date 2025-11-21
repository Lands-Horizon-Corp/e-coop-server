package report

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Reports) loanTransactionReport(ctx context.Context, data ReportData) (result []byte, err error) {

	switch data.generated.GeneratedReportType {
	case core.GeneratedReportTypeExcel:
	case core.GeneratedReportTypePDF:
		fmt.Printf("[DEBUG] loanTransactionReport: entering PDF handler for URL=%s\n", data.generated.URL)
		result, err = data.extractor.MatchableRoute("/api/v1/loan-transaction/:loan_transaction_id", func(params ...string) ([]byte, error) {
			fmt.Printf("[DEBUG] handler: params=%#v\n", params)

			// validate params
			if len(params) == 0 || params[0] == "" {
				fmt.Printf("[DEBUG] handler: missing loan transaction id in route params\n")
				return nil, eris.New("missing loan transaction id in route params")
			}

			fmt.Printf("[DEBUG] handler: parsing uuid %s\n", params[0])
			loanTransactionID, err := uuid.Parse(params[0])
			if err != nil {
				fmt.Printf("[DEBUG] handler: uuid parse error: %v\n", err)
				return nil, eris.Wrapf(err, "invalid loan transaction ID: %s", params[0])
			}
			fmt.Printf("[DEBUG] handler: parsed loanTransactionID=%s\n", loanTransactionID)

			loanTransaction, getErr := r.core.LoanTransactionManager.GetByID(
				ctx, loanTransactionID,
				"ReleasedBy", "MemberProfile",
				"Account.Currency",
			)
			if getErr != nil {
				fmt.Printf("[DEBUG] handler: GetByID error: %v\n", getErr)
				return nil, eris.Wrapf(getErr, "Failed to get loan transaction by ID: %s", loanTransactionID)
			}
			fmt.Printf("[DEBUG] handler: fetched loanTransaction=%#v\n", loanTransaction)

			// ensure we have a member profile id before dereferencing
			if loanTransaction == nil {
				fmt.Printf("[DEBUG] handler: loanTransaction is nil -> returning error\n")
				return nil, eris.Wrapf(nil, "loan transaction %s not found", loanTransactionID)
			}
			if loanTransaction.MemberProfileID == nil {
				fmt.Printf("[DEBUG] handler: loanTransaction.MemberProfileID is nil -> returning error\n")
				return nil, eris.Wrapf(nil, "loan transaction %s has no member profile id", loanTransactionID)
			}

			branch, err := r.core.BranchManager.GetByID(ctx, loanTransaction.BranchID)
			if err != nil {
				fmt.Printf("[DEBUG] handler: BranchManager.GetByID error: %v\n", err)
				return nil, eris.Wrapf(err, "Failed to get branch by ID: %s", loanTransaction.BranchID)
			}
			fmt.Printf("[DEBUG] handler: fetched branch=%#v\n", branch)

			memberProfile, err := r.core.MemberProfileManager.GetByID(ctx, *loanTransaction.MemberProfileID)
			if err != nil {
				fmt.Printf("[DEBUG] handler: MemberProfileManager.GetByID error: %v\n", err)
				return nil, eris.Wrapf(err, "Failed to get member profile by ID: %s", loanTransaction.MemberProfileID)
			}
			fmt.Printf("[DEBUG] handler: fetched memberProfile=%#v\n", memberProfile)

			loanTransactionEntries, err := r.core.LoanTransactionEntryManager.Find(ctx, &core.LoanTransactionEntry{
				BranchID:          loanTransaction.BranchID,
				OrganizationID:    loanTransaction.OrganizationID,
				LoanTransactionID: loanTransaction.ID,
			}, "Account.Currency")
			if err != nil {
				fmt.Printf("[DEBUG] handler: LoanTransactionEntryManager.Find error: %v\n", err)
				return nil, eris.Wrapf(err, "Failed to find loan transaction entries for loan transaction ID: %s", loanTransaction.ID)
			}
			fmt.Printf("[DEBUG] handler: found %d entries\n", len(loanTransactionEntries))

			loan_transaction_entries := make([]map[string]any, 0)
			var total_debit float64
			var total_credit float64
			for _, entry := range loanTransactionEntries {
				// be defensive about nested fields
				var accountName string
				var debitFmt, creditFmt string
				if entry.Account != nil {
					accountName = entry.Account.Name
					if entry.Account.Currency != nil {
						debitFmt = entry.Account.Currency.FormatValue(entry.Debit)
						creditFmt = entry.Account.Currency.FormatValue(entry.Credit)
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
			fmt.Printf("[DEBUG] handler: totals computed debit=%f credit=%f amount=%f\n", total_debit, total_credit, amount)

			// defensive access for nested LoanTransaction fields used below
			var processorName string
			if loanTransaction.ReleasedBy != nil {
				processorName = loanTransaction.ReleasedBy.FullName
			}
			var currencyAmountWords string
			if loanTransaction.Account != nil && loanTransaction.Account.Currency != nil {
				currencyAmountWords = loanTransaction.Account.Currency.AmountInWordsSimple(amount)
			}

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

				"cash_on_hand_total_debit":  total_debit,
				"cash_on_hand_total_credit": total_credit,
				"total_amount_in_words":     currencyAmountWords,

				"prepared_by":          loanTransaction.PreparedByName,
				"payeee":               memberProfile.FullName,
				"cetified_correct":     loanTransaction.CertifiedByName,
				"paid_by":              "",
				"approved_for_payment": loanTransaction.ApprovedByName,
			}

			res, genErr := data.report.Generate(ctx, loanReleaseVoucher)
			if genErr != nil {
				fmt.Printf("[DEBUG] handler: report.Generate error: %v\n", genErr)
				return nil, genErr
			}
			fmt.Printf("[DEBUG] handler: report.Generate succeeded (bytes=%d)\n", len(res))
			return res, nil
		})
		fmt.Printf("[DEBUG] loanTransactionReport: MatchableRoute returned err=%v len(result)=%d\n", err, len(result))
	}
	return result, err
}
