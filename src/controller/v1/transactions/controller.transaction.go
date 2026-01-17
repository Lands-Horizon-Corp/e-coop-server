package transactions

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func TransactionController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction",
		Method:       "POST",
		RequestType: types.TransactionRequest{},
		ResponseType: types.TransactionResponse{},
		Note:         "Creates a new transaction record with provided details, allowing subsequent deposit or withdrawal actions.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Failed to get user organization",
			})
		}

		var req types.TransactionRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid request body",
			})
		}

		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "validation-error",
				Description: "Validation failed (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"error": "Validation failed",
			})
		}

		transactionBatch, err := core.TransactionBatchCurrent(
			context,
			service,
			userOrg.UserID,
			userOrg.OrganizationID,
			*userOrg.BranchID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, echo.Map{
				"error": "Failed to retrieve transaction batch",
			})
		}
		tx, endTx := service.Database.StartTransaction(context)
		transaction := &types.Transaction{
			CreatedAt:   time.Now().UTC(),
			CreatedByID: userOrg.UserID,
			UpdatedAt:   time.Now().UTC(),
			UpdatedByID: userOrg.UserID,
			BranchID:    *userOrg.BranchID,

			OrganizationID:   userOrg.OrganizationID,
			SignatureMediaID: req.SignatureMediaID,

			TransactionBatchID: &transactionBatch.ID,
			EmployeeUserID:     &userOrg.UserID,

			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			CurrencyID:           req.CurrencyID,

			LoanBalance: 0,
			LoanDue:     0,
			TotalDue:    0,
			FinesDue:    0,
			TotalLoan:   0,
			InterestDue: 0,
			Amount:      0,

			ReferenceNumber: req.ReferenceNumber,
		}
		if req.IsReferenceNumberChecked {
			if err := event.IncrementOfficialReceipt(context, service, tx, transaction.ReferenceNumber, core.GeneralLedgerSourcePayment, userOrg); err != nil {
				return ctx.JSON(http.StatusConflict, echo.Map{"error": endTx(err).Error()})
			}
		}
		if err := core.TransactionManager(service).CreateWithTx(context, tx, transaction); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Transaction creation failed (/transaction): " + endTx(err).Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create transaction"})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Transaction commit failed (/transaction): " + endTx(err).Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to commit transaction"})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Transaction created successfully (/transaction), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})

		return ctx.JSON(http.StatusCreated, core.TransactionManager(service).ToModel(transaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id",
		Method:       "PUT",
		RequestType: types.TransactionRequestEdit{},
		ResponseType: types.TransactionResponse{},
		Note:         "Modifies the description of an existing transaction, allowing updates to its memo or comment field.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "param-error",
				Description: "Invalid transaction ID (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req types.TransactionRequestEdit
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change request failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		transaction, err := core.TransactionManager(service).GetByID(context, *transactionID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "not-found-error",
				Description: "Transaction not found or lock failed (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		transaction.Description = req.Description
		transaction.ReferenceNumber = req.ReferenceNumber
		transaction.UpdatedAt = time.Now().UTC()
		transaction.UpdatedByID = userOrg.UserID
		if err := core.TransactionManager(service).UpdateByID(context, transaction.ID, transaction); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update transaction (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Transaction description updated successfully (/transaction/:transaction_id), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, core.TransactionManager(service).ToModel(transaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Retrieves detailed information for the specified transaction by its unique identifier.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		transaction, err := core.TransactionManager(service).GetByID(context, *transactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TransactionManager(service).ToModel(transaction))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/current/search",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user (automatically adjusted for employee, admin, and member) within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Failed to get user organization: " + err.Error(),
			})
		}
		var filter core.Transaction
		if userOrg.UserType == core.UserOrganizationTypeMember {
			memberProfile, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID: &userOrg.UserID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to retrieve transactions: " + err.Error(),
				})
			}
			filter.MemberProfileID = &memberProfile.ID
		} else {
			filter.EmployeeUserID = &userOrg.UserID
		}

		filter.OrganizationID = userOrg.OrganizationID
		filter.BranchID = *userOrg.BranchID
		transactionPagination, err := core.TransactionManager(service).NormalPagination(context, ctx, &filter)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to paginate transactions: " + err.Error(),
			})
		}
		return ctx.JSON(http.StatusOK, transactionPagination)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/current",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user (automatically adjusted for employee, admin, and member) within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Failed to get user organization: " + err.Error(),
			})
		}
		var filter core.Transaction
		if userOrg.UserType == core.UserOrganizationTypeMember {
			memberProfile, err := core.MemberProfileManager(service).FindOne(context, &types.MemberProfile{
				UserID: &userOrg.UserID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to retrieve member profile: " + err.Error(),
				})
			}
			filter.MemberProfileID = &memberProfile.ID
		} else {
			filter.EmployeeUserID = &userOrg.UserID
		}
		filter.OrganizationID = userOrg.OrganizationID
		filter.BranchID = *userOrg.BranchID

		transactions, err := core.TransactionManager(service).Find(context, &filter)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve transactions: " + err.Error(),
			})
		}
		return ctx.JSON(http.StatusOK, core.TransactionManager(service).ToModels(transactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Fetches all transactions handled by the specified employee, filtered by organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		userOrganizationID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Employee not found: " + err.Error()})
		}
		transactions, err := core.TransactionManager(service).NormalPagination(context, ctx, &types.Transaction{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, transactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Retrieves all transactions related to the given member profile within the user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		transactions, err := core.TransactionManager(service).Find(context, &types.Transaction{
			MemberProfileID: &memberProfile.ID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TransactionManager(service).ToModels(transactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/branch/search",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Provides a paginated list of all transactions recorded for the current branch of the user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactions, err := core.TransactionManager(service).NormalPagination(context, ctx, &types.Transaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve branch transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, transactions)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: types.TransactionResponse{},
		Note:         "Retrieves all transactions associated with a specific transaction batch, allowing for batch-level analysis.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := helpers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactions, err := core.TransactionManager(service).NormalPagination(context, ctx, &types.Transaction{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, transactions)
	})

}
