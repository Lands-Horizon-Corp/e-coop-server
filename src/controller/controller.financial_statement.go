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
		Route:    "/financial-statement",
		Method:   "GET",
		Response: "FinancialStatement[]",
		Note:     "List all financial statements for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fs, err := c.model.FinancialStatementDefinitionManager.FindRaw(context, &model.FinancialStatementDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, fs)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement",
		Method:   "POST",
		Request:  "FinancialStatementRequest",
		Response: "FinancialStatementResponse",
		Note:     "Create a new financial statement",
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
		fs := &model.FinancialStatementDefinition{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CreatedByID:    userOrg.UserID,
			UpdatedByID:    userOrg.UserID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			// Add other fields as needed
		}
		if err := c.model.FinancialStatementDefinitionManager.Create(context, fs); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fs))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement/:financial_statement_id",
		Method:   "PUT",
		Request:  "FinancialStatementRequest",
		Response: "FinancialStatementResponse",
		Note:     "Update an existing financial statement",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsID, err := horizon.EngineUUIDParam(ctx, "financial_statement_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement ID")
		}
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
		fs, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement")
		}
		fs.Name = req.Name
		fs.Description = req.Description
		fs.UpdatedAt = time.Now().UTC()
		fs.UpdatedByID = userOrg.UserID
		// Update other fields as needed
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fs.ID, fs); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fs))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/financial-statement/:financial_statement_id/index/:index",
		Method:   "PUT",
		Response: "FinancialStatementResponse",
		Note:     "Update the index of a financial statement",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fsID, err := horizon.EngineUUIDParam(ctx, "financial_statement_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid financial statement ID")
		}
		index, err := strconv.Atoi(ctx.Param("index"))
		if err != nil {
			return c.BadRequest(ctx, "Invalid index value")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		fs, err := c.model.FinancialStatementDefinitionManager.GetByID(context, *fsID)
		if err != nil {
			return c.NotFound(ctx, "Financial Statement")
		}
		fs.Index = index
		fs.UpdatedAt = time.Now().UTC()
		fs.UpdatedByID = userOrg.UserID
		if err := c.model.FinancialStatementDefinitionManager.UpdateFields(context, fs.ID, fs); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.FinancialStatementDefinitionManager.ToModel(fs))
	})
}
