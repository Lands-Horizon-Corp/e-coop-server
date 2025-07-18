package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberTypeReferenceController() {
	req := c.provider.Service.Request

	// Get all member type references by member_type_id for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/member-type/:member_type_id/search",
		Method:   "GET",
		Response: "TMemberTypeReference[]",
		Note:     "Returns all member type references for the specified member_type_id in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := horizon.EngineUUIDParam(ctx, "member_type_id")
		if err != nil || memberTypeID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch ID is required"})
		}
		refs, err := c.model.MemberTypeReferenceManager.Find(context, &model.MemberTypeReference{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			MemberTypeID:   *memberTypeID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "MemberTypeReference not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.Pagination(context, ctx, refs))
	})

	// Get a single member type reference by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "GET",
		Response: "TMemberTypeReference",
		Note:     "Returns a specific member type reference by member_type_reference_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}
		ref, err := c.model.MemberTypeReferenceManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "MemberTypeReference not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, ref)
	})

	// Create a new member type reference
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference",
		Method:   "POST",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
		Note:     "Creates a new member type reference record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type reference failed: validation error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type reference failed: user org error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		ref := &model.MemberTypeReference{
			AccountID:                  req.AccountID,
			MemberTypeID:               req.MemberTypeID,
			MaintainingBalance:         req.MaintainingBalance,
			Description:                req.Description,
			InterestRate:               req.InterestRate,
			MinimumBalance:             req.MinimumBalance,
			Charges:                    req.Charges,
			ActiveMemberMinimumBalance: req.ActiveMemberMinimumBalance,
			ActiveMemberRatio:          req.ActiveMemberRatio,
			OtherInterestOnSavingComputationMinimumBalance: req.OtherInterestOnSavingComputationMinimumBalance,
			OtherInterestOnSavingComputationInterestRate:   req.OtherInterestOnSavingComputationInterestRate,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberTypeReferenceManager.Create(context, ref); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member type reference failed: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member type reference: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member type reference for member_type_id: " + ref.MemberTypeID.String(),
			Module:      "MemberTypeReference",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	// Update an existing member type reference by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-type-reference/:member_type_reference_id",
		Method:   "PUT",
		Request:  "TMemberTypeReference",
		Response: "TMemberTypeReference",
		Note:     "Updates an existing member type reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: invalid member_type_reference_id: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}

		req, err := c.model.MemberTypeReferenceManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: validation error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: user org error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		ref, err := c.model.MemberTypeReferenceManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: record not found: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "MemberTypeReference not found: " + err.Error()})
		}
		ref.AccountID = req.AccountID
		ref.MemberTypeID = req.MemberTypeID
		ref.MaintainingBalance = req.MaintainingBalance
		ref.Description = req.Description
		ref.InterestRate = req.InterestRate
		ref.MinimumBalance = req.MinimumBalance
		ref.Charges = req.Charges
		ref.ActiveMemberMinimumBalance = req.ActiveMemberMinimumBalance
		ref.ActiveMemberRatio = req.ActiveMemberRatio
		ref.OtherInterestOnSavingComputationMinimumBalance = req.OtherInterestOnSavingComputationMinimumBalance
		ref.OtherInterestOnSavingComputationInterestRate = req.OtherInterestOnSavingComputationInterestRate
		ref.UpdatedAt = time.Now().UTC()
		ref.UpdatedByID = user.UserID
		if err := c.model.MemberTypeReferenceManager.UpdateFields(context, ref.ID, ref); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: update error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member type reference: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member type reference for member_type_reference_id: " + ref.ID.String(),
			Module:      "MemberTypeReference",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberTypeReferenceManager.ToModel(ref))
	})

	// Delete a member type reference by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-type-reference/:member_type_reference_id",
		Method: "DELETE",
		Note:   "Deletes a member type reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := horizon.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type reference failed: invalid member_type_reference_id: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}
		if err := c.model.MemberTypeReferenceManager.DeleteByID(context, *id); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type reference failed: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member type reference: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member type reference for member_type_reference_id: " + id.String(),
			Module:      "MemberTypeReference",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member type references by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/member-type-reference/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple member type reference records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed: invalid request body.",
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed: no IDs provided.",
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed: begin tx error: " + tx.Error.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			id, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member type references failed: invalid UUID: " + rawID,
					Module:      "MemberTypeReference",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}
			ref, err := c.model.MemberTypeReferenceManager.GetByID(context, id)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member type references failed: record not found: " + rawID,
					Module:      "MemberTypeReference",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("MemberTypeReference with ID %s not found: %v", rawID, err)})
			}
			names += ref.Description + ","
			if err := c.model.MemberTypeReferenceManager.DeleteByIDWithTx(context, tx, id); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member type references failed: delete error: " + err.Error(),
					Module:      "MemberTypeReference",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member type reference with ID %s: %v", rawID, err)})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed: commit tx error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member type references: " + names,
			Module:      "MemberTypeReference",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
