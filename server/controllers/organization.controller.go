package controllers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /organization
func (c *Controller) OrganizationList(ctx echo.Context) error {
	organization, err := c.organization.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationModels(organization))
}

// GET /organization/:organization_id
func (c *Controller) OrganizationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	organization, err := c.organization.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationModel(organization))
}

// POST /organization
func (c *Controller) OrganizationCreate(ctx echo.Context) error {
	req, err := c.model.OrganizationValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}
	subscription, err := c.subscriptionPlan.Manager.GetByID(*req.SubscriptionPlanID)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Subscription plan not found"})
	}
	tx := c.database.Client().Begin()
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
	if err := c.organization.Manager.CreateWithTx(tx, organization); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	userOrganization := &model.UserOrganization{
		CreatedAt:              time.Now().UTC(),
		CreatedByID:            user.ID,
		UpdatedAt:              time.Now().UTC(),
		UpdatedByID:            user.ID,
		OrganizationID:         organization.ID,
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
	if err := c.userOrganization.Manager.CreateWithTx(tx, userOrganization); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	for _, category := range req.OrganizationCategories {

		id, err := uuid.Parse(category.CategoryID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.organizationCategory.Manager.CreateWithTx(tx, &model.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     &id,
		}); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationModel(organization))

}

// PUT /organization/:organization_id
func (c *Controller) OrganizationUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	req, err := c.model.OrganizationValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUser(ctx)
	if err != nil {
		return err
	}

	organization, err := c.organization.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
	}
	if organization.CreatedByID != user.ID {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
	}

	tx := c.database.Client().Begin()
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

	if err := c.organization.Manager.UpdateByIDWithTx(tx, organization.ID, organization); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	for _, category := range req.OrganizationCategories {
		id, err := uuid.Parse(category.CategoryID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.organizationCategory.Manager.UpsertWithTx(tx, &model.OrganizationCategory{
			ID:             horizon.ParseUUID(category.ID),
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     &id,
		}); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationModel(organization))

}

// DELETE /organization/:organization
func (c *Controller) OrganizationDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}

	organization, err := c.organization.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
	}

	currentTime := time.Now().UTC()
	if organization.SubscriptionEndDate.After(currentTime) {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
	}

	userOrganizationsCount, err := c.userOrganization.CountByOrganization(organization.ID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if userOrganizationsCount >= 3 {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
	}
	tx := c.database.Client().Begin()
	if tx.Error != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}

	for _, category := range organization.OrganizationCategories {
		if err := c.organizationCategory.Manager.DeleteByIDWithTx(tx, category.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	if err := c.organization.Manager.DeleteByIDWithTx(tx, *id); err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
