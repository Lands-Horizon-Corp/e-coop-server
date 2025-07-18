package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberGovernmentBenefitController() {
	req := c.provider.Service.Request

	// Create a new government benefit record for a member profile
	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Creates a new government benefit record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &model.MemberGovernmentBenefit{
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

		if err := c.model.MemberGovernmentBenefitManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create government benefit record: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	// Update an existing government benefit record by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/:member_government_benefit_id",
		Method:   "PUT",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Updates an existing government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberGovernmentBenefitManager.GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
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

		if err := c.model.MemberGovernmentBenefitManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update government benefit record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	// Delete a government benefit record by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-government-benefit/:member_government_benefit_id",
		Method: "DELETE",
		Note:   "Deletes a member's government benefit record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_government_benefit_id: " + err.Error()})
		}
		if err := c.model.MemberGovernmentBenefitManager.DeleteByID(context, *memberGovernmentBenefitID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete government benefit record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
