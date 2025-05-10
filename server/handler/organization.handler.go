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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
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

func (h *Handler) OrganizationSubscribe(c echo.Context) error {
	// 1. Validate request payload (bind path & body)
	req, err := h.model.OrganizationSubscriptionValidate(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// 2. Authenticated user
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	// 3. Load organization
	org, err := h.repository.OrganizationGetByID(req.OrganizationID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "organization not found"})
	}
	if org.CreatedByID != user.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "not authorized to update this organization"})
	}

	// 4. Load new plan
	plan, err := h.repository.SubscriptionPlanGetByID(req.SubscriptionPlanID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "subscription plan not found"})
	}

	// 5. Determine whether this is a yearly or monthly renewal
	useYearly := false
	if req.SubscriptionPlanIsYearly != nil {
		useYearly = *req.SubscriptionPlanIsYearly
	}

	// 6. Compute new start & end dates
	now := time.Now().UTC()
	// if still active, top up from existing end date; else start now
	var newStart time.Time
	if org.SubscriptionEndDate.After(now) {
		newStart = org.SubscriptionEndDate
	} else {
		newStart = now
		org.SubscriptionStartDate = now
	}

	var newEnd time.Time
	if useYearly {
		newEnd = newStart.AddDate(1, 0, 0)
	} else {
		newEnd = newStart.AddDate(0, 1, 0)
	}

	// 7. Apply update in a transaction
	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}

	org.SubscriptionPlanID = &plan.ID
	org.SubscriptionEndDate = newEnd

	// copy over plan limits
	org.SubscriptionPlanMaxBranches = plan.MaxBranches
	org.SubscriptionPlanMaxEmployees = plan.MaxEmployees
	org.SubscriptionPlanMaxMembersPerBranch = plan.MaxMembersPerBranch

	if err := h.repository.OrganizationUpdateCreateTransaction(tx, org); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	// 8. Return updated organization
	return c.JSON(http.StatusOK, h.model.OrganizationModel(org))

}

func (h *Handler) BranchListByOrganization(c echo.Context) error {
	if _, err := h.provider.CurrentUser(c); err != nil {
		return err
	}
	orgIDParam := c.Param("id")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branches, err := h.repository.BranchListByOrganizationID(orgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.BranchModels(branches))
}

func (h *Handler) BranchCreate(c echo.Context) error {
	req, err := h.model.BranchValidate(c)
	if err != nil {
		return err
	}
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	br := &model.Branch{
		CreatedAt:      now,
		CreatedByID:    user.ID,
		UpdatedAt:      now,
		UpdatedByID:    user.ID,
		OrganizationID: orgID,

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
	if err := h.repository.BranchCreate(br); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.BranchModel(br))
}

func (h *Handler) BranchUpdate(c echo.Context) error {
	req, err := h.model.BranchValidate(c)
	if err != nil {
		return err
	}
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	br, err := h.repository.BranchGetByID(branchID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}
	if br.OrganizationID != orgID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch does not belong to this organization"})
	}
	now := time.Now().UTC()
	br.UpdatedAt = now
	br.UpdatedByID = user.ID
	br.MediaID = req.MediaID
	br.Type = req.Type
	br.Name = req.Name
	br.Email = req.Email
	br.Description = req.Description
	br.CountryCode = req.CountryCode
	br.ContactNumber = req.ContactNumber
	br.Address = req.Address
	br.Province = req.Province
	br.City = req.City
	br.Region = req.Region
	br.Barangay = req.Barangay
	br.PostalCode = req.PostalCode
	br.Latitude = req.Latitude
	br.Longitude = req.Longitude
	if err := h.repository.BranchUpdate(br); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.BranchModel(br))
}

func (h *Handler) BranchDelete(c echo.Context) error {
	// 1) parse IDs
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	br, err := h.repository.BranchGetByID(branchID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}
	if br.OrganizationID != orgID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch does not belong to this organization"})
	}
	count, err := h.repository.UserOrganizationCountByOrgBranch(orgID, branchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if count > 2 {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "cannot delete branch with more than 2 members",
		})
	}
	if err := h.repository.BranchDelete(br); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
