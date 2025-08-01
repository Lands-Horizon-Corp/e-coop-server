package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberGenderController() {
	req := c.provider.Service.Request

	// Get all member gender history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender-history",
		Method:       "GET",
		ResponseType: model.MemberGenderHistory{},
		Note:         "Returns all member gender history entries for the current user's branch.",
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
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.Filtered(context, ctx, memberGenderHistory))
	})

	// Get member gender history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.MemberGenderHistoryResponse{},
		Note:         "Returns member gender history for a specific member profile ID.",
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
		memberGenderHistory, err := c.model.MemberGenderHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member gender history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGenderHistoryManager.Pagination(context, ctx, memberGenderHistory))
	})

	// Get all member genders for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender",
		Method:       "GET",
		ResponseType: model.MemberGenderResponse{},
		Note:         "Returns all member genders for the current user's branch.",
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
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.Filtered(context, ctx, memberGender))
	})

	// Get paginated member genders
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender/search",
		Method:       "GET",
		ResponseType: model.MemberGenderResponse{},
		Note:         "Returns paginated member genders for the current user's branch.",
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
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender",
		Method:       "POST",
		ResponseType: model.MemberGenderResponse{},
		RequestType:  model.MemberGenderRequest{},
		Note:         "Creates a new member gender record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), validation error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), user org error: " + err.Error(),
				Module:      "MemberGender",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member gender failed (/member-gender), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member gender: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member gender (/member-gender): " + memberGender.Name,
			Module:      "MemberGender",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	// Update an existing member gender by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-gender/:member_gender_id",
		Method:       "PUT",
		ResponseType: model.MemberGenderResponse{},
		RequestType:  model.MemberGenderRequest{},
		Note:         "Updates an existing member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := handlers.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), invalid member_gender_id: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), user org error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.MemberGenderManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), validation error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGender, err := c.model.MemberGenderManager.GetByID(context, *memberGenderID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), not found: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member gender not found: " + err.Error()})
		}
		memberGender.UpdatedAt = time.Now().UTC()
		memberGender.UpdatedByID = user.UserID
		memberGender.OrganizationID = user.OrganizationID
		memberGender.BranchID = *user.BranchID
		memberGender.Name = req.Name
		memberGender.Description = req.Description
		if err := c.model.MemberGenderManager.UpdateFields(context, memberGender.ID, memberGender); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member gender failed (/member-gender/:member_gender_id), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member gender: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member gender (/member-gender/:member_gender_id): " + memberGender.Name,
			Module:      "MemberGender",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberGenderManager.ToModel(memberGender))
	})

	// Delete a member gender by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/member-gender/:member_gender_id",
		Method: "DELETE",
		Note:   "Deletes a member gender record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGenderID, err := handlers.EngineUUIDParam(ctx, "member_gender_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), invalid member_gender_id: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_gender_id: " + err.Error()})
		}
		value, err := c.model.MemberGenderManager.GetByID(context, *memberGenderID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), record not found: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member gender not found: " + err.Error()})
		}
		if err := c.model.MemberGenderManager.DeleteByID(context, *memberGenderID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member gender failed (/member-gender/:member_gender_id), db error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member gender: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member gender (/member-gender/:member_gender_id): " + value.Name,
			Module:      "MemberGender",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member genders by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/member-gender/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member gender records by their IDs.",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete), invalid request body.",
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete), no IDs provided.",
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			memberGenderID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member genders failed (/member-gender/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberGender",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			value, err := c.model.MemberGenderManager.GetByID(context, memberGenderID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member genders failed (/member-gender/bulk-delete), not found: " + rawID,
					Module:      "MemberGender",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member gender with ID '%s' not found: %s", rawID, err.Error())})
			}

			names += value.Name + ","
			if err := c.model.MemberGenderManager.DeleteByIDWithTx(context, tx, memberGenderID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member genders failed (/member-gender/bulk-delete), db error: " + err.Error(),
					Module:      "MemberGender",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member gender with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member genders failed (/member-gender/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberGender",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member genders (/member-gender/bulk-delete): " + names,
			Module:      "MemberGender",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
