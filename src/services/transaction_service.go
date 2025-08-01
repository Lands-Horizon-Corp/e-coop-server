package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src/model"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// TransactionService handles all transaction-related business logic
type TransactionService struct {
	model                 *model.Model
	userOrganizationToken model.UserOrganization
	footstepLogger        model.Footstep
	ipBlocker             IPBlockerInterface
}

// IPBlockerInterface defines the interface for IP blocking functionality
type IPBlockerInterface interface {
	HandleIPBlocker(context context.Context, ctx echo.Context) (func(string), bool, error)
}

// FootstepEvent represents an audit log event
type FootstepEvent struct {
	Activity    string
	Description string
	Module      string
}

// PaymentRequest represents a payment transaction request
type PaymentRequest struct {
	Amount                *float64
	SignatureMediaID      *uuid.UUID
	ProofOfPaymentMediaID *uuid.UUID
	BankID                *uuid.UUID
	BankReferenceNumber   *string
	EntryDate             *time.Time
	AccountID             *uuid.UUID
	PaymentTypeID         *uuid.UUID
	MemberProfileID       *uuid.UUID
	MemberJointAccountID  *uuid.UUID
	ReferenceNumber       *string
	Description           *string
}

// TransactionResult represents the result of a transaction operation
type TransactionResult struct {
	GeneralLedgerID uuid.UUID
	TransactionID   uuid.UUID
	Amount          float64
	NewBalance      float64
	EffectiveSource model.GeneralLedgerSource
	IsNegative      bool
}

// NewTransactionService creates a new instance of TransactionService
func NewTransactionService(
	model *model.Model,
	userOrgToken UserOrganizationTokenInterface,
	footstepLogger FootstepLoggerInterface,
	ipBlocker IPBlockerInterface,
) *TransactionService {
	return &TransactionService{
		model:                 model,
		userOrganizationToken: userOrgToken,
		footstepLogger:        footstepLogger,
		ipBlocker:             ipBlocker,
	}
}

