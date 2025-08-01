package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// calculateAccountBalances handles negative amounts and calculates debit/credit based on account type
func (ts *TransactionService) calculateAccountBalances(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	originalAmount float64,
	ledgerSource model.GeneralLedgerSource,
	account *model.Account,
	generalLedger *model.GeneralLedger,
	block func(string),
) (credit, debit, newBalance float64, effectiveSource model.GeneralLedgerSource, processedAmount float64, isNegative bool, err error) {
	// Handle negative amounts by flipping ledger source and making amount positive
	amount := originalAmount
	effectiveSource = ledgerSource
	var negativeAmountReason string

	if amount < 0 {
		amount = -amount // Make amount positive for calculations
		isNegative = true
		negativeAmountReason = fmt.Sprintf("Negative amount %.2f converted to positive %.2f", originalAmount, amount)

		// Flip the ledger source when amount is negative
		if ledgerSource == model.GeneralLedgerSourceDeposit {
			effectiveSource = model.GeneralLedgerSourceWithdraw
			negativeAmountReason += " and ledger source changed from deposit to withdraw"
		} else if ledgerSource == model.GeneralLedgerSourceWithdraw {
			effectiveSource = model.GeneralLedgerSourceDeposit
			negativeAmountReason += " and ledger source changed from withdraw to deposit"
		}

		// Log the negative amount handling
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "negative-amount-handled",
			Description: negativeAmountReason,
			Module:      "TransactionService",
		})
	}

	processedAmount = amount

	switch account.Type {
	case model.AccountTypeDeposit:
		// Deposit accounts: Credits increase balance, Debits decrease balance
		if effectiveSource == model.GeneralLedgerSourceDeposit {
			credit = amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + amount
			} else {
				newBalance = amount
			}
		} else { // Withdraw
			// Check sufficient balance for withdrawal
			if generalLedger == nil || generalLedger.Balance < amount {
				tx.Rollback()
				reason := fmt.Sprintf("Insufficient balance for withdrawal. Available: %.2f, Required: %.2f",
					func() float64 {
						if generalLedger != nil {
							return generalLedger.Balance
						} else {
							return 0
						}
					}(), amount)
				block(reason)
				ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
					Activity:    "balance-error",
					Description: reason,
					Module:      "TransactionService",
				})
				return 0, 0, 0, effectiveSource, 0, isNegative, eris.New("insufficient balance for withdrawal")
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	case model.AccountTypeLoan:
		// Loan accounts: Deposits reduce loan balance (debit), Withdrawals increase loan balance (credit)
		if effectiveSource == model.GeneralLedgerSourceDeposit {
			credit = 0
			debit = amount // Payment reduces loan balance
			if generalLedger != nil {
				newBalance = generalLedger.Balance - amount
				// Prevent overpayment - loan balance shouldn't go below 0
				if newBalance < 0 {
					tx.Rollback()
					reason := fmt.Sprintf("Loan overpayment not allowed. Current balance: %.2f, Payment: %.2f", generalLedger.Balance, amount)
					block(reason)
					ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
						Activity:    "loan-overpayment-error",
						Description: reason,
						Module:      "TransactionService",
					})
					return 0, 0, 0, effectiveSource, 0, isNegative, eris.New("payment exceeds loan balance")
				}
			} else {
				newBalance = -amount
			}
		} else {
			credit = amount // Loan disbursement increases loan balance
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + amount
			} else {
				newBalance = amount
			}
		}

	case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
		// Receivable/Payable accounts: Deposits reduce balance (debit), Withdrawals increase balance (credit)
		if effectiveSource == model.GeneralLedgerSourceDeposit {
			credit = 0
			debit = amount // Payment reduces receivable balance
			if generalLedger != nil {
				newBalance = generalLedger.Balance - amount
			} else {
				newBalance = -amount
			}
		} else {
			credit = amount // Charge increases receivable balance
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + amount
			} else {
				newBalance = amount
			}
		}

	case model.AccountTypeOther:
		// Other accounts: Similar to deposit accounts
		if effectiveSource == model.GeneralLedgerSourceDeposit {
			credit = amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + amount
			} else {
				newBalance = amount
			}
		} else {
			// Check sufficient balance for withdrawal
			if generalLedger == nil || generalLedger.Balance < amount {
				tx.Rollback()
				reason := fmt.Sprintf("Insufficient balance for withdrawal from Other account. Available: %.2f, Required: %.2f",
					func() float64 {
						if generalLedger != nil {
							return generalLedger.Balance
						} else {
							return 0
						}
					}(), amount)
				block(reason)
				ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
					Activity:    "balance-error",
					Description: reason,
					Module:      "TransactionService",
				})
				return 0, 0, 0, effectiveSource, 0, isNegative, eris.New("insufficient balance for withdrawal from Other account")
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	default:
		// Unsupported account type
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-type-error",
			Description: fmt.Sprintf("Unsupported account type: %s", account.Type),
			Module:      "TransactionService",
		})
		tx.Rollback()
		reason := fmt.Sprintf("Unsupported account type: %s", account.Type)
		block(reason)
		return 0, 0, 0, effectiveSource, 0, isNegative, eris.New("unsupported account type")
	}

	// Check account balance limits
	if (account.MinAmount != 0 || account.MaxAmount != 0) &&
		(newBalance < account.MinAmount || newBalance > account.MaxAmount) {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity: "balance-limit-error",
			Description: fmt.Sprintf(
				"Balance %.2f exceeds account limits [%.2f-%.2f]",
				newBalance,
				account.MinAmount,
				account.MaxAmount,
			),
			Module: "TransactionService",
		})
		block(fmt.Sprintf("Account balance limits exceeded: %.2f not in [%.2f-%.2f]", newBalance, account.MinAmount, account.MaxAmount))
		return 0, 0, 0, effectiveSource, 0, isNegative, eris.New("account balance limits exceeded")
	}

	return credit, debit, newBalance, effectiveSource, processedAmount, isNegative, nil
}

