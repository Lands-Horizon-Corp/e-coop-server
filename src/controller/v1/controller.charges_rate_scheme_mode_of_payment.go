package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// ChargesRateSchemeModelOfPaymentController registers routes for managing charges rate scheme model of payment.
func (c *Controller) ChargesRateSchemeModelOfPaymentController() {
	req := c.provider.Service.Request

	// POST /charges-rate-scheme-model-of-payment: Create a new charges rate scheme model of payment. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme-model-of-payment/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "POST",
		Note:         "Creates a new charges rate scheme model of payment for the current user's organization and branch.",
		RequestType:  model.ChargesRateSchemeModelOfPaymentRequest{},
		ResponseType: model.ChargesRateSchemeModelOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-model-of-payment), invalid charges rate scheme ID.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		req, err := c.model.ChargesRateSchemeModelOfPaymentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-model-of-payment), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-model-of-payment), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-model-of-payment), user not assigned to branch.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateSchemeModelOfPayment := &model.ChargesRateSchemeModelOfPayment{
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

		if err := c.model.ChargesRateSchemeModelOfPaymentManager.Create(context, chargesRateSchemeModelOfPayment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-model-of-payment), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate scheme model of payment (/charges-rate-scheme-model-of-payment): " + chargesRateSchemeModelOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModelOfPayment",
		})
		return ctx.JSON(http.StatusCreated, c.model.ChargesRateSchemeModelOfPaymentManager.ToModel(chargesRateSchemeModelOfPayment))
	})

	// PUT /charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id: Update charges rate scheme model of payment by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme model of payment by its ID.",
		RequestType:  model.ChargesRateSchemeModelOfPaymentRequest{},
		ResponseType: model.ChargesRateSchemeModelOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModelOfPaymentID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}

		req, err := c.model.ChargesRateSchemeModelOfPaymentManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateSchemeModelOfPayment, err := c.model.ChargesRateSchemeModelOfPaymentManager.GetByID(context, *chargesRateSchemeModelOfPaymentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), charges rate scheme model of payment not found.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
		chargesRateSchemeModelOfPayment.ChargesRateSchemeID = req.ChargesRateSchemeID
		chargesRateSchemeModelOfPayment.From = req.From
		chargesRateSchemeModelOfPayment.To = req.To
		chargesRateSchemeModelOfPayment.Column1 = req.Column1
		chargesRateSchemeModelOfPayment.Column2 = req.Column2
		chargesRateSchemeModelOfPayment.Column3 = req.Column3
		chargesRateSchemeModelOfPayment.Column4 = req.Column4
		chargesRateSchemeModelOfPayment.Column5 = req.Column5
		chargesRateSchemeModelOfPayment.Column6 = req.Column6
		chargesRateSchemeModelOfPayment.Column7 = req.Column7
		chargesRateSchemeModelOfPayment.Column8 = req.Column8
		chargesRateSchemeModelOfPayment.Column9 = req.Column9
		chargesRateSchemeModelOfPayment.Column10 = req.Column10
		chargesRateSchemeModelOfPayment.Column11 = req.Column11
		chargesRateSchemeModelOfPayment.Column12 = req.Column12
		chargesRateSchemeModelOfPayment.Column13 = req.Column13
		chargesRateSchemeModelOfPayment.Column14 = req.Column14
		chargesRateSchemeModelOfPayment.Column15 = req.Column15
		chargesRateSchemeModelOfPayment.Column16 = req.Column16
		chargesRateSchemeModelOfPayment.Column17 = req.Column17
		chargesRateSchemeModelOfPayment.Column18 = req.Column18
		chargesRateSchemeModelOfPayment.Column19 = req.Column19
		chargesRateSchemeModelOfPayment.Column20 = req.Column20
		chargesRateSchemeModelOfPayment.Column21 = req.Column21
		chargesRateSchemeModelOfPayment.Column22 = req.Column22
		chargesRateSchemeModelOfPayment.UpdatedAt = time.Now().UTC()
		chargesRateSchemeModelOfPayment.UpdatedByID = user.UserID
		if err := c.model.ChargesRateSchemeModelOfPaymentManager.UpdateFields(context, chargesRateSchemeModelOfPayment.ID, chargesRateSchemeModelOfPayment); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme model of payment (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModelOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModelOfPayment",
		})
		return ctx.JSON(http.StatusOK, c.model.ChargesRateSchemeModelOfPaymentManager.ToModel(chargesRateSchemeModelOfPayment))
	})

	// DELETE /charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id: Delete a charges rate scheme model of payment by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate scheme model of payment by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModelOfPaymentID, err := handlers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}
		chargesRateSchemeModelOfPayment, err := c.model.ChargesRateSchemeModelOfPaymentManager.GetByID(context, *chargesRateSchemeModelOfPaymentID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), not found.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
		if err := c.model.ChargesRateSchemeModelOfPaymentManager.DeleteByID(context, *chargesRateSchemeModelOfPaymentID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme model of payment: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate scheme model of payment (/charges-rate-scheme-model-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModelOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModelOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /charges-rate-scheme-model-of-payment/bulk-delete: Bulk delete charges rate scheme model of payment by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/charges-rate-scheme-model-of-payment/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate scheme model of payment by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), invalid request body.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), no IDs provided.",
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate scheme model of payment IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		ids := ""
		for _, rawID := range reqBody.IDs {
			chargesRateSchemeModelOfPaymentID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), invalid UUID: " + rawID,
					Module:      "ChargesRateSchemeModelOfPayment",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			chargesRateSchemeModelOfPayment, err := c.model.ChargesRateSchemeModelOfPaymentManager.GetByID(context, chargesRateSchemeModelOfPaymentID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), not found: " + rawID,
					Module:      "ChargesRateSchemeModelOfPayment",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Charges rate scheme model of payment not found with ID: %s", rawID)})
			}
			ids += chargesRateSchemeModelOfPayment.ID.String() + ","
			if err := c.model.ChargesRateSchemeModelOfPaymentManager.DeleteByIDWithTx(context, tx, chargesRateSchemeModelOfPaymentID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), db error: " + err.Error(),
					Module:      "ChargesRateSchemeModelOfPayment",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme model of payment: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-model-of-payment/bulk-delete), commit error: " + err.Error(),
				Module:      "ChargesRateSchemeModelOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate scheme model of payment (/charges-rate-scheme-model-of-payment/bulk-delete): " + ids,
			Module:      "ChargesRateSchemeModelOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
