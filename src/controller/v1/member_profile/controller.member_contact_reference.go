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

func MemberContactReferenceController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-contact-reference/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: core.MemberContactReferenceResponse{},
		RequestType:  core.MemberContactReferenceRequest{},
		Note:         "Creates a new contact reference entry for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberContactReferenceManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &core.MemberContactReference{
			MemberProfileID: *memberProfileID,
			Name:            req.Name,
			Description:     req.Description,
			ContactNumber:   req.ContactNumber,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}

		if err := core.MemberContactReferenceManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create contact reference failed (/member-contact-reference/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create contact reference: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created contact reference (/member-contact-reference/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberContactReference",
		})

		return ctx.JSON(http.StatusOK, core.MemberContactReferenceManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-contact-reference/:member_contact_reference_id",
		Method:       "PUT",
		ResponseType: core.MemberContactReferenceResponse{},
		RequestType:  core.MemberContactReferenceRequest{},
		Note:         "Updates an existing contact reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := helpers.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), invalid member_contact_reference_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_contact_reference_id: " + err.Error()})
		}
		req, err := core.MemberContactReferenceManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), validation error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), user org error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberContactReferenceManager(service).GetByID(context, *memberContactReferenceID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), record not found: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact reference not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.Name = req.Name
		value.Description = req.Description
		value.ContactNumber = req.ContactNumber

		if err := core.MemberContactReferenceManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update contact reference failed (/member-contact-reference/:member_contact_reference_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update contact reference: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated contact reference (/member-contact-reference/:member_contact_reference_id): " + value.Name,
			Module:      "MemberContactReference",
		})
		return ctx.JSON(http.StatusOK, core.MemberContactReferenceManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-contact-reference/:member_contact_reference_id",
		Method: "DELETE",
		Note:   "Deletes a contact reference entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := helpers.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), invalid member_contact_reference_id: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_contact_reference_id: " + err.Error()})
		}
		value, err := core.MemberContactReferenceManager(service).GetByID(context, *memberContactReferenceID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), record not found: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Contact reference not found: " + err.Error()})
		}
		if err := core.MemberContactReferenceManager(service).Delete(context, *memberContactReferenceID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete contact reference failed (/member-contact-reference/:member_contact_reference_id), db error: " + err.Error(),
				Module:      "MemberContactReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete contact reference: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted contact reference (/member-contact-reference/:member_contact_reference_id): " + value.Name,
			Module:      "MemberContactReference",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
