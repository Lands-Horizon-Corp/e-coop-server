package controllers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// Get All
func (c *Controller) UserOrganizationGetAll(ctx echo.Context) error {
	userOrganization, err := c.userOrganization.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModels(userOrganization))
}

// GET /user-organization/:user_organization_id
func (c *Controller) UserOrganizationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return err
	}
	userOrganization, err := c.userOrganization.Manager.GetByID(*id)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrganization))
}

// POST user-organization/:organization_id/seed
func (c *Controller) UserOrganizationSeeder(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization ID")
	}

	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to get current user")
	}

	userOrganizations, err := c.userOrganization.ListByOrganization(*orgId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch user organization list")
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

		if _, err := c.memberClassification.Seeder(user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed classifications")
		}
		if _, err := c.memberGender.Seeder(user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed genders")
		}
		if _, err := c.memberGroup.Seeder(user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed groups")
		}
		if _, err := c.memberOccupation.Seeder(user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed occupations")
		}
		if _, err := c.memberType.Seeder(user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed member types")
		}

		userOrganization.IsSeeded = true
		if err := c.userOrganization.Manager.Update(userOrganization); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user organization seed status")
		}
		c.provider.Notification(ctx, "Organization creation", "Finished setting up your Organization", "info")
	}

	return ctx.NoContent(http.StatusOK)
}

// GET  user-organization/unswitch
func (c *Controller) UserOrganizationUnSwitch(ctx echo.Context) error {
	c.provider.CleanCustomToken(ctx)
	return ctx.NoContent(http.StatusOK)
}

// GET  user-organization/:user_organization_id/switch
func (c *Controller) UserOrganizationSwitch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		c.provider.CleanCustomToken(ctx)
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	userOrganization, err := c.userOrganization.Manager.GetByID(*id)
	if err != nil {
		return err
	}
	if user.ID != userOrganization.UserID {
		return echo.NewHTTPError(http.StatusNotAcceptable, "the user is not part of the organization")
	}
	if err := c.provider.SetCustom(ctx, userOrganization); err != nil {
		return echo.NewHTTPError(http.StatusNotAcceptable, "failed to set authentication token")
	}

	return ctx.JSON(http.StatusOK, model.CurrentUserResponse{
		UserID:           user.ID,
		User:             c.model.UserModel(user),
		UserOrganization: c.model.UserOrganizationModel(userOrganization),
	})
}

// PUT /user-organization/:user_organization_id/developer-key-refresh
func (c *Controller) UserOrganizationRegenerateDeveloperKey(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_organization_id")
	}

	model, err := c.userOrganization.Manager.GetByID(*id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user organization not found")
	}

	currentUserOrg, err := c.provider.UserOwnerEmployee(ctx, model.OrganizationID.String(), model.BranchID.String())
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized access")
	}

	if currentUserOrg.ID != model.ID && currentUserOrg.UserType != "owner" {
		return echo.NewHTTPError(http.StatusForbidden, "cannot refresh developer key")
	}

	regenKey := uuid.New().String()
	newToken := c.security.GenerateToken(regenKey)
	model.DeveloperSecretKey = newToken

	if err := c.userOrganization.Manager.Update(model); err != nil {
		return echo.NewHTTPError(http.StatusNotAcceptable, err.Error())
	}

	c.provider.Notification(ctx, "Security Update", "Developer key regenerated successfully", "info")
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(model))
}

// PUT user-organization/:user_organization_id
func (c *Controller) UserOrganizationUpdate(ctx echo.Context) error {
	userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return err
	}

	req, err := c.model.UserOrganizationValidate(ctx)
	if err != nil {
		return err
	}

	model, err := c.userOrganization.Manager.GetByID(*userOrgId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "membership not found"})
	}

	model.UserType = req.UserType
	model.Description = req.Description
	model.ApplicationDescription = req.ApplicationDescription
	model.ApplicationStatus = req.ApplicationStatus
	model.PermissionName = req.PermissionName
	model.PermissionDescription = req.PermissionDescription
	model.Permissions = req.Permissions
	model.UpdatedAt = time.Now().UTC()

	model.UserSettingDescription = req.UserSettingDescription
	model.UserSettingStartOR = req.UserSettingStartOR
	model.UserSettingEndOR = req.UserSettingEndOR
	model.UserSettingUsedOR = req.UserSettingUsedOR
	model.UserSettingStartVoucher = req.UserSettingStartVoucher
	model.UserSettingEndVoucher = req.UserSettingEndVoucher
	model.UserSettingUsedVoucher = req.UserSettingUsedVoucher

	currentUserOrg, err := c.provider.UserOwnerEmployee(ctx, model.OrganizationID.String(), model.BranchID.String())
	if err != nil {
		return err
	}

	model.UpdatedByID = currentUserOrg.UserID
	if err := c.userOrganization.Manager.Update(model); err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}

	c.provider.Notification(ctx, "Profile Update", "User organization updated successfully", "info")
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(model))
}

