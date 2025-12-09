package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// BankController registers routes for managing banks.
func (c *Controller) bankController() {
	req := c.provider.Service.Request

	// GET /bank: List all banks for the current user's branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/bank",
		Method:       "GET",
		Note:         "Returns all banks for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := c.core.BankCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No banks found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.BankManager.ToModels(banks))
	})

	// GET /bank/search: Paginated search of banks for the current branch. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/bank/search",
		Method:       "GET",
		Note:         "Returns a paginated list of banks for the current user's organization and branch.",
		ResponseType: core.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := c.core.BankManager.NormalPagination(context, ctx, &core.Bank{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch banks for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, banks)
	})

	// GET /bank/:bank_id: Get specific bank by ID. (NO footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "GET",
		Note:         "Returns a single bank by its ID.",
		ResponseType: core.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := c.core.BankManager.GetByIDRaw(context, *bankID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		return ctx.JSON(http.StatusOK, bank)
	})

	// POST /bank: Create a new bank. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/bank",
		Method:       "POST",
		Note:         "Creates a new bank for the current user's organization and branch.",
		RequestType:  core.BankRequest{},
		ResponseType: core.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.BankManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user not assigned to branch.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		bank := &core.Bank{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.BankManager.Create(context, bank); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bank: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created bank (/bank): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusCreated, c.core.BankManager.ToModel(bank))
	})

	// PUT /bank/:bank_id: Update bank by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "PUT",
		Note:         "Updates an existing bank by its ID.",
		RequestType:  core.BankRequest{},
		ResponseType: core.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}

		req, err := c.core.BankManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		bank, err := c.core.BankManager.GetByID(context, *bankID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), bank not found.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		bank.MediaID = req.MediaID
		bank.Name = req.Name
		bank.Description = req.Description
		bank.UpdatedAt = time.Now().UTC()
		bank.UpdatedByID = userOrg.UserID
		if err := c.core.BankManager.UpdateByID(context, bank.ID, bank); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update bank: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusOK, c.core.BankManager.ToModel(bank))
	})

	// DELETE /bank/:bank_id: Delete a bank by ID. (WITH footstep)
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/bank/:bank_id",
		Method: "DELETE",
		Note:   "Deletes the specified bank by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := c.core.BankManager.GetByID(context, *bankID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), not found.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		if err := c.core.BankManager.Delete(context, *bankID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bank: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/bank/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple banks by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | no IDs provided",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No bank IDs provided for bulk delete"})
		}

		if err := c.core.BankManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete banks: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted banks (/bank/bulk-delete)",
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
