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

func (c *Controller) LoanPurposeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose",
		Method:   "GET",
		Response: "TLoanPurpose[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		purposes, err := c.model.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "LoanPurpose")
		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.ToModels(purposes))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/search",
		Method:   "GET",
		Request:  "Filter<ILoanPurpose>",
		Response: "Paginated<ILoanPurpose>",
		Note:     "Get pagination loan purpose",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.LoanPurposeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/:loan_purpose_id",
		Method:   "GET",
		Response: "TLoanPurpose",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan purpose ID")
		}
		purpose, err := c.model.LoanPurposeManager.GetByIDRaw(context, *id)
		if err != nil {
			return c.NotFound(ctx, "LoanPurpose")
		}
		return ctx.JSON(http.StatusOK, purpose)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose",
		Method:   "POST",
		Request:  "TLoanPurpose",
		Response: "TLoanPurpose",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.LoanPurposeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		purpose := &model.LoanPurpose{
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.LoanPurposeManager.Create(context, purpose); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.ToModel(purpose))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-purpose/:loan_purpose_id",
		Method:   "PUT",
		Request:  "TLoanPurpose",
		Response: "TLoanPurpose",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan purpose ID")
		}

		req, err := c.model.LoanPurposeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		purpose, err := c.model.LoanPurposeManager.GetByID(context, *id)
		if err != nil {
			return c.NotFound(ctx, "LoanPurpose")
		}
		purpose.Description = req.Description
		purpose.Icon = req.Icon
		purpose.UpdatedAt = time.Now().UTC()
		purpose.UpdatedByID = user.UserID
		if err := c.model.LoanPurposeManager.UpdateFields(context, purpose.ID, purpose); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.LoanPurposeManager.ToModel(purpose))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/loan-purpose/:loan_purpose_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_purpose_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan purpose ID")
		}
		if err := c.model.LoanPurposeManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/loan-purpose/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple loan purpose records",
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
			if _, err := c.model.LoanPurposeManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("LoanPurpose with ID %s", rawID))
			}
			if err := c.model.LoanPurposeManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
