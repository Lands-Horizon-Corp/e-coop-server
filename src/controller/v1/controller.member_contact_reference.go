package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberContactReferenceController() {
	req := c.provider.Service.Request

	// Create a new contact reference for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-contact-reference/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: model.MemberContactReferenceResponse{},
		RequestType:  model.MemberContactReferenceRequest{},
		Note:         "Creates a new contact reference entry for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &model.MemberContactReference{
			MemberProfileID: *memberProfileID,
			Name:            req.Name,
			Description:     req.Description,
			ContactNumber:   req.ContactNumber,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}

		if err := c.model.MemberContactReferenceManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact reference: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created contact reference (/member-contact-reference/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberContactReference",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	// Update an existing contact reference by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-contact-reference/:member_contact_reference_id",
		Method:       "PUT",
		ResponseType: model.MemberContactReferenceResponse{},
		RequestType:  model.MemberContactReferenceRequest{},
		Note:         "Updates an existing contact reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := handlers.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), invalid member_contact_reference_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_contact_reference_id: " + err.Error()})
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), validation error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), user org error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberContactReferenceManager.GetByID(context, *memberContactReferenceID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), record not found: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact reference not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.Name = req.Name
		value.Description = req.Description
		value.ContactNumber = req.ContactNumber

		if err := c.model.MemberContactReferenceManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update contact reference: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated contact reference (/member-contact-reference/:member_contact_reference_id): " + value.Name,
			Module:      "MemberContactReference",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	// Delete a contact reference by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/member-contact-reference/:member_contact_reference_id",
		Method: "DELETE",
		Note:   "Deletes a contact reference entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := handlers.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), invalid member_contact_reference_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_contact_reference_id: " + err.Error()})
		}
		value, err := c.model.MemberContactReferenceManager.GetByID(context, *memberContactReferenceID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), record not found: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact reference not found: " + err.Error()})
		}
		if err := c.model.MemberContactReferenceManager.DeleteByID(context, *memberContactReferenceID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact reference: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted contact reference (/member-contact-reference/:member_contact_reference_id): " + value.Name,
			Module:      "MemberContactReference",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
