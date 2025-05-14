package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-type-history
func (c *Controller) MemberTypeHistoryList(ctx echo.Context) error {
	member_type_history, err := c.memberTypeHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryModels(member_type_history))
}

// GET /member-type-history/:member_type_history_id
func (c *Controller) MemberTypeHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_type_history_id")
	if err != nil {
		return err
	}
	member_type_history, err := c.memberTypeHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryModel(member_type_history))
}

// DELETE /member-type-history/member_type_history_id
func (c *Controller) MemberTypeHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_type_history_id")
	if err != nil {
		return err
	}
	if err := c.memberTypeHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-type-history/branch/:branch_id
func (c *Controller) MemberTypeHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_type_history, err := c.memberTypeHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryModels(member_type_history))
}

// GET member-type-history/organization/:organization_id
func (c *Controller) MemberTypeHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_type_history, err := c.memberTypeHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryModels(member_type_history))
}

// GET member_type_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberTypeHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_type_history, err := c.memberTypeHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryModels(member_type_history))
}
