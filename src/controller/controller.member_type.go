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

func (c *Controller) MemberTypeController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-history",
		Method:   "GET",
		Response: "TMemberTypeHistory[]",
		Note:     "Get member type history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberTypeHistory, err := c.model.MemberTypeHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryManager.ToModels(memberTypeHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberTypeHistory[]",
		Note:     "Get member type history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberTypeHistory, err := c.model.MemberTypeHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeHistoryManager.Pagination(context, ctx, memberTypeHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type",
		Method:   "GET",
		Response: "TMemberType[]",
		Note:     "Get all member type records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberType, err := c.model.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModels(memberType))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type/search",
		Method:   "GET",
		Request:  "Filter<IMemberType>",
		Response: "Paginated<IMemberType>",
		Note:     "Get pagination member type",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type",
		Method:   "POST",
		Request:  "TMemberType",
		Response: "TMemberType",
		Note:     "Create a new member type record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberType := &model.MemberType{
			Name:           req.Name,
			Description:    req.Description,
			Prefix:         req.Prefix,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberTypeManager.Create(context, memberType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModel(memberType))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-type/:member_type_id",
		Method:   "PUT",
		Request:  "TMemberType",
		Response: "TMemberType",
		Note:     "Update an existing member type record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberTypeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberType, err := c.model.MemberTypeManager.GetByID(context, *memberTypeID)
		if err != nil {
			return c.NotFound(ctx, "MemberType")
		}

		memberType.UpdatedAt = time.Now().UTC()
		memberType.UpdatedByID = user.UserID
		memberType.OrganizationID = user.OrganizationID
		memberType.BranchID = *user.BranchID
		memberType.Name = req.Name
		memberType.Description = req.Description
		memberType.Prefix = req.Prefix
		if err := c.model.MemberTypeManager.UpdateFields(context, memberType.ID, memberType); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeManager.ToModel(memberType))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-type/:member_type_id",
		Method: "DELETE",
		Note:   "Delete a member type record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member type ID")
		}
		if err := c.model.MemberTypeManager.DeleteByID(context, *memberTypeID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-type/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member type records by their IDs",
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
			memberTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberTypeManager.GetByID(context, memberTypeID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberType with ID %s", rawID))
			}

			if err := c.model.MemberTypeManager.DeleteByIDWithTx(context, tx, memberTypeID); err != nil {
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
