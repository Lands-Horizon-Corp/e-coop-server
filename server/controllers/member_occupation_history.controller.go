package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-occupation-history
func (c *Controller) MemberOccupationHistoryList(ctx echo.Context) error {
	member_occupation_history, err := c.memberOccupationHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryModels(member_occupation_history))
}

// GET /member-occupation-history/:member_occupation_history_id
func (c *Controller) MemberOccupationHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_occupation_history_id")
	if err != nil {
		return err
	}
	member_occupation_history, err := c.memberOccupationHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryModel(member_occupation_history))
}

// DELETE /member-occupation-history/member_occupation_history_id
func (c *Controller) MemberOccupationHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_occupation_history_id")
	if err != nil {
		return err
	}
	if err := c.memberOccupationHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-occupation-history/branch/:branch_id
func (c *Controller) MemberOccupationHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_occupation_history, err := c.memberOccupationHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryModels(member_occupation_history))
}

// GET member-occupation-history/organization/:organization_id
func (c *Controller) MemberOccupationHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_occupation_history, err := c.memberOccupationHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryModels(member_occupation_history))
}

// GET member_occupation_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberOccupationHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_occupation_history, err := c.memberOccupationHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryModels(member_occupation_history))
}