// handleTransactionAndMember handles transaction creation/retrieval and determines effective member profile ID
func (ts *TransactionService) handleTransactionAndMember(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	transactionId *uuid.UUID,
	data *model.PaymentQuickRequest,
	userOrg *model.UserOrganization,
	transactionBatch *model.TransactionBatch,
	block func(string),
) (*model.Transaction, *uuid.UUID, error) {
	var transaction *model.Transaction
	var effectiveMemberProfileID *uuid.UUID

	if transactionId != nil {
		// Try to retrieve existing transaction
		existingTransaction, err := ts.model.TransactionManager.GetByID(context, *transactionId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new transaction since it doesn't exist
				transaction, effectiveMemberProfileID = ts.createNewTransaction(data, userOrg, transactionBatch)
				if err := ts.model.TransactionManager.CreateWithTx(context, tx, transaction); err != nil {
					tx.Rollback()
					ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
						Activity:    "transaction-create-error",
						Description: "Failed to create transaction: " + err.Error(),
						Module:      "TransactionService",
					})
					block("Failed to create transaction: " + err.Error())
					return nil, nil, eris.Wrap(err, "failed to create transaction")
				}
			} else {
				// Actual database error occurred
				tx.Rollback()
				ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
					Activity:    "transaction-error",
					Description: "Failed to retrieve transaction: " + err.Error(),
					Module:      "TransactionService",
				})
				block("Failed to retrieve transaction: " + err.Error())
				return nil, nil, eris.Wrap(err, "failed to retrieve transaction")
			}
		} else {
			// Transaction exists - validate and determine member
			if err := ts.validateExistingTransaction(existingTransaction, transactionBatch, ctx, tx, block); err != nil {
				return nil, nil, err
			}

			transaction = existingTransaction
			effectiveMemberProfileID, err = ts.determineEffectiveMemberID(context, ctx, tx, data, transaction, userOrg, block)
			if err != nil {
				return nil, nil, err
			}
		}
	} else {
		// Create new transaction
		transaction, effectiveMemberProfileID = ts.createNewTransaction(data, userOrg, transactionBatch)
		if err := ts.model.TransactionManager.CreateWithTx(context, tx, transaction); err != nil {
			tx.Rollback()
			ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
				Activity:    "transaction-create-error",
				Description: "Failed to create transaction: " + err.Error(),
				Module:      "TransactionService",
			})
			block("Failed to create transaction: " + err.Error())
			return nil, nil, eris.Wrap(err, "failed to create transaction")
		}
	}

	return transaction, effectiveMemberProfileID, nil
}

