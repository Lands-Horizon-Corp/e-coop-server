package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// ComputationSheetController registers routes for managing computation sheets.
func (c *Controller) computationSheetController() {
	req := c.provider.Service.Request

	// POST /computation-sheet/:computation_sheet_id/calculator: Returns sample calculation data.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:computation_sheet_id/calculator",
		Method:       "POST",
		Note:         "Returns sample payment calculation data for a computation sheet.",
		RequestType:  event.LoanComputationSheetCalculatorRequest{},
		ResponseType: event.ComputationSheetAmortizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var request event.LoanComputationSheetCalculatorRequest
		computationSheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		if computationSheetID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Computation sheet ID is required"})
		}
		if err := ctx.Bind(&request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		computed, err := c.event.ComputationSheetCalculator(
			context,
			event.LoanComputationSheetCalculatorRequest{
				AccountID:           request.AccountID,
				Applied1:            request.Applied1,
				Terms:               request.Terms,
				MemberTypeID:        request.MemberTypeID,
				IsAddOn:             request.IsAddOn,
				CashOnHandAccountID: cashOnHandAccountID,
				ComputationSheetID:  computationSheetID,
			},
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Computation failed: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, computed)
	})

	// GET /computation-sheet: List all computation sheets for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "GET",
		Note:         "Returns all computation sheets for the current user's organization and branch.",
		ResponseType: core.ComputationSheetResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheets, err := c.core.ComputationSheetCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No computation sheets found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.ComputationSheetManager.ToModels(sheets))
	})

	// GET /computation-sheet/:id: Get specific computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "GET",
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Returns a single computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		return ctx.JSON(http.StatusOK, sheet)
	})

	// POST /computation-sheet: Create a new computation sheet.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "POST",
		RequestType:  core.ComputationSheetRequest{},
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Creates a new computation sheet for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user not assigned to branch.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		sheet := &core.ComputationSheet{
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
			CurrencyID:        req.CurrencyID,
		}

		if err := c.core.ComputationSheetManager.Create(context, sheet); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created computation sheet (/computation-sheet): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusCreated, c.core.ComputationSheetManager.ToModel(sheet))
	})

	// PUT /computation-sheet/:id: Update computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "PUT",
		RequestType:  core.ComputationSheetRequest{},
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Updates an existing computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}

		req, err := c.core.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.ComputationSheetManager.UpdateByID(context, sheet.ID, sheet); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusOK, c.core.ComputationSheetManager.ToModel(sheet))
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
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		if err := c.core.ComputationSheetManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/computation-sheet/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple computation sheets by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | no IDs provided",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No computation sheet IDs provided for bulk delete"})
		}

		if err := c.core.ComputationSheetManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete computation sheets: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted computation sheets (/computation-sheet/bulk-delete)",
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// POST - api/v1/computation-sheeet/:computation-sheet-id/account/:account-id/connect
	// PUT - api/v1/computation-sheeet/:computation-sheet-id/account/:account-id/disconnect
}
