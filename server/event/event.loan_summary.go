package event

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
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

	TotalNumberOfDeductions int `json:"total_number_of_deductions"`
	TotalNumberOfAdditions  int `json:"total_number_of_additions"`

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
	if loanTransactionID == nil {
		return nil, eris.New("loan transaction id is required")
	}
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}
	loanTransaction, err := e.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
	if err != nil {
		return nil, eris.Wrapf(err, "loan transaction not found: %s", *loanTransactionID)
	}

	entries, err := e.core.GeneralLedgerByLoanTransaction(
		context,
		*loanTransactionID,
		userOrg.OrganizationID,
		*userOrg.BranchID,
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve general ledger entries for loan transaction id: %s", *loanTransactionID)
	}

	accounts, err := e.core.AccountManager.Find(context, &core.Account{
		BranchID:       *userOrg.BranchID,
		OrganizationID: userOrg.OrganizationID,
		LoanAccountID:  loanTransaction.AccountID,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to retrieve accounts for loan transaction id: %s", *loanTransactionID)
	}

	arrears := 0.0
	accountsummary := []LoanAccountSummaryResponse{}
	for _, entry := range accounts {
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
		credit, debit, balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: entries,
			AccountID:      &entry.ID,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for account id: %s", entry.ID)
		}
		accountsummary = append(accountsummary, LoanAccountSummaryResponse{
			AccountHistoryID: accountHistory.ID,
			AccountHistory:   *e.core.AccountHistoryManager.ToModel(accountHistory),

			TotalDebit:  debit,
			TotalCredit: credit,
			Balance:     balance,

			TotalNumberOfDeductions: 0,
			TotalNumberOfAdditions:  0,

			DueDate:               nil,
			LastPayment:           nil,
			TotalNumberOfPayments: 0,

			TotalAccountPrincipal:          0,
			TotalAccountAdvancedPayment:    0,
			TotalAccountPrincipalPaid:      0,
			TotalAccountRemainingPrincipal: 0,
			LoanTransactionID:              loanTransaction.ID,
		})
	}
	return &LoanTransactionSummaryResponse{
		GeneralLedger: e.core.GeneralLedgerManager.ToModels(
			e.usecase.GeneralLedgerAddBalanceByAccount(entries),
		),
		AccountSummary:    accountsummary,
		Arrears:           arrears,
		AmountGranted:     loanTransaction.Applied1,
		LoanTransactionID: loanTransaction.ID,
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
		_, _, balance, err := e.usecase.Balance(usecase.Balance{
			GeneralLedgers: generalLedgers,
		})
		if err != nil {
			return nil, eris.Wrapf(err, "failed to compute balance for loan account id: %s", *loanTransaction.AccountID)
		}
		total = e.provider.Service.Decimal.Add(total, balance)
	}
	return &total, nil
}
