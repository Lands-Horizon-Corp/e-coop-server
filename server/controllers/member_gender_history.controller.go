package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-gender-history
func (c *Controller) MemberGenderHistoryList(ctx echo.Context) error {
	member_gender_history, err := c.memberGenderHistory.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryModels(member_gender_history))
}

// GET /member-gender-history/:member_gender_history_id
func (c *Controller) MemberGenderHistoryGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_gender_history_id")
	if err != nil {
		return err
	}
	member_gender_history, err := c.memberGenderHistory.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryModel(member_gender_history))
}

// DELETE /member-gender-history/member_gender_history_id
func (c *Controller) MemberGenderHistoryDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_gender_history_id")
	if err != nil {
		return err
	}
	if err := c.memberGenderHistory.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-gender-history/branch/:branch_id
func (c *Controller) MemberGenderHistoryListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_gender_history, err := c.memberGenderHistory.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryModels(member_gender_history))
}

// GET member-gender-history/organization/:organization_id
func (c *Controller) MemberGenderHistoryListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_gender_history, err := c.memberGenderHistory.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryModels(member_gender_history))
}

// GET member_gender_history/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberGenderHistoryListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_gender_history, err := c.memberGenderHistory.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryModels(member_gender_history))
}
