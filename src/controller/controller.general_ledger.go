package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// GeneralLedgerController manages endpoints for general ledger accounts, definitions, and member ledgers.
func (c *Controller) GeneralLedgerController() {
	req := c.provider.Service.Request

	// GET /general-ledger-accounts-grouping (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-accounts-grouping",
		Method:   "GET",
		Response: "GeneralLedgerAccountsGrouping[]",
		Note:     "Returns all general ledger account groupings for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view general ledger account groupings"})
		}
		gl, err := c.model.GeneralLedgerAccountsGroupingManager.Find(context, &model.GeneralLedgerAccountsGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger account groupings: " + err.Error()})
		}
		for _, grouping := range gl {
			if grouping != nil {
				grouping.GeneralLedgerDefinitionEntries = []*model.GeneralLedgerDefinition{}
				entries, err := c.model.GeneralLedgerDefinitionManager.FindWithConditions(context, map[string]any{
					"organization_id":                     userOrg.OrganizationID,
					"branch_id":                           *userOrg.BranchID,
					"general_ledger_accounts_grouping_id": &grouping.ID,
				})
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger definitions: " + err.Error()})
				}
				var filteredEntries []*model.GeneralLedgerDefinition
				for _, entry := range entries {
					if entry.GeneralLedgerDefinitionEntryID == nil {
						filteredEntries = append(filteredEntries, entry)
					}
				}
				grouping.GeneralLedgerDefinitionEntries = filteredEntries
			}
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerAccountsGroupingManager.Filtered(context, ctx, gl))
	})

	// PUT /general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id",
		Method:   "PUT",
		Request:  "GeneralLedgerAccountsGroupingRequest",
		Response: "GeneralLedgerAccountsGroupingResponse",
		Note:     "Updates an existing general ledger account grouping by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		groupingID, err := horizon.EngineUUIDParam(ctx, "general_ledger_accounts_grouping_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), invalid grouping ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger account grouping ID"})
		}
		reqBody, err := c.model.GeneralLedgerAccountsGroupingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid grouping data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for general ledger account grouping (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger account groupings"})
		}
		grouping, err := c.model.GeneralLedgerAccountsGroupingManager.GetByID(context, *groupingID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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

		if debit, err := strconv.ParseFloat(reqBody.Debit, 64); err == nil {
			grouping.Debit = debit
		}
		if credit, err := strconv.ParseFloat(reqBody.Credit, 64); err == nil {
			grouping.Credit = credit
		}

		if err := c.model.GeneralLedgerAccountsGroupingManager.UpdateFields(context, grouping.ID, grouping); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger account grouping update failed (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger account grouping: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated general ledger account grouping (/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id): " + grouping.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerAccountsGroupingManager.ToModel(grouping))
	})

	// GET /general-ledger-definition (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition",
		Method:   "GET",
		Response: "GeneralLedgerDefinition[]",
		Note:     "Returns all general ledger definitions for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view general ledger definitions"})
		}
		gl, err := c.model.GeneralLedgerDefinitionManager.FindRaw(context, &model.GeneralLedgerDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger definitions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, gl)
	})

	// POST /general-ledger-definition (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition",
		Method:   "POST",
		Request:  "GeneralLedgerDefinitionRequest",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Creates a new general ledger definition for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.GeneralLedgerDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for general ledger definition (/general-ledger-definition)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create general ledger definitions"})
		}
		glDefinition := &model.GeneralLedgerDefinition{
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
		if err := c.model.GeneralLedgerDefinitionManager.Create(context, glDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "General ledger definition creation failed (/general-ledger-definition), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create general ledger definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created general ledger definition (/general-ledger-definition): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusCreated, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-definition/:general_ledger_definition_id (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id",
		Method:   "PUT",
		Request:  "GeneralLedgerDefinitionRequest",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Updates an existing general ledger definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		req, err := c.model.GeneralLedgerDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), validation error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for general ledger definition (/general-ledger-definition/:general_ledger_definition_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger definitions"})
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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

		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "General ledger definition update failed (/general-ledger-definition/:general_ledger_definition_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated general ledger definition (/general-ledger-definition/:general_ledger_definition_id): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// POST /general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect",
		Method:   "POST",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Connects an account to a general ledger definition by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), invalid GL definition ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), invalid account ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized connect attempt for account to general ledger definition (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to connect accounts"})
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		if glDefinition.OrganizationID != userOrg.OrganizationID || glDefinition.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), wrong org/branch.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "General ledger definition does not belong to your organization/branch"})
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account wrong org/branch.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to your organization/branch"})
		}
		account.GeneralLedgerDefinitionID = glDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID
		if err := c.model.AccountManager.UpdateFields(context, account.ID, account); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), account db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID
		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Account connect to general ledger definition failed (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect), GL def db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Connected account to GL definition (/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect): " + account.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-definition/:general_ledger_definition_id/index/:index (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id/index/:index",
		Method:   "PUT",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Updates the index of a general ledger definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), invalid index.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index value"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized index update attempt for GL definition (/general-ledger-definition/:general_ledger_definition_id/index/:index)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update general ledger definition index"})
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		glDefinition.Index = index
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID
		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL definition index update failed (/general-ledger-definition/:general_ledger_definition_id/index/:index), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update index: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated GL definition index (/general-ledger-definition/:general_ledger_definition_id/index/:index): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// PUT /general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index",
		Method:   "PUT",
		Request:  "UpdateAccountIndexRequest {general_ledger_definition_index: int, account_index: int}",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Updates the index of an account within a general ledger definition and reorders accordingly.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), invalid GL definition ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), invalid payload.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized GL grouping/account index update attempt (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update account index"})
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), GL definition not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), account not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.GeneralLedgerDefinitionID == nil || *account.GeneralLedgerDefinitionID != *glDefinitionID {
			account.GeneralLedgerDefinitionID = glDefinitionID
		}
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			GeneralLedgerDefinitionID: glDefinitionID,
			OrganizationID:            userOrg.OrganizationID,
			BranchID:                  *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), account find error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		var updatedAccounts []*model.Account
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
		updatedAccounts = append(updatedAccounts[:reqBody.AccountIndex], append([]*model.Account{account}, updatedAccounts[reqBody.AccountIndex:]...)...)
		for idx, acc := range updatedAccounts {
			acc.Index = idx
			acc.UpdatedAt = time.Now().UTC()
			acc.UpdatedByID = userOrg.UserID
			if err := c.model.AccountManager.UpdateFields(context, acc.ID, acc); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
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
			if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "GL grouping/account index update failed (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index), update GL definition error: " + err.Error(),
					Module:      "GeneralLedger",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update general ledger definition index: " + err.Error()})
			}
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account index within GL definition (/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index): " + account.Name,
			Module:      "GeneralLedger",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	// The rest GET endpoints (member-accounting-ledger/member-profile/:member_profile_id/total, search, total, account/:account_id/search) and DELETE (general-ledger-definition/:general_definition_id)
	// (NO footstep for GET endpoints, WITH footstep for DELETE)
	req.RegisterRoute(horizon.Route{
		Route:    "/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Returns the total amount for a specific member profile's general ledger entries.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		var totalAmount float64
		for _, entry := range entries {
			totalAmount += entry.Balance
		}
		return ctx.JSON(http.StatusOK, map[string]any{
			"total_amount":                     totalAmount,
			"total_share_capital_plus_savings": 0,
			"total_loans":                      totalAmount,
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-accounting-ledger/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Returns paginated general ledger entries for a specific member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is missing for user organization"})
		}
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member accounting ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Returns paginated general ledger entries for a specific member profile and account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Returns the total amount for a specific member profile's general ledger entries for an account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view member general ledger totals"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		var totalAmount float64
		var debit float64
		var credit float64
		for _, entry := range entries {
			totalAmount += entry.Debit - entry.Credit
			debit += entry.Debit
			credit += entry.Credit
		}
		return ctx.JSON(http.StatusOK, map[string]any{
			"balance":      totalAmount,
			"total_debit":  debit,
			"total_credit": credit,
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger/account/:account_id/search",
		Method:   "GET",
		Response: "GeneralLedger[]",
		Note:     "Returns paginated general ledger entries for a specific account.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view general ledger entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve general ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// DELETE /general-ledger-definition/:general_definition_id (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_definition_id",
		Method:   "DELETE",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Deletes a general ledger definition by its ID, only if no accounts are linked.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), invalid ID.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger definition ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), user org error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for GL definition (/general-ledger-definition/:general_definition_id)",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete general ledger definitions"})
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), not found.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger definition not found"})
		}
		if len(glDefinition.GeneralLedgerDefinitionEntries) > 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), has sub-entries.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: general ledger definition has sub-entries"})
		}
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			GeneralLedgerDefinitionID: glDefinitionID,
			OrganizationID:            userOrg.OrganizationID,
			BranchID:                  *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), account find error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check linked accounts: " + err.Error()})
		}
		if len(accounts) > 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), has linked accounts.",
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: accounts are linked to this general ledger definition"})
		}
		if err := c.model.GeneralLedgerDefinitionManager.DeleteByID(context, glDefinition.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "GL definition delete failed (/general-ledger-definition/:general_definition_id), db error: " + err.Error(),
				Module:      "GeneralLedger",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete general ledger definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted GL definition (/general-ledger-definition/:general_definition_id): " + glDefinition.Name,
			Module:      "GeneralLedger",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
