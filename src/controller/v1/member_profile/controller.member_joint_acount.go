package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func MemberJointAccountController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-joint-account/member-profile/:member_profile_id",
		Method:       "POST",
		ResponseType: core.MemberJointAccountResponse{},
		RequestType:  core.MemberJointAccountRequest{},
		Note:         "Creates a new joint account record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), invalid member_profile_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberJointAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &core.MemberJointAccount{
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
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
		}

		if err := core.MemberJointAccountManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create joint account failed (/member-joint-account/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create joint account record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created joint account (/member-joint-account/member-profile/:member_profile_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})

		return ctx.JSON(http.StatusOK, core.MemberJointAccountManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-joint-account/:member_joint_account_id",
		Method:       "PUT",
		ResponseType: core.MemberJointAccountResponse{},
		RequestType:  core.MemberJointAccountRequest{},
		Note:         "Updates an existing joint account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := helpers.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), invalid member_joint_account_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_joint_account_id: " + err.Error()})
		}
		req, err := core.MemberJointAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), validation error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), user org error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberJointAccountManager(service).GetByID(context, *memberJointAccountID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), record not found: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Joint account record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
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

		if err := core.MemberJointAccountManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update joint account failed (/member-joint-account/:member_joint_account_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update joint account record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated joint account (/member-joint-account/:member_joint_account_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})

		return ctx.JSON(http.StatusOK, core.MemberJointAccountManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-joint-account/:member_joint_account_id",
		Method: "DELETE",
		Note:   "Deletes a joint account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberJointAccountID, err := helpers.EngineUUIDParam(ctx, "member_joint_account_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), invalid member_joint_account_id: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_joint_account_id: " + err.Error()})
		}
		value, err := core.MemberJointAccountManager(service).GetByID(context, *memberJointAccountID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), record not found: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Joint account record not found: " + err.Error()})
		}
		if err := core.MemberJointAccountManager(service).Delete(context, *memberJointAccountID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete joint account failed (/member-joint-account/:member_joint_account_id), db error: " + err.Error(),
				Module:      "MemberJointAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete joint account record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted joint account (/member-joint-account/:member_joint_account_id): " + value.FullName,
			Module:      "MemberJointAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
