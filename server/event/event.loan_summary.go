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

type LoanAccountSummaryResponse struct {
	AccountHistoryID uuid.UUID                   `json:"account_history_id"`
	AccountHistory   core.AccountHistoryResponse `json:"account_history"`

	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
	Balance     float64 `json:"balance"`

	DueDate     *time.Time `json:"due_date,omitempty"`
	LastPayment *time.Time `json:"last_payment,omitempty"`

	TotalNumberOfPayments int `json:"total_number_of_payments"`

	TotalNumberOfDeductions int     `json:"total_number_of_deductions"`
	TotalDeductions         float64 `json:"total_deductions"`
	TotalNumberOfAdditions  int     `json:"total_number_of_additions"`
	TotalAdditions          float64 `json:"total_additions"`

	TotalAccountPrincipal          float64   `json:"total_account_principal"`
	TotalAccountAdvancedPayment    float64   `json:"total_account_advanced_payment"`
	TotalAccountPrincipalPaid      float64   `json:"total_account_principal_paid"`
	TotalAccountRemainingPrincipal float64   `json:"total_remaining_principal"`
	LoanTransactionID              uuid.UUID `json:"loan_transaction_id"`
}

type LoanTransactionSummaryResponse struct {
	LoanTransactionID uuid.UUID                     `json:"loan_transaction_id"`
	AmountGranted     float64                       `json:"amount_granted"`
	AddOnAmount       float64                       `json:"add_on_amount"`
	AccountSummary    []LoanAccountSummaryResponse  `json:"account_summary"`
	GeneralLedger     []*core.GeneralLedgerResponse `json:"general_ledger"`

	Arrears float64 `json:"arrears"`

	LastPayment             *time.Time `json:"last_payment,omitempty"`
	FirstDeliquencyDate     *time.Time `json:"first_deliquency_date,omitempty"`
	FirstIrregularityDate   *time.Time `json:"first_irregularity_date,omitempty"`
	TotalPrincipal          float64    `json:"total_principal"`
	TotalAdvancedPayment    float64    `json:"total_advanced_payment"`
	TotalPrincipalPaid      float64    `json:"total_principal_paid"`
	TotalRemainingPrincipal float64    `json:"total_remaining_principal"`
}

