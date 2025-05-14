package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-classification
func (c *Controller) MemberClassificationList(ctx echo.Context) error {
	member_classification, err := c.memberClassification.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationModels(member_classification))
}

// GET /member-classification/:member_classification_id
func (c *Controller) MemberClassificationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
	if err != nil {
		return err
	}
	member_classification, err := c.memberClassification.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationModel(member_classification))
}

// POST /member-classification
func (c *Controller) MemberClassificationCreate(ctx echo.Context) error {
	req, err := c.model.MemberClassificationValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberClassification{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}
	if err := c.memberClassification.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-classification", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberClassificationModel(model))
}

// PUT /member-classification/member_classification_id
func (c *Controller) MemberClassificationUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberClassificationValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberClassification{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	}
	if err := c.memberClassification.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-classification", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberClassificationModel(model))
}

// DELETE /member-classification/member_classification_id
func (c *Controller) MemberClassificationDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
	if err != nil {
		return err
	}
	if err := c.memberClassification.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-classification/branch/:branch_id
func (c *Controller) MemberClassificationListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_classification, err := c.memberClassification.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationModels(member_classification))
}

// GET member-classification/organization/:organization_id
func (c *Controller) MemberClassificationListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_classification, err := c.memberClassification.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationModels(member_classification))
}

// GET member_classification/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberClassificationListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_classification, err := c.memberClassification.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberClassificationModels(member_classification))
}
