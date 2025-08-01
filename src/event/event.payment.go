package event

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

func (e *Event) Payment(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	data *model.PaymentQuickRequest,
	transactionId *uuid.UUID,
	ledgerSource model.GeneralLedgerSource,
) (*model.GeneralLedger, error) {
	// ================================================================================
	// STEP 1: INITIALIZATION & PERFORMANCE MONITORING
	// ================================================================================
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if duration > 5*time.Second {
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "performance-warning",
				Description: fmt.Sprintf("Payment operation took %.2fs - potential performance issue (/transaction/payment/:transaction_id)", duration.Seconds()),
				Module:      "Transaction",
			})
		}
	}()

	// ================================================================================
	// STEP 2: SECURITY & ACCESS CONTROL
	// ================================================================================
	// IP Block Check
	block, blocked, err := e.HandleIPBlocker(context, ctx)
	if err != nil {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "ip-block-check-error",
			Description: "IP blocker check failed (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		tx.Rollback()
		return nil, eris.Wrap(err, "internal error during IP block check")
	}

	if blocked {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "ip-blocked",
			Description: "IP is temporarily blocked due to repeated errors (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		tx.Rollback()
		return nil, eris.New("IP is temporarily blocked due to repeated errors")
	}

	// Get User Organization
	userOrg, err := e.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to get user organization: " + err.Error())
		return nil, eris.Wrap(err, "failed to get user organization")
	}

	// ================================================================================
	// STEP 3: INPUT VALIDATION
	// ================================================================================
	// Validate Payment Amount
	if data.Amount == 0 {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "payment-error",
			Description: "Payment amount cannot be zero (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		block("Payment amount cannot be zero")
		return nil, eris.New("payment amount cannot be zero")
	}

	// Validate required fields
	if data.AccountID == nil || data.PaymentTypeID == nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "validation-error",
			Description: "Missing required fields (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		block("Missing required fields")
		return nil, eris.New("missing required fields: AccountID and PaymentTypeID are required")
	}

	// ================================================================================
	// STEP 4: ENTITY VALIDATION & RETRIEVAL
	// ================================================================================
	// Validate Member Profile (if provided)
	if data.MemberProfileID != nil {
		memberProfile, err := e.model.MemberProfileManager.GetByID(context, *data.MemberProfileID)
		if err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "member-error",
				Description: "Failed to retrieve member profile (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to retrieve member profile: " + err.Error())
			return nil, eris.Wrap(err, "failed to retrieve member profile")
		}

		if memberProfile.BranchID != *userOrg.BranchID {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "branch-mismatch",
				Description: "Member does not belong to the current branch (/transaction/payment/:transaction_id)",
				Module:      "Transaction",
			})
			block("Member does not belong to the current branch")
			return nil, eris.New("member does not belong to the current branch")
		}
		if memberProfile.OrganizationID != userOrg.OrganizationID {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "organization-mismatch",
				Description: "Member does not belong to the current organization (/transaction/payment/:transaction_id)",
				Module:      "Transaction",
			})
			block("Member does not belong to the current organization")
			return nil, eris.New("member does not belong to the current organization")
		}
	}

	// Get Transaction Batch
	transactionBatch, err := e.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve transaction batch: " + err.Error())
		return nil, eris.Wrap(err, "failed to retrieve transaction batch")
	}

	// Validate Account
	account, err := e.model.AccountManager.GetByID(context, *data.AccountID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-error",
			Description: "Failed to retrieve account (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve account: " + err.Error())
		return nil, eris.Wrap(err, "failed to retrieve account")
	}

	if account.BranchID != *userOrg.BranchID {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "branch-mismatch",
			Description: "Account does not belong to the current branch (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		block("Account does not belong to the current branch")
		return nil, eris.New("account does not belong to the current branch")
	}

	if account.OrganizationID != userOrg.OrganizationID {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "organization-mismatch",
			Description: "Account does not belong to the current organization (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		tx.Rollback()
		block("Account does not belong to the current organization")
		return nil, eris.New("account does not belong to the current organization")
	}

	// Validate Payment Type
	paymentType, err := e.model.PaymentTypeManager.GetByID(context, *data.PaymentTypeID)
	if err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "payment-type-error",
			Description: "Failed to retrieve payment type (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to retrieve payment type: " + err.Error())
		return nil, eris.Wrap(err, "failed to retrieve payment type")
	}

	if paymentType.OrganizationID != userOrg.OrganizationID {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "organization-mismatch",
			Description: "Payment type does not belong to the current organization (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
		block("Payment type does not belong to the current organization")
		return nil, eris.New("payment type does not belong to the current organization")
	}

	// ================================================================================
	// STEP 5: CONCURRENCY PROTECTION & LOCKING
	// ================================================================================
	// Lock account for concurrent protection
	lockedAccount, err := e.model.AccountLockWithValidation(context, tx, *data.AccountID, account)
	if err != nil {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-lock-error",
			Description: "Failed to acquire account lock for concurrent protection (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		tx.Rollback()
		block("Failed to acquire account lock: " + err.Error())
		return nil, eris.Wrap(err, "failed to acquire account lock for concurrent protection")
	}
	account = lockedAccount

	// ================================================================================
	// STEP 6: TRANSACTION SOURCE DETERMINATION
	// ================================================================================
	var transactionSource model.GeneralLedgerSource
	var existingTransaction *model.Transaction

	if transactionId != nil {
		// Check if transaction exists and get its source
		existingTransaction, err = e.model.TransactionManager.GetByID(context, *transactionId)
		if err == nil && existingTransaction.Source != "" {
			// Use source from existing transaction
			transactionSource = existingTransaction.Source
			e.Footstep(context, ctx, FootstepEvent{
				Activity: "transaction-source-from-existing",
				Description: fmt.Sprintf("Using transaction source '%s' from existing transaction %s (/transaction/payment/:transaction_id)",
					string(existingTransaction.Source), existingTransaction.ID.String()),
				Module: "Transaction",
			})
		} else {
			// Use the effective ledger source from parameters
			transactionSource = ledgerSource
			if err != nil {
				e.Footstep(context, ctx, FootstepEvent{
					Activity: "transaction-not-found",
					Description: fmt.Sprintf("Transaction not found, will create new one with source '%s' (/transaction/payment/:transaction_id)",
						string(ledgerSource)),
					Module: "Transaction",
				})
			}
		}
	} else {
		// New transaction - use the ledger source from parameters
		transactionSource = ledgerSource
		e.Footstep(context, ctx, FootstepEvent{
			Activity: "new-transaction-source",
			Description: fmt.Sprintf("Creating new transaction with source '%s' (/transaction/payment/:transaction_id)",
				string(ledgerSource)),
			Module: "Transaction",
		})
	}

	// ================================================================================
	// STEP 7: MEMBER PROFILE DETERMINATION & LEDGER LOCKING
	// ================================================================================
	// Determine effective member profile ID for ledger retrieval
	var effectiveMemberProfileID *uuid.UUID
	if data.MemberProfileID != nil {
		effectiveMemberProfileID = data.MemberProfileID
	} else if existingTransaction != nil && existingTransaction.MemberProfileID != nil {
		effectiveMemberProfileID = existingTransaction.MemberProfileID
		e.Footstep(context, ctx, FootstepEvent{
			Activity: "member-from-existing-transaction",
			Description: fmt.Sprintf("Using member profile ID %s from existing transaction for ledger retrieval (/transaction/payment/:transaction_id)",
				existingTransaction.MemberProfileID.String()),
			Module: "Transaction",
		})
	}

	// Lock the latest general ledger entry
	var generalLedger *model.GeneralLedger
	if effectiveMemberProfileID != nil {
		// Member-specific transaction
		generalLedger, err = e.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *effectiveMemberProfileID, *data.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve member general ledger (FOR UPDATE) (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to retrieve member general ledger: " + err.Error())
			return nil, eris.Wrap(err, "failed to retrieve member general ledger")
		}

		if generalLedger == nil {
			e.Footstep(context, ctx, FootstepEvent{
				Activity: "new-member-account",
				Description: fmt.Sprintf("No existing general ledger found for member %s on account %s - starting with zero balance (/transaction/payment/:transaction_id)",
					effectiveMemberProfileID.String(), data.AccountID.String()),
				Module: "Transaction",
			})
		}
	} else {
		// Subsidiary ledger transaction (non-member/cooperative-level)
		generalLedger, err = e.model.GeneralLedgerCurrentSubsidiaryAccountForUpdate(
			context, tx, *data.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve subsidiary general ledger (FOR UPDATE) (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to retrieve subsidiary general ledger: " + err.Error())
			return nil, eris.Wrap(err, "failed to retrieve subsidiary general ledger")
		}

		if generalLedger == nil {
			e.Footstep(context, ctx, FootstepEvent{
				Activity: "new-subsidiary-account",
				Description: fmt.Sprintf("No existing subsidiary ledger found for account %s - starting with zero balance (/transaction/payment/:transaction_id)",
					data.AccountID.String()),
				Module: "Transaction",
			})
		}

		e.Footstep(context, ctx, FootstepEvent{
			Activity: "subsidiary-transaction",
			Description: fmt.Sprintf("Subsidiary ledger transaction on account %s - no member profile (/transaction/payment/:transaction_id)",
				data.AccountID.String()),
			Module: "Transaction",
		})
	}

	// ================================================================================
	// STEP 8: AMOUNT PROCESSING & TRANSACTION CALCULATION
	// ================================================================================
	// Handle negative amounts and determine effective transaction source
	amount := data.Amount
	effectiveTransactionSource := transactionSource
	var negativeAmountReason string

	if amount < 0 {
		amount = -amount // Make amount positive for calculations
		negativeAmountReason = fmt.Sprintf("Negative amount %.2f converted to positive %.2f", data.Amount, amount)

		// Flip the transaction source when amount is negative
		switch transactionSource {
		case model.GeneralLedgerSourceDeposit:
			effectiveTransactionSource = model.GeneralLedgerSourceWithdraw
			negativeAmountReason += " and transaction source changed from deposit to withdraw"
		case model.GeneralLedgerSourceWithdraw:
			effectiveTransactionSource = model.GeneralLedgerSourceDeposit
			negativeAmountReason += " and transaction source changed from withdraw to deposit"
		}

		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "negative-amount-handled",
			Description: negativeAmountReason + " (/transaction/payment/:transaction_id)",
			Module:      "Transaction",
		})
	}

	// Calculate debit/credit and new balance based on account type
	var credit, debit, newBalance float64
	isDeposited := effectiveTransactionSource == model.GeneralLedgerSourcePayment || effectiveTransactionSource == model.GeneralLedgerSourceDeposit
	switch account.Type {
	case model.AccountTypeDeposit:
		if isDeposited {
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
				return nil, eris.New("insufficient balance for withdrawal")
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	case model.AccountTypeLoan:
		if isDeposited {
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
					return nil, eris.New("payment exceeds loan balance")
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
		if isDeposited {
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
		if isDeposited {
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
				return nil, eris.New("insufficient balance for withdrawal from Other account")
			}
			credit = 0
			debit = amount
			newBalance = generalLedger.Balance - amount
		}

	default:
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "account-type-error",
			Description: fmt.Sprintf("Unsupported account type: %s (/transaction/payment/:transaction_id)", account.Type),
			Module:      "Transaction",
		})
		tx.Rollback()
		reason := fmt.Sprintf("Unsupported account type: %s", account.Type)
		block(reason)
		return nil, eris.New("unsupported account type")
	}

	// Validate account balance limits
	if (account.MinAmount != 0 || account.MaxAmount != 0) &&
		(newBalance < account.MinAmount || newBalance > account.MaxAmount) {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity: "balance-limit-error",
			Description: fmt.Sprintf(
				"Balance %.2f exceeds account limits [%.2f-%.2f] (/transaction/payment/:transaction_id)",
				newBalance,
				account.MinAmount,
				account.MaxAmount,
			),
			Module: "Transaction",
		})
		block(fmt.Sprintf("Account balance limits exceeded: %.2f not in [%.2f-%.2f]", newBalance, account.MinAmount, account.MaxAmount))
		return nil, eris.New("account balance limits exceeded")
	}

	// ================================================================================
	// STEP 9: TRANSACTION MANAGEMENT
	// ================================================================================
	// Handle existing transaction or create new one
	transaction := &model.Transaction{}

	if transactionId != nil {
		var err error
		// Use the existingTransaction we already fetched
		if existingTransaction != nil {
			transaction = existingTransaction
		} else {
			// Try to fetch again if not already fetched
			transaction, err = e.model.TransactionManager.GetByID(context, *transactionId)
		}

		if err != nil {
			// If transaction doesn't exist, create a new one
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new transaction since it doesn't exist
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
					MemberProfileID:      effectiveMemberProfileID,
					MemberJointAccountID: data.MemberJointAccountID,
					Amount:               data.Amount,
					ReferenceNumber:      data.ReferenceNumber,
					Source:               transactionSource,
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
					return nil, eris.Wrap(err, "failed to create transaction")
				}
			} else {
				// Actual database error occurred
				tx.Rollback()
				e.Footstep(context, ctx, FootstepEvent{
					Activity:    "transaction-error",
					Description: "Failed to retrieve transaction (/transaction/payment/:transaction_id): " + err.Error(),
					Module:      "Transaction",
				})
				block("Failed to retrieve transaction: " + err.Error())
				return nil, eris.Wrap(err, "failed to retrieve transaction")
			}
		} else {
			// Transaction exists, validate it belongs to current batch
			if transaction.TransactionBatchID == nil || *transaction.TransactionBatchID != transactionBatch.ID {
				tx.Rollback()
				e.Footstep(context, ctx, FootstepEvent{
					Activity:    "transaction-batch-mismatch",
					Description: "Transaction does not belong to the current transaction batch (/transaction/payment/:transaction_id)",
					Module:      "Transaction",
				})
				block("Transaction batch mismatch")
				return nil, eris.New("transaction does not belong to the current transaction batch")
			}

			// Update effective member profile ID from existing transaction if needed
			if data.MemberProfileID != nil {
				effectiveMemberProfileID = data.MemberProfileID
			} else if transaction.MemberProfileID != nil {
				// No member in data, but transaction has member - use transaction's member
				effectiveMemberProfileID = transaction.MemberProfileID
				e.Footstep(context, ctx, FootstepEvent{
					Activity: "member-from-transaction",
					Description: fmt.Sprintf("Using member profile ID %s from existing transaction (/transaction/payment/:transaction_id)",
						transaction.MemberProfileID.String()),
					Module: "Transaction",
				})
			} else {
				// No member in data and no member in transaction - subsidiary ledger
				effectiveMemberProfileID = nil
			}
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
			MemberProfileID:      effectiveMemberProfileID,
			MemberJointAccountID: data.MemberJointAccountID,
			Amount:               data.Amount,
			ReferenceNumber:      data.ReferenceNumber,
			Source:               transactionSource,
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
			return nil, eris.Wrap(err, "failed to create transaction")
		}
	}

	// ================================================================================
	// STEP 10: GENERAL LEDGER CREATION
	// ================================================================================
	referenceNumber := data.ReferenceNumber
	if transaction.ReferenceNumber != "" {
		referenceNumber = transaction.ReferenceNumber
	}
	newGeneralLedger := &model.GeneralLedger{
		CreatedAt:                  time.Now().UTC(),
		CreatedByID:                userOrg.UserID,
		UpdatedAt:                  time.Now().UTC(),
		UpdatedByID:                userOrg.UserID,
		BranchID:                   *userOrg.BranchID,
		OrganizationID:             userOrg.OrganizationID,
		TransactionBatchID:         &transactionBatch.ID,
		ReferenceNumber:            referenceNumber,
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

	if err := e.model.GeneralLedgerManager.CreateWithTx(context, tx, newGeneralLedger); err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "ledger-create-error",
			Description: "Failed to create general ledger entry (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to create general ledger entry: " + err.Error())
		return nil, eris.Wrap(err, "failed to create general ledger entry")
	}

	// ================================================================================
	// STEP 11: TRANSACTION AMOUNT UPDATE
	// ================================================================================
	// Update transaction amount based on effective operation
	switch effectiveTransactionSource {
	case model.GeneralLedgerSourceDeposit:
		transaction.Amount += amount // Add the positive amount
	case model.GeneralLedgerSourceWithdraw:
		transaction.Amount -= amount // Subtract the positive amount
	}
	if data.ORAutoGenerated {
		if userOrg.UserSettingUsedOR > userOrg.UserSettingEndOR {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "used-or-limit-exceeded",
				Description: "Used OR limit exceeded for user organization (/transaction/payment/:transaction_id)",
				Module:      "Transaction",
			})
			block("Used OR limit exceeded for user organization")
			return nil, eris.New("used OR limit exceeded for user organization")
		}
		userOrg.UserSettingUsedOR = userOrg.UserSettingUsedOR + 1
		if err := e.model.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrg.ID, userOrg); err != nil {
			tx.Rollback()
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "used-or-update-error",
				Description: "Failed to update used OR count (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			block("Failed to update used OR count: " + err.Error())
			return nil, eris.Wrap(err, "failed to update used OR count")
		}
	}

	if err := e.model.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
		tx.Rollback()
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "transaction-update-error",
			Description: "Failed to update transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		block("Failed to update transaction: " + err.Error())
		return nil, eris.Wrap(err, "failed to update transaction")
	}

	// ================================================================================
	// STEP 12: MEMBER ACCOUNTING LEDGER UPDATE
	// ================================================================================
	// Update member accounting ledger (only for member-specific transactions)
	if effectiveMemberProfileID != nil {
		lastPayTime := time.Now().UTC()
		if data.EntryDate != nil {
			lastPayTime = *data.EntryDate
		}

		_, err := e.model.MemberAccountingLedgerUpdateOrCreate(
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
			e.Footstep(context, ctx, FootstepEvent{
				Activity:    "member-accounting-ledger-error",
				Description: "Failed to update member accounting ledger (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			tx.Rollback()
			block("Failed to update member accounting ledger: " + err.Error())
			return nil, eris.Wrap(err, "failed to update member accounting ledger")
		}

		e.Footstep(context, ctx, FootstepEvent{
			Activity: "member-accounting-ledger-updated",
			Description: fmt.Sprintf("Member accounting ledger updated successfully for member %s, account %s, new balance: %.2f (/transaction/payment/:transaction_id)",
				effectiveMemberProfileID.String(), data.AccountID.String(), newBalance),
			Module: "Transaction",
		})
	}

	// ================================================================================
	// STEP 13: TRANSACTION COMMIT & SUCCESS LOGGING
	// ================================================================================
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		e.Footstep(context, ctx, FootstepEvent{
			Activity:    "db-commit-error",
			Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
			Module:      "Transaction",
		})
		return nil, eris.Wrap(err, "failed to commit transaction")
	}

	// Success - log completion with performance metrics
	duration := time.Since(startTime)
	e.Footstep(context, ctx, FootstepEvent{
		Activity: "payment-success",
		Description: fmt.Sprintf("Payment completed successfully. Amount: %.2f, Account: %s, Balance: %.2f, Duration: %.3fs (/transaction/payment/:transaction_id)",
			amount, data.AccountID.String(), newBalance, duration.Seconds()),
		Module: "Transaction",
	})

	return newGeneralLedger, nil
}
