package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserOrganinzationController() {

	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/permission",
		Method: "PUT",
		Note:   "Update the permission fields of a user organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_organization_id"})
		}

		var payload struct {
			PermissionName        string   `json:"permission_name" validate:"required"`
			PermissionDescription string   `json:"permission_description" validate:"required"`
			Permissions           []string `json:"permissions" validate:"required,min=1,dive,required"`
		}
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		}

		validate := validator.New()
		if err := validate.Struct(payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Get user organization
		userOrg, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "user organization not found"})
		}

		// Optionally: check if current user is allowed to update (owner/admin)
		currentUserOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		// Update fields
		userOrg.PermissionName = payload.PermissionName
		userOrg.PermissionDescription = payload.PermissionDescription
		userOrg.Permissions = payload.Permissions
		userOrg.UpdatedAt = time.Now().UTC()
		userOrg.UpdatedByID = currentUserOrg.UserID

		if err := c.model.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update permissions: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrg))
	})
	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:organization_id/seed",
		Method: "POST",
		Note:   "Seed all branches inside an organization when first created.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid organization ID")
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		userOrganizations, err := c.model.GetUserOrganizationByOrganization(context, *orgId, nil)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if len(userOrganizations) == 0 || userOrganizations == nil {
			return echo.NewHTTPError(http.StatusNotFound, "user organization not found")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, userOrganization := range userOrganizations {
			if userOrganization.UserID != user.ID {
				continue
			}
			if userOrganization.UserType != "owner" {
				return echo.NewHTTPError(http.StatusForbidden, "only owners can seed the organization")
			}
			if userOrganization.BranchID == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "branch is missing")
			}
			if userOrganization.IsSeeded {
				continue
			}
			if err := c.model.OrganizationSeeder(context, tx, user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			userOrganization.IsSeeded = true
			if err := c.model.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user organization seed status")
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusOK)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/employee/search",
		Method:   "GET",
		Request:  "Filter<TUserOrganization>",
		Response: "Paginated<TUserOrganization>",
		Note:     "Get pagination user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       "employee",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/member/search",
		Method:   "GET",
		Request:  "Filter<TUserOrganization>",
		Response: "Paginated<TUserOrganization>",
		Note:     "Get pagination user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       "member",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/none-member-profle/search",
		Method:   "GET",
		Request:  "Filter<TUserOrganization>",
		Response: "Paginated<TUserOrganization>",
		Note:     "Get pagination user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       "member",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		filteredUserOrganizations := []*model.UserOrganization{}
		for _, uo := range userOrganization {
			if uo.BranchID == nil {
				continue
			}
			userProfile, _ := c.model.MemberProfileFindUserByID(context, uo.UserID, uo.OrganizationID, *uo.BranchID)
			if userProfile == nil {
				filteredUserOrganizations = append(filteredUserOrganizations, uo)
			}
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.Pagination(context, ctx, filteredUserOrganizations))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/user/:user_id",
		Method:   "GET",
		Response: "TUserOrganization[]",
		Note:     "Retrieve all user organizations. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}
		user, err := c.model.UserManager.GetByID(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		userOrganization, err := c.model.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/current",
		Method:   "GET",
		Response: "TUserOrganization[]",
		Note:     "Retrieve all user organizations of the user logged in",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		isPending := false
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/join-request/paginated",
		Method:   "GET",
		Request:  "Filter<TUserOrganization>",
		Response: "Paginated<TUserOrganization>",
		Note:     "Get pagination user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			OrganizationID:    user.OrganizationID,
			BranchID:          user.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/join-request",
		Method:   "GET",
		Request:  "Filter<TUserOrganization>",
		Response: "Paginated<TUserOrganization>",
		Note:     "Get pagination user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		userOrganization, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			OrganizationID:    user.OrganizationID,
			BranchID:          user.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/organization/:organization_id",
		Method:   "GET",
		Response: "TUserOrganization[]",
		Note:     "Retrieve all user organizations across all branches of a specific organization. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}

		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		userOrganization, err := c.model.GetUserOrganizationByOrganization(context, organization.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/branch/:branch_id",
		Method:   "GET",
		Response: "TUserOrganization[]",
		Note:     "Retrieve all user organizations from a specific branch. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}
		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		userOrganization, err := c.model.GetUserOrganizationByBranch(context, branch.OrganizationID, branch.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/switch",
		Method: "GET",
		Note:   "Switch organization and branch stored in JWT (no database impact).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return err
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if user.ID != userOrganization.UserID {
			return ctx.NoContent(http.StatusForbidden)
		}
		if userOrganization.ApplicationStatus == "accepted" {
			if err := c.userOrganizationToken.SetUserOrganization(context, ctx, userOrganization); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrganization))
		}
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "switching forbidden - user is " + userOrganization.UserType})

	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/unswitch",
		Method: "POST",
		Note:   "Remove organization and branch from JWT (no database impact).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		c.userOrganizationToken.Token.CleanToken(context, ctx)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/developer-key-refresh",
		Method: "POST",
		Note:   "Refresh developer key associated with the user organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userOrg.UserID.String())
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "something wrong generting developer key"})
		}
		userOrg.DeveloperSecretKey = developerKey + uuid.NewString() + "-horizon"
		if err := c.model.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "failed to update user: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrg))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/invitation-code/:code/join",
		Method: "POST",
		Note:   "Join organization and branch using an invitation code.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		code := ctx.Param("code")
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		invitationCode, err := c.model.VerifyInvitationCodeByCode(context, code)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		if invitationCode == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "invitation code not found",
			})
		}
		switch invitationCode.UserType {
		case "member":
			if !c.model.UserOrganizationMemberCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot join as member"})
			}
		case "employee":
			if !c.model.UserOrganizationEmployeeCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot join as employee"})
			}
		default:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot join as employee"})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "something wrong generting developer key"})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         invitationCode.OrganizationID,
			BranchID:               &invitationCode.BranchID,
			UserID:                 user.ID,
			UserType:               invitationCode.UserType,
			Description:            invitationCode.Description,
			ApplicationDescription: "anything",
			ApplicationStatus:      "pending",
			DeveloperSecretKey:     developerKey,

			PermissionName:        invitationCode.PermissionName,
			PermissionDescription: invitationCode.PermissionDescription,
			Permissions:           invitationCode.Permissions,

			UserSettingDescription:  "user settings",
			UserSettingStartOR:      0,
			UserSettingEndOR:        0,
			UserSettingUsedOR:       0,
			UserSettingStartVoucher: 0,
			UserSettingEndVoucher:   0,
			UserSettingUsedVoucher:  0,
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": tx.Error.Error()})
		}
		if err := c.model.RedeemInvitationCode(context, tx, invitationCode.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})

		}
		if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, userOrg); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrg))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/organization/:organization_id/branch/:branch_id/join",
		Method: "POST",
		Note:   "Join an organization and branch that is already created.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return err
		}
		req, err := c.model.UserOrganizationManager.Validate(ctx)
		if err != nil {
			return err
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		if req.ApplicationStatus == "member" {
			if !c.model.UserOrganizationMemberCanJoin(context, user.ID, *orgId, *branchId) {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as member"})
			}
		}
		if req.ApplicationStatus == "employee" {
			if !c.model.UserOrganizationEmployeeCanJoin(context, user.ID, *orgId, *branchId) {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as employee"})
			}
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "something wrong generting developer key"})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         *orgId,
			BranchID:               branchId,
			UserID:                 user.ID,
			UserType:               "member",
			Description:            req.Description,
			ApplicationDescription: "",
			ApplicationStatus:      "pending",
			DeveloperSecretKey:     developerKey,
			PermissionName:         "member",
			PermissionDescription:  "just a member",
			Permissions:            []string{},

			UserSettingDescription:  "",
			UserSettingStartOR:      0,
			UserSettingEndOR:        0,
			UserSettingUsedOR:       0,
			UserSettingStartVoucher: 0,
			UserSettingEndVoucher:   0,
			UserSettingUsedVoucher:  0,
		}

		if err := c.model.UserOrganizationManager.Create(context, userOrg); err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrg))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/leave",
		Method: "POST",
		Note:   "Leave a specific organization and branch that is already joined. (Must have Current Organization)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		switch userOrg.UserType {
		case "owner", "employee":
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "owners and employees cannot leave an organization"})
		}

		if err := c.model.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/organization/:organization_id/branch/:branch_id/can-join-member",
		Method: "GET",
		Note:   "Check if the user can join as an member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return err
		}
		if !c.model.UserOrganizationMemberCanJoin(context, user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as member"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/organization/:organization_id/branch/:branch_id/can-join-employee",
		Method: "GET",
		Note:   "Check if the user can join as a empolyee.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return err
		}
		if !c.model.UserOrganizationEmployeeCanJoin(context, user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as employee"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/:user_organization_id",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve specific user organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return err
		}
		userOrg, err := c.model.UserOrganizationManager.GetByIDRaw(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, userOrg)
	})

	// USER ORGANIZATION

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/accept",
		Method: "POST",
		Note:   "Accept an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_organization_id"})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		userOrg, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		// Only allow org admins/owners to accept applications
		if user.UserType != "owner" && user.UserType != "admin" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "only organization owners or admins can accept applications"})
		}

		// Prevent users from accepting their own application
		if user.UserID == userOrg.UserID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "you cannot accept your own application"})
		}

		userOrg.ApplicationStatus = "accepted"
		if err := c.model.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user org for accept: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/reject",
		Method: "DELETE",
		Note:   "Reject an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_organization_id"})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		userOrg, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		// Only allow org admins/owners to reject applications
		if user.UserType != "owner" && user.UserType != "admin" && user.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "only organization owners or admins can reject applications"})
		}

		// Prevent users from rejecting their own application
		if user.UserID == userOrg.UserID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "you cannot reject your own application"})
		}

		if err := c.model.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user org for reject: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id",
		Method: "DELETE",
		Note:   "Delete a user organization by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return err
		}
		userOrg, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err := c.model.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/bulk-delete",
		Method: "DELETE",
		Note:   "Delete multiple user organizations by providing an array of IDs in the request body.",
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
			userOrgId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.UserOrganizationManager.GetByID(context, userOrgId); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("User organization with ID %s", rawID))
			}

			if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrgId); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/employee",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve all employees of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		employees, err := c.model.Employees(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(employees))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/members",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve all members of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		members, err := c.model.Members(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(members))
	})
}
