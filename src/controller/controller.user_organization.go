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
		if user.UserType != "owner" && user.UserType != "admin" {
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

func (c *Controller) BranchController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/branch",
		Method:   "GET",
		Response: "TBranch[]",
		Note:     "If there's no user organization (e.g., unauthenticated), return all branches. If a user organization exists (from JWT), filter branches by that organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg == nil {
			branches, err := c.model.BranchManager.ListRaw(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := c.model.GetBranchesByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModels(branches))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/organization/:organization_id",
		Method:   "GET",
		Response: "TBranch[]",
		Note:     "Returns branches filtered by a specific organization ID provided in the URL path parameter.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		branches, err := c.model.GetBranchesByOrganization(context, *orgId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModels(branches))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/organization/:organization_id",
		Method:   "POST",
		Request:  "TBranch[]",
		Response: "{branch: TBranch, user_organization: TUserOrganization}",
		Note:     "Creates a new branch under a user organization. If the user organization doesn't have a branch yet, it will be updated. Otherwise, a new user organization record is created with the new branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request payload
		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		organzationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
		}

		userOrganization, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:         user.ID,
			OrganizationID: *organzationId,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.UserType != "owner" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User must be an owner of this organization"})
		}

		organization, err := c.model.OrganizationManager.GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Associated organization not found"})
		}

		branchCount, err := c.model.GetBranchesByOrganizationCount(context, organization.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count organization branches"})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for current subscription plan"})
		}

		branch := &model.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: userOrganization.OrganizationID,

			MediaID:       req.MediaID,
			Type:          req.Type,
			Name:          req.Name,
			Email:         req.Email,
			Description:   req.Description,
			CountryCode:   req.CountryCode,
			ContactNumber: req.ContactNumber,
			Address:       req.Address,
			Province:      req.Province,
			City:          req.City,
			Region:        req.Region,
			Barangay:      req.Barangay,
			PostalCode:    req.PostalCode,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			IsMainBranch:  req.IsMainBranch,
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		if err := c.model.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}

		if userOrganization.BranchID == nil {
			// Update existing userOrganization
			userOrganization.BranchID = &branch.ID
			userOrganization.UpdatedAt = time.Now().UTC()
			userOrganization.UpdatedByID = user.ID

			if err := c.model.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + err.Error()})
			}
		} else {
			// Create new userOrganization with new branch
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key"})
			}

			newUserOrg := &model.UserOrganization{
				CreatedAt:          time.Now().UTC(),
				CreatedByID:        user.ID,
				UpdatedAt:          time.Now().UTC(),
				UpdatedByID:        user.ID,
				OrganizationID:     userOrganization.OrganizationID,
				BranchID:           &branch.ID,
				UserID:             user.ID,
				UserType:           "owner",
				ApplicationStatus:  "accepted",
				DeveloperSecretKey: developerKey + uuid.NewString() + "-horizon",
				PermissionName:     "owner",
				Permissions:        []string{},
			}

			if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed: " + err.Error()})
		}

		// EVENT
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("%s: %s", "Create: ", branch.Name),
			Description: fmt.Sprintf("%s: %s", "Creates a new branch", branch.Name),
		})

		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModel(branch))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/:branch_id",
		Method:   "PUT",
		Request:  "TBranch",
		Response: "{branch: TBranch, user_organization: TUserOrganization}",
		Note:     "Updates the branch information under the specified user organization. Only allowed if the user is an 'owner' and the user organization already has an existing branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request body
		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		// Get currently authenticated user
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required: " + err.Error()})
		}

		// Parse and validate user organization ID
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}

		userOrg, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:   user.ID,
			BranchID: branchId,
		})
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user user org ID: " + err.Error()})
		}
		if userOrg.UserType != "owner" {
			return c.BadRequest(ctx, "Unauthorized")
		}

		// Retrieve the branch
		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Associated branch not found: " + err.Error()})
		}

		// Update branch fields
		branch.UpdatedAt = time.Now().UTC()
		branch.UpdatedByID = user.ID
		branch.MediaID = req.MediaID
		branch.Type = req.Type
		branch.Name = req.Name
		branch.Email = req.Email
		branch.Description = req.Description
		branch.CountryCode = req.CountryCode
		branch.ContactNumber = req.ContactNumber
		branch.Address = req.Address
		branch.Province = req.Province
		branch.City = req.City
		branch.Region = req.Region
		branch.Barangay = req.Barangay
		branch.PostalCode = req.PostalCode
		branch.Latitude = req.Latitude
		branch.Longitude = req.Longitude
		branch.IsMainBranch = req.IsMainBranch

		// Save changes to the branch
		if err := c.model.BranchManager.UpdateFields(context, branch.ID, branch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		// EVENT
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("%s: %s", "Update: ", branch.Name),
			Description: fmt.Sprintf("%s: %s", "Updates the branch", branch.Name),
		})

		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModel(branch))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/branch/:branch_id",
		Method: "DELETE",
		Note:   "Deletes a branch and the associated user organization if the user is the owner and fewer than 3 members exist under that branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required: " + err.Error()})
		}
		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found"})
		}

		userOrganization, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:         user.ID,
			BranchID:       branchId,
			OrganizationID: branch.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.UserType != "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owners can delete branches"})
		}
		count, err := c.model.CountUserOrganizationPerBranch(context, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count user organizations: " + err.Error()})
		}
		if count > 2 {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "Cannot delete branch with more than 2 members",
			})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
		}
		if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed: " + err.Error()})
		}
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("%s: %s", "Update: ", branch.Name),
			Description: fmt.Sprintf("%s: %s", "Updates the branch", branch.Name),
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}

