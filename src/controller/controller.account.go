package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) AccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/account/search",
		Method:   "GET",
		Response: "IAccount[]",
		Note:     "List all accounts for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		account, err := c.model.AccountManager.Find(context, &model.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountManager.Pagination(context, ctx, account))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account",
		Method:   "POST",
		Response: "IAccount",
		Note:     "Create a new account",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountManager.Validate(ctx)
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

		account := &model.Account{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,

			GeneralLedgerDefinitionID:             req.GeneralLedgerDefinitionID,
			FinancialStatementDefinitionID:        req.FinancialStatementDefinitionID,
			AccountClassificationID:               req.AccountClassificationID,
			AccountCategoryID:                     req.AccountCategoryID,
			MemberTypeID:                          req.MemberTypeID,
			Name:                                  req.Name,
			Description:                           req.Description,
			MinAmount:                             req.MinAmount,
			MaxAmount:                             req.MaxAmount,
			Index:                                 req.Index,
			Type:                                  req.Type,
			IsInternal:                            req.IsInternal,
			CashOnHand:                            req.CashOnHand,
			PaidUpShareCapital:                    req.PaidUpShareCapital,
			ComputationType:                       req.ComputationType,
			FinesAmort:                            req.FinesAmort,
			FinesMaturity:                         req.FinesMaturity,
			InterestStandard:                      req.InterestStandard,
			InterestSecured:                       req.InterestSecured,
			ComputationSheetID:                    req.ComputationSheetID,
			CohCibFinesGracePeriodEntryCashHand:   req.CohCibFinesGracePeriodEntryCashHand,
			CohCibFinesGracePeriodEntryCashInBank: req.CohCibFinesGracePeriodEntryCashInBank,
			CohCibFinesGracePeriodEntryDailyAmortization:       req.CohCibFinesGracePeriodEntryDailyAmortization,
			CohCibFinesGracePeriodEntryDailyMaturity:           req.CohCibFinesGracePeriodEntryDailyMaturity,
			CohCibFinesGracePeriodEntryWeeklyAmortization:      req.CohCibFinesGracePeriodEntryWeeklyAmortization,
			CohCibFinesGracePeriodEntryWeeklyMaturity:          req.CohCibFinesGracePeriodEntryWeeklyMaturity,
			CohCibFinesGracePeriodEntryMonthlyAmortization:     req.CohCibFinesGracePeriodEntryMonthlyAmortization,
			CohCibFinesGracePeriodEntryMonthlyMaturity:         req.CohCibFinesGracePeriodEntryMonthlyMaturity,
			CohCibFinesGracePeriodEntrySemiMonthlyAmortization: req.CohCibFinesGracePeriodEntrySemiMonthlyAmortization,
			CohCibFinesGracePeriodEntrySemiMonthlyMaturity:     req.CohCibFinesGracePeriodEntrySemiMonthlyMaturity,
			CohCibFinesGracePeriodEntryQuarterlyAmortization:   req.CohCibFinesGracePeriodEntryQuarterlyAmortization,
			CohCibFinesGracePeriodEntryQuarterlyMaturity:       req.CohCibFinesGracePeriodEntryQuarterlyMaturity,
			CohCibFinesGracePeriodEntrySemiAnualAmortization:   req.CohCibFinesGracePeriodEntrySemiAnualAmortization,
			CohCibFinesGracePeriodEntrySemiAnualMaturity:       req.CohCibFinesGracePeriodEntrySemiAnualMaturity,
			CohCibFinesGracePeriodEntryLumpsumAmortization:     req.CohCibFinesGracePeriodEntryLumpsumAmortization,
			CohCibFinesGracePeriodEntryLumpsumMaturity:         req.CohCibFinesGracePeriodEntryLumpsumMaturity,
			FinancialStatementType:                             string(req.FinancialStatementType),
			GeneralLedgerType:                                  req.GeneralLedgerType,
			AlternativeCode:                                    req.AlternativeCode,
			FinesGracePeriodAmortization:                       req.FinesGracePeriodAmortization,
			AdditionalGracePeriod:                              req.AdditionalGracePeriod,
			NumberGracePeriodDaily:                             req.NumberGracePeriodDaily,
			FinesGracePeriodMaturity:                           req.FinesGracePeriodMaturity,
			YearlySubscriptionFee:                              req.YearlySubscriptionFee,
			LoanCutOffDays:                                     req.LoanCutOffDays,
			LumpsumComputationType:                             string(req.LumpsumComputationType),
			InterestFinesComputationDiminishing:                string(req.InterestFinesComputationDiminishing),
			InterestFinesComputationDiminishingStraightYearly:  string(req.InterestFinesComputationDiminishingStraightYearly),
			EarnedUnearnedInterest:                             string(req.EarnedUnearnedInterest),
			LoanSavingType:                                     string(req.LoanSavingType),
			InterestDeduction:                                  string(req.InterestDeduction),
			OtherDeductionEntry:                                string(req.OtherDeductionEntry),
			InterestSavingTypeDiminishingStraight:              string(req.InterestSavingTypeDiminishingStraight),
			OtherInformationOfAnAccount:                        string(req.OtherInformationOfAnAccount),
			HeaderRow:                                          req.HeaderRow,
			CenterRow:                                          req.CenterRow,
			TotalRow:                                           req.TotalRow,
			GeneralLedgerGroupingExcludeAccount:                req.GeneralLedgerGroupingExcludeAccount,
		}

		if err := c.model.AccountManager.Create(context, account); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, account)
	})

	// Account Category Search
	req.RegisterRoute(horizon.Route{
		Route:    "/account-category/search",
		Method:   "GET",
		Response: "IAccount[]",
		Note:     "List all accounts for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		account, err := c.model.AccountManager.Find(context, &model.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountManager.Pagination(context, ctx, account))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-category/:account_category_id",
		Method:   "GET",
		Response: "IAccountCategory",
		Note:     "Get an account category by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountCategoryID, err := horizon.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return err
		}
		accountCategory, err := c.model.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountCategoryManager.ToModel(accountCategory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-category",
		Method:   "POST",
		Response: "IAccount",
		Note:     "Create a new account category",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountCategoryManager.Validate(ctx)
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

		accountCategory := &model.AccountCategory{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.model.AccountCategoryManager.Create(context, accountCategory); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, accountCategory)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-category/:account_category_id",
		Method:   "PUT",
		Response: "IAccountCategory",
		Note:     "Update an account category",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountCategoryManager.Validate(ctx)
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
		accountCategoryID, err := horizon.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return err
		}
		accountCategory, err := c.model.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		accountCategory.UpdatedByID = userOrg.UserID
		accountCategory.UpdatedAt = time.Now().UTC()
		accountCategory.Name = req.Name
		accountCategory.Description = req.Description
		accountCategory.BranchID = *userOrg.BranchID
		accountCategory.OrganizationID = userOrg.OrganizationID
		if err := c.model.AccountCategoryManager.UpdateFields(context, accountCategory.ID, accountCategory); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountCategoryManager.ToModel(accountCategory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-category/:account_category_id",
		Method:   "DELETE",
		Response: "IAccountCategory",
		Note:     "Delete an account category",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		accountCategoryID, err := horizon.EngineUUIDParam(ctx, "account_category_id")
		if err != nil {
			return err
		}
		accountCategory, err := c.model.AccountCategoryManager.GetByID(context, *accountCategoryID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.model.AccountCategoryManager.DeleteByID(context, accountCategory.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountCategoryManager.ToModel(accountCategory))
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/account-category/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple account categories",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.AccountCategoryManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("AccountCategory with ID %s", rawID))
			}
			if err := c.model.AccountCategoryManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Add this code to the AccountController function
	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification/search",
		Method:   "GET",
		Response: "IAccount[]",
		Note:     "List all accounts classification for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		account, err := c.model.AccountClassificationManager.Find(context, &model.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.Pagination(context, ctx, account))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification/:account_classification_id",
		Method:   "GET",
		Response: "IAccountClassification",
		Note:     "Get an account classification by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountClassificationID, err := horizon.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return err
		}
		accountClassification, err := c.model.AccountClassificationManager.GetByID(context, *accountClassificationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(accountClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification",
		Method:   "POST",
		Response: "IAccountClassification",
		Note:     "Create a new account classification",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountClassificationManager.Validate(ctx)
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

		accountClassification := &model.AccountClassification{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Name:           req.Name,
			Description:    req.Description,
		}

		if err := c.model.AccountClassificationManager.Create(context, accountClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusCreated, accountClassification)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification/:account_classification_id",
		Method:   "PUT",
		Response: "IAccountClassification",
		Note:     "Update an account classification",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountClassificationManager.Validate(ctx)
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
		accountClassificationID, err := horizon.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return err
		}
		accountClassification, err := c.model.AccountClassificationManager.GetByID(context, *accountClassificationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		accountClassification.UpdatedByID = userOrg.UserID
		accountClassification.UpdatedAt = time.Now().UTC()
		accountClassification.Name = req.Name
		accountClassification.Description = req.Description
		accountClassification.BranchID = *userOrg.BranchID
		accountClassification.OrganizationID = userOrg.OrganizationID
		if err := c.model.AccountClassificationManager.UpdateFields(context, accountClassification.ID, accountClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(accountClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification/:account_classification_id",
		Method:   "DELETE",
		Response: "IAccountClassification",
		Note:     "Delete an account classification",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		accountClassificationID, err := horizon.EngineUUIDParam(ctx, "account_classification_id")
		if err != nil {
			return err
		}
		accountClassification, err := c.model.AccountClassificationManager.GetByID(context, *accountClassificationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.model.AccountClassificationManager.DeleteByID(context, accountClassification.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.ToModel(accountClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/account-classification/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple account classifications",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.AccountClassificationManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("AccountClassification with ID %s", rawID))
			}
			if err := c.model.AccountClassificationManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Also add a search endpoint for account classifications
	req.RegisterRoute(horizon.Route{
		Route:    "/account-classification/search",
		Method:   "GET",
		Response: "IAccountClassification[]",
		Note:     "List all account classifications for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		accountClassifications, err := c.model.AccountClassificationManager.Find(context, &model.AccountClassification{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountClassificationManager.Pagination(context, ctx, accountClassifications))
	})

}
