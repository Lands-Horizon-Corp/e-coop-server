package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

/*
	Create organization()

SubscriptionPlan
UserOrganization
OrganizationCategory
OrganizationDailyUsage
*/
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

	return c.JSON(http.StatusCreated, organization)
}
