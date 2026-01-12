package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func cancelledCashCheckVoucherController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher",
		Method:       "GET",
		Note:         "Returns all cancelled cash check vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := core.CancelledCashCheckVoucherCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cancelled cash check vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.CancelledCashCheckVoucherManager(service).ToModels(cancelledVouchers))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cancelled cash check vouchers for the current user's organization and branch.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := core.CancelledCashCheckVoucherManager(service).NormalPagination(context, ctx, &core.CancelledCashCheckVoucher{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cancelled cash check vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, cancelledVouchers)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns a single cancelled cash check voucher by its ID.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := helpers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}
		cancelledVoucher, err := core.CancelledCashCheckVoucherManager(service).GetByIDRaw(context, *cancelledVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		return ctx.JSON(http.StatusOK, cancelledVoucher)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher",
		Method:       "POST",
		Note:         "Creates a new cancelled cash check voucher for the current user's organization and branch.",
		RequestType:  core.CancelledCashCheckVoucherRequest{},
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CancelledCashCheckVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), validation error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), user org error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), user not assigned to branch.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		cancelledVoucher := &core.CancelledCashCheckVoucher{
			CheckNumber:    req.CheckNumber,
			EntryDate:      req.EntryDate,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.CancelledCashCheckVoucherManager(service).Create(context, cancelledVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cancelled cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cancelled cash check voucher (/cancelled-cash-check-voucher): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.JSON(http.StatusCreated, core.CancelledCashCheckVoucherManager(service).ToModel(cancelledVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing cancelled cash check voucher by its ID.",
		RequestType:  core.CancelledCashCheckVoucherRequest{},
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := helpers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), invalid ID.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}

		req, err := core.CancelledCashCheckVoucherManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), validation error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), user org error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		cancelledVoucher, err := core.CancelledCashCheckVoucherManager(service).GetByID(context, *cancelledVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), not found.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		cancelledVoucher.CheckNumber = req.CheckNumber
		cancelledVoucher.EntryDate = req.EntryDate
		cancelledVoucher.Description = req.Description
		cancelledVoucher.UpdatedAt = time.Now().UTC()
		cancelledVoucher.UpdatedByID = userOrg.UserID
		if err := core.CancelledCashCheckVoucherManager(service).UpdateByID(context, cancelledVoucher.ID, cancelledVoucher); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cancelled cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cancelled cash check voucher (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.JSON(http.StatusOK, core.CancelledCashCheckVoucherManager(service).ToModel(cancelledVoucher))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified cancelled cash check voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := helpers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), invalid ID.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}
		cancelledVoucher, err := core.CancelledCashCheckVoucherManager(service).GetByID(context, *cancelledVoucherID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), not found.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		if err := core.CancelledCashCheckVoucherManager(service).Delete(context, *cancelledVoucherID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cancelled cash check voucher: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cancelled cash check voucher (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/cancelled-cash-check-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cancelled cash check vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | no IDs provided",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No cancelled cash check voucher IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CancelledCashCheckVoucherManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete cancelled cash check vouchers: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete)",
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
