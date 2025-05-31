package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) UserOrganinzationController() {

	req := c.provider.Service.Request

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
			if err := c.model.OrganizationSeeder(context, tx, user.ID, userOrganization.ID, *userOrganization.BranchID); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			userOrganization.IsSeeded = true
			if err := c.model.UserOrganizationManager.UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
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
		Route:    "/user-organization/user/:user_id",
		Method:   "GET",
		Response: "TUserOrganization",
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
		Route:    "/user-organization/organization/:organization_id",
		Method:   "GET",
		Response: "TUserOrganization",
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
		Response: "TUserOrganization",
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/switch",
		Method: "POST",
		Note:   "Switch organization and branch stored in JWT (no database impact).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return err
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.userOrganizationToken.SetUserOrganization(context, ctx, userOrganization); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModel(userOrganization))
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
		if err := c.model.UserOrganizationManager.UpdateByID(context, userOrg.ID, userOrg); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if invitationCode == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "invitation code not found",
			})
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
			OrganizationID:         invitationCode.OrganizationID,
			BranchID:               &invitationCode.BranchID,
			UserID:                 user.ID,
			UserType:               invitationCode.UserType,
			Description:            invitationCode.Description,
			ApplicationDescription: "anything",
			ApplicationStatus:      "pending",
			DeveloperSecretKey:     developerKey,
			PermissionName:         invitationCode.UserType,
			PermissionDescription:  "",
			Permissions:            []string{},

			UserSettingDescription:  "user settings",
			UserSettingStartOR:      0,
			UserSettingEndOR:        0,
			UserSettingUsedOR:       0,
			UserSettingStartVoucher: 0,
			UserSettingEndVoucher:   0,
			UserSettingUsedVoucher:  0,
		}
		tx := c.provider.Service.Database.Client().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		if err := c.model.RedeemInvitationCode(context, tx, invitationCode.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})

		}
		if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, userOrg); err != nil {
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
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
			UserType:               req.UserType,
			Description:            req.Description,
			ApplicationDescription: "",
			ApplicationStatus:      "pending",
			DeveloperSecretKey:     developerKey,
			PermissionName:         req.UserType,
			PermissionDescription:  "",
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
}
