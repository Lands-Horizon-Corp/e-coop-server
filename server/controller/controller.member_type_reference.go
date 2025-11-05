package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberTypeReferenceController() {
	req := c.provider.Service.Request

	// Get all member type references by member_type_id for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-reference/member-type/:member_type_id/search",
		Method:       "GET",
		ResponseType: core.MemberTypeReferenceResponse{},
		Note:         "Returns all member type references for the specified member_type_id in the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberTypeID, err := handlers.EngineUUIDParam(ctx, "member_type_id")
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
		refs, err := c.core.MemberTypeReferenceManager.PaginationWithFields(context, ctx, &core.MemberTypeReference{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
			MemberTypeID:   *memberTypeID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "MemberTypeReference not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, refs)
	})

	// Get a single member type reference by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-reference/:member_type_reference_id",
		Method:       "GET",
		ResponseType: core.MemberTypeReferenceResponse{},
		Note:         "Returns a specific member type reference by member_type_reference_id.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}
		ref, err := c.core.MemberTypeReferenceManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "MemberTypeReference not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, ref)
	})

	// Create a new member type reference
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-reference",
		Method:       "POST",
		ResponseType: core.MemberTypeReferenceResponse{},
		RequestType:  core.MemberTypeReferenceRequest{},
		Note:         "Creates a new member type reference record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberTypeReferenceManager.Validate(ctx)
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

		ref := &core.MemberTypeReference{
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

		if err := c.core.MemberTypeReferenceManager.Create(context, ref); err != nil {
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

		return ctx.JSON(http.StatusOK, c.core.MemberTypeReferenceManager.ToModel(ref))
	})

	// Update an existing member type reference by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-type-reference/:member_type_reference_id",
		Method:       "PUT",
		ResponseType: core.MemberTypeReferenceResponse{},
		RequestType:  core.MemberTypeReferenceRequest{},
		Note:         "Updates an existing member type reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member type reference failed: invalid member_type_reference_id: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}

		req, err := c.core.MemberTypeReferenceManager.Validate(ctx)
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

		ref, err := c.core.MemberTypeReferenceManager.GetByID(context, *id)
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
		if err := c.core.MemberTypeReferenceManager.UpdateByID(context, ref.ID, ref); err != nil {
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
		return ctx.JSON(http.StatusOK, c.core.MemberTypeReferenceManager.ToModel(ref))
	})

	// Delete a member type reference by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-type-reference/:member_type_reference_id",
		Method: "DELETE",
		Note:   "Deletes a member type reference by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "member_type_reference_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member type reference failed: invalid member_type_reference_id: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_type_reference_id: " + err.Error()})
		}
		if err := c.core.MemberTypeReferenceManager.Delete(context, *id); err != nil {
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
	// Simplified bulk-delete handler for member type references (mirrors feedback/holiday pattern)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-type-reference/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member type reference records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed (/member-type-reference/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed (/member-type-reference/bulk-delete) | no IDs provided",
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.MemberTypeReferenceManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member type references failed (/member-type-reference/bulk-delete) | error: " + err.Error(),
				Module:      "MemberTypeReference",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member type references: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member type references (/member-type-reference/bulk-delete)",
			Module:      "MemberTypeReference",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
