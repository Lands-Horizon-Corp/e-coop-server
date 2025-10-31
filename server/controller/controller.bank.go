package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BankController registers routes for managing banks.
func (c *Controller) bankController() {
	req := c.provider.Service.Request

	// GET /bank: List all banks for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/bank",
		Method:       "GET",
		Note:         "Returns all banks for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: modelcore.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := c.modelcore.BankCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No banks found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.BankManager.Filtered(context, ctx, banks))
	})

	// GET /bank/search: Paginated search of banks for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/bank/search",
		Method:       "GET",
		Note:         "Returns a paginated list of banks for the current user's organization and branch.",
		ResponseType: modelcore.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := c.modelcore.BankCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch banks for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.BankManager.Pagination(context, ctx, banks))
	})

	// GET /bank/:bank_id: Get specific bank by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "GET",
		Note:         "Returns a single bank by its ID.",
		ResponseType: modelcore.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := c.modelcore.BankManager.GetByIDRaw(context, *bankID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		return ctx.JSON(http.StatusOK, bank)
	})

	// POST /bank: Create a new bank. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/bank",
		Method:       "POST",
		Note:         "Creates a new bank for the current user's organization and branch.",
		RequestType:  modelcore.BankRequest{},
		ResponseType: modelcore.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.BankManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user not assigned to branch.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		bank := &modelcore.Bank{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.BankManager.Create(context, bank); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bank: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created bank (/bank): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusCreated, c.modelcore.BankManager.ToModel(bank))
	})

	// PUT /bank/:bank_id: Update bank by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "PUT",
		Note:         "Updates an existing bank by its ID.",
		RequestType:  modelcore.BankRequest{},
		ResponseType: modelcore.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}

		req, err := c.modelcore.BankManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		bank, err := c.modelcore.BankManager.GetByID(context, *bankID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		bank.UpdatedByID = user.UserID
		if err := c.modelcore.BankManager.UpdateFields(context, bank.ID, bank); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update bank: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.BankManager.ToModel(bank))
	})

	// DELETE /bank/:bank_id: Delete a bank by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/bank/:bank_id",
		Method: "DELETE",
		Note:   "Deletes the specified bank by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := handlers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := c.modelcore.BankManager.GetByID(context, *bankID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), not found.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		if err := c.modelcore.BankManager.DeleteByID(context, *bankID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bank: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /bank/bulk-delete: Bulk delete banks by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/bank/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple banks by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/bank/bulk-delete), invalid request body.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/bank/bulk-delete), no IDs provided.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No bank IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/bank/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			bankID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/bank/bulk-delete), invalid UUID: " + rawID,
					Module:      "Bank",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			bank, err := c.modelcore.BankManager.GetByID(context, bankID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/bank/bulk-delete), not found: " + rawID,
					Module:      "Bank",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Bank not found with ID: %s", rawID)})
			}
			names += bank.Name + ","
			if err := c.modelcore.BankManager.DeleteByIDWithTx(context, tx, bankID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/bank/bulk-delete), db error: " + err.Error(),
					Module:      "Bank",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bank: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/bank/bulk-delete), commit error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted banks (/bank/bulk-delete): " + names,
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
