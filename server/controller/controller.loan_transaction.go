package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LoanTransactionTotalResponse represents the total calculations for a loan transaction

func (c *Controller) loanTransactionController() {
	req := c.provider.Service.Request

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/account/:account_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/account/:account_id",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionResponse{},
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
		loanTransactions, err := c.modelcore.LoanTransactionsMemberAccount(
			context, *memberProfileID, *accountID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/search",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for the current user's branch with pagination and filtering. Query params: has_print_date, has_approved_date, has_release_date (true/false)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.modelcore.LoanTransactionManager.Find(context, &modelcore.LoanTransaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.Pagination(context, ctx, loanTransactions))
	})

	// GET /api/v1/loan-transaction/member-profile/:member_profile_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.modelcore.LoanTransactionManager.Find(context, &modelcore.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.Pagination(context, ctx, loanTransactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.modelcore.LoanTransactionManager.FindRaw(context, &modelcore.LoanTransaction{
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
		ResponseType: modelcore.LoanTransactionResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionWithDatesNotNull(
			context, *memberProfileID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.Pagination(context, ctx, loanTransactions))
	})

	// GET /api/v1/loan-transaction/draft
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/draft",
		Method:       "GET",
		Note:         "Fetches draft loan transactions for the current user's organization and branch.",
		ResponseType: modelcore.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "draft-error",
				Description: "Loan transaction draft failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionDraft(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/printed
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/printed",
		Method:       "GET",
		Note:         "Fetches printed loan transactions for the current user's organization and branch.",
		ResponseType: modelcore.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Loan transaction printed fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionPrinted(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/approved
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/approved",
		Method:       "GET",
		Note:         "Fetches approved loan transactions for the current user's organization and branch.",
		ResponseType: modelcore.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Loan transaction approved fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionApproved(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/released/today
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/released/today",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: modelcore.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionReleasedCurrentDay(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})
	// GET /api/v1/loan-transaction/released
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/released",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: modelcore.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := c.modelcore.LoanTransactionReleased(context, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModels(loanTransactions))
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModel(loanTransaction))
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/total
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/total",
		Method:       "GET",
		ResponseType: modelcore.LoanTransactionTotalResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction totals"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Get all loan transaction entries for this loan transaction
		loanTransactionEntries, err := c.modelcore.LoanTransactionEntryManager.Find(context, &modelcore.LoanTransactionEntry{
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

		return ctx.JSON(http.StatusOK, modelcore.LoanTransactionTotalResponse{
			TotalInterest: totalInterest,
			TotalDebit:    totalDebit,
			TotalCredit:   totalCredit,
		})
	})

	// GET /api/v1/loan-transaction/:loan_transaction_id/amortization-schedule
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/amortization-schedule",
		Method:       "GET",
		ResponseType: modelcore.AmortizationScheduleResponse{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan amortization schedule"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Generate amortization schedule
		schedule, err := c.modelcore.GenerateLoanAmortizationSchedule(context, loanTransaction)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate amortization schedule: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, schedule)
	})

	// POST /api/v1/loan-transaction
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction",
		Method:       "POST",
		ResponseType: modelcore.LoanTransactionResponse{},
		Note:         "Creates a new loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transactions"})
		}

		request, err := c.modelcore.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionBatch, err := c.modelcore.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		loanTransaction := &modelcore.LoanTransaction{
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

		if err := c.modelcore.LoanTransactionManager.CreateWithTx(context, tx, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction: " + err.Error()})
		}
		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandAccountID == nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cash on hand account is not set for the branch"})
		}
		if err := c.modelcore.LoanTransactionManager.UpdateFieldsWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				clearanceAnalysis := &modelcore.LoanClearanceAnalysis{
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

				if err := c.modelcore.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(created) Failed to create loan clearance analysis: " + err.Error()})
				}
			}
		}
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				institution := &modelcore.LoanClearanceAnalysisInstitution{
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

				if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + err.Error()})
				}
			}
		}
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				suggestedPayment := &modelcore.LoanTermsAndConditionSuggestedPayment{

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

				if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition suggested payment: " + err.Error()})
				}
			}
		}
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				amountReceipt := &modelcore.LoanTermsAndConditionAmountReceipt{
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

				if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition amount receipt: " + err.Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles
		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				comakerMemberProfile := &modelcore.ComakerMemberProfile{
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

				if err := c.modelcore.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + err.Error()})
				}
			}
		}

		// Handle ComakerCollaterals
		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				comakerCollateral := &modelcore.ComakerCollateral{
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

				if err := c.modelcore.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + err.Error()})
				}
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		newTx := c.provider.Service.Database.Client().Begin()
		if newTx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + newTx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + newTx.Error.Error()})
		}
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, newTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: *cashOnHandAccountID,
			LoanTransactionID:              loanTransaction.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModel(newLoanTransaction))
	})

	// PUT /api/v1/loan-transaction/:id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "PUT",
		ResponseType: modelcore.LoanTransactionResponse{},
		RequestType:  modelcore.LoanTransactionRequest{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transactions"})
		}

		request, err := c.modelcore.LoanTransactionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		account, err := c.modelcore.AccountManager.GetByID(context, *request.AccountID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		transactionBatch, err := c.modelcore.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		cashOnHandAccount, err := c.modelcore.GetCashOnCashEquivalence(
			context, *loanTransactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash on cash equivalence account: " + err.Error()})
		}
		cashOnCashEquivalenceAccountID := cashOnHandAccount.ID
		if !handlers.UuidPtrEqual(account.CurrencyID, &cashOnCashEquivalenceAccountID) {
			accounts, err := c.modelcore.AccountManager.Find(context, &modelcore.Account{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CurrencyID:             account.CurrencyID,
				CashAndCashEquivalence: true,
			})
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts for currency conversion: " + err.Error()})
			}
			if len(accounts) == 0 {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "No account found for currency conversion"})
			}
			cashOnCashEquivalenceAccountID = accounts[0].ID
			fmt.Println("Converted Cash on Cash Equivalence Account ID:", cashOnCashEquivalenceAccountID)
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
				clearanceAnalysis, err := c.modelcore.LoanClearanceAnalysisManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis for deletion: " + err.Error()})
				}
				if clearanceAnalysis.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis that doesn't belong to this loan transaction"})
				}
				clearanceAnalysis.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + err.Error()})
				}
			}
		}

		if request.LoanClearanceAnalysisInstitutionDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisInstitutionDeleted {
				institution, err := c.modelcore.LoanClearanceAnalysisInstitutionManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution for deletion: " + err.Error()})
				}
				if institution.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis institution that doesn't belong to this loan transaction"})
				}
				institution.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + err.Error()})
				}
			}
		}

		if request.LoanTermsAndConditionSuggestedPaymentDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionSuggestedPaymentDeleted {
				suggestedPayment, err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment for deletion: " + err.Error()})
				}
				if suggestedPayment.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms suggested payment that doesn't belong to this loan transaction"})
				}
				suggestedPayment.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + err.Error()})
				}
			}
		}

		if request.LoanTermsAndConditionAmountReceiptDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionAmountReceiptDeleted {
				amountReceipt, err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt for deletion: " + err.Error()})
				}
				if amountReceipt.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms amount receipt that doesn't belong to this loan transaction"})
				}
				amountReceipt.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + err.Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles deletions
		if request.ComakerMemberProfilesDeleted != nil {
			for _, deletedID := range request.ComakerMemberProfilesDeleted {
				comakerMemberProfile, err := c.modelcore.ComakerMemberProfileManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile for deletion: " + err.Error()})
				}
				if comakerMemberProfile.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker member profile that doesn't belong to this loan transaction"})
				}
				comakerMemberProfile.DeletedByID = &userOrg.UserID
				if err := c.modelcore.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + err.Error()})
				}
			}
		}

		// Handle ComakerCollaterals deletions
		if request.ComakerCollateralsDeleted != nil {
			for _, deletedID := range request.ComakerCollateralsDeleted {
				comakerCollateral, err := c.modelcore.ComakerCollateralManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker collateral for deletion: " + err.Error()})
				}
				if comakerCollateral.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker collateral that doesn't belong to this loan transaction"})
				}
				comakerCollateral.DeletedByID = &userOrg.UserID
				if err := c.modelcore.ComakerCollateralManager.DeleteWithTx(context, tx, comakerCollateral); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker collateral: " + err.Error()})
				}
			}
		}

		// Create/Update LoanClearanceAnalysis records
		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				if clearanceAnalysisReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.LoanClearanceAnalysisManager.GetByID(context, *clearanceAnalysisReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.RegularDeductionDescription = clearanceAnalysisReq.RegularDeductionDescription
					existingRecord.RegularDeductionAmount = clearanceAnalysisReq.RegularDeductionAmount
					existingRecord.BalancesDescription = clearanceAnalysisReq.BalancesDescription
					existingRecord.BalancesAmount = clearanceAnalysisReq.BalancesAmount
					existingRecord.BalancesCount = clearanceAnalysisReq.BalancesCount

					if err := c.modelcore.LoanClearanceAnalysisManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis: " + err.Error()})
					}
				} else {
					// Create new record
					clearanceAnalysis := &modelcore.LoanClearanceAnalysis{
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

					if err := c.modelcore.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(updated) Failed to create loan clearance analysis: " + err.Error()})
					}
				}
			}
		}

		// Create/Update LoanClearanceAnalysisInstitution records
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				if institutionReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.LoanClearanceAnalysisInstitutionManager.GetByID(context, *institutionReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis institution: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis institution that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = institutionReq.Name
					existingRecord.Description = institutionReq.Description

					if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis institution: " + err.Error()})
					}
				} else {
					// Create new record
					institution := &modelcore.LoanClearanceAnalysisInstitution{
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

					if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + err.Error()})
					}
				}
			}
		}

		// Create/Update LoanTermsAndConditionSuggestedPayment records
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				if suggestedPaymentReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, *suggestedPaymentReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms suggested payment: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan terms suggested payment that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = suggestedPaymentReq.Name
					existingRecord.Description = suggestedPaymentReq.Description

					if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms suggested payment: " + err.Error()})
					}
				} else {
					// Create new record
					suggestedPayment := &modelcore.LoanTermsAndConditionSuggestedPayment{
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

					if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms suggested payment: " + err.Error()})
					}
				}
			}
		}

		// Create/Update LoanTermsAndConditionAmountReceipt records
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				if amountReceiptReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.GetByID(context, *amountReceiptReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms amount receipt: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan terms amount receipt that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.AccountID = amountReceiptReq.AccountID
					existingRecord.Amount = amountReceiptReq.Amount

					if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms amount receipt: " + err.Error()})
					}
				} else {
					// Create new record
					amountReceipt := &modelcore.LoanTermsAndConditionAmountReceipt{
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

					if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms amount receipt: " + err.Error()})
					}
				}
			}
		}

		// Create/Update ComakerMemberProfile records
		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				if comakerReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.ComakerMemberProfileManager.GetByID(context, *comakerReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker member profile: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update comaker member profile that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.MemberProfileID = comakerReq.MemberProfileID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := c.modelcore.ComakerMemberProfileManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker member profile: " + err.Error()})
					}
				} else {
					// Create new record
					comakerMemberProfile := &modelcore.ComakerMemberProfile{
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

					if err := c.modelcore.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + err.Error()})
					}
				}
			}
		}

		// Create/Update ComakerCollateral records
		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				if comakerReq.ID != nil {
					// Update existing record
					existingRecord, err := c.modelcore.ComakerCollateralManager.GetByID(context, *comakerReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker collateral: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update comaker collateral that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.CollateralID = comakerReq.CollateralID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := c.modelcore.ComakerCollateralManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker collateral: " + err.Error()})
					}
				} else {
					// Create new record
					comakerCollateral := &modelcore.ComakerCollateral{
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

					if err := c.modelcore.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + err.Error()})
					}
				}
			}
		}
		if !handlers.UuidPtrEqual(account.CurrencyID, loanTransaction.Account.CurrencyID) {
			loanTransactionEntries, err := c.modelcore.LoanTransactionEntryManager.Find(context, &modelcore.LoanTransactionEntry{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + err.Error()})
			}
			// Process currency conversion for each loan transaction entry
			for _, entry := range loanTransactionEntries {
				if err := c.modelcore.LoanTransactionEntryManager.DeleteByID(context, entry.ID); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
				}
			}
			fmt.Println("Converted loan transaction entries due to currency change.")
		} else {
			loanTransactionEntry, err := c.modelcore.GetLoanEntryAccount(context, loanTransaction.ID, userOrg.OrganizationID, *userOrg.BranchID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to find loan transaction entry (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change): " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry: " + err.Error()})
			}
			loanTransactionEntry.AccountID = &account.ID
			loanTransactionEntry.Name = account.Name
			loanTransactionEntry.Description = account.Description
			if err := c.modelcore.LoanTransactionEntryManager.UpdateFieldsWithTx(context, tx, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
			}

		}

		if err := c.modelcore.LoanTransactionManager.UpdateFieldsWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		newTx := c.provider.Service.Database.Client().Begin()
		if newTx.Error != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + newTx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + newTx.Error.Error()})
		}
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, newTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: cashOnCashEquivalenceAccountID,
			LoanTransactionID:              loanTransaction.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.LoanTransactionManager.ToModel(newLoanTransaction))
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		// Check if the loan transaction belongs to the user's organization and branch
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		// Start transaction for cascading deletes
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		// Delete all LoanClearanceAnalysis records
		clearanceAnalysisList, err := c.modelcore.LoanClearanceAnalysisManager.Find(context, &modelcore.LoanClearanceAnalysis{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis records: " + err.Error()})
		}

		for _, clearanceAnalysis := range clearanceAnalysisList {
			clearanceAnalysis.DeletedByID = &userOrg.UserID
			if err := c.modelcore.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + err.Error()})
			}
		}

		// Delete all LoanClearanceAnalysisInstitution records
		institutionList, err := c.modelcore.LoanClearanceAnalysisInstitutionManager.Find(context, &modelcore.LoanClearanceAnalysisInstitution{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution records: " + err.Error()})
		}

		for _, institution := range institutionList {
			institution.DeletedByID = &userOrg.UserID
			if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + err.Error()})
			}
		}

		// Delete all LoanTermsAndConditionSuggestedPayment records
		suggestedPaymentList, err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.Find(context, &modelcore.LoanTermsAndConditionSuggestedPayment{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment records: " + err.Error()})
		}

		for _, suggestedPayment := range suggestedPaymentList {
			suggestedPayment.DeletedByID = &userOrg.UserID
			if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + err.Error()})
			}
		}

		// Delete all LoanTermsAndConditionAmountReceipt records
		amountReceiptList, err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.Find(context, &modelcore.LoanTermsAndConditionAmountReceipt{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt records: " + err.Error()})
		}

		for _, amountReceipt := range amountReceiptList {
			amountReceipt.DeletedByID = &userOrg.UserID
			if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + err.Error()})
			}
		}

		// Delete all LoanTransactionEntry records
		transactionEntryList, err := c.modelcore.LoanTransactionEntryManager.Find(context, &modelcore.LoanTransactionEntry{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry records: " + err.Error()})
		}

		for _, transactionEntry := range transactionEntryList {
			transactionEntry.DeletedByID = &userOrg.UserID
			if err := c.modelcore.LoanTransactionEntryManager.DeleteWithTx(context, tx, transactionEntry); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
			}
		}

		// Delete all ComakerMemberProfile records
		comakerMemberProfileList, err := c.modelcore.ComakerMemberProfileManager.Find(context, &modelcore.ComakerMemberProfile{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile records: " + err.Error()})
		}

		for _, comakerMemberProfile := range comakerMemberProfileList {
			comakerMemberProfile.DeletedByID = &userOrg.UserID
			if err := c.modelcore.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + err.Error()})
			}
		}

		// Set deleted by user for main loan transaction
		loanTransaction.DeletedByID = &userOrg.UserID

		// Delete the main loan transaction
		if err := c.modelcore.LoanTransactionManager.DeleteWithTx(context, tx, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction: " + err.Error()})
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Successfully deleted loan transaction %s and all related records", loanTransaction.ID),
			Module:      "LoanTransaction",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction and all related records deleted successfully"})
	})

	// DELETE /api/v1/loan-transaction/bulk-delete
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/loan-transaction/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan transactions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete), invalid request body.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete), no IDs provided.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No loan transaction IDs provided for bulk delete"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			loanTransactionID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), invalid UUID: " + rawID,
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}

			loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, loanTransactionID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), not found: " + rawID,
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Loan transaction not found with ID: %s", rawID)})
			}

			// Check if the loan transaction belongs to the user's organization and branch
			if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
				tx.Rollback()
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": fmt.Sprintf("Access denied to loan transaction: %s", rawID)})
			}

			names += fmt.Sprintf("LT-%s,", loanTransaction.ID.String()[:8])

			// Delete all LoanClearanceAnalysis records
			clearanceAnalysisList, err := c.modelcore.LoanClearanceAnalysisManager.Find(context, &modelcore.LoanClearanceAnalysis{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), clearance analysis find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis records: " + err.Error()})
			}

			for _, clearanceAnalysis := range clearanceAnalysisList {
				clearanceAnalysis.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), clearance analysis delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + err.Error()})
				}
			}

			// Delete all LoanClearanceAnalysisInstitution records
			institutionList, err := c.modelcore.LoanClearanceAnalysisInstitutionManager.Find(context, &modelcore.LoanClearanceAnalysisInstitution{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), institution find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution records: " + err.Error()})
			}

			for _, institution := range institutionList {
				institution.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), institution delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + err.Error()})
				}
			}

			// Delete all LoanTermsAndConditionSuggestedPayment records
			suggestedPaymentList, err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.Find(context, &modelcore.LoanTermsAndConditionSuggestedPayment{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), suggested payment find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment records: " + err.Error()})
			}

			for _, suggestedPayment := range suggestedPaymentList {
				suggestedPayment.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), suggested payment delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + err.Error()})
				}
			}

			// Delete all LoanTermsAndConditionAmountReceipt records
			amountReceiptList, err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.Find(context, &modelcore.LoanTermsAndConditionAmountReceipt{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), amount receipt find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt records: " + err.Error()})
			}

			for _, amountReceipt := range amountReceiptList {
				amountReceipt.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), amount receipt delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + err.Error()})
				}
			}

			// Delete all LoanTransactionEntry records
			transactionEntryList, err := c.modelcore.LoanTransactionEntryManager.Find(context, &modelcore.LoanTransactionEntry{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), transaction entry find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry records: " + err.Error()})
			}

			for _, transactionEntry := range transactionEntryList {
				transactionEntry.DeletedByID = &userOrg.UserID
				if err := c.modelcore.LoanTransactionEntryManager.DeleteWithTx(context, tx, transactionEntry); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), transaction entry delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
				}
			}

			// Delete all ComakerMemberProfile records
			comakerMemberProfileList, err := c.modelcore.ComakerMemberProfileManager.Find(context, &modelcore.ComakerMemberProfile{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), comaker member profile find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile records: " + err.Error()})
			}

			for _, comakerMemberProfile := range comakerMemberProfileList {
				comakerMemberProfile.DeletedByID = &userOrg.UserID
				if err := c.modelcore.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), comaker member profile delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + err.Error()})
				}
			}

			// Delete all ComakerCollateral records
			comakerCollateralList, err := c.modelcore.ComakerCollateralManager.Find(context, &modelcore.ComakerCollateral{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), comaker collateral find error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker collateral records: " + err.Error()})
			}

			for _, comakerCollateral := range comakerCollateralList {
				comakerCollateral.DeletedByID = &userOrg.UserID
				if err := c.modelcore.ComakerCollateralManager.DeleteWithTx(context, tx, comakerCollateral); err != nil {
					tx.Rollback()
					c.event.Footstep(context, ctx, event.FootstepEvent{
						Activity:    "bulk-delete-error",
						Description: "Bulk delete failed (/loan-transaction/bulk-delete), comaker collateral delete error: " + err.Error(),
						Module:      "LoanTransaction",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker collateral: " + err.Error()})
				}
			}

			// Set deleted by user for main loan transaction
			loanTransaction.DeletedByID = &userOrg.UserID

			// Delete the main loan transaction
			if err := c.modelcore.LoanTransactionManager.DeleteWithTx(context, tx, loanTransaction); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/loan-transaction/bulk-delete), main transaction delete error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction: " + err.Error()})
			}
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete), commit error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan transactions (/loan-transaction/bulk-delete): " + names,
			Module:      "LoanTransaction",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// PUT /api/v1/loan-transaction/:loan_transaction_id/print
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed by ID.",
		RequestType:  modelcore.LoanTransactionPrintRequest{},
		ResponseType: modelcore.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req modelcore.LoanTransactionPrintRequest
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print loan transactions"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.PrintedDate = handlers.Ptr(time.Now().UTC())
		loanTransaction.PrintedByID = &userOrg.UserID
		loanTransaction.Voucher = req.Voucher
		loanTransaction.CheckNumber = req.CheckNumber
		loanTransaction.CheckDate = req.CheckDate
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID

		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print on loan transactions"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: modelcore.LoanTransaction{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to mark loan transactions as printed"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: modelcore.LoanTransaction{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to approve loan transactions"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.ApprovedDate = &now
		loanTransaction.ApprovedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/:id/approve-undo
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/approve-undo",
		Method:       "PUT",
		Note:         "Reverts the approval status of a loan transaction by ID.",
		ResponseType: modelcore.LoanTransaction{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo approval on loan transactions"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: modelcore.LoanTransaction{},
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to release loan transactions"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		now := time.Now().UTC()
		loanTransaction.ReleasedDate = &now
		loanTransaction.ReleasedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return err
		}
		newLoanTransaction, err := c.event.LoanRelease(context, ctx, tx, event.LoanBalanceEvent{
			LoanTransactionID: loanTransaction.ID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// Put /api/v1/loan-transaction/:loan_transaction_id/signature
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/signature",
		Method:       "PUT",
		Note:         "Updates the signature of a loan transaction by ID.",
		RequestType:  modelcore.LoanTransactionSignatureRequest{},
		ResponseType: modelcore.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req modelcore.LoanTransactionSignatureRequest
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
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transaction signatures"})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByID(context, *loanTransactionID)
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

		if err := c.modelcore.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: modelcore.LoanTransaction{},
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
		account, err := c.modelcore.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		loanTransactionEntry, err := c.modelcore.LoanTransactionEntryManager.FindOneWithFilters(context, []services.Filter{
			{Field: "loan_transaction_entries.organization_id", Op: services.OpEq, Value: userOrg.OrganizationID},
			{Field: "loan_transaction_entries.branch_id", Op: services.OpEq, Value: userOrg.BranchID},
			{Field: "loan_transaction_entries.index", Op: services.OpEq, Value: 0},
			{Field: "loan_transaction_entries.debit", Op: services.OpEq, Value: 0},
			{Field: "loan_transaction_entries.loan_transaction_id", Op: services.OpEq, Value: loanTransactionID},
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction entry not found (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found: " + err.Error()})
		}
		loanTransactionEntry.AccountID = &account.ID
		loanTransactionEntry.Name = account.Name
		loanTransactionEntry.Description = account.Description
		if err := c.modelcore.LoanTransactionEntryManager.UpdateFields(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
		}
		loanTransaction, err := c.modelcore.LoanTransactionManager.GetByIDRaw(context, loanTransactionEntry.LoanTransactionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		RequestType:  modelcore.LoanTransactionSuggestedRequest{},
		ResponseType: modelcore.LoanTransactionSuggestedResponse{},
		Note:         "Updates the suggested payment details for a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req modelcore.LoanTransactionSuggestedRequest
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

		return ctx.JSON(http.StatusOK, &modelcore.LoanTransactionSuggestedResponse{
			Terms: suggestedTerms,
		})
	})

}
