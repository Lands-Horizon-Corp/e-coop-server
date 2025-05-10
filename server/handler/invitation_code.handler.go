package handler

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (h *Handler) InvitationCodeListByOrgBranch(c echo.Context) error {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	codes, err := h.repository.InvitationCodeListByOrgBranch(orgID, branchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.InvitationCodeModels(codes))
}

func (h *Handler) InvitationCodeCreateByOrgBranch(c echo.Context) error {
	req, err := h.model.InvitationCodeValidate(c)
	if err != nil {
		return err
	}
	user, err := h.provider.CurrentUser(c)
	if err != nil {
		return err
	}
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid organization ID"})
	}
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid branch ID"})
	}
	var code string
	for {
		code = func() string {
			b := make([]byte, 6)
			for i := range b {
				b[i] = codeChars[rand.Intn(len(codeChars))]
			}
			return string(b)
		}()
		exists, err := h.repository.InvitationCodeExists(orgID, branchID, code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if !exists {
			break
		}
	}

	now := time.Now().UTC()
	model := &model.InvitationCode{
		CreatedByID:    user.ID,
		UpdatedByID:    user.ID,
		OrganizationID: orgID,
		BranchID:       branchID,

		UserType:       req.UserType,
		Code:           code,
		ExpirationDate: req.ExpirationDate,
		MaxUse:         req.MaxUse,
		Description:    req.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.repository.InvitationCodeCreate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.InvitationCodeModel(model))
}

func (h *Handler) InvitationCodeDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid invitation_code ID"})
	}
	model := &model.InvitationCode{ID: id}
	if err := h.repository.InvitationCodeDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetInvitationCode(c echo.Context) error {
	code := c.Param("code")
	ic, err := h.repository.InvitationCodeGetByCode(code)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.InvitationCodeModel(ic))
}
