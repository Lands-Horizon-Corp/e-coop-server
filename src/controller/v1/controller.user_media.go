package controller_v1

import (
	"net/http"
	"strconv"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

// UserMediaController registers routes for managing user media.
func (c *Controller) UserMediaController() {
	req := c.provider.Service.Request

	// GET /api/v1/user-media/search: Get all media of the current user includes all branches
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/search",
		Method:       "GET",
		Note:         "Get all media of the current user across all branches.",
		ResponseType: []model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "search-error",
				Description: "User media search failed (/user-media/search), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		// Search across all branches for the current user and organization
		userMediaList, err := c.model.UserMediaManager.Find(context, &model.UserMedia{
			UserID: &user.UserID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "search-error",
				Description: "User media search failed (/user-media/search), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "search-success",
			Description: "User media search successful (/user-media/search), found " + strconv.Itoa(len(userMediaList)) + " media items.",
			Module:      "UserMedia",
		})

		return ctx.JSON(http.StatusOK, c.model.UserMediaManager.Pagination(context, ctx, userMediaList))
	})

	// GET /api/v1/user-media/current/search: Get all media of the current user of specific branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/current/search",
		Method:       "GET",
		Note:         "Get all media of the current user for their current branch.",
		ResponseType: []model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "current-search-error",
				Description: "User media current search failed (/user-media/current/search), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "current-search-error",
				Description: "User media current search failed (/user-media/current/search), user not assigned to branch.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		// Search for current user's media in their current branch
		userMediaList, err := c.model.UserMediaManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "user_medias.created_by_id", Op: horizon_services.OpEq, Value: user.UserID},
			{Field: "user_medias.organization_id", Op: horizon_services.OpEq, Value: &user.OrganizationID},
			{Field: "user_medias.branch_id", Op: horizon_services.OpEq, Value: user.BranchID},
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "current-search-error",
				Description: "User media current search failed (/user-media/current/search), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "current-search-success",
			Description: "User media current search successful (/user-media/current/search), found " + strconv.Itoa(len(userMediaList)) + " media items.",
			Module:      "UserMedia",
		})

		return ctx.JSON(http.StatusOK, c.model.UserMediaManager.Pagination(context, ctx, userMediaList))
	})

	// GET /api/v1/user-media/branch/:branch_id/search: Get all media of all users from the branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/branch/:branch_id/search",
		Method:       "GET",
		Note:         "Get all user media from a specific branch.",
		ResponseType: []model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "branch-search-error",
				Description: "User media branch search failed (/user-media/branch/:branch_id/search), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "branch-search-error",
				Description: "User media branch search failed (/user-media/branch/:branch_id/search), invalid branch ID.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID"})
		}

		// Verify branch belongs to user's organization
		branch, err := c.model.BranchManager.GetByID(context, *branchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "branch-search-error",
				Description: "User media branch search failed (/user-media/branch/:branch_id/search), branch not found.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found"})
		}

		if branch.OrganizationID != user.OrganizationID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "branch-search-error",
				Description: "User media branch search failed (/user-media/branch/:branch_id/search), branch access denied.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this branch"})
		}

		// Search for all user media in the specified branch
		userMediaList, err := c.model.UserMediaManager.FindWithFilters(context, []horizon_services.Filter{
			{Field: "user_medias.organization_id", Op: horizon_services.OpEq, Value: &user.OrganizationID},
			{Field: "user_medias.branch_id", Op: horizon_services.OpEq, Value: branchID},
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "branch-search-error",
				Description: "User media branch search failed (/user-media/branch/:branch_id/search), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "branch-search-success",
			Description: "User media branch search successful (/user-media/branch/:branch_id/search), found " + strconv.Itoa(len(userMediaList)) + " media items.",
			Module:      "UserMedia",
		})

		return ctx.JSON(http.StatusOK, c.model.UserMediaManager.Pagination(context, ctx, userMediaList))
	})

	// GET /api/v1/user-media/member-profile/:member_profile_id/search: Get all media for a specific member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/member-profile/:member_profile_id/search",
		Method:       "GET",
		Note:         "Get all user media for a specific member profile.",
		ResponseType: []model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "User media member profile search failed (/user-media/member-profile/:member_profile_id/search), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "User media member profile search failed (/user-media/member-profile/:member_profile_id/search), invalid member profile ID.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}

		// Verify member profile belongs to user's organization
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "User media member profile search failed (/user-media/member-profile/:member_profile_id/search), member profile not found.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found"})
		}

		if memberProfile.OrganizationID != user.OrganizationID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "User media member profile search failed (/user-media/member-profile/:member_profile_id/search), member profile access denied.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this member profile"})
		}

		// Search for all user media for the specified member profile
		userMediaList, err := c.model.UserMediaManager.Find(context, &model.UserMedia{
			BranchID:       user.BranchID,
			OrganizationID: &user.OrganizationID,
			UserID:         memberProfile.UserID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "member-profile-search-error",
				Description: "User media member profile search failed (/user-media/member-profile/:member_profile_id/search), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "member-profile-search-success",
			Description: "User media member profile search successful (/user-media/member-profile/:member_profile_id/search), found " + strconv.Itoa(len(userMediaList)) + " media items.",
			Module:      "UserMedia",
		})

		return ctx.JSON(http.StatusOK, c.model.UserMediaManager.Pagination(context, ctx, userMediaList))
	})

	// GET /api/v1/user-media/:user_media_id: Get a specific user media by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/:user_media_id",
		Method:       "GET",
		Note:         "Get a specific user media by ID.",
		ResponseType: model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userMediaID, err := handlers.EngineUUIDParam(ctx, "user_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media ID"})
		}

		userMedia, err := c.model.UserMediaManager.GetByIDRaw(context, *userMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User media not found"})
		}

		return ctx.JSON(http.StatusOK, userMedia)
	})

	// POST /api/v1/user-media: Create a new user media
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media",
		Method:       "POST",
		Note:         "Creates a new user media for the current user's organization and branch.",
		RequestType:  model.UserMediaRequest{},
		ResponseType: model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		reqData, err := c.model.UserMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "User media creation failed (/user-media), validation error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media data: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "User media creation failed (/user-media), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "User media creation failed (/user-media), user not assigned to branch.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		userMedia := &model.UserMedia{
			MediaID:        reqData.MediaID,
			Name:           reqData.Name,
			Description:    reqData.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       user.BranchID,
			OrganizationID: &user.OrganizationID,
		}

		if err := c.model.UserMediaManager.Create(context, userMedia); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "User media creation failed (/user-media), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "User media created successfully (/user-media), ID: " + userMedia.ID.String(),
			Module:      "UserMedia",
		})

		result, err := c.model.UserMediaManager.GetByID(context, userMedia.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve created user media: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, result)
	})

	// PUT /api/v1/user-media/:user_media_id: Update a user media
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/:user_media_id",
		Method:       "PUT",
		Note:         "Update a user media by ID.",
		RequestType:  model.UserMediaRequest{},
		ResponseType: model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userMediaID, err := handlers.EngineUUIDParam(ctx, "user_media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "User media update failed (/user-media/:user_media_id), invalid user media ID.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media ID"})
		}

		reqData, err := c.model.UserMediaManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "User media update failed (/user-media/:user_media_id), validation error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media data: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "User media update failed (/user-media/:user_media_id), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		_, err = c.model.UserMediaManager.GetByID(context, *userMediaID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "User media update failed (/user-media/:user_media_id), user media not found.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User media not found"})
		}

		updateData := &model.UserMedia{
			MediaID:     reqData.MediaID,
			Name:        reqData.Name,
			Description: reqData.Description,
			UpdatedAt:   time.Now().UTC(),
			UpdatedByID: user.UserID,
		}

		if err := c.model.UserMediaManager.UpdateByID(context, *userMediaID, updateData); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "User media update failed (/user-media/:user_media_id), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "User media updated successfully (/user-media/:user_media_id), ID: " + userMediaID.String(),
			Module:      "UserMedia",
		})

		result, err := c.model.UserMediaManager.GetByID(context, *userMediaID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated user media: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// DELETE /api/v1/user-media/:user_media_id: Delete a user media
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/:user_media_id",
		Method:       "DELETE",
		Note:         "Delete a user media by ID.",
		ResponseType: map[string]string{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userMediaID, err := handlers.EngineUUIDParam(ctx, "user_media_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "User media delete failed (/user-media/:user_media_id), invalid user media ID.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media ID"})
		}

		_, err = c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "User media delete failed (/user-media/:user_media_id), user org error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		userMedia, err := c.model.UserMediaManager.GetByID(context, *userMediaID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "User media delete failed (/user-media/:user_media_id), not found.",
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User media not found"})
		}
		if err := c.model.MediaDelete(context, *userMedia.MediaID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
				Module:      "Media",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
		}

		if err := c.model.UserMediaManager.DeleteByID(context, userMedia.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "User media delete failed (/user-media/:user_media_id), db error: " + err.Error(),
				Module:      "UserMedia",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user media: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User media deleted successfully (/user-media/:user_media_id), ID: " + userMediaID.String(),
			Module:      "UserMedia",
		})

		return ctx.JSON(http.StatusOK, map[string]string{"message": "User media deleted successfully"})
	})
	// GET /api/v1/user-media/:user_media_id: Get a specific user media by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-media/:user_media_id",
		Method:       "GET",
		Note:         "Get a specific user media by ID.",
		ResponseType: model.UserMediaResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userMediaID, err := handlers.EngineUUIDParam(ctx, "user_media_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user media ID"})
		}

		userMedia, err := c.model.UserMediaManager.GetByIDRaw(context, *userMediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User media not found"})
		}

		return ctx.JSON(http.StatusOK, userMedia)
	})
}
