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

func (c *Controller) MemberOccupationController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Get member occupation history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupationHistory, err := c.model.MemberOccupationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.ToModels(memberOccupationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Get member occupation history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupationHistory, err := c.model.MemberOccupationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.Pagination(context, ctx, memberOccupationHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "GET",
		Response: "TMemberOccupation[]",
		Note:     "Get all member occupation records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberOccupation, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModels(memberOccupation))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/search",
		Method:   "GET",
		Request:  "Filter<IMemberOccupation>",
		Response: "Paginated<IMemberOccupation>",
		Note:     "Get pagination member occupation",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "POST",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Create a new member occupation record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberOccupation := &model.MemberOccupation{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberOccupationManager.Create(context, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/:member_occupation_id",
		Method:   "PUT",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Update an existing member occupation record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberOccupation, err := c.model.MemberOccupationManager.GetByID(context, *memberOccupationID)
		if err != nil {
			return c.NotFound(ctx, "MemberOccupation")
		}

		memberOccupation.UpdatedAt = time.Now().UTC()
		memberOccupation.UpdatedByID = user.UserID
		memberOccupation.OrganizationID = user.OrganizationID
		memberOccupation.BranchID = *user.BranchID
		memberOccupation.Name = req.Name
		memberOccupation.Description = req.Description
		if err := c.model.MemberOccupationManager.UpdateFields(context, memberOccupation.ID, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Delete a member occupation record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member occupation ID")
		}
		if err := c.model.MemberOccupationManager.DeleteByID(context, *memberOccupationID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-occupation/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member occupation records by their IDs",
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
			memberOccupationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberOccupationManager.GetByID(context, memberOccupationID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberOccupation with ID %s", rawID))
			}

			if err := c.model.MemberOccupationManager.DeleteByIDWithTx(context, tx, memberOccupationID); err != nil {
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
