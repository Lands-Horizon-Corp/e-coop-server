package controller_v1

import (
	"fmt"
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
		Note:         "Returns all loan transactions for the current user's branch with pagination and filtering. Query params: has_print_date, has_approved_date, has_release_date (true/false)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := c.model.LoanTransactionManager.FindRaw(context, &model.LoanTransaction{
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
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "LoanTransaction",
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
			ModeOfPaymentFixedDays:                 request.ModeOfPaymentFixedDays,
			TotalCredit:                            request.Applied1,
			TotalDebit:                             request.Applied1,
			ModeOfPaymentMonthlyExactDay:           request.ModeOfPaymentMonthlyExactDay,
		}

		if err := c.model.LoanTransactionManager.CreateWithTx(context, tx, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction: " + err.Error()})
		}
		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandAccountID == nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cash on hand account is not set for the branch"})
		}
		cashOnHand, err := c.model.AccountManager.GetByID(context, *cashOnHandAccountID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash on hand account: " + err.Error()})
		}
		account, err := c.model.AccountManager.GetByID(context, *request.AccountID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		automaticLoanDeduction, err := c.model.AutomaticLoanDeductionManager.Find(context, &model.AutomaticLoanDeduction{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: account.ComputationSheetID,
		})
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve automatic loan deduction: " + err.Error()})
		}

		loanTransactionEntries := []*model.LoanTransactionEntry{
			{
				CreatedByID:       userOrg.UserID,
				UpdatedByID:       userOrg.UserID,
				CreatedAt:         time.Now().UTC(),
				UpdatedAt:         time.Now().UTC(),
				AccountID:         &cashOnHand.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				LoanTransactionID: loanTransaction.ID,
				Credit:            request.Applied1,
				Debit:             0,
				Description:       cashOnHand.Description,
				Name:              cashOnHand.Name,
				Index:             0,
				Type:              model.LoanTransactionStatic,
			},
			{
				CreatedByID:       userOrg.UserID,
				UpdatedByID:       userOrg.UserID,
				CreatedAt:         time.Now().UTC(),
				UpdatedAt:         time.Now().UTC(),
				AccountID:         &cashOnHand.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				LoanTransactionID: loanTransaction.ID,
				Credit:            request.Applied1,
				Debit:             0,
				Description:       cashOnHand.Description,
				Name:              cashOnHand.Name,
				Index:             0,
				Type:              model.LoanTransactionStatic,
			},
		}

		addOnEntry := &model.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: loanTransaction.ID,
			Credit:            0,
			Debit:             0,
			Description:       "ADD ON Interest",
			Name:              "add on interest loan",
			Index:             1,
			Type:              model.LoanTransactionAddOn,
			IsAddOn:           true,
		}
		total_non_add_ons, total_add_ons := 0.0, 0.0
		for i, ald := range automaticLoanDeduction {
			if ald.AccountID == nil {
				continue
			}
			ald.Account, err = c.model.AccountManager.GetByID(context, *ald.AccountID)
			if err != nil {
				continue
			}

			entry := &model.LoanTransactionEntry{
				CreatedByID:       userOrg.UserID,
				UpdatedByID:       userOrg.UserID,
				CreatedAt:         time.Now().UTC(),
				UpdatedAt:         time.Now().UTC(),
				AccountID:         ald.AccountID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
				LoanTransactionID: loanTransaction.ID,
				Credit:            0,
				Debit:             0,
				Description:       ald.Description,
				Name:              ald.Name,
				Index:             i + 2,
				Type:              model.LoanTransactionStatic,
				IsAddOn:           ald.AddOn,
			}
			entry.Credit = c.service.LoanComputation(context, *ald, *loanTransaction)
			if ald.ChargesPercentage1 == 0 && ald.ChargesPercentage2 == 0 {
				if ald.Account.Type == model.AccountTypeInterest && ald.Account.InterestStandard > 0 {
					entry.Credit = request.Applied1 * (ald.Account.InterestStandard / 100) * float64(request.Terms)
				}
			}

			if loanTransaction.IsAddOn && entry.IsAddOn {
				entry.Debit += entry.Credit
			}
			if !entry.IsAddOn {
				total_non_add_ons += entry.Credit
			} else {
				total_add_ons += entry.Credit
			}
			loanTransactionEntries = append(loanTransactionEntries, entry)
		}

		if loanTransaction.IsAddOn {
			loanTransactionEntries[0].Credit = request.Applied1 - total_non_add_ons
		} else {
			loanTransactionEntries[0].Credit = request.Applied1 - (total_non_add_ons + total_add_ons)
		}
		if loanTransaction.IsAddOn {
			loanTransactionEntries = append(loanTransactionEntries, addOnEntry)
		}

		for i, entry := range loanTransactionEntries {
			entry.Index = i + 1
			if err := c.model.LoanTransactionEntryManager.CreateWithTx(context, tx, entry); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create loan transaction entry for account ID %s: %s", entry.AccountID.String(), err.Error())})
			}
		}

		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				clearanceAnalysis := &model.LoanClearanceAnalysis{
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

				if err := c.model.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(created) Failed to create loan clearance analysis: " + err.Error()})
				}
			}
		}
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				institution := &model.LoanClearanceAnalysisInstitution{
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

				if err := c.model.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + err.Error()})
				}
			}
		}
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				suggestedPayment := &model.LoanTermsAndConditionSuggestedPayment{

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

				if err := c.model.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition suggested payment: " + err.Error()})
				}
			}
		}
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				amountReceipt := &model.LoanTermsAndConditionAmountReceipt{
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

				if err := c.model.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition amount receipt: " + err.Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles
		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				comakerMemberProfile := &model.ComakerMemberProfile{
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

				if err := c.model.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + err.Error()})
				}
			}
		}

		// Handle ComakerCollaterals
		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				comakerCollateral := &model.ComakerCollateral{
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

				if err := c.model.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + err.Error()})
				}
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		loanTransactionUpdated, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactionUpdated)
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

		// Update fields
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
		loanTransaction.ModeOfPaymentFixedDays = request.ModeOfPaymentFixedDays
		loanTransaction.ModeOfPaymentMonthlyExactDay = request.ModeOfPaymentMonthlyExactDay

		loanTransaction.UpdatedAt = time.Now().UTC()

		if request.LoanTransactionEntriesDeleted != nil {
			for _, deletedID := range request.LoanTransactionEntriesDeleted {
				loanTransactionEntry, err := c.model.LoanTransactionEntryManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry for deletion: " + err.Error()})
				}
				if loanTransactionEntry.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis that doesn't belong to this loan transaction"})
				}
				loanTransactionEntry.DeletedByID = &userOrg.UserID
				if err := c.model.LoanTransactionEntryManager.DeleteWithTx(context, tx, loanTransactionEntry); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
				}
			}
		}
		// Handle deletions first (same as before)
		if request.LoanClearanceAnalysisDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisDeleted {
				clearanceAnalysis, err := c.model.LoanClearanceAnalysisManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis for deletion: " + err.Error()})
				}
				if clearanceAnalysis.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis that doesn't belong to this loan transaction"})
				}
				clearanceAnalysis.DeletedByID = &userOrg.UserID
				if err := c.model.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + err.Error()})
				}
			}
		}

		if request.LoanClearanceAnalysisInstitutionDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisInstitutionDeleted {
				institution, err := c.model.LoanClearanceAnalysisInstitutionManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution for deletion: " + err.Error()})
				}
				if institution.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis institution that doesn't belong to this loan transaction"})
				}
				institution.DeletedByID = &userOrg.UserID
				if err := c.model.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + err.Error()})
				}
			}
		}

		if request.LoanTermsAndConditionSuggestedPaymentDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionSuggestedPaymentDeleted {
				suggestedPayment, err := c.model.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment for deletion: " + err.Error()})
				}
				if suggestedPayment.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms suggested payment that doesn't belong to this loan transaction"})
				}
				suggestedPayment.DeletedByID = &userOrg.UserID
				if err := c.model.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + err.Error()})
				}
			}
		}

		if request.LoanTermsAndConditionAmountReceiptDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionAmountReceiptDeleted {
				amountReceipt, err := c.model.LoanTermsAndConditionAmountReceiptManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt for deletion: " + err.Error()})
				}
				if amountReceipt.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms amount receipt that doesn't belong to this loan transaction"})
				}
				amountReceipt.DeletedByID = &userOrg.UserID
				if err := c.model.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + err.Error()})
				}
			}
		}

		// Handle ComakerMemberProfiles deletions
		if request.ComakerMemberProfilesDeleted != nil {
			for _, deletedID := range request.ComakerMemberProfilesDeleted {
				comakerMemberProfile, err := c.model.ComakerMemberProfileManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile for deletion: " + err.Error()})
				}
				if comakerMemberProfile.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker member profile that doesn't belong to this loan transaction"})
				}
				comakerMemberProfile.DeletedByID = &userOrg.UserID
				if err := c.model.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile); err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + err.Error()})
				}
			}
		}

		// Handle ComakerCollaterals deletions
		if request.ComakerCollateralsDeleted != nil {
			for _, deletedID := range request.ComakerCollateralsDeleted {
				comakerCollateral, err := c.model.ComakerCollateralManager.GetByID(context, deletedID)
				if err != nil {
					tx.Rollback()
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker collateral for deletion: " + err.Error()})
				}
				if comakerCollateral.LoanTransactionID != loanTransaction.ID {
					tx.Rollback()
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker collateral that doesn't belong to this loan transaction"})
				}
				comakerCollateral.DeletedByID = &userOrg.UserID
				if err := c.model.ComakerCollateralManager.DeleteWithTx(context, tx, comakerCollateral); err != nil {
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
					existingRecord, err := c.model.LoanClearanceAnalysisManager.GetByID(context, *clearanceAnalysisReq.ID)
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

					if err := c.model.LoanClearanceAnalysisManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis: " + err.Error()})
					}
				} else {
					// Create new record
					clearanceAnalysis := &model.LoanClearanceAnalysis{
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

					if err := c.model.LoanClearanceAnalysisManager.CreateWithTx(context, tx, clearanceAnalysis); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(updated) Failed to create loan clearance analysis: " + err.Error()})
					}
				}
			}
		}

		if request.LoanTransactionEntries != nil {
			for _, entryReq := range request.LoanTransactionEntries {
				if entryReq.ID != nil {
					// Update existing record
					existingRecord, err := c.model.LoanTransactionEntryManager.GetByID(context, *entryReq.ID)
					if err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan transaction entry: " + err.Error()})
					}
					// Verify ownership
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						tx.Rollback()
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan transaction entry that doesn't belong to this loan transaction"})
					}
					// Update fields
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.AccountID = entryReq.AccountID
					existingRecord.Credit = entryReq.Credit
					existingRecord.Debit = entryReq.Debit
					existingRecord.IsAddOn = entryReq.IsAddOn

					// Validate AccountID if provided
					if entryReq.AccountID != nil {
						account, err := c.model.AccountManager.GetByID(context, *entryReq.AccountID)
						if err != nil {
							tx.Rollback()
							return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: account not found"})
						}

						existingRecord.Description = account.Description
						existingRecord.Name = account.Name
					} else {
						existingRecord.Description = entryReq.Description
						existingRecord.Name = entryReq.Name
					}
					existingRecord.Index = entryReq.Index
					existingRecord.Type = entryReq.Type
					if err := c.model.LoanTransactionEntryManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
					}
				} else {
					var accountName, accountDescription string

					if entryReq.AccountID != nil {
						account, err := c.model.AccountManager.GetByID(context, *entryReq.AccountID)
						if err != nil {
							accountName = entryReq.Name
							accountDescription = entryReq.Description
						} else {
							accountName = account.Name
							accountDescription = account.Description
						}
					} else {
						// No account ID provided, use provided name and description
						accountName = entryReq.Name
						accountDescription = entryReq.Description
					}

					newEntry := &model.LoanTransactionEntry{
						CreatedByID:       userOrg.UserID,
						UpdatedByID:       userOrg.UserID,
						CreatedAt:         time.Now().UTC(),
						UpdatedAt:         time.Now().UTC(),
						AccountID:         entryReq.AccountID, // This can be nil
						OrganizationID:    userOrg.OrganizationID,
						BranchID:          *userOrg.BranchID,
						LoanTransactionID: loanTransaction.ID,
						Credit:            entryReq.Credit,
						Debit:             entryReq.Debit,
						Description:       accountDescription,
						Name:              accountName,
						Index:             entryReq.Index,
						Type:              entryReq.Type,
						IsAddOn:           entryReq.IsAddOn,
					}

					if err := c.model.LoanTransactionEntryManager.CreateWithTx(context, tx, newEntry); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction entry: " + err.Error()})
					}
				}
			}
		}

		// Create/Update LoanClearanceAnalysisInstitution records
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				if institutionReq.ID != nil {
					// Update existing record
					existingRecord, err := c.model.LoanClearanceAnalysisInstitutionManager.GetByID(context, *institutionReq.ID)
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

					if err := c.model.LoanClearanceAnalysisInstitutionManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis institution: " + err.Error()})
					}
				} else {
					// Create new record
					institution := &model.LoanClearanceAnalysisInstitution{
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

					if err := c.model.LoanClearanceAnalysisInstitutionManager.CreateWithTx(context, tx, institution); err != nil {
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
					existingRecord, err := c.model.LoanTermsAndConditionSuggestedPaymentManager.GetByID(context, *suggestedPaymentReq.ID)
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

					if err := c.model.LoanTermsAndConditionSuggestedPaymentManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms suggested payment: " + err.Error()})
					}
				} else {
					// Create new record
					suggestedPayment := &model.LoanTermsAndConditionSuggestedPayment{
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

					if err := c.model.LoanTermsAndConditionSuggestedPaymentManager.CreateWithTx(context, tx, suggestedPayment); err != nil {
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
					existingRecord, err := c.model.LoanTermsAndConditionAmountReceiptManager.GetByID(context, *amountReceiptReq.ID)
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

					if err := c.model.LoanTermsAndConditionAmountReceiptManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms amount receipt: " + err.Error()})
					}
				} else {
					// Create new record
					amountReceipt := &model.LoanTermsAndConditionAmountReceipt{
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

					if err := c.model.LoanTermsAndConditionAmountReceiptManager.CreateWithTx(context, tx, amountReceipt); err != nil {
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
					existingRecord, err := c.model.ComakerMemberProfileManager.GetByID(context, *comakerReq.ID)
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

					if err := c.model.ComakerMemberProfileManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker member profile: " + err.Error()})
					}
				} else {
					// Create new record
					comakerMemberProfile := &model.ComakerMemberProfile{
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

					if err := c.model.ComakerMemberProfileManager.CreateWithTx(context, tx, comakerMemberProfile); err != nil {
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
					existingRecord, err := c.model.ComakerCollateralManager.GetByID(context, *comakerReq.ID)
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

					if err := c.model.ComakerCollateralManager.UpdateFieldsWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker collateral: " + err.Error()})
					}
				} else {
					// Create new record
					comakerCollateral := &model.ComakerCollateral{
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

					if err := c.model.ComakerCollateralManager.CreateWithTx(context, tx, comakerCollateral); err != nil {
						tx.Rollback()
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + err.Error()})
					}
				}
			}
		}

		totalCredit, totalDebit := 0.0, 0.0
		for _, entry := range request.LoanTransactionEntries {
			totalCredit += entry.Credit
			totalDebit += entry.Debit
		}

		loanTransaction.TotalCredit = totalCredit
		loanTransaction.TotalDebit = totalDebit

		if err := c.model.LoanTransactionManager.UpdateFieldsWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
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
		clearanceAnalysisList, err := c.model.LoanClearanceAnalysisManager.Find(context, &model.LoanClearanceAnalysis{
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
			if err := c.model.LoanClearanceAnalysisManager.DeleteWithTx(context, tx, clearanceAnalysis); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + err.Error()})
			}
		}

		// Delete all LoanClearanceAnalysisInstitution records
		institutionList, err := c.model.LoanClearanceAnalysisInstitutionManager.Find(context, &model.LoanClearanceAnalysisInstitution{
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
			if err := c.model.LoanClearanceAnalysisInstitutionManager.DeleteWithTx(context, tx, institution); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + err.Error()})
			}
		}

		// Delete all LoanTermsAndConditionSuggestedPayment records
		suggestedPaymentList, err := c.model.LoanTermsAndConditionSuggestedPaymentManager.Find(context, &model.LoanTermsAndConditionSuggestedPayment{
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
			if err := c.model.LoanTermsAndConditionSuggestedPaymentManager.DeleteWithTx(context, tx, suggestedPayment); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + err.Error()})
			}
		}

		// Delete all LoanTermsAndConditionAmountReceipt records
		amountReceiptList, err := c.model.LoanTermsAndConditionAmountReceiptManager.Find(context, &model.LoanTermsAndConditionAmountReceipt{
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
			if err := c.model.LoanTermsAndConditionAmountReceiptManager.DeleteWithTx(context, tx, amountReceipt); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + err.Error()})
			}
		}

		// Delete all LoanTransactionEntry records
		transactionEntryList, err := c.model.LoanTransactionEntryManager.Find(context, &model.LoanTransactionEntry{
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
			if err := c.model.LoanTransactionEntryManager.DeleteWithTx(context, tx, transactionEntry); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
			}
		}

		// Delete all ComakerMemberProfile records
		comakerMemberProfileList, err := c.model.ComakerMemberProfileManager.Find(context, &model.ComakerMemberProfile{
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
			if err := c.model.ComakerMemberProfileManager.DeleteWithTx(context, tx, comakerMemberProfile); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + err.Error()})
			}
		}

		// Set deleted by user for main loan transaction
		loanTransaction.DeletedByID = &userOrg.UserID

		// Delete the main loan transaction
		if err := c.model.LoanTransactionManager.DeleteWithTx(context, tx, loanTransaction); err != nil {
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

	// PUT /api/v1/loan-transaction/:loan_transaction_id/print
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed by ID.",
		RequestType:  model.LoanTransactionPrintRequest{},
		ResponseType: model.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req model.LoanTransactionPrintRequest
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
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print loan transactions"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		if loanTransaction.PrintedDate != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Loan transaction has already been marked printed, you can undo it by clicking undo print"})
		}
		loanTransaction.PrintNumber = loanTransaction.PrintNumber + 1
		loanTransaction.PrintedDate = handlers.Ptr(time.Now().UTC())
		loanTransaction.Voucher = req.Voucher
		loanTransaction.CheckNumber = req.CheckNumber
		loanTransaction.CheckDate = req.CheckDate
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID

		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print on loan transactions"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.PrintNumber = 0
		loanTransaction.Voucher = ""
		loanTransaction.CheckNumber = ""
		loanTransaction.CheckDate = nil
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: model.LoanTransaction{},
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to mark loan transactions as printed"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		loanTransaction.PrintNumber = loanTransaction.PrintNumber + 1
		loanTransaction.PrintedDate = handlers.Ptr(time.Now().UTC())
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: model.LoanTransaction{},
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to approve loan transactions"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: model.LoanTransaction{},
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo approval on loan transactions"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
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
		ResponseType: model.LoanTransaction{},
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to release loan transactions"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
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
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// Put /api/v1/loan-transaction/:loan_transaction_id/signature
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/signature",
		Method:       "PUT",
		Note:         "Updates the signature of a loan transaction by ID.",
		RequestType:  model.LoanTransactionSignatureRequest{},
		ResponseType: model.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req model.LoanTransactionSignatureRequest
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
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transaction signatures"})
		}
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
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

		if err := c.model.LoanTransactionManager.UpdateFields(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

}
