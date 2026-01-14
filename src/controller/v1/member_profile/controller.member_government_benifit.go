package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func MemberGovernmentBenefitController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-government-benefit/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: core.MemberGovernmentBenefitResponse{},
		RequestType:  core.MemberGovernmentBenefitRequest{},
		Note:         "Creates a new government benefit record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberGovernmentBenefitManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &core.MemberGovernmentBenefit{
			MemberProfileID: *memberProfileID,
			FrontMediaID:    req.FrontMediaID,
			BackMediaID:     req.BackMediaID,
			CountryCode:     req.CountryCode,
			Description:     req.Description,
			Name:            req.Name,
			Value:           req.Value,
			ExpiryDate:      req.ExpiryDate,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     &userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     &userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}

		if err := core.MemberGovernmentBenefitManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create government benefit failed (/member-government-benefit/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create government benefit record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created government benefit (/member-government-benefit/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})

		return ctx.JSON(http.StatusOK, core.MemberGovernmentBenefitManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-government-benefit/:member_government_benefit_id",
		Method:       "PUT",
		ResponseType: core.MemberGovernmentBenefitResponse{},
		RequestType:  core.MemberGovernmentBenefitRequest{},
		Note:         "Updates an existing government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := helpers.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), invalid member_government_benefit_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		req, err := core.MemberGovernmentBenefitManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), validation error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), user org error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberGovernmentBenefitManager(service).GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), record not found: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Government benefit record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = &userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.FrontMediaID = req.FrontMediaID
		value.BackMediaID = req.BackMediaID
		value.CountryCode = req.CountryCode
		value.Description = req.Description
		value.Name = req.Name
		value.Value = req.Value
		value.ExpiryDate = req.ExpiryDate

		if err := core.MemberGovernmentBenefitManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update government benefit failed (/member-government-benefit/:member_government_benefit_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update government benefit record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated government benefit (/member-government-benefit/:member_government_benefit_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})
		return ctx.JSON(http.StatusOK, core.MemberGovernmentBenefitManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-government-benefit/:member_government_benefit_id",
		Method: "DELETE",
		Note:   "Deletes a member's government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := helpers.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), invalid member_government_benefit_id: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		value, err := core.MemberGovernmentBenefitManager(service).GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), record not found: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Government benefit record not found: " + err.Error()})
		}
		if err := core.MemberGovernmentBenefitManager(service).Delete(context, *memberGovernmentBenefitID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete government benefit failed (/member-government-benefit/:member_government_benefit_id), db error: " + err.Error(),
				Module:      "MemberGovernmentBenefit",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete government benefit record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted government benefit (/member-government-benefit/:member_government_benefit_id): " + value.Name,
			Module:      "MemberGovernmentBenefit",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
