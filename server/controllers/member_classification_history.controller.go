package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-classification-history
func (c *Controller) MemberClassificationHistoryList(ctx echo.Context) error {
	member_classification_history, err := c.memberClassificationHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryModels(member_classification_history))
}

// GET /member-classification-history/:member_classification_history_id
func (c *Controller) MemberClassificationHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_classification_history_id")
	if err != nil {
		return err
	}
	member_classification_history, err := c.memberClassificationHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryModel(member_classification_history))
}

// DELETE /member-classification-history/member_classification_history_id
func (c *Controller) MemberClassificationHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_classification_history_id")
	if err != nil {
		return err
	}
	if err := c.memberClassificationHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-classification-history/branch/:branch_id
func (c *Controller) MemberClassificationHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_classification_history, err := c.memberClassificationHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryModels(member_classification_history))
}

// GET member-classification-history/organization/:organization_id
func (c *Controller) MemberClassificationHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_classification_history, err := c.memberClassificationHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryModels(member_classification_history))
}

// GET member_classification_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberClassificationHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_classification_history, err := c.memberClassificationHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryModels(member_classification_history))
}
