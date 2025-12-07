package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// GeneralLedgerController manages endpoints for general ledger accounts, definitions, and member ledgers.
func (c *Controller) generalLedgerGroupingController() {
	req := c.provider.Service.WebRequest

	// GET /general-ledger-accounts-grouping
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-accounts-grouping",
		Method:       "GET",
		ResponseType: core.GeneralLedgerAccountsGroupingResponse{},
		Note:         "Returns all general ledger account groupings for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view general ledger account groupings"})
		}
		gl, err := c.core.GeneralLedgerAlignments(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger account groupings: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerAccountsGroupingManager.ToModels(gl))
	})

	// PUT /general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id",
		Method:       "PUT",
		ResponseType: core.GeneralLedgerAccountsGroupingResponse{},
		RequestType:  core.GeneralLedgerAccountsGroupingRequest{},
		Note:         "Updates an existing general ledger account grouping by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		groupingID, err := handlers.EngineUUIDParam(ctx, "general_ledger_accounts_grouping_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), invalid grouping ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger account grouping ID"})
		}
		reqBody, err := c.core.GeneralLedgerAccountsGroupingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid grouping data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for general ledger account grouping (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger account groupings"})
		}
		grouping, err := c.core.GeneralLedgerAccountsGroupingManager.GetByID(context, *groupingID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger account grouping not found"})
		}
		grouping.Name = reqBody.Name
		grouping.Description = reqBody.Description
		grouping.UpdatedAt = time.Now().UTC()
		grouping.UpdatedByID = userOrg.UserID
		grouping.FromCode = reqBody.FromCode
		grouping.ToCode = reqBody.ToCode
		grouping.Debit = reqBody.Debit
		grouping.Credit = reqBody.Credit

		if err := c.core.GeneralLedgerAccountsGroupingManager.UpdateByID(context, grouping.ID, grouping); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger account grouping: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated general ledger account grouping (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id): " + grouping.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerAccountsGroupingManager.ToModel(grouping))
	})

	// GET /general-ledger-definition
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-definition",
		Method:       "GET",
		ResponseType: core.GeneralLedgerDefinitionResponse{},
		Note:         "Returns all general ledger definitions for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view general ledger definitions"})
		}
		gl, err := c.core.GeneralLedgerDefinitionManager.FindRaw(context, &core.GeneralLedgerDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger definitions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, gl)
	})

	// POST /general-ledger-definition (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-definition",
		Method:       "POST",
		RequestType:  core.GeneralLedgerDefinitionRequest{},
		ResponseType: core.GeneralLedgerDefinitionResponse{},
		Note:         "Creates a new general ledger definition for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.GeneralLedgerDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for general ledger definition (/general-ledger-definition)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create general ledger definitions"})
		}
		glDefinition := &core.GeneralLedgerDefinition{
			OrganizationID:                  userOrg.OrganizationID,
			BranchID:                        *userOrg.BranchID,
			CreatedByID:                     userOrg.UserID,
			UpdatedByID:                     userOrg.UserID,
			GeneralLedgerDefinitionEntryID:  req.GeneralLedgerDefinitionEntryID,
			GeneralLedgerAccountsGroupingID: req.GeneralLedgerAccountsGroupingID,
			Name:                            req.Name,
			Description:                     req.Description,
			Index:                           req.Index,
			NameInTotal:                     req.NameInTotal,
			IsPosting:                       req.IsPosting,
			GeneralLedgerType:               req.GeneralLedgerType,
			BeginningBalanceOfTheYearCredit: req.BeginningBalanceOfTheYearCredit,
			BeginningBalanceOfTheYearDebit:  req.BeginningBalanceOfTheYearDebit,
			CreatedAt:                       time.Now().UTC(),
			UpdatedAt:                       time.Now().UTC(),
		}
		if err := c.core.GeneralLedgerDefinitionManager.Create(context, glDefinition); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create general ledger definition: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created general ledger definition (/general-ledger-definition): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusCreated, c.core.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-definition/:general_ledger_definition_id (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-definition/:general_ledger_definition_id",
		Method:       "PUT",
		RequestType:  core.GeneralLedgerDefinitionRequest{},
		ResponseType: core.GeneralLedgerDefinitionResponse{},
		Note:         "Updates an existing general ledger definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := handlers.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		req, err := c.core.GeneralLedgerDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for general ledger definition (/general-ledger-definition/:general_ledger_definition_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger definitions"})
		}
		glDefinition, err := c.core.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		glDefinition.GeneralLedgerDefinitionEntryID = req.GeneralLedgerDefinitionEntryID
		glDefinition.Name = req.Name
		glDefinition.Description = req.Description
		glDefinition.Index = req.Index
		glDefinition.NameInTotal = req.NameInTotal
		glDefinition.IsPosting = req.IsPosting
		glDefinition.GeneralLedgerType = req.GeneralLedgerType
		glDefinition.BeginningBalanceOfTheYearCredit = req.BeginningBalanceOfTheYearCredit
		glDefinition.BeginningBalanceOfTheYearDebit = req.BeginningBalanceOfTheYearDebit
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID

		if err := c.core.GeneralLedgerDefinitionManager.UpdateByID(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated general ledger definition (/general-ledger-definition/:general_ledger_definition_id): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// POST /general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect",
		Method:       "POST",
		Note:         "Connects an account to a general ledger definition by their IDs.",
		ResponseType: core.GeneralLedgerDefinitionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := handlers.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), invalid GL definition ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), invalid account ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized connect attempt for account to general ledger definition (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to connect accounts"})
		}
		glDefinition, err := c.core.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		if glDefinition.OrganizationID != userOrg.OrganizationID || glDefinition.BranchID != *userOrg.BranchID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), wrong org/branch.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "General ledger definition does not belong to your organization/branch"})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account wrong org/branch.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to your organization/branch"})
		}
		account.GeneralLedgerDefinitionID = glDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID
		if err := c.core.AccountManager.UpdateByID(context, account.ID, account); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID
		if err := c.core.GeneralLedgerDefinitionManager.UpdateByID(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), GL def db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Connected account to GL definition (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect): " + account.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-definition/:general_ledger_definition_id/index/:index (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-definition/:general_ledger_definition_id/index/:index",
		Method:       "PUT",
		ResponseType: core.GeneralLedgerDefinitionResponse{},
		Note:         "Updates the index of a general ledger definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := handlers.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), invalid index.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index value"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized index update attempt for GL definition (/general-ledger-definition/:general_ledger_definition_id/index/:index)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger definition index"})
		}
		glDefinition, err := c.core.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		glDefinition.Index = index
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID
		if err := c.core.GeneralLedgerDefinitionManager.UpdateByID(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update index: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated GL definition index (/general-ledger-definition/:general_ledger_definition_id/index/:index): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index",
		Method:       "PUT",
		Note:         "Updates the index of an account within a general ledger definition and reorders accordingly.",
		ResponseType: core.GeneralLedgerDefinitionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := handlers.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), invalid GL definition ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), invalid account ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		type UpdateAccountIndexRequest struct {
			GeneralLedgerDefinitionIndex int `json:"general_ledger_definition_index"`
			AccountIndex                 int `json:"account_index"`
		}
		var reqBody UpdateAccountIndexRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), invalid payload.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized GL grouping/account index update attempt (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update account index"})
		}
		glDefinition, err := c.core.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), GL definition not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		account, err := c.core.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), account not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.GeneralLedgerDefinitionID == nil || *account.GeneralLedgerDefinitionID != *glDefinitionID {
			account.GeneralLedgerDefinitionID = glDefinitionID
		}
		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			GeneralLedgerDefinitionID: glDefinitionID,
			OrganizationID:            userOrg.OrganizationID,
			BranchID:                  *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), account find error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		var updatedAccounts []*core.Account
		for _, acc := range accounts {
			if acc.ID != account.ID {
				updatedAccounts = append(updatedAccounts, acc)
			}
		}
		if reqBody.AccountIndex < 0 {
			reqBody.AccountIndex = 0
		}
		if reqBody.AccountIndex > len(updatedAccounts) {
			reqBody.AccountIndex = len(updatedAccounts)
		}
		updatedAccounts = append(updatedAccounts[:reqBody.AccountIndex], append([]*core.Account{account}, updatedAccounts[reqBody.AccountIndex:]...)...)
		for idx, acc := range updatedAccounts {
			acc.Index = idx
			acc.UpdatedAt = time.Now().UTC()
			acc.UpdatedByID = userOrg.UserID
			if err := c.core.AccountManager.UpdateByID(context, acc.ID, acc); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), update account error: " + err.Error(),
					Module:      "GeneralLedger",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account: " + err.Error()})
			}
		}
		if glDefinition.Index != reqBody.GeneralLedgerDefinitionIndex {
			glDefinition.Index = reqBody.GeneralLedgerDefinitionIndex
			glDefinition.UpdatedAt = time.Now().UTC()
			glDefinition.UpdatedByID = userOrg.UserID
			if err := c.core.GeneralLedgerDefinitionManager.UpdateByID(context, glDefinition.ID, glDefinition); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), update GL definition error: " + err.Error(),
					Module:      "GeneralLedger",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition index: " + err.Error()})
			}
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account index within GL definition (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index): " + account.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// DELETE /general-ledger-definition/:general_definition_id (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/general-ledger-definition/:general_definition_id",
		Method: "DELETE",
		Note:   "Deletes a general ledger definition by its ID, only if no accounts are linked.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := handlers.EngineUUIDParam(ctx, "general_definition_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for GL definition (/general-ledger-definition/:general_definition_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete general ledger definitions"})
		}
		glDefinition, err := c.core.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		if len(glDefinition.GeneralLedgerDefinitionEntries) > 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), has sub-entries.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: general ledger definition has sub-entries"})
		}
		accounts, err := c.core.AccountManager.Find(context, &core.Account{
			GeneralLedgerDefinitionID: glDefinitionID,
			OrganizationID:            userOrg.OrganizationID,
			BranchID:                  *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), account find error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check linked accounts: " + err.Error()})
		}
		if len(accounts) > 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), has linked accounts.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: accounts are linked to this general ledger definition"})
		}
		if err := c.core.GeneralLedgerDefinitionManager.Delete(context, glDefinition.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete general ledger definition: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted GL definition (/general-ledger-definition/:general_definition_id): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