// createNewTransaction creates a new transaction with the provided data
func (ts *TransactionService) createNewTransaction(
	data *model.PaymentQuickRequest,
	userOrg *model.UserOrganization,
	transactionBatch *model.TransactionBatch,
) (*model.Transaction, *uuid.UUID) {
	effectiveMemberProfileID := data.MemberProfileID

	transaction := &model.Transaction{
		CreatedAt:            time.Now().UTC(),
		CreatedByID:          userOrg.UserID,
		UpdatedAt:            time.Now().UTC(),
		UpdatedByID:          userOrg.UserID,
		BranchID:             *userOrg.BranchID,
		OrganizationID:       userOrg.OrganizationID,
		SignatureMediaID:     data.SignatureMediaID,
		TransactionBatchID:   &transactionBatch.ID,
		EmployeeUserID:       &userOrg.UserID,
		MemberProfileID:      effectiveMemberProfileID,
		MemberJointAccountID: data.MemberJointAccountID,
		Amount:               data.Amount,
		ReferenceNumber:      data.ReferenceNumber,
		Source:               model.GeneralLedgerSourceDeposit,
		Description:          data.Description,
		LoanBalance:          0,
		LoanDue:              0,
		TotalDue:             0,
		FinesDue:             0,
		TotalLoan:            0,
		InterestDue:          0,
	}

	return transaction, effectiveMemberProfileID
}

// validateExistingTransaction validates that the existing transaction belongs to current batch
func (ts *TransactionService) validateExistingTransaction(
	transaction *model.Transaction,
	transactionBatch *model.TransactionBatch,
	ctx echo.Context,
	tx *gorm.DB,
	block func(string),
) error {
	if transaction.TransactionBatchID == nil || *transaction.TransactionBatchID != transactionBatch.ID {
		tx.Rollback()
		ts.footstepLogger.Footstep(context.Background(), ctx, FootstepEvent{
			Activity:    "transaction-batch-mismatch",
			Description: "Transaction does not belong to the current transaction batch",
			Module:      "TransactionService",
		})
		block("Transaction batch mismatch")
		return eris.New("transaction does not belong to the current transaction batch")
	}
	return nil
}

// determineEffectiveMemberID determines the effective member profile ID from data or transaction
func (ts *TransactionService) determineEffectiveMemberID(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	data *model.PaymentQuickRequest,
	transaction *model.Transaction,
	userOrg *model.UserOrganization,
	block func(string),
) (*uuid.UUID, error) {
	// Priority: data member ID > transaction member ID > nil (subsidiary)
	if data.MemberProfileID != nil {
		return data.MemberProfileID, nil
	}

	if transaction.MemberProfileID != nil {
		// Validate member from transaction for security
		if err := ts.validateTransactionMember(context, ctx, tx, *transaction.MemberProfileID, userOrg, block); err != nil {
			return nil, err
		}

		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity: "member-from-transaction",
			Description: fmt.Sprintf("Using member profile ID %s from existing transaction",
				transaction.MemberProfileID.String()),
			Module: "TransactionService",
		})

		return transaction.MemberProfileID, nil
	}

	// No member - subsidiary ledger
	return nil, nil
}

// validateTransactionMember validates member profile from transaction
func (ts *TransactionService) validateTransactionMember(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	memberProfileID uuid.UUID,
	userOrg *model.UserOrganization,
	block func(string),
) error {
	memberProfile, err := ts.model.MemberProfileManager.GetByID(context, memberProfileID)
	if err != nil {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-member-error",
			Description: "Failed to retrieve member profile from transaction: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to retrieve member profile from transaction: " + err.Error())
		return eris.Wrap(err, "failed to retrieve member profile from transaction")
	}

	if memberProfile.BranchID != *userOrg.BranchID {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-member-branch-mismatch",
			Description: "Transaction member does not belong to the current branch",
			Module:      "TransactionService",
		})
		block("Transaction member does not belong to the current branch")
		return eris.New("transaction member does not belong to the current branch")
	}

	if memberProfile.OrganizationID != userOrg.OrganizationID {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-member-organization-mismatch",
			Description: "Transaction member does not belong to the current organization",
			Module:      "TransactionService",
		})
		block("Transaction member does not belong to the current organization")
		return eris.New("transaction member does not belong to the current organization")
	}

	return nil
}

