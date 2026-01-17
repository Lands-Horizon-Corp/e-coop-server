package voucher

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func BillAndCoinsController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bills-and-coins",
		Method:       "GET",
		Note:         "Returns all bills and coins for the current user's organization and branch. Returns error if not authenticated.",
		ResponseType: core.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: user org error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionBatch, err := core.TransactionBatchCurrent(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create transaction batch failed: current batch error: " + err.Error(),
				Module:      "TransactionBatch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current transaction batch: " + err.Error()})
		}
		billAndCoins, err := core.BillAndCoinsManager(service).FindRaw(context, &types.BillAndCoins{
			CurrencyID:     transactionBatch.CurrencyID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bills and coins: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, billAndCoins)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bills-and-coins/search",
		Method:       "GET",
		Note:         "Returns a paginated list of bills and coins for the current user's organization and branch.",
		ResponseType: core.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		billAndCoins, err := core.BillAndCoinsManager(service).NormalPagination(context, ctx, &types.BillAndCoins{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bills and coins: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, billAndCoins)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bills-and-coins/:bills_and_coins_id",
		Method:       "GET",
		Note:         "Returns a bills and coins record by its ID.",
		ResponseType: core.BillAndCoinsResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := helpers.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}
		billAndCoins, err := core.BillAndCoinsManager(service).GetByIDRaw(context, *types.BillAndCoinsID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		return ctx.JSON(http.StatusOK, billAndCoins)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bills-and-coins",
		Method:       "POST",
		RequestType:  core.BillAndCoinsRequest{},
		ResponseType: core.BillAndCoinsResponse{},
		Note:         "Creates a new bills and coins record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.BillAndCoinsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), validation error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), user org error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), user not assigned to branch.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		billAndCoins := &types.BillAndCoins{
			MediaID:    req.MediaID,
			Name:       req.Name,
			Value:      req.Value,
			CurrencyID: req.CurrencyID,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.BillAndCoinsManager(service).Create(context, billAndCoins); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Bills and coins creation failed (/bills-and-coins), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bills and coins record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created bills and coins (/bills-and-coins): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.JSON(http.StatusCreated, core.BillAndCoinsManager(service).ToModel(billAndCoins))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/bills-and-coins/:bills_and_coins_id",
		Method:       "PUT",
		RequestType:  core.BillAndCoinsRequest{},
		ResponseType: core.BillAndCoinsResponse{},
		Note:         "Updates an existing bills and coins record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := helpers.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), invalid ID.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}

		req, err := core.BillAndCoinsManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), validation error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), user org error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		billAndCoins, err := core.BillAndCoinsManager(service).GetByID(context, *types.BillAndCoinsID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), record not found.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		billAndCoins.MediaID = req.MediaID
		billAndCoins.Name = req.Name
		billAndCoins.Value = req.Value
		billAndCoins.CurrencyID = req.CurrencyID

		billAndCoins.UpdatedAt = time.Now().UTC()
		billAndCoins.UpdatedByID = userOrg.UserID
		if err := core.BillAndCoinsManager(service).UpdateByID(context, billAndCoins.ID, billAndCoins); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bills and coins update failed (/bills-and-coins/:bills_and_coins_id), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "Failed to update bills and coins record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated bills and coins (/bills-and-coins/:bills_and_coins_id): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.JSON(http.StatusOK, core.BillAndCoinsManager(service).ToModel(billAndCoins))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/bills-and-coins/:bills_and_coins_id",
		Method: "DELETE",
		Note:   "Deletes the specified bills and coins record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		billAndCoinsID, err := helpers.EngineUUIDParam(ctx, "bills_and_coins_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), invalid ID.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bills and coins ID"})
		}
		billAndCoins, err := core.BillAndCoinsManager(service).GetByID(context, *types.BillAndCoinsID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), record not found.",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Bills and coins record not found"})
		}
		if err := core.BillAndCoinsManager(service).Delete(context, *types.BillAndCoinsID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bills and coins delete failed (/bills-and-coins/:bills_and_coins_id), db error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bills and coins record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted bills and coins (/bills-and-coins/:bills_and_coins_id): " + billAndCoins.Name,
			Module:      "BillAndCoins",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/bills-and-coins/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple bills and coins records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete bills and coins (/bills-and-coins/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete bills and coins (/bills-and-coins/bulk-delete) | no IDs provided",
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.BillAndCoinsManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete bills and coins (/bills-and-coins/bulk-delete) | error: " + err.Error(),
				Module:      "BillAndCoins",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete bills and coins records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted bills and coins (/bills-and-coins/bulk-delete)",
			Module:      "BillAndCoins",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
