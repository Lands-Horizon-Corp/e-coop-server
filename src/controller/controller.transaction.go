package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionController() {
	req := c.provider.Service.Request
	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/deposit/:transaction_id",
		Method:       "POST",
		RequestType:  model.PaymentRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Registers an online deposit against a specific transaction, updating both the general ledger and transaction record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Transaction error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		return c.event.Payment(context, ctx, event.PaymentEvent{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			ReferenceNumber:       req.ReferenceNumber,
			AccountID:             req.AccountID,
			MemberProfileID:       req.MemberProfileID,
			MemberJointAccountID:  req.MemberJointAccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
			TransactionType:       event.TransactionTypeDeposit,
		}, transactionID)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/withdraw/:transaction_id",
		Method:       "POST",
		RequestType:  model.PaymentRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Processes an online withdrawal for the given transaction, updating the general ledger and transaction amounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Transaction error: " + err.Error(),
				Module:      "User",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		return c.event.Payment(context, ctx, event.PaymentEvent{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			ReferenceNumber:       req.ReferenceNumber,
			AccountID:             req.AccountID,
			MemberProfileID:       req.MemberProfileID,
			MemberJointAccountID:  req.MemberJointAccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
			TransactionType:       event.TransactionTypeWithdraw,
		}, transactionID)
	})

	// Create transaction
	req.RegisterRoute(handlers.Route{
		Route:        "/transaction",
		Method:       "POST",
		RequestType:  model.TransactionRequest{},
		ResponseType: model.TransactionResponse{},
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
		var req model.TransactionRequest
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
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		transaction := &model.Transaction{
			CreatedAt:            time.Now().UTC(),
			CreatedByID:          userOrg.UserID,
			UpdatedAt:            time.Now().UTC(),
			UpdatedByID:          userOrg.UserID,
			BranchID:             *userOrg.BranchID,
			OrganizationID:       userOrg.OrganizationID,
			SignatureMediaID:     req.SignatureMediaID,
			TransactionBatchID:   &transactionBatch.ID,
			EmployeeUserID:       &userOrg.UserID,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			LoanBalance:          0,
			LoanDue:              0,
			TotalDue:             0,
			FinesDue:             0,
			TotalLoan:            0,
			InterestDue:          0,
			Amount:               0,
			ReferenceNumber:      req.ReferenceNumber,
			Source:               req.Source,
			Description:          req.Description,
		}
		if err := c.model.TransactionManager.Create(context, transaction); err != nil {
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
		return ctx.JSON(http.StatusCreated, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/deposit",
		Method:       "POST",
		RequestType:  model.PaymentQuickRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Performs a quick deposit operation using minimal information, creating both a transaction and related ledger entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/deposit): " + err.Error(),
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
		return c.event.Withdraw(context, ctx, event.PaymentEvent{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			ReferenceNumber:       req.ReferenceNumber,
			AccountID:             req.AccountID,
			MemberProfileID:       req.MemberProfileID,
			MemberJointAccountID:  req.MemberJointAccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
			TransactionType:       event.TransactionTypeWithdraw,
		})
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/withdraw",
		Method:       "POST",
		RequestType:  model.PaymentQuickRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Executes a quick withdrawal with minimal required info, generating both a transaction and related ledger entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/withdraw): " + err.Error(),
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
		return c.event.Deposit(context, ctx, event.PaymentEvent{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			ReferenceNumber:       req.ReferenceNumber,
			AccountID:             req.AccountID,
			MemberProfileID:       req.MemberProfileID,
			MemberJointAccountID:  req.MemberJointAccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
			TransactionType:       event.TransactionTypeDeposit,
		})
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/:transaction_id",
		Method:       "PUT",
		RequestType:  model.TransactionRequestEdit{},
		ResponseType: model.TransactionResponse{},
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
		var req model.TransactionRequestEdit
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

		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
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
		transaction.UpdatedAt = time.Now().UTC()
		transaction.UpdatedByID = userOrg.UserID
		if err := c.model.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
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
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Retrieves detailed information for the specified transaction by its unique identifier.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/current/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user (automatically adjusted for employee, admin, and member) within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Failed to get user organization: " + err.Error(),
			})
		}
		var filter model.Transaction
		if userOrg.UserType == "member" {
			memberProfile, err := c.model.MemberProfileManager.FindOne(context, &model.MemberProfile{
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

		transactions, err := c.model.TransactionManager.Find(context, &filter)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve transactions: " + err.Error(),
			})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/employee/:employee_id/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Fetches all transactions handled by the specified employee, filtered by organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		employeeID, err := handlers.EngineUUIDParam(ctx, "employee_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid employee ID: " + err.Error()})
		}
		employee, err := c.model.UserManager.GetByID(context, *employeeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Employee not found: " + err.Error()})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			EmployeeUserID: &employee.ID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Retrieves all transactions related to the given member profile within the user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID: " + err.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			MemberProfileID: &memberProfile.ID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/branch/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Provides a paginated list of all transactions recorded for the current branch of the user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve branch transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Pagination(context, ctx, transactions))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
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
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Pagination(context, ctx, transactions))
	})
}
