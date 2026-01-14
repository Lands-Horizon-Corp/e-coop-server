package member_profile

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

func MemberProfileMediaController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-media/member-profile/:member_profile_id",
		Method:       "GET",
		Note:         "Get all member profile media for a specific member profile.",
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), invalid member profile ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), member profile not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
		}
		memberProfileMediaList, err := core.MemberProfileMediaManager(service).FindRaw(context, &core.MemberProfileMedia{
			BranchID:        userOrg.BranchID,
			OrganizationID:  &userOrg.OrganizationID,
			MemberProfileID: &memberProfile.ID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "Member profile media member profile search failed (/member-profile-media/member-profile/:member_profile_id/search), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search member profile media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "member-profile-search-success",
			Description: "Member profile media member profile search successful (/member-profile-media/member-profile/:member_profile_id/search), found " + strconv.Itoa(len(memberProfileMediaList)) + " media items.",
			Module:      "MemberProfileMedia",
		})

		return ctx.JSON(http.StatusOK, memberProfileMediaList)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-media",
		Method:       "POST",
		Note:         "Creates a new member profile media for the current user's organization and branch.",
		RequestType:  core.MemberProfileMediaRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := core.MemberProfileMediaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), validation error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), user not assigned to branch.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		memberProfileMedia := &core.MemberProfileMedia{
			MediaID:        req.MediaID,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		}

		if err := core.MemberProfileMediaManager(service).Create(context, memberProfileMedia); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Member profile media creation failed (/member-profile-media), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Member profile media created successfully (/member-profile-media), ID: " + memberProfileMedia.ID.String(),
			Module:      "MemberProfileMedia",
		})

		result, err := core.MemberProfileMediaManager(service).GetByID(context, memberProfileMedia.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created member profile media: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-media/:member_profile_media_id",
		Method:       "PUT",
		Note:         "Update a member profile media by ID.",
		RequestType:  core.MemberProfileMediaRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := helpers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), invalid member profile media ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		req, err := core.MemberProfileMediaManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), validation error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media data: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileMedia, err := core.MemberProfileMediaManager(service).GetByID(context, *memberProfileMediaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), member profile media not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}

		memberProfileMedia.Name = req.Name
		memberProfileMedia.Description = req.Description
		memberProfileMedia.UpdatedAt = time.Now().UTC()
		memberProfileMedia.UpdatedByID = userOrg.UserID

		if err := core.MemberProfileMediaManager(service).UpdateByID(context, memberProfileMedia.ID, memberProfileMedia); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Member profile media update failed (/member-profile-media/:member_profile_media_id), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Member profile media updated successfully (/member-profile-media/:member_profile_media_id), ID: " + memberProfileMediaID.String(),
			Module:      "MemberProfileMedia",
		})

		result, err := core.MemberProfileMediaManager(service).GetByID(context, *memberProfileMediaID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated member profile media: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-profile-media/:member_profile_media_id",
		Method: "DELETE",
		Note:   "Delete a member profile media by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := helpers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), invalid member profile media ID.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		_, err = event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), user org error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileMedia, err := core.MemberProfileMediaManager(service).GetByID(context, *memberProfileMediaID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), not found.",
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}

		if memberProfileMedia.MediaID != nil {
			if err := core.MediaDelete(context, service, *memberProfileMedia.MediaID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
			}

		}

		if err := core.MemberProfileMediaManager(service).Delete(context, memberProfileMedia.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Member profile media delete failed (/member-profile-media/:member_profile_media_id), db error: " + err.Error(),
				Module:      "MemberProfileMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile media: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Member profile media deleted successfully (/member-profile-media/:member_profile_media_id), ID: " + memberProfileMediaID.String(),
			Module:      "MemberProfileMedia",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Member profile media deleted successfully"})
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-media/:member_profile_media_id",
		Method:       "GET",
		Note:         "Get a specific member profile media by ID.",
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileMediaID, err := helpers.EngineUUIDParam(ctx, "member_profile_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile media ID"})
		}

		memberProfileMedia, err := core.MemberProfileMediaManager(service).GetByIDRaw(context, *memberProfileMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile media not found"})
		}

		return ctx.JSON(http.StatusOK, memberProfileMedia)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile-media/bulk/member-profile/:member_profile_id",
		Method:       "POST",
		Note:         "Bulk create member profile media for a specific member profile.",
		RequestType:  core.IDSRequest{},
		ResponseType: core.MemberProfileMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		var req core.IDSRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data: " + err.Error()})
		}

		var createdMedia []*core.MemberProfileMedia
		for _, mediaID := range req.IDs {
			media, err := core.MediaManager(service).GetByID(context, mediaID)
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

			if err := core.MemberProfileMediaManager(service).Create(context, memberProfileMedia); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member profile media: " + err.Error()})
			}

			createdMedia = append(createdMedia, memberProfileMedia)
		}

		return ctx.JSON(http.StatusCreated, core.MemberProfileMediaManager(service).ToModels(createdMedia))
	})
}
