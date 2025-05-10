package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

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
	req, err := h.model.UserOrganizationValidate(c)
	if err != nil {
		return err
	}
	codeParam := c.Param("invitation_code")

	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	tx := h.database.Client().Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
	}
	if codeParam != "" {
		ic, err := h.repository.InvitationCodeGetByCode(codeParam)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid invitation code"})
		}
		if ic.OrganizationID != req.OrganizationID || ic.BranchID != req.BranchID {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invitation code not valid for this organization or branch",
			})
		}
		if err := h.repository.InvitationCodeRedeemTransaction(tx, ic); err != nil {
			tx.Rollback()
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
	}
	exists, err := h.repository.UserOrganizationExists(user.ID, req.OrganizationID)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if exists {
		tx.Rollback()
		return c.JSON(http.StatusConflict, map[string]string{"error": "user already in organization"})
	}
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

func (h *Handler) UserOrganizationLeave(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user_organization ID"})
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	uo, err := h.repository.UserOrganizationGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "membership not found"})
	}

	if uo.UserID != user.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "you can only leave your own memberships"})
	}

	switch uo.UserType {
	case "owner", "employee":
		return c.JSON(http.StatusForbidden, map[string]string{"error": "owners and employees cannot leave an organization"})
	}

	if err := h.repository.UserOrganizationDelete(uo); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UserOrganizationUpdate(c echo.Context) error {
	req := new(model.UserOrganizationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	idParam := c.Param("id")
	targetID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid membership ID"})
	}

	me, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}

	myUo, err := h.repository.UserOrganizationGetByUserOrgBranch(
		me.ID, req.OrganizationID, req.BranchID,
	)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "you are not a member of this branch"})
	}

	if myUo.UserType != "owner" && myUo.UserType != "employee" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "only owners or employees can manage members"})
	}

	targetUo, err := h.repository.UserOrganizationGetByID(targetID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "membership not found"})
	}

	if targetUo.OrganizationID != req.OrganizationID || targetUo.BranchID != req.BranchID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "membership does not belong to given org/branch"})
	}

	targetUo.UserType = req.UserType
	targetUo.Description = req.Description
	targetUo.ApplicationDescription = req.ApplicationDescription
	targetUo.ApplicationStatus = req.ApplicationStatus
	targetUo.PermissionName = req.PermissionName
	targetUo.PermissionDescription = req.PermissionDescription
	targetUo.Permissions = req.Permissions
	targetUo.UpdatedAt = time.Now().UTC()
	targetUo.UpdatedByID = me.ID

	if err := h.repository.UserOrganizationUpdate(targetUo); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, h.model.UserOrganizationModel(targetUo))
}
