package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// OrganizationMediaController registers routes for managing organization media.
func (c *Controller) organizationMediaController() {
	req := c.provider.Service.Request

	// GET /organization-media: List all organization media for the current user's organization. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media",
		Method:       "GET",
		Note:         "Returns all organization media for the current user's organization. Returns empty if not authenticated.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		organizationMedia, err := c.core.OrganizationMediaFindByOrganization(context, user.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No organization media found for the current organization"})
		}
		return ctx.JSON(http.StatusOK, c.core.OrganizationMediaManager.Filtered(context, ctx, organizationMedia))
	})

	// GET /organization-media/search: Paginated search of organization media for the current organization. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media/search",
		Method:       "GET",
		Note:         "Returns a paginated list of organization media for the current user's organization.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		organizationMedia, err := c.core.OrganizationMediaFindByOrganization(context, user.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch organization media for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.OrganizationMediaManager.Pagination(context, ctx, organizationMedia))
	})

	// GET /organization-media/organization/:organization_id: Get all media for a specific organization by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media/organization/:organization_id",
		Method:       "GET",
		Note:         "Returns all organization media for a specific organization by its ID.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID"})
		}
		organizationMedias, err := c.core.OrganizationMediaManager.FindRaw(context, &core.OrganizationMedia{
			OrganizationID: *organizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No organization media found for the specified organization"})
		}
		return ctx.JSON(http.StatusOK, organizationMedias)
	})

	// GET /organization-media/:media_id: Get specific organization media by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media/:media_id",
		Method:       "GET",
		Note:         "Returns a single organization media by its ID.",
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaID, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}
		organizationMedia, err := c.core.OrganizationMediaManager.GetByIDRaw(context, *mediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}
		return ctx.JSON(http.StatusOK, organizationMedia)
	})

	// POST /organization-media: Create a new organization media. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media",
		Method:       "POST",
		Note:         "Creates a new organization media for the current user's organization.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.OrganizationMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		organizationMedia := &core.OrganizationMedia{
			Name:           req.Name,
			Description:    req.Description,
			OrganizationID: req.OrganizationID,
			MediaID:        req.MediaID,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		if err := c.core.OrganizationMediaManager.Create(context, organizationMedia); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Organization media creation failed (/organization-media), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created organization media (/organization-media): " + organizationMedia.Name,
			Module:      "OrganizationMedia",
		})
		return ctx.JSON(http.StatusCreated, c.core.OrganizationMediaManager.ToModel(organizationMedia))
	})

	// PUT /organization-media/: Update organization media by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-media/:organization_media_id",
		Method:       "PUT",
		Note:         "Updates an existing organization media by its ID.",
		RequestType:  core.OrganizationMediaRequest{},
		ResponseType: core.OrganizationMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationMeiaID, err := handlers.EngineUUIDParam(ctx, "organization_media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:media_id), invalid media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		req, err := c.core.OrganizationMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:media_id), validation error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media data: " + err.Error()})
		}

		organizationMedia, err := c.core.OrganizationMediaManager.GetByID(context, *organizationMeiaID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:media_id), media not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		organizationMedia.Name = req.Name
		organizationMedia.Description = req.Description
		organizationMedia.MediaID = req.MediaID
		organizationMedia.UpdatedAt = time.Now().UTC()

		if err := c.core.OrganizationMediaManager.UpdateByID(context, organizationMedia.ID, organizationMedia); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Organization media update failed (/organization-media/:media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization media: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated organization media (/organization-media/:media_id): " + organizationMedia.Name,
			Module:      "OrganizationMedia",
		})
		return ctx.JSON(http.StatusOK, c.core.OrganizationMediaManager.ToModel(organizationMedia))
	})

	// DELETE /organization-media/:media_id: Delete an organization media by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/organization-media/:media_id",
		Method: "DELETE",
		Note:   "Deletes the specified organization media by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		mediaID, err := handlers.EngineUUIDParam(ctx, "media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:media_id), invalid media ID.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization media ID"})
		}

		organizationMedia, err := c.core.OrganizationMediaManager.GetByID(context, *mediaID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:media_id), not found.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization media not found"})
		}

		if err := c.core.OrganizationMediaManager.Delete(context, *mediaID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Organization media delete failed (/organization-media/:media_id), db error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization media: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted organization media (/organization-media/:media_id): " + organizationMedia.Name,
			Module:      "OrganizationMedia",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /organization-media/bulk-delete: Bulk delete organization media by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/organization-media/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple organization media by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/organization-media/bulk-delete), invalid request body.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/organization-media/bulk-delete), no IDs provided.",
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No organization media IDs provided for bulk delete"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/organization-media/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		var namesSlice []string
		for _, rawID := range reqBody.IDs {
			mediaID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/organization-media/bulk-delete), invalid UUID: " + rawID,
					Module:      "OrganizationMedia",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			organizationMedia, err := c.core.OrganizationMediaManager.GetByID(context, mediaID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/organization-media/bulk-delete), not found: " + rawID,
					Module:      "OrganizationMedia",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Organization media not found with ID: %s", rawID)})
			}

			namesSlice = append(namesSlice, organizationMedia.Name)
			if err := c.core.OrganizationMediaManager.DeleteWithTx(context, tx, mediaID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/organization-media/bulk-delete), db error: " + err.Error(),
					Module:      "OrganizationMedia",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization media: " + err.Error()})
			}
		}
		names := strings.Join(namesSlice, ",")

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/organization-media/bulk-delete), commit error: " + err.Error(),
				Module:      "OrganizationMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted organization media (/organization-media/bulk-delete): " + names,
			Module:      "OrganizationMedia",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
