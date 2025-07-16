package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) PaymentTypeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/payment-type",
		Method:   "GET",
		Response: "TPaymentType[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		paymentTypes, err := c.model.PaymentTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "PaymentType")
		}
		return ctx.JSON(http.StatusOK, c.model.PaymentTypeManager.ToModels(paymentTypes))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/payment-type/search",
		Method:   "GET",
		Request:  "Filter<IPaymentType>",
		Response: "Paginated<IPaymentType>",
		Note:     "Get pagination payment types",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.PaymentTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.PaymentTypeManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/payment-type/:payment_type_id",
		Method:   "GET",
		Response: "TPaymentType",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := horizon.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid payment type ID")
		}
		paymentType, err := c.model.PaymentTypeManager.GetByIDRaw(context, *paymentTypeID)
		if err != nil {
			return c.NotFound(ctx, "PaymentType")
		}
		return ctx.JSON(http.StatusOK, paymentType)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/payment-type",
		Method:   "POST",
		Request:  "TPaymentType",
		Response: "TPaymentType",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.PaymentTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		paymentType := &model.PaymentType{
			Name:           req.Name,
			Description:    req.Description,
			NumberOfDays:   req.NumberOfDays,
			Type:           req.Type,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.PaymentTypeManager.Create(context, paymentType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.PaymentTypeManager.ToModel(paymentType))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/payment-type/:payment_type_id",
		Method:   "PUT",
		Request:  "TPaymentType",
		Response: "TPaymentType",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := horizon.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid payment type ID")
		}

		req, err := c.model.PaymentTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *paymentTypeID)
		if err != nil {
			return c.NotFound(ctx, "PaymentType")
		}
		paymentType.Name = req.Name
		paymentType.Description = req.Description
		paymentType.NumberOfDays = req.NumberOfDays
		paymentType.Type = req.Type
		paymentType.UpdatedAt = time.Now().UTC()
		paymentType.UpdatedByID = user.UserID
		if err := c.model.PaymentTypeManager.UpdateFields(context, paymentType.ID, paymentType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.PaymentTypeManager.ToModel(paymentType))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/payment-type/:payment_type_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := horizon.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid payment type ID")
		}
		if err := c.model.PaymentTypeManager.DeleteByID(context, *paymentTypeID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/payment-type/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple payment type records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			paymentTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.PaymentTypeManager.GetByID(context, paymentTypeID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("PaymentType with ID %s", rawID))
			}
			if err := c.model.PaymentTypeManager.DeleteByIDWithTx(context, tx, paymentTypeID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
