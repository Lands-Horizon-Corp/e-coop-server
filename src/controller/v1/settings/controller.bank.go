package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func BankController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bank",
		Method:       "GET",
		Note:         "Returns all banks for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := core.BankCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No banks found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.BankManager(service).ToModels(banks))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bank/search",
		Method:       "GET",
		Note:         "Returns a paginated list of banks for the current user's organization and branch.",
		ResponseType: types.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		banks, err := core.BankManager(service).NormalPagination(context, ctx, &types.Bank{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch banks for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, banks)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "GET",
		Note:         "Returns a single bank by its ID.",
		ResponseType: types.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := helpers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := core.BankManager(service).GetByIDRaw(context, *bankID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		return ctx.JSON(http.StatusOK, bank)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bank",
		Method:       "POST",
		Note:         "Creates a new bank for the current user's organization and branch.",
		RequestType:  types.BankRequest{},
		ResponseType: types.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.BankManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), user not assigned to branch.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		bank := &types.Bank{
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

		if err := core.BankManager(service).Create(context, bank); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bank creation failed (/bank), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bank: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created bank (/bank): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusCreated, core.BankManager(service).ToModel(bank))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bank/:bank_id",
		Method:       "PUT",
		Note:         "Updates an existing bank by its ID.",
		RequestType:  types.BankRequest{},
		ResponseType: types.BankResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := helpers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}

		req, err := core.BankManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), validation error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		bank, err := core.BankManager(service).GetByID(context, *bankID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.BankManager(service).UpdateByID(context, bank.ID, bank); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update bank: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.JSON(http.StatusOK, core.BankManager(service).ToModel(bank))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/bank/:bank_id",
		Method: "DELETE",
		Note:   "Deletes the specified bank by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		bankID, err := helpers.EngineUUIDParam(ctx, "bank_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), invalid bank ID.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bank ID"})
		}
		bank, err := core.BankManager(service).GetByID(context, *bankID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), not found.",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bank not found"})
		}
		if err := core.BankManager(service).Delete(context, *bankID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bank delete failed (/bank/:bank_id), db error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bank: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted bank (/bank/:bank_id): " + bank.Name,
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/bank/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple banks by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | no IDs provided",
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No bank IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}

		if err := core.BankManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete banks (/bank/bulk-delete) | error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete banks: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted banks (/bank/bulk-delete)",
			Module:      "Bank",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
