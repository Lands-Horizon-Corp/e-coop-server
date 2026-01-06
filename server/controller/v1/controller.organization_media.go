package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) organizationMediaController() {
	req := c.provider.Service.Request

	// Get all organization media for a specific organization
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-media/organization/:organization_id",
		Method:       "GET",
		Note:         "Get all organization media for a specific organization.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), invalid organization ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID"})
		}

		// Optional: verify organization exists (using core.OrganizationManager().if needed)
		// For parity with member profile controller we call GetByID
		organization, err := c.core.OrganizationManager().GetByID(context, *organizationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), organization not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}

		organizationMediaList, err := c.core.OrganizationMediaManager().FindRaw(context, &core.OrganizationMedia{
			OrganizationID: organization.ID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "organization-media-search-error",
				Description: "Organization media organization search failed (/organization-media/organization/:organization_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search organization media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "organization-media-search-success",
			Description: "Organization media organization search successful (/organization-media/organization/:organization_id), found " + strconv.Itoa(len(organizationMediaList)) + " media items.",
			Module:      "OrganizationMedia",
		})

		return ctx.JSON(http.StatusOK, organizationMediaList)
	})

	// Create organization media
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-media",
		Method:       "POST",
		Note:         "Creates a new organization media for the current user's organization and branch.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.core.OrganizationMediaManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), user org error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.OrganizationMediaManager().Create(context, organizationMedia); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Organization media created successfully (/organization-media), ID: " + organizationMedia.ID.String(),
			Module:      "OrganizationMedia",
		})

		result, err := c.core.OrganizationMediaManager().GetByID(context, organizationMedia.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created organization media: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	// Update organization media by ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-media/:organization_media_id",
		Method:       "PUT",
		Note:         "Update an organization media by ID.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := handlers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), invalid organization media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		req, err := c.core.OrganizationMediaManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		organizationMedia, err := c.core.OrganizationMediaManager().GetByID(context, *organizationMediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), organization media not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		organizationMedia.Name = req.Name
		organizationMedia.Description = req.Description
		organizationMedia.UpdatedAt = time.Now().UTC()

		if err := c.core.OrganizationMediaManager().UpdateByID(context, organizationMedia.ID, organizationMedia); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:organization_media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Organization media updated successfully (/organization-media/:organization_media_id), ID: " + organizationMediaID.String(),
			Module:      "OrganizationMedia",
		})

		result, err := c.core.OrganizationMediaManager().GetByID(context, *organizationMediaID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated organization media: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// Delete organization media by ID
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/organization-media/:organization_media_id",
		Method: "DELETE",
		Note:   "Delete an organization media by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := handlers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), invalid organization media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		organizationMedia, err := c.core.OrganizationMediaManager().GetByID(context, *organizationMediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		if err := c.core.MediaDelete(context, organizationMedia.MediaID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := c.core.OrganizationMediaManager().Delete(context, organizationMedia.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:organization_media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Organization media deleted successfully (/organization-media/:organization_media_id), ID: " + organizationMediaID.String(),
			Module:      "OrganizationMedia",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Organization media deleted successfully"})
	})

	// Get organization media by ID (raw)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-media/:organization_media_id",
		Method:       "GET",
		Note:         "Get a specific organization media by ID.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		organizationMediaID, err := handlers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		organizationMedia, err := c.core.OrganizationMediaManager().GetByIDRaw(context, *organizationMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		return ctx.JSON(http.StatusOK, organizationMedia)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-media/bulk/organization/:organization_id",
		Method:       "POST",
		Note:         "Bulk create organization media for a specific organization.",
		RequestType:  core.IDSRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID"})
		}

		var req core.IDSRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*core.OrganizationMedia
		for _, mediaID := range req.IDs {
			media, err := c.core.MediaManager().GetByID(context, mediaID)
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

			if err := c.core.OrganizationMediaManager().Create(context, organizationMedia); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
			}

			createdMedia = append(createdMedia, organizationMedia)
		}

		return ctx.JSON(http.StatusCreated, c.core.OrganizationMediaManager().ToModels(createdMedia))
	})
}
