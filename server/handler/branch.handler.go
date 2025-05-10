package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

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

func (h *Handler) BranchGet(c echo.Context) error {
	// 1. Parse org and branch IDs
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}

	// 2. Fetch the branch
	br, err := h.repository.BranchGetByID(branchID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}

	// 3. Ensure it belongs to the org
	if br.OrganizationID != orgID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch does not belong to this organization"})
	}

	// 4. Return it
	return c.JSON(http.StatusOK, h.model.BranchModel(br))
}
