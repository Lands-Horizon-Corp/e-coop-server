package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberJointAccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-joint-account/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberJointAccount",
		Response: "TMemberJointAccount",
		Note:     "Create a new joint account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberJointAccount{
			MemberProfileID: *memberProfileID,

			PictureMediaID:     req.PictureMediaID,
			SignatureMediaID:   req.SignatureMediaID,
			Description:        req.Description,
			FirstName:          req.FirstName,
			MiddleName:         req.MiddleName,
			LastName:           req.LastName,
			FullName:           req.FullName,
			Suffix:             req.Suffix,
			Birthday:           req.Birthday,
			FamilyRelationship: req.FamilyRelationship,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberJointAccountManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberJointAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-joint-account/:member_joint_account_id",
		Method:   "PUT",
		Request:  "TMemberJointAccount",
		Response: "TMemberJointAccount",
		Note:     "Update an existing joint account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := horizon.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member joint account ID")
		}
		req, err := c.model.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberJointAccountManager.GetByID(context, *memberJointAccountID)
		if err != nil {
			return c.NotFound(ctx, "MemberJointAccount")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.PictureMediaID = req.PictureMediaID
		value.SignatureMediaID = req.SignatureMediaID
		value.Description = req.Description
		value.FirstName = req.FirstName
		value.MiddleName = req.MiddleName
		value.LastName = req.LastName
		value.FullName = req.FullName
		value.Suffix = req.Suffix
		value.Birthday = req.Birthday
		value.FamilyRelationship = req.FamilyRelationship

		if err := c.model.MemberJointAccountManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberJointAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-joint-account/:member_joint_account_id",
		Method: "DELETE",
		Note:   "Delete a member's joint account record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := horizon.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member joint account ID")
		}
		if err := c.model.MemberJointAccountManager.DeleteByID(context, *memberJointAccountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
