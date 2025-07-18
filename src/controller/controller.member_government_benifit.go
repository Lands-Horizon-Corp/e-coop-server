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

	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Create a new government benefit record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberGovernmentBenefit{
			MemberProfileID: *memberProfileID,

			FrontMediaID: req.FrontMediaID,
			BackMediaID:  req.BackMediaID,
			CountryCode:  req.CountryCode,
			Description:  req.Description,
			Name:         req.Name,
			Value:        req.Value,
			ExpiryDate:   req.ExpiryDate,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGovernmentBenefitManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/:member_government_benefit_id",
		Method:   "PUT",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Update an existing government benefit record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member government benefit ID")
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberGovernmentBenefitManager.GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			return c.NotFound(ctx, "MemberGovernmentBenefit")
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-government-benefit/:member_government_benefit_id",
		Method: "DELETE",
		Note:   "Delete a member's government benefit record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member government benefit ID")
		}

		if err := c.model.MemberGovernmentBenefitManager.DeleteByID(context, *memberGovernmentBenefitID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