func (e *Event) LoanSummary(
	context context.Context, loanTransactionID *uuid.UUID, userOrg *core.UserOrganization,
) (*LoanTransactionSummaryResponse, error) {
	// ===============================================================================================
	// STEP 1: VALIDATE INPUT PARAMETERS
	// ===============================================================================================
	if loanTransactionID == nil {
		return nil, eris.New("loan transaction id is required")
	}
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}

	// ===============================================================================================
	// STEP 2: FETCH LOAN TRANSACTION DATA
	// ===============================================================================================
	// Retrieve the main loan transaction record
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, eris.Wrapf(err, "loan transaction not found: %s", *loanTransactionID)
	}

	// ===============================================================================================
	// STEP 3: FETCH GENERAL LEDGER ENTRIES
	// ===============================================================================================
	// Get all general ledger entries associated with this loan transaction
	// These entries contain all debit/credit transactions for the loan
	entries, err := e.core.GeneralLedgerByLoanTransaction(
		context,
		*loanTransactionID,
		userOrg.OrganizationID,
		*userOrg.BranchID,
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve general ledger entries for loan transaction id: %s", *loanTransactionID)
	}

	// ===============================================================================================
	// STEP 4: FETCH RELATED ACCOUNTS
	// ===============================================================================================
	// Retrieve all accounts linked to this loan (principal, interest, fees, etc.)
	// LoanAccountID points to the main loan account, and this finds all sub-accounts
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
	// STEP 5: FETCH LOAN AMORTIZATION SCHEDULE
	// ===============================================================================================
	// Get the complete payment schedule to calculate principal amounts and due dates
	// This schedule contains the planned payment breakdown by period
	amortizationSchedule, err := e.LoanAmortizationSchedule(context, *loanTransactionID, userOrg)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve amortization schedule for loan transaction id: %s", *loanTransactionID)
	}

	// ===============================================================================================
	// STEP 6: INITIALIZE SUMMARY VARIABLES
	// ===============================================================================================
	arrears := 0.0 // Total outstanding balance across all accounts
	var lastPayment *time.Time
	accountsummary := []LoanAccountSummaryResponse{}

	// ===============================================================================================
	// STEP 7: PROCESS EACH ACCOUNT AND CALCULATE SUMMARIES
	// ===============================================================================================
	for _, entry := range accounts {
		// -------------------------------------------------------------------------------------------
		// 7.1: Fetch Account History
		// -------------------------------------------------------------------------------------------
		// Get the historical snapshot of account settings at the time of loan printing
		accountHistory, err := e.core.GetAccountHistoryLatestByTimeHistory(
			context,
			entry.ID,
			entry.OrganizationID,
			entry.BranchID,
			loanTransaction.PrintedDate,
		)
		if err != nil {
			return nil, eris.Wrapf(err, "failed to retrieve account history: %s", err.Error())
		}

		// -------------------------------------------------------------------------------------------
		// 7.2: Calculate Account Balance
		// -------------------------------------------------------------------------------------------
		// Compute total debits, credits, and running balance for this specific account
		balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: entries,
			AccountID:      &entry.ID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for account id: %s", entry.ID)
		}

		// Add this account's balance to the total arrears
		arrears = e.provider.Service.Decimal.Add(arrears, balance.Balance)

		// Track the most recent payment date across all accounts
		if balance.LastPayment != nil && (lastPayment == nil || balance.LastPayment.After(*lastPayment)) {
			lastPayment = balance.LastPayment
		}

		// -------------------------------------------------------------------------------------------
		// 7.3: Calculate Principal Amounts from Amortization Schedule
		// -------------------------------------------------------------------------------------------
		totalAccountPrincipal := 0.0     // Total scheduled principal for this account
		totalAccountPrincipalPaid := 0.0 // Actual principal paid so far
		var dueDate *time.Time           // Next payment due date

		if amortizationSchedule != nil && amortizationSchedule.Schedule != nil {
			now := userOrg.UserOrgTime()

			// Loop through each scheduled payment period
			for _, schedule := range amortizationSchedule.Schedule {
				if schedule.Accounts != nil {
					// Find this account in the schedule's account breakdown
					for _, accountValue := range schedule.Accounts {
						if accountValue.Account != nil && handlers.UUIDPtrEqual(&accountValue.Account.ID, &entry.ID) {
							// Accumulate total scheduled principal (cumulative total from amortization)
							totalAccountPrincipal = e.provider.Service.Decimal.Add(totalAccountPrincipal, accountValue.Total)

							// Capture the first future scheduled date as the due date
							if dueDate == nil && schedule.ScheduledDate.After(now) {
								scheduledDate := schedule.ScheduledDate
								dueDate = &scheduledDate
							}
						}
					}
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 7.4: Calculate Principal Paid
		// -------------------------------------------------------------------------------------------
		// For loan-type accounts, credits represent principal payments
		if accountHistory.Type == core.AccountTypeLoan {
			totalAccountPrincipalPaid = balance.Credit
		}

		// -------------------------------------------------------------------------------------------
		// 7.5: Calculate Remaining Principal
		// -------------------------------------------------------------------------------------------
		// Remaining principal = Total scheduled - Amount paid
		totalAccountRemainingPrincipal := e.provider.Service.Decimal.Subtract(totalAccountPrincipal, totalAccountPrincipalPaid)

		// -------------------------------------------------------------------------------------------
		// 7.6: Calculate Advanced Payments
		// -------------------------------------------------------------------------------------------
		// Advanced payment occurs when paid amount exceeds scheduled amount
		totalAccountAdvancedPayment := 0.0
		if totalAccountPrincipalPaid > totalAccountPrincipal {
			totalAccountAdvancedPayment = e.provider.Service.Decimal.Subtract(totalAccountPrincipalPaid, totalAccountPrincipal)
		}

		// -------------------------------------------------------------------------------------------
		// 7.7: Build Account Summary Response
		// -------------------------------------------------------------------------------------------
		accountsummary = append(accountsummary, LoanAccountSummaryResponse{
			AccountHistoryID:        accountHistory.ID,
			AccountHistory:          *e.core.AccountHistoryManager.ToModel(accountHistory),
			TotalDebit:              balance.Debit,
			TotalCredit:             balance.Credit,
			Balance:                 balance.Balance,
			TotalNumberOfDeductions: balance.CountDeductions,
			TotalDeductions:         balance.Deductions,
			TotalNumberOfAdditions:  balance.CountAdded,
			TotalAdditions:          balance.Added,
			TotalNumberOfPayments:   balance.CountCredit,
			LoanTransactionID:       loanTransaction.ID,
			LastPayment:             balance.LastPayment,

			DueDate:                        dueDate,
			TotalAccountPrincipal:          totalAccountPrincipal,
			TotalAccountAdvancedPayment:    totalAccountAdvancedPayment,
			TotalAccountPrincipalPaid:      totalAccountPrincipalPaid,
			TotalAccountRemainingPrincipal: totalAccountRemainingPrincipal,
		})
	}

	// ===============================================================================================
	// STEP 8: FETCH LOAN TRANSACTION ENTRIES
	// ===============================================================================================
	// Retrieve loan transaction entries for processing add-on amounts and automatic deductions
	loanTransactionEntries, err := e.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
		LoanTransactionID: loanTransaction.ID,
		OrganizationID:    loanTransaction.OrganizationID,
		BranchID:          loanTransaction.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve loan transaction entries")
	}

	// ===============================================================================================
	// STEP 9: CALCULATE LOAN BALANCE INCLUDING ADD-ON AMOUNTS
	// ===============================================================================================
	loanBalance, err := e.usecase.Balance(usecase.Balance{
		LoanTransactionEntries: loanTransactionEntries,
		IsAddOn:                loanTransaction.IsAddOn,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to compute loan balance from transaction entries")
	}

	// ===============================================================================================
	// STEP 10: AGGREGATE TRANSACTION-LEVEL TOTALS
	// ===============================================================================================
	// Sum up all account-level values to get transaction-wide totals
	totalPrincipal := 0.0          // Total principal across all accounts
	totalAdvancedPayment := 0.0    // Total advanced payments across all accounts
	totalPrincipalPaid := 0.0      // Total principal paid across all accounts
	totalRemainingPrincipal := 0.0 // Total remaining principal across all accounts

	for _, accSummary := range accountsummary {
		totalPrincipal = e.provider.Service.Decimal.Add(totalPrincipal, accSummary.TotalAccountPrincipal)
		totalAdvancedPayment = e.provider.Service.Decimal.Add(totalAdvancedPayment, accSummary.TotalAccountAdvancedPayment)
		totalPrincipalPaid = e.provider.Service.Decimal.Add(totalPrincipalPaid, accSummary.TotalAccountPrincipalPaid)
		totalRemainingPrincipal = e.provider.Service.Decimal.Add(totalRemainingPrincipal, accSummary.TotalAccountRemainingPrincipal)
	}

	// ===============================================================================================
	// STEP 11: BUILD AND RETURN FINAL SUMMARY RESPONSE
	// ===============================================================================================
	return &LoanTransactionSummaryResponse{
		GeneralLedger: e.core.GeneralLedgerManager.ToModels(
			e.usecase.GeneralLedgerAddBalanceByAccount(entries),
		),
		AccountSummary:          accountsummary,
		Arrears:                 arrears,
		AmountGranted:           loanTransaction.Applied1,
		LoanTransactionID:       loanTransaction.ID,
		LastPayment:             lastPayment,
		AddOnAmount:             loanBalance.AddOnAmount,
		TotalPrincipal:          totalPrincipal,
		TotalAdvancedPayment:    totalAdvancedPayment,
		TotalPrincipalPaid:      totalPrincipalPaid,
		TotalRemainingPrincipal: totalRemainingPrincipal,
	}, nil
}

func (e *Event) LoanTotalMemberProfile(context context.Context, memberProfileID uuid.UUID) (*float64, error) {
	memberProfile, err := e.core.MemberProfileManager.GetByID(context, memberProfileID)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to get member profile by id: %s", memberProfileID)
	}
	loanTransactions, err := e.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
		MemberProfileID: &memberProfile.ID,
		OrganizationID:  memberProfile.OrganizationID,
		BranchID:        memberProfile.BranchID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to find loan transactions for member profile id: %s", memberProfileID)
	}

	total := 0.0
	for _, loanTransaction := range loanTransactions {
		generalLedgers, err := e.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
			AccountID:      loanTransaction.AccountID,
			OrganizationID: memberProfile.OrganizationID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to get latest general ledger for loan account id: %s", *loanTransaction.AccountID)
		}
		balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: generalLedgers,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for loan account id: %s", *loanTransaction.AccountID)
		}
		total = e.provider.Service.Decimal.Add(total, balance.Balance)
	}
	return &total, nil
}
