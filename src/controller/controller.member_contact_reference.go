package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberContactReferenceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-contact-reference/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberContactReference",
		Response: "TMemberContactReference",
		Note:     "Create a new contact reference for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberContactReference{
			MemberProfileID: *memberProfileID,

			Name:          req.Name,
			Description:   req.Description,
			ContactNumber: req.ContactNumber,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberContactReferenceManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-contact-reference/:member_contact_reference_id",
		Method:   "PUT",
		Request:  "TMemberContactReference",
		Response: "TMemberContactReference",
		Note:     "Update an existing contact reference for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := horizon.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member contact reference ID")
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberContactReferenceManager.GetByID(context, *memberContactReferenceID)
		if err != nil {
			return c.NotFound(ctx, "MemberContactReference")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.Name = req.Name
		value.Description = req.Description
		value.ContactNumber = req.ContactNumber

		if err := c.model.MemberContactReferenceManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-contact-reference/:member_contact_reference_id",
		Method: "DELETE",
		Note:   "Delete a member's contact reference by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := horizon.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member contact reference ID")
		}
		if err := c.model.MemberContactReferenceManager.DeleteByID(context, *memberContactReferenceID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
