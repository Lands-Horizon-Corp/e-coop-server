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

func (c *Controller) MemberEducationalAttainmentController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Create a new educational attainment record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberEducationalAttainment{
			MemberProfileID:       *memberProfileID,
			SchoolName:            req.SchoolName,
			SchoolYear:            req.SchoolYear,
			ProgramCourse:         req.ProgramCourse,
			EducationalAttainment: req.EducationalAttainment,
			Description:           req.Description,
			CreatedAt:             time.Now().UTC(),
			CreatedByID:           user.UserID,
			UpdatedAt:             time.Now().UTC(),
			UpdatedByID:           user.UserID,
			BranchID:              *user.BranchID,
			OrganizationID:        user.OrganizationID,
		}

		if err := c.model.MemberEducationalAttainmentManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-educational-attainment/:member_educational_attainment_id",
		Method:   "PUT",
		Request:  "TMemberEducationalAttainment",
		Response: "TMemberEducationalAttainment",
		Note:     "Update an existing educational attainment record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		req, err := c.model.MemberEducationalAttainmentManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberEducationalAttainmentManager.GetByID(context, *memberEducationalAttainmentID)
		if err != nil {
			return c.NotFound(ctx, "MemberEducationalAttainment")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.SchoolName = req.SchoolName
		value.SchoolYear = req.SchoolYear
		value.ProgramCourse = req.ProgramCourse
		value.EducationalAttainment = req.EducationalAttainment
		value.Description = req.Description
		if err := c.model.MemberEducationalAttainmentManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberEducationalAttainmentManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-educational-attainment/:member_educational_attainment_id",
		Method: "DELETE",
		Note:   "Delete a member's educational attainment record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberEducationalAttainmentID, err := horizon.EngineUUIDParam(ctx, "member_educational_attainment_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member educational attainment ID")
		}
		if err := c.model.MemberEducationalAttainmentManager.DeleteByID(context, *memberEducationalAttainmentID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-educational-attainment/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member educational attainment records",
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
			bankID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.MemberEducationalAttainmentManager.GetByID(context, bankID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberEducationalAttainment with ID %s", rawID))
			}
			if err := c.model.MemberEducationalAttainmentManager.DeleteByIDWithTx(context, tx, bankID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberAddressController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Create a new address record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberAddress{
			MemberProfileID: memberProfileID,

			Label:         req.Label,
			City:          req.City,
			CountryCode:   req.CountryCode,
			PostalCode:    req.PostalCode,
			ProvinceState: req.ProvinceState,
			Barangay:      req.Barangay,
			Landmark:      req.Landmark,
			Address:       req.Address,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberAddressManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberAddressManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/:member_address_id",
		Method:   "PUT",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Update an existing address record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member address ID")
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberAddressManager.GetByID(context, *memberAddressID)
		if err != nil {
			return c.NotFound(ctx, "MemberAddress")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.Label = req.Label
		value.City = req.City
		value.CountryCode = req.CountryCode
		value.PostalCode = req.PostalCode
		value.ProvinceState = req.ProvinceState
		value.Barangay = req.Barangay
		value.Landmark = req.Landmark
		value.Address = req.Address
		if err := c.model.MemberAddressManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberAddressManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-address/:member_address_id",
		Method: "DELETE",
		Note:   "Delete a member's address record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member address ID")
		}
		if err := c.model.MemberAddressManager.DeleteByID(context, *memberAddressID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberContactReferenceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-contact-reference/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberContactReference",
		Response: "TMemberContactReference",
		Note:     "Create a new contact reference for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberContactReference{
			MemberProfileID: *memberProfileID,

			Name:          req.Name,
			Description:   req.Description,
			ContactNumber: req.ContactNumber,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberContactReferenceManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-contact-reference/:member_contact_reference_id",
		Method:   "PUT",
		Request:  "TMemberContactReference",
		Response: "TMemberContactReference",
		Note:     "Update an existing contact reference for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := horizon.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member contact reference ID")
		}
		req, err := c.model.MemberContactReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberContactReferenceManager.GetByID(context, *memberContactReferenceID)
		if err != nil {
			return c.NotFound(ctx, "MemberContactReference")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.Name = req.Name
		value.Description = req.Description
		value.ContactNumber = req.ContactNumber

		if err := c.model.MemberContactReferenceManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberContactReferenceManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-contact-reference/:member_contact_reference_id",
		Method: "DELETE",
		Note:   "Delete a member's contact reference by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberContactReferenceID, err := horizon.EngineUUIDParam(ctx, "member_contact_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member contact reference ID")
		}
		if err := c.model.MemberContactReferenceManager.DeleteByID(context, *memberContactReferenceID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberAssetController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Create a new asset record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberAsset{
			MemberProfileID: memberProfileID,

			MediaID:     req.MediaID,
			Name:        req.Name,
			EntryDate:   req.EntryDate,
			Description: req.Description,
			Cost:        req.Cost,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberAssetManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberAssetManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/:member_asset_id",
		Method:   "PUT",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Update an existing asset record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member asset ID")
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			return c.NotFound(ctx, "MemberAsset")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MediaID = req.MediaID
		value.Name = req.Name
		value.EntryDate = req.EntryDate
		value.Description = req.Description
		value.Cost = req.Cost

		if err := c.model.MemberAssetManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberAssetManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-asset/:member_asset_id",
		Method: "DELETE",
		Note:   "Delete a member's asset record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member asset ID")
		}
		if err := c.model.MemberAssetManager.DeleteByID(context, *memberAssetID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberIncomeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-income/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberIncome",
		Response: "TMemberIncome",
		Note:     "Create a new income record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberIncome{
			MemberProfileID: *memberProfileID,

			MediaID:     req.MediaID,
			Name:        req.Name,
			Amount:      req.Amount,
			ReleaseDate: req.ReleaseDate,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberIncomeManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-income/:member_income_id",
		Method:   "PUT",
		Request:  "TMemberIncome",
		Response: "TMemberIncome",
		Note:     "Update an existing income record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := horizon.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member income ID")
		}
		req, err := c.model.MemberIncomeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberIncomeManager.GetByID(context, *memberIncomeID)
		if err != nil {
			return c.NotFound(ctx, "MemberIncome")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MediaID = req.MediaID
		value.Name = req.Name
		value.Amount = req.Amount
		value.ReleaseDate = req.ReleaseDate
		if err := c.model.MemberIncomeManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberIncomeManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-income/:member_income_id",
		Method: "DELETE",
		Note:   "Delete a member's income record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberIncomeID, err := horizon.EngineUUIDParam(ctx, "member_income_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member income ID")
		}
		if err := c.model.MemberIncomeManager.DeleteByID(context, *memberIncomeID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberExpenseController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-expense/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberExpense",
		Response: "TMemberExpense",
		Note:     "Create a new expense record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberExpense{
			MemberProfileID: *memberProfileID,

			Name:        req.Name,
			Amount:      req.Amount,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberExpenseManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-expense/:member_expense_id",
		Method:   "PUT",
		Request:  "TMemberExpense",
		Response: "TMemberExpense",
		Note:     "Update an existing expense record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := horizon.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member expense ID")
		}
		req, err := c.model.MemberExpenseManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberExpenseManager.GetByID(context, *memberExpenseID)
		if err != nil {
			return c.NotFound(ctx, "MemberExpense")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.Name = req.Name
		value.Amount = req.Amount
		value.Description = req.Description
		if err := c.model.MemberExpenseManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberExpenseManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-expense/:member_expense_id",
		Method: "DELETE",
		Note:   "Delete a member's expense record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberExpenseID, err := horizon.EngineUUIDParam(ctx, "member_expense_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member expense ID")
		}
		if err := c.model.MemberExpenseManager.DeleteByID(context, *memberExpenseID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberGovernmentBenefitController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Create a new government benefit record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberGovernmentBenefit{
			MemberProfileID: *memberProfileID,

			FrontMediaID: req.FrontMediaID,
			BackMediaID:  req.BackMediaID,
			CountryCode:  req.CountryCode,
			Description:  req.Description,
			Name:         req.Name,
			Value:        req.Value,
			ExpiryDate:   req.ExpiryDate,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGovernmentBenefitManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-government-benefit/:member_government_benefit_id",
		Method:   "PUT",
		Request:  "TMemberGovernmentBenefit",
		Response: "TMemberGovernmentBenefit",
		Note:     "Update an existing government benefit record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member government benefit ID")
		}
		req, err := c.model.MemberGovernmentBenefitManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberGovernmentBenefitManager.GetByID(context, *memberGovernmentBenefitID)
		if err != nil {
			return c.NotFound(ctx, "MemberGovernmentBenefit")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.FrontMediaID = req.FrontMediaID
		value.BackMediaID = req.BackMediaID
		value.CountryCode = req.CountryCode
		value.Description = req.Description
		value.Name = req.Name
		value.Value = req.Value
		value.ExpiryDate = req.ExpiryDate

		if err := c.model.MemberGovernmentBenefitManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGovernmentBenefitManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-government-benefit/:member_government_benefit_id",
		Method: "DELETE",
		Note:   "Delete a member's government benefit record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGovernmentBenefitID, err := horizon.EngineUUIDParam(ctx, "member_government_benefit_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member government benefit ID")
		}

		if err := c.model.MemberGovernmentBenefitManager.DeleteByID(context, *memberGovernmentBenefitID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberJointAccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-joint-account/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberJointAccount",
		Response: "TMemberJointAccount",
		Note:     "Create a new joint account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberJointAccount{
			MemberProfileID: *memberProfileID,

			PictureMediaID:     req.PictureMediaID,
			SignatureMediaID:   req.SignatureMediaID,
			Description:        req.Description,
			FirstName:          req.FirstName,
			MiddleName:         req.MiddleName,
			LastName:           req.LastName,
			FullName:           req.FullName,
			Suffix:             req.Suffix,
			Birthday:           req.Birthday,
			FamilyRelationship: req.FamilyRelationship,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberJointAccountManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberJointAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-joint-account/:member_joint_account_id",
		Method:   "PUT",
		Request:  "TMemberJointAccount",
		Response: "TMemberJointAccount",
		Note:     "Update an existing joint account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := horizon.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member joint account ID")
		}
		req, err := c.model.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberJointAccountManager.GetByID(context, *memberJointAccountID)
		if err != nil {
			return c.NotFound(ctx, "MemberJointAccount")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.PictureMediaID = req.PictureMediaID
		value.SignatureMediaID = req.SignatureMediaID
		value.Description = req.Description
		value.FirstName = req.FirstName
		value.MiddleName = req.MiddleName
		value.LastName = req.LastName
		value.FullName = req.FullName
		value.Suffix = req.Suffix
		value.Birthday = req.Birthday
		value.FamilyRelationship = req.FamilyRelationship

		if err := c.model.MemberJointAccountManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberJointAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-joint-account/:member_joint_account_id",
		Method: "DELETE",
		Note:   "Delete a member's joint account record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := horizon.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member joint account ID")
		}
		if err := c.model.MemberJointAccountManager.DeleteByID(context, *memberJointAccountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberRelativeAccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Create a new relative account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberRelativeAccount{
			MemberProfileID:         *memberProfileID,
			RelativeMemberProfileID: req.RelativeMemberProfileID,
			FamilyRelationship:      req.FamilyRelationship,
			Description:             req.Description,
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             user.UserID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             user.UserID,
			BranchID:                *user.BranchID,
			OrganizationID:          user.OrganizationID,
		}

		if err := c.model.MemberRelativeAccountManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/:member_relative_account_id",
		Method:   "PUT",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Update an existing relative account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member relative account ID")
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberRelativeAccountManager.GetByID(context, *memberRelativeAccountID)
		if err != nil {
			return c.NotFound(ctx, "MemberRelativeAccount")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.RelativeMemberProfileID = req.RelativeMemberProfileID
		value.FamilyRelationship = req.FamilyRelationship
		value.Description = req.Description
		value.FamilyRelationship = req.FamilyRelationship

		if err := c.model.MemberRelativeAccountManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-relative-account/:member_relative_account_id",
		Method: "DELETE",
		Note:   "Delete a member's relative account record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member relative account ID")
		}
		if err := c.model.MemberRelativeAccountManager.DeleteByID(context, *memberRelativeAccountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberGenderController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Get member gender history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.ToModels(memberGenderHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Get member gender history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.Pagination(context, ctx, memberGenderHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "GET",
		Response: "TMemberGender[]",
		Note:     "Get all member gender records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModels(memberGender))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/search",
		Method:   "GET",
		Request:  "Filter<IMemberGender>",
		Response: "Paginated<IMemberGender>",
		Note:     "Get pagination member gender",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.Pagination(context, ctx, memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "POST",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Create a new member gender record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberGender := &model.MemberGender{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGenderManager.Create(context, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/:member_gender_id",
		Method:   "PUT",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Update an existing member gender record by ID",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberGender, err := c.model.MemberGenderManager.GetByID(context, *memberGenderID)
		if err != nil {
			return c.NotFound(ctx, "MemberGender")
		}

		memberGender.UpdatedAt = time.Now().UTC()
		memberGender.UpdatedByID = user.UserID
		memberGender.OrganizationID = user.OrganizationID
		memberGender.BranchID = *user.BranchID
		memberGender.Name = req.Name
		memberGender.Description = req.Description
		if err := c.model.MemberGenderManager.UpdateFields(context, memberGender.ID, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-gender/:member_gender_id",
		Method: "DELETE",
		Note:   "Delete a member gender record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		if err := c.model.MemberGenderManager.DeleteByID(context, *memberGenderID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-gender/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member gender records by their IDs",
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
			memberGenderID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberGenderManager.GetByID(context, memberGenderID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberGender with ID %s", rawID))
			}

			if err := c.model.MemberGenderManager.DeleteByIDWithTx(context, tx, memberGenderID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberCenterController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Get member center history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenterHistory, err := c.model.MemberCenterHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.ToModels(memberCenterHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Get member center history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenterHistory, err := c.model.MemberCenterHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.Pagination(context, ctx, memberCenterHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "GET",
		Response: "TMemberCenter[]",
		Note:     "Get all member center records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenter, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModels(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/search",
		Method:   "GET",
		Request:  "Filter<IMemberCenter>",
		Response: "Paginated<IMemberCenter>",
		Note:     "Get pagination member center",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "POST",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Create a new member center record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberCenter := &model.MemberCenter{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberCenterManager.Create(context, memberCenter); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/:member_center_id",
		Method:   "PUT",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Update an existing member center record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberCenter, err := c.model.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			return c.NotFound(ctx, "MemberCenter")
		}

		memberCenter.UpdatedAt = time.Now().UTC()
		memberCenter.UpdatedByID = user.UserID
		memberCenter.OrganizationID = user.OrganizationID
		memberCenter.BranchID = *user.BranchID
		memberCenter.Name = req.Name
		memberCenter.Description = req.Description
		if err := c.model.MemberCenterManager.UpdateFields(context, memberCenter.ID, memberCenter); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Delete a member center record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		if err := c.model.MemberCenterManager.DeleteByID(context, *memberCenterID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-center/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member center records by their IDs",
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
			memberCenterID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberCenterManager.GetByID(context, memberCenterID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberCenter with ID %s", rawID))
			}

			if err := c.model.MemberCenterManager.DeleteByIDWithTx(context, tx, memberCenterID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberTypeReferenceController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/member-type/:member_type_id",
		Method:   "GET",
		Response: "TMemberTypeReference[]",
		Note:     "Get all member type references by member_type_id for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		refs, err := c.model.MemberTypeReferenceManager.Find(context, &model.MemberTypeReference{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			MemberTypeID:   *memberTypeID,
		})
		if err != nil {
			return c.NotFound(ctx, "MemberTypeReference")
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModels(refs))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/member-type/:member_type_id/search",
		Method:   "GET",
		Response: "TMemberTypeReference[]",
		Note:     "Get all member type references by member_type_id for the current branch",
	}, func(ctx echo.Context) error {
		fmt.Println("DEBUG: Handler entered") // 1
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			fmt.Println("DEBUG: memberTypeID error:", err) // 2
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		if memberTypeID == nil {
			fmt.Println("DEBUG: memberTypeID is nil") // 3
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			fmt.Println("DEBUG: userOrganizationToken error:", err) // 4
			return ctx.NoContent(http.StatusNoContent)
		}
		if user.BranchID == nil {
			fmt.Println("DEBUG: user.BranchID is nil") // 5
			return c.BadRequest(ctx, "Branch ID is required")
		}
		fmt.Println("DEBUG: About to call Find with org:", user.OrganizationID, "branch:", *user.BranchID, "memberTypeID:", *memberTypeID) // 6
		refs, err := c.model.MemberTypeReferenceManager.Find(context, &model.MemberTypeReference{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			MemberTypeID:   *memberTypeID,
		})
		if err != nil {
			fmt.Println("DEBUG: Find error:", err) // 7
			return c.NotFound(ctx, "MemberTypeReference")
		}
		fmt.Println("DEBUG: Success, returning refs") // 8
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.Pagination(context, ctx, refs))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "GET",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}
		ref, err := c.model.MemberTypeReferenceManager.GetByIDRaw(context, *id)
		if err != nil {
			return c.NotFound(ctx, "MemberTypeReference")
		}
		return ctx.JSON(http.StatusOK, ref)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference",
		Method:   "POST",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		ref := &model.MemberTypeReference{
			AccountID:                  req.AccountID,
			MemberTypeID:               req.MemberTypeID,
			MaintainingBalance:         req.MaintainingBalance,
			Description:                req.Description,
			InterestRate:               req.InterestRate,
			MinimumBalance:             req.MinimumBalance,
			Charges:                    req.Charges,
			ActiveMemberMinimumBalance: req.ActiveMemberMinimumBalance,
			ActiveMemberRatio:          req.ActiveMemberRatio,
			OtherInterestOnSavingComputationMinimumBalance: req.OtherInterestOnSavingComputationMinimumBalance,
			OtherInterestOnSavingComputationInterestRate:   req.OtherInterestOnSavingComputationInterestRate,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberTypeReferenceManager.Create(context, ref); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "PUT",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}

		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		ref, err := c.model.MemberTypeReferenceManager.GetByID(context, *id)
		if err != nil {
			return c.NotFound(ctx, "MemberTypeReference")
		}
		ref.AccountID = req.AccountID
		ref.MemberTypeID = req.MemberTypeID
		ref.MaintainingBalance = req.MaintainingBalance
		ref.Description = req.Description
		ref.InterestRate = req.InterestRate
		ref.MinimumBalance = req.MinimumBalance
		ref.Charges = req.Charges
		ref.ActiveMemberMinimumBalance = req.ActiveMemberMinimumBalance
		ref.ActiveMemberRatio = req.ActiveMemberRatio
		ref.OtherInterestOnSavingComputationMinimumBalance = req.OtherInterestOnSavingComputationMinimumBalance
		ref.OtherInterestOnSavingComputationInterestRate = req.OtherInterestOnSavingComputationInterestRate
		ref.UpdatedAt = time.Now().UTC()
		ref.UpdatedByID = user.UserID
		if err := c.model.MemberTypeReferenceManager.UpdateFields(context, ref.ID, ref); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-type-reference/:member_type_reference_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type reference ID")
		}
		if err := c.model.MemberTypeReferenceManager.DeleteByID(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-type-reference/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member type reference records",
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
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.MemberTypeReferenceManager.GetByID(context, id); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberTypeReference with ID %s", rawID))
			}
			if err := c.model.MemberTypeReferenceManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
func (c *Controller) MemberTypeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-history",
		Method:   "GET",
		Response: "TMemberTypeHistory[]",
		Note:     "Get member type history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberTypeHistory, err := c.model.MemberTypeHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryManager.ToModels(memberTypeHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberTypeHistory[]",
		Note:     "Get member type history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberTypeHistory, err := c.model.MemberTypeHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryManager.Pagination(context, ctx, memberTypeHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type",
		Method:   "GET",
		Response: "TMemberType[]",
		Note:     "Get all member type records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberType, err := c.model.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModels(memberType))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type/search",
		Method:   "GET",
		Request:  "Filter<IMemberType>",
		Response: "Paginated<IMemberType>",
		Note:     "Get pagination member type",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type",
		Method:   "POST",
		Request:  "TMemberType",
		Response: "TMemberType",
		Note:     "Create a new member type record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberType := &model.MemberType{
			Name:           req.Name,
			Description:    req.Description,
			Prefix:         req.Prefix,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberTypeManager.Create(context, memberType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModel(memberType))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type/:member_type_id",
		Method:   "PUT",
		Request:  "TMemberType",
		Response: "TMemberType",
		Note:     "Update an existing member type record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberType, err := c.model.MemberTypeManager.GetByID(context, *memberTypeID)
		if err != nil {
			return c.NotFound(ctx, "MemberType")
		}

		memberType.UpdatedAt = time.Now().UTC()
		memberType.UpdatedByID = user.UserID
		memberType.OrganizationID = user.OrganizationID
		memberType.BranchID = *user.BranchID
		memberType.Name = req.Name
		memberType.Description = req.Description
		memberType.Prefix = req.Prefix
		if err := c.model.MemberTypeManager.UpdateFields(context, memberType.ID, memberType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModel(memberType))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-type/:member_type_id",
		Method: "DELETE",
		Note:   "Delete a member type record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		if err := c.model.MemberTypeManager.DeleteByID(context, *memberTypeID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-type/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member type records by their IDs",
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
			memberTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberTypeManager.GetByID(context, memberTypeID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberType with ID %s", rawID))
			}

			if err := c.model.MemberTypeManager.DeleteByIDWithTx(context, tx, memberTypeID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberClassificationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification-history",
		Method:   "GET",
		Response: "TMemberClassificationHistory[]",
		Note:     "Get member classification history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassificationHistory, err := c.model.MemberClassificationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryManager.ToModels(memberClassificationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberClassificationHistory[]",
		Note:     "Get member classification history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassificationHistory, err := c.model.MemberClassificationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryManager.Pagination(context, ctx, memberClassificationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification",
		Method:   "GET",
		Response: "TMemberClassification[]",
		Note:     "Get all member classification records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassification, err := c.model.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModels(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification/search",
		Method:   "GET",
		Request:  "Filter<IMemberClassification>",
		Response: "Paginated<IMemberClassification>",
		Note:     "Get pagination member classification",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification",
		Method:   "POST",
		Request:  "TMemberClassification",
		Response: "TMemberClassification",
		Note:     "Create a new member classification record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberClassificationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberClassification := &model.MemberClassification{
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberClassificationManager.Create(context, memberClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification/:member_classification_id",
		Method:   "PUT",
		Request:  "TMemberClassification",
		Response: "TMemberClassification",
		Note:     "Update an existing member classification record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberClassificationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberClassification, err := c.model.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			return c.NotFound(ctx, "MemberClassification")
		}

		memberClassification.UpdatedAt = time.Now().UTC()
		memberClassification.UpdatedByID = user.UserID
		memberClassification.OrganizationID = user.OrganizationID
		memberClassification.BranchID = *user.BranchID
		memberClassification.Name = req.Name
		memberClassification.Description = req.Description
		memberClassification.Icon = req.Icon
		if err := c.model.MemberClassificationManager.UpdateFields(context, memberClassification.ID, memberClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-classification/:member_classification_id",
		Method: "DELETE",
		Note:   "Delete a member classification record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		if err := c.model.MemberClassificationManager.DeleteByID(context, *memberClassificationID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-classification/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member classification records by their IDs",
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
			memberClassificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberClassificationManager.GetByID(context, memberClassificationID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberClassification with ID %s", rawID))
			}

			if err := c.model.MemberClassificationManager.DeleteByIDWithTx(context, tx, memberClassificationID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberOccupationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Get member occupation history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupationHistory, err := c.model.MemberOccupationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.ToModels(memberOccupationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Get member occupation history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupationHistory, err := c.model.MemberOccupationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.Pagination(context, ctx, memberOccupationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "GET",
		Response: "TMemberOccupation[]",
		Note:     "Get all member occupation records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupation, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModels(memberOccupation))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/search",
		Method:   "GET",
		Request:  "Filter<IMemberOccupation>",
		Response: "Paginated<IMemberOccupation>",
		Note:     "Get pagination member occupation",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "POST",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Create a new member occupation record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberOccupation := &model.MemberOccupation{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberOccupationManager.Create(context, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/:member_occupation_id",
		Method:   "PUT",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Update an existing member occupation record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberOccupation, err := c.model.MemberOccupationManager.GetByID(context, *memberOccupationID)
		if err != nil {
			return c.NotFound(ctx, "MemberOccupation")
		}

		memberOccupation.UpdatedAt = time.Now().UTC()
		memberOccupation.UpdatedByID = user.UserID
		memberOccupation.OrganizationID = user.OrganizationID
		memberOccupation.BranchID = *user.BranchID
		memberOccupation.Name = req.Name
		memberOccupation.Description = req.Description
		if err := c.model.MemberOccupationManager.UpdateFields(context, memberOccupation.ID, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Delete a member occupation record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		if err := c.model.MemberOccupationManager.DeleteByID(context, *memberOccupationID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-occupation/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member occupation records by their IDs",
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
			memberOccupationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberOccupationManager.GetByID(context, memberOccupationID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberOccupation with ID %s", rawID))
			}

			if err := c.model.MemberOccupationManager.DeleteByIDWithTx(context, tx, memberOccupationID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) MemberGroupController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group-history",
		Method:   "GET",
		Response: "TMemberGroupHistory[]",
		Note:     "Get member group history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroupHistory, err := c.model.MemberGroupHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryManager.ToModels(memberGroupHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberGroupHistory[]",
		Note:     "Get member group history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroupHistory, err := c.model.MemberGroupHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryManager.Pagination(context, ctx, memberGroupHistory))

	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group",
		Method:   "GET",
		Response: "TMemberGroup[]",
		Note:     "Get all member group records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroup, err := c.model.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModels(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group/search",
		Method:   "GET",
		Request:  "Filter<IMemberGroup>",
		Response: "Paginated<IMemberGroup>",
		Note:     "Get pagination member group",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group",
		Method:   "POST",
		Request:  "TMemberGroup",
		Response: "TMemberGroup",
		Note:     "Create a new member group record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGroupManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberGroup := &model.MemberGroup{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGroupManager.Create(context, memberGroup); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModel(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group/:member_group_id",
		Method:   "PUT",
		Request:  "TMemberGroup",
		Response: "TMemberGroup",
		Note:     "Update an existing member group record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := horizon.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberGroupManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberGroup, err := c.model.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			return c.NotFound(ctx, "MemberGroup")
		}

		memberGroup.UpdatedAt = time.Now().UTC()
		memberGroup.UpdatedByID = user.UserID
		memberGroup.OrganizationID = user.OrganizationID
		memberGroup.BranchID = *user.BranchID
		memberGroup.Name = req.Name
		memberGroup.Description = req.Description
		if err := c.model.MemberGroupManager.UpdateFields(context, memberGroup.ID, memberGroup); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModel(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-group/:member_group_id",
		Method: "DELETE",
		Note:   "Delete a member group record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := horizon.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		if err := c.model.MemberGroupManager.DeleteByID(context, *memberGroupID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-group/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member group records by their IDs",
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
			memberGroupID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberGroupManager.GetByID(context, memberGroupID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberGroup with ID %s", rawID))
			}

			if err := c.model.MemberGroupManager.DeleteByIDWithTx(context, tx, memberGroupID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
