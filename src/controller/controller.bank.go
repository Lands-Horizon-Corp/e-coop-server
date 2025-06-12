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

func (c *Controller) BankController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/bank",
		Method:   "GET",
		Response: "TBank[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		bank, err := c.model.BankCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "Bank")
		}

		return ctx.JSON(http.StatusOK, c.model.BankManager.ToModels(bank))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bank/search",
		Method:   "GET",
		Request:  "Filter<IBank>",
		Response: "Paginated<IBank>",
		Note:     "Get pagination bank",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.BankCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BankManager.Pagination(context, ctx, value))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/bank/:bank_id",
		Method:   "GET",
		Response: "TBank",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := horizon.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bank ID")
		}
		bank, err := c.model.BankManager.GetByIDRaw(context, *bankID)
		if err != nil {
			return c.NotFound(ctx, "Bank")
		}
		return ctx.JSON(http.StatusOK, bank)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bank",
		Method:   "POST",
		Request:  "TBank",
		Response: "TBank",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.BankManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		bank := &model.Bank{
			MediaID:     req.MediaID,
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.BankManager.Create(context, bank); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.BankManager.ToModel(bank))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/bank/:bank_id",
		Method:   "PUT",
		Request:  "TBank",
		Response: "TBank",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := horizon.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bank ID")
		}

		req, err := c.model.BankManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		bank, err := c.model.BankManager.GetByID(context, *bankID)
		if err != nil {
			return c.NotFound(ctx, "Bank")
		}
		bank.MediaID = req.MediaID
		bank.Name = req.Name
		bank.Description = req.Description
		bank.UpdatedAt = time.Now().UTC()
		bank.UpdatedByID = user.UserID
		if err := c.model.BankManager.UpdateFields(context, bank.ID, bank); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.BankManager.ToModel(bank))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/bank/:bank_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := horizon.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid bank ID")
		}
		if err := c.model.BankManager.DeleteByID(context, *bankID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/bank/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple bank records",
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
			bankID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.BankManager.GetByID(context, bankID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Bank with ID %s", rawID))
			}
			if err := c.model.BankManager.DeleteByIDWithTx(context, tx, bankID); err != nil {
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
