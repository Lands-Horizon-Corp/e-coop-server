package controller_v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberClassificationController() {
	req := c.provider.Service.Request

	// Get all member classification history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification-history",
		Method:       "GET",
		ResponseType: modelcore.MemberClassificationHistoryResponse{},
		Note:         "Returns all member classification history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassificationHistory, err := c.modelcore.MemberClassificationHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationHistoryManager.Filtered(context, ctx, memberClassificationHistory))
	})

	// Get member classification history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelcore.MemberClassificationHistoryResponse{},
		Note:         "Returns member classification history for a specific member profile ID.",
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
		memberClassificationHistory, err := c.modelcore.MemberClassificationHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classification history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationHistoryManager.Pagination(context, ctx, memberClassificationHistory))
	})

	// Get all member classifications for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification",
		Method:       "GET",
		ResponseType: modelcore.MemberClassificationResponse{},
		Note:         "Returns all member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberClassification, err := c.modelcore.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationManager.Filtered(context, ctx, memberClassification))
	})

	// Get paginated member classifications
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification/search",
		Method:       "GET",
		ResponseType: modelcore.MemberClassificationResponse{},
		Note:         "Returns paginated member classifications for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.modelcore.MemberClassificationCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member classifications for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationManager.Pagination(context, ctx, value))
	})

	// Create a new member classification
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification",
		Method:       "POST",
		ResponseType: modelcore.MemberClassificationResponse{},
		RequestType:  modelcore.MemberClassificationRequest{},
		Note:         "Creates a new member classification record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.MemberClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberClassification := &modelcore.MemberClassification{
			Name:           req.Name,
			Description:    req.Description,
			Icon:           req.Icon,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.MemberClassificationManager.Create(context, memberClassification); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member classification failed (/member-classification), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member classification: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member classification (/member-classification): " + memberClassification.Name,
			Module:      "MemberClassification",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationManager.ToModel(memberClassification))
	})

	// Update an existing member classification by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-classification/:member_classification_id",
		Method:       "PUT",
		ResponseType: modelcore.MemberClassificationResponse{},
		RequestType:  modelcore.MemberClassificationRequest{},
		Note:         "Updates an existing member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := handlers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), user org error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.modelcore.MemberClassificationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), validation error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberClassification, err := c.modelcore.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), not found: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member classification not found: " + err.Error()})
		}
		memberClassification.UpdatedAt = time.Now().UTC()
		memberClassification.UpdatedByID = user.UserID
		memberClassification.OrganizationID = user.OrganizationID
		memberClassification.BranchID = *user.BranchID
		memberClassification.Name = req.Name
		memberClassification.Description = req.Description
		memberClassification.Icon = req.Icon
		if err := c.modelcore.MemberClassificationManager.UpdateFields(context, memberClassification.ID, memberClassification); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member classification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member classification (/member-classification/:member_classification_id): " + memberClassification.Name,
			Module:      "MemberClassification",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.MemberClassificationManager.ToModel(memberClassification))
	})

	// Delete a member classification by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-classification/:member_classification_id",
		Method: "DELETE",
		Note:   "Deletes a member classification record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberClassificationID, err := handlers.EngineUUIDParam(ctx, "member_classification_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), invalid member_classification_id: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_classification_id: " + err.Error()})
		}
		value, err := c.modelcore.MemberClassificationManager.GetByID(context, *memberClassificationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), not found: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member classification not found: " + err.Error()})
		}
		if err := c.modelcore.MemberClassificationManager.DeleteByID(context, *memberClassificationID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member classification failed (/member-classification/:member_classification_id), db error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member classification: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member classification (/member-classification/:member_classification_id): " + value.Name,
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member classifications by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-classification/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member classification records by their IDs.",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), invalid request body.",
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), no IDs provided.",
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		var namesBuilder strings.Builder
		for _, rawID := range reqBody.IDs {
			memberClassificationID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberClassification",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}
			value, err := c.modelcore.MemberClassificationManager.GetByID(context, memberClassificationID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), not found: " + rawID,
					Module:      "MemberClassification",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member classification with ID '%s' not found: %s", rawID, err.Error())})
			}
			namesBuilder.WriteString(value.Name)
			namesBuilder.WriteString(",")
			if err := c.modelcore.MemberClassificationManager.DeleteByIDWithTx(context, tx, memberClassificationID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), db error: " + err.Error(),
					Module:      "MemberClassification",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member classification with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member classifications failed (/member-classification/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberClassification",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member classifications (/member-classification/bulk-delete): " + namesBuilder.String(),
			Module:      "MemberClassification",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
