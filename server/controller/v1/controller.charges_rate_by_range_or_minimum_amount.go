package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// ChargesRateByRangeOrMinimumAmountController registers routes for managing charges rate by range or minimum amount.
func (c *Controller) chargesRateByRangeOrMinimumAmountController() {
	req := c.provider.Service.WebRequest

	// POST /charges-rate-by-range-or-minimum-amount: Create a new charges rate by range or minimum amount. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-by-range-or-minimum-amount/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "POST",
		Note:         "Creates a new charges rate by range or minimum amount for the current user's organization and branch.",
		RequestType:  core.ChargesRateByRangeOrMinimumAmountRequest{},
		ResponseType: core.ChargesRateByRangeOrMinimumAmountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by range or minimum amount creation failed (/charges-rate-by-range-or-minimum-amount), invalid charges rate scheme ID.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		req, err := c.core.ChargesRateByRangeOrMinimumAmountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by range or minimum amount creation failed (/charges-rate-by-range-or-minimum-amount), validation error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by range or minimum amount data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by range or minimum amount creation failed (/charges-rate-by-range-or-minimum-amount), user org error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by range or minimum amount creation failed (/charges-rate-by-range-or-minimum-amount), user not assigned to branch.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateByRangeOrMinimumAmount := &core.ChargesRateByRangeOrMinimumAmount{
			ChargesRateSchemeID: *chargesRateSchemeID,
			From:                req.From,
			To:                  req.To,
			Charge:              req.Charge,
			Amount:              req.Amount,
			MinimumAmount:       req.MinimumAmount,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         userOrg.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         userOrg.UserID,
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
		}

		if err := c.core.ChargesRateByRangeOrMinimumAmountManager.Create(context, chargesRateByRangeOrMinimumAmount); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate by range or minimum amount creation failed (/charges-rate-by-range-or-minimum-amount), db error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate by range or minimum amount: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate by range or minimum amount (/charges-rate-by-range-or-minimum-amount): " + chargesRateByRangeOrMinimumAmount.ID.String(),
			Module:      "ChargesRateByRangeOrMinimumAmount",
		})
		return ctx.JSON(http.StatusCreated, c.core.ChargesRateByRangeOrMinimumAmountManager.ToModel(chargesRateByRangeOrMinimumAmount))
	})

	// PUT /charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id: Update charges rate by range or minimum amount by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate by range or minimum amount by its ID.",
		RequestType:  core.ChargesRateByRangeOrMinimumAmountRequest{},
		ResponseType: core.ChargesRateByRangeOrMinimumAmountResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateByRangeOrMinimumAmountID, err := handlers.EngineUUIDParam(ctx, "charges_rate_by_range_or_minimum_amount_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by range or minimum amount update failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), invalid charges rate by range or minimum amount ID.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by range or minimum amount ID"})
		}

		req, err := c.core.ChargesRateByRangeOrMinimumAmountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by range or minimum amount update failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), validation error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by range or minimum amount data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by range or minimum amount update failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), user org error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateByRangeOrMinimumAmount, err := c.core.ChargesRateByRangeOrMinimumAmountManager.GetByID(context, *chargesRateByRangeOrMinimumAmountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by range or minimum amount update failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), charges rate by range or minimum amount not found.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate by range or minimum amount not found"})
		}
		chargesRateByRangeOrMinimumAmount.From = req.From
		chargesRateByRangeOrMinimumAmount.To = req.To
		chargesRateByRangeOrMinimumAmount.Charge = req.Charge
		chargesRateByRangeOrMinimumAmount.Amount = req.Amount
		chargesRateByRangeOrMinimumAmount.MinimumAmount = req.MinimumAmount
		chargesRateByRangeOrMinimumAmount.UpdatedAt = time.Now().UTC()
		chargesRateByRangeOrMinimumAmount.UpdatedByID = userOrg.UserID
		if err := c.core.ChargesRateByRangeOrMinimumAmountManager.UpdateByID(context, chargesRateByRangeOrMinimumAmount.ID, chargesRateByRangeOrMinimumAmount); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate by range or minimum amount update failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), db error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate by range or minimum amount: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate by range or minimum amount (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id): " + chargesRateByRangeOrMinimumAmount.ID.String(),
			Module:      "ChargesRateByRangeOrMinimumAmount",
		})
		return ctx.JSON(http.StatusOK, c.core.ChargesRateByRangeOrMinimumAmountManager.ToModel(chargesRateByRangeOrMinimumAmount))
	})

	// DELETE /charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id: Delete a charges rate by range or minimum amount by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate by range or minimum amount by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateByRangeOrMinimumAmountID, err := handlers.EngineUUIDParam(ctx, "charges_rate_by_range_or_minimum_amount_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by range or minimum amount delete failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), invalid charges rate by range or minimum amount ID.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate by range or minimum amount ID"})
		}
		chargesRateByRangeOrMinimumAmount, err := c.core.ChargesRateByRangeOrMinimumAmountManager.GetByID(context, *chargesRateByRangeOrMinimumAmountID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by range or minimum amount delete failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), not found.",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate by range or minimum amount not found"})
		}
		if err := c.core.ChargesRateByRangeOrMinimumAmountManager.Delete(context, *chargesRateByRangeOrMinimumAmountID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate by range or minimum amount delete failed (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id), db error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate by range or minimum amount: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate by range or minimum amount (/charges-rate-by-range-or-minimum-amount/:charges_rate_by_range_or_minimum_amount_id): " + chargesRateByRangeOrMinimumAmount.ID.String(),
			Module:      "ChargesRateByRangeOrMinimumAmount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/charges-rate-by-range-or-minimum-amount/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate by range or minimum amount by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-range-or-minimum-amount/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-range-or-minimum-amount/bulk-delete) | no IDs provided",
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate by range or minimum amount IDs provided for bulk delete"})
		}

		if err := c.core.ChargesRateByRangeOrMinimumAmountManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-by-range-or-minimum-amount/bulk-delete) | error: " + err.Error(),
				Module:      "ChargesRateByRangeOrMinimumAmount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete charges rate by range or minimum amount: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate by range or minimum amount (/charges-rate-by-range-or-minimum-amount/bulk-delete)",
			Module:      "ChargesRateByRangeOrMinimumAmount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
