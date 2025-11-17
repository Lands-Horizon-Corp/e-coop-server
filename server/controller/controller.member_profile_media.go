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

// MemberProfileMediaController registers routes for managing member profile media.
func (c *Controller) memberProfileMediaController() {
	req := c.provider.Service.Request

	// GET /api/v1/member-profile-media/member-profile/:member_profile_id/search: Get all media for a specific member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-media/member-profile/:member_profile_id",
		Method:       "GET",
		Note:         "Get all member profile media for a specific member profile.",
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), invalid member profile ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		// Verify member profile belongs to user's organization
		memberProfile, err := c.core.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), member profile not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
		}
		// Search for all member profile media for the specified member profile
		memberProfileMediaList, err := c.core.MemberProfileMediaManager.FindRaw(context, &core.MemberProfileMedia{
			BranchID:        userOrg.BranchID,
			OrganizationID:  &userOrg.OrganizationID,
			MemberProfileID: &memberProfile.ID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search member profile media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "member-profile-search-success",
			Description: "Member profile media member profile search successful (/member-profile-media/member-profile/:member_profile_id/search), found " + strconv.Itoa(len(memberProfileMediaList)) + " media items.",
			Module:      "MemberProfileMedia",
		})

		return ctx.JSON(http.StatusOK, memberProfileMediaList)
	})

	// POST /api/v1/member-profile-media: Create a new member profile media
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-media",
		Method:       "POST",
		Note:         "Creates a new member profile media for the current user's organization and branch.",
		RequestType:  core.MemberProfileMediaRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		reqData, err := c.core.MemberProfileMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), validation error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), user not assigned to branch.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileMedia := &core.MemberProfileMedia{
			MediaID:        reqData.MediaID,
			Name:           reqData.Name,
			Description:    reqData.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		}

		if err := c.core.MemberProfileMediaManager.Create(context, memberProfileMedia); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Member profile media created successfully (/member-profile-media), ID: " + memberProfileMedia.ID.String(),
			Module:      "MemberProfileMedia",
		})

		result, err := c.core.MemberProfileMediaManager.GetByID(context, memberProfileMedia.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created member profile media: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	// PUT /api/v1/member-profile-media/:member_profile_media_id: Update a member profile media
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-media/:member_profile_media_id",
		Method:       "PUT",
		Note:         "Update a member profile media by ID.",
		RequestType:  core.MemberProfileMediaRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := handlers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), invalid member profile media ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		reqData, err := c.core.MemberProfileMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), validation error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media data: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileMedia, err := c.core.MemberProfileMediaManager.GetByID(context, *memberProfileMediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), member profile media not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}
		if memberProfileMedia.MediaID != reqData.MediaID {
			if err := c.core.MediaDelete(context, *memberProfileMedia.MediaID); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
			}
		}

		updateData := &core.MemberProfileMedia{
			MediaID:     reqData.MediaID,
			Name:        reqData.Name,
			Description: reqData.Description,
			UpdatedAt:   time.Now().UTC(),
			UpdatedByID: userOrg.UserID,
		}

		if err := c.core.MemberProfileMediaManager.UpdateByID(context, *memberProfileMediaID, updateData); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Member profile media updated successfully (/member-profile-media/:member_profile_media_id), ID: " + memberProfileMediaID.String(),
			Module:      "MemberProfileMedia",
		})

		result, err := c.core.MemberProfileMediaManager.GetByID(context, *memberProfileMediaID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated member profile media: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// DELETE /api/v1/member-profile-media/:member_profile_media_id: Delete a member profile media
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-profile-media/:member_profile_media_id",
		Method: "DELETE",
		Note:   "Delete a member profile media by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := handlers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), invalid member profile media ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		_, err = c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileMedia, err := c.core.MemberProfileMediaManager.GetByID(context, *memberProfileMediaID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}
		if err := c.core.MediaDelete(context, *memberProfileMedia.MediaID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := c.core.MemberProfileMediaManager.Delete(context, memberProfileMedia.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile media: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Member profile media deleted successfully (/member-profile-media/:member_profile_media_id), ID: " + memberProfileMediaID.String(),
			Module:      "MemberProfileMedia",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Member profile media deleted successfully"})
	})
	// GET /api/v1/member-profile-media/:member_profile_media_id: Get a specific member profile media by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-media/:member_profile_media_id",
		Method:       "GET",
		Note:         "Get a specific member profile media by ID.",
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := handlers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		memberProfileMedia, err := c.core.MemberProfileMediaManager.GetByIDRaw(context, *memberProfileMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}

		return ctx.JSON(http.StatusOK, memberProfileMedia)
	})

	// POST /api/v1/member-profile-media/bulk/member-profile/:member_profile_id: Bulk create member profile media for a specific member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-profile-media/bulk/member-profile/:member_profile_id",
		Method:       "POST",
		Note:         "Bulk create member profile media for a specific member profile.",
		RequestType:  core.MemberProfileBulkMediaRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		var reqData core.MemberProfileBulkMediaRequest
		if err := ctx.Bind(&reqData); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*core.MemberProfileMedia
		for _, mediaID := range *reqData.MediaIDs {
			media, err := c.core.MediaManager.GetByID(context, mediaID)
			if err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found: " + mediaID.String()})
			}
			memberProfileMedia := &core.MemberProfileMedia{
				MediaID:         &mediaID,
				CreatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       time.Now().UTC(),
				UpdatedByID:     userOrg.UserID,
				BranchID:        userOrg.BranchID,
				OrganizationID:  &userOrg.OrganizationID,
				MemberProfileID: memberProfileID,
				Name:            media.FileName,
				Description:     media.FileName + " at " + time.Now().Format(time.RFC3339),
			}

			if err := c.core.MemberProfileMediaManager.Create(context, memberProfileMedia); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile media: " + err.Error()})
			}

			createdMedia = append(createdMedia, memberProfileMedia)
		}

		return ctx.JSON(http.StatusCreated, c.core.MemberProfileMediaManager.ToModels(createdMedia))
	})
}