// ProcessPayment processes a payment transaction with the specified general ledger source
func (ts *TransactionService) ProcessPayment(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
	transactionID *uuid.UUID,
	ledgerSource model.GeneralLedgerSource,
) (*TransactionResult, error) {
	// Convert PaymentRequest to PaymentQuickRequest
	quickRequest := &model.PaymentQuickRequest{
		Amount:                request.Amount,
		SignatureMediaID:      request.SignatureMediaID,
		ProofOfPaymentMediaID: request.ProofOfPaymentMediaID,
		BankID:                request.BankID,
		BankReferenceNumber:   request.BankReferenceNumber,
		EntryDate:             request.EntryDate,
		AccountID:             request.AccountID,
		PaymentTypeID:         request.PaymentTypeID,
		MemberProfileID:       request.MemberProfileID,
		MemberJointAccountID:  request.MemberJointAccountID,
		ReferenceNumber:       request.ReferenceNumber,
		Description:           request.Description,
	}

	// Performance monitoring
	defer func() {
		duration := time.Since(startTime)
		if duration > 5*time.Second {
			ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
				Activity:    "performance-warning",
				Description: fmt.Sprintf("Payment operation took %.2fs - potential performance issue", duration.Seconds()),
				Module:      "TransactionService",
			})
		}
	}()

	// IP Block Check
	block, blocked, err := ts.ipBlocker.HandleIPBlocker(context, ctx)
	if err != nil {
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "ip-block-check-error",
			Description: "IP blocker check failed: " + err.Error(),
			Module:      "TransactionService",
		})
		tx.Rollback()
		return nil, eris.Wrap(err, "internal error during IP block check")
	}
	if blocked {
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "ip-blocked",
			Description: "IP is temporarily blocked due to repeated errors",
			Module:      "TransactionService",
		})
		tx.Rollback()
		return nil, eris.New("IP is temporarily blocked due to repeated errors")
	}

	// Validate payment request
	if err := ts.validatePaymentRequest(context, ctx, data, block); err != nil {
		return nil, err
	}

	// Get user organization
	userOrg, err := ts.userOrganizationToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to get user organization: " + err.Error())
		return nil, eris.Wrap(err, "failed to get user organization")
	}

	// Validate member profile if provided
	if err := ts.validateMemberProfile(context, ctx, data, userOrg, block); err != nil {
		return nil, err
	}

	// Get transaction batch
	transactionBatch, err := ts.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		tx.Rollback()
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Failed to retrieve transaction batch: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to retrieve transaction batch: " + err.Error())
		return nil, eris.Wrap(err, "failed to retrieve transaction batch")
	}

	// Validate and lock account
	account, err := ts.validateAndLockAccount(context, ctx, tx, data, userOrg, block)
	if err != nil {
		return nil, err
	}

	// Validate payment type
	if err := ts.validatePaymentType(context, ctx, data, userOrg, block); err != nil {
		return nil, err
	}

	// Get and lock general ledger
	generalLedger, err := ts.getAndLockGeneralLedger(context, ctx, tx, data, transactionId, userOrg, block)
	if err != nil {
		return nil, err
	}

	// Calculate account balances
	credit, debit, newBalance, effectiveType, processedAmount, isNegative, err := ts.calculateAccountBalances(
		context, ctx, tx, data.Amount, transactionType, account, generalLedger, block)
	if err != nil {
		return nil, err
	}

	// Handle transaction creation/retrieval
	transaction, effectiveMemberProfileID, err := ts.handleTransactionAndMember(
		context, ctx, tx, transactionId, data, userOrg, transactionBatch, block)
	if err != nil {
		return nil, err
	}

	// Determine transaction source
	transactionSource := model.GeneralLedgerSourceDeposit
	if effectiveType == TransactionTypeWithdraw {
		transactionSource = model.GeneralLedgerSourceWithdraw
	}

	// Create general ledger entry and finalize
	generalLedgerID, err := ts.finalizePaymentTransaction(
		context, ctx, tx, transaction, effectiveMemberProfileID, data, userOrg,
		transactionBatch, credit, debit, newBalance, transactionSource,
		effectiveType, processedAmount, block)
	if err != nil {
		return nil, err
	}

	// Success - log completion
	duration := time.Since(startTime)
	ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
		Activity: "payment-success",
		Description: fmt.Sprintf("Payment completed successfully. Amount: %.2f, Account: %s, Balance: %.2f, Duration: %.3fs",
			processedAmount, data.AccountID.String(), newBalance, duration.Seconds()),
		Module: "TransactionService",
	})

	// Return transaction result
	return &TransactionResult{
		TransactionID:     transaction.ID,
		GeneralLedgerID:   generalLedgerID,
		MemberProfileID:   effectiveMemberProfileID,
		AccountID:         *data.AccountID,
		OriginalAmount:    data.Amount,
		ProcessedAmount:   processedAmount,
		Credit:            credit,
		Debit:             debit,
		NewBalance:        newBalance,
		EffectiveType:     effectiveType,
		TransactionSource: transactionSource,
		Duration:          duration,
		IsNegativeAmount:  isNegative,
	}, nil
}

// ProcessDeposit is a convenience method for deposit transactions
func (ts *TransactionService) ProcessDeposit(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
) (*TransactionResult, error) {
	return ts.ProcessPayment(context, ctx, tx, request, nil, TransactionTypeDeposit)
}

// ProcessWithdrawal is a convenience method for withdrawal transactions
func (ts *TransactionService) ProcessWithdrawal(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
) (*TransactionResult, error) {
	return ts.ProcessPayment(context, ctx, tx, request, nil, TransactionTypeWithdraw)
}

// ProcessPaymentWithTransactionID processes payment for an existing transaction
func (ts *TransactionService) ProcessPaymentWithTransactionID(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
	transactionId uuid.UUID,
	transactionType TransactionType,
) (*TransactionResult, error) {
	return ts.ProcessPayment(context, ctx, tx, request, &transactionId, transactionType)
}
