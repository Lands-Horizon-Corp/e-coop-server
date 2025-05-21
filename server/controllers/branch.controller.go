package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /branch
func (c *Controller) BranchList(ctx echo.Context) error {
	branch, err := c.branch.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.BranchModels(branch))
}

// GET /branch/branch_id
func (c *Controller) BranchGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	branch, err := c.branch.Manager.GetByID(*id)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, c.model.BranchModel(branch))
}

// POST /branch/user-organization/:user_organization_id
func (c *Controller) BranchCreate(ctx echo.Context) error {
	req, err := c.model.BranchValidate(ctx)
	if err != nil {
		return err
	}
	userOrganizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	userOrganization, err := c.userOrganization.Manager.GetByID(*userOrganizationId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "User organization doesn't exist"})
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
	}

	tx := c.database.Client().Begin()
	if tx.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}

	if err := c.branch.Manager.CreateWithTx(tx, branch); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if userOrganization.BranchID == nil {
		// Update existing userOrganization with the new branch ID
		userOrganization.BranchID = &branch.ID
		userOrganization.UpdatedAt = time.Now().UTC()
		userOrganization.UpdatedByID = user.ID

		if err := c.userOrganization.Manager.UpdateFieldsWithTx(tx, userOrganization.ID, userOrganization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	} else {
		// Create a new userOrganization linked to the new branch
		userOrganizationModel := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         userOrganization.OrganizationID,
			BranchID:               &branch.ID,
			UserID:                 user.ID,
			UserType:               "owner",
			Description:            "",
			ApplicationDescription: "",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     c.security.GenerateToken(user.ID.String()),
			PermissionName:         "owner",
			PermissionDescription:  "",
			Permissions:            []string{},
		}

		if err := c.userOrganization.Manager.CreateWithTx(tx, userOrganizationModel); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"branch":            c.model.BranchModel(branch),
		"user_organization": c.model.UserOrganizationModel(userOrganization),
	})
}

// PUT /branch/user-organization/:user_organization_id
func (c *Controller) BranchUpdate(ctx echo.Context) error {
	req, err := c.model.BranchValidate(ctx)
	if err != nil {
		return err
	}
	userOrganizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return err
	}
	userOrganization, err := c.userOrganization.Manager.GetByID(*userOrganizationId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "User organization doesnt exists"})
	}
	user, err := c.provider.UserOwner(ctx, userOrganization.OrganizationID.String(), userOrganization.OrganizationID.String())
	if err != nil {
		return err
	}
	branch := &model.Branch{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		OrganizationID: user.OrganizationID,
		MediaID:        req.MediaID,
		Type:           req.Type,
		Name:           req.Name,
		Email:          req.Email,
		Description:    req.Description,
		CountryCode:    req.CountryCode,
		ContactNumber:  req.ContactNumber,
		Address:        req.Address,
		Province:       req.Province,
		City:           req.City,
		Region:         req.Region,
		Barangay:       req.Barangay,
		PostalCode:     req.PostalCode,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
	}
	if err := c.branch.Manager.UpdateByID(userOrganization.OrganizationID, branch); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, map[string]any{
		"branch":            c.model.BranchModel(branch),
		"user_organization": c.model.UserOrganizationModel(userOrganization),
	})
}

// DELETE /branch/:branch_id/:user_organization_id
func (c *Controller) BranchDelete(ctx echo.Context) error {
	organizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	br, err := c.branch.Manager.GetByID(*branchId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}
	_, err = c.provider.UserOwner(ctx, organizationId.String(), branchId.String())
	if err != nil {
		return err
	}
	if br.OrganizationID != *organizationId {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "branch does not belong to this organization"})
	}
	count, err := c.userOrganization.CountByOrganizationBranch(*organizationId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if count > 2 {
		return ctx.JSON(http.StatusForbidden, map[string]string{
			"error": "cannot delete branch with more than 2 members",
		})
	}
	if err := c.branch.Manager.DeleteByID(*branchId); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET /branch/organization/:organization_id
func (c *Controller) BranchOrganizations(ctx echo.Context) error {
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branch, err := c.branch.ByOrganizations(*organizationId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.BranchModels(branch))
}
