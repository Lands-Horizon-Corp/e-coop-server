package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberExpenseController() {
	req := c.provider.Service.Request

	// Create a new expense record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-expense/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: model.MemberExpenseResponse{},
		RequestType:  model.MemberExpenseRequest{},
		Note:         "Creates a new expense record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &model.MemberExpense{
			MemberProfileID: *memberProfileID,
			Name:            req.Name,
			Amount:          req.Amount,
			Description:     req.Description,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}

		if err := c.model.MemberExpenseManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member expense: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member expense (/member-expense/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberExpense",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	// Update an existing expense record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-expense/:member_expense_id",
		Method:       "PUT",
		RequestType:  model.MemberExpenseRequest{},
		ResponseType: model.MemberExpenseResponse{},
		Note:         "Updates an existing expense record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := handlers.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), invalid member_expense_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_expense_id: " + err.Error()})
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), validation error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), user org error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberExpenseManager.GetByID(context, *memberExpenseID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), record not found: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member expense not found: " + err.Error()})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member expense: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member expense (/member-expense/:member_expense_id): " + value.Name,
			Module:      "MemberExpense",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	// Delete a member's expense record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/member-expense/:member_expense_id",
		Method: "DELETE",
		Note:   "Deletes a member's expense record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := handlers.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), invalid member_expense_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_expense_id: " + err.Error()})
		}
		value, err := c.model.MemberExpenseManager.GetByID(context, *memberExpenseID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), record not found: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member expense not found: " + err.Error()})
		}
		if err := c.model.MemberExpenseManager.DeleteByID(context, *memberExpenseID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member expense: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member expense (/member-expense/:member_expense_id): " + value.Name,
			Module:      "MemberExpense",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
