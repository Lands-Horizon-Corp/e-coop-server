package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// FinancialStatementController manages endpoints for financial statement groupings and definitions.
func (c *Controller) financialStatementController() {
	req := c.provider.Service.Request

	// GET /financial-statement-grouping: List all financial statement groupings for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-grouping",
		Method:       "GET",
		ResponseType: modelcore.FinancialStatementGroupingResponse{},
		Note:         "Returns all financial statement groupings for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement groupings"})
		}
		fsGroupings, err := c.modelcore.FinancialStatementGroupingManager.Find(context, &modelcore.FinancialStatementGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve financial statement groupings: " + err.Error()})
		}
		for _, grouping := range fsGroupings {
			if grouping != nil {
				grouping.FinancialStatementDefinitionEntries = []*modelcore.FinancialStatementDefinition{}
				entries, err := c.modelcore.FinancialStatementDefinitionManager.FindWithConditions(context, map[string]any{
					"organization_id":                 userOrg.OrganizationID,
					"branch_id":                       *userOrg.BranchID,
					"financial_statement_grouping_id": &grouping.ID,
				})
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve definitions: " + err.Error()})
				}

				var filteredEntries []*modelcore.FinancialStatementDefinition
				for _, entry := range entries {
					if entry.FinancialStatementDefinitionEntriesID == nil {
						filteredEntries = append(filteredEntries, entry)
					}
				}

				grouping.FinancialStatementDefinitionEntries = filteredEntries
			}
		}
		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementGroupingManager.Filtered(context, ctx, fsGroupings))
	})

	// PUT /financial-statement-grouping/:financial_statement_grouping_id: Update a financial statement grouping. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-grouping/:financial_statement_grouping_id",
		Method:       "PUT",
		RequestType:  modelcore.FinancialStatementGroupingRequest{},
		ResponseType: modelcore.FinancialStatementGroupingResponse{},
		Note:         "Updates an existing financial statement grouping by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		groupingID, err := handlers.EngineUUIDParam(ctx, "financial_statement_grouping_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement grouping update failed (/financial-statement-grouping/:financial_statement_grouping_id), invalid grouping ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement grouping ID"})
		}
		reqBody, err := c.modelcore.FinancialStatementGroupingManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement grouping update failed (/financial-statement-grouping/:financial_statement_grouping_id), validation error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid grouping data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement grouping update failed (/financial-statement-grouping/:financial_statement_grouping_id), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for financial statement grouping (/financial-statement-grouping/:financial_statement_grouping_id)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement groupings"})
		}
		grouping, err := c.modelcore.FinancialStatementGroupingManager.GetByID(context, *groupingID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement grouping update failed (/financial-statement-grouping/:financial_statement_grouping_id), not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement grouping not found"})
		}
		grouping.Name = reqBody.Name
		grouping.Description = reqBody.Description
		grouping.Debit = reqBody.Debit
		grouping.Credit = reqBody.Credit
		grouping.IconMediaID = reqBody.IconMediaID
		grouping.UpdatedAt = time.Now().UTC()
		grouping.UpdatedByID = userOrg.UserID

		if err := c.modelcore.FinancialStatementGroupingManager.UpdateFields(context, grouping.ID, grouping); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement grouping update failed (/financial-statement-grouping/:financial_statement_grouping_id), db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update group: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated financial statement grouping (/financial-statement-grouping/:financial_statement_grouping_id): " + grouping.Name,
			Module:      "FinancialStatement",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementGroupingManager.ToModel(grouping))
	})

	// GET /financial-statement-definition: List all financial statement definitions for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-definition",
		Method:       "GET",
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
		Note:         "Returns all financial statement definitions for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement definitions"})
		}
		fsDefs, err := c.modelcore.FinancialStatementDefinitionManager.FindRaw(context, &modelcore.FinancialStatementDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve financial statement definitions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, fsDefs)
	})

	// POST /financial-statement-definition: Create a new financial statement definition. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-definition",
		Method:       "POST",
		RequestType:  modelcore.FinancialStatementDefinitionRequest{},
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
		Note:         "Creates a new financial statement definition for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Financial statement definition creation failed (/financial-statement-definition), validation error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Financial statement definition creation failed (/financial-statement-definition), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for financial statement definition (/financial-statement-definition)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create financial statement definitions"})
		}
		fsDefinition := &modelcore.FinancialStatementDefinition{
			OrganizationID:                        userOrg.OrganizationID,
			BranchID:                              *userOrg.BranchID,
			CreatedByID:                           userOrg.UserID,
			UpdatedByID:                           userOrg.UserID,
			FinancialStatementDefinitionEntriesID: req.FinancialStatementDefinitionEntriesID,
			FinancialStatementGroupingID:          req.FinancialStatementGroupingID,
			Name:                                  req.Name,
			Description:                           req.Description,
			Index:                                 req.Index,
			NameInTotal:                           req.NameInTotal,
			IsPosting:                             req.IsPosting,
			FinancialStatementType:                req.FinancialStatementType,
			CreatedAt:                             time.Now().UTC(),
			UpdatedAt:                             time.Now().UTC(),
		}
		if err := c.modelcore.FinancialStatementDefinitionManager.Create(context, fsDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Financial statement definition creation failed (/financial-statement-definition), db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create financial statement definition: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created financial statement definition (/financial-statement-definition): " + fsDefinition.Name,
			Module:      "FinancialStatement",
		})

		return ctx.JSON(http.StatusCreated, c.modelcore.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-definition/:financial_statement_definition_id: Update a financial statement definition. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-definition/:financial_statement_definition_id",
		Method:       "PUT",
		Note:         "Updates an existing financial statement definition by its ID.",
		RequestType:  modelcore.FinancialStatementDefinitionRequest{},
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := handlers.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement definition update failed (/financial-statement-definition/:financial_statement_definition_id), invalid ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		req, err := c.modelcore.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement definition update failed (/financial-statement-definition/:financial_statement_definition_id), validation error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement definition update failed (/financial-statement-definition/:financial_statement_definition_id), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized update attempt for financial statement definition (/financial-statement-definition/:financial_statement_definition_id)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement definitions"})
		}
		fsDefinition, err := c.modelcore.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement definition update failed (/financial-statement-definition/:financial_statement_definition_id), not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		fsDefinition.FinancialStatementDefinitionEntriesID = req.FinancialStatementDefinitionEntriesID
		fsDefinition.Name = req.Name
		fsDefinition.Description = req.Description
		fsDefinition.Index = req.Index
		fsDefinition.NameInTotal = req.NameInTotal
		fsDefinition.IsPosting = req.IsPosting
		fsDefinition.FinancialStatementType = req.FinancialStatementType
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID

		if err := c.modelcore.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Financial statement definition update failed (/financial-statement-definition/:financial_statement_definition_id), db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated financial statement definition (/financial-statement-definition/:financial_statement_definition_id): " + fsDefinition.Name,
			Module:      "FinancialStatement",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// POST /financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect: Connect an account to a financial statement definition. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect",
		Method:       "POST",
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
		Note:         "Connects an account to a financial statement definition by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := handlers.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), invalid FS definition ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), invalid account ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized connect attempt for account to FS definition (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to connect accounts"})
		}
		fsDefinition, err := c.modelcore.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), FS definition not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		if fsDefinition.OrganizationID != userOrg.OrganizationID || fsDefinition.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), FS definition wrong org/branch.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Financial statement definition does not belong to your organization/branch"})
		}
		account, err := c.modelcore.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), account not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), account wrong org/branch.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to your organization/branch"})
		}
		account.FinancialStatementDefinitionID = fsDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID
		if err := c.modelcore.AccountManager.UpdateFields(context, account.ID, account); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), account db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID
		if err := c.modelcore.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), FS definition db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Connected account to FS definition (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect) for account: " + account.Name,
			Module:      "FinancialStatement",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-definition/:financial_statement_definition_id/index/:index: Update the index of a financial statement definition. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-definition/:financial_statement_definition_id/index/:index",
		Method:       "PUT",
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
		Note:         "Updates the index of a financial statement definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := handlers.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS definition index update failed (/financial-statement-definition/:financial_statement_definition_id/index/:index), invalid ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS definition index update failed (/financial-statement-definition/:financial_statement_definition_id/index/:index), invalid index.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index value"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS definition index update failed (/financial-statement-definition/:financial_statement_definition_id/index/:index), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized FS definition index update attempt (/financial-statement-definition/:financial_statement_definition_id/index/:index)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement definition index"})
		}
		fsDefinition, err := c.modelcore.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS definition index update failed (/financial-statement-definition/:financial_statement_definition_id/index/:index), not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		fsDefinition.Index = index
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID
		if err := c.modelcore.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS definition index update failed (/financial-statement-definition/:financial_statement_definition_id/index/:index), db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update index: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated FS definition index (/financial-statement-definition/:financial_statement_definition_id/index/:index): " + fsDefinition.Name,
			Module:      "FinancialStatement",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index: Update the index of an account within a financial statement definition. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index",
		Method:       "PUT",
		ResponseType: modelcore.FinancialStatementDefinitionResponse{},
		Note:         "Updates the index of an account within a financial statement definition and reorders accordingly.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := handlers.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), invalid FS definition ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), invalid account ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		type UpdateAccountIndexRequest struct {
			FinancialStatementDefinitionIndex int `json:"financial_statement_definition_index"`
			AccountIndex                      int `json:"account_index"`
		}
		var reqBody UpdateAccountIndexRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), invalid payload.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Unauthorized FS grouping/account index update attempt (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update account index"})
		}
		fsDefinition, err := c.modelcore.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), FS definition not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		account, err := c.modelcore.AccountManager.GetByID(context, *accountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), account not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.FinancialStatementDefinitionID == nil || *account.FinancialStatementDefinitionID != *fsDefinitionID {
			account.FinancialStatementDefinitionID = fsDefinitionID
		}
		accounts, err := c.modelcore.AccountManager.Find(context, &modelcore.Account{
			FinancialStatementDefinitionID: fsDefinitionID,
			OrganizationID:                 userOrg.OrganizationID,
			BranchID:                       *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), account find error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		var updatedAccounts []*modelcore.Account
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
		updatedAccounts = append(updatedAccounts[:reqBody.AccountIndex], append([]*modelcore.Account{account}, updatedAccounts[reqBody.AccountIndex:]...)...)
		for idx, acc := range updatedAccounts {
			acc.Index = idx
			acc.UpdatedAt = time.Now().UTC()
			acc.UpdatedByID = userOrg.UserID
			if err := c.modelcore.AccountManager.UpdateFields(context, acc.ID, acc); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), update account error: " + err.Error(),
					Module:      "FinancialStatement",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account: " + err.Error()})
			}
		}
		if fsDefinition.Index != reqBody.FinancialStatementDefinitionIndex {
			fsDefinition.Index = reqBody.FinancialStatementDefinitionIndex
			fsDefinition.UpdatedAt = time.Now().UTC()
			fsDefinition.UpdatedByID = userOrg.UserID
			if err := c.modelcore.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "FS grouping/account index update failed (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index), update FS definition error: " + err.Error(),
					Module:      "FinancialStatement",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition index: " + err.Error()})
			}
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated account index within FS definition (/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index): " + account.Name,
			Module:      "FinancialStatement",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// DELETE /financial-statement-definition/:financial_statement_definition_id: Delete a financial statement definition by ID, only if no accounts are linked. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/financial-statement-definition/:financial_statement_definition_id",
		Method: "DELETE",
		Note:   "Deletes a financial statement definition by its ID, only if no accounts are linked.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := handlers.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), invalid ID.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), user org error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != modelcore.UserOrganizationTypeOwner && userOrg.UserType != modelcore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Unauthorized delete attempt for FS definition (/financial-statement-definition/:financial_statement_definition_id)",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete financial statement definitions"})
		}
		fsDefinition, err := c.modelcore.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), not found.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		if len(fsDefinition.FinancialStatementDefinitionEntries) > 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), has sub-entries.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: financial statement definition has sub-entries"})
		}
		accounts, err := c.modelcore.AccountManager.Find(context, &modelcore.Account{
			FinancialStatementDefinitionID: fsDefinitionID,
			OrganizationID:                 userOrg.OrganizationID,
			BranchID:                       *userOrg.BranchID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), account find error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check accounts linked: " + err.Error()})
		}
		if len(accounts) > 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), has linked accounts.",
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: accounts are linked to this financial statement definition"})
		}
		if err := c.modelcore.FinancialStatementDefinitionManager.DeleteByID(context, fsDefinition.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "FS definition delete failed (/financial-statement-definition/:financial_statement_definition_id), db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete financial statement definition: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted FS definition (/financial-statement-definition/:financial_statement_definition_id): " + fsDefinition.Name,
			Module:      "FinancialStatement",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
