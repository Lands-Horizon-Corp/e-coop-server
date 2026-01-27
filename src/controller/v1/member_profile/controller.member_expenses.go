package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberExpenseController(service *horizon.HorizonService) {
	

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-expense/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: types.MemberExpenseResponse{},
		RequestType:  types.MemberExpenseRequest{},
		Note:         "Creates a new expense record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberExpenseManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &types.MemberExpense{
			MemberProfileID: *memberProfileID,
			Name:            req.Name,
			Amount:          req.Amount,
			Description:     req.Description,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}

		if err := core.MemberExpenseManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member expense failed (/member-expense/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member expense: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member expense (/member-expense/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberExpense",
		})

		return ctx.JSON(http.StatusOK, core.MemberExpenseManager(service).ToModel(value))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-expense/:member_expense_id",
		Method:       "PUT",
		RequestType:  types.MemberExpenseRequest{},
		ResponseType: types.MemberExpenseResponse{},
		Note:         "Updates an existing expense record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := helpers.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), invalid member_expense_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_expense_id: " + err.Error()})
		}
		req, err := core.MemberExpenseManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), validation error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), user org error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberExpenseManager(service).GetByID(context, *memberExpenseID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), record not found: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member expense not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.Name = req.Name
		value.Amount = req.Amount
		value.Description = req.Description

		if err := core.MemberExpenseManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member expense failed (/member-expense/:member_expense_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member expense: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member expense (/member-expense/:member_expense_id): " + value.Name,
			Module:      "MemberExpense",
		})
		return ctx.JSON(http.StatusOK, core.MemberExpenseManager(service).ToModel(value))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-expense/:member_expense_id",
		Method: "DELETE",
		Note:   "Deletes a member's expense record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := helpers.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), invalid member_expense_id: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_expense_id: " + err.Error()})
		}
		value, err := core.MemberExpenseManager(service).GetByID(context, *memberExpenseID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), record not found: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member expense not found: " + err.Error()})
		}
		if err := core.MemberExpenseManager(service).Delete(context, *memberExpenseID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member expense failed (/member-expense/:member_expense_id), db error: " + err.Error(),
				Module:      "MemberExpense",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member expense: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member expense (/member-expense/:member_expense_id): " + value.Name,
			Module:      "MemberExpense",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
