package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

// LoanTransactionTotalResponse represents the total calculations for a loan transaction

func (c *Controller) LoanTransactionController() {
	req := c.provider.Service.Request

	// GET /api/v1/loan-transaction/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/search",
		Method:       "GET",
		ResponseType: model.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for the current user's branch with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.model.LoanTransactionCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionManager.Pagination(context, ctx, loanTransactions))
	})

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for a specific member profile with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.model.LoanTransactionManager.Find(context, &model.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionManager.Pagination(context, ctx, loanTransactions))
	})
	// GET /api/v1/loan-transaction/:loan_transaction_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: model.LoanTransactionResponse{},
		Note:         "Returns a specific loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionManager.ToModel(loanTransaction))
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/total",
		Method:       "GET",
		ResponseType: model.LoanTransactionTotalResponse{},
		Note:         "Returns total calculations for a specific loan transaction including total interest, debit, and credit.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction totals"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Get all loan transaction entries for this loan transaction
		loanTransactionEntries, err := c.model.LoanTransactionEntryManager.Find(context, &model.LoanTransactionEntry{
			LoanTransactionID: *loanTransactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + err.Error()})
		}

		// Calculate totals
		var totalCredit, totalDebit float64
		for _, entry := range loanTransactionEntries {
			totalCredit += entry.Credit
			totalDebit += entry.Debit
		}

		// Calculate total interest (assuming interest is the difference between debit and credit)
		totalInterest := totalDebit - totalCredit

		return ctx.JSON(http.StatusOK, model.LoanTransactionTotalResponse{
			TotalInterest: totalInterest,
			TotalDebit:    totalDebit,
			TotalCredit:   totalCredit,
		})
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/amortization-schedule
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/amortization-schedule",
		Method:       "GET",
		ResponseType: model.AmortizationScheduleResponse{},
		Note:         "Returns the amortization schedule for a specific loan transaction with payment details.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan amortization schedule"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Generate amortization schedule
		schedule, err := c.model.GenerateLoanAmortizationSchedule(context, loanTransaction)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate amortization schedule: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, schedule)
	})

	// POST /api/v1/loan-transaction
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction",
		Method:       "POST",
		ResponseType: model.LoanTransactionResponse{},
		Note:         "Creates a new loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transactions"})
		}

		request, err := c.model.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		loanTransaction := &model.LoanTransaction{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),

			CreatedByID:                            userOrg.UserID,
			UpdatedByID:                            userOrg.UserID,
			OrganizationID:                         userOrg.OrganizationID,
			BranchID:                               *userOrg.BranchID,
			TransactionBatchID:                     &transactionBatch.ID,
			OfficialReceiptNumber:                  request.OfficialReceiptNumber,
			Voucher:                                request.Voucher,
			EmployeeUserID:                         &userOrg.UserID,
			LoanPurposeID:                          request.LoanPurposeID,
			LoanStatusID:                           request.LoanStatusID,
			ModeOfPayment:                          string(request.ModeOfPayment),
			ModeOfPaymentWeekly:                    string(request.ModeOfPaymentWeekly),
			ModeOfPaymentSemiMonthlyPay1:           request.ModeOfPaymentSemiMonthlyPay1,
			ModeOfPaymentSemiMonthlyPay2:           request.ModeOfPaymentSemiMonthlyPay2,
			ComakerType:                            string(request.ComakerType),
			ComakerDepositMemberAccountingLedgerID: request.ComakerDepositMemberAccountingLedgerID,
			ComakerCollateralID:                    request.ComakerCollateralID,
			ComakerCollateralDescription:           request.ComakerCollateralDescription,
			CollectorPlace:                         string(request.CollectorPlace),
			LoanType:                               string(request.LoanType),
			PreviousLoanID:                         request.PreviousLoanID,
			Terms:                                  request.Terms,
			AmortizationAmount:                     request.AmortizationAmount,
			IsAddOn:                                request.IsAddOn,
			Applied1:                               request.Applied1,
			Applied2:                               request.Applied2,
			AccountID:                              request.AccountID,
			MemberProfileID:                        request.MemberProfileID,
			MemberJointAccountID:                   request.MemberJointAccountID,
			SignatureMediaID:                       request.SignatureMediaID,
			MountToBeClosed:                        request.MountToBeClosed,
			DamayanFund:                            request.DamayanFund,
			ShareCapital:                           request.ShareCapital,
			LengthOfService:                        request.LengthOfService,
			ExcludeSunday:                          request.ExcludeSunday,
			ExcludeHoliday:                         request.ExcludeHoliday,
			ExcludeSaturday:                        request.ExcludeSaturday,
			RemarksOtherTerms:                      request.RemarksOtherTerms,
			RemarksPayrollDeduction:                request.RemarksPayrollDeduction,
			RecordOfLoanPaymentsOrLoanStatus:       request.RecordOfLoanPaymentsOrLoanStatus,
			CollateralOffered:                      request.CollateralOffered,
			AppraisedValue:                         request.AppraisedValue,
			AppraisedValueDescription:              request.AppraisedValueDescription,
			PrintedDate:                            request.PrintedDate,
			ApprovedDate:                           request.ApprovedDate,
			ReleasedDate:                           request.ReleasedDate,
			ApprovedBySignatureMediaID:             request.ApprovedBySignatureMediaID,
			ApprovedByName:                         request.ApprovedByName,
			ApprovedByPosition:                     request.ApprovedByPosition,
			PreparedBySignatureMediaID:             request.PreparedBySignatureMediaID,
			PreparedByName:                         request.PreparedByName,
			PreparedByPosition:                     request.PreparedByPosition,
			CertifiedBySignatureMediaID:            request.CertifiedBySignatureMediaID,
			CertifiedByName:                        request.CertifiedByName,
			CertifiedByPosition:                    request.CertifiedByPosition,
			VerifiedBySignatureMediaID:             request.VerifiedBySignatureMediaID,
			VerifiedByName:                         request.VerifiedByName,
			VerifiedByPosition:                     request.VerifiedByPosition,
			CheckBySignatureMediaID:                request.CheckBySignatureMediaID,
			CheckByName:                            request.CheckByName,
			CheckByPosition:                        request.CheckByPosition,
			AcknowledgeBySignatureMediaID:          request.AcknowledgeBySignatureMediaID,
			AcknowledgeByName:                      request.AcknowledgeByName,
			AcknowledgeByPosition:                  request.AcknowledgeByPosition,
			NotedBySignatureMediaID:                request.NotedBySignatureMediaID,
			NotedByName:                            request.NotedByName,
			NotedByPosition:                        request.NotedByPosition,
			PostedBySignatureMediaID:               request.PostedBySignatureMediaID,
			PostedByName:                           request.PostedByName,
			PostedByPosition:                       request.PostedByPosition,
			PaidBySignatureMediaID:                 request.PaidBySignatureMediaID,
			PaidByName:                             request.PaidByName,
			PaidByPosition:                         request.PaidByPosition,
		}

		if err := c.model.LoanTransactionManager.CreateWithTx(context, tx, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction: " + err.Error()})
		}
		cashOnHandTransactionEntry := &model.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			AccountID:         userOrg.Branch.BranchSetting.CashOnHandAccountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: loanTransaction.ID,
			Credit:            request.Applied1,
			Debit:             0,
			Description:       "Loan Disbursement",
			Index:             0,
		}

		if err := c.model.LoanTransactionEntryManager.CreateWithTx(context, tx, cashOnHandTransactionEntry); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction entry: " + err.Error()})
		}

		loanTransactionEntry := &model.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			AccountID:         request.AccountID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: loanTransaction.ID,
			Credit:            0,
			Debit:             request.Applied1,
			Description:       "Loan Disbursement",
			Index:             1,
		}
		if err := c.model.LoanTransactionEntryManager.CreateWithTx(context, tx, loanTransactionEntry); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction entry: " + err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/browse-exclude-include-accounts/bulk-delete), commit error: " + err.Error(),
				Module:      "BrowseExcludeIncludeAccounts",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.LoanTransactionManager.ToModel(loanTransaction))
	})

	// PUT /api/v1/loan-transaction/:id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "PUT",
		ResponseType: model.LoanTransactionResponse{},
		Note:         "Updates an existing loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transactions"})
		}

		request, err := c.model.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		// Update fields
		loanTransaction.UpdatedByID = userOrg.UserID
		loanTransaction.TransactionBatchID = &transactionBatch.ID
		loanTransaction.OfficialReceiptNumber = request.OfficialReceiptNumber
		loanTransaction.Voucher = request.Voucher
		loanTransaction.EmployeeUserID = &userOrg.UserID
		loanTransaction.LoanPurposeID = request.LoanPurposeID
		loanTransaction.LoanStatusID = request.LoanStatusID
		loanTransaction.ModeOfPayment = string(request.ModeOfPayment)
		loanTransaction.ModeOfPaymentWeekly = string(request.ModeOfPaymentWeekly)
		loanTransaction.ModeOfPaymentSemiMonthlyPay1 = request.ModeOfPaymentSemiMonthlyPay1
		loanTransaction.ModeOfPaymentSemiMonthlyPay2 = request.ModeOfPaymentSemiMonthlyPay2
		loanTransaction.ComakerType = string(request.ComakerType)
		loanTransaction.ComakerDepositMemberAccountingLedgerID = request.ComakerDepositMemberAccountingLedgerID
		loanTransaction.ComakerCollateralID = request.ComakerCollateralID
		loanTransaction.ComakerCollateralDescription = request.ComakerCollateralDescription
		loanTransaction.CollectorPlace = string(request.CollectorPlace)
		loanTransaction.LoanType = string(request.LoanType)
		loanTransaction.PreviousLoanID = request.PreviousLoanID
		loanTransaction.Terms = request.Terms
		loanTransaction.AmortizationAmount = request.AmortizationAmount
		loanTransaction.IsAddOn = request.IsAddOn
		loanTransaction.Applied1 = request.Applied1
		loanTransaction.Applied2 = request.Applied2
		loanTransaction.AccountID = request.AccountID
		loanTransaction.MemberProfileID = request.MemberProfileID
		loanTransaction.MemberJointAccountID = request.MemberJointAccountID
		loanTransaction.SignatureMediaID = request.SignatureMediaID
		loanTransaction.MountToBeClosed = request.MountToBeClosed
		loanTransaction.DamayanFund = request.DamayanFund
		loanTransaction.ShareCapital = request.ShareCapital
		loanTransaction.LengthOfService = request.LengthOfService
		loanTransaction.ExcludeSunday = request.ExcludeSunday
		loanTransaction.ExcludeHoliday = request.ExcludeHoliday
		loanTransaction.ExcludeSaturday = request.ExcludeSaturday
		loanTransaction.RemarksOtherTerms = request.RemarksOtherTerms
		loanTransaction.RemarksPayrollDeduction = request.RemarksPayrollDeduction
		loanTransaction.RecordOfLoanPaymentsOrLoanStatus = request.RecordOfLoanPaymentsOrLoanStatus
		loanTransaction.CollateralOffered = request.CollateralOffered
		loanTransaction.AppraisedValue = request.AppraisedValue
		loanTransaction.AppraisedValueDescription = request.AppraisedValueDescription
		loanTransaction.PrintedDate = request.PrintedDate
		loanTransaction.ApprovedDate = request.ApprovedDate
		loanTransaction.ReleasedDate = request.ReleasedDate
		loanTransaction.ApprovedBySignatureMediaID = request.ApprovedBySignatureMediaID
		loanTransaction.ApprovedByName = request.ApprovedByName
		loanTransaction.ApprovedByPosition = request.ApprovedByPosition
		loanTransaction.PreparedBySignatureMediaID = request.PreparedBySignatureMediaID
		loanTransaction.PreparedByName = request.PreparedByName
		loanTransaction.PreparedByPosition = request.PreparedByPosition
		loanTransaction.CertifiedBySignatureMediaID = request.CertifiedBySignatureMediaID
		loanTransaction.CertifiedByName = request.CertifiedByName
		loanTransaction.CertifiedByPosition = request.CertifiedByPosition
		loanTransaction.VerifiedBySignatureMediaID = request.VerifiedBySignatureMediaID
		loanTransaction.VerifiedByName = request.VerifiedByName
		loanTransaction.VerifiedByPosition = request.VerifiedByPosition
		loanTransaction.CheckBySignatureMediaID = request.CheckBySignatureMediaID
		loanTransaction.CheckByName = request.CheckByName
		loanTransaction.CheckByPosition = request.CheckByPosition
		loanTransaction.AcknowledgeBySignatureMediaID = request.AcknowledgeBySignatureMediaID
		loanTransaction.AcknowledgeByName = request.AcknowledgeByName
		loanTransaction.AcknowledgeByPosition = request.AcknowledgeByPosition
		loanTransaction.NotedBySignatureMediaID = request.NotedBySignatureMediaID
		loanTransaction.NotedByName = request.NotedByName
		loanTransaction.NotedByPosition = request.NotedByPosition
		loanTransaction.PostedBySignatureMediaID = request.PostedBySignatureMediaID
		loanTransaction.PostedByName = request.PostedByName
		loanTransaction.PostedByPosition = request.PostedByPosition
		loanTransaction.PaidBySignatureMediaID = request.PaidBySignatureMediaID
		loanTransaction.PaidByName = request.PaidByName
		loanTransaction.PaidByPosition = request.PaidByPosition
		loanTransaction.UpdatedAt = time.Now().UTC()

		if err := c.model.LoanTransactionManager.Update(context, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionManager.ToModel(loanTransaction))
	})

	// DELETE /api/v1/loan-transaction/:id
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/loan-transaction/:loan_transaction_id",
		Method: "DELETE",
		Note:   "Deletes a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Set deleted by user
		loanTransaction.DeletedByID = &userOrg.UserID

		if err := c.model.LoanTransactionManager.Delete(context, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction deleted successfully"})
	})
}
