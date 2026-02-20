package member_profile

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func MemberProfileController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/pending",
		Method:       "GET",
		ResponseType: types.MemberProfileResponse{},
		Note:         "Returns all pending member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		memberProfile, err := core.FindLatestMembers(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get pending member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModels(memberProfile))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/summary",
		Method:       "GET",
		ResponseType: types.MemberProfileDashboardSummaryResponse{},
		Note:         "Returns total number of  member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}
		totalMembers, err := core.MemberProfileManager(service).Find(context, &types.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Status:         types.MemberStatusVerified,
		}, "MemberType", "Media")
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles: " + err.Error()})
		}
		femaleCount, maleCount := 0, 0
		memberTypeCount := []types.MemberTypeCountResponse{}
		for _, member := range totalMembers {
			if member.Sex == types.MemberFemale {
				femaleCount++
			} else if member.Sex == types.MemberMale {
				maleCount++
			}
			found := false
			for i, count := range memberTypeCount {
				if &count.MemberTypeID == member.MemberTypeID {
					memberTypeCount[i].Count++
					found = true
					break
				}
			}
			if !found {
				memberTypeCount = append(memberTypeCount, types.MemberTypeCountResponse{
					MemberTypeID: *member.MemberTypeID,
					Count:        1,
				})
			}
		}
		return ctx.JSON(http.StatusOK, types.MemberProfileDashboardSummaryResponse{
			TotalMembers:       int64(len(totalMembers)),
			TotalMaleMembers:   int64(maleCount),
			TotalFemaleMembers: int64(femaleCount),
			MemberTypeCounts:   memberTypeCount,
		})
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/user-account",
		Method:       "POST",
		RequestType:  types.MemberProfileUserAccountRequest{},
		ResponseType: types.MemberProfileResponse{},
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
		var req types.MemberProfileUserAccountRequest
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
		userProfile := &types.User{
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
		newUserOrg := &types.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            userOrg.UserID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            userOrg.UserID,
			OrganizationID:         userOrg.OrganizationID,
			BranchID:               userOrg.BranchID,
			UserID:                 userProfile.ID,
			UserType:               types.UserOrganizationTypeMember,
			Description:            "",
			ApplicationDescription: "anything",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey,
			PermissionName:         string(types.UserOrganizationTypeMember),
			PermissionDescription:  "",
			Permissions:            []string{},
			UserSettingDescription: "user settings",
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/approve",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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

		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
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
		tx, endTx := service.Database.StartTransaction(context)
		memberProfile.Status = types.MemberStatusVerified
		memberProfile.MemberVerifiedByEmployeeUserID = &userOrg.UserID
		if err := core.MemberProfileManager(service).UpdateByIDWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: update error: " + endTx(err).Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member profile: " + err.Error()})
		}
		if memberProfile.UserID != nil {
			branchSetting, err := core.BranchSettingManager(service).FindOne(context, &types.BranchSetting{
				BranchID: memberProfile.BranchID,
			})
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "approve-error",
					Description: "Approve member profile failed: branch setting not found: " + endTx(err).Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch settings not found"})
			}

			_, err = core.MemberAccountingLedgerUpdateOrCreate(
				context,
				service,
				tx, 0, types.MemberAccountingLedgerUpdateOrCreateParams{
					MemberProfileID: memberProfile.ID,
					AccountID:       *branchSetting.AccountWalletID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					UserID:          *memberProfile.UserID,
					LastPayTime:     time.Now().UTC(),
					Wallet:          true,
				},
			)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "approve-error",
					Description: "Approve member profile failed: create member accounting ledger error: " + endTx(err).Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member accounting ledger: " + err.Error()})
			}

		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Approve member profile failed: commit tx error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Approved member profile: " + memberProfile.FullName,
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(memberProfile))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/reject",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile",
		Method:       "GET",
		ResponseType: types.MemberProfileResponse{},
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/search",
		Method:       "GET",
		ResponseType: types.MemberProfileResponse{},
		Note:         "Returns paginated member profiles for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberProfileManager(service).NormalPagination(context, ctx, &types.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member profiles for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id",
		Method:       "GET",
		ResponseType: types.MemberProfileResponse{},
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

	service.API.RegisterWebRoute(horizon.Route{
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-profile/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member profiles and all their connections by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect-user",
		Method:       "POST",
		RequestType:  types.MemberProfileAccountRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Connects the specified member profile to a user account by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfileAccountRequest
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
	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/quick-create",
		Method:       "POST",
		RequestType:  types.MemberProfileQuickCreateRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Quickly creates a new member profile with minimal required fields.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfileQuickCreateRequest
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
		branchSetting, err := core.BranchSettingManager(service).FindOneWithLock(context, tx, &types.BranchSetting{
			BranchID: *userOrg.BranchID,
		})
		if req.PBAutoGenerated {
			if branchSetting.MemberProfilePassbookORUnique {
				memberProfiles, err := core.MemberProfileManager(service).Find(context, &types.MemberProfile{
					Passbook: req.Passbook,
				})
				if err != nil {
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Quick create member profile failed: failed to check passbook uniqueness: " + endTx(err).Error(),
						Module:      "MemberProfile",
					})
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check passbook uniqueness"})
				}
				if len(memberProfiles) > 0 {
					endTx(eris.New("member profile with the same passbook already exists"))
					event.Footstep(ctx, service, event.FootstepEvent{
						Activity:    "create-error",
						Description: "Quick create member profile failed: member profile with the same passbook already exists",
						Module:      "MemberProfile",
					})
					return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Member profile with the same passbook already exists"})
				}
			}
			branchSetting.MemberProfilePassbookORCurrent++
			if err := core.BranchSettingManager(service).UpdateByIDWithTx(context, tx, branchSetting.ID, branchSetting); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch settings: " + endTx(err).Error()})
			}
		}

		var userProfile *types.User
		var userProfileID *uuid.UUID
		if req.AccountInfo == nil && req.Status == types.MemberStatusVerified {
			if req.AccountInfo != nil && req.Status == types.MemberStatusVerified {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: cannot create verified member without account info",
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot create a verified member without providing account info"})
			}

		}
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
			userProfile = &types.User{
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

		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: branch setting not found: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Branch settings not found",
			})
		}

		profile := &types.MemberProfile{
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
			Sex:                  req.Sex,
			BirthPlace:           req.BirthPlace,
		}
		if err := core.MemberProfileManager(service).CreateWithTx(context, tx, profile); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Quick create member profile failed: create profile error: " + err.Error(),
				Module:      "MemberProfile",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Could not create member profile: " + endTx(err).Error(),
			})
		}

		if req.Status == types.MemberStatusVerified {
			_, err := core.MemberAccountingLedgerUpdateOrCreate(
				context,
				service,
				tx, 0, types.MemberAccountingLedgerUpdateOrCreateParams{
					MemberProfileID: profile.ID,
					AccountID:       *branchSetting.AccountWalletID,
					OrganizationID:  userOrg.OrganizationID,
					BranchID:        *userOrg.BranchID,
					UserID:          userProfile.ID,
					LastPayTime:     time.Now().UTC(),
					Wallet:          true,
				},
			)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Quick create member profile failed: create member accounting ledger error: " + err.Error(),
					Module:      "MemberProfile",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member accounting ledger: " + endTx(err).Error()})
			}

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
			userOrg := &types.UserOrganization{
				CreatedAt:              time.Now().UTC(),
				CreatedByID:            userOrg.UserID,
				UpdatedAt:              time.Now().UTC(),
				UpdatedByID:            userOrg.UserID,
				OrganizationID:         userOrg.OrganizationID,
				BranchID:               userOrg.BranchID,
				UserID:                 *userProfileID,
				UserType:               types.UserOrganizationTypeMember,
				Description:            "",
				ApplicationDescription: "anything",
				ApplicationStatus:      "accepted",
				DeveloperSecretKey:     developerKey,
				PermissionName:         string(types.UserOrganizationTypeMember),
				PermissionDescription:  "",
				Permissions:            []string{},
				UserSettingDescription: "user settings",
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/personal-info",
		Method:       "PUT",
		RequestType:  types.MemberProfilePersonalInfoRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Updates the personal information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfilePersonalInfoRequest
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
		profile.Sex = req.Sex
		if req.MemberGenderID != nil && !helpers.UUIDPtrEqual(profile.MemberGenderID, req.MemberGenderID) {
			data := &types.MemberGenderHistory{
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
			data := &types.MemberOccupationHistory{
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
		if req.MemberAddressDeletedID != nil {
			for _, deletedID := range *req.MemberAddressDeletedID {
				address, err := core.MemberAddressManager(service).GetByID(context, deletedID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find member address for deletion: " + err.Error()})
				}
				if err := core.MemberAddressManager(service).Delete(context, address.ID); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member address: " + err.Error()})
				}
			}
		}
		if req.MemberAddress != nil {
			for _, addrReq := range req.MemberAddress {
				if addrReq.ID != uuid.Nil {
					existingRecord, err := core.MemberAddressManager(service).GetByID(context, addrReq.ID)
					if err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find existing member address: " + err.Error()})
					}
					if existingRecord.MemberProfileID == nil || *existingRecord.MemberProfileID != profile.ID {
						return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot update member address that doesn't belong to this member profile"})
					}
					existingRecord.UpdatedAt = time.Now().UTC()
					existingRecord.UpdatedByID = &userOrg.UserID
					existingRecord.Label = addrReq.Label
					existingRecord.City = addrReq.City
					existingRecord.CountryCode = addrReq.CountryCode
					existingRecord.PostalCode = addrReq.PostalCode
					existingRecord.ProvinceState = addrReq.ProvinceState
					existingRecord.AreaID = addrReq.AreaID
					existingRecord.Barangay = addrReq.Barangay
					existingRecord.Landmark = addrReq.Landmark
					existingRecord.Address = addrReq.Address
					existingRecord.Latitude = addrReq.Latitude
					existingRecord.Longitude = addrReq.Longitude
					if err := core.MemberAddressManager(service).UpdateByID(context, existingRecord.ID, existingRecord); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member address: " + err.Error()})
					}
				} else {
					newAddress := &types.MemberAddress{
						CreatedAt:       time.Now().UTC(),
						UpdatedAt:       time.Now().UTC(),
						CreatedByID:     &userOrg.UserID,
						UpdatedByID:     &userOrg.UserID,
						OrganizationID:  userOrg.OrganizationID,
						BranchID:        *userOrg.BranchID,
						MemberProfileID: &profile.ID,
						Label:           addrReq.Label,
						City:            addrReq.City,
						CountryCode:     addrReq.CountryCode,
						PostalCode:      addrReq.PostalCode,
						ProvinceState:   addrReq.ProvinceState,
						AreaID:          addrReq.AreaID,
						Barangay:        addrReq.Barangay,
						Landmark:        addrReq.Landmark,
						Address:         addrReq.Address,
						Latitude:        addrReq.Latitude,
						Longitude:       addrReq.Longitude,
					}

					if err := core.MemberAddressManager(service).Create(context, newAddress); err != nil {
						return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member address: " + err.Error()})
					}
				}
			}
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: fmt.Sprintf("Updated member profile personal info: %s", profile.FullName),
			Module:      "MemberProfile",
		})
		return ctx.JSON(http.StatusOK, core.MemberProfileManager(service).ToModel(profile))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/membership-info",
		Method:       "PUT",
		RequestType:  types.MemberProfileMembershipInfoRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Updates the membership information of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfileMembershipInfoRequest
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
			data := &types.MemberDepartmentHistory{
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
			data := &types.MemberTypeHistory{
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
			data := &types.MemberGroupHistory{
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
			data := &types.MemberClassificationHistory{
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
			data := &types.MemberCenterHistory{
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/disconnect",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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
		if userOrg.UserType != types.UserOrganizationTypeOwner && userOrg.UserType != types.UserOrganizationTypeEmployee {
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect-user/:user_id",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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
		if currentUserOrg.UserType != types.UserOrganizationTypeOwner && currentUserOrg.UserType != types.UserOrganizationTypeEmployee {
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/close",
		Method:       "POST",
		RequestType:  types.MemberCloseRemarkRequest{},
		ResponseType: types.MemberCloseRemarkResponse{},
		Note:         "Close the specified member profile by member_profile_id. Accepts multiple remarks for closing.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req []types.MemberCloseRemarkRequest
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

		var createdRemarks []*types.MemberCloseRemark
		for _, remarkReq := range req {
			value := &types.MemberCloseRemark{
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/connect",
		Method:       "POST",
		RequestType:  types.MemberProfileAccountRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Connect the specified member profile to a user account using member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfileAccountRequest
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/coordinates",
		Method:       "PUT",
		RequestType:  types.MemberProfileCoordinatesRequest{},
		ResponseType: types.MemberProfileResponse{},
		Note:         "Updates the coordinates (latitude and longitude) of a member profile by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req types.MemberProfileCoordinatesRequest
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/member-type/:member_type_id/search",
		Method:       "GET",
		ResponseType: types.MemberProfileArchiveResponse{},
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
		memberProfiles, err := core.MemberProfileManager(service).NormalPagination(context, ctx, &types.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			MemberTypeID:   memberTypeID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch member profiles: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberProfiles)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/member-type/:member_type_id/link",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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
			data := &types.MemberTypeHistory{
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

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-profile/:member_profile_id/unlink",
		Method:       "PUT",
		ResponseType: types.MemberProfileResponse{},
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
