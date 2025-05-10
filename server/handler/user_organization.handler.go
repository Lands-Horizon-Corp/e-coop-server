package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
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
