package event

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
)

func (e *Event) Payment(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	data *model.PaymentQuickRequest,
	transactionId *uuid.UUID,
	transactionType TransactionType,
) error {
	// IP Block Check
	block, blocked, err := e.HandleIPBlocker(context, ctx)
	if err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal error: " + err.Error()})
	}
	if blocked {
		tx.Rollback()
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Your IP is temporarily blocked due to repeated errors."})
	}
	// Validate Payment Amount
	if data.Amount == 0 {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "payment-error",
			Description: "Payment amount cannot be zero (/transaction/withdraw/:transaction_id)",
			Module:      "Transaction",
		})
		block("Payment amount cannot be zero")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment amount cannot be zero"})
	}

	// Get User Organization
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/withdraw/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to get user organization: " + err.Error())
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
	}

	// Member Profile Checks (Subsidiary Ledger)
	if data.MemberProfileID != nil {
		memberProfile, err := e.model.MemberProfileManager.GetByID(context, *data.MemberProfileID)
		if err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "member-error",
				Description: "Failed to retrieve member profile (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to retrieve member profile: " + err.Error())
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve member profile: " + err.Error()})
		}

		if memberProfile.BranchID != *userOrg.BranchID {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "branch-mismatch",
				Description: "Member does not belong to the current branch (/transaction/withdraw/:transaction_id)",
				Module:      "Transaction",
			})
			block("Member does not belong to the current branch")
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Member does not belong to the current branch"})
		}

		if memberProfile.OrganizationID != userOrg.OrganizationID {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "organization-mismatch",
				Description: "Member does not belong to the current organization (/transaction/withdraw/:transaction_id)",
				Module:      "Transaction",
			})
			block("Member does not belong to the current organization")
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Member does not belong to the current organization"})
		}
	}

	// Get current TransactionBatch
	transactionBatch, err := e.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Failed to retrieve transaction batch (/transaction/withdraw/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve transaction batch: " + err.Error())
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
	}

	// Add null validation for required fields (MemberProfileID is optional for cooperative-level transactions)
	if data.AccountID == nil || data.PaymentTypeID == nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Missing required fields (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		block("Missing required fields")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields: AccountID and PaymentTypeID are required"})
	}

	// Check account ownership and organization
	account, err := e.model.AccountManager.GetByID(context, *data.AccountID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-error",
			Description: "Failed to retrieve account (/transaction/withdraw/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve account: " + err.Error())
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
	}

	if account.BranchID != *userOrg.BranchID {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "branch-mismatch",
			Description: "Account does not belong to the current branch (/transaction/withdraw/:transaction_id)",
			Module:      "Transaction",
		})
		block("Account does not belong to the current branch")
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to the current branch"})
	}

	if account.OrganizationID != userOrg.OrganizationID {
		tx.Rollback() // ADD THIS LINE
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "organization-mismatch",
			Description: "Account does not belong to the current organization (/transaction/withdraw/:transaction_id)",
			Module:      "Transaction",
		})
		block("Account does not belong to the current organization")
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to the current organization"})
	}

	// Check if payment type is valid
	paymentType, err := e.model.PaymentTypeManager.GetByID(context, *data.PaymentTypeID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "payment-type-error",
			Description: "Failed to retrieve payment type (/transaction/withdraw/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve payment type: " + err.Error())
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
	}

	if paymentType.OrganizationID != userOrg.OrganizationID {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "organization-mismatch",
			Description: "Payment type does not belong to the current organization (/transaction/withdraw/:transaction_id)",
			Module:      "Transaction",
		})
		block("Payment type does not belong to the current organization")
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Payment type does not belong to the current organization"})
	}

	// Lock the latest general ledger entry for update (only if member-specific transaction)
	var generalLedger *model.GeneralLedger
	if data.MemberProfileID != nil {
		var err error
		generalLedger, err = e.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *data.MemberProfileID, *data.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			// Check if the error is because no general ledger exists (new member/account)
			// This is normal for new members who haven't made any transactions yet
			if err == gorm.ErrRecordNotFound {
				// No general ledger exists yet - this is fine for new members
				// Set generalLedger to nil to indicate starting balance of 0
				generalLedger = nil
				e.Footstep(context, ctx, FootstepEvent{
					Activity: "new-member-account",
					Description: fmt.Sprintf("No existing general ledger found for member %s on account %s - starting with zero balance (/transaction/payment/:transaction_id)",
						data.MemberProfileID.String(), data.AccountID.String()),
					Module: "Transaction",
				})
			} else {
				// Actual database error occurred
				tx.Rollback()
				e.Footstep(context, ctx, FootstepEvent{
					Activity:    "ledger-error",
					Description: "Failed to retrieve general ledger (FOR UPDATE) (/transaction/payment/:transaction_id): " + err.Error(),
					Module:      "Transaction",
				})
				block("Failed to retrieve general ledger: " + err.Error())
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
			}
		}
	} else {
		// Cooperative-level transaction - no member-specific ledger needed
		generalLedger = nil
		e.Footstep(context, ctx, FootstepEvent{
			Activity: "cooperative-transaction",
			Description: fmt.Sprintf("Cooperative-level transaction on account %s - no member profile (/transaction/payment/:transaction_id)",
				data.AccountID.String()),
			Module: "Transaction",
		})
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Computation - Handle debit/credit logic based on account type and transaction type
	var credit, debit, newBalance float64

	// Handle negative amounts by flipping transaction type and making amount positive
	amount := data.Amount
	effectiveTransactionType := transactionType
	var negativeAmountReason string

	if amount < 0 {
		amount = -amount // Make amount positive for calculations
		negativeAmountReason = fmt.Sprintf("Negative amount %.2f converted to positive %.2f", data.Amount, amount)

		// Flip the transaction type when amount is negative
		if transactionType == TransactionTypeDeposit {
			effectiveTransactionType = TransactionTypeWithdraw
			negativeAmountReason += " and transaction type changed from deposit to withdraw"
		} else {
			effectiveTransactionType = TransactionTypeDeposit
			negativeAmountReason += " and transaction type changed from withdraw to deposit"
		}

		// Log the negative amount handling
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "negative-amount-handled",
			Description: negativeAmountReason + " (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
	}

	switch account.Type {
	case model.AccountTypeDeposit:
		// Deposit accounts: Credits increase balance, Debits decrease balance
		if effectiveTransactionType == TransactionTypeDeposit {
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
				e.Footstep(context, ctx, FootstepEvent{
					Activity:    "balance-error",
					Description: reason + " (/transaction/payment/:transaction_id)",
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Insufficient balance for withdrawal"})
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	case model.AccountTypeLoan:
		// Loan accounts: Deposits reduce loan balance (debit), Withdrawals increase loan balance (credit)
		if effectiveTransactionType == TransactionTypeDeposit {
			credit = 0
			debit = amount // Payment reduces loan balance
			if generalLedger != nil {
				newBalance = generalLedger.Balance - amount
				// Prevent overpayment - loan balance shouldn't go below 0
				if newBalance < 0 {
					tx.Rollback()
					reason := fmt.Sprintf("Loan overpayment not allowed. Current balance: %.2f, Payment: %.2f", generalLedger.Balance, amount)
					block(reason)
					e.Footstep(context, ctx, FootstepEvent{
						Activity:    "loan-overpayment-error",
						Description: reason + " (/transaction/payment/:transaction_id)",
						Module:      "Transaction",
					})
					return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment exceeds loan balance"})
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
		if effectiveTransactionType == TransactionTypeDeposit {
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
		if effectiveTransactionType == TransactionTypeDeposit {
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
				e.Footstep(context, ctx, FootstepEvent{
					Activity:    "balance-error",
					Description: reason + " (/transaction/payment/:transaction_id)",
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Insufficient balance for withdrawal"})
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	default:
		// Unsupported account type
		tx.Rollback()
		reason := fmt.Sprintf("Unsupported account type: %s", account.Type)
		block(reason)
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-type-error",
			Description: reason + " (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported account type"})
	}

	if (account.MinAmount != 0 || account.MaxAmount != 0) &&
		(newBalance < account.MinAmount || newBalance > account.MaxAmount) {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity: "balance-limit-error",
			Description: fmt.Sprintf(
				"Balance %.2f exceeds account limits [%.2f-%.2f] (%s)",
				newBalance,
				account.MinAmount,
				account.MaxAmount,
				"/transaction/payment/:transaction_id",
			),
			Module: "Transaction",
		})
		block(fmt.Sprintf("Account balance limits exceeded: %.2f not in [%.2f-%.2f]", newBalance, account.MinAmount, account.MaxAmount))
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf(
				"Account balance must be between %.2f and %.2f. Result would be %.2f",
				account.MinAmount,
				account.MaxAmount,
				newBalance,
			),
		})
	}

	// FIX 3: Complete the transaction instead of returning nil
	// Handle existing transaction or create new one
	transaction := &model.Transaction{}

	// Determine correct transaction source
	transactionSource := model.GeneralLedgerSourceDeposit
	if effectiveTransactionType == TransactionTypeWithdraw {
		transactionSource = model.GeneralLedgerSourceWithdraw
	}

	if transactionId != nil {
		var err error
		transaction, err = e.model.TransactionManager.GetByID(context, *transactionId)
		if err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "transaction-error",
				Description: "Failed to retrieve transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to retrieve transaction: " + err.Error())
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction: " + err.Error()})
		}

		// Validate transaction belongs to current batch
		if transaction.TransactionBatchID == nil || *transaction.TransactionBatchID != transactionBatch.ID {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "transaction-batch-mismatch",
				Description: "Transaction does not belong to the current transaction batch (/transaction/payment/:transaction_id)",
				Module:      "Transaction",
			})
			block("Transaction batch mismatch")
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "Transaction does not belong to the current transaction batch",
			})
		}
	} else {
		// Create new transaction
		transaction = &model.Transaction{
			CreatedAt:            time.Now().UTC(),
			CreatedByID:          userOrg.UserID,
			UpdatedAt:            time.Now().UTC(),
			UpdatedByID:          userOrg.UserID,
			BranchID:             *userOrg.BranchID,
			OrganizationID:       userOrg.OrganizationID,
			SignatureMediaID:     data.SignatureMediaID,
			TransactionBatchID:   &transactionBatch.ID,
			EmployeeUserID:       &userOrg.UserID,
			MemberProfileID:      data.MemberProfileID,
			MemberJointAccountID: data.MemberJointAccountID,
			Amount:               0, // Will be updated below
			ReferenceNumber:      data.ReferenceNumber,
			Source:               transactionSource, // Use correct source
			Description:          data.Description,
			LoanBalance:          0,
			LoanDue:              0,
			TotalDue:             0,
			FinesDue:             0,
			TotalLoan:            0,
			InterestDue:          0,
		}

		if err := e.model.TransactionManager.CreateWithTx(context, tx, transaction); err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "transaction-create-error",
				Description: "Failed to create transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to create transaction: " + err.Error())
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction: " + err.Error()})
		}
	}

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
		MemberProfileID:            data.MemberProfileID,
		MemberJointAccountID:       data.MemberJointAccountID,
		PaymentTypeID:              data.PaymentTypeID,
		TransactionReferenceNumber: data.ReferenceNumber,
		Source:                     transactionSource, // Use correct source
		BankReferenceNumber:        data.BankReferenceNumber,
		EmployeeUserID:             &userOrg.UserID,
		Description:                data.Description,
	}

	if err := e.model.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "ledger-create-error",
			Description: "Failed to create general ledger entry (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to create general ledger entry: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create general ledger entry: " + err.Error()})
	}

	// Update transaction amount - handle all combinations of original intent and effective operations
	// We want to track the net effect on the transaction batch amount
	if effectiveTransactionType == TransactionTypeDeposit {
		// Effective operation is a deposit (increases account balance)
		transaction.Amount += amount // Add the positive amount
	} else {
		// Effective operation is a withdrawal (decreases account balance)
		transaction.Amount -= amount // Subtract the positive amount
	}

	if err := e.model.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-update-error",
			Description: "Failed to update transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to update transaction: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to commit transaction: " + err.Error(),
		})
	}

	// Return success with the created ledger entry
	return ctx.JSON(http.StatusOK, e.model.GeneralLedgerManager.ToModel(newGeneralLedger))
}
