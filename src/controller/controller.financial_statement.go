package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) FinancialStatementController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statements-grouping",
		Method:   "GET",
		Response: "FinancialStatementGrouping[]",
		Note:     "List all financial statement groupings for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fsGroupings, err := c.model.FinancialStatementGroupingManager.Find(context, &model.FinancialStatementGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition",
		Method:   "GET",
		Response: "FinancialStatementDefinition[]",
		Note:     "List all financial statement definitions for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fsDefs, err := c.model.FinancialStatementDefinitionManager.FindRaw(context, &model.FinancialStatementDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, fsDefs)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition",
		Method:   "POST",
		Request:  "FinancialStatementDefinitionRequest",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Create a new financial statement definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
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
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id",
		Method:   "PUT",
		Request:  "FinancialStatementDefinitionRequest",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Update an existing financial statement definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get and validate financial statement definition ID
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement definition ID")
		}

		// Validate request
		req, err := c.model.FinancialStatementDefinitionManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		// Check authorization
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get existing financial statement definition
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement Definition")
		}

		// Update fields
		fsDefinition.FinancialStatementDefinitionEntriesID = req.FinancialStatementDefinitionEntriesID
		fsDefinition.Name = req.Name
		fsDefinition.Description = req.Description
		fsDefinition.Index = req.Index
		fsDefinition.NameInTotal = req.NameInTotal
		fsDefinition.IsPosting = req.IsPosting
		fsDefinition.FinancialStatementType = req.FinancialStatementType
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID

		// Update the financial statement definition
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect",
		Method:   "POST",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Connect an account to a financial statement definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement definition ID")
		}

		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account ID")
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement Definition")
		}

		if fsDefinition.OrganizationID != userOrg.OrganizationID || fsDefinition.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Financial statement definition not found in your organization")
		}

		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return c.NotFound(ctx, "Account")
		}

		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Account not found in your organization")
		}

		account.FinancialStatementDefinitionID = fsDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		if err := c.model.AccountManager.UpdateFields(context, account.ID, account); err != nil {
			return c.InternalServerError(ctx, err)
		}

		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID

		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id/index/:index",
		Method:   "PUT",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Update the index of a financial statement definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get and validate financial statement definition ID
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement definition ID")
		}

		// Get and validate index parameter
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			return c.BadRequest(ctx, "Invalid index value")
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}

		// Check authorization
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get existing financial statement definition
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement Definition")
		}

		// Update only the index and audit fields
		fsDefinition.Index = index
		fsDefinition.UpdatedAt = time.Now().UTC()
		fsDefinition.UpdatedByID = userOrg.UserID

		// Update the financial statement definition
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-grouping/financial-statement-definition/:financial_statement_definition_id/account/:account_id/index",
		Method:   "PUT",
		Request:  "UpdateAccountIndexRequest {financial_statement_definition_index: int, account_index: int}",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Update the index of an account within a financial statement definition and reorder accordingly",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement definition ID")
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account ID")
		}
		type UpdateAccountIndexRequest struct {
			FinancialStatementDefinitionIndex int `json:"financial_statement_definition_index"`
			AccountIndex                      int `json:"account_index"`
		}
		var reqBody UpdateAccountIndexRequest
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid payload")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement Definition")
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return c.NotFound(ctx, "Account")
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
			return c.InternalServerError(ctx, err)
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
				return c.InternalServerError(ctx, err)
			}
		}

		// Optionally, update the FS Definition's index if needed
		if fsDefinition.Index != reqBody.FinancialStatementDefinitionIndex {
			fsDefinition.Index = reqBody.FinancialStatementDefinitionIndex
			fsDefinition.UpdatedAt = time.Now().UTC()
			fsDefinition.UpdatedByID = userOrg.UserID
			if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fsDefinition.ID, fsDefinition); err != nil {
				return c.InternalServerError(ctx, err)
			}
		}

		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fsDefinition))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement/account/:account_id/search",
		Method:   "GET",
		Response: "FinancialStatement[]",
		Note:     "Get all financial statement entries for an account with pagination",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			AccountID:      accountID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement-definition/:financial_statement_definition_id",
		Method:   "DELETE",
		Response: "FinancialStatementDefinitionResponse",
		Note:     "Delete a financial statement definition by ID, only if no accounts are linked",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsDefinitionID, err := horizon.EngineUUIDParam(ctx, "financial_statement_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement definition ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fsDefinition, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement Definition")
		}
		if len(fsDefinition.FinancialStatementDefinitionEntries) > 0 {
			return c.BadRequest(ctx, "Cannot delete: financial statement definition has entries")
		}
		// Check if any accounts are linked to this financial statement definition
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			FinancialStatementDefinitionID: fsDefinitionID,
			OrganizationID:                 userOrg.OrganizationID,
			BranchID:                       *userOrg.BranchID,
		})
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		if len(accounts) > 0 {
			return c.BadRequest(ctx, "Cannot delete: accounts are linked to this financial statement definition")
		}
		if err := c.model.FinancialStatementDefinitionManager.DeleteByID(context, fsDefinition.ID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
