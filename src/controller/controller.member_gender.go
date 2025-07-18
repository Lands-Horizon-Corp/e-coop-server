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

	// Get all member gender history for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Returns all member gender history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member gender history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.ToModels(memberGenderHistory))
	})

	// Get member gender history by member profile ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberGenderHistory[]",
		Note:     "Returns member gender history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGenderHistory, err := c.model.MemberGenderHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member gender history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.Pagination(context, ctx, memberGenderHistory))
	})

	// Get all member genders for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "GET",
		Response: "TMemberGender[]",
		Note:     "Returns all member genders for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member genders: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModels(memberGender))
	})

	// Get paginated member genders
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/search",
		Method:   "GET",
		Request:  "Filter<IMemberGender>",
		Response: "Paginated<IMemberGender>",
		Note:     "Returns paginated member genders for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGender, err := c.model.MemberGenderCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member genders for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.Pagination(context, ctx, memberGender))
	})

	// Create a new member gender
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender",
		Method:   "POST",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Creates a new member gender record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberGender := &model.MemberGender{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGenderManager.Create(context, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member gender: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	// Update an existing member gender by ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-gender/:member_gender_id",
		Method:   "PUT",
		Request:  "TMemberGender",
		Response: "TMemberGender",
		Note:     "Updates an existing member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGender, err := c.model.MemberGenderManager.GetByID(context, *memberGenderID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member gender not found: " + err.Error()})
		}
		memberGender.UpdatedAt = time.Now().UTC()
		memberGender.UpdatedByID = user.UserID
		memberGender.OrganizationID = user.OrganizationID
		memberGender.BranchID = *user.BranchID
		memberGender.Name = req.Name
		memberGender.Description = req.Description
		if err := c.model.MemberGenderManager.UpdateFields(context, memberGender.ID, memberGender); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member gender: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	// Delete a member gender by ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-gender/:member_gender_id",
		Method: "DELETE",
		Note:   "Deletes a member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		if err := c.model.MemberGenderManager.DeleteByID(context, *memberGenderID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member gender: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member genders by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/member-gender/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple member gender records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			memberGenderID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			if _, err := c.model.MemberGenderManager.GetByID(context, memberGenderID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member gender with ID '%s' not found: %s", rawID, err.Error())})
			}

			if err := c.model.MemberGenderManager.DeleteByIDWithTx(context, tx, memberGenderID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member gender with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
