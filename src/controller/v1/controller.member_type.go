package controller_v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberTypeController() {
	req := c.provider.Service.Request

	// Get all member type history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-history",
		Method:       "GET",
		ResponseType: modelCore.MemberTypeHistoryResponse{},
		Note:         "Returns all member type history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberTypeHistory, err := c.modelCore.MemberTypeHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member type history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeHistoryManager.Filtered(context, ctx, memberTypeHistory))
	})

	// Get member type history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelCore.MemberTypeHistoryResponse{},
		Note:         "Returns member type history for a specific member profile ID.",
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
		memberTypeHistory, err := c.modelCore.MemberTypeHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member type history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeHistoryManager.Pagination(context, ctx, memberTypeHistory))
	})

	// Get all member types for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type",
		Method:       "GET",
		ResponseType: modelCore.MemberTypeResponse{},
		Note:         "Returns all member types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberType, err := c.modelCore.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeManager.Filtered(context, ctx, memberType))
	})

	// Get paginated member types for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type/search",
		Method:       "GET",
		ResponseType: modelCore.MemberTypeResponse{},
		Note:         "Returns paginated member types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.modelCore.MemberTypeCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeManager.Pagination(context, ctx, value))
	})

	// Create a new member type
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type",
		Method:       "POST",
		RequestType:  modelCore.MemberTypeRequest{},
		ResponseType: modelCore.MemberTypeResponse{},
		Note:         "Creates a new member type record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.MemberTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: validation error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: user org error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberType := &modelCore.MemberType{
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

		if err := c.modelCore.MemberTypeManager.Create(context, memberType); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type failed: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member type: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member type: " + memberType.Name,
			Module:      "MemberType",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeManager.ToModel(memberType))
	})

	// Update an existing member type by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type/:member_type_id",
		Method:       "PUT",
		RequestType:  modelCore.MemberTypeRequest{},
		ResponseType: modelCore.MemberTypeResponse{},
		Note:         "Updates an existing member type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := handlers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: invalid member_type_id: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: user org error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		req, err := c.modelCore.MemberTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: validation error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		memberType, err := c.modelCore.MemberTypeManager.GetByID(context, *memberTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: fmt.Sprintf("Update member type failed: not found (ID: %s): %v", memberTypeID, err),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberType with ID %s not found: %v", memberTypeID, err)})
		}

		memberType.UpdatedAt = time.Now().UTC()
		memberType.UpdatedByID = user.UserID
		memberType.OrganizationID = user.OrganizationID
		memberType.BranchID = *user.BranchID
		memberType.Name = req.Name
		memberType.Description = req.Description
		memberType.Prefix = req.Prefix
		if err := c.modelCore.MemberTypeManager.UpdateFields(context, memberType.ID, memberType); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type failed: update error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member type: " + memberType.Name,
			Module:      "MemberType",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.MemberTypeManager.ToModel(memberType))
	})

	// Delete a member type by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-type/:member_type_id",
		Method: "DELETE",
		Note:   "Deletes a member type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := handlers.EngineUUIDParam(ctx, "member_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type failed: invalid member_type_id: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		memberType, err := c.modelCore.MemberTypeManager.GetByID(context, *memberTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: fmt.Sprintf("Delete member type failed: not found (ID: %s): %v", memberTypeID, err),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberType with ID %s not found: %v", memberTypeID, err)})
		}
		if err := c.modelCore.MemberTypeManager.DeleteByID(context, *memberTypeID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type failed: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member type: " + memberType.Name,
			Module:      "MemberType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member types by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-type/bulk-delete",
		Method:      "DELETE",
		RequestType: modelCore.IDSRequest{},
		Note:        "Deletes multiple member type records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelCore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed: invalid request body.",
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed: no IDs provided.",
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed: begin tx error: " + tx.Error.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		var namesBuilder strings.Builder
		for _, rawID := range reqBody.IDs {
			memberTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member types failed: invalid UUID: " + rawID,
					Module:      "MemberType",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			memberType, err := c.modelCore.MemberTypeManager.GetByID(context, memberTypeID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: fmt.Sprintf("Bulk delete member types failed: not found (ID: %s): %v", rawID, err),
					Module:      "MemberType",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberType with ID %s not found: %v", rawID, err)})
			}

			namesBuilder.WriteString(memberType.Name)
			namesBuilder.WriteString(",")
			if err := c.modelCore.MemberTypeManager.DeleteByIDWithTx(context, tx, memberTypeID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member types failed: delete error: " + err.Error(),
					Module:      "MemberType",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member type with ID %s: %v", rawID, err)})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member types failed: commit tx error: " + err.Error(),
				Module:      "MemberType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member types: " + namesBuilder.String(),
			Module:      "MemberType",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
