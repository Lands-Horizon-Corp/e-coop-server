package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// FinancialStatementController manages endpoints for financial statement groupings and definitions.
func (c *Controller) FinancialStatementController() {
	req := c.provider.Service.Request

	// GET /financial-statement-grouping: List all financial statement groupings for the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-grouping",
		Method:   "GET",
		Response: "FinancialStatementGrouping[]",
		Note:     "Returns all financial statement groupings for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement groupings"})
		}
		fsGroupings, err := c.model.FinancialStatementGroupingManager.Find(context, &model.FinancialStatementGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve financial statement groupings: " + err.Error()})
		}
		for _, grouping := range fsGroupings {
			if grouping != nil {
				grouping.FinancialStatementDefinitionEntries = []*model.FinancialStatementDefinition{}
				entries, err := c.model.FinancialStatementDefinitionManager.FindWithConditions(context, map[string]any{
					"organization_id":                 userOrg.OrganizationID,
					"branch_id":                       *userOrg.BranchID,
					"financial_statement_grouping_id": &grouping.ID,
				})
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve definitions: " + err.Error()})
				}

				var filteredEntries []*model.FinancialStatementDefinition
				for _, entry := range entries {
					if entry.FinancialStatementDefinitionEntriesID == nil {
						filteredEntries = append(filteredEntries, entry)
					}
				}

				grouping.FinancialStatementDefinitionEntries = filteredEntries
			}
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementGroupingManager.ToModels(fsGroupings))
	})

	// PUT /financial-statement-grouping/:financial_statement_grouping_id: Update a financial statement grouping.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-grouping/:financial_statement_grouping_id",
		Method:   "PUT",
		Request:  "FinancialStatementGroupingRequest",
		Response: "FinancialStatementGroupingResponse",
		Note:     "Updates an existing financial statement grouping by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		groupingID, err := horizon.EngineUUIDParam(ctx, "financial_statement_grouping_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement grouping ID"})
		}
		reqBody, err := c.model.FinancialStatementGroupingManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid grouping data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement groupings"})
		}
		grouping, err := c.model.FinancialStatementGroupingManager.GetByID(context, *groupingID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement grouping not found"})
		}
		grouping.Name = reqBody.Name
		grouping.Description = reqBody.Description
		grouping.Debit = reqBody.Debit
		grouping.Credit = reqBody.Credit
		grouping.Code = reqBody.Code
		grouping.IconMediaID = reqBody.IconMediaID
		grouping.UpdatedAt = time.Now().UTC()
		grouping.UpdatedByID = userOrg.UserID

		if err := c.model.FinancialStatementGroupingManager.UpdateFields(context, grouping.ID, grouping); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update group: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementGroupingManager.ToModel(grouping))
	})

	// GET /financial-statement-definition: List all financial statement definitions for the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition",
		Method:   "GET",
		Response: "FinancialStatementDefinition[]",
		Note:     "Returns all financial statement definitions for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement definitions"})
		}
		fsDefs, err := c.model.FinancialStatementDefinitionManager.FindRaw(context, &model.FinancialStatementDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve financial statement definitions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, fsDefs)
	})

	// POST /financial-statement-definition: Create a new financial statement definition.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition",
		Method:   "POST",
		Request:  "FinancialStatementDefinitionRequest",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Creates a new financial statement definition for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create financial statement definitions"})
		}
		fsDefinition := &model.FinancialStatementDefinition{
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
		if err := c.model.FinancialStatementDefinitionManager.Create(context, fsDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create financial statement definition: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-definition/:financial_statement_definition_id: Update a financial statement definition.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id",
		Method:   "PUT",
		Request:  "FinancialStatementDefinitionRequest",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Updates an existing financial statement definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		req, err := c.model.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement definitions"})
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
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

		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// POST /financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect: Connect an account to a financial statement definition.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect",
		Method:   "POST",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Connects an account to a financial statement definition by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to connect accounts"})
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		if fsDefinition.OrganizationID != userOrg.OrganizationID || fsDefinition.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Financial statement definition does not belong to your organization/branch"})
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Account does not belong to your organization/branch"})
		}
		account.FinancialStatementDefinitionID = fsDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID
		if err := c.model.AccountManager.UpdateFields(context, account.ID, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-definition/:financial_statement_definition_id/index/:index: Update the index of a financial statement definition.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id/index/:index",
		Method:   "PUT",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Updates the index of a financial statement definition by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index value"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update financial statement definition index"})
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		fsDefinition.Index = index
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update index: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// PUT /financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index: Update the index of an account within a financial statement definition.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index",
		Method:   "PUT",
		Request:  "UpdateAccountIndexRequest {financial_statement_definition_index: int, account_index: int}",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Updates the index of an account within a financial statement definition and reorders accordingly.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		type UpdateAccountIndexRequest struct {
			FinancialStatementDefinitionIndex int `json:"financial_statement_definition_index"`
			AccountIndex                      int `json:"account_index"`
		}
		var reqBody UpdateAccountIndexRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update account index"})
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if account.FinancialStatementDefinitionID == nil || *account.FinancialStatementDefinitionID != *fsDefinitionID {
			account.FinancialStatementDefinitionID = fsDefinitionID
		}
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			FinancialStatementDefinitionID: fsDefinitionID,
			OrganizationID:                 userOrg.OrganizationID,
			BranchID:                       *userOrg.BranchID,
		})
		if err != nil {
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
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update account: " + err.Error()})
			}
		}
		if fsDefinition.Index != reqBody.FinancialStatementDefinitionIndex {
			fsDefinition.Index = reqBody.FinancialStatementDefinitionIndex
			fsDefinition.UpdatedAt = time.Now().UTC()
			fsDefinition.UpdatedByID = userOrg.UserID
			if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update financial statement definition index: " + err.Error()})
			}
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	// GET /financial-statement/account/:account_id/search: Get all financial statement entries for an account with pagination.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement/account/:account_id/search",
		Method:   "GET",
		Response: "FinancialStatement[]",
		Note:     "Returns all financial statement entries for an account with pagination.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view financial statement entries"})
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve ledger entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})

	// DELETE /financial-statement-definition/:financial_statement_definition_id: Delete a financial statement definition by ID, only if no accounts are linked.
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id",
		Method:   "DELETE",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Deletes a financial statement definition by its ID, only if no accounts are linked.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid financial statement definition ID"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete financial statement definitions"})
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Financial statement definition not found"})
		}
		if len(fsDefinition.FinancialStatementDefinitionEntries) > 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: financial statement definition has sub-entries"})
		}
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			FinancialStatementDefinitionID: fsDefinitionID,
			OrganizationID:                 userOrg.OrganizationID,
			BranchID:                       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check accounts linked: " + err.Error()})
		}
		if len(accounts) > 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot delete: accounts are linked to this financial statement definition"})
		}
		if err := c.model.FinancialStatementDefinitionManager.DeleteByID(context, fsDefinition.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete financial statement definition: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
