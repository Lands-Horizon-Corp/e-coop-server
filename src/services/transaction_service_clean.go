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
	model          *model.Model
	userOrgToken   UserOrganizationTokenInterface
	footstepLogger FootstepLoggerInterface
	ipBlocker      IPBlockerInterface
}

// UserOrganizationTokenInterface defines the interface for user organization token operations
type UserOrganizationTokenInterface interface {
	CurrentUserOrganization(context context.Context, ctx echo.Context) (*model.UserOrganization, error)
}

// FootstepLoggerInterface defines the interface for footstep logging
type FootstepLoggerInterface interface {
	Footstep(context context.Context, ctx echo.Context, event FootstepEvent)
}

// IPBlockerInterface defines the interface for IP blocking functionality
type IPBlockerInterface interface {
	HandleIPBlocker(context context.Context, ctx echo.Context) (func(string), bool, error)
}

// FootstepEvent represents a footstep event for audit logging
type FootstepEvent struct {
	Activity    string
	Description string
	Module      string
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount                *float64   `json:"amount"`
	SignatureMediaID      *uuid.UUID `json:"signature_media_id"`
	ProofOfPaymentMediaID *uuid.UUID `json:"proof_of_payment_media_id"`
	BankID                *uuid.UUID `json:"bank_id"`
	BankReferenceNumber   *string    `json:"bank_reference_number"`
	EntryDate             *time.Time `json:"entry_date"`
	AccountID             *uuid.UUID `json:"account_id"`
	PaymentTypeID         *uuid.UUID `json:"payment_type_id"`
	MemberProfileID       *uuid.UUID `json:"member_profile_id"`
	MemberJointAccountID  *uuid.UUID `json:"member_joint_account_id"`
	ReferenceNumber       *string    `json:"reference_number"`
	Description           *string    `json:"description"`
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

// NewTransactionService creates a new transaction service instance
func NewTransactionService(
	model *model.Model,
	userOrgToken UserOrganizationTokenInterface,
	footstepLogger FootstepLoggerInterface,
	ipBlocker IPBlockerInterface,
) *TransactionService {
	return &TransactionService{
		model:          model,
		userOrgToken:   userOrgToken,
		footstepLogger: footstepLogger,
		ipBlocker:      ipBlocker,
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
	// Convert PaymentRequest to PaymentQuickRequest for compatibility
	quickRequest := &model.PaymentQuickRequest{
		Amount:                getValue(request.Amount, 0.0),
		SignatureMediaID:      request.SignatureMediaID,
		ProofOfPaymentMediaID: request.ProofOfPaymentMediaID,
		BankID:                request.BankID,
		BankReferenceNumber:   getValue(request.BankReferenceNumber, ""),
		EntryDate:             request.EntryDate,
		AccountID:             request.AccountID,
		PaymentTypeID:         request.PaymentTypeID,
		MemberProfileID:       request.MemberProfileID,
		MemberJointAccountID:  request.MemberJointAccountID,
		ReferenceNumber:       getValue(request.ReferenceNumber, ""),
		Description:           getValue(request.Description, ""),
	}

	// Performance monitoring
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if duration > 5*time.Second {
			ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
				Activity:    "performance-warning",
				Description: fmt.Sprintf("Transaction service operation took %.2fs - potential performance issue", duration.Seconds()),
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
	if err := ts.validatePaymentRequest(context, ctx, quickRequest, block); err != nil {
		return nil, err
	}

	// Get user organization
	userOrg, err := ts.userOrgToken.CurrentUserOrganization(context, ctx)
	if err != nil {
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "auth-error",
			Description: "Failed to get user organization: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to get user organization: " + err.Error())
		return nil, eris.Wrap(err, "failed to get user organization")
	}

	// Validate member profile if provided
	if err := ts.validateMemberProfile(context, ctx, quickRequest, userOrg, block); err != nil {
		return nil, err
	}

	// Get transaction batch
	transactionBatch, err := ts.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
	if err != nil {
		ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
			Activity:    "batch-error",
			Description: "Failed to retrieve transaction batch: " + err.Error(),
			Module:      "TransactionService",
		})
		block("Failed to retrieve transaction batch: " + err.Error())
		return nil, eris.Wrap(err, "failed to retrieve transaction batch")
	}

	// Validate and lock account
	account, err := ts.validateAndLockAccount(context, ctx, tx, quickRequest, userOrg, block)
	if err != nil {
		return nil, err
	}

	// Validate payment type
	if err := ts.validatePaymentType(context, ctx, quickRequest, userOrg, block); err != nil {
		return nil, err
	}

	// Get and lock general ledger
	generalLedger, err := ts.getAndLockGeneralLedger(context, ctx, tx, quickRequest, transactionID, userOrg, block)
	if err != nil {
		return nil, err
	}

	// Calculate account balances
	credit, debit, newBalance, effectiveSource, processedAmount, isNegative, err := ts.calculateAccountBalances(
		context, ctx, tx, quickRequest.Amount, ledgerSource, account, generalLedger, block)
	if err != nil {
		return nil, err
	}

	// Handle transaction creation/retrieval
	transaction, effectiveMemberProfileID, err := ts.handleTransactionAndMember(
		context, ctx, tx, transactionID, quickRequest, userOrg, transactionBatch, block)
	if err != nil {
		return nil, err
	}

	// Finalize transaction
	generalLedgerID, err := ts.finalizePaymentTransaction(
		context, ctx, tx, transaction, effectiveMemberProfileID, quickRequest, userOrg, transactionBatch,
		credit, debit, newBalance, effectiveSource, processedAmount, block)
	if err != nil {
		return nil, err
	}

	// Success logging
	duration := time.Since(startTime)
	ts.footstepLogger.Footstep(context, ctx, FootstepEvent{
		Activity: "payment-success",
		Description: fmt.Sprintf("Transaction completed successfully. Amount: %.2f, Account: %s, Balance: %.2f, Duration: %.3fs",
			processedAmount, quickRequest.AccountID.String(), newBalance, duration.Seconds()),
		Module: "TransactionService",
	})

	return &TransactionResult{
		GeneralLedgerID: generalLedgerID,
		TransactionID:   transaction.ID,
		Amount:          processedAmount,
		NewBalance:      newBalance,
		EffectiveSource: effectiveSource,
		IsNegative:      isNegative,
	}, nil
}

// ProcessDeposit processes a deposit transaction
func (ts *TransactionService) ProcessDeposit(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
	userOrg *model.UserOrganization,
	transactionBatch *model.TransactionBatch,
) (*TransactionResult, error) {
	return ts.ProcessPayment(context, ctx, tx, request, nil, model.GeneralLedgerSourceDeposit)
}

// ProcessWithdrawal processes a withdrawal transaction
func (ts *TransactionService) ProcessWithdrawal(
	context context.Context,
	ctx echo.Context,
	tx *gorm.DB,
	request *PaymentRequest,
	userOrg *model.UserOrganization,
	transactionBatch *model.TransactionBatch,
) (*TransactionResult, error) {
	return ts.ProcessPayment(context, ctx, tx, request, nil, model.GeneralLedgerSourceWithdraw)
}

// Helper function to get value from pointer with default
func getValue[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
