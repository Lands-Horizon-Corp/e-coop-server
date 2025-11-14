package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// LoanTransactionTotalResponse represents the total calculations for a loan transaction

func (c *Controller) loanTransactionController() {
	req := c.provider.Service.Request

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/account/:account_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/account/:account_id",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
		Note:         "Returns the latest loan transaction for a specific member profile and account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		loanTransactions, err := c.core.LoanTransactionsMemberAccount(
			context, *memberProfileID, *accountID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/search",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for the current user's branch with pagination and filtering. Query params: has_print_date, has_approved_date, has_release_date (true/false)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.core.LoanTransactionManager.PaginationWithFields(context, ctx, &core.LoanTransaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.core.LoanTransactionManager.PaginationWithFields(context, ctx, &core.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.core.LoanTransactionManager.FindRaw(context, &core.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/release/search",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}
		loanTransactions, err := c.core.LoanTransactionWithDatesNotNull(
			context, *memberProfileID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	// GET /api/v1/loan-transaction/draft
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/draft",
		Method:       "GET",
		Note:         "Fetches draft loan transactions for the current user's organization and branch.",
		ResponseType: core.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "draft-error",
				Description: "Loan transaction draft failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.core.LoanTransactionDraft(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/printed
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/printed",
		Method:       "GET",
		Note:         "Fetches printed loan transactions for the current user's organization and branch.",
		ResponseType: core.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Loan transaction printed fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.core.LoanTransactionPrinted(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/approved
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/approved",
		Method:       "GET",
		Note:         "Fetches approved loan transactions for the current user's organization and branch.",
		ResponseType: core.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Loan transaction approved fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.core.LoanTransactionApproved(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/released/today
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/released/today",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: core.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.core.LoanTransactionReleasedCurrentDay(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})
	// GET /api/v1/loan-transaction/released
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/released",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: core.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.core.LoanTransactionReleased(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: core.LoanTransactionResponse{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(loanTransaction))
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/total",
		Method:       "GET",
		ResponseType: core.LoanTransactionTotalResponse{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction totals"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Get all loan transaction entries for this loan transaction
		loanTransactionEntries, err := c.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
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
			totalCredit = c.provider.Service.Decimal.Add(totalCredit, entry.Credit)
			totalDebit = c.provider.Service.Decimal.Add(totalDebit, entry.Debit)
		}

		// Calculate total interest (assuming interest is the difference between debit and credit)
		totalInterest := c.provider.Service.Decimal.Subtract(totalDebit, totalCredit)

		return ctx.JSON(http.StatusOK, core.LoanTransactionTotalResponse{
			TotalInterest: totalInterest,
			TotalDebit:    totalDebit,
			TotalCredit:   totalCredit,
		})
	})

	// POST /api/v1/loan-transaction
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction",
		Method:       "POST",
		ResponseType: core.LoanTransactionResponse{},
		Note:         "Creates a new loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transactions"})
		}

		request, err := c.core.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		loanTransaction := &core.LoanTransaction{
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
			ModeOfPayment:                          request.ModeOfPayment,
			ModeOfPaymentWeekly:                    request.ModeOfPaymentWeekly,
			ModeOfPaymentSemiMonthlyPay1:           request.ModeOfPaymentSemiMonthlyPay1,
			ModeOfPaymentSemiMonthlyPay2:           request.ModeOfPaymentSemiMonthlyPay2,
			ComakerType:                            request.ComakerType,
			ComakerDepositMemberAccountingLedgerID: request.ComakerDepositMemberAccountingLedgerID,
			CollectorPlace:                         request.CollectorPlace,
			LoanType:                               request.LoanType,
			PreviousLoanID:                         request.PreviousLoanID,
			Terms:                                  request.Terms,
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
			ModeOfPaymentFixedDays:                 request.ModeOfPaymentFixedDays,
			TotalCredit:                            request.Applied1,
			TotalDebit:                             request.Applied1,
			ModeOfPaymentMonthlyExactDay:           request.ModeOfPaymentMonthlyExactDay,
		}

		if err := c.core.LoanTransactionManager.CreateWithTx(context, tx, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction: " + endTx(err).Error()})
		}
		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandAccountID == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cash on hand account is not set for the branch: " + endTx(eris.New("cash on hand account not set")).Error()})
		}
		if err := c.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + endTx(err).Error()})
		}
		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				clearanceAnalysis := &core.LoanClearanceAnalysis{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedByID:    userOrg.UserID,
					OrganizationID: userOrg.OrganizationID,
					BranchID:       *userOrg.BranchID,

					RegularDeductionDescription: clearanceAnalysisReq.RegularDeductionDescription,
					RegularDeductionAmount:      clearanceAnalysisReq.RegularDeductionAmount,
					BalancesDescription:         clearanceAnalysisReq.BalancesDescription,
					BalancesAmount:              clearanceAnalysisReq.BalancesAmount,
					BalancesCount:               clearanceAnalysisReq.BalancesCount,
				}

				if err := c.core.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(created) Failed to create loan clearance analysis: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				institution := &core.LoanClearanceAnalysisInstitution{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					LoanTransactionID: loanTransaction.ID,
					Name:              institutionReq.Name,
					Description:       institutionReq.Description,
				}

				if err := c.core.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				suggestedPayment := &core.LoanTermsAndConditionSuggestedPayment{

					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					LoanTransactionID: loanTransaction.ID,
					Name:              suggestedPaymentReq.Name,
					Description:       suggestedPaymentReq.Description,
				}

				if err := c.core.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition suggested payment: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				amountReceipt := &core.LoanTermsAndConditionAmountReceipt{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					LoanTransactionID: loanTransaction.ID,
					AccountID:         amountReceiptReq.AccountID,
					Amount:            amountReceiptReq.Amount,
				}

				if err := c.core.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition amount receipt: " + endTx(err).Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles
		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				comakerMemberProfile := &core.ComakerMemberProfile{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					LoanTransactionID: loanTransaction.ID,
					MemberProfileID:   comakerReq.MemberProfileID,
					Amount:            comakerReq.Amount,
					Description:       comakerReq.Description,
					MonthsCount:       comakerReq.MonthsCount,
					YearCount:         comakerReq.YearCount,
				}

				if err := c.core.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + endTx(err).Error()})
				}
			}
		}

		// Handle ComakerCollaterals
		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				comakerCollateral := &core.ComakerCollateral{
					CreatedAt:         time.Now().UTC(),
					UpdatedAt:         time.Now().UTC(),
					CreatedByID:       userOrg.UserID,
					UpdatedByID:       userOrg.UserID,
					OrganizationID:    userOrg.OrganizationID,
					BranchID:          *userOrg.BranchID,
					LoanTransactionID: loanTransaction.ID,
					CollateralID:      comakerReq.CollateralID,
					Amount:            comakerReq.Amount,
					Description:       comakerReq.Description,
					MonthsCount:       comakerReq.MonthsCount,
					YearCount:         comakerReq.YearCount,
				}

				if err := c.core.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + endTx(err).Error()})
				}
			}
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		newTx, newEndTx := c.provider.Service.Database.StartTransaction(context)
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, newTx, newEndTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: *cashOnHandAccountID,
			LoanTransactionID:              loanTransaction.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v: %v", err, newEndTx(err))})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(newLoanTransaction))
	})

	// PUT /api/v1/loan-transaction/:id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "PUT",
		ResponseType: core.LoanTransactionResponse{},
		RequestType:  core.LoanTransactionRequest{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transactions"})
		}

		request, err := c.core.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *request.AccountID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		transactionBatch, err := c.core.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		cashOnHandAccount, err := c.core.GetCashOnCashEquivalence(
			context, *loanTransactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash on cash equivalence account: " + endTx(err).Error()})
		}
		cashOnCashEquivalenceAccountID := cashOnHandAccount.ID
		if !handlers.UUIDPtrEqual(account.CurrencyID, &cashOnCashEquivalenceAccountID) {
			accounts, err := c.core.AccountManager.Find(context, &core.Account{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CurrencyID:             account.CurrencyID,
				CashAndCashEquivalence: true,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts for currency conversion: " + endTx(err).Error()})
			}
			if len(accounts) == 0 {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "No account found for currency conversion: " + endTx(eris.New("no account found")).Error()})
			}
			cashOnCashEquivalenceAccountID = accounts[0].ID
		}

		// Update fields
		loanTransaction.AccountID = request.AccountID
		loanTransaction.UpdatedByID = userOrg.UserID
		loanTransaction.TransactionBatchID = &transactionBatch.ID
		loanTransaction.OfficialReceiptNumber = request.OfficialReceiptNumber
		loanTransaction.Voucher = request.Voucher
		loanTransaction.EmployeeUserID = &userOrg.UserID
		loanTransaction.LoanPurposeID = request.LoanPurposeID
		loanTransaction.LoanStatusID = request.LoanStatusID
		loanTransaction.ModeOfPayment = request.ModeOfPayment
		loanTransaction.ModeOfPaymentWeekly = request.ModeOfPaymentWeekly
		loanTransaction.ModeOfPaymentSemiMonthlyPay1 = request.ModeOfPaymentSemiMonthlyPay1
		loanTransaction.ModeOfPaymentSemiMonthlyPay2 = request.ModeOfPaymentSemiMonthlyPay2
		loanTransaction.ComakerType = request.ComakerType
		loanTransaction.ComakerDepositMemberAccountingLedgerID = request.ComakerDepositMemberAccountingLedgerID
		loanTransaction.CollectorPlace = request.CollectorPlace
		loanTransaction.LoanType = request.LoanType
		loanTransaction.PreviousLoanID = request.PreviousLoanID
		loanTransaction.Terms = request.Terms
		loanTransaction.IsAddOn = request.IsAddOn
		loanTransaction.Applied1 = request.Applied1
		loanTransaction.Applied2 = request.Applied2
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
		loanTransaction.ModeOfPaymentFixedDays = request.ModeOfPaymentFixedDays
		loanTransaction.ModeOfPaymentMonthlyExactDay = request.ModeOfPaymentMonthlyExactDay
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.PreviousLoanID = request.PreviousLoanID

		// Handle deletions first (same as before)
		if request.LoanClearanceAnalysisDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisDeleted {
				clearanceAnalysis, err := c.core.LoanClearanceAnalysisManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis for deletion: " + endTx(err).Error()})
				}
				if clearanceAnalysis.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				clearanceAnalysis.DeletedByID = &userOrg.UserID
				if err := c.core.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanClearanceAnalysisInstitutionDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisInstitutionDeleted {
				institution, err := c.core.LoanClearanceAnalysisInstitutionManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution for deletion: " + endTx(err).Error()})
				}
				if institution.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis institution that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				institution.DeletedByID = &userOrg.UserID
				if err := c.core.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanTermsAndConditionSuggestedPaymentDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionSuggestedPaymentDeleted {
				suggestedPayment, err := c.core.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment for deletion: " + endTx(err).Error()})
				}
				if suggestedPayment.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms suggested payment that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				suggestedPayment.DeletedByID = &userOrg.UserID
				if err := c.core.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanTermsAndConditionAmountReceiptDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionAmountReceiptDeleted {
				amountReceipt, err := c.core.LoanTermsAndConditionAmountReceiptManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt for deletion: " + endTx(err).Error()})
				}
				if amountReceipt.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms amount receipt that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				amountReceipt.DeletedByID = &userOrg.UserID
				if err := c.core.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + endTx(err).Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles deletions
		if request.ComakerMemberProfilesDeleted != nil {
			for _, deletedID := range request.ComakerMemberProfilesDeleted {
				comakerMemberProfile, err := c.core.ComakerMemberProfileManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile for deletion: " + endTx(err).Error()})
				}
				if comakerMemberProfile.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker member profile that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				comakerMemberProfile.DeletedByID = &userOrg.UserID
				if err := c.core.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + endTx(err).Error()})
				}
			}
		}

		// Handle ComakerCollaterals deletions
		if request.ComakerCollateralsDeleted != nil {
			for _, deletedID := range request.ComakerCollateralsDeleted {
				comakerCollateral, err := c.core.ComakerCollateralManager.GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker collateral for deletion: " + endTx(err).Error()})
				}
				if comakerCollateral.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker collateral that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				comakerCollateral.DeletedByID = &userOrg.UserID
				if err := c.core.ComakerCollateralManager.DeleteWithTx(context, tx, comakerCollateral.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker collateral: " + endTx(err).Error()})
				}
			}
		}

		// Create/Update LoanClearanceAnalysis records
		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				if clearanceAnalysisReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.LoanClearanceAnalysisManager.GetByID(context, *clearanceAnalysisReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.RegularDeductionDescription = clearanceAnalysisReq.RegularDeductionDescription
					existingRecord.RegularDeductionAmount = clearanceAnalysisReq.RegularDeductionAmount
					existingRecord.BalancesDescription = clearanceAnalysisReq.BalancesDescription
					existingRecord.BalancesAmount = clearanceAnalysisReq.BalancesAmount
					existingRecord.BalancesCount = clearanceAnalysisReq.BalancesCount

					if err := c.core.LoanClearanceAnalysisManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					clearanceAnalysis := &core.LoanClearanceAnalysis{
						CreatedAt:                   time.Now().UTC(),
						UpdatedAt:                   time.Now().UTC(),
						CreatedByID:                 userOrg.UserID,
						UpdatedByID:                 userOrg.UserID,
						OrganizationID:              userOrg.OrganizationID,
						BranchID:                    *userOrg.BranchID,
						LoanTransactionID:           loanTransaction.ID,
						RegularDeductionDescription: clearanceAnalysisReq.RegularDeductionDescription,
						RegularDeductionAmount:      clearanceAnalysisReq.RegularDeductionAmount,
						BalancesDescription:         clearanceAnalysisReq.BalancesDescription,
						BalancesAmount:              clearanceAnalysisReq.BalancesAmount,
						BalancesCount:               clearanceAnalysisReq.BalancesCount,
					}

					if err := c.core.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(updated) Failed to create loan clearance analysis: " + endTx(err).Error()})
					}
				}
			}
		}

		// Create/Update LoanClearanceAnalysisInstitution records
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				if institutionReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.LoanClearanceAnalysisInstitutionManager.GetByID(context, *institutionReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis institution: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis institution that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = institutionReq.Name
					existingRecord.Description = institutionReq.Description

					if err := c.core.LoanClearanceAnalysisInstitutionManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis institution: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					institution := &core.LoanClearanceAnalysisInstitution{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						Name:              institutionReq.Name,
						Description:       institutionReq.Description,
					}

					if err := c.core.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + endTx(err).Error()})
					}
				}
			}
		}

		// Create/Update LoanTermsAndConditionSuggestedPayment records
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				if suggestedPaymentReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, *suggestedPaymentReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms suggested payment: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan terms suggested payment that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = suggestedPaymentReq.Name
					existingRecord.Description = suggestedPaymentReq.Description

					if err := c.core.LoanTermsAndConditionSuggestedPaymentManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms suggested payment: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					suggestedPayment := &core.LoanTermsAndConditionSuggestedPayment{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						Name:              suggestedPaymentReq.Name,
						Description:       suggestedPaymentReq.Description,
					}

					if err := c.core.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms suggested payment: " + endTx(err).Error()})
					}
				}
			}
		}

		// Create/Update LoanTermsAndConditionAmountReceipt records
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				if amountReceiptReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.LoanTermsAndConditionAmountReceiptManager.GetByID(context, *amountReceiptReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms amount receipt: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot update loan terms amount receipt that doesn't belong to this loan transaction: " + endTx(eris.New("cannot update loan terms amount receipt that doesn't belong to this loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.AccountID = amountReceiptReq.AccountID
					existingRecord.Amount = amountReceiptReq.Amount

					if err := c.core.LoanTermsAndConditionAmountReceiptManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms amount receipt: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					amountReceipt := &core.LoanTermsAndConditionAmountReceipt{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						AccountID:         amountReceiptReq.AccountID,
						Amount:            amountReceiptReq.Amount,
					}

					if err := c.core.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms amount receipt: " + endTx(err).Error()})
					}
				}
			}
		}

		// Create/Update ComakerMemberProfile records
		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				if comakerReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.ComakerMemberProfileManager.GetByID(context, *comakerReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker member profile: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update comaker member profile that doesn't belong to this loan transaction: " + endTx(eris.New("Cannot update comaker member profile that doesn't belong to this loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.MemberProfileID = comakerReq.MemberProfileID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := c.core.ComakerMemberProfileManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker member profile: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					comakerMemberProfile := &core.ComakerMemberProfile{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						MemberProfileID:   comakerReq.MemberProfileID,
						Amount:            comakerReq.Amount,
						Description:       comakerReq.Description,
						MonthsCount:       comakerReq.MonthsCount,
						YearCount:         comakerReq.YearCount,
					}

					if err := c.core.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + endTx(err).Error()})
					}
				}
			}
		}

		// Create/Update ComakerCollateral records
		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				if comakerReq.ID != nil {
					// Update existing record
					existingRecord, err := c.core.ComakerCollateralManager.GetByID(context, *comakerReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker collateral: " + endTx(err).Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden,
							map[string]string{"error": "Cannot update comaker collateral that doesn't belong to this loan transaction: " + endTx(eris.New("Cannot update comaker collateral that doesn't belong to this loan transaction")).Error()})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.CollateralID = comakerReq.CollateralID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := c.core.ComakerCollateralManager.UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker collateral: " + endTx(err).Error()})
					}
				} else {
					// Create new record
					comakerCollateral := &core.ComakerCollateral{
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						CollateralID:      comakerReq.CollateralID,
						Amount:            comakerReq.Amount,
						Description:       comakerReq.Description,
						MonthsCount:       comakerReq.MonthsCount,
						YearCount:         comakerReq.YearCount,
					}

					if err := c.core.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + endTx(err).Error()})
					}
				}
			}
		}
		if !handlers.UUIDPtrEqual(account.CurrencyID, loanTransaction.Account.CurrencyID) {
			loanTransactionEntries, err := c.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + endTx(err).Error()})
			}
			// Process currency conversion for each loan transaction entry
			for _, entry := range loanTransactionEntries {
				if err := c.core.LoanTransactionEntryManager.Delete(context, entry.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + endTx(err).Error()})
				}
			}
		} else {
			loanTransactionEntry, err := c.core.GetLoanEntryAccount(context, loanTransaction.ID, userOrg.OrganizationID, *userOrg.BranchID)
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to find loan transaction entry (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change): " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry: " + endTx(err).Error()})
			}
			loanTransactionEntry.AccountID = &account.ID
			loanTransactionEntry.Name = account.Name
			loanTransactionEntry.Description = account.Description
			if err := c.core.LoanTransactionEntryManager.UpdateByIDWithTx(context, tx, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
			}

		}

		if err := c.core.LoanTransactionManager.UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		newTx, newEndTx := c.provider.Service.Database.StartTransaction(context)
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, newTx, newEndTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: cashOnCashEquivalenceAccountID,
			LoanTransactionID:              loanTransaction.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to retrieve updated loan transaction: %v: %v", err, newEndTx(err))})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(newLoanTransaction))
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Start transaction for cascading deletes
		tx, endTx := c.provider.Service.Database.StartTransaction(context)

		// Delete all LoanClearanceAnalysis records
		clearanceAnalysisList, err := c.core.LoanClearanceAnalysisManager.Find(context, &core.LoanClearanceAnalysis{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis records: " + endTx(err).Error()})
		}

		for _, clearanceAnalysis := range clearanceAnalysisList {
			clearanceAnalysis.DeletedByID = &userOrg.UserID
			if err := c.core.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + endTx(err).Error()})
			}
		}

		// Delete all LoanClearanceAnalysisInstitution records
		institutionList, err := c.core.LoanClearanceAnalysisInstitutionManager.Find(context, &core.LoanClearanceAnalysisInstitution{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution records: " + endTx(err).Error()})
		}

		for _, institution := range institutionList {
			institution.DeletedByID = &userOrg.UserID
			if err := c.core.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + endTx(err).Error()})
			}
		}

		// Delete all LoanTermsAndConditionSuggestedPayment records
		suggestedPaymentList, err := c.core.LoanTermsAndConditionSuggestedPaymentManager.Find(context, &core.LoanTermsAndConditionSuggestedPayment{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment records: " + endTx(err).Error()})
		}

		for _, suggestedPayment := range suggestedPaymentList {
			suggestedPayment.DeletedByID = &userOrg.UserID
			if err := c.core.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + endTx(err).Error()})
			}
		}

		// Delete all LoanTermsAndConditionAmountReceipt records
		amountReceiptList, err := c.core.LoanTermsAndConditionAmountReceiptManager.Find(context, &core.LoanTermsAndConditionAmountReceipt{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt records: " + endTx(err).Error()})
		}

		for _, amountReceipt := range amountReceiptList {
			amountReceipt.DeletedByID = &userOrg.UserID
			if err := c.core.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + endTx(err).Error()})
			}
		}

		// Delete all LoanTransactionEntry records
		transactionEntryList, err := c.core.LoanTransactionEntryManager.Find(context, &core.LoanTransactionEntry{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry records: " + endTx(err).Error()})
		}

		for _, transactionEntry := range transactionEntryList {
			transactionEntry.DeletedByID = &userOrg.UserID
			if err := c.core.LoanTransactionEntryManager.DeleteWithTx(context, tx, transactionEntry.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + endTx(err).Error()})
			}
		}

		// Delete all ComakerMemberProfile records
		comakerMemberProfileList, err := c.core.ComakerMemberProfileManager.Find(context, &core.ComakerMemberProfile{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile records: " + endTx(err).Error()})
		}

		for _, comakerMemberProfile := range comakerMemberProfileList {
			comakerMemberProfile.DeletedByID = &userOrg.UserID
			if err := c.core.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + endTx(err).Error()})
			}
		}

		// Set deleted by user for main loan transaction
		loanTransaction.DeletedByID = &userOrg.UserID

		// Delete the main loan transaction
		if err := c.core.LoanTransactionManager.DeleteWithTx(context, tx, loanTransaction.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction: " + endTx(err).Error()})
		}

		// Commit the transaction
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Successfully deleted loan transaction %s and all related records", loanTransaction.ID),
			Module:      "LoanTransaction",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction and all related records deleted successfully"})
	})

	// Simplified bulk-delete handler for loan transactions.
	// Keeps authorization in the handler but moves heavy deletion logic into the manager.
	// Expects c.core.LoanTransactionManager.BulkDeleteWithOrg(ctx, ids, userOrg) or similar to exist.
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/loan-transaction/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan transactions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | no IDs provided",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Authorization / org+branch resolution stays in the handler
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | auth/organization error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | unauthorized user type",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		// Delegate complex deletion (transaction, related records, ownership checks) to the manager.
		// Manager signature assumed: BulkDeleteWithOrg(ctx context.Context, ids []string, userOrg core.UserOrganization) error
		if err := c.core.LoanTransactionManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			// Manager should return appropriate error types (validation/not-found/forbidden/internal)
			// Map manager errors to HTTP status as needed (here we return 500 for generic failure).
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan transactions: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan transactions (/loan-transaction/bulk-delete)",
			Module:      "LoanTransaction",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// PUT /api/v1/loan-transaction/:loan_transaction_id/print
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed by ID.",
		RequestType:  core.LoanTransactionPrintRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionPrintRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction print request: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print loan transactions"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.PrintedDate != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Loan transaction has already been marked printed, you can undo it by clicking undo print"})
		}
		loanTransaction.PrintNumber++
		timeNow := time.Now().UTC()
		if userOrg.TimeMachineTime != nil {
			timeNow = userOrg.UserOrgTime()
		}
		loanTransaction.PrintedDate = &timeNow
		loanTransaction.PrintedByID = &userOrg.UserID
		loanTransaction.Voucher = req.Voucher
		loanTransaction.CheckNumber = req.CheckNumber
		loanTransaction.CheckDate = req.CheckDate
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID

		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT - api/v1/loan-transaction/:id/print-undo
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/loan-transaction/:loan_transaction_id/print-undo",
		Method: "PUT",
		Note:   "Reverts the print status of a loan transaction by ID.",
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print on loan transactions"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.ApprovedDate != nil || loanTransaction.ReleasedDate != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot undo print on an approved or released loan transaction"})
		}
		loanTransaction.PrintedDate = nil
		loanTransaction.PrintedByID = nil
		loanTransaction.PrintNumber = 0
		loanTransaction.Voucher = ""
		loanTransaction.CheckNumber = ""
		loanTransaction.CheckDate = nil
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/:id/print-only
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print-only",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed without additional details by ID.",
		ResponseType: core.LoanTransaction{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to mark loan transactions as printed"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		loanTransaction.PrintNumber++
		loanTransaction.PrintedDate = handlers.Ptr(time.Now().UTC())
		loanTransaction.PrintedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/:id/approve\
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/approve",
		Method:       "PUT",
		Note:         "Approves a loan transaction by ID.",
		ResponseType: core.LoanTransaction{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to approve loan transactions"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction must be printed before approval"})
		}
		if loanTransaction.ApprovedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction is already approved"})
		}
		now := time.Now().UTC()
		timeNow := time.Now().UTC()
		if userOrg.TimeMachineTime != nil {
			timeNow = userOrg.UserOrgTime()
		}
		loanTransaction.ApprovedDate = &timeNow
		loanTransaction.ApprovedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Description:      fmt.Sprintf("Loan transaction has been approved by %s and is waiting to be released", *userOrg.User.FirstName),
			Title:            "Loan Transaction Approved - Pending Release",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/:id/approve-undo
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/approve-undo",
		Method:       "PUT",
		Note:         "Reverts the approval status of a loan transaction by ID.",
		ResponseType: core.LoanTransaction{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo approval on loan transactions"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot undo approval on a released loan transaction"})
		}
		if loanTransaction.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction is not approved"})
		}
		loanTransaction.ApprovedDate = nil
		loanTransaction.ApprovedByID = nil
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT - api/v1/loan-transaction/:id/release
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/release",
		Method:       "PUT",
		Note:         "Releases a loan transaction by ID. RELEASED SHOULD NOT BE UNAPPROVE",
		ResponseType: core.LoanTransaction{},
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
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to release loan transactions"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.PrintedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction must be printed before release"})
		}
		if loanTransaction.ApprovedDate == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction must be approved before release"})
		}
		if loanTransaction.ReleasedDate != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction is already released"})
		}

		newLoanTransaction, err := c.event.LoanRelease(context, ctx, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}

		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Description:      fmt.Sprintf("Loan transaction has been released by %s", *userOrg.User.FirstName),
			Title:            "Loan Transaction Released",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// Put /api/v1/loan-transaction/:loan_transaction_id/signature
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/signature",
		Method:       "PUT",
		Note:         "Updates the signature of a loan transaction by ID.",
		RequestType:  core.LoanTransactionSignatureRequest{},
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req core.LoanTransactionSignatureRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction signature request: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transaction signatures"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		loanTransaction.ApprovedBySignatureMediaID = req.ApprovedBySignatureMediaID
		loanTransaction.ApprovedByName = req.ApprovedByName
		loanTransaction.ApprovedByPosition = req.ApprovedByPosition
		loanTransaction.PreparedBySignatureMediaID = req.PreparedBySignatureMediaID
		loanTransaction.PreparedByName = req.PreparedByName
		loanTransaction.PreparedByPosition = req.PreparedByPosition
		loanTransaction.CertifiedBySignatureMediaID = req.CertifiedBySignatureMediaID
		loanTransaction.CertifiedByName = req.CertifiedByName
		loanTransaction.CertifiedByPosition = req.CertifiedByPosition
		loanTransaction.VerifiedBySignatureMediaID = req.VerifiedBySignatureMediaID
		loanTransaction.VerifiedByName = req.VerifiedByName
		loanTransaction.VerifiedByPosition = req.VerifiedByPosition
		loanTransaction.CheckBySignatureMediaID = req.CheckBySignatureMediaID
		loanTransaction.CheckByName = req.CheckByName
		loanTransaction.CheckByPosition = req.CheckByPosition
		loanTransaction.AcknowledgeBySignatureMediaID = req.AcknowledgeBySignatureMediaID
		loanTransaction.AcknowledgeByName = req.AcknowledgeByName
		loanTransaction.AcknowledgeByPosition = req.AcknowledgeByPosition
		loanTransaction.NotedBySignatureMediaID = req.NotedBySignatureMediaID
		loanTransaction.NotedByName = req.NotedByName
		loanTransaction.NotedByPosition = req.NotedByPosition
		loanTransaction.PostedBySignatureMediaID = req.PostedBySignatureMediaID
		loanTransaction.PostedByName = req.PostedByName
		loanTransaction.PostedByPosition = req.PostedByPosition
		loanTransaction.PaidBySignatureMediaID = req.PaidBySignatureMediaID
		loanTransaction.PaidByName = req.PaidByName
		loanTransaction.PaidByPosition = req.PaidByPosition
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID

		if err := c.core.LoanTransactionManager.UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/:loan_transaction_id/cash-and_cash-equivalence-account/:account_id/change
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change",
		Method:       "PUT",
		Note:         "Changes the cash and cash equivalence account for a loan transaction by ID.",
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		loanTransactionEntry, err := c.core.GetCashOnCashEquivalence(context, *loanTransactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction entry not found (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found: " + err.Error()})
		}
		loanTransactionEntry.AccountID = &account.ID
		loanTransactionEntry.Name = account.Name
		loanTransactionEntry.Description = account.Description
		if err := c.core.LoanTransactionEntryManager.UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByIDRaw(context, loanTransactionEntry.LoanTransactionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction not found after entry update (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found after entry update: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, loanTransaction)
	})

	// PUT /api/v1/loan-transaction/:loan_transaction_id/suggested/
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/suggested",
		Method:       "POST",
		RequestType:  core.LoanTransactionSuggestedRequest{},
		ResponseType: core.LoanTransactionSuggestedResponse{},
		Note:         "Updates the suggested payment details for a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.LoanTransactionSuggestedRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction suggested request: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		suggestedTerms, err := c.usecase.SuggestedNumberOfTerms(context, req.Amount, req.Principal, req.ModeOfPayment, req.FixedDays)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to calculate suggested terms: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, &core.LoanTransactionSuggestedResponse{
			Terms: suggestedTerms,
		})
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/schedule
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/schedule",
		Method:       "GET",
		ResponseType: event.LoanTransactionAmortizationResponse{},
		Note:         "Retrieves the payment schedule for a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		schedule, err := c.event.LoanAmortizationSchedule(context, ctx, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction schedule: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, schedule)
	})

	// POST /api/v1/loan-transaction/:loan_transaction_id/process
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/process",
		Method:       "POST",
		Note:         "Processes a loan transaction by ID.",
		ResponseType: core.LoanTransaction{},
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
		processedLoanTransaction, err := c.event.LoanProcessing(context, userOrg, loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.LoanTransactionManager.ToModel(processedLoanTransaction))
	})

	// POST /api/v1/loan-transaction/process
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/process",
		Method:       "POST",
		Note:         "All Loan transactions that are pending to be processed will be processed",
		ResponseType: core.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "loan-processing-started",
			Description: "Loan processing started",
			Module:      "Loan Processing",
		})
		c.event.OrganizationAdminsNotification(ctx, event.NotificationEvent{
			Title:       "Loan Processing",
			Description: "Loan processing started",
		})

		if err := c.event.ProcessAllLoans(context, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process loan transactions: " + err.Error()})
		}
		// return ctx.JSON(http.StatusOK, processedLoanTransaction)
		return ctx.NoContent(http.StatusNoContent)
	})

	// POST /api/v1/loan-transaction/:loan_transaction_id/adjustment
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/adjustment",
		Method:       "POST",
		Note:         "Creates an adjustment for a loan transaction by ID.",
		ResponseType: core.LoanTransactionAdjustmentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.LoanTransactionAdjustmentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "adjustment-create-error",
				Description: "Loan transaction adjustment creation failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction adjustment payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "adjustment-create-error",
				Description: "Loan transaction adjustment creation failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transaction adjustments"})
		}
		processedLoanTransaction, err := c.event.LoanProcessing(context, userOrg, loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process loan transaction: " + err.Error()})
		}
		if err := c.event.LoanAdjustment(context, *userOrg, processedLoanTransaction.ID, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction adjustment: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// GET api/v1/loan-transaction/:loan_transaction/summary
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/summary",
		Method:       "GET",
		Note:         "Retrieves a summary of loan transactions based on query parameters.",
		ResponseType: core.LoanTransactionSummaryResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransaction, err := c.core.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		entries, err := c.core.GeneralLedgerByLoanTransaction(
			context,
			*loanTransactionID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			LoanAccountID:  loanTransaction.AccountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		arrears := 0.0
		accountsummary := []core.LoanAccountSummaryResponse{}
		for _, entry := range accounts {
			accountHistory, err := c.core.GetAccountHistoryLatestByTimeHistory(
				context,
				entry.ID,
				entry.OrganizationID,
				entry.BranchID,
				loanTransaction.PrintedDate,
			)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account history: " + err.Error()})
			}
			accountsummary = append(accountsummary, core.LoanAccountSummaryResponse{
				AccountHistoryID:               accountHistory.ID,
				AccountHistory:                 *c.core.AccountHistoryManager.ToModel(accountHistory),
				TotalDebit:                     0,
				TotalCredit:                    0,
				Balance:                        0,
				DueDate:                        nil,
				LastPayment:                    nil,
				TotalNumberOfPayments:          0,
				TotalNumberOfDeductions:        0,
				TotalNumberOfAdditions:         0,
				TotalAccountPrincipal:          0,
				TotalAccountAdvancedPayment:    0,
				TotalAccountPrincipalPaid:      0,
				TotalAccountRemainingPrincipal: 0,
			})
		}
		return ctx.JSON(http.StatusOK, &core.LoanTransactionSummaryResponse{
			GeneralLedger:  c.core.GeneralLedgerManager.ToModels(entries),
			AccountSummary: accountsummary,
			Arrears:        arrears,
			AmountGranted:  loanTransaction.Applied1,
		})
	})
}
