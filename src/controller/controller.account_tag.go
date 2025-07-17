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

func (c *Controller) AccountTagController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag",
		Method:   "GET",
		Response: "TAccountTag[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		accountTag, err := c.model.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "AccountTag")
		}

		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModels(accountTag))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/search",
		Method:   "GET",
		Request:  "Filter<IAccountTag>",
		Response: "Paginated<IAccountTag>",
		Note:     "Get pagination account tag",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.AccountTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/:account_tag_id",
		Method:   "GET",
		Response: "TAccountTag",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account tag ID")
		}
		accountTag, err := c.model.AccountTagManager.GetByIDRaw(context, *accountTagID)
		if err != nil {
			return c.NotFound(ctx, "AccountTag")
		}
		return ctx.JSON(http.StatusOK, accountTag)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag",
		Method:   "POST",
		Request:  "TAccountTag",
		Response: "TAccountTag",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AccountTagManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		accountTag := &model.AccountTag{
			AccountID:   req.AccountID,
			Name:        req.Name,
			Description: req.Description,
			Category:    req.Category,
			Color:       req.Color,
			Icon:        req.Icon,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.AccountTagManager.Create(context, accountTag); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModel(accountTag))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/account-tag/:account_tag_id",
		Method:   "PUT",
		Request:  "TAccountTag",
		Response: "TAccountTag",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account tag ID")
		}

		req, err := c.model.AccountTagManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		accountTag, err := c.model.AccountTagManager.GetByID(context, *accountTagID)
		if err != nil {
			return c.NotFound(ctx, "AccountTag")
		}
		accountTag.AccountID = req.AccountID
		accountTag.Name = req.Name
		accountTag.Description = req.Description
		accountTag.Category = req.Category
		accountTag.Color = req.Color
		accountTag.Icon = req.Icon
		accountTag.UpdatedAt = time.Now().UTC()
		accountTag.UpdatedByID = user.UserID

		if err := c.model.AccountTagManager.UpdateFields(context, accountTag.ID, accountTag); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.AccountTagManager.ToModel(accountTag))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/account-tag/:account_tag_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountTagID, err := horizon.EngineUUIDParam(ctx, "account_tag_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid account tag ID")
		}
		if err := c.model.AccountTagManager.DeleteByID(context, *accountTagID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/account-tag/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple account tag records",
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
			accountTagID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.AccountTagManager.GetByID(context, accountTagID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("AccountTag with ID %s", rawID))
			}
			if err := c.model.AccountTagManager.DeleteByIDWithTx(context, tx, accountTagID); err != nil {
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
