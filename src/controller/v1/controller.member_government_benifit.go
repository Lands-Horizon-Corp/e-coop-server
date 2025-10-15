package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberGovernmentBenefitController() {
	req := c.provider.Service.Request

	// Create a new government benefit record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-government-benefit/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: model_core.MemberGovernmentBenefitResponse{},
		RequestType:  model_core.MemberGovernmentBenefitRequest{},
		Note:         "Creates a new government benefit record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model_core.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &model_core.MemberGovernmentBenefit{
			MemberProfileID: *memberProfileID,
			FrontMediaID:    req.FrontMediaID,
			BackMediaID:     req.BackMediaID,
			CountryCode:     req.CountryCode,
			Description:     req.Description,
			Name:            req.Name,
			Value:           req.Value,
			ExpiryDate:      req.ExpiryDate,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}

		if err := c.model_core.MemberGovernmentBenefitManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create government benefit record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created government benefit (/member-government-benefit/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})

		return ctx.JSON(http.StatusOK, c.model_core.MemberGovernmentBenefitManager.ToModel(value))
	})

	// Update an existing government benefit record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-government-benefit/:member_government_benefit_id",
		Method:       "PUT",
		ResponseType: model_core.MemberGovernmentBenefitResponse{},
		RequestType:  model_core.MemberGovernmentBenefitRequest{},
		Note:         "Updates an existing government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := handlers.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), invalid member_government_benefit_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		req, err := c.model_core.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), validation error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), user org error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model_core.MemberGovernmentBenefitManager.GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), record not found: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Government benefit record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.FrontMediaID = req.FrontMediaID
		value.BackMediaID = req.BackMediaID
		value.CountryCode = req.CountryCode
		value.Description = req.Description
		value.Name = req.Name
		value.Value = req.Value
		value.ExpiryDate = req.ExpiryDate

		if err := c.model_core.MemberGovernmentBenefitManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update government benefit record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated government benefit (/member-government-benefit/:member_government_benefit_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})
		return ctx.JSON(http.StatusOK, c.model_core.MemberGovernmentBenefitManager.ToModel(value))
	})

	// Delete a government benefit record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-government-benefit/:member_government_benefit_id",
		Method: "DELETE",
		Note:   "Deletes a member's government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := handlers.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), invalid member_government_benefit_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		value, err := c.model_core.MemberGovernmentBenefitManager.GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), record not found: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Government benefit record not found: " + err.Error()})
		}
		if err := c.model_core.MemberGovernmentBenefitManager.DeleteByID(context, *memberGovernmentBenefitID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete government benefit record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted government benefit (/member-government-benefit/:member_government_benefit_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
