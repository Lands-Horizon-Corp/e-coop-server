package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberExpenseController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-expense/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberExpense",
		Response: "TMemberExpense",
		Note:     "Create a new expense record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberExpense{
			MemberProfileID: *memberProfileID,

			Name:        req.Name,
			Amount:      req.Amount,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberExpenseManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-expense/:member_expense_id",
		Method:   "PUT",
		Request:  "TMemberExpense",
		Response: "TMemberExpense",
		Note:     "Update an existing expense record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := horizon.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member expense ID")
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberExpenseManager.GetByID(context, *memberExpenseID)
		if err != nil {
			return c.NotFound(ctx, "MemberExpense")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.Name = req.Name
		value.Amount = req.Amount
		value.Description = req.Description
		if err := c.model.MemberExpenseManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-expense/:member_expense_id",
		Method: "DELETE",
		Note:   "Delete a member's expense record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := horizon.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member expense ID")
		}
		if err := c.model.MemberExpenseManager.DeleteByID(context, *memberExpenseID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
