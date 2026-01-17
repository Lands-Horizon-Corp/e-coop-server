package charges

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func ChargesRateSchemeModeOfPaymentController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme-mode-of-payment/charges-rate-scheme/:charges_rate_scheme_id",
		Method:       "POST",
		Note:         "Creates a new charges rate scheme model of payment for the current user's organization and branch.",
		RequestType: types.ChargesRateSchemeModeOfPaymentRequest{},
		ResponseType: types.ChargesRateSchemeModeOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), invalid charges rate scheme ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme ID"})
		}
		req, err := core.ChargesRateSchemeModeOfPaymentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), user not assigned to branch.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		chargesRateSchemeModeOfPayment := &types.ChargesRateSchemeModeOfPayment{
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
			CreatedByID:         userOrg.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         userOrg.UserID,
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
		}

		if err := core.ChargesRateSchemeModeOfPaymentManager(service).Create(context, chargesRateSchemeModeOfPayment); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Charges rate scheme model of payment creation failed (/charges-rate-scheme-mode-of-payment), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create charges rate scheme model of payment: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.JSON(http.StatusCreated, core.ChargesRateSchemeModeOfPaymentManager(service).ToModel(chargesRateSchemeModeOfPayment))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method:       "PUT",
		Note:         "Updates an existing charges rate scheme model of payment by its ID.",
		RequestType: types.ChargesRateSchemeModeOfPaymentRequest{},
		ResponseType: types.ChargesRateSchemeModeOfPaymentResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModeOfPaymentID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}

		req, err := core.ChargesRateSchemeModeOfPaymentManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), validation error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), user org error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		chargesRateSchemeModeOfPayment, err := core.ChargesRateSchemeModeOfPaymentManager(service).GetByID(context, *chargesRateSchemeModeOfPaymentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), charges rate scheme model of payment not found.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
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
		chargesRateSchemeModeOfPayment.UpdatedByID = userOrg.UserID
		if err := core.ChargesRateSchemeModeOfPaymentManager(service).UpdateByID(context, chargesRateSchemeModeOfPayment.ID, chargesRateSchemeModeOfPayment); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Charges rate scheme model of payment update failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update charges rate scheme model of payment: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.JSON(http.StatusOK, core.ChargesRateSchemeModeOfPaymentManager(service).ToModel(chargesRateSchemeModeOfPayment))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id",
		Method: "DELETE",
		Note:   "Deletes the specified charges rate scheme model of payment by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		chargesRateSchemeModeOfPaymentID, err := helpers.EngineUUIDParam(ctx, "charges_rate_scheme_model_of_payment_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), invalid charges rate scheme model of payment ID.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid charges rate scheme model of payment ID"})
		}
		chargesRateSchemeModeOfPayment, err := core.ChargesRateSchemeModeOfPaymentManager(service).GetByID(context, *chargesRateSchemeModeOfPaymentID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), not found.",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Charges rate scheme model of payment not found"})
		}
		if err := core.ChargesRateSchemeModeOfPaymentManager(service).Delete(context, *chargesRateSchemeModeOfPaymentID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Charges rate scheme model of payment delete failed (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id), db error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete charges rate scheme model of payment: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted charges rate scheme model of payment (/charges-rate-scheme-mode-of-payment/:charges_rate_scheme_model_of_payment_id): " + chargesRateSchemeModeOfPayment.ID.String(),
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/charges-rate-scheme-mode-of-payment/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple charges rate scheme mode of payment by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete) | no IDs provided",
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No charges rate scheme mode of payment IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.ChargesRateSchemeModeOfPaymentManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/charges-rate-scheme-mode-of-payment/bulk-delete) | error: " + err.Error(),
				Module:      "ChargesRateSchemeModeOfPayment",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete charges rate scheme mode of payment: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted charges rate scheme mode of payment (/charges-rate-scheme-mode-of-payment/bulk-delete)",
			Module:      "ChargesRateSchemeModeOfPayment",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
