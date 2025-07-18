package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberClassificationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification-history",
		Method:   "GET",
		Response: "TMemberClassificationHistory[]",
		Note:     "Get member classification history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassificationHistory, err := c.model.MemberClassificationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryManager.ToModels(memberClassificationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberClassificationHistory[]",
		Note:     "Get member classification history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassificationHistory, err := c.model.MemberClassificationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationHistoryManager.Pagination(context, ctx, memberClassificationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification",
		Method:   "GET",
		Response: "TMemberClassification[]",
		Note:     "Get all member classification records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberClassification, err := c.model.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModels(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification/search",
		Method:   "GET",
		Request:  "Filter<IMemberClassification>",
		Response: "Paginated<IMemberClassification>",
		Note:     "Get pagination member classification",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification",
		Method:   "POST",
		Request:  "TMemberClassification",
		Response: "TMemberClassification",
		Note:     "Create a new member classification record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberClassificationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberClassification := &model.MemberClassification{
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberClassificationManager.Create(context, memberClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-classification/:member_classification_id",
		Method:   "PUT",
		Request:  "TMemberClassification",
		Response: "TMemberClassification",
		Note:     "Update an existing member classification record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberClassificationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberClassification, err := c.model.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			return c.NotFound(ctx, "MemberClassification")
		}

		memberClassification.UpdatedAt = time.Now().UTC()
		memberClassification.UpdatedByID = user.UserID
		memberClassification.OrganizationID = user.OrganizationID
		memberClassification.BranchID = *user.BranchID
		memberClassification.Name = req.Name
		memberClassification.Description = req.Description
		memberClassification.Icon = req.Icon
		if err := c.model.MemberClassificationManager.UpdateFields(context, memberClassification.ID, memberClassification); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberClassificationManager.ToModel(memberClassification))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-classification/:member_classification_id",
		Method: "DELETE",
		Note:   "Delete a member classification record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := horizon.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member classification ID")
		}
		if err := c.model.MemberClassificationManager.DeleteByID(context, *memberClassificationID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-classification/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member classification records by their IDs",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			memberClassificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberClassificationManager.GetByID(context, memberClassificationID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberClassification with ID %s", rawID))
			}

			if err := c.model.MemberClassificationManager.DeleteByIDWithTx(context, tx, memberClassificationID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
