package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/model"
)

func (h *Handler) SubscriptionPlanList(c echo.Context) error {
	subscription_plan, err := h.repository.SubscriptionPlanList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.SubscriptionPlanModels(subscription_plan))
}

func (h *Handler) SubscriptionPlanGet(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid subscription_plan ID"})
	}
	subscription_plan, err := h.repository.SubscriptionPlanGetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.SubscriptionPlanModel(subscription_plan))
}

func (h *Handler) SubscriptionPlanCreate(c echo.Context) error {
	req, err := h.model.SubscriptionPlanValidate(c)
	if err != nil {
		return err
	}
	model := &model.SubscriptionPlan{
		Name:                req.Name,
		Description:         req.Description,
		Cost:                req.Cost,
		Timespan:            req.Timespan,
		MaxBranches:         req.MaxBranches,
		MaxEmployees:        req.MaxEmployees,
		MaxMembersPerBranch: req.MaxMembersPerBranch,
		Discount:            req.Discount,
		YearlyDiscount:      req.YearlyDiscount,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := h.repository.SubscriptionPlanCreate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, h.model.SubscriptionPlanModel(model))
}

func (h *Handler) SubscriptionPlanUpdate(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid subscription_plan ID"})
	}
	req, err := h.model.SubscriptionPlanValidate(c)
	if err != nil {
		return err
	}
	model := &model.SubscriptionPlan{
		ID:                  id,
		Name:                req.Name,
		Description:         req.Description,
		Cost:                req.Cost,
		Timespan:            req.Timespan,
		MaxBranches:         req.MaxBranches,
		MaxEmployees:        req.MaxEmployees,
		MaxMembersPerBranch: req.MaxMembersPerBranch,
		Discount:            req.Discount,
		YearlyDiscount:      req.YearlyDiscount,
		UpdatedAt:           time.Now().UTC(),
	}
	if err := h.repository.SubscriptionPlanUpdate(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, h.model.SubscriptionPlanModel(model))

}

func (h *Handler) SubscriptionPlanDelete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid subscription_plan ID"})
	}
	model := &model.SubscriptionPlan{ID: id}
	if err := h.repository.SubscriptionPlanDelete(model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
