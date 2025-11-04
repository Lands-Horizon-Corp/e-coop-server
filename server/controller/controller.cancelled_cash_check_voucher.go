package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// CancelledCashCheckVoucherController registers routes for managing cancelled cash check vouchers.
func (c *Controller) cancelledCashCheckVoucherController() {
	req := c.provider.Service.Request

	// GET /cancelled-cash-check-voucher: List all cancelled cash check vouchers for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher",
		Method:       "GET",
		Note:         "Returns all cancelled cash check vouchers for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := c.core.CancelledCashCheckVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cancelled cash check vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.CancelledCashCheckVoucherManager.ToModels(cancelledVouchers))
	})

	// GET /cancelled-cash-check-voucher/search: Paginated search of cancelled cash check vouchers for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cancelled cash check vouchers for the current user's organization and branch.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := c.core.CancelledCashCheckVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cancelled cash check vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.CancelledCashCheckVoucherManager.Pagination(context, ctx, cancelledVouchers))
	})

	// GET /cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id: Get specific cancelled cash check voucher by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns a single cancelled cash check voucher by its ID.",
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := handlers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}
		cancelledVoucher, err := c.core.CancelledCashCheckVoucherManager.GetByIDRaw(context, *cancelledVoucherID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		return ctx.JSON(http.StatusOK, cancelledVoucher)
	})

	// POST /cancelled-cash-check-voucher: Create a new cancelled cash check voucher. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher",
		Method:       "POST",
		Note:         "Creates a new cancelled cash check voucher for the current user's organization and branch.",
		RequestType:  core.CancelledCashCheckVoucherRequest{},
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.CancelledCashCheckVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), validation error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), user org error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.core.CancelledCashCheckVoucherManager.Create(context, cancelledVoucher); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cancelled cash check voucher creation failed (/cancelled-cash-check-voucher), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cancelled cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cancelled cash check voucher (/cancelled-cash-check-voucher): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.JSON(http.StatusCreated, c.core.CancelledCashCheckVoucherManager.ToModel(cancelledVoucher))
	})

	// PUT /cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id: Update cancelled cash check voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing cancelled cash check voucher by its ID.",
		RequestType:  core.CancelledCashCheckVoucherRequest{},
		ResponseType: core.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := handlers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), invalid ID.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}

		req, err := c.core.CancelledCashCheckVoucherManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), validation error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), user org error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		cancelledVoucher, err := c.core.CancelledCashCheckVoucherManager.GetByID(context, *cancelledVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		cancelledVoucher.UpdatedByID = user.UserID
		if err := c.core.CancelledCashCheckVoucherManager.UpdateByID(context, cancelledVoucher.ID, cancelledVoucher); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancelled cash check voucher update failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cancelled cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cancelled cash check voucher (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.JSON(http.StatusOK, c.core.CancelledCashCheckVoucherManager.ToModel(cancelledVoucher))
	})

	// DELETE /cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id: Delete a cancelled cash check voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method: "DELETE",
		Note:   "Deletes the specified cancelled cash check voucher by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := handlers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), invalid ID.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}
		cancelledVoucher, err := c.core.CancelledCashCheckVoucherManager.GetByID(context, *cancelledVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), not found.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		if err := c.core.CancelledCashCheckVoucherManager.Delete(context, *cancelledVoucherID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), db error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cancelled cash check voucher: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cancelled cash check voucher (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id): " + cancelledVoucher.CheckNumber,
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/cancelled-cash-check-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cancelled cash check vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | no IDs provided",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No cancelled cash check voucher IDs provided for bulk delete"})
		}

		if err := c.core.CancelledCashCheckVoucherManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete) | error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete cancelled cash check vouchers: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete)",
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
