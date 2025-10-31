package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
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
		ResponseType: modelcore.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := c.modelcore.CancelledCashCheckVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cancelled cash check vouchers found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CancelledCashCheckVoucherManager.Filtered(context, ctx, cancelledVouchers))
	})

	// GET /cancelled-cash-check-voucher/search: Paginated search of cancelled cash check vouchers for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cancelled cash check vouchers for the current user's organization and branch.",
		ResponseType: modelcore.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		cancelledVouchers, err := c.modelcore.CancelledCashCheckVoucherCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cancelled cash check vouchers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.CancelledCashCheckVoucherManager.Pagination(context, ctx, cancelledVouchers))
	})

	// GET /cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id: Get specific cancelled cash check voucher by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns a single cancelled cash check voucher by its ID.",
		ResponseType: modelcore.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cancelledVoucherID, err := handlers.EngineUUIDParam(ctx, "cancelled_cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cancelled cash check voucher ID"})
		}
		cancelledVoucher, err := c.modelcore.CancelledCashCheckVoucherManager.GetByIDRaw(context, *cancelledVoucherID)
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
		RequestType:  modelcore.CancelledCashCheckVoucherRequest{},
		ResponseType: modelcore.CancelledCashCheckVoucherResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.CancelledCashCheckVoucherManager.Validate(ctx)
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

		cancelledVoucher := &modelcore.CancelledCashCheckVoucher{
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

		if err := c.modelcore.CancelledCashCheckVoucherManager.Create(context, cancelledVoucher); err != nil {
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
		return ctx.JSON(http.StatusCreated, c.modelcore.CancelledCashCheckVoucherManager.ToModel(cancelledVoucher))
	})

	// PUT /cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id: Update cancelled cash check voucher by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id",
		Method:       "PUT",
		Note:         "Updates an existing cancelled cash check voucher by its ID.",
		RequestType:  modelcore.CancelledCashCheckVoucherRequest{},
		ResponseType: modelcore.CancelledCashCheckVoucherResponse{},
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

		req, err := c.modelcore.CancelledCashCheckVoucherManager.Validate(ctx)
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
		cancelledVoucher, err := c.modelcore.CancelledCashCheckVoucherManager.GetByID(context, *cancelledVoucherID)
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
		if err := c.modelcore.CancelledCashCheckVoucherManager.UpdateFields(context, cancelledVoucher.ID, cancelledVoucher); err != nil {
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
		return ctx.JSON(http.StatusOK, c.modelcore.CancelledCashCheckVoucherManager.ToModel(cancelledVoucher))
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
		cancelledVoucher, err := c.modelcore.CancelledCashCheckVoucherManager.GetByID(context, *cancelledVoucherID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cancelled cash check voucher delete failed (/cancelled-cash-check-voucher/:cancelled_cash_check_voucher_id), not found.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cancelled cash check voucher not found"})
		}
		if err := c.modelcore.CancelledCashCheckVoucherManager.DeleteByID(context, *cancelledVoucherID); err != nil {
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

	// DELETE /cancelled-cash-check-voucher/bulk-delete: Bulk delete cancelled cash check vouchers by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/cancelled-cash-check-voucher/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cancelled cash check vouchers by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), invalid request body.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), no IDs provided.",
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No cancelled cash check voucher IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		var sb strings.Builder
		for _, rawID := range reqBody.IDs {
			cancelledVoucherID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), invalid UUID: " + rawID,
					Module:      "CancelledCashCheckVoucher",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			cancelledVoucher, err := c.modelcore.CancelledCashCheckVoucherManager.GetByID(context, cancelledVoucherID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), not found: " + rawID,
					Module:      "CancelledCashCheckVoucher",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Cancelled cash check voucher not found with ID: %s", rawID)})
			}
			sb.WriteString(cancelledVoucher.CheckNumber)
			sb.WriteByte(',')
			if err := c.modelcore.CancelledCashCheckVoucherManager.DeleteByIDWithTx(context, tx, cancelledVoucherID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), db error: " + err.Error(),
					Module:      "CancelledCashCheckVoucher",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cancelled cash check voucher: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/cancelled-cash-check-voucher/bulk-delete), commit error: " + err.Error(),
				Module:      "CancelledCashCheckVoucher",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cancelled cash check vouchers (/cancelled-cash-check-voucher/bulk-delete): " + sb.String(),
			Module:      "CancelledCashCheckVoucher",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
