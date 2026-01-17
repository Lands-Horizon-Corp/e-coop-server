package loan

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func LoanTransactionController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/account/:account_id",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns the latest loan transaction for a specific member profile and account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		loanTransactions, err := core.LoanTransactionsMemberAccount(
			context, service, *memberProfileID, *accountID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/search",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for the current user's branch with pagination and filtering. Query params: has_print_date, has_approved_date, has_release_date (true/false)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := core.LoanTransactionManager(service).NormalPagination(context, ctx, &types.LoanTransaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for a specific member profile with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := core.LoanTransactionManager(service).NormalPagination(context, ctx, &types.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for a specific member profile with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransactions, err := core.LoanTransactionManager(service).FindRaw(context, &types.LoanTransaction{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/member-profile/:member_profile_id/release/search",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns all loan transactions for a specific member profile with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}
		loanTransactions, err := core.LoanTransactionWithDatesNotNull(
			context, service, *memberProfileID, *userOrg.BranchID, userOrg.OrganizationID,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transactions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanTransactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/draft",
		Method:       "GET",
		Note:         "Fetches draft loan transactions for the current user's organization and branch.",
		ResponseType: types.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "draft-error",
				Description: "Loan transaction draft failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := core.LoanTransactionDraft(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch draft loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/printed",
		Method:       "GET",
		Note:         "Fetches printed loan transactions for the current user's organization and branch.",
		ResponseType: types.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "printed-error",
				Description: "Loan transaction printed fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := core.LoanTransactionPrinted(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch printed loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/approved",
		Method:       "GET",
		Note:         "Fetches approved loan transactions for the current user's organization and branch.",
		ResponseType: types.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approved-error",
				Description: "Loan transaction approved fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := core.LoanTransactionApproved(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch approved loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/released/today",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: types.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := core.LoanTransactionReleasedCurrentDay(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/released",
		Method:       "GET",
		Note:         "Fetches released loan transactions for the current user's organization and branch.",
		ResponseType: types.LoanTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "released-error",
				Description: "Loan transaction released fetch failed, user org error.",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		loanTransactions, err := core.LoanTransactionReleased(context, service, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch released loan transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModels(loanTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "GET",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Returns a specific loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transactions"})
		}

		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(loanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/total",
		Method:       "GET",
		ResponseType: types.LoanTransactionTotalResponse{},
		Note:         "Returns total calculations for a specific loan transaction including total interest, debit, and credit.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction totals"})
		}

		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		entries, err := core.LoanTransactionEntryManager(service).Find(context, &types.LoanTransactionEntry{
			LoanTransactionID: *loanTransactionID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + err.Error()})
		}

		balance, err := usecase.CalculateBalance(usecase.Balance{
			LoanTransactionEntries: entries,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to compute loan transaction totals: %s", err.Error())})
		}
		return ctx.JSON(http.StatusOK, types.LoanTransactionTotalResponse{
			Balance:     balance.Balance,
			TotalDebit:  balance.Debit,
			TotalCredit: balance.Credit,
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction",
		Method:       "POST",
		ResponseType: types.LoanTransactionResponse{},
		Note:         "Creates a new loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transactions"})
		}

		request, err := core.LoanTransactionManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)

		loanTransaction := &types.LoanTransaction{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),

			CreatedByID:                  userOrg.UserID,
			UpdatedByID:                  userOrg.UserID,
			OrganizationID:               userOrg.OrganizationID,
			BranchID:                     *userOrg.BranchID,
			OfficialReceiptNumber:        request.OfficialReceiptNumber,
			Voucher:                      request.Voucher,
			EmployeeUserID:               &userOrg.UserID,
			LoanPurposeID:                request.LoanPurposeID,
			LoanStatusID:                 request.LoanStatusID,
			ModeOfPayment:                request.ModeOfPayment,
			ModeOfPaymentWeekly:          request.ModeOfPaymentWeekly,
			ModeOfPaymentSemiMonthlyPay1: request.ModeOfPaymentSemiMonthlyPay1,
			ModeOfPaymentSemiMonthlyPay2: request.ModeOfPaymentSemiMonthlyPay2,

			CollectorPlace:                   request.CollectorPlace,
			LoanType:                         request.LoanType,
			PreviousLoanID:                   request.PreviousLoanID,
			Terms:                            request.Terms,
			IsAddOn:                          request.IsAddOn,
			Applied1:                         request.Applied1,
			Applied2:                         request.Applied2,
			AccountID:                        request.AccountID,
			MemberProfileID:                  request.MemberProfileID,
			MemberJointAccountID:             request.MemberJointAccountID,
			SignatureMediaID:                 request.SignatureMediaID,
			MountToBeClosed:                  request.MountToBeClosed,
			DamayanFund:                      request.DamayanFund,
			ShareCapital:                     request.ShareCapital,
			LengthOfService:                  request.LengthOfService,
			ExcludeSunday:                    request.ExcludeSunday,
			ExcludeHoliday:                   request.ExcludeHoliday,
			ExcludeSaturday:                  request.ExcludeSaturday,
			RemarksOtherTerms:                request.RemarksOtherTerms,
			RemarksPayrollDeduction:          request.RemarksPayrollDeduction,
			RecordOfLoanPaymentsOrLoanStatus: request.RecordOfLoanPaymentsOrLoanStatus,
			CollateralOffered:                request.CollateralOffered,
			AppraisedValue:                   request.AppraisedValue,
			AppraisedValueDescription:        request.AppraisedValueDescription,
			PrintedDate:                      request.PrintedDate,
			ApprovedDate:                     request.ApprovedDate,
			ReleasedDate:                     request.ReleasedDate,
			ApprovedBySignatureMediaID:       request.ApprovedBySignatureMediaID,
			ApprovedByName:                   request.ApprovedByName,
			ApprovedByPosition:               request.ApprovedByPosition,
			PreparedBySignatureMediaID:       request.PreparedBySignatureMediaID,
			PreparedByName:                   request.PreparedByName,
			PreparedByPosition:               request.PreparedByPosition,
			CertifiedBySignatureMediaID:      request.CertifiedBySignatureMediaID,
			CertifiedByName:                  request.CertifiedByName,
			CertifiedByPosition:              request.CertifiedByPosition,
			VerifiedBySignatureMediaID:       request.VerifiedBySignatureMediaID,
			VerifiedByName:                   request.VerifiedByName,
			VerifiedByPosition:               request.VerifiedByPosition,
			CheckBySignatureMediaID:          request.CheckBySignatureMediaID,
			CheckByName:                      request.CheckByName,
			CheckByPosition:                  request.CheckByPosition,
			AcknowledgeBySignatureMediaID:    request.AcknowledgeBySignatureMediaID,
			AcknowledgeByName:                request.AcknowledgeByName,
			AcknowledgeByPosition:            request.AcknowledgeByPosition,
			NotedBySignatureMediaID:          request.NotedBySignatureMediaID,
			NotedByName:                      request.NotedByName,
			NotedByPosition:                  request.NotedByPosition,
			PostedBySignatureMediaID:         request.PostedBySignatureMediaID,
			PostedByName:                     request.PostedByName,
			PostedByPosition:                 request.PostedByPosition,
			PaidBySignatureMediaID:           request.PaidBySignatureMediaID,
			PaidByName:                       request.PaidByName,
			PaidByPosition:                   request.PaidByPosition,
			ModeOfPaymentFixedDays:           request.ModeOfPaymentFixedDays,
			TotalCredit:                      request.Applied1,
			TotalDebit:                       request.Applied1,
			ModeOfPaymentMonthlyExactDay:     request.ModeOfPaymentMonthlyExactDay,
			ComakerType:                      request.ComakerType,
		}

		if err := core.LoanTransactionManager(service).CreateWithTx(context, tx, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction: " + endTx(err).Error()})
		}
		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		if cashOnHandAccountID == nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Cash on hand account is not set for the branch: " + endTx(eris.New("cash on hand account not set")).Error()})
		}
		if err := core.LoanTransactionManager(service).UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + endTx(err).Error()})
		}
		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				clearanceAnalysis := &types.LoanClearanceAnalysis{
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

				if err := core.LoanClearanceAnalysisManager(service).CreateWithTx(context, tx, clearanceAnalysis); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(created) Failed to create loan clearance analysis: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				institution := &types.LoanClearanceAnalysisInstitution{
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

				if err := core.LoanClearanceAnalysisInstitutionManager(service).CreateWithTx(context, tx, institution); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				suggestedPayment := &types.LoanTermsAndConditionSuggestedPayment{

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

				if err := core.LoanTermsAndConditionSuggestedPaymentManager(service).CreateWithTx(context, tx, suggestedPayment); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition suggested payment: " + endTx(err).Error()})
				}
			}
		}
		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				amountReceipt := &types.LoanTermsAndConditionAmountReceipt{
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

				if err := core.LoanTermsAndConditionAmountReceiptManager(service).CreateWithTx(context, tx, amountReceipt); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms and condition amount receipt: " + endTx(err).Error()})
				}
			}
		}

		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				comakerMemberProfile := &types.ComakerMemberProfile{
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

				if err := core.ComakerMemberProfileManager(service).CreateWithTx(context, tx, comakerMemberProfile); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + endTx(err).Error()})
				}
			}
		}

		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				comakerCollateral := &types.ComakerCollateral{
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

				if err := core.ComakerCollateralManager(service).CreateWithTx(context, tx, comakerCollateral); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + endTx(err).Error()})
				}
			}
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		newTx, newEndTx := service.Database.StartTransaction(context)
		newLoanTransaction, err := event.LoanBalancing(context, service, newTx, newEndTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: *cashOnHandAccountID,
			LoanTransactionID:              loanTransaction.ID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v: %v", err, newEndTx(err))})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id",
		Method:       "PUT",
		ResponseType: types.LoanTransactionResponse{},
		RequestType: types.LoanTransactionRequest{},
		Note:         "Updates an existing loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transactions"})
		}

		request, err := core.LoanTransactionManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		account, err := core.AccountManager(service).GetByID(context, *request.AccountID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		tx, endTx := service.Database.StartTransaction(context)
		cashOnHandAccount, err := core.GetCashOnCashEquivalence(
			context, service, *loanTransactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash on cash equivalence account: " + endTx(err).Error()})
		}
		cashOnCashEquivalenceAccountID := cashOnHandAccount.ID
		if !helpers.UUIDPtrEqual(account.CurrencyID, &cashOnCashEquivalenceAccountID) {
			accounts, err := core.AccountManager(service).Find(context, &types.Account{
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

		loanTransaction.AccountID = request.AccountID
		loanTransaction.UpdatedByID = userOrg.UserID
		loanTransaction.OfficialReceiptNumber = request.OfficialReceiptNumber
		loanTransaction.Voucher = request.Voucher
		loanTransaction.EmployeeUserID = &userOrg.UserID
		loanTransaction.LoanPurposeID = request.LoanPurposeID
		loanTransaction.LoanStatusID = request.LoanStatusID
		loanTransaction.ModeOfPayment = request.ModeOfPayment
		loanTransaction.ModeOfPaymentWeekly = request.ModeOfPaymentWeekly
		loanTransaction.ModeOfPaymentSemiMonthlyPay1 = request.ModeOfPaymentSemiMonthlyPay1
		loanTransaction.ModeOfPaymentSemiMonthlyPay2 = request.ModeOfPaymentSemiMonthlyPay2
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
		loanTransaction.ComakerType = request.ComakerType
		loanTransaction.PreviousLoanID = request.PreviousLoanID

		if request.LoanClearanceAnalysisDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisDeleted {
				clearanceAnalysis, err := core.LoanClearanceAnalysisManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis for deletion: " + endTx(err).Error()})
				}
				if clearanceAnalysis.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				clearanceAnalysis.DeletedByID = &userOrg.UserID
				if err := core.LoanClearanceAnalysisManager(service).DeleteWithTx(context, tx, clearanceAnalysis.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanClearanceAnalysisInstitutionDeleted != nil {
			for _, deletedID := range request.LoanClearanceAnalysisInstitutionDeleted {
				institution, err := core.LoanClearanceAnalysisInstitutionManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution for deletion: " + endTx(err).Error()})
				}
				if institution.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan clearance analysis institution that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				institution.DeletedByID = &userOrg.UserID
				if err := core.LoanClearanceAnalysisInstitutionManager(service).DeleteWithTx(context, tx, institution.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanTermsAndConditionSuggestedPaymentDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionSuggestedPaymentDeleted {
				suggestedPayment, err := core.LoanTermsAndConditionSuggestedPaymentManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment for deletion: " + endTx(err).Error()})
				}
				if suggestedPayment.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms suggested payment that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				suggestedPayment.DeletedByID = &userOrg.UserID
				if err := core.LoanTermsAndConditionSuggestedPaymentManager(service).DeleteWithTx(context, tx, suggestedPayment.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanTermsAndConditionAmountReceiptDeleted != nil {
			for _, deletedID := range request.LoanTermsAndConditionAmountReceiptDeleted {
				amountReceipt, err := core.LoanTermsAndConditionAmountReceiptManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt for deletion: " + endTx(err).Error()})
				}
				if amountReceipt.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete loan terms amount receipt that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				amountReceipt.DeletedByID = &userOrg.UserID
				if err := core.LoanTermsAndConditionAmountReceiptManager(service).DeleteWithTx(context, tx, amountReceipt.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + endTx(err).Error()})
				}
			}
		}

		if request.ComakerMemberProfilesDeleted != nil {
			for _, deletedID := range request.ComakerMemberProfilesDeleted {
				comakerMemberProfile, err := core.ComakerMemberProfileManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile for deletion: " + endTx(err).Error()})
				}
				if comakerMemberProfile.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker member profile that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				comakerMemberProfile.DeletedByID = &userOrg.UserID
				if err := core.ComakerMemberProfileManager(service).DeleteWithTx(context, tx, comakerMemberProfile.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + endTx(err).Error()})
				}
			}
		}

		if request.ComakerCollateralsDeleted != nil {
			for _, deletedID := range request.ComakerCollateralsDeleted {
				comakerCollateral, err := core.ComakerCollateralManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker collateral for deletion: " + endTx(err).Error()})
				}
				if comakerCollateral.LoanTransactionID != loanTransaction.ID {
					return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete comaker collateral that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
				}
				comakerCollateral.DeletedByID = &userOrg.UserID
				if err := core.ComakerCollateralManager(service).DeleteWithTx(context, tx, comakerCollateral.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker collateral: " + endTx(err).Error()})
				}
			}
		}

		if request.LoanClearanceAnalysis != nil {
			for _, clearanceAnalysisReq := range request.LoanClearanceAnalysis {
				if clearanceAnalysisReq.ID != nil {
					existingRecord, err := core.LoanClearanceAnalysisManager(service).GetByID(context, *clearanceAnalysisReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.RegularDeductionDescription = clearanceAnalysisReq.RegularDeductionDescription
					existingRecord.RegularDeductionAmount = clearanceAnalysisReq.RegularDeductionAmount
					existingRecord.BalancesDescription = clearanceAnalysisReq.BalancesDescription
					existingRecord.BalancesAmount = clearanceAnalysisReq.BalancesAmount
					existingRecord.BalancesCount = clearanceAnalysisReq.BalancesCount

					if err := core.LoanClearanceAnalysisManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis: " + endTx(err).Error()})
					}
				} else {
					clearanceAnalysis := &types.LoanClearanceAnalysis{
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

					if err := core.LoanClearanceAnalysisManager(service).CreateWithTx(context, tx, clearanceAnalysis); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "(updated) Failed to create loan clearance analysis: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.LoanClearanceAnalysisInstitution != nil {
			for _, institutionReq := range request.LoanClearanceAnalysisInstitution {
				if institutionReq.ID != nil {
					existingRecord, err := core.LoanClearanceAnalysisInstitutionManager(service).GetByID(context, *institutionReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan clearance analysis institution: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan clearance analysis institution that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = institutionReq.Name
					existingRecord.Description = institutionReq.Description

					if err := core.LoanClearanceAnalysisInstitutionManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan clearance analysis institution: " + endTx(err).Error()})
					}
				} else {
					institution := &types.LoanClearanceAnalysisInstitution{
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

					if err := core.LoanClearanceAnalysisInstitutionManager(service).CreateWithTx(context, tx, institution); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan clearance analysis institution: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.LoanTermsAndConditionSuggestedPayment != nil {
			for _, suggestedPaymentReq := range request.LoanTermsAndConditionSuggestedPayment {
				if suggestedPaymentReq.ID != nil {
					existingRecord, err := core.LoanTermsAndConditionSuggestedPaymentManager(service).GetByID(context, *suggestedPaymentReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms suggested payment: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update loan terms suggested payment that doesn't belong to this loan transaction: " + endTx(eris.New("invalid loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.Name = suggestedPaymentReq.Name
					existingRecord.Description = suggestedPaymentReq.Description

					if err := core.LoanTermsAndConditionSuggestedPaymentManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms suggested payment: " + endTx(err).Error()})
					}
				} else {
					suggestedPayment := &types.LoanTermsAndConditionSuggestedPayment{
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

					if err := core.LoanTermsAndConditionSuggestedPaymentManager(service).CreateWithTx(context, tx, suggestedPayment); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms suggested payment: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.LoanTermsAndConditionAmountReceipt != nil {
			for _, amountReceiptReq := range request.LoanTermsAndConditionAmountReceipt {
				if amountReceiptReq.ID != nil {
					existingRecord, err := core.LoanTermsAndConditionAmountReceiptManager(service).GetByID(context, *amountReceiptReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing loan terms amount receipt: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot update loan terms amount receipt that doesn't belong to this loan transaction: " + endTx(eris.New("cannot update loan terms amount receipt that doesn't belong to this loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.AccountID = amountReceiptReq.AccountID
					existingRecord.Amount = amountReceiptReq.Amount

					if err := core.LoanTermsAndConditionAmountReceiptManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan terms amount receipt: " + endTx(err).Error()})
					}
				} else {
					amountReceipt := &types.LoanTermsAndConditionAmountReceipt{
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

					if err := core.LoanTermsAndConditionAmountReceiptManager(service).CreateWithTx(context, tx, amountReceipt); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan terms amount receipt: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.ComakerMemberProfiles != nil {
			for _, comakerReq := range request.ComakerMemberProfiles {
				if comakerReq.ID != nil {
					existingRecord, err := core.ComakerMemberProfileManager(service).GetByID(context, *comakerReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker member profile: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update comaker member profile that doesn't belong to this loan transaction: " + endTx(eris.New("Cannot update comaker member profile that doesn't belong to this loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.MemberProfileID = comakerReq.MemberProfileID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := core.ComakerMemberProfileManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker member profile: " + endTx(err).Error()})
					}
				} else {
					comakerMemberProfile := &types.ComakerMemberProfile{
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

					if err := core.ComakerMemberProfileManager(service).CreateWithTx(context, tx, comakerMemberProfile); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker member profile: " + endTx(err).Error()})
					}
				}
			}
		}

		if request.ComakerCollaterals != nil {
			for _, comakerReq := range request.ComakerCollaterals {
				if comakerReq.ID != nil {
					existingRecord, err := core.ComakerCollateralManager(service).GetByID(context, *comakerReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing comaker collateral: " + endTx(err).Error()})
					}
					if existingRecord.LoanTransactionID != loanTransaction.ID {
						return ctx.JSON(http.StatusForbidden,
							map[string]string{"error": "Cannot update comaker collateral that doesn't belong to this loan transaction: " + endTx(eris.New("Cannot update comaker collateral that doesn't belong to this loan transaction")).Error()})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = userOrg.UserID
					existingRecord.CollateralID = comakerReq.CollateralID
					existingRecord.Amount = comakerReq.Amount
					existingRecord.Description = comakerReq.Description
					existingRecord.MonthsCount = comakerReq.MonthsCount
					existingRecord.YearCount = comakerReq.YearCount

					if err := core.ComakerCollateralManager(service).UpdateByIDWithTx(context, tx, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comaker collateral: " + endTx(err).Error()})
					}
				} else {
					comakerCollateral := &types.ComakerCollateral{
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

					if err := core.ComakerCollateralManager(service).CreateWithTx(context, tx, comakerCollateral); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create comaker collateral: " + endTx(err).Error()})
					}
				}
			}
		}
		if !helpers.UUIDPtrEqual(account.CurrencyID, loanTransaction.Account.CurrencyID) {
			loanTransactionEntries, err := core.LoanTransactionEntryManager(service).Find(context, &types.LoanTransactionEntry{
				LoanTransactionID: loanTransaction.ID,
				OrganizationID:    userOrg.OrganizationID,
				BranchID:          *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + endTx(err).Error()})
			}
			for _, entry := range loanTransactionEntries {
				if err := core.LoanTransactionEntryManager(service).Delete(context, entry.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + endTx(err).Error()})
				}
			}
		} else {
			loanTransactionEntry, err := core.GetLoanEntryAccount(context, service, loanTransaction.ID, userOrg.OrganizationID, *userOrg.BranchID)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to find loan transaction entry (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change): " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry: " + endTx(err).Error()})
			}
			loanTransactionEntry.AccountID = &account.ID
			loanTransactionEntry.Name = account.Name
			loanTransactionEntry.Description = account.Description
			if err := core.LoanTransactionEntryManager(service).UpdateByIDWithTx(context, tx, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
					Module:      "LoanTransaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
			}

		}

		if err := core.LoanTransactionManager(service).UpdateByIDWithTx(context, tx, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "db-commit-error",
				Description: "Failed to commit transaction (/transaction/payment/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit database transaction: " + err.Error()})
		}

		tx, newEndTx := service.Database.StartTransaction(context)
		newLoanTransaction, err := event.LoanBalancing(context, service, tx, newEndTx, event.LoanBalanceEvent{
			CashOnCashEquivalenceAccountID: cashOnCashEquivalenceAccountID,
			LoanTransactionID:              loanTransaction.ID,
		}, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to retrieve updated loan transaction: %v: %v", err, newEndTx(err))})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-transaction/:loan_transaction_id",
		Method: "DELETE",
		Note:   "Deletes a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}

		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		tx, endTx := service.Database.StartTransaction(context)

		clearanceAnalysisList, err := core.LoanClearanceAnalysisManager(service).Find(context, &types.LoanClearanceAnalysis{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis records: " + endTx(err).Error()})
		}

		for _, clearanceAnalysis := range clearanceAnalysisList {
			clearanceAnalysis.DeletedByID = &userOrg.UserID
			if err := core.LoanClearanceAnalysisManager(service).DeleteWithTx(context, tx, clearanceAnalysis.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis: " + endTx(err).Error()})
			}
		}

		institutionList, err := core.LoanClearanceAnalysisInstitutionManager(service).Find(context, &types.LoanClearanceAnalysisInstitution{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan clearance analysis institution records: " + endTx(err).Error()})
		}

		for _, institution := range institutionList {
			institution.DeletedByID = &userOrg.UserID
			if err := core.LoanClearanceAnalysisInstitutionManager(service).DeleteWithTx(context, tx, institution.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan clearance analysis institution: " + endTx(err).Error()})
			}
		}

		suggestedPaymentList, err := core.LoanTermsAndConditionSuggestedPaymentManager(service).Find(context, &types.LoanTermsAndConditionSuggestedPayment{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms suggested payment records: " + endTx(err).Error()})
		}

		for _, suggestedPayment := range suggestedPaymentList {
			suggestedPayment.DeletedByID = &userOrg.UserID
			if err := core.LoanTermsAndConditionSuggestedPaymentManager(service).DeleteWithTx(context, tx, suggestedPayment.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms suggested payment: " + endTx(err).Error()})
			}
		}

		amountReceiptList, err := core.LoanTermsAndConditionAmountReceiptManager(service).Find(context, &types.LoanTermsAndConditionAmountReceipt{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan terms amount receipt records: " + endTx(err).Error()})
		}

		for _, amountReceipt := range amountReceiptList {
			amountReceipt.DeletedByID = &userOrg.UserID
			if err := core.LoanTermsAndConditionAmountReceiptManager(service).DeleteWithTx(context, tx, amountReceipt.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan terms amount receipt: " + endTx(err).Error()})
			}
		}

		transactionEntryList, err := core.LoanTransactionEntryManager(service).Find(context, &types.LoanTransactionEntry{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find loan transaction entry records: " + endTx(err).Error()})
		}

		for _, transactionEntry := range transactionEntryList {
			transactionEntry.DeletedByID = &userOrg.UserID
			if err := core.LoanTransactionEntryManager(service).DeleteWithTx(context, tx, transactionEntry.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + endTx(err).Error()})
			}
		}

		comakerMemberProfileList, err := core.ComakerMemberProfileManager(service).Find(context, &types.ComakerMemberProfile{
			LoanTransactionID: loanTransaction.ID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find comaker member profile records: " + endTx(err).Error()})
		}

		for _, comakerMemberProfile := range comakerMemberProfileList {
			comakerMemberProfile.DeletedByID = &userOrg.UserID
			if err := core.ComakerMemberProfileManager(service).DeleteWithTx(context, tx, comakerMemberProfile.ID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comaker member profile: " + endTx(err).Error()})
			}
		}

		loanTransaction.DeletedByID = &userOrg.UserID

		if err := core.LoanTransactionManager(service).DeleteWithTx(context, tx, loanTransaction.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Successfully deleted loan transaction %s and all related records", loanTransaction.ID),
			Module:      "LoanTransaction",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction and all related records deleted successfully"})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/loan-transaction/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple loan transactions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | no IDs provided",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | auth/organization error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | unauthorized user type",
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transactions"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.LoanTransactionManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/loan-transaction/bulk-delete) | error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete loan transactions: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted loan transactions (/loan-transaction/bulk-delete)",
			Module:      "LoanTransaction",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed by ID.",
		RequestType: types.LoanTransactionPrintRequest{},
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req types.LoanTransactionPrintRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction print request: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to print loan transactions"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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
		timeNow := userOrg.UserOrgTime()
		loanTransaction.PrintedDate = &timeNow
		loanTransaction.PrintedByID = &userOrg.UserID
		loanTransaction.Voucher = req.Voucher
		loanTransaction.CheckNumber = req.CheckNumber
		loanTransaction.CheckDate = req.CheckDate
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID

		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/loan-transaction/:loan_transaction_id/print-undo",
		Method: "PUT",
		Note:   "Reverts the print status of a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo print on loan transactions"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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
		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/print-only",
		Method:       "PUT",
		Note:         "Marks a loan transaction as printed without additional details by ID.",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to mark loan transactions as printed"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}
		loanTransaction.PrintNumber++
		loanTransaction.PrintedDate = helpers.Ptr(time.Now().UTC())
		loanTransaction.PrintedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = time.Now().UTC()
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/approve",
		Method:       "PUT",
		Note:         "Approves a loan transaction by ID.",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to approve loan transactions"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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
		timeNow := userOrg.UserOrgTime()
		loanTransaction.ApprovedDate = &timeNow
		loanTransaction.ApprovedByID = &userOrg.UserID
		loanTransaction.UpdatedAt = now
		loanTransaction.UpdatedByID = userOrg.UserID
		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Description:      fmt.Sprintf("Loan transaction has been approved by %s and is waiting to be released", *userOrg.User.FirstName),
			Title:            "Loan Transaction Approved - Pending Release",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/approve-undo",
		Method:       "PUT",
		Note:         "Reverts the approval status of a loan transaction by ID.",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to undo approval on loan transactions"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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
		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/release",
		Method:       "PUT",
		Note:         "Releases a loan transaction by ID. RELEASED SHOULD NOT BE UNAPPROVE",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to release loan transactions"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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

		newLoanTransaction, err := event.LoanRelease(context, service, loanTransaction.ID, userOrg)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}

		if err := event.TransactionBatchBalancing(context, service, newLoanTransaction.TransactionBatchID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to balance transaction batch: " + err.Error()})
		}

		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Description:      fmt.Sprintf("Loan transaction has been released by %s", *userOrg.User.FirstName),
			Title:            "Loan Transaction Released",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(newLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/signature",
		Method:       "PUT",
		Note:         "Updates the signature of a loan transaction by ID.",
		RequestType: types.LoanTransactionSignatureRequest{},
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req types.LoanTransactionSignatureRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction signature request: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transaction signatures"})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByID(context, *loanTransactionID)
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

		if err := core.LoanTransactionManager(service).UpdateByID(context, loanTransaction.ID, loanTransaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction: " + err.Error()})
		}
		newLoanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransaction.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change",
		Method:       "PUT",
		Note:         "Changes the cash and cash equivalence account for a loan transaction by ID.",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		account, err := core.AccountManager(service).GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		loanTransactionEntry, err := core.GetCashOnCashEquivalence(context, service, *loanTransactionID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction entry not found (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found: " + err.Error()})
		}
		loanTransactionEntry.AccountID = &account.ID
		loanTransactionEntry.Name = account.Name
		loanTransactionEntry.Description = account.Description
		if err := core.LoanTransactionEntryManager(service).UpdateByID(context, loanTransactionEntry.ID, loanTransactionEntry); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction entry update failed (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
		}
		loanTransaction, err := core.LoanTransactionManager(service).GetByIDRaw(context, loanTransactionEntry.LoanTransactionID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction not found after entry update (/loan-transaction/:loan_transaction_id/cash-and-cash-equivalence-account/:account_id/change), db error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found after entry update: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, loanTransaction)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/suggested",
		Method:       "POST",
		RequestType: types.LoanTransactionSuggestedRequest{},
		ResponseType: types.LoanTransactionSuggestedResponse{},
		Note:         "Updates the suggested payment details for a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req types.LoanTransactionSuggestedRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction suggested request: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		suggestedTerms, err := usecase.SuggestedNumberOfTerms(context, req.Amount, req.Principal, req.ModeOfPayment, req.FixedDays)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to calculate suggested terms: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, &types.LoanTransactionSuggestedResponse{
			Terms: suggestedTerms,
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/schedule",
		Method:       "GET",
		ResponseType: event.LoanTransactionAmortizationResponse{},
		Note:         "Retrieves the payment schedule for a loan transaction by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "authentication-failed",
				Description: "Failed to authenticate user organization for loan amortization schedule generation: " + err.Error(),
				Module:      "Loan Amortization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to authenticate user organization for loan amortization schedule"})
		}
		schedule, err := event.LoanAmortization(context, service, *loanTransactionID, userOrg)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "schedule-retrieval-failed",
				Description: "Failed to retrieve loan transaction schedule: " + err.Error(),
				Module:      "Loan Amortization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction schedule: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, schedule)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/adjustment",
		Method:       "POST",
		RequestType: types.LoanTransactionAdjustmentRequest{},
		ResponseType: types.LoanTransaction{},
		Note:         "Creates a loan transaction adjustment.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.LoanTransactionAdjustmentRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction adjustment request: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transaction adjustments"})
		}

		if err := event.LoanAdjustment(context, service, *userOrg, req); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction adjustment: " + err.Error()})
		}
		return ctx.NoContent(http.StatusCreated)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/process",
		Method:       "POST",
		Note:         "Processes a loan transaction by ID.",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		processedLoanTransaction, err := event.LoanProcessing(context, service, userOrg, loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process loan transaction: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.LoanTransactionManager(service).ToModel(processedLoanTransaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/:loan_transaction_id/guide",
		Method:       "GET",
		ResponseType: event.LoanGuideResponse{},
		Note:         "Returns comprehensive loan payment guide with schedules, statuses, and real-time balance tracking for a specific loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := helpers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan guides"})
		}

		loanGuide, err := event.LoanGuide(context, service, userOrg, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan guide: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, loanGuide)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/loan-transaction/process",
		Method:       "POST",
		Note:         "All Loan transactions that are pending to be processed will be processed",
		ResponseType: types.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "loan-processing-started",
			Description: "Loan processing started",
			Module:      "Loan Processing",
		})
		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Title:       "Loan Processing",
			Description: "Loan processing started",
		})

		if err := event.ProcessAllLoans(context, service, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process loan transactions: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
