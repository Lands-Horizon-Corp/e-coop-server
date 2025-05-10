package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) PermissionTemplateList(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	if _, err := h.provider.EnsureEmployeeOrOwner(c, orgID, branchID); err != nil {
		return err
	}
	pts, err := h.repository.PermissionTemplateListByOrgBranch(orgID, branchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.PermissionTemplateModels(pts))
}

func (h *Handler) PermissionTemplateGet(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	if _, err := h.provider.EnsureEmployeeOrOwner(c, orgID, branchID); err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("permission_templates_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid permission_template ID"})
	}
	pt, err := h.repository.PermissionTemplateGetByID(id)
	if err != nil || pt.OrganizationID != orgID || pt.BranchID != branchID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	return c.JSON(http.StatusOK, h.model.PermissionTemplateModel(pt))
}

// POST /organization/:org_id/branch/:branch_id/permission-templates
func (h *Handler) PermissionTemplateCreate(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	uo, err := h.provider.EnsureEmployeeOrOwner(c, orgID, branchID)
	if err != nil {
		return err
	}
	req, err := h.model.PermissionTemplateValidate(c)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	pt := &model.PermissionTemplate{
		OrganizationID: orgID,
		BranchID:       branchID,
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    req.Permissions,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedByID:    uo.UserID,
		UpdatedByID:    uo.UserID,
	}
	if err := h.repository.PermissionTemplateCreate(pt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.PermissionTemplateModel(pt))
}

// PUT /organization/:org_id/branch/:branch_id/permission-templates/:id
func (h *Handler) PermissionTemplateUpdate(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	uo, err := h.provider.EnsureEmployeeOrOwner(c, orgID, branchID)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("permission_templates_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid permission_template ID"})
	}
	req, err := h.model.PermissionTemplateValidate(c)
	if err != nil {
		return err
	}
	pt, err := h.repository.PermissionTemplateGetByID(id)
	if err != nil || pt.OrganizationID != orgID || pt.BranchID != branchID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	pt.Name = req.Name
	pt.Description = req.Description
	pt.Permissions = req.Permissions
	pt.UpdatedAt = time.Now().UTC()
	pt.UpdatedByID = uo.UserID
	if err := h.repository.PermissionTemplateUpdate(pt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.PermissionTemplateModel(pt))
}

// DELETE /organization/:org_id/branch/:branch_id/permission-templates/:id
func (h *Handler) PermissionTemplateDelete(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	if _, err := h.provider.EnsureEmployeeOrOwner(c, orgID, branchID); err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("permission_templates_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid permission_template ID"})
	}
	pt, err := h.repository.PermissionTemplateGetByID(id)
	if err != nil || pt.OrganizationID != orgID || pt.BranchID != branchID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if err := h.repository.PermissionTemplateDelete(pt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