func (c *Controller) UserOrganizationJoin(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	req, err := c.model.UserOrganizationValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}

	if req.ApplicationStatus == "member" {
		if !c.userOrganization.MemberCanJoin(user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as member"})
		}
	}

	if req.ApplicationStatus == "employee" {
		if !c.userOrganization.EmployeeCanJoin(user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as employee"})
		}
	}

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
		ApplicationDescription: req.ApplicationDescription,
		ApplicationStatus:      "pending",
		DeveloperSecretKey:     c.security.GenerateToken(user.ID.String()),
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

	if err := c.userOrganization.Manager.Create(userOrg); err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}

	c.provider.Notification(ctx, "Membership", "Successfully joined organization", "success")
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrg))
}

// POST user-organization/join/invitation-code/:code
func (c *Controller) UserOrganizationJoinByCode(ctx echo.Context) error {
	code := ctx.Param("code")
	exists, err := c.invitationCode.Exists(code)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{
			"error": "failed to check invitation code",
		})
	}
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "invitation code not found",
		})
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}

	invitationCode, err := c.invitationCode.Verify(code)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": "failed to retrieve invitation code"})
	}

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
		DeveloperSecretKey:     c.security.GenerateToken(user.ID.String()),
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

	tx := c.database.Client().Begin()
	_, err = c.invitationCode.Redeem(tx, code)
	if err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	if err := c.userOrganization.Manager.CreateWithTx(tx, userOrg); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}

	c.provider.Notification(ctx, "Membership", "Joined organization using invitation code", "success")
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrg))
}

// POST user-organization/leave/organization/:organization_id/branch/:branch_id
func (c *Controller) UserOrganizationLeave(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	userOrg, err := c.provider.UserOrganization(ctx, orgId.String(), branchId.String())
	if err != nil {
		return err
	}

	switch userOrg.UserType {
	case "owner", "employee":
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "owners and employees cannot leave an organization"})
	}

	if err := c.userOrganization.Manager.DeleteByID(userOrg.ID); err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}

	c.provider.Notification(ctx, "Membership", "Successfully left organization", "info")
	return ctx.NoContent(http.StatusNoContent)
}

// GET user-organization/user/:user_id
func (c *Controller) UserOrganizationListByUser(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByUser(*userId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModels(userOrg))
}

// GET user-organization/branch/:branch_id
func (c *Controller) UserOrganizationListByBranch(ctx echo.Context) error {
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByBranch(*branchId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModels(userOrg))
}

// GET user-organization/organization/:organization_id
func (c *Controller) UserOrganizationListByOrganization(ctx echo.Context) error {
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByOrganization(*organizationId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModels(userOrg))
}

// GET user-organization/organization/:organization_id/branch/:branch_id
func (c *Controller) UserOrganizationListByOrganizationBranch(ctx echo.Context) error {
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByOrganizationBranch(*organizationId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrg))
}

// GET user-organization/user/:user_id/organization/:organization_id
func (c *Controller) UserOrganizationListByUserOrganization(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByUserOrganization(*userId, *organizationId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModels(userOrg))
}

// GET user-organization/user/:user_id/organization/:organization_id/branch/:branch_id
func (c *Controller) UserOrganizationByUserOrganizationBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}

	userOrg, err := c.userOrganization.ByUserOrganizationBranch(*userId, *organizationId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrg))
}

// GET user-organization/organization/:organization_id/branch/:branch_id/can-join-member
func (c *Controller) UserOrganizationCanJoinMember(ctx echo.Context) error {
	user, err := c.provider.CurrentUser(ctx)
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

	if !c.userOrganization.MemberCanJoin(user.ID, *orgId, *branchId) {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as member"})
	}

	return ctx.NoContent(http.StatusOK)
}

// GET user-organization/organization/:organization_id/branch/:branch_id/can-join-employee
func (c *Controller) UserOrganizationCanJoinEmployee(ctx echo.Context) error {
	user, err := c.provider.CurrentUser(ctx)
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
	if !c.userOrganization.EmployeeCanJoin(user.ID, *orgId, *branchId) {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "cannot join as employee"})
	}
	return ctx.NoContent(http.StatusOK)
}

// GET user-organization/user/:user_id/branch/:branch_id
func (c *Controller) UserOrganizationListByUserBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	userOrg, err := c.userOrganization.ListByUserBranch(*userId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.UserOrganizationModel(userOrg))
}
