package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// BillAndCoinsController handles endpoints for managing bills and coins.
func (c *Controller) BillAndCoinsController() {
	req := c.provider.Service.Request

	// GET /bills-and-coins: List all bills and coins for the current user's branch. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/bills-and-coins",
		Method:       "GET",
		Note:         "Returns all bills and coins for the current user's organization and branch. Returns error if not authenticated.",
		ResponseType: model.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		billAndCoins, err := c.model.BillAndCoinsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No bills and coins found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.Filtered(context, ctx, billAndCoins))
	})

	// GET /bills-and-coins/search: Paginated search of bills and coins for current branch. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/bills-and-coins/search",
		Method:       "GET",
		Note:         "Returns a paginated list of bills and coins for the current user's organization and branch.",
		ResponseType: model.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		billAndCoins, err := c.model.BillAndCoinsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bills and coins: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.Pagination(context, ctx, billAndCoins))
	})

	// GET /bills-and-coins/:bills_and_coins_id: Get a specific bills and coins record by ID. (NO footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/bills-and-coins/:bills_and_coins_id",
		Method:       "GET",
		Note:         "Returns a bills and coins record by its ID.",
		ResponseType: model.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}
		billAndCoins, err := c.model.BillAndCoinsManager.GetByIDRaw(context, *billAndCoinsID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		return ctx.JSON(http.StatusOK, billAndCoins)
	})

	// POST /bills-and-coins: Create a new bills and coins record. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/bills-and-coins",
		Method:       "POST",
		RequestType:  model.BillAndCoinsRequest{},
		ResponseType: model.BillAndCoinsResponse{},
		Note:         "Creates a new bills and coins record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.BillAndCoinsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), validation error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), user org error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), user not assigned to branch.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		billAndCoins := &model.BillAndCoins{
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

		if err := c.model.BillAndCoinsManager.Create(context, billAndCoins); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bills and coins record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created bills and coins (/bills-and-coins): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.JSON(http.StatusCreated, c.model.BillAndCoinsManager.ToModel(billAndCoins))
	})

	// PUT /bills-and-coins/:bills_and_coins_id: Update a bills and coins record by ID. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:        "/bills-and-coins/:bills_and_coins_id",
		Method:       "PUT",
		RequestType:  model.BillAndCoinsRequest{},
		ResponseType: model.BillAndCoinsResponse{},
		Note:         "Updates an existing bills and coins record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), invalid ID.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}

		req, err := c.model.BillAndCoinsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), validation error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), user org error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		billAndCoins, err := c.model.BillAndCoinsManager.GetByID(context, *billAndCoinsID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), record not found.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		billAndCoins.MediaID = req.MediaID
		billAndCoins.Name = req.Name
		billAndCoins.Value = req.Value
		billAndCoins.CountryCode = req.CountryCode

		billAndCoins.UpdatedAt = time.Now().UTC()
		billAndCoins.UpdatedByID = user.UserID
		if err := c.model.BillAndCoinsManager.UpdateFields(context, billAndCoins.ID, billAndCoins); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Failed to update bills and coins record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated bills and coins (/bills-and-coins/:bills_and_coins_id): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.JSON(http.StatusOK, c.model.BillAndCoinsManager.ToModel(billAndCoins))
	})

	// DELETE /bills-and-coins/:bills_and_coins_id: Delete a bills and coins record by ID. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:  "/bills-and-coins/:bills_and_coins_id",
		Method: "DELETE",
		Note:   "Deletes the specified bills and coins record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := horizon.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), invalid ID.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}
		billAndCoins, err := c.model.BillAndCoinsManager.GetByID(context, *billAndCoinsID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), record not found.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		if err := c.model.BillAndCoinsManager.DeleteByID(context, *billAndCoinsID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bills and coins record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted bills and coins (/bills-and-coins/:bills_and_coins_id): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /bills-and-coins/bulk-delete: Bulk delete bills and coins records by IDs. (WITH footstep)
	req.RegisterRoute(horizon.Route{
		Route:       "/bills-and-coins/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple bills and coins records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), invalid request body.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), no IDs provided.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			billAndCoinsID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), invalid UUID: " + rawID,
					Module:      "BillAndCoins",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			billAndCoins, err := c.model.BillAndCoinsManager.GetByID(context, billAndCoinsID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), record not found: " + rawID,
					Module:      "BillAndCoins",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Bills and coins record not found with ID: %s", rawID)})
			}
			names += billAndCoins.Name + ","
			if err := c.model.BillAndCoinsManager.DeleteByIDWithTx(context, tx, billAndCoinsID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), db error: " + err.Error(),
					Module:      "BillAndCoins",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bills and coins record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bills and coins bulk delete failed (/bills-and-coins/bulk-delete), commit error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted bills and coins (/bills-and-coins/bulk-delete): " + names,
			Module:      "BillAndCoins",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
