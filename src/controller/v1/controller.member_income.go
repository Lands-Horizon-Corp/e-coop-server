package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberIncomeController() {
	req := c.provider.Service.Request

	// Create a new income record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-income/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: model.MemberIncomeResponse{},
		RequestType:  model.MemberIncomeRequest{},
		Note:         "Creates a new income record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &model.MemberIncome{
			MemberProfileID: *memberProfileID,
			MediaID:         req.MediaID,
			Name:            req.Name,
			Source:          req.Source,
			Amount:          req.Amount,
			ReleaseDate:     req.ReleaseDate,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}

		if err := c.model.MemberIncomeManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member income failed (/member-income/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member income: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member income (/member-income/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberIncome",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	// Update an existing income record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-income/:member_income_id",
		Method:       "PUT",
		ResponseType: model.MemberIncomeResponse{},
		RequestType:  model.MemberIncomeRequest{},
		Note:         "Updates an existing income record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := handlers.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), invalid member_income_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_income_id: " + err.Error()})
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), validation error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), user org error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberIncomeManager.GetByID(context, *memberIncomeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), record not found: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member income not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MediaID = req.MediaID
		value.Name = req.Name
		value.Amount = req.Amount
		value.ReleaseDate = req.ReleaseDate

		if err := c.model.MemberIncomeManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member income failed (/member-income/:member_income_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member income: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member income (/member-income/:member_income_id): " + value.Name,
			Module:      "MemberIncome",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	// Delete a member's income record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-income/:member_income_id",
		Method: "DELETE",
		Note:   "Deletes a member's income record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := handlers.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), invalid member_income_id: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_income_id: " + err.Error()})
		}
		value, err := c.model.MemberIncomeManager.GetByID(context, *memberIncomeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), record not found: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member income not found: " + err.Error()})
		}
		if err := c.model.MemberIncomeManager.DeleteByID(context, *memberIncomeID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member income failed (/member-income/:member_income_id), db error: " + err.Error(),
				Module:      "MemberIncome",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member income: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member income (/member-income/:member_income_id): " + value.Name,
			Module:      "MemberIncome",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