// finalizePaymentTransaction creates ledger entry, updates transaction, and commits
func (ts *TransactionService) finalizePaymentTransaction(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	transaction *model.Transaction,
	effectiveMemberProfileID *uuid.UUID,
	data *model.PaymentQuickRequest,
	userOrg *model.UserOrganization,
	transactionBatch *model.TransactionBatch,
	credit, debit, newBalance float64,
	transactionSource model.GeneralLedgerSource,
	amount float64,
	block func(string),
) (uuid.UUID, error) {
	// Create general ledger entry
	newGeneralLedger := &model.GeneralLedger{
		CreatedAt:                  time.Now().UTC(),
		CreatedByID:                userOrg.UserID,
		UpdatedAt:                  time.Now().UTC(),
		UpdatedByID:                userOrg.UserID,
		BranchID:                   *userOrg.BranchID,
		OrganizationID:             userOrg.OrganizationID,
		TransactionBatchID:         &transactionBatch.ID,
		ReferenceNumber:            data.ReferenceNumber,
		TransactionID:              &transaction.ID,
		EntryDate:                  data.EntryDate,
		SignatureMediaID:           data.SignatureMediaID,
		ProofOfPaymentMediaID:      data.ProofOfPaymentMediaID,
		BankID:                     data.BankID,
		AccountID:                  data.AccountID,
		Credit:                     credit,
		Debit:                      debit,
		Balance:                    newBalance,
		MemberProfileID:            effectiveMemberProfileID,
		MemberJointAccountID:       data.MemberJointAccountID,
		PaymentTypeID:              data.PaymentTypeID,
		TransactionReferenceNumber: data.ReferenceNumber,
		Source:                     transactionSource,
		BankReferenceNumber:        data.BankReferenceNumber,
		EmployeeUserID:             &userOrg.UserID,
		Description:                data.Description,
	}

	if err := ts.model.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "ledger-create-error",
			Description: "Failed to create general ledger entry: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to create general ledger entry: " + err.Error())
		return uuid.Nil, eris.Wrap(err, "failed to create general ledger entry")
	}

	// Update transaction amount - track the net effect on the transaction batch
	if transactionSource == model.GeneralLedgerSourceDeposit {
		transaction.Amount += amount // Add the positive amount
	} else if transactionSource == model.GeneralLedgerSourceWithdraw {
		transaction.Amount -= amount // Subtract the positive amount
	}

	if err := ts.model.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-update-error",
			Description: "Failed to update transaction: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to update transaction: " + err.Error())
		return uuid.Nil, eris.Wrap(err, "failed to update transaction")
	}

	// Update or create member accounting ledger (only for member-specific transactions)
	if effectiveMemberProfileID != nil {
		lastPayTime := time.Now().UTC()
		if data.EntryDate != nil {
			lastPayTime = *data.EntryDate
		}

		_, err := ts.model.MemberAccountingLedgerUpdateOrCreate(
			context, tx,
			*effectiveMemberProfileID,
			*data.AccountID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
			userOrg.UserID,
			newBalance,
			lastPayTime,
		)
		if err != nil {
			ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
				Activity:    "member-accounting-ledger-error",
				Description: "Failed to update member accounting ledger: " + err.Error(),
				Module:      "TransactionService",
			})
			tx.Rollback()
			block("Failed to update member accounting ledger: " + err.Error())
			return uuid.Nil, eris.Wrap(err, "failed to update member accounting ledger")
		}

		// Log successful member accounting ledger update
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity: "member-accounting-ledger-updated",
			Description: fmt.Sprintf("Member accounting ledger updated successfully for member %s, account %s, new balance: %.2f",
				effectiveMemberProfileID.String(), data.AccountID.String(), newBalance),
			Module: "TransactionService",
		})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction: " + err.Error(),
			Module:      "TransactionService",
		})
		return uuid.Nil, eris.Wrap(err, "failed to commit transaction")
	}

	return newGeneralLedger.ID, nil
}
