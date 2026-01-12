package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func memberProfileController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/pending",
		Method:       "GET",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Returns all pending member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := core.MemberProfileManager(service).Find(context, &core.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Status:         core.MemberStatusPending,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get pending member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModels(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/user-account",
		Method:       "POST",
		RequestType:  core.MemberProfileUserAccountRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Links a minimal user account to a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		var req core.MemberProfileUserAccountRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		hashedPwd, err := service.Security.HashPassword(req.Password)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: hash password error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + endTx(err).Error()})
		}
		userProfile := &core.User{
			Email:             req.Email,
			Username:          req.Username,
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
			Birthdate:         req.BirthDate,
		}
		if err := core.UserManager(service).CreateWithTx(context, tx, userProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: create user error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create user profile: " + endTx(err).Error()})
		}
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: database error: " + tx.Error.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error: " + endTx(tx.Error).Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, endTx(err))})
		}
		memberProfile.UserID = &userProfile.ID
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = &userOrg.UserID

		if err := core.MemberProfileManager(service).UpdateByIDWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: update member profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + endTx(err).Error()})
		}

		developerKey, err := service.Security.GenerateUUIDv5(userProfile.ID.String())
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: generate developer key error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + endTx(err).Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		newUserOrg := &core.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              userOrg.UserID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              userOrg.UserID,
			OrganizationID:           userOrg.OrganizationID,
			BranchID:                 userOrg.BranchID,
			UserID:                   userProfile.ID,
			UserType:                 core.UserOrganizationTypeMember,
			Description:              "",
			ApplicationDescription:   "anything",
			ApplicationStatus:        "accepted",
			DeveloperSecretKey:       developerKey,
			PermissionName:           string(core.UserOrganizationTypeMember),
			PermissionDescription:    "",
			Permissions:              []string{},
			UserSettingDescription:   "user settings",
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}
		if err := core.UserOrganizationManager(service).CreateWithTx(context, tx, newUserOrg); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: create user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create UserOrganization: " + endTx(err).Error()})
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create user account for member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created user account for member profile: " + userProfile.Username,
			Module:      "MemberProfile",
		})

		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/approve",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Approve a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: user not authorized",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.Status = "verified"
		memberProfile.MemberVerifiedByEmployeeUserID = &userOrg.UserID
		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Approved member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/reject",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Reject a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: user not authorized",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.Status = "not allowed"
		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "reject-error",
				Description: "Reject member profile failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "reject-success",
			Description: "Rejected member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile",
		Method:       "GET",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Returns all member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModels(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/search",
		Method:       "GET",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Returns paginated member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberProfileManager(service).NormalPagination(context, ctx, &core.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Returns a specific member profile by its member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByIDRaw(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		return ctx.JSON(http.StatusOK, memberProfile)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-profile/:member_profile_id",
		Method: "DELETE",
		Note:   "Deletes a specific member profile and all its connections by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: begin tx error: " + tx.Error.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + endTx(tx.Error).Error()})
		}

		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID.String(), endTx(err))})
		}
		if err := core.MemberProfileDestroy(context, service, tx, memberProfile.ID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: destroy error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member profile: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-profile/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member profiles and all their connections by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed (/member-profile/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed (/member-profile/bulk-delete) | no IDs provided",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberProfileManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member profiles failed (/member-profile/bulk-delete) | error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member profiles: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member profiles (/member-profile/bulk-delete)",
			Module:      "MemberProfile",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect-user",
		Method:       "POST",
		RequestType:  core.MemberProfileAccountRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Connects the specified member profile to a user account by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfileAccountRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		memberProfile.UserID = req.UserID
		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect member profile to user account failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Connected member profile (%s) to user account.", memberProfile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/quick-create",
		Method:       "POST",
		RequestType:  core.MemberProfileQuickCreateRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Quickly creates a new member profile with minimal required fields.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfileQuickCreateRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)

		var userProfile *core.User
		var userProfileID *uuid.UUID

		if req.AccountInfo != nil {
			hashedPwd, err := service.Security.HashPassword(req.AccountInfo.Password)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: hash password error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password: " + endTx(err).Error()})
			}
			userProfile = &core.User{
				Email:             req.AccountInfo.Email,
				Username:          req.AccountInfo.Username,
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
				Birthdate:         req.BirthDate,
			}
			if err := core.UserManager(service).CreateWithTx(context, tx, userProfile); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: create user error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create user profile: " + endTx(err).Error()})
			}
			if tx.Error != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: database error: " + tx.Error.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error: " + endTx(tx.Error).Error()})
			}
			userProfileID = &userProfile.ID
		}

		profile := &core.MemberProfile{
			OrganizationID:       userOrg.OrganizationID,
			BranchID:             *userOrg.BranchID,
			CreatedAt:            time.Now().UTC(),
			UpdatedAt:            time.Now().UTC(),
			CreatedByID:          &userOrg.UserID,
			UpdatedByID:          &userOrg.UserID,
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
			BirthPlace:           req.BirthPlace,
		}
		if err := core.MemberProfileManager(service).CreateWithTx(context, tx, profile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: create profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not create member profile: " + endTx(err).Error()})
		}

		if userProfile != nil {
			developerKey, err := service.Security.GenerateUUIDv5(userOrg.ID.String())
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: generate developer key error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + endTx(err).Error()})
			}
			developerKey = developerKey + uuid.NewString() + "-horizon"
			userOrg := &core.UserOrganization{
				CreatedAt:                time.Now().UTC(),
				CreatedByID:              userOrg.UserID,
				UpdatedAt:                time.Now().UTC(),
				UpdatedByID:              userOrg.UserID,
				OrganizationID:           userOrg.OrganizationID,
				BranchID:                 userOrg.BranchID,
				UserID:                   *userProfileID,
				UserType:                 core.UserOrganizationTypeMember,
				Description:              "",
				ApplicationDescription:   "anything",
				ApplicationStatus:        "accepted",
				DeveloperSecretKey:       developerKey,
				PermissionName:           string(core.UserOrganizationTypeMember),
				PermissionDescription:    "",
				Permissions:              []string{},
				UserSettingDescription:   "user settings",
				UserSettingStartOR:       0,
				UserSettingEndOR:         1000,
				UserSettingUsedOR:        0,
				UserSettingStartVoucher:  0,
				UserSettingEndVoucher:    0,
				UserSettingUsedVoucher:   0,
				UserSettingNumberPadding: 7,
			}
			if err := core.UserOrganizationManager(service).CreateWithTx(context, tx, userOrg); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: create user org error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create UserOrganization: " + endTx(err).Error()})
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Quick created member profile: " + profile.FullName,
			Module:      "MemberProfile",
		})

		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(profile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/personal-info",
		Method:       "PUT",
		RequestType:  core.MemberProfilePersonalInfoRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Updates the personal information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfilePersonalInfoRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		profile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = &userOrg.UserID
		profile.FirstName = req.FirstName
		profile.MiddleName = req.MiddleName
		profile.LastName = req.LastName
		profile.FullName = req.FullName
		profile.Suffix = req.Suffix
		profile.BirthDate = req.BirthDate
		profile.BirthPlace = req.BirthPlace
		profile.ContactNumber = req.ContactNumber
		profile.CivilStatus = req.CivilStatus

		if req.MemberGenderID != nil && !helpers.UUIDPtrEqual(profile.MemberGenderID, req.MemberGenderID) {
			data := &core.MemberGenderHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileID,
				MemberGenderID:  *req.MemberGenderID,
			}
			if err := core.MemberGenderHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile personal info failed: update gender history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member gender history: " + err.Error()})
			}
			profile.MemberGenderID = req.MemberGenderID
		}
		if req.MemberOccupationID != nil && !helpers.UUIDPtrEqual(profile.MemberOccupationID, req.MemberOccupationID) {
			data := &core.MemberOccupationHistory{
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				CreatedAt:          time.Now().UTC(),
				UpdatedAt:          time.Now().UTC(),
				CreatedByID:        userOrg.UserID,
				UpdatedByID:        userOrg.UserID,
				MemberProfileID:    *memberProfileID,
				MemberOccupationID: *req.MemberOccupationID,
			}
			if err := core.MemberOccupationHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.MemberProfileManager(service).UpdateByID(context, profile.ID, profile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile personal info failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile personal info: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(profile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/membership-info",
		Method:       "PUT",
		RequestType:  core.MemberProfileMembershipInfoRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Updates the membership information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfileMembershipInfoRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		profile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = &userOrg.UserID
		profile.Passbook = req.Passbook
		profile.OldReferenceID = req.OldReferenceID
		profile.RecruitedByMemberProfileID = req.RecruitedByMemberProfileID
		profile.Status = req.Status

		if req.MemberDepartmentID != nil && !helpers.UUIDPtrEqual(profile.MemberDepartmentID, req.MemberDepartmentID) {
			data := &core.MemberDepartmentHistory{
				OrganizationID:     userOrg.OrganizationID,
				BranchID:           *userOrg.BranchID,
				CreatedAt:          time.Now().UTC(),
				UpdatedAt:          time.Now().UTC(),
				CreatedByID:        userOrg.UserID,
				UpdatedByID:        userOrg.UserID,
				MemberProfileID:    *memberProfileID,
				MemberDepartmentID: *req.MemberDepartmentID,
			}
			if err := core.MemberDepartmentHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member department history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member department history: " + err.Error()})
			}
			profile.MemberDepartmentID = req.MemberDepartmentID
		}

		if req.MemberTypeID != nil && !helpers.UUIDPtrEqual(profile.MemberTypeID, req.MemberTypeID) {
			data := &core.MemberTypeHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileID,
				MemberTypeID:    *req.MemberTypeID,
			}
			if err := core.MemberTypeHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member type history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member type history: " + err.Error()})
			}
			profile.MemberTypeID = req.MemberTypeID
		}
		if req.MemberGroupID != nil && !helpers.UUIDPtrEqual(profile.MemberGroupID, req.MemberGroupID) {
			data := &core.MemberGroupHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileID,
				MemberGroupID:   *req.MemberGroupID,
			}
			if err := core.MemberGroupHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member group history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member group history: " + err.Error()})
			}
			profile.MemberGroupID = req.MemberGroupID
		}
		if req.MemberClassificationID != nil && !helpers.UUIDPtrEqual(profile.MemberClassificationID, req.MemberClassificationID) {
			data := &core.MemberClassificationHistory{
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               *userOrg.BranchID,
				CreatedAt:              time.Now().UTC(),
				UpdatedAt:              time.Now().UTC(),
				CreatedByID:            userOrg.UserID,
				UpdatedByID:            userOrg.UserID,
				MemberProfileID:        *memberProfileID,
				MemberClassificationID: *req.MemberClassificationID,
			}
			if err := core.MemberClassificationHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member classification history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member classification history: " + err.Error()})
			}
			profile.MemberClassificationID = req.MemberClassificationID
		}
		if req.MemberCenterID != nil && !helpers.UUIDPtrEqual(profile.MemberCenterID, req.MemberCenterID) {
			data := &core.MemberCenterHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileID,
				MemberCenterID:  *req.MemberCenterID,
			}
			if err := core.MemberCenterHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.MemberProfileManager(service).UpdateByID(context, profile.ID, profile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile membership info failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile membership info: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(profile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/disconnect",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		memberProfile.UserID = nil
		memberProfile.User = nil
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = &userOrg.UserID
		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect-user/:user_id",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Connect the specified member profile to a user organization by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: invalid user_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}

		currentUserOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if currentUserOrg.UserType != core.UserOrganizationTypeOwner && currentUserOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: user not authorized",
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}

		memberProfile.UserID = userID
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = &currentUserOrg.UserID

		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}

		member, err := core.MemberProfileManager(service).GetByID(context, memberProfile.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "connect-error",
				Description: "Connect member profile to user organization failed: fetch updated member profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "connect-success",
			Description: fmt.Sprintf("Connected member profile (%s) to user (%s)", memberProfile.FullName, userID.String()),
			Module:      "MemberProfile",
		})

		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(member))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/close",
		Method:       "POST",
		RequestType:  core.MemberCloseRemarkRequest{},
		ResponseType: core.MemberCloseRemarkResponse{},
		Note:         "Close the specified member profile by member_profile_id. Accepts multiple remarks for closing.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req []core.MemberCloseRemarkRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		for i, remark := range req {
			if err := service.Validator.Struct(remark); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Validation failed for remark %d: %s", i+1, err.Error()))
			}
		}

		if len(req) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "At least one close remark is required")
		}

		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found", memberProfileID)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusNoContent, map[string]string{"error": "Current user organization not found"})
		}

		tx, endTx := service.Database.StartTransaction(context)

		memberProfile.IsClosed = true
		if err := core.MemberProfileManager(service).UpdateByIDWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to close member profile: "+endTx(err).Error())
		}

		var createdRemarks []*core.MemberCloseRemark
		for _, remarkReq := range req {
			value := &core.MemberCloseRemark{
				MemberProfileID: &memberProfile.ID,
				Reason:          remarkReq.Reason,
				Description:     remarkReq.Description,
				CreatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       time.Now().UTC(),
				UpdatedByID:     userOrg.UserID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			}

			if err := core.MemberCloseRemarkManager(service).CreateWithTx(context, tx, value); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create member close remark: " + endTx(err).Error()})
			}
			createdRemarks = append(createdRemarks, value)
		}

		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.MemberCloseRemarkManager(service).ToModels(createdRemarks))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect",
		Method:       "POST",
		RequestType:  core.MemberProfileAccountRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Connect the specified member profile to a user account using member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfileAccountRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found", memberProfileID)})
		}
		memberProfile.UserID = req.UserID
		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update member profile by specifying user connection: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/coordinates",
		Method:       "PUT",
		RequestType:  core.MemberProfileCoordinatesRequest{},
		ResponseType: core.MemberProfileResponse{},
		Note:         "Updates the coordinates (latitude and longitude) of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.MemberProfileCoordinatesRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: invalid request body: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: validation error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		profile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: member profile not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: user org error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = &userOrg.UserID
		profile.Latitude = &req.Latitude
		profile.Longitude = &req.Longitude

		if err := core.MemberProfileManager(service).UpdateByID(context, profile.ID, profile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member profile coordinates failed: update error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile coordinates: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(profile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/member-type/:member_type_id/search",
		Method:       "GET",
		ResponseType: core.MemberProfileArchiveResponse{},
		Note:         "Searches member profiles by member type ID with optional query parameters.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := helpers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberProfiles, err := core.MemberProfileManager(service).NormalPagination(context, ctx, &core.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			MemberTypeID:   memberTypeID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberProfiles)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/member-type/:member_type_id/link",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Links a member profile to a member type by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberTypeID, err := helpers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if !helpers.UUIDPtrEqual(memberProfile.MemberTypeID, memberTypeID) {
			data := &core.MemberTypeHistory{
				OrganizationID:  userOrg.OrganizationID,
				BranchID:        *userOrg.BranchID,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedByID:     userOrg.UserID,
				MemberProfileID: *memberProfileID,
				MemberTypeID:    *memberTypeID,
			}
			if err := core.MemberTypeHistoryManager(service).Create(context, data); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update member profile membership info failed: update member type history error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member type history: " + err.Error()})
			}
		}
		memberProfile.MemberTypeID = memberTypeID
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = &userOrg.UserID

		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/unlink",
		Method:       "PUT",
		ResponseType: core.MemberProfileResponse{},
		Note:         "Unlinks a member profile from its member type by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		memberProfile, err := core.MemberProfileManager(service).GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberProfile with ID %s not found: %v", memberProfileID, err)})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberProfile.MemberTypeID = nil
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = &userOrg.UserID

		if err := core.MemberProfileManager(service).UpdateByID(context, memberProfile.ID, memberProfile); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Could not update member profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})
}
