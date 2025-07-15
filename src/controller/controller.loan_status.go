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

func (c *Controller) LoanStatusController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status",
		Method:   "GET",
		Response: "TLoanStatus[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		statuses, err := c.model.LoanStatusCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "LoanStatus")
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModels(statuses))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/search",
		Method:   "GET",
		Request:  "Filter<ILoanStatus>",
		Response: "Paginated<ILoanStatus>",
		Note:     "Get pagination loan status",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.LoanStatusCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/:loan_status_id",
		Method:   "GET",
		Response: "TLoanStatus",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan status ID")
		}
		status, err := c.model.LoanStatusManager.GetByIDRaw(context, *id)
		if err != nil {
			return c.NotFound(ctx, "LoanStatus")
		}
		return ctx.JSON(http.StatusOK, status)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status",
		Method:   "POST",
		Request:  "TLoanStatus",
		Response: "TLoanStatus",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		status := &model.LoanStatus{
			Name:           req.Name,
			Icon:           req.Icon,
			Color:          req.Color,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.LoanStatusManager.Create(context, status); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModel(status))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/loan-status/:loan_status_id",
		Method:   "PUT",
		Request:  "TLoanStatus",
		Response: "TLoanStatus",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan status ID")
		}

		req, err := c.model.LoanStatusManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		status, err := c.model.LoanStatusManager.GetByID(context, *id)
		if err != nil {
			return c.NotFound(ctx, "LoanStatus")
		}
		status.Name = req.Name
		status.Icon = req.Icon
		status.Color = req.Color
		status.Description = req.Description
		status.UpdatedAt = time.Now().UTC()
		status.UpdatedByID = user.UserID
		if err := c.model.LoanStatusManager.UpdateFields(context, status.ID, status); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.LoanStatusManager.ToModel(status))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/loan-status/:loan_status_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "loan_status_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid loan status ID")
		}
		if err := c.model.LoanStatusManager.DeleteByID(context, *id); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/loan-status/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple loan status records",
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
			if _, err := c.model.LoanStatusManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("LoanStatus with ID %s", rawID))
			}
			if err := c.model.LoanStatusManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
