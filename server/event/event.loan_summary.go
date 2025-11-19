package event

import (
	"context"
	"encoding/json"
	"fmt"
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
							totalAccountPrincipal = e.provider.Service.Decimal.Add(totalAccountPrincipal, accountValue.Value)

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

// MemberLoanSummary represents loan summary for a single member
type MemberLoanSummary struct {
	MemberProfileID   uuid.UUID  `json:"member_profile_id"`
	TotalLoans        int        `json:"total_loans"`
	TotalArrears      float64    `json:"total_arrears"`
	TotalPrincipal    float64    `json:"total_principal"`
	TotalPaid         float64    `json:"total_paid"`
	TotalRemaining    float64    `json:"total_remaining"`
	ActiveLoans       int        `json:"active_loans"`
	FullyPaidLoans    int        `json:"fully_paid_loans"`
	OverdueLoans      int        `json:"overdue_loans"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty"`
	LastPaymentAmount float64    `json:"last_payment_amount"`
}

// AllMembersLoanSummaryResponse represents loan summaries for all members
type AllMembersLoanSummaryResponse struct {
	MemberSummaries     []MemberLoanSummary `json:"member_summaries"`
	TotalMembers        int                 `json:"total_members"`
	TotalLoans          int                 `json:"total_loans"`
	TotalArrears        float64             `json:"total_arrears"`
	TotalPrincipal      float64             `json:"total_principal"`
	TotalPaid           float64             `json:"total_paid"`
	TotalRemaining      float64             `json:"total_remaining"`
	TotalActiveLoans    int                 `json:"total_active_loans"`
	TotalFullyPaidLoans int                 `json:"total_fully_paid_loans"`
	TotalOverdueLoans   int                 `json:"total_overdue_loans"`
	MembersWithLoans    int                 `json:"members_with_loans"`
	MembersWithOverdue  int                 `json:"members_with_overdue"`
	MembersFullyPaid    int                 `json:"members_fully_paid"`
	OrganizationID      uuid.UUID           `json:"organization_id"`
	BranchID            uuid.UUID           `json:"branch_id"`
	GeneratedAt         time.Time           `json:"generated_at"`
}

// AllMembersLoanSummary retrieves comprehensive loan summaries for all members in a branch/organization
func (e *Event) AllMembersLoanSummary(
	context context.Context,
	userOrg *core.UserOrganization,
) (*AllMembersLoanSummaryResponse, error) {
	// ===============================================================================================
	// STEP 1: VALIDATE INPUT PARAMETERS
	// ===============================================================================================
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}
	if userOrg.BranchID == nil {
		return nil, eris.New("branch ID is required")
	}

	// ===============================================================================================
	// STEP 1.5: CHECK CACHE
	// ===============================================================================================
	// Generate cache key based on organization, branch, and current date (daily cache)
	currentDate := userOrg.UserOrgTime().Format("2006-01-02")
	cacheKey := fmt.Sprintf("loan_summary:all_members:%s:%s:%s",
		userOrg.OrganizationID.String(),
		userOrg.BranchID.String(),
		currentDate,
	)

	// Try to get from cache
	if e.provider.Service.Cache != nil {
		cachedData, err := e.provider.Service.Cache.Get(context, cacheKey)
		if err == nil && cachedData != nil {
			var cachedResponse AllMembersLoanSummaryResponse
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				return &cachedResponse, nil
			}
		}
	}

	// ===============================================================================================
	// STEP 2: FETCH ALL LOAN TRANSACTIONS FOR THE BRANCH
	// ===============================================================================================
	allLoanTransactions, err := e.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve loan transactions")
	}

	// ===============================================================================================
	// STEP 3: GROUP LOAN TRANSACTIONS BY MEMBER PROFILE
	// ===============================================================================================
	memberLoanMap := make(map[uuid.UUID][]*core.LoanTransaction)
	for _, loan := range allLoanTransactions {
		if loan.MemberProfileID != nil {
			memberLoanMap[*loan.MemberProfileID] = append(memberLoanMap[*loan.MemberProfileID], loan)
		}
	}

	// ===============================================================================================
	// STEP 4: INITIALIZE AGGREGATION VARIABLES
	// ===============================================================================================
	memberSummaries := []MemberLoanSummary{}
	totalMembers := len(memberLoanMap)
	totalLoans := 0
	totalArrears := 0.0
	totalPrincipal := 0.0
	totalPaid := 0.0
	totalRemaining := 0.0
	totalActiveLoans := 0
	totalFullyPaidLoans := 0
	totalOverdueLoans := 0
	membersWithLoans := 0
	membersWithOverdue := 0
	membersFullyPaid := 0

	// ===============================================================================================
	// STEP 5: PROCESS EACH MEMBER'S LOANS
	// ===============================================================================================
	for memberProfileID, loans := range memberLoanMap {
		// -------------------------------------------------------------------------------------------
		// 5.1: Initialize Member-Level Aggregates
		// -------------------------------------------------------------------------------------------
		memberTotalArrears := 0.0
		memberTotalPrincipal := 0.0
		memberTotalPaid := 0.0
		memberTotalRemaining := 0.0
		memberActiveLoans := 0
		memberFullyPaidLoans := 0
		memberOverdueLoans := 0
		var memberLastPaymentDate *time.Time
		memberLastPaymentAmount := 0.0

		// -------------------------------------------------------------------------------------------
		// 5.2: Process Each Loan Transaction for This Member
		// -------------------------------------------------------------------------------------------
		for _, loan := range loans {
			// Get detailed loan summary
			loanSummary, err := e.LoanSummary(context, &loan.ID, userOrg)
			if err != nil {
				// Log error but continue processing other loans
				continue
			}

			// Aggregate member-level metrics
			memberTotalArrears = e.provider.Service.Decimal.Add(memberTotalArrears, loanSummary.Arrears)
			memberTotalPrincipal = e.provider.Service.Decimal.Add(memberTotalPrincipal, loanSummary.TotalPrincipal)
			memberTotalPaid = e.provider.Service.Decimal.Add(memberTotalPaid, loanSummary.TotalPrincipalPaid)
			memberTotalRemaining = e.provider.Service.Decimal.Add(memberTotalRemaining, loanSummary.TotalRemainingPrincipal)

			// Track loan status
			if loanSummary.TotalRemainingPrincipal > 0.01 {
				memberActiveLoans++
			} else {
				memberFullyPaidLoans++
			}

			// Track overdue status (arrears > 0)
			if loanSummary.Arrears > 0.01 {
				memberOverdueLoans++
			}

			// Track latest payment date
			if loanSummary.LastPayment != nil {
				if memberLastPaymentDate == nil || loanSummary.LastPayment.After(*memberLastPaymentDate) {
					memberLastPaymentDate = loanSummary.LastPayment
					// Estimate payment amount from arrears change (simplified)
					memberLastPaymentAmount = loanSummary.TotalPrincipalPaid
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 5.3: Fetch Member Profile Details
		// -------------------------------------------------------------------------------------------

		// -------------------------------------------------------------------------------------------
		// 5.4: Build Member Summary
		// -------------------------------------------------------------------------------------------
		memberSummary := MemberLoanSummary{
			MemberProfileID:   memberProfileID,
			TotalLoans:        len(loans),
			TotalArrears:      memberTotalArrears,
			TotalPrincipal:    memberTotalPrincipal,
			TotalPaid:         memberTotalPaid,
			TotalRemaining:    memberTotalRemaining,
			ActiveLoans:       memberActiveLoans,
			FullyPaidLoans:    memberFullyPaidLoans,
			OverdueLoans:      memberOverdueLoans,
			LastPaymentDate:   memberLastPaymentDate,
			LastPaymentAmount: memberLastPaymentAmount,
		}
		memberSummaries = append(memberSummaries, memberSummary)

		// -------------------------------------------------------------------------------------------
		// 5.5: Aggregate Organization-Level Totals
		// -------------------------------------------------------------------------------------------
		totalLoans += len(loans)
		totalArrears = e.provider.Service.Decimal.Add(totalArrears, memberTotalArrears)
		totalPrincipal = e.provider.Service.Decimal.Add(totalPrincipal, memberTotalPrincipal)
		totalPaid = e.provider.Service.Decimal.Add(totalPaid, memberTotalPaid)
		totalRemaining = e.provider.Service.Decimal.Add(totalRemaining, memberTotalRemaining)
		totalActiveLoans += memberActiveLoans
		totalFullyPaidLoans += memberFullyPaidLoans
		totalOverdueLoans += memberOverdueLoans

		if len(loans) > 0 {
			membersWithLoans++
		}
		if memberOverdueLoans > 0 {
			membersWithOverdue++
		}
		if memberActiveLoans == 0 && len(loans) > 0 {
			membersFullyPaid++
		}
	}

	// ===============================================================================================
	// STEP 6: BUILD RESPONSE
	// ===============================================================================================
	response := &AllMembersLoanSummaryResponse{
		MemberSummaries:     memberSummaries,
		TotalMembers:        totalMembers,
		TotalLoans:          totalLoans,
		TotalArrears:        totalArrears,
		TotalPrincipal:      totalPrincipal,
		TotalPaid:           totalPaid,
		TotalRemaining:      totalRemaining,
		TotalActiveLoans:    totalActiveLoans,
		TotalFullyPaidLoans: totalFullyPaidLoans,
		TotalOverdueLoans:   totalOverdueLoans,
		MembersWithLoans:    membersWithLoans,
		MembersWithOverdue:  membersWithOverdue,
		MembersFullyPaid:    membersFullyPaid,
		OrganizationID:      userOrg.OrganizationID,
		BranchID:            *userOrg.BranchID,
		GeneratedAt:         userOrg.UserOrgTime(),
	}

	// ===============================================================================================
	// STEP 7: CACHE THE RESPONSE
	// ===============================================================================================
	if e.provider.Service.Cache != nil {
		responseData, err := json.Marshal(response)
		if err == nil {
			// Cache for 24 hours
			_ = e.provider.Service.Cache.Set(context, cacheKey, responseData, 24*time.Hour)
		}
	}

	return response, nil
}
