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

func (c *Controller) CollateralController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/collateral",
		Method:   "GET",
		Response: "TCollateral[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		collaterals, err := c.model.CollateralCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "Collateral")
		}
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.ToModels(collaterals))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/collateral/search",
		Method:   "GET",
		Request:  "Filter<ICollateral>",
		Response: "Paginated<ICollateral>",
		Note:     "Get pagination collateral",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.CollateralCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/collateral/:collateral_id",
		Method:   "GET",
		Response: "TCollateral",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := horizon.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid collateral ID")
		}
		collateral, err := c.model.CollateralManager.GetByIDRaw(context, *collateralID)
		if err != nil {
			return c.NotFound(ctx, "Collateral")
		}
		return ctx.JSON(http.StatusOK, collateral)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/collateral",
		Method:   "POST",
		Request:  "TCollateral",
		Response: "TCollateral",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.CollateralManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		collateral := &model.Collateral{
			Icon:           req.Icon,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.CollateralManager.Create(context, collateral); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.CollateralManager.ToModel(collateral))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/collateral/:collateral_id",
		Method:   "PUT",
		Request:  "TCollateral",
		Response: "TCollateral",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := horizon.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid collateral ID")
		}

		req, err := c.model.CollateralManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		collateral, err := c.model.CollateralManager.GetByID(context, *collateralID)
		if err != nil {
			return c.NotFound(ctx, "Collateral")
		}
		collateral.Icon = req.Icon
		collateral.Name = req.Name
		collateral.Description = req.Description
		collateral.UpdatedAt = time.Now().UTC()
		collateral.UpdatedByID = user.UserID
		if err := c.model.CollateralManager.UpdateFields(context, collateral.ID, collateral); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.ToModel(collateral))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/collateral/:collateral_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := horizon.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid collateral ID")
		}
		if err := c.model.CollateralManager.DeleteByID(context, *collateralID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/collateral/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple collateral records",
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
			collateralID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.CollateralManager.GetByID(context, collateralID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Collateral with ID %s", rawID))
			}
			if err := c.model.CollateralManager.DeleteByIDWithTx(context, tx, collateralID); err != nil {
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
