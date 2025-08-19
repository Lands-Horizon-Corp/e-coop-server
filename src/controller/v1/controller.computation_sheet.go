package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ComputationSheetController registers routes for managing computation sheets.
func (c *Controller) ComputationSheetController() {
	req := c.provider.Service.Request

	// POST /computation-sheet/:computation_sheet_id/calculator: Returns sample calculation data.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/computation-sheet/:computation_sheet_id/calculator",
		Method: "POST",
		Note:   "Returns sample payment calculation data for a computation sheet.",
	}, func(ctx echo.Context) error {
		// You can parse computation_sheet_id if needed, but for sample data, we ignore it.
		// id, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		// if err != nil {
		//     return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		// }

		// Generate sample data
		now := time.Now().UTC()
		sample := map[string]any{
			"payments": []map[string]any{
				{
					"date":   now.Format("2006-01-02"),
					"amount": 100,
				},
				{
					"date":   now.AddDate(0, 0, 30).Format("2006-01-02"),
					"amount": 100,
				},
			},
			"total_amount":   200,
			"total_interest": 40,
		}
		return ctx.JSON(http.StatusOK, sample)
	})

	// GET /computation-sheet: List all computation sheets for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "GET",
		Note:         "Returns all computation sheets for the current user's organization and branch.",
		ResponseType: model.ComputationSheetResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheets, err := c.model.ComputationSheetCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No computation sheets found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.ComputationSheetManager.Filtered(context, ctx, sheets))
	})

	// GET /computation-sheet/:id: Get specific computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "GET",
		ResponseType: model.ComputationSheetResponse{},
		Note:         "Returns a single computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.model.ComputationSheetManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		return ctx.JSON(http.StatusOK, sheet)
	})

	// POST /computation-sheet: Create a new computation sheet.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "POST",
		RequestType:  model.ComputationSheetRequest{},
		ResponseType: model.ComputationSheetResponse{},
		Note:         "Creates a new computation sheet for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user not assigned to branch.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		sheet := &model.ComputationSheet{
			Name:              req.Name,
			Description:       req.Description,
			DeliquentAccount:  req.DeliquentAccount,
			FinesAccount:      req.FinesAccount,
			InterestAccountID: req.InterestAccountID,
			ComakerAccount:    req.ComakerAccount,
			ExistAccount:      req.ExistAccount,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.model.ComputationSheetManager.Create(context, sheet); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create computation sheet: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created computation sheet (/computation-sheet): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusCreated, c.model.ComputationSheetManager.ToModel(sheet))
	})

	// PUT /computation-sheet/:id: Update computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "PUT",
		RequestType:  model.ComputationSheetRequest{},
		ResponseType: model.ComputationSheetResponse{},
		Note:         "Updates an existing computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}

		req, err := c.model.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		sheet, err := c.model.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		sheet.Name = req.Name
		sheet.Description = req.Description
		sheet.DeliquentAccount = req.DeliquentAccount
		sheet.FinesAccount = req.FinesAccount
		sheet.InterestAccountID = req.InterestAccountID
		sheet.ComakerAccount = req.ComakerAccount
		sheet.ExistAccount = req.ExistAccount
		sheet.UpdatedAt = time.Now().UTC()
		sheet.UpdatedByID = user.UserID

		if err := c.model.ComputationSheetManager.UpdateFields(context, sheet.ID, sheet); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update computation sheet: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusOK, c.model.ComputationSheetManager.ToModel(sheet))
	})

	// DELETE /computation-sheet/:id: Delete a computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/computation-sheet/:id",
		Method: "DELETE",
		Note:   "Deletes the specified computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.model.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		if err := c.model.ComputationSheetManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete computation sheet: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /computation-sheet/bulk-delete: Bulk delete computation sheets by IDs.
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/computation-sheet/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple computation sheets by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete), invalid request body.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete), no IDs provided.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No computation sheet IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/computation-sheet/bulk-delete), invalid UUID: " + rawID,
					Module:      "ComputationSheet",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			sheet, err := c.model.ComputationSheetManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/computation-sheet/bulk-delete), not found: " + rawID,
					Module:      "ComputationSheet",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Computation sheet not found with ID: %s", rawID)})
			}
			names += sheet.Name + ","
			if err := c.model.ComputationSheetManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/computation-sheet/bulk-delete), db error: " + err.Error(),
					Module:      "ComputationSheet",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete computation sheet: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete), commit error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted computation sheets (/computation-sheet/bulk-delete): " + names,
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
