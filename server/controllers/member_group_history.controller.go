package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-group-history
func (c *Controller) MemberGroupHistoryList(ctx echo.Context) error {
	member_group_history, err := c.memberGroupHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryModels(member_group_history))
}

// GET /member-group-history/:member_group_history_id
func (c *Controller) MemberGroupHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_group_history_id")
	if err != nil {
		return err
	}
	member_group_history, err := c.memberGroupHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryModel(member_group_history))
}

// DELETE /member-group-history/member_group_history_id
func (c *Controller) MemberGroupHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_group_history_id")
	if err != nil {
		return err
	}
	if err := c.memberGroupHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-group-history/branch/:branch_id
func (c *Controller) MemberGroupHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_group_history, err := c.memberGroupHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryModels(member_group_history))
}

// GET member-group-history/organization/:organization_id
func (c *Controller) MemberGroupHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_group_history, err := c.memberGroupHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryModels(member_group_history))
}

// GET member_group_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberGroupHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_group_history, err := c.memberGroupHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryModels(member_group_history))
}
