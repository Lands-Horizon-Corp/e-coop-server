package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func organizationMediaController(service *horizon.HorizonService) {
	req := service.API

	// Get all organization media for a specific organization
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-media/organization/:organization_id",
		Method:       "GET",
		Note:         "Get all organization media for a specific organization.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), invalid organization ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID"})
		}

		// Optional: verify organization exists (using core.OrganizationManager(service).if needed)
		// For parity with member profile controller we call GetByID
		organization, err := core.OrganizationManager(service).GetByID(context, *organizationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), organization not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}

		organizationMediaList, err := core.OrganizationMediaManager(service).FindRaw(context, &core.OrganizationMedia{
			OrganizationID: organization.ID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search organization media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "organization-media-search-success",
			Description: "Organization media organization search successful (/organization-media/organization/:organization_id), found " + strconv.Itoa(len(organizationMediaList)) + " media items.",
			Module:      "OrganizationMedia",
		})

		return ctx.JSON(http.StatusOK, organizationMediaList)
	})

	// Create organization media
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-media",
		Method:       "POST",
		Note:         "Creates a new organization media for the current user's organization and branch.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.OrganizationMediaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), user org error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), user not assigned to branch.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		organizationMedia := &core.OrganizationMedia{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.OrganizationMediaManager(service).Create(context, organizationMedia); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Organization media created successfully (/organization-media), ID: " + organizationMedia.ID.String(),
			Module:      "OrganizationMedia",
		})

		result, err := core.OrganizationMediaManager(service).GetByID(context, organizationMedia.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created organization media: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	// Update organization media by ID
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-media/:organization_media_id",
		Method:       "PUT",
		Note:         "Update an organization media by ID.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := helpers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), invalid organization media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		req, err := core.OrganizationMediaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		organizationMedia, err := core.OrganizationMediaManager(service).GetByID(context, *organizationMediaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), organization media not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		organizationMedia.Name = req.Name
		organizationMedia.Description = req.Description
		organizationMedia.UpdatedAt = time.Now().UTC()

		if err := core.OrganizationMediaManager(service).UpdateByID(context, organizationMedia.ID, organizationMedia); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Organization media updated successfully (/organization-media/:organization_media_id), ID: " + organizationMediaID.String(),
			Module:      "OrganizationMedia",
		})

		result, err := core.OrganizationMediaManager(service).GetByID(context, *organizationMediaID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated organization media: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// Delete organization media by ID
	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/organization-media/:organization_media_id",
		Method: "DELETE",
		Note:   "Delete an organization media by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := helpers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), invalid organization media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		organizationMedia, err := core.OrganizationMediaManager(service).GetByID(context, *organizationMediaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		if err := core.MediaDelete(context, service, organizationMedia.MediaID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := core.OrganizationMediaManager(service).Delete(context, organizationMedia.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Organization media deleted successfully (/organization-media/:organization_media_id), ID: " + organizationMediaID.String(),
			Module:      "OrganizationMedia",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Organization media deleted successfully"})
	})

	// Get organization media by ID (raw)
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-media/:organization_media_id",
		Method:       "GET",
		Note:         "Get a specific organization media by ID.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := helpers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		organizationMedia, err := core.OrganizationMediaManager(service).GetByIDRaw(context, *organizationMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		return ctx.JSON(http.StatusOK, organizationMedia)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-media/bulk/organization/:organization_id",
		Method:       "POST",
		Note:         "Bulk create organization media for a specific organization.",
		RequestType:  core.IDSRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID"})
		}

		var req core.IDSRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*core.OrganizationMedia
		for _, mediaID := range req.IDs {
			media, err := core.MediaManager(service).GetByID(context, mediaID)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found: " + mediaID.String()})
			}
			descruption := media.FileName + " at " + time.Now().Format(time.RFC3339)
			organizationMedia := &core.OrganizationMedia{
				MediaID:        mediaID,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: *organizationID,
				Name:           media.FileName,
				Description:    &descruption,
			}

			if err := core.OrganizationMediaManager(service).Create(context, organizationMedia); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
			}

			createdMedia = append(createdMedia, organizationMedia)
		}

		return ctx.JSON(http.StatusCreated, core.OrganizationMediaManager(service).ToModels(createdMedia))
	})
}
