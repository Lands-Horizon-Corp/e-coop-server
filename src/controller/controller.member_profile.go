package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberProfileController() {
	req := c.provider.Service.Request

	// Get all pending member profiles in the current branch
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/pending",
		Method:       "GET",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Returns all pending member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := c.model.MemberProfileManager.Find(context, &model.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Status:         "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get pending member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.Filtered(context, ctx, memberProfile))
	})

	// Quickly create a new user account and link it to a member profile by ID
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/user-account",
		Method:       "POST",
		RequestType:  model.MemberProfileUserAccountRequest{},
		ResponseType: model.MemberProfileResponse{},
		Note:         "Links a minimal user account to a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		var req model.MemberProfileUserAccountRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.Password)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: hash password error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
		}
		userProfile := &model.User{
			Email:             req.Email,
			UserName:          req.UserName,
			ContactNumber:     req.ContactNumber,
			Password:          hashedPwd,
			FullName:          req.FullName,
			FirstName:         &req.FirstName,
			MiddleName:        &req.MiddleName,
			LastName:          &req.LastName,
			Suffix:            &req.Suffix,
			IsEmailVerified:   false,
			IsContactVerified: false,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		if err := c.model.UserManager.CreateWithTx(context, tx, userProfile); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: create user error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create user profile: " + err.Error()})
		}
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: database error: " + tx.Error.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error: " + tx.Error.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.UserID = &userProfile.ID
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = userOrg.UserID

		if err := c.model.MemberProfileManager.UpdateFieldsWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: update member profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userProfile.ID.String())
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: generate developer key error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		newUserOrg := &model.UserOrganization{
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             userOrg.UserID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             userOrg.UserID,
			OrganizationID:          userOrg.OrganizationID,
			BranchID:                userOrg.BranchID,
			UserID:                  userProfile.ID,
			UserType:                "member",
			Description:             "",
			ApplicationDescription:  "anything",
			ApplicationStatus:       "accepted",
			DeveloperSecretKey:      developerKey,
			PermissionName:          "member",
			PermissionDescription:   "",
			Permissions:             []string{},
			UserSettingDescription:  "user settings",
			UserSettingStartOR:      0,
			UserSettingEndOR:        0,
			UserSettingUsedOR:       0,
			UserSettingStartVoucher: 0,
			UserSettingEndVoucher:   0,
			UserSettingUsedVoucher:  0,
		}
		if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: create user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create UserOrganization: " + err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created user account for member profile: " + userProfile.UserName,
			Module:      "MemberProfile",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	// Approve a member profile by ID
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/approve",
		Method:       "PUT",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Approve a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: user not authorized",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.Status = "verified"
		memberProfile.MemberVerifiedByEmployeeUserID = &userOrg.UserID
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Approved member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	// Reject a member profile by ID
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/reject",
		Method:       "PUT",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Reject a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: user not authorized",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.Status = "not allowed"
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "reject-success",
			Description: "Rejected member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	// Retrieve a list of all member profiles in the current branch
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile",
		Method:       "GET",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Returns all member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberProfile, err := c.model.MemberProfileCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.Filtered(context, ctx, memberProfile))
	})

	// Retrieve paginated member profiles for the current branch
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/search",
		Method:       "GET",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Returns paginated member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.MemberProfileCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.Pagination(context, ctx, value))
	})

	// Retrieve a specific member profile by member_profile_id
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: model.MemberProfileResponse{},
		Note:         "Returns a specific member profile by its member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByIDRaw(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		return ctx.JSON(http.StatusOK, memberProfile)
	})

	// Delete a specific member profile by its member_profile_id
	req.RegisterRoute(horizon.Route{
		Route:  "/member-profile/:member_profile_id",
		Method: "DELETE",
		Note:   "Deletes a specific member profile and all its connections by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: begin tx error: " + tx.Error.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID.String(), err)})
		}
		if err := c.model.MemberProfileDestroy(context, tx, memberProfile.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: destroy error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member profiles by IDs
	req.RegisterRoute(horizon.Route{
		Route:       "/member-profile/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member profiles and all their connections by their IDs.",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed: invalid request body.",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed: no IDs provided.",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed: begin tx error: " + tx.Error.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			if rawID == "" {
				continue
			}
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member profiles failed: invalid UUID: " + rawID,
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}
			memberProfile, err := c.model.MemberProfileManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member profiles failed: member profile not found: " + rawID,
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", rawID, err)})
			}
			names += memberProfile.FullName + ","
			if err := c.model.MemberProfileDestroy(context, tx, memberProfile.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member profiles failed: destroy error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member profile with ID %s: %v", rawID, err)})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member profiles: " + names,
			Module:      "MemberProfile",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Connect the specified member profile to a user account
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/connect-user",
		Method:       "POST",
		RequestType:  model.MemberProfileAccountRequest{},
		ResponseType: model.MemberProfileResponse{},
		Note:         "Connects the specified member profile to a user account by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfileAccountRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileId, err)})
		}
		memberProfile.UserID = req.UserID
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Connected member profile (%s) to user account.", memberProfile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})
	// Quickly create a new member profile with minimal required fields
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/quick-create",
		Method:       "POST",
		RequestType:  model.MemberProfileQuickCreateRequest{},
		ResponseType: model.MemberProfileResponse{},
		Note:         "Quickly creates a new member profile with minimal required fields.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfileQuickCreateRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()

		var userProfile *model.User
		var userProfileID *uuid.UUID

		if req.AccountInfo != nil {
			hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.AccountInfo.Password)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: hash password error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + err.Error()})
			}
			userProfile = &model.User{
				Email:             req.AccountInfo.Email,
				UserName:          req.AccountInfo.UserName,
				ContactNumber:     req.ContactNumber,
				Password:          hashedPwd,
				FullName:          req.FullName,
				FirstName:         &req.FirstName,
				MiddleName:        &req.MiddleName,
				LastName:          &req.LastName,
				Suffix:            &req.Suffix,
				IsEmailVerified:   false,
				IsContactVerified: false,
				CreatedAt:         time.Now().UTC(),
				UpdatedAt:         time.Now().UTC(),
			}
			if err := c.model.UserManager.CreateWithTx(context, tx, userProfile); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: create user error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create user profile: " + err.Error()})
			}
			if tx.Error != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: database error: " + tx.Error.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error: " + tx.Error.Error()})
			}
			userProfileID = &userProfile.ID
		}

		profile := &model.MemberProfile{
			OrganizationID:       user.OrganizationID,
			BranchID:             *user.BranchID,
			CreatedAt:            time.Now().UTC(),
			UpdatedAt:            time.Now().UTC(),
			CreatedByID:          user.UserID,
			UpdatedByID:          user.UserID,
			UserID:               userProfileID,
			OldReferenceID:       req.OldReferenceID,
			Passbook:             req.Passbook,
			FirstName:            req.FirstName,
			MiddleName:           req.MiddleName,
			LastName:             req.LastName,
			FullName:             req.FullName,
			Suffix:               req.Suffix,
			MemberGenderID:       req.MemberGenderID,
			BirthDate:            req.BirthDate,
			ContactNumber:        req.ContactNumber,
			CivilStatus:          req.CivilStatus,
			MemberOccupationID:   req.MemberOccupationID,
			Status:               req.Status,
			IsMutualFundMember:   req.IsMutualFundMember,
			IsMicroFinanceMember: req.IsMicroFinanceMember,
			MemberTypeID:         req.MemberTypeID,
		}
		if err := c.model.MemberProfileManager.CreateWithTx(context, tx, profile); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: create profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create member profile: " + err.Error()})
		}

		if userProfile != nil {
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: generate developer key error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
			}
			developerKey = developerKey + uuid.NewString() + "-horizon"
			userOrg := &model.UserOrganization{
				CreatedAt:               time.Now().UTC(),
				CreatedByID:             user.UserID,
				UpdatedAt:               time.Now().UTC(),
				UpdatedByID:             user.UserID,
				OrganizationID:          user.OrganizationID,
				BranchID:                user.BranchID,
				UserID:                  *userProfileID,
				UserType:                "member",
				Description:             "",
				ApplicationDescription:  "anything",
				ApplicationStatus:       "accepted",
				DeveloperSecretKey:      developerKey,
				PermissionName:          "member",
				PermissionDescription:   "",
				Permissions:             []string{},
				UserSettingDescription:  "user settings",
				UserSettingStartOR:      0,
				UserSettingEndOR:        0,
				UserSettingUsedOR:       0,
				UserSettingStartVoucher: 0,
				UserSettingEndVoucher:   0,
				UserSettingUsedVoucher:  0,
			}
			if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, userOrg); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: create user org error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create UserOrganization: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Quick created member profile: " + profile.FullName,
			Module:      "MemberProfile",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})

	// Update the personal information of a member profile by ID
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/personal-info",
		Method:       "PUT",
		RequestType:  model.MemberProfilePersonalInfoRequest{},
		ResponseType: model.MemberProfileResponse{},
		Note:         "Updates the personal information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfilePersonalInfoRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		profile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileId, err)})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = userOrg.UserID
		profile.FirstName = req.FirstName
		profile.MiddleName = req.MiddleName
		profile.LastName = req.LastName
		profile.FullName = req.FullName
		profile.Suffix = req.Suffix
		profile.BirthDate = req.BirthDate
		profile.ContactNumber = req.ContactNumber
		profile.CivilStatus = req.CivilStatus

		if req.MemberGenderID != nil && !uuidPtrEqual(profile.MemberGenderID, req.MemberGenderID) {
			data := &model.MemberGenderHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileId,
				MemberGenderID:  *req.MemberGenderID,
			}
			if err := c.model.MemberGenderHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile personal info failed: update gender history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member gender history: " + err.Error()})
			}
			profile.MemberGenderID = req.MemberGenderID
		}
		if req.MemberOccupationID != nil && !uuidPtrEqual(profile.MemberOccupationID, req.MemberOccupationID) {
			data := &model.MemberOccupationHistory{
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				CreatedAt:          time.Now().UTC(),
				UpdatedAt:          time.Now().UTC(),
				CreatedByID:        userOrg.UserID,
				UpdatedByID:        userOrg.UserID,
				MemberProfileID:    *memberProfileId,
				MemberOccupationID: *req.MemberOccupationID,
			}
			if err := c.model.MemberOccupationHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile personal info failed: update occupation history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member occupation history: " + err.Error()})
			}
			profile.MemberOccupationID = req.MemberOccupationID
		}

		profile.BusinessAddress = req.BusinessAddress
		profile.BusinessContactNumber = req.BusinessContactNumber
		profile.Notes = req.Notes
		profile.Description = req.Description
		profile.MediaID = req.MediaID
		profile.SignatureMediaID = req.SignatureMediaID
		if err := c.model.MemberProfileManager.UpdateFields(context, profile.ID, profile); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile personal info: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})

	// Update the membership information of a member profile by ID
	req.RegisterRoute(horizon.Route{
		Route:        "/member-profile/:member_profile_id/membership-info",
		Method:       "PUT",
		RequestType:  model.MemberProfileMembershipInfoRequest{},
		ResponseType: model.MemberProfileResponse{},
		Note:         "Updates the membership information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfileMembershipInfoRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		profile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileId, err)})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = userOrg.UserID
		profile.Passbook = req.Passbook
		profile.OldReferenceID = req.OldReferenceID
		profile.RecruitedByMemberProfileID = req.RecruitedByMemberProfileID
		profile.Status = req.Status

		if req.MemberTypeID != nil && !uuidPtrEqual(profile.MemberTypeID, req.MemberTypeID) {
			data := &model.MemberTypeHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileId,
				MemberTypeID:    *req.MemberTypeID,
			}
			if err := c.model.MemberTypeHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member type history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member type history: " + err.Error()})
			}
			profile.MemberTypeID = req.MemberTypeID
		}
		if req.MemberGroupID != nil && !uuidPtrEqual(profile.MemberGroupID, req.MemberGroupID) {
			data := &model.MemberGroupHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileId,
				MemberGroupID:   *req.MemberGroupID,
			}
			if err := c.model.MemberGroupHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member group history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member group history: " + err.Error()})
			}
			profile.MemberGroupID = req.MemberGroupID
		}
		if req.MemberClassificationID != nil && !uuidPtrEqual(profile.MemberClassificationID, req.MemberClassificationID) {
			data := &model.MemberClassificationHistory{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CreatedAt:              time.Now().UTC(),
				UpdatedAt:              time.Now().UTC(),
				CreatedByID:            userOrg.UserID,
				UpdatedByID:            userOrg.UserID,
				MemberProfileID:        *memberProfileId,
				MemberClassificationID: *req.MemberClassificationID,
			}
			if err := c.model.MemberClassificationHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member classification history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member classification history: " + err.Error()})
			}
			profile.MemberClassificationID = req.MemberClassificationID
		}
		if req.MemberCenterID != nil && !uuidPtrEqual(profile.MemberCenterID, req.MemberCenterID) {
			data := &model.MemberCenterHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileId,
				MemberCenterID:  *req.MemberCenterID,
			}
			if err := c.model.MemberCenterHistoryManager.Create(context, data); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member center history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member center history: " + err.Error()})
			}
			profile.MemberCenterID = req.MemberCenterID
		}

		profile.IsMutualFundMember = req.IsMutualFundMember
		profile.IsMicroFinanceMember = req.IsMicroFinanceMember

		if err := c.model.MemberProfileManager.UpdateFields(context, profile.ID, profile); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile membership info: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})
}
