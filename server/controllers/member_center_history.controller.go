package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-center-history
func (c *Controller) MemberCenterHistoryList(ctx echo.Context) error {
	member_center_history, err := c.memberCenterHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryModels(member_center_history))
}

// GET /member-center-history/:member_center_history_id
func (c *Controller) MemberCenterHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_center_history_id")
	if err != nil {
		return err
	}
	member_center_history, err := c.memberCenterHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryModel(member_center_history))
}

// DELETE /member-center-history/member_center_history_id
func (c *Controller) MemberCenterHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_center_history_id")
	if err != nil {
		return err
	}
	if err := c.memberCenterHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-center-history/branch/:branch_id
func (c *Controller) MemberCenterHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_center_history, err := c.memberCenterHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryModels(member_center_history))
}

// GET member-center-history/organization/:organization_id
func (c *Controller) MemberCenterHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_center_history, err := c.memberCenterHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryModels(member_center_history))
}

// GET member_center_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberCenterHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_center_history, err := c.memberCenterHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryModels(member_center_history))
}
