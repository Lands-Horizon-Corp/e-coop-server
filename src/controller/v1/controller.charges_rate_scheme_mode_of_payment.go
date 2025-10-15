package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ChargesRateSchemeModeOfPaymentController registers routes for managing charges rate scheme model of payment.
func (c *Controller) ChargesRateSchemeModeOfPaymentController() {
	req := c.provider.Service.Request

	// POST /charges-rate-scheme-mode-of-payment: Create a new charges rate scheme model of payment. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme-mode-of-payment/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "POST",
		Note:         "Creates a new charges rate scheme model of payment for the current user's organization and branch.",
		RequestType:  model_core.ModeOfPayment{},
		ResponseType: model_core.ChargesRateSchemeModeOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), invalid charges rate scheme ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		req, err := c.model_core.ChargesRateSchemeModeOfPaymentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), user not assigned to branch.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateSchemeModeOfPayment := &model_core.ChargesRateSchemeModeOfPayment{
			ChargesRateSchemeID: *chargesRateSchemeID,
			From:                req.From,
			To:                  req.To,
			Column1:             req.Column1,
			Column2:             req.Column2,
			Column3:             req.Column3,
			Column4:             req.Column4,
			Column5:             req.Column5,
			Column6:             req.Column6,
			Column7:             req.Column7,
			Column8:             req.Column8,
			Column9:             req.Column9,
			Column10:            req.Column10,
			Column11:            req.Column11,
			Column12:            req.Column12,
			Column13:            req.Column13,
			Column14:            req.Column14,
			Column15:            req.Column15,
			Column16:            req.Column16,
			Column17:            req.Column17,
			Column18:            req.Column18,
			Column19:            req.Column19,
			Column20:            req.Column20,
			Column21:            req.Column21,
			Column22:            req.Column22,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         user.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         user.UserID,
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
		}

		if err := c.model_core.ChargesRateSchemeModeOfPaymentManager.Create(context, chargesRateSchemeModeOfPayment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.JSON(http.StatusCreated, c.model_core.ChargesRateSchemeModeOfPaymentManager.ToModel(chargesRateSchemeModeOfPayment))
	})

	// PUT /charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id: Update charges rate scheme model of payment by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme model of payment by its ID.",
		RequestType:  model_core.ModeOfPayment{},
		ResponseType: model_core.ChargesRateSchemeModeOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModeOfPaymentID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}

		req, err := c.model_core.ChargesRateSchemeModeOfPaymentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateSchemeModeOfPayment, err := c.model_core.ChargesRateSchemeModeOfPaymentManager.GetByID(context, *chargesRateSchemeModeOfPaymentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), charges rate scheme model of payment not found.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
		chargesRateSchemeModeOfPayment.ChargesRateSchemeID = req.ChargesRateSchemeID
		chargesRateSchemeModeOfPayment.From = req.From
		chargesRateSchemeModeOfPayment.To = req.To
		chargesRateSchemeModeOfPayment.Column1 = req.Column1
		chargesRateSchemeModeOfPayment.Column2 = req.Column2
		chargesRateSchemeModeOfPayment.Column3 = req.Column3
		chargesRateSchemeModeOfPayment.Column4 = req.Column4
		chargesRateSchemeModeOfPayment.Column5 = req.Column5
		chargesRateSchemeModeOfPayment.Column6 = req.Column6
		chargesRateSchemeModeOfPayment.Column7 = req.Column7
		chargesRateSchemeModeOfPayment.Column8 = req.Column8
		chargesRateSchemeModeOfPayment.Column9 = req.Column9
		chargesRateSchemeModeOfPayment.Column10 = req.Column10
		chargesRateSchemeModeOfPayment.Column11 = req.Column11
		chargesRateSchemeModeOfPayment.Column12 = req.Column12
		chargesRateSchemeModeOfPayment.Column13 = req.Column13
		chargesRateSchemeModeOfPayment.Column14 = req.Column14
		chargesRateSchemeModeOfPayment.Column15 = req.Column15
		chargesRateSchemeModeOfPayment.Column16 = req.Column16
		chargesRateSchemeModeOfPayment.Column17 = req.Column17
		chargesRateSchemeModeOfPayment.Column18 = req.Column18
		chargesRateSchemeModeOfPayment.Column19 = req.Column19
		chargesRateSchemeModeOfPayment.Column20 = req.Column20
		chargesRateSchemeModeOfPayment.Column21 = req.Column21
		chargesRateSchemeModeOfPayment.Column22 = req.Column22
		chargesRateSchemeModeOfPayment.UpdatedAt = time.Now().UTC()
		chargesRateSchemeModeOfPayment.UpdatedByID = user.UserID
		if err := c.model_core.ChargesRateSchemeModeOfPaymentManager.UpdateFields(context, chargesRateSchemeModeOfPayment.ID, chargesRateSchemeModeOfPayment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.JSON(http.StatusOK, c.model_core.ChargesRateSchemeModeOfPaymentManager.ToModel(chargesRateSchemeModeOfPayment))
	})

	// DELETE /charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id: Delete a charges rate scheme model of payment by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate scheme model of payment by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModeOfPaymentID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}
		chargesRateSchemeModeOfPayment, err := c.model_core.ChargesRateSchemeModeOfPaymentManager.GetByID(context, *chargesRateSchemeModeOfPaymentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), not found.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
		if err := c.model_core.ChargesRateSchemeModeOfPaymentManager.DeleteByID(context, *chargesRateSchemeModeOfPaymentID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /charges-rate-scheme-mode-of-payment/bulk-delete: Bulk delete charges rate scheme model of payment by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/charges-rate-scheme-mode-of-payment/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate scheme model of payment by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model_core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model_core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), invalid request body.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), no IDs provided.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate scheme model of payment IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		ids := ""
		for _, rawID := range reqBody.IDs {
			chargesRateSchemeModeOfPaymentID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), invalid UUID: " + rawID,
					Module:      "ChargesRateSchemeModeOfPayment",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			chargesRateSchemeModeOfPayment, err := c.model_core.ChargesRateSchemeModeOfPaymentManager.GetByID(context, chargesRateSchemeModeOfPaymentID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), not found: " + rawID,
					Module:      "ChargesRateSchemeModeOfPayment",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Charges rate scheme model of payment not found with ID: %s", rawID)})
			}
			ids += chargesRateSchemeModeOfPayment.ID.String() + ","
			if err := c.model_core.ChargesRateSchemeModeOfPaymentManager.DeleteByIDWithTx(context, tx, chargesRateSchemeModeOfPaymentID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), db error: " + err.Error(),
					Module:      "ChargesRateSchemeModeOfPayment",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme model of payment: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete), commit error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment/bulk-delete): " + ids,
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
