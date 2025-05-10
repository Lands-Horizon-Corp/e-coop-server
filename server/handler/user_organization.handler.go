package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

// list of all organization list by user
func (h *Handler) UserOrganizationList(c echo.Context) error {
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	user_organization, err := h.repository.UserOrganizationListByUserID(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.UserOrganizationModels(user_organization))
}

func (h *Handler) UserOrganizationJoin(c echo.Context) error {
	// 1) bind + validate your request body (org_id, branch_id, etc.)
	req, err := h.model.UserOrganizationValidate(c)
	if err != nil {
		return err
	}

	// 2) pull the code from the URL path, if provided
	codeParam := c.Param("invitation_code")

	// 3) get the logged-in user
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	// 4) start a single transaction
	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}

	// 5) if a code was provided, load it and make sure it belongs to this org+branch
	if codeParam != "" {
		ic, err := h.repository.InvitationCodeGetByCode(codeParam)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid invitation code"})
		}
		// validate the code’s scope
		if ic.OrganizationID != req.OrganizationID || ic.BranchID != req.BranchID {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invitation code not valid for this organization or branch",
			})
		}
		// only now redeem (increment usage)
		if err := h.repository.InvitationCodeRedeemTransaction(tx, ic); err != nil {
			tx.Rollback()
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
	}

	// 6) make sure the user isn’t already in the organization
	exists, err := h.repository.UserOrganizationExists(user.ID, req.OrganizationID)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if exists {
		tx.Rollback()
		return c.JSON(http.StatusConflict, map[string]string{"error": "user already in organization"})
	}

	// 7) create the UserOrganization
	now := time.Now().UTC()
	uo := &model.UserOrganization{
		CreatedAt:              now,
		CreatedByID:            user.ID,
		UpdatedAt:              now,
		UpdatedByID:            user.ID,
		OrganizationID:         req.OrganizationID,
		BranchID:               req.BranchID,
		UserID:                 user.ID,
		UserType:               req.UserType,
		Description:            req.Description,
		ApplicationDescription: req.ApplicationDescription,
		ApplicationStatus:      "accepted",
		DeveloperSecretKey:     h.security.GenerateToken(user.ID.String()),
		PermissionName:         req.UserType,
		PermissionDescription:  "",
		Permissions:            []string{},
	}
	if err := h.repository.UserOrganizationUpdateCreateTransaction(tx, uo); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.UserOrganizationModel(uo))
}
