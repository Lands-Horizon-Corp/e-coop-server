package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (c *Controller) userOrganinzationController() {

	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id/permission",
		Method:       "PUT",
		Note:         "Updates the permission fields of a user organization.",
		RequestType:  core.UserOrganizationPermissionPayload{},
		ResponseType: core.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		var payload core.UserOrganizationPermissionPayload
		if err := ctx.Bind(&payload); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload: " + err.Error()})
		}

		validate := validator.New()
		if err := validate.Struct(payload); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		currentUserOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}

		userOrg.PermissionName = payload.PermissionName
		userOrg.PermissionDescription = payload.PermissionDescription
		userOrg.Permissions = payload.Permissions
		userOrg.UpdatedAt = time.Now().UTC()
		userOrg.UpdatedByID = currentUserOrg.UserID

		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update permissions: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated permission for user organization " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:organization_id/seed",
		Method: "POST",
		Note:   "Seeds all branches inside an organization when first created.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: invalid organization ID: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganizations, err := c.core.GetUserOrganizationByOrganization(context, *organizationID, nil)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: get user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		if len(userOrganizations) == 0 || userOrganizations == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: user org not found",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + endTx(tx.Error).Error()})
		}
		seededAny := false
		for _, userOrganization := range userOrganizations {
			if userOrganization.UserID != user.ID {
				continue
			}
			if userOrganization.UserType != core.UserOrganizationTypeOwner {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: not owner",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners can seed the organization"})
			}
			if userOrganization.BranchID == nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: branch missing",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch is missing"})
			}
			if userOrganization.IsSeeded {
				continue
			}
			if err := c.core.OrganizationSeeder(context, tx, user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: seeder error: " + err.Error(),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to seed organization: " + endTx(err).Error()})
			}
			userOrganization.IsSeeded = true
			if err := c.core.UserOrganizationManager().UpdateByIDWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: update seed status error: " + err.Error(),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization seed status: " + endTx(err).Error()})
			}
			seededAny = true
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: commit tx error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		if seededAny {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-success",
				Description: "Seeded all branches for organization " + organizationID.String(),
				Module:      "UserOrganization",
			})
		}
		return ctx.NoContent(http.StatusOK)
	})
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/employee/search",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns paginated employee user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().NormalPagination(context, ctx, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
			UserType:       core.UserOrganizationTypeEmployee,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employee user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrganization)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/owner/search",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns paginated employee user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().NormalPagination(context, ctx, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
			UserType:       core.UserOrganizationTypeOwner,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employee user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrganization)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/member/search",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns paginated member user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().NormalPagination(context, ctx, &core.UserOrganization{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       userOrg.BranchID,
			UserType:       core.UserOrganizationTypeMember,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrganization)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/none-member-profile/search",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns paginated member user organizations without a member profile for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().RawPagination(
			context,
			ctx,
			func(db *gorm.DB) *gorm.DB {
				return db.Model(&core.UserOrganization{}).
					Where("organization_id = ?", userOrg.OrganizationID).
					Where("branch_id = ?", userOrg.BranchID).
					Where("user_type = ?", core.UserOrganizationTypeMember).
					Where(`NOT EXISTS (
				SELECT 1 FROM member_profiles mp
				WHERE mp.user_id = user_organizations.user_id
				AND mp.organization_id = user_organizations.organization_id
				AND mp.branch_id = user_organizations.branch_id
			)`)
			},
			"User", "User.Media", "Branch", "Organization",
		)

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve none member user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrganization)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/user/:user_id",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific user. Use quer	y param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := handlers.EngineUUIDParam(ctx, "user_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.core.UserManager().GetByID(context, *userID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found: " + err.Error()})
		}
		userOrganization, err := c.core.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(userOrganization))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/current",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all user organizations for the currently logged-in user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		isPending := false
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := c.core.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(userOrganization))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/join-request/paginated",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns paginated join requests for user organizations (pending applications) for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().NormalPagination(context, ctx, &core.UserOrganization{
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          userOrg.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve join requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrganization)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/join-request",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all join requests for user organizations (pending applications) for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().Find(context, &core.UserOrganization{
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          userOrg.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve join requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(userOrganization))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/organization/:organization_id",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific organization. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}

		organization, err := c.core.OrganizationManager().GetByID(context, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}

		userOrganization, err := c.core.GetUserOrganizationByOrganization(context, organization.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(userOrganization))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/branch/:branch_id",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific branch. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		branch, err := c.core.BranchManager().GetByID(context, *branchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}
		userOrganization, err := c.core.GetUserOrganizationBybranch(context, branch.OrganizationID, branch.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(userOrganization))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id/switch",
		ResponseType: core.UserOrganizationResponse{},
		Method:       "GET",
		Note:         "Switches organization and branch in Cache for the current user. No database impact.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if user.ID != userOrganization.UserID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: You do not own this user organization"})
		}
		if userOrganization.ApplicationStatus == "accepted" {
			if err := c.event.SetUserOrganization(context, ctx, userOrganization); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user organization: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrganization))
		}
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Switching forbidden - user is " + string(userOrganization.UserType)})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/unswitch",
		Method: "POST",
		Note:   "Removes organization and branch from Cache for the current user. No database impact.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		c.event.ClearCurrentToken(context, ctx)

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "User organization and branch removed from Cache (unswitch)",
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/developer-key-refresh",
		Method:       "POST",
		Note:         "Refreshes the developer key associated with the current user organization.",
		ResponseType: core.DeveloperSecretKeyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userOrg.UserID.String())
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: generate key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		userOrg.DeveloperSecretKey = developerKey + uuid.NewString() + "-horizon"
		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update developer key: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Refreshed developer key for user organization " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})
		return ctx.JSON(http.StatusOK, core.DeveloperSecretKeyResponse{
			DeveloperSecretKey: userOrg.DeveloperSecretKey,
		})
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/invitation-code/:code/join",
		Method:       "POST",
		Note:         "Joins an organization and branch using an invitation code.",
		ResponseType: core.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		invitationCode, err := c.core.VerifyInvitationCodeByCode(context, code)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: verify invitation code error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to verify invitation code: " + err.Error()})
		}
		if invitationCode == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invitation code not found",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		switch invitationCode.UserType {
		case core.UserOrganizationTypeMember:
			if !c.core.UserOrganizationMemberCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as member",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
			}
		case core.UserOrganizationTypeEmployee:
			if !c.core.UserOrganizationEmployeeCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as employee",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
			}
		default:
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: cannot join as employee (default)",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: generate developer key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &core.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              user.ID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              user.ID,
			OrganizationID:           invitationCode.OrganizationID,
			BranchID:                 &invitationCode.BranchID,
			UserID:                   user.ID,
			UserType:                 core.UserOrganizationType(invitationCode.UserType),
			Description:              invitationCode.Description,
			ApplicationDescription:   "anything",
			ApplicationStatus:        "pending",
			DeveloperSecretKey:       developerKey,
			PermissionName:           invitationCode.PermissionName,
			PermissionDescription:    invitationCode.PermissionDescription,
			Permissions:              invitationCode.Permissions,
			UserSettingDescription:   "user settings",
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + endTx(tx.Error).Error()})
		}
		if err := c.core.RedeemInvitationCode(context, tx, invitationCode.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: redeem invitation code error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to redeem invitation code: " + endTx(err).Error()})
		}
		if err := c.core.UserOrganizationManager().CreateWithTx(context, tx, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: create user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create user organization: " + endTx(err).Error()})
		}
		if err := endTx(nil); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: commit tx error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Joined organization and branch using invitation code " + code,
			Module:      "UserOrganization",
		})
		c.event.OrganizationAdminsDirectNotification(ctx, invitationCode.OrganizationID, event.NotificationEvent{
			Description: fmt.Sprintf(
				"New %s joined using invitation code: %s %s",
				string(userOrg.UserType),
				func() string {
					if user.FirstName != nil {
						return *user.FirstName
					}
					return ""
				}(),
				func() string {
					if user.LastName != nil {
						return *user.LastName
					}
					return ""
				}(),
			),
			Title:            "New Member Joined via Invitation",
			NotificationType: core.NotificationInfo,
		})

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/join",
		Method:       "POST",
		Note:         "Joins an existing organization and branch.",
		ResponseType: core.UserOrganizationResponse{},
		RequestType:  core.UserOrganizationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invalid organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invalid branch_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		req, err := c.core.UserOrganizationManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if req.UserType == core.UserOrganizationTypeMember {
			if !c.core.UserOrganizationMemberCanJoin(context, user.ID, *organizationID, *branchID) {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as member",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
			}
		}
		if req.UserType == core.UserOrganizationTypeEmployee {
			if !c.core.UserOrganizationEmployeeCanJoin(context, user.ID, *organizationID, *branchID) {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as employee",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
			}
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: generate developer key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &core.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              user.ID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              user.ID,
			OrganizationID:           *organizationID,
			BranchID:                 branchID,
			UserID:                   user.ID,
			UserType:                 core.UserOrganizationTypeMember,
			Description:              req.Description,
			ApplicationDescription:   "",
			ApplicationStatus:        "pending",
			DeveloperSecretKey:       developerKey,
			PermissionName:           string(core.UserOrganizationTypeMember),
			PermissionDescription:    "just a member",
			Permissions:              []string{},
			UserSettingDescription:   "users settings description",
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}

		if err := c.core.UserOrganizationManager().Create(context, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: create user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": "Failed to create user organization: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Joined organization and branch " + organizationID.String() + " - " + branchID.String() + " as member",
			Module:      "UserOrganization",
		})
		c.event.OrganizationAdminsDirectNotification(ctx, *organizationID, event.NotificationEvent{
			Description: fmt.Sprintf(
				"New member application received from %s %s",
				func() string {
					if user.FirstName != nil {
						return *user.FirstName
					}
					return ""
				}(),
				func() string {
					if user.LastName != nil {
						return *user.LastName
					}
					return ""
				}(),
			),
			Title:            "New Member Application",
			NotificationType: core.NotificationInfo,
		})
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/leave",
		Method: "POST",
		Note:   "Leaves the current organization and branch (must have current organization token set).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		switch userOrg.UserType {
		case core.UserOrganizationTypeOwner, core.UserOrganizationTypeEmployee:
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: forbidden for owner or employee",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Owners and employees cannot leave an organization"})
		}

		if err := c.core.UserOrganizationManager().Delete(context, userOrg.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": "Failed to leave organization: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User left organization and branch: " + userOrg.OrganizationID.String(),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/can-join-member",
		Method: "GET",
		Note:   "Checks if the user can join as a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		if !c.core.UserOrganizationMemberCanJoin(context, user.ID, *organizationID, *branchID) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/can-join-employee",
		Method: "GET",
		Note:   "Checks if the user can join as an employee.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.event.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchID, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		if !c.core.UserOrganizationEmployeeCanJoin(context, user.ID, *organizationID, *branchID) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id",
		Method:       "GET",
		Note:         "Returns a specific user organization by ID.",
		ResponseType: core.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrg, err := c.core.UserOrganizationManager().GetByIDRaw(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrg)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id/accept",
		Method: "POST",
		Note:   "Accepts an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}

		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: user organization not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != "admin" {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: not owner or admin",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners or admins can accept applications"})
		}

		if userOrg.UserID == userOrganization.UserID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: cannot accept own application",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot accept your own application"})
		}

		userOrganization.ApplicationStatus = "accepted"
		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrganization.ID, userOrganization); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to accept user organization application: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Accepted user organization application for user " + userOrganization.UserID.String(),
			Module:      "UserOrganization",
		})

		c.event.OrganizationDirectNotification(ctx, userOrganization.OrganizationID, event.NotificationEvent{
			Description:      fmt.Sprintf("Your %s application has been accepted", string(userOrganization.UserType)),
			Title:            "Application Accepted",
			NotificationType: core.NotificationSuccess,
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id/reject",
		Method: "DELETE",
		Note:   "Rejects an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}

		userOrganization, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: user organization not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != "admin" && userOrg.UserType != core.UserOrganizationTypeEmployee {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: not allowed for user type " + string(userOrg.UserType),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners, admins, or employees can reject applications"})
		}

		if userOrg.UserID == userOrganization.UserID {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: cannot reject own application",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot reject your own application"})
		}

		if err := c.core.UserOrganizationManager().Delete(context, userOrganization.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to reject user organization application: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Rejected user organization application for user " + userOrg.UserID.String(),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id",
		Method: "DELETE",
		Note:   "Deletes a user organization by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrg, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if err := c.core.UserOrganizationManager().Delete(context, userOrg.ID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/user-organization/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple user organizations by providing an array of IDs in the request body.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "UserOrganization bulk delete failed (/user-organization/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "UserOrganization bulk delete failed (/user-organization/bulk-delete) | no IDs provided",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := c.core.UserOrganizationManager().BulkDelete(context, ids); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "UserOrganization bulk delete failed (/user-organization/bulk-delete) | error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete user organizations: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted user organizations (/user-organization/bulk-delete)",
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/employee",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all employees of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		employees, err := c.core.Employees(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employees: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(employees))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/members",
		Method:       "GET",
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Returns all members of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		members, err := c.core.Members(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve members: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModels(members))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/settings/:user_organization_id",
		Method:       "PUT",
		RequestType:  core.UserOrganizationSettingsRequest{},
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Updates the user organization settings.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		var req core.UserOrganizationSettingsRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid settings payload: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		userOrg.UserType = req.UserType
		userOrg.Description = req.Description
		userOrg.ApplicationDescription = req.ApplicationDescription
		userOrg.ApplicationStatus = req.ApplicationStatus
		userOrg.UserSettingDescription = req.UserSettingDescription
		userOrg.UserSettingStartOR = req.UserSettingStartOR
		userOrg.UserSettingEndOR = req.UserSettingEndOR
		userOrg.UserSettingUsedOR = req.UserSettingUsedOR
		userOrg.UserSettingStartVoucher = req.UserSettingStartVoucher
		userOrg.UserSettingEndVoucher = req.UserSettingEndVoucher
		userOrg.UserSettingUsedVoucher = req.UserSettingUsedVoucher
		userOrg.UserSettingNumberPadding = req.UserSettingNumberPadding
		userOrg.SettingsAccountingPaymentDefaultValueID = req.SettingsAccountingPaymentDefaultValueID
		userOrg.SettingsAccountingDepositDefaultValueID = req.SettingsAccountingDepositDefaultValueID
		userOrg.SettingsAccountingWithdrawDefaultValueID = req.SettingsAccountingWithdrawDefaultValueID
		userOrg.SettingsPaymentTypeDefaultValueID = req.SettingsPaymentTypeDefaultValueID

		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization settings: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated settings for user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/settings/current",
		Method:       "PUT",
		RequestType:  core.UserOrganizationSelfSettingsRequest{},
		ResponseType: core.UserOrganizationResponse{},
		Note:         "Updates the user organization settings.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.UserOrganizationSelfSettingsRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid settings payload: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		userOrg.Description = req.Description
		userOrg.UserSettingDescription = req.UserSettingDescription
		userOrg.UserSettingStartOR = req.UserSettingStartOR
		userOrg.UserSettingEndOR = req.UserSettingEndOR
		userOrg.UserSettingUsedOR = req.UserSettingUsedOR
		userOrg.UserSettingStartVoucher = req.UserSettingStartVoucher
		userOrg.UserSettingEndVoucher = req.UserSettingEndVoucher
		userOrg.UserSettingUsedVoucher = req.UserSettingUsedVoucher
		userOrg.SettingsAllowWithdrawNegativeBalance = req.SettingsAllowWithdrawNegativeBalance
		userOrg.SettingsAllowWithdrawExactBalance = req.SettingsAllowWithdrawExactBalance
		userOrg.SettingsMaintainingBalance = req.SettingsMaintainingBalance
		userOrg.UserSettingNumberPadding = req.UserSettingNumberPadding
		userOrg.SettingsAccountingPaymentDefaultValueID = req.SettingsAccountingPaymentDefaultValueID
		userOrg.SettingsAccountingDepositDefaultValueID = req.SettingsAccountingDepositDefaultValueID
		userOrg.SettingsAccountingWithdrawDefaultValueID = req.SettingsAccountingWithdrawDefaultValueID
		userOrg.SettingsPaymentTypeDefaultValueID = req.SettingsPaymentTypeDefaultValueID
		userOrg.TimeMachineTime = req.TimeMachineTime

		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization settings: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated settings for user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/time-machine/cancel",
		Method:       "PUT",
		Note:         "Cancels time machine by setting TimeMachineTime to nil for current user organization.",
		ResponseType: core.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancel time machine failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}

		userOrg.TimeMachineTime = nil

		if err := c.core.UserOrganizationManager().UpdateByID(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cancel time machine failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to cancel time machine: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Cancelled time machine for user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.core.UserOrganizationManager().ToModel(userOrg))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/user-organization/employee",
		Method:       "POST",
		Note:         "Creates a new employee user and user organization record.",
		RequestType:  core.EmployeeCreateRequest{},
		ResponseType: core.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}

		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != "admin" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners or admins can create employees"})
		}
		var payload core.EmployeeCreateRequest
		if err := ctx.Bind(&payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload: " + err.Error()})
		}
		validate := validator.New()
		if err := validate.Struct(payload); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		hashedPwd, err := c.provider.Service.Security.HashPassword(context, payload.Password)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		now := time.Now().UTC()
		user := &core.User{
			Email:             payload.Email,
			Password:          hashedPwd,
			Username:          payload.Username,
			FirstName:         &payload.FirstName,
			MiddleName:        &payload.MiddleName,
			LastName:          &payload.LastName,
			Suffix:            &payload.Suffix,
			FullName:          payload.FullName,
			Birthdate:         payload.BirthDate,
			ContactNumber:     payload.ContactNumber,
			MediaID:           payload.MediaID,
			IsEmailVerified:   false,
			IsContactVerified: false,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := c.core.UserManager().CreateWithTx(context, tx, user); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": endTx(err).Error()})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(
			context,
			fmt.Sprintf("%s-%s-%s", user.ID, userOrg.OrganizationID, userOrg.BranchID),
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key"})
		}
		employeeOrg := &core.UserOrganization{
			CreatedAt:             now,
			CreatedByID:           userOrg.UserID,
			UpdatedAt:             now,
			UpdatedByID:           userOrg.UserID,
			BranchID:              userOrg.BranchID,
			OrganizationID:        userOrg.OrganizationID,
			UserID:                user.ID,
			UserType:              core.UserOrganizationTypeEmployee,
			ApplicationStatus:     "accepted",
			DeveloperSecretKey:    fmt.Sprintf("%s-%s-employee-horizon", developerKey, uuid.NewString()),
			Status:                core.UserOrganizationStatusOffline,
			LastOnlineAt:          now,
			PermissionName:        payload.PermissionName,
			PermissionDescription: payload.PermissionDescription,
			Permissions:           payload.Permissions,
		}
		if err := c.core.UserOrganizationManager().CreateWithTx(context, tx, employeeOrg); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": endTx(err).Error(),
			})
		}
		if err := endTx(nil); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create",
			Description: "Employee user created",
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusCreated, map[string]string{"message": "Employee user created successfully"})
	})

}
