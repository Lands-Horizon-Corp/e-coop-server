package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) TransactionController() {
	req := c.provider.Service.Request

	// Create transaction
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction",
		Method:       "POST",
		RequestType:  modelCore.TransactionRequest{},
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Creates a new transaction record with provided details, allowing subsequent deposit or withdrawal actions.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		var req modelCore.TransactionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change request failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		account, err := c.modelCore.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "account-error",
				Description: "Failed to retrieve member joint account (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve member joint account: " + err.Error()})
		}
		transactionBatch, err := c.modelCore.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		transaction := &modelCore.Transaction{
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
			CurrencyID:           *account.CurrencyID,

			LoanBalance:     0,
			LoanDue:         0,
			TotalDue:        0,
			FinesDue:        0,
			TotalLoan:       0,
			InterestDue:     0,
			Amount:          0,
			ReferenceNumber: req.ReferenceNumber,
			Description:     req.Description,
		}
		if req.IsReferenceNumberChecked {
			userOrg.UserSettingUsedOR = userOrg.UserSettingUsedOR + 1
			if err := c.modelCore.UserOrganizationManager.Update(context, userOrg); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Failed to update user organization (/transaction): " + err.Error(),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + err.Error()})
			}
		}
		if err := c.modelCore.TransactionManager.Create(context, transaction); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Transaction creation failed (/transaction), db error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Transaction created successfully (/transaction), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusCreated, c.modelCore.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id",
		Method:       "PUT",
		RequestType:  modelCore.TransactionRequestEdit{},
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Modifies the description of an existing transaction, allowing updates to its memo or comment field.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "param-error",
				Description: "Invalid transaction ID (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req modelCore.TransactionRequestEdit
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Change request failed: validation error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// Begin transaction for row-level locking
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Failed to start transaction (/transaction/:transaction_id): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		transaction, err := c.modelCore.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		if err := c.modelCore.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update transaction (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Transaction description updated successfully (/transaction/:transaction_id), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Retrieves detailed information for the specified transaction by its unique identifier.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		transaction, err := c.modelCore.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/current/search",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user (automatically adjusted for employee, admin, and member) within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Failed to get user organization: " + err.Error(),
			})
		}
		var filter modelCore.Transaction
		if userOrg.UserType == modelCore.UserOrganizationTypeMember {
			memberProfile, err := c.modelCore.MemberProfileManager.FindOne(context, &modelCore.MemberProfile{
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

		transactions, err := c.modelCore.TransactionManager.Find(context, &filter)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve transactions: " + err.Error(),
			})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Pagination(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/current",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user (automatically adjusted for employee, admin, and member) within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Failed to get user organization: " + err.Error(),
			})
		}
		var filter modelCore.Transaction
		if userOrg.UserType == modelCore.UserOrganizationTypeMember {
			memberProfile, err := c.modelCore.MemberProfileManager.FindOne(context, &modelCore.MemberProfile{
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

		transactions, err := c.modelCore.TransactionManager.Find(context, &filter)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve transactions: " + err.Error(),
			})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Fetches all transactions handled by the specified employee, filtered by organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		userOrganizationID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Employee not found: " + err.Error()})
		}
		transactions, err := c.modelCore.TransactionManager.Find(context, &modelCore.Transaction{
			EmployeeUserID: &userOrganization.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Pagination(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Retrieves all transactions related to the given member profile within the user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID: " + err.Error()})
		}
		memberProfile, err := c.modelCore.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		transactions, err := c.modelCore.TransactionManager.Find(context, &modelCore.Transaction{
			MemberProfileID: &memberProfile.ID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/branch/search",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Provides a paginated list of all transactions recorded for the current branch of the user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != modelCore.UserOrganizationTypeOwner && userOrg.UserType != modelCore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactions, err := c.modelCore.TransactionManager.Find(context, &modelCore.Transaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve branch transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Pagination(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: modelCore.TransactionResponse{},
		Note:         "Retrieves all transactions associated with a specific transaction batch, allowing for batch-level analysis.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactions, err := c.modelCore.TransactionManager.Find(context, &modelCore.Transaction{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.TransactionManager.Pagination(context, ctx, transactions))
	})

}
