package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberOccupationController() {
	req := c.provider.Service.Request

	// Get all member occupation history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation-history",
		Method:       "GET",
		ResponseType: model_core.MemberOccupationHistoryResponse{},
		Note:         "Returns all member occupation history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := c.model_core.MemberOccupationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationHistoryManager.Filtered(context, ctx, memberOccupationHistory))
	})

	// Get member occupation history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model_core.MemberOccupationHistoryResponse{},
		Note:         "Returns member occupation history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupationHistory, err := c.model_core.MemberOccupationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupation history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationHistoryManager.Pagination(context, ctx, memberOccupationHistory))
	})

	// Get all member occupations for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "GET",
		ResponseType: model_core.MemberOccupationResponse{},
		Note:         "Returns all member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberOccupation, err := c.model_core.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationManager.Filtered(context, ctx, memberOccupation))
	})

	// Get paginated member occupations
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation/search",
		Method:       "GET",
		ResponseType: model_core.MemberOccupationResponse{},
		Note:         "Returns paginated member occupations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model_core.MemberOccupationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member occupations for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationManager.Pagination(context, ctx, value))
	})

	// Create a new member occupation
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation",
		Method:       "POST",
		ResponseType: model_core.MemberOccupationResponse{},
		RequestType:  model_core.MemberOccupationRequest{},
		Note:         "Creates a new member occupation record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model_core.MemberOccupationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberOccupation := &model_core.MemberOccupation{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model_core.MemberOccupationManager.Create(context, memberOccupation); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member occupation failed (/member-occupation), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member occupation: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member occupation (/member-occupation): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})

		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationManager.ToModel(memberOccupation))
	})

	// Update an existing member occupation by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-occupation/:member_occupation_id",
		Method:       "PUT",
		ResponseType: model_core.MemberOccupationResponse{},
		RequestType:  model_core.MemberOccupationRequest{},
		Note:         "Updates an existing member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := handlers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), user org error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model_core.MemberOccupationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), validation error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberOccupation, err := c.model_core.MemberOccupationManager.GetByID(context, *memberOccupationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), not found: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		memberOccupation.UpdatedAt = time.Now().UTC()
		memberOccupation.UpdatedByID = user.UserID
		memberOccupation.OrganizationID = user.OrganizationID
		memberOccupation.BranchID = *user.BranchID
		memberOccupation.Name = req.Name
		memberOccupation.Description = req.Description
		if err := c.model_core.MemberOccupationManager.UpdateFields(context, memberOccupation.ID, memberOccupation); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member occupation: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member occupation (/member-occupation/:member_occupation_id): " + memberOccupation.Name,
			Module:      "MemberOccupation",
		})
		return ctx.JSON(http.StatusOK, c.model_core.MemberOccupationManager.ToModel(memberOccupation))
	})

	// Delete a member occupation by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-occupation/:member_occupation_id",
		Method: "DELETE",
		Note:   "Deletes a member occupation record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberOccupationID, err := handlers.EngineUUIDParam(ctx, "member_occupation_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), invalid member_occupation_id: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_occupation_id: " + err.Error()})
		}
		value, err := c.model_core.MemberOccupationManager.GetByID(context, *memberOccupationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), record not found: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member occupation not found: " + err.Error()})
		}
		if err := c.model_core.MemberOccupationManager.DeleteByID(context, *memberOccupationID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member occupation failed (/member-occupation/:member_occupation_id), db error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member occupation: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member occupation (/member-occupation/:member_occupation_id): " + value.Name,
			Module:      "MemberOccupation",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member occupations by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-occupation/bulk-delete",
		Method:      "DELETE",
		RequestType: model_core.IDSRequest{},
		Note:        "Deletes multiple member occupation records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model_core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), invalid request body.",
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), no IDs provided.",
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			memberOccupationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberOccupation",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			value, err := c.model_core.MemberOccupationManager.GetByID(context, memberOccupationID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), not found: " + rawID,
					Module:      "MemberOccupation",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member occupation with ID '%s' not found: %s", rawID, err.Error())})
			}

			names += value.Name + ","
			if err := c.model_core.MemberOccupationManager.DeleteByIDWithTx(context, tx, memberOccupationID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), db error: " + err.Error(),
					Module:      "MemberOccupation",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member occupation with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member occupations failed (/member-occupation/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberOccupation",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member occupations (/member-occupation/bulk-delete): " + names,
			Module:      "MemberOccupation",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
