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

func (c *Controller) MemberGenderController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Get member gender history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.ToModels(memberGenderHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Get member gender history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.Pagination(context, ctx, memberGenderHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "GET",
		Response: "TMemberGender[]",
		Note:     "Get all member gender records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModels(memberGender))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/search",
		Method:   "GET",
		Request:  "Filter<IMemberGender>",
		Response: "Paginated<IMemberGender>",
		Note:     "Get pagination member gender",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.Pagination(context, ctx, memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "POST",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Create a new member gender record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberGender := &model.MemberGender{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGenderManager.Create(context, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/:member_gender_id",
		Method:   "PUT",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Update an existing member gender record by ID",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberGender, err := c.model.MemberGenderManager.GetByID(context, *memberGenderID)
		if err != nil {
			return c.NotFound(ctx, "MemberGender")
		}

		memberGender.UpdatedAt = time.Now().UTC()
		memberGender.UpdatedByID = user.UserID
		memberGender.OrganizationID = user.OrganizationID
		memberGender.BranchID = *user.BranchID
		memberGender.Name = req.Name
		memberGender.Description = req.Description
		if err := c.model.MemberGenderManager.UpdateFields(context, memberGender.ID, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-gender/:member_gender_id",
		Method: "DELETE",
		Note:   "Delete a member gender record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member gender ID")
		}
		if err := c.model.MemberGenderManager.DeleteByID(context, *memberGenderID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-gender/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member gender records by their IDs",
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
			memberGenderID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberGenderManager.GetByID(context, memberGenderID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberGender with ID %s", rawID))
			}

			if err := c.model.MemberGenderManager.DeleteByIDWithTx(context, tx, memberGenderID); err != nil {
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
