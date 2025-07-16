package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) GeneralLedgerController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-accounts-grouping",
		Method:   "GET",
		Response: "GeneralLedgerAccountsGrouping[]",
		Note:     "List all general ledger accounts grouping for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		gl, err := c.model.GeneralLedgerAccountsGroupingManager.Find(context, &model.GeneralLedgerAccountsGrouping{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerAccountsGroupingManager.ToModels(gl))
	})
	// ...existing code...

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-accounts-grouping/:general_ledger_accounts_grouping_id",
		Method:   "PUT",
		Request:  "GeneralLedgerAccountsGroupingRequest",
		Response: "GeneralLedgerAccountsGroupingResponse",
		Note:     "Update an existing general ledger accounts grouping",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get and validate grouping ID
		groupingID, err := horizon.EngineUUIDParam(ctx, "general_ledger_accounts_grouping_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid grouping ID")
		}

		// Validate request
		reqBody, err := c.model.GeneralLedgerAccountsGroupingManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		// Get current user organization
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		// Get existing grouping
		grouping, err := c.model.GeneralLedgerAccountsGroupingManager.GetByID(context, *groupingID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Accounts Grouping")
		}

		// Update fields
		grouping.Name = reqBody.Name
		grouping.Description = reqBody.Description
		grouping.UpdatedAt = time.Now().UTC()
		grouping.UpdatedByID = userOrg.UserID
		grouping.FromCode = reqBody.FromCode
		grouping.ToCode = reqBody.ToCode

		// Parse Debit and Credit from string to float64 if needed
		if debit, err := strconv.ParseFloat(reqBody.Debit, 64); err == nil {
			grouping.Debit = debit
		}
		if credit, err := strconv.ParseFloat(reqBody.Credit, 64); err == nil {
			grouping.Credit = credit
		}

		// Save update
		if err := c.model.GeneralLedgerAccountsGroupingManager.UpdateFields(context, grouping.ID, grouping); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerAccountsGroupingManager.ToModel(grouping))
	})
	// ...existing code...

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition",
		Method:   "GET",
		Response: "GeneralLedgerDefinition[]",
		Note:     "List all general ledger definitions for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		gl, err := c.model.GeneralLedgerDefinitionManager.FindRaw(context, &model.GeneralLedgerDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, gl)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition",
		Method:   "POST",
		Request:  "GeneralLedgerDefinitionRequest",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Create a new general ledger definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request
		req, err := c.model.GeneralLedgerDefinitionManager.Validate(ctx)
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

		// Create general ledger definition
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

		// Create the general ledger definition
		if err := c.model.GeneralLedgerDefinitionManager.Create(context, glDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id",
		Method:   "PUT",
		Request:  "GeneralLedgerDefinitionRequest",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Update an existing general ledger definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get and validate general ledger definition ID
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid general ledger definition ID")
		}

		// Validate request
		req, err := c.model.GeneralLedgerDefinitionManager.Validate(ctx)
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

		// Get existing general ledger definition
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Definition")
		}

		// Update fields
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

		// Update the general ledger definition
		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id/account/:account_id/connect",
		Method:   "POST",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Connect an account to a general ledger definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid general ledger definition ID")
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

		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Definition")
		}

		if glDefinition.OrganizationID != userOrg.OrganizationID || glDefinition.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "General ledger definition not found in your organization")
		}

		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return c.NotFound(ctx, "Account")
		}

		if account.OrganizationID != userOrg.OrganizationID || account.BranchID != *userOrg.BranchID {
			return c.BadRequest(ctx, "Account not found in your organization")
		}

		account.GeneralLedgerDefinitionID = glDefinitionID
		account.UpdatedAt = time.Now().UTC()
		account.UpdatedByID = userOrg.UserID

		if err := c.model.AccountManager.UpdateFields(context, account.ID, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID

		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_ledger_definition_id/index/:index",
		Method:   "PUT",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Update the index of a general ledger definition",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Get and validate general ledger definition ID
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid general ledger definition ID")
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

		// Get existing general ledger definition
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Definition")
		}

		// Update only the index and audit fields
		glDefinition.Index = index
		glDefinition.UpdatedAt = time.Now().UTC()
		glDefinition.UpdatedByID = userOrg.UserID

		// Update the general ledger definition
		if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-grouping/general-ledger-definition/:general_ledger_definition_id/account/:account_id/index",
		Method:   "PUT",
		Request:  "UpdateAccountIndexRequest {general_ledger_definition_index: int, account_index: int}",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Update the index of an account within a general ledger definition and reorder accordingly",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_ledger_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid general ledger definition ID")
		}
		accountID, err := horizon.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account ID")
		}
		type UpdateAccountIndexRequest struct {
			GeneralLedgerDefinitionIndex int `json:"general_ledger_definition_index"`
			AccountIndex                 int `json:"account_index"`
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
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Definition")
		}
		account, err := c.model.AccountManager.GetByID(context, *accountID)
		if err != nil {
			return c.NotFound(ctx, "Account")
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

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
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		// Optionally, update the GL Definition's index if needed
		if glDefinition.Index != reqBody.GeneralLedgerDefinitionIndex {
			glDefinition.Index = reqBody.GeneralLedgerDefinitionIndex
			glDefinition.UpdatedAt = time.Now().UTC()
			glDefinition.UpdatedByID = userOrg.UserID
			if err := c.model.GeneralLedgerDefinitionManager.UpdateFields(context, glDefinition.ID, glDefinition); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerDefinitionManager.ToModel(glDefinition))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-accounting-ledger/member-profile/:member_profile_id/total",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Get total amount for a specific member profile's general ledger entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		Note:     "Get total amount for a specific member profile's general ledger entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		if userOrg.BranchID == nil {
			return c.BadRequest(ctx, "Branch ID is missing for user organization")
		}
		entries, err := c.model.MemberAccountingLedgerManager.Find(context, &model.MemberAccountingLedger{
			MemberProfileID: *memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAccountingLedgerManager.Pagination(context, ctx, entries))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger/member-profile/:member_profile_id/account/:account_id/search",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Get total amount for a specific member profile's general ledger entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
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
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.Pagination(context, ctx, entries))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger/member-profile/:member_profile_id/account/:account_id/total",
		Method:   "GET",
		Response: "MemberGeneralLedgerTotal",
		Note:     "Get total amount for a specific member profile's general ledger entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
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
		entries, err := c.model.GeneralLedgerManager.Find(context, &model.GeneralLedger{
			MemberProfileID: memberProfileID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			AccountID:       accountID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		Note:     "Get all general ledger entries for an account with pagination",
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
	// Delete a general ledger definition by ID, only if no accounts are linked
	req.RegisterRoute(horizon.Route{
		Route:    "/general-ledger-definition/:general_definition_id",
		Method:   "DELETE",
		Response: "GeneralLedgerDefinitionResponse",
		Note:     "Delete a general ledger definition by ID, only if no accounts are linked",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		glDefinitionID, err := horizon.EngineUUIDParam(ctx, "general_definition_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid general ledger definition ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		glDefinition, err := c.model.GeneralLedgerDefinitionManager.GetByID(context, *glDefinitionID)
		if err != nil {
			return c.NotFound(ctx, "General Ledger Definition")
		}
		if len(glDefinition.GeneralLedgerDefinitionEntries) > 0 {
			return c.BadRequest(ctx, "Cannot delete: general ledger definition has entries")
		}
		// Check if any accounts are linked to this general ledger definition
		accounts, err := c.model.AccountManager.Find(context, &model.Account{
			GeneralLedgerDefinitionID: glDefinitionID,
			OrganizationID:            userOrg.OrganizationID,
			BranchID:                  *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		if len(accounts) > 0 {
			return c.BadRequest(ctx, "Cannot delete: accounts are linked to this general ledger definition")
		}
		if err := c.model.GeneralLedgerDefinitionManager.DeleteByID(context, glDefinition.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
