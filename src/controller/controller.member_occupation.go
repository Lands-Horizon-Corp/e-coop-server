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

	// Get all member occupation history for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Returns all member occupation history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := c.model.MemberOccupationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.ToModels(memberOccupationHistory))
	})

	// Get member occupation history by member profile ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberOccupationHistory[]",
		Note:     "Returns member occupation history for a specific member profile ID.",
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
		memberOccupationHistory, err := c.model.MemberOccupationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationHistoryManager.Pagination(context, ctx, memberOccupationHistory))
	})

	// Get all member occupations for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "GET",
		Response: "TMemberOccupation[]",
		Note:     "Returns all member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupation, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModels(memberOccupation))
	})

	// Get paginated member occupations
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/search",
		Method:   "GET",
		Request:  "Filter<IMemberOccupation>",
		Response: "Paginated<IMemberOccupation>",
		Note:     "Returns paginated member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.Pagination(context, ctx, value))
	})

	// Create a new member occupation
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation",
		Method:   "POST",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Creates a new member occupation record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberOccupation := &model.MemberOccupation{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberOccupationManager.Create(context, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member occupation: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	// Update an existing member occupation by ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-occupation/:member_occupation_id",
		Method:   "PUT",
		Request:  "TMemberOccupation",
		Response: "TMemberOccupation",
		Note:     "Updates an existing member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.MemberOccupationManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberOccupation, err := c.model.MemberOccupationManager.GetByID(context, *memberOccupationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		memberOccupation.UpdatedAt = time.Now().UTC()
		memberOccupation.UpdatedByID = user.UserID
		memberOccupation.OrganizationID = user.OrganizationID
		memberOccupation.BranchID = *user.BranchID
		memberOccupation.Name = req.Name
		memberOccupation.Description = req.Description
		if err := c.model.MemberOccupationManager.UpdateFields(context, memberOccupation.ID, memberOccupation); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member occupation: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberOccupationManager.ToModel(memberOccupation))
	})

	// Delete a member occupation by ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Deletes a member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		if err := c.model.MemberOccupationManager.DeleteByID(context, *memberOccupationID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member occupation: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member occupations by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/member-occupation/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple member occupation records by their IDs.",
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
			memberOccupationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			if _, err := c.model.MemberOccupationManager.GetByID(context, memberOccupationID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member occupation with ID '%s' not found: %s", rawID, err.Error())})
			}

			if err := c.model.MemberOccupationManager.DeleteByIDWithTx(context, tx, memberOccupationID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member occupation with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
