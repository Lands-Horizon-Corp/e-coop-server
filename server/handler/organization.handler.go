package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) OrganizationList(c echo.Context) error {
	organizations, err := h.repository.OrganizationList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.OrganizationModels(organizations))
}

func (h *Handler) OrganizationGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	organization, err := h.repository.OrganizationGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.OrganizationModel(organization))
}

func (h *Handler) OrganizationCreate(c echo.Context) error {
	req, err := h.model.OrganizationValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	subscription, err := h.repository.SubscriptionPlanGetByID(*req.SubscriptionPlanID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Subscription plan not found"})
	}

	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}
	var subscriptionEndDate time.Time
	if req.SubscriptionPlanIsYearly {
		subscriptionEndDate = time.Now().UTC().AddDate(1, 0, 0)
	} else {
		subscriptionEndDate = time.Now().UTC().Add(30 * 24 * time.Hour)
	}
	organization := &model.Organization{
		CreatedAt:             time.Now().UTC(),
		CreatedByID:           user.ID,
		UpdatedAt:             time.Now().UTC(),
		UpdatedByID:           user.ID,
		Name:                  req.Name,
		Address:               req.Address,
		Email:                 req.Email,
		ContactNumber:         req.ContactNumber,
		Description:           req.Description,
		Color:                 req.Color,
		TermsAndConditions:    req.TermsAndConditions,
		PrivacyPolicy:         req.PrivacyPolicy,
		CookiePolicy:          req.CookiePolicy,
		RefundPolicy:          req.RefundPolicy,
		UserAgreement:         req.UserAgreement,
		IsPrivate:             req.IsPrivate,
		MediaID:               req.MediaID,
		CoverMediaID:          req.CoverMediaID,
		SubscriptionPlanID:    &subscription.ID,
		SubscriptionStartDate: time.Now().UTC(),
		SubscriptionEndDate:   subscriptionEndDate,
	}
	if err := h.repository.OrganizationUpdateCreateTransaction(tx, organization); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		DeveloperSecretKey:     h.security.GenerateToken(user.ID.String()),
		PermissionName:         "owner",
		PermissionDescription:  "",
		Permissions:            []string{},
	}
	if err := h.repository.UserOrganizationUpdateCreateTransaction(tx, userOrganization); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	for _, category := range req.OrganizationCategories {
		if err := h.repository.OrganizationCategoryUpdateCreateTransaction(tx, &model.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     category,
		}); err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, h.model.OrganizationModel(organization))
}

func (h *Handler) OrganizationUpdate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid footstep ID"})
	}
	req, err := h.model.OrganizationValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	organization, err := h.repository.OrganizationGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
	}
	if organization.CreatedByID != user.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
	}
	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
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
	if err := h.repository.OrganizationUpdateCreateTransaction(tx, organization); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	for _, category := range organization.OrganizationCategories {
		if err := h.repository.OrganizationCategoryDeleteTransaction(tx, category); err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	for _, category := range req.OrganizationCategories {
		if err := h.repository.OrganizationCategoryUpdateCreateTransaction(tx, &model.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     category,
		}); err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.OrganizationModel(organization))
}

func (h *Handler) OrganizationDelete(c echo.Context) error {

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}

	_, err = h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	organization, err := h.repository.OrganizationGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
	}

	currentTime := time.Now().UTC()
	if organization.SubscriptionEndDate.After(currentTime) {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
	}

	userOrganizationsCount, err := h.repository.UserOrganizationsCount(organization.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if userOrganizationsCount >= 3 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
	}

	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}

	if err := h.repository.OrganizationDeleteTransaction(tx, organization); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
