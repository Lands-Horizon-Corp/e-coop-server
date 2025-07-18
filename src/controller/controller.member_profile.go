package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberProfileController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/pending",
		Method:   "GET",
		Response: "[]MemberProfile",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		memberProfile, err := c.model.MemberProfileManager.Find(context, &model.MemberProfile{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			Status:         "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModels(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/user-account",
		Method:   "POST",
		Request:  "MemberProfilePersonalInfoRequest",
		Response: "MemberProfile",
		Note:     "Quickly create a new member profile with minimal required fields.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		var req model.MemberProfileUserAccountRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		tx := c.provider.Service.Database.Client().Begin()
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.Password)
		if err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not create user profile: %v", err))
		}
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			tx.Rollback()
			return c.NotFound(ctx, fmt.Sprintf("MemberProfile with ID %s not found", memberProfileID))
		}
		memberProfile.UserID = &userProfile.ID
		memberProfile.UpdatedAt = time.Now().UTC()
		memberProfile.UpdatedByID = userOrg.UserID

		if err := c.model.MemberProfileManager.UpdateFieldsWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not create member profile: %v", err))
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userProfile.ID.String())
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "something wrong generating developer key"})
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/close",
		Method:   "PUT",
		Request:  "[]TMemberCloseRemarkRequest",
		Response: "TMemberProfile",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		var reqs []model.MemberCloseRemarkRequest
		if err := ctx.Bind(&reqs); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, req := range reqs {
			if err := c.provider.Service.Validator.Struct(req); err != nil {
				tx.Rollback()
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			value := &model.MemberCloseRemark{
				Reason:          req.Reason,
				Description:     req.Description,
				MemberProfileID: memberProfileID,
				CreatedAt:       time.Now().UTC(),
				CreatedByID:     userOrg.UserID,
				UpdatedAt:       time.Now().UTC(),
				UpdatedByID:     userOrg.UserID,
				BranchID:        *userOrg.BranchID,
				OrganizationID:  userOrg.OrganizationID,
			}
			if err := c.model.MemberCloseRemarkManager.CreateWithTx(context, tx, value); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		memberProfile.IsClosed = true
		if err := c.model.MemberProfileManager.UpdateFieldsWithTx(context, tx, memberProfile.ID, memberProfile); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/connect-user-account/:user_id",
		Method:   "PUT",
		Response: "TMemberProfile",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		userID, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		user, err := c.model.UserManager.GetByID(context, *userID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		memberProfile.UserID = &user.ID
		memberProfile.MemberVerifiedByEmployeeUserID = &userOrg.UserID
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/disconnect",
		Method:   "PUT",
		Response: "TMemberProfile",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized"})
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		memberProfile.UserID = nil
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/approve",
		Method:   "PUT",
		Response: "MemberProfile",
		Note:     "Approve member profiles",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return c.NotFound(ctx, "MemberProfile")
		}
		memberProfile.Status = "verified"
		memberProfile.MemberVerifiedByEmployeeUserID = &userOrg.UserID
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update member profile: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/reject",
		Method:   "PUT",
		Response: "MemberProfile",
		Note:     "Reject member profile",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return c.BadRequest(ctx, "User is not authorized")
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return c.NotFound(ctx, "MemberProfile")
		}
		memberProfile.Status = "not allowed	"
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update member profile: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile",
		Method:   "GET",
		Response: "[]IMemberProfile",
		Note:     "Retrieve a list of all member profiles.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		memberProfile, err := c.model.MemberProfileCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModels(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/search",
		Method:   "GET",
		Request:  "Filter<IMemberProfile>",
		Response: "Paginated<IMemberProfile>",
		Note:     "Get pagination member occupation",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberProfileCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id",
		Method:   "GET",
		Response: "MemberProfile",
		Note:     "Retrieve a specific member profile by its member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		memberProfile, err := c.model.MemberProfileManager.GetByIDRaw(context, *memberProfileID)
		if err != nil {
			return c.NotFound(ctx, "MemberProfile")
		}
		return ctx.JSON(http.StatusOK, memberProfile)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-profile/:member_profile_id",
		Method: "DELETE",
		Note:   "Delete a specific member-profile by its member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return c.NotFound(ctx, fmt.Sprintf("MemberProfile with ID %s not found", memberProfileID.String()))
		}
		if err := c.model.MemberProfileDestroy(context, tx, memberProfile.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// ...existing code...
	req.RegisterRoute(horizon.Route{
		Route:   "/member-profile/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member profile records and all their connections",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			if rawID == "" {
				continue
			}
			id := uuid.MustParse(rawID)
			memberProfile, err := c.model.MemberProfileManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if err := c.model.MemberProfileDestroy(context, tx, memberProfile.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
	// ...existing code...

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/connect-user",
		Method:   "POST",
		Request:  "MemberProfileAccountRequest",
		Response: "MemberProfile",
		Note:     "Connect the specified member profile to a user account using member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfileAccountRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			return c.NotFound(ctx, fmt.Sprintf("MemberProfile with ID %s not found", memberProfileId))
		}
		memberProfile.UserID = req.UserID
		if err := c.model.MemberProfileManager.UpdateFields(context, memberProfile.ID, memberProfile); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update member profile by specifying user connection: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(memberProfile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/quick-create",
		Method:   "POST",
		Request:  "MemberProfilePersonalInfoRequest",
		Response: "MemberProfile",
		Note:     "Quickly create a new member profile with minimal required fields.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		// ...existing code...
		var req model.MemberProfileQuickCreateRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		tx := c.provider.Service.Database.Client().Begin()

		var userProfile *model.User
		var userProfileID *uuid.UUID

		if req.AccountInfo != nil {
			hashedPwd, err := c.provider.Service.Security.HashPassword(context, req.AccountInfo.Password)
			if err != nil {
				tx.Rollback()
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not create user profile: %v", err))
			}
			if tx.Error != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not create member profile: %v", err))
		}

		if userProfile != nil {
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "something wrong generating developer key"})
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
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/personal-info",
		Method:   "PUT",
		Request:  "MemberProfilePersonalInfoRequest",
		Response: "MemberProfile",
		Note:     "Update the personal information of a member profile identified by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfilePersonalInfoRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		profile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			return c.NotFound(ctx, fmt.Sprintf("MemberProfile with ID %s not found", memberProfileId))
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not update member gender history: %v", err))
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not update member occupation history: %v", err))
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could update member profile: %v", err))
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-profile/:member_profile_id/membership-info",
		Method:   "PUT",
		Request:  "MemberProfileMembershipInfoRequest",
		Response: "MemberProfile",
		Note:     "Update the membership information of a member profile identified by member_profile_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.MemberProfileMembershipInfoRequest
		if err := ctx.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		profile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			return c.NotFound(ctx, fmt.Sprintf("MemberProfile with ID %s not found", memberProfileId))
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		profile.UpdatedAt = time.Now().UTC()
		profile.UpdatedByID = userOrg.UserID
		profile.Passbook = req.Passbook
		profile.OldReferenceID = req.OldReferenceID
		profile.RecruitedByMemberProfileID = req.RecruitedByMemberProfileID
		profile.Status = req.Status

		// MemberTypeID
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could update member profile: %v", err))
			}
			profile.MemberTypeID = req.MemberTypeID
		}

		// MemberGroupID
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not update member group history: %v", err))
			}
			profile.MemberGroupID = req.MemberGroupID
		}

		// MemberClassificationID
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not update member classification history: %v", err))
			}
			profile.MemberClassificationID = req.MemberClassificationID
		}

		// MemberCenterID
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could not update member center history: %v", err))
			}
			profile.MemberCenterID = req.MemberCenterID
		}

		profile.IsMutualFundMember = req.IsMutualFundMember
		profile.IsMicroFinanceMember = req.IsMicroFinanceMember

		if err := c.model.MemberProfileManager.UpdateFields(context, profile.ID, profile); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("could update member profile: %v", err))
		}
		return ctx.JSON(http.StatusOK, c.model.MemberProfileManager.ToModel(profile))
	})
}
