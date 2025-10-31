package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberJointAccountController() {
	req := c.provider.Service.Request

	// Create a new joint account record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-joint-account/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: modelcore.MemberJointAccountResponse{},
		RequestType:  modelcore.MemberJointAccountRequest{},
		Note:         "Creates a new joint account record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.modelcore.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &modelcore.MemberJointAccount{
			MemberProfileID:    *memberProfileID,
			PictureMediaID:     req.PictureMediaID,
			SignatureMediaID:   req.SignatureMediaID,
			Description:        req.Description,
			FirstName:          req.FirstName,
			MiddleName:         req.MiddleName,
			LastName:           req.LastName,
			FullName:           req.FullName,
			Suffix:             req.Suffix,
			Birthday:           req.Birthday,
			FamilyRelationship: req.FamilyRelationship,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.UserID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		}

		if err := c.modelcore.MemberJointAccountManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create joint account record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created joint account (/member-joint-account/member-profile/:member_profile_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberJointAccountManager.ToModel(value))
	})

	// Update an existing joint account record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-joint-account/:member_joint_account_id",
		Method:       "PUT",
		ResponseType: modelcore.MemberJointAccountResponse{},
		RequestType:  modelcore.MemberJointAccountRequest{},
		Note:         "Updates an existing joint account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := handlers.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), invalid member_joint_account_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_joint_account_id: " + err.Error()})
		}
		req, err := c.modelcore.MemberJointAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), validation error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), user org error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.modelcore.MemberJointAccountManager.GetByID(context, *memberJointAccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), record not found: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Joint account record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.PictureMediaID = req.PictureMediaID
		value.SignatureMediaID = req.SignatureMediaID
		value.Description = req.Description
		value.FirstName = req.FirstName
		value.MiddleName = req.MiddleName
		value.LastName = req.LastName
		value.FullName = req.FullName
		value.Suffix = req.Suffix
		value.Birthday = req.Birthday
		value.FamilyRelationship = req.FamilyRelationship

		if err := c.modelcore.MemberJointAccountManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update joint account record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated joint account (/member-joint-account/:member_joint_account_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberJointAccountManager.ToModel(value))
	})

	// Delete a member's joint account record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-joint-account/:member_joint_account_id",
		Method: "DELETE",
		Note:   "Deletes a joint account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := handlers.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), invalid member_joint_account_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_joint_account_id: " + err.Error()})
		}
		value, err := c.modelcore.MemberJointAccountManager.GetByID(context, *memberJointAccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), record not found: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Joint account record not found: " + err.Error()})
		}
		if err := c.modelcore.MemberJointAccountManager.DeleteByID(context, *memberJointAccountID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete joint account record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted joint account (/member-joint-account/:member_joint_account_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