func (c *Controller) OrganizationController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/organization",
		Method:   "GET",
		Response: "TOrganization[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organization, err := c.model.GetPublicOrganization(context)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModels(organization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization/:organization_id",
		Method:   "GET",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid organization ID")
		}
		organization, err := c.model.OrganizationManager.GetByIDRaw(context, *organizationID)
		if err != nil {
			return c.NotFound(ctx, "Organization")
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization",
		Method:   "POST",
		Request:  "TOrganization",
		Response: "{organization: TOrganization, user_organization: TUserOrganization}",
		Note:     "(User must be logged in) This will be use to create an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			return err
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		subscription, err := c.model.SubscriptionPlanManager.GetByID(context, *req.SubscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Subscription plan not found"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		var subscriptionEndDate time.Time
		if req.SubscriptionPlanIsYearly {
			subscriptionEndDate = time.Now().UTC().AddDate(1, 0, 0)
		} else {
			subscriptionEndDate = time.Now().UTC().Add(30 * 24 * time.Hour)
		}

		organization := &model.Organization{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.ID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.ID,
			Name:               req.Name,
			Address:            req.Address,
			Email:              req.Email,
			ContactNumber:      req.ContactNumber,
			Description:        req.Description,
			Color:              req.Color,
			TermsAndConditions: req.TermsAndConditions,
			PrivacyPolicy:      req.PrivacyPolicy,
			CookiePolicy:       req.CookiePolicy,
			RefundPolicy:       req.RefundPolicy,
			UserAgreement:      req.UserAgreement,
			IsPrivate:          req.IsPrivate,
			MediaID:            req.MediaID,
			CoverMediaID:       req.CoverMediaID,

			SubscriptionPlanMaxBranches:         subscription.MaxBranches,
			SubscriptionPlanMaxEmployees:        subscription.MaxEmployees,
			SubscriptionPlanMaxMembersPerBranch: subscription.MaxMembersPerBranch,

			SubscriptionPlanID:    &subscription.ID,
			SubscriptionStartDate: time.Now().UTC(),
			SubscriptionEndDate:   subscriptionEndDate,
		}

		if err := c.model.OrganizationManager.CreateWithTx(context, tx, organization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		var longitude float64 = 0
		var latitude float64 = 0

		branch := &model.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: organization.ID,

			MediaID:       req.MediaID,
			Name:          req.Name,
			Email:         *req.Email,
			Description:   req.Description,
			CountryCode:   "",
			ContactNumber: req.ContactNumber,
			Latitude:      &latitude,
			Longitude:     &longitude,
		}
		if err := c.model.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "something wrong generting developer key"})
		}
		userOrganization := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         organization.ID,
			UserID:                 user.ID,
			BranchID:               &branch.ID,
			UserType:               "owner",
			Description:            "",
			ApplicationDescription: "",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey + uuid.NewString() + "-horizon",
			PermissionName:         "owner",
			PermissionDescription:  "",
			Permissions:            []string{},
		}
		if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, userOrganization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, category := range req.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.CreateWithTx(context, tx, &model.OrganizationCategory{
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, map[string]any{
			"organization":      c.model.OrganizationManager.ToModel(organization),
			"user_organization": c.model.UserOrganizationManager.ToModel(userOrganization),
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization/:organization_id",
		Method:   "PUT",
		Request:  "TOrganization",
		Response: "{organization: TOrganization, user_organization: TUserOrganization}",
		Note:     "(User must be logged in) This will be use to update an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			return err
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}

		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}
		if organization.CreatedByID != user.ID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		organization.Name = req.Name
		organization.Address = req.Address
		organization.Email = req.Email
		organization.ContactNumber = req.ContactNumber
		organization.Description = req.Description
		organization.Color = req.Color
		organization.TermsAndConditions = req.TermsAndConditions
		organization.PrivacyPolicy = req.PrivacyPolicy
		organization.CookiePolicy = req.CookiePolicy
		organization.RefundPolicy = req.RefundPolicy
		organization.UserAgreement = req.UserAgreement
		organization.IsPrivate = req.IsPrivate
		organization.MediaID = req.MediaID
		organization.CoverMediaID = req.CoverMediaID
		organization.UpdatedAt = time.Now().UTC()
		organization.UpdatedByID = user.ID
		if err := c.model.OrganizationManager.UpdateFieldsWithTx(context, tx, organization.ID, organization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		organizationsFromCategory, err := c.model.GetOrganizationCategoryByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		for _, category := range organizationsFromCategory {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}

		for _, category := range req.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.CreateWithTx(context, tx, &model.OrganizationCategory{
				ID:             *category.ID,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModel(organization))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/organization/:organization_id",
		Method: "DELETE",
		Note:   "(User must be logged in) This will be use to DELETE an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}
		currentTime := time.Now().UTC()
		if organization.SubscriptionEndDate.After(currentTime) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
		}
		userOrganizations, err := c.model.GetUserOrganizationByOrganization(context, organization.ID, nil)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if len(userOrganizations) >= 3 {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, category := range organization.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		branches, err := c.model.GetBranchesByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, branch := range branches {
			if err := c.model.OrganizationDestroyer(context, tx, user.ID, *organizationId, branch.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := c.model.OrganizationManager.DeleteByIDWithTx(context, tx, *organizationId); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, userOrganization := range userOrganizations {

			if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) InvitationCode() {
	req := c.provider.Service.Request

	// Retrieve all invitation codes for the current user's organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "GET",
		Response: "IInvitationCode[]",
		Note:     "Retrieves a list of all invitation codes for the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModels(invitationCode))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/search",
		Method:   "GET",
		Request:  "Filter<TInvitationCode>",
		Response: "Paginated<TInvitationCode>",
		Note:     "Get pagination gender",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.Pagination(context, ctx, invitationCode))
	})
	// Retrieve all invitation codes that match a specific code in the current organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/code/:code",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Retrieves invitation code matching the specified code for the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		invitationCode, err := c.model.GetInvitationCodeByCode(context, code)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Retrieve a specific invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/:invitation_code_id",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Retrieves details of a specific invitation code by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Create a new invitation code for the current user's organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "POST",
		Response: "IInvitationCode",
		Request:  "IInvitationCode",
		Note:     "Creates a new invitation code under the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "only owners and employees can perform this action"})
		}
		if req.UserType == "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot create invitation code type owner"})
		}
		data := &model.InvitationCode{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			UserType:       req.UserType,
			Code:           req.Code,
			ExpirationDate: req.ExpirationDate,
			MaxUse:         req.MaxUse,
			CurrentUse:     0,
			Description:    req.Description,
		}

		if err := c.model.InvitationCodeManager.Create(context, data); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(data))
	})

	// Update an existing invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/:invitation_code_id",
		Method:   "PUT",
		Response: "IInvitationCode",
		Request:  "IInvitationCode",
		Note:     "Updates an existing invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		invitationCode.UpdatedAt = time.Now().UTC()
		invitationCode.UpdatedByID = userOrg.UserID
		invitationCode.OrganizationID = userOrg.OrganizationID
		invitationCode.BranchID = *userOrg.BranchID
		invitationCode.UserType = req.UserType
		invitationCode.Code = req.Code
		invitationCode.ExpirationDate = req.ExpirationDate
		invitationCode.MaxUse = req.MaxUse
		invitationCode.Description = req.Description

		if err := c.model.InvitationCodeManager.UpdateFields(context, invitationCode.ID, invitationCode); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Delete a specific invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/invitation-code/:invitation_code_id",
		Method: "DELETE",
		Note:   "Deletes a specific invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		if err := c.model.InvitationCodeManager.DeleteByID(context, *invitationCodeId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/invitation-code/bulk-delete",
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
			invitationCodeId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.InvitationCodeManager.GetByID(context, invitationCodeId); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("InvitationCode with ID %s", rawID))
			}

			if err := c.model.InvitationCodeManager.DeleteByIDWithTx(context, tx, invitationCodeId); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}

func (c *Controller) OrganizationDailyUsage() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/organization-daily-usage",
		Method:   "GET",
		Response: "TOrganizationDailyUsage[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		dailyUsage, err := c.model.GetOrganizationDailyUsageByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationDailyUsageManager.ToModels(dailyUsage))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization-daily-usage/:organization_daily_usage_id",
		Method:   "GET",
		Response: "TOrganizationDailyUsage",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		dailyUsageId, err := horizon.EngineUUIDParam(ctx, "organization_daily_usage_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid Organization daily usage ID")
		}
		dailyUsage, err := c.model.OrganizationDailyUsageManager.GetByIDRaw(context, *dailyUsageId)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, dailyUsage)
	})
}
