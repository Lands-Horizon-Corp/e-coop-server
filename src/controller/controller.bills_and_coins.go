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

func (c *Controller) BillAndCoinsController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/bills-and-coins",
		Method:   "GET",
		Response: "TBillAndCoins[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		BillAndCoins, err := c.model.BillAndCoinsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "Bills and Coins")
		}
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.ToModels(BillAndCoins))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bills-and-coins/search",
		Method:   "GET",
		Request:  "Filter<IBillAndCoins>",
		Response: "Paginated<IBillAndCoins>",
		Note:     "Get pagination bills and coins",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.BillAndCoinsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bills-and-coins/:bills_and_coins_id",
		Method:   "GET",
		Response: "TBillAndCoins",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		BillAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bills and coins ID")
		}
		BillAndCoins, err := c.model.BillAndCoinsManager.GetByIDRaw(context, *BillAndCoinsID)
		if err != nil {
			return c.NotFound(ctx, "Bills and Coins")
		}
		return ctx.JSON(http.StatusOK, BillAndCoins)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bills-and-coins",
		Method:   "POST",
		Request:  "TBillAndCoins",
		Response: "TBillAndCoins",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.BillAndCoinsManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		BillAndCoins := &model.BillAndCoins{
			MediaID:     req.MediaID,
			Name:        req.Name,
			Value:       req.Value,
			CountryCode: req.CountryCode,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.BillAndCoinsManager.Create(context, BillAndCoins); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.ToModel(BillAndCoins))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bills-and-coins/:bills_and_coins_id",
		Method:   "PUT",
		Request:  "TBillAndCoins",
		Response: "TBillAndCoins",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		BillAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bills and coins ID")
		}

		req, err := c.model.BillAndCoinsManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		BillAndCoins, err := c.model.BillAndCoinsManager.GetByID(context, *BillAndCoinsID)
		if err != nil {
			return c.NotFound(ctx, "Bills and Coins")
		}
		BillAndCoins.MediaID = req.MediaID
		BillAndCoins.Name = req.Name
		BillAndCoins.Value = req.Value
		BillAndCoins.CountryCode = req.CountryCode

		BillAndCoins.UpdatedAt = time.Now().UTC()
		BillAndCoins.UpdatedByID = user.UserID
		if err := c.model.BillAndCoinsManager.UpdateFields(context, BillAndCoins.ID, BillAndCoins); err != nil {
			return ctx.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.ToModel(BillAndCoins))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/bills-and-coins/:bills_and_coins_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		BillAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bills and coins ID")
		}
		if err := c.model.BillAndCoinsManager.DeleteByID(context, *BillAndCoinsID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/bills-and-coins/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple bills and coins records",
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
			BillAndCoinsID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.BillAndCoinsManager.GetByID(context, BillAndCoinsID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Bills and Coins with ID %s", rawID))
			}
			if err := c.model.BillAndCoinsManager.DeleteByIDWithTx(context, tx, BillAndCoinsID); err != nil {
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
