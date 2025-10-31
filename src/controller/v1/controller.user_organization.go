package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) UserOrganinzationController() {

	req := c.provider.Service.Request

	// Update the permission fields of a user organization
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id/permission",
		Method:       "PUT",
		Note:         "Updates the permission fields of a user organization.",
		RequestType:  modelCore.UserOrganizationPermissionPayload{},
		ResponseType: modelCore.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		var payload modelCore.UserOrganizationPermissionPayload
		if err := ctx.Bind(&payload); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload: " + err.Error()})
		}

		validate := validator.New()
		if err := validate.Struct(payload); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		currentUserOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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

		if err := c.modelCore.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update permission failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update permissions: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated permission for user organization " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrg))
	})

	// Seed all branches inside an organization when first created
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:organization_id/seed",
		Method: "POST",
		Note:   "Seeds all branches inside an organization when first created.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: invalid organization ID: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganizations, err := c.modelCore.GetUserOrganizationByOrganization(context, *orgId, nil)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: get user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		if len(userOrganizations) == 0 || userOrganizations == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: user org not found",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		seededAny := false
		for _, userOrganization := range userOrganizations {
			if userOrganization.UserID != user.ID {
				continue
			}
			if userOrganization.UserType != modelCore.UserOrganizationTypeOwner {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: not owner",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners can seed the organization"})
			}
			if userOrganization.BranchID == nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: branch missing",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Branch is missing"})
			}
			if userOrganization.IsSeeded {
				continue
			}
			if err := c.modelCore.OrganizationSeeder(context, tx, user.ID, userOrganization.OrganizationID, *userOrganization.BranchID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: seeder error: " + err.Error(),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to seed organization: " + err.Error()})
			}
			userOrganization.IsSeeded = true
			if err := c.modelCore.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Seed organization failed: update seed status error: " + err.Error(),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization seed status: " + err.Error()})
			}
			seededAny = true
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Seed organization failed: commit tx error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		if seededAny {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-success",
				Description: "Seeded all branches for organization " + orgId.String(),
				Module:      "UserOrganization",
			})
		}
		return ctx.NoContent(http.StatusOK)
	})
	// Get paginated user organizations for employees on current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/employee/search",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns paginated employee user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       modelCore.UserOrganizationTypeEmployee,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employee user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/owner/search",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns paginated employee user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       modelCore.UserOrganizationTypeOwner,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employee user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	// Get paginated user organizations for members on current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/member/search",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns paginated member user organizations for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       modelCore.UserOrganizationTypeMember,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	// Get paginated user organizations for members without profiles on current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/none-member-profile/search",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns paginated member user organizations without a member profile for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID: user.OrganizationID,
			BranchID:       user.BranchID,
			UserType:       modelCore.UserOrganizationTypeMember,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve member user organizations: " + err.Error()})
		}
		filteredUserOrganizations := []*modelCore.UserOrganization{}
		for _, uo := range userOrganization {
			if uo.BranchID == nil {
				continue
			}
			userProfile, _ := c.modelCore.MemberProfileFindUserByID(context, uo.UserID, uo.OrganizationID, *uo.BranchID)
			if userProfile == nil {
				filteredUserOrganizations = append(filteredUserOrganizations, uo)
			}
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Pagination(context, ctx, filteredUserOrganizations))
	})

	// Retrieve all user organizations for a user (optionally including pending)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/user/:user_id",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific user. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userId, err := handlers.EngineUUIDParam(ctx, "user_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.modelCore.UserManager.GetByID(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found: " + err.Error()})
		}
		userOrganization, err := c.modelCore.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganization))
	})

	// Retrieve all user organizations for the logged-in user (not pending)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/current",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all user organizations for the currently logged-in user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		isPending := false
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := c.modelCore.GetUserOrganizationByUser(context, user.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganization))
	})

	// Get paginated join requests for user organizations in the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/join-request/paginated",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns paginated join requests for user organizations (pending applications) for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID:    user.OrganizationID,
			BranchID:          user.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve join requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Pagination(context, ctx, userOrganization))
	})

	// Get all join requests for user organizations in the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/join-request",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all join requests for user organizations (pending applications) for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.Find(context, &modelCore.UserOrganization{
			OrganizationID:    user.OrganizationID,
			BranchID:          user.BranchID,
			ApplicationStatus: "pending",
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve join requests: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganization))
	})

	// Retrieve all user organizations for a specific organization (optionally including pending)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/organization/:organization_id",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific organization. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}

		organization, err := c.modelCore.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}

		userOrganization, err := c.modelCore.GetUserOrganizationByOrganization(context, organization.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganization))
	})

	// Retrieve all user organizations for a specific branch (optionally including pending)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/branch/:branch_id",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all user organizations for a specific branch. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		branch, err := c.modelCore.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}
		userOrganization, err := c.modelCore.GetUserOrganizationByBranch(context, branch.OrganizationID, branch.ID, &isPending)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve user organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, userOrganization))
	})

	// Switch organization and branch stored in JWT (no database impact)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id/switch",
		ResponseType: modelCore.UserOrganizationResponse{},
		Method:       "GET",
		Note:         "Switches organization and branch in JWT for the current user. No database impact.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		userOrganization, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if user.ID != userOrganization.UserID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: You do not own this user organization"})
		}
		if userOrganization.ApplicationStatus == "accepted" {
			if err := c.userOrganizationToken.SetUserOrganization(context, ctx, userOrganization); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set user organization: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrganization))
		}
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Switching forbidden - user is " + string(userOrganization.UserType)})
	})

	// Remove organization and branch from JWT (no database impact)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/unswitch",
		Method: "POST",
		Note:   "Removes organization and branch from JWT for the current user. No database impact.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Remove the token
		c.userOrganizationToken.ClearCurrentToken(context, ctx)

		// Log the footstep event
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "User organization and branch removed from JWT (unswitch)",
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Refresh developer key associated with the user organization
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/developer-key-refresh",
		Method:       "POST",
		Note:         "Refreshes the developer key associated with the current user organization.",
		ResponseType: modelCore.DeveloperSecretKeyResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, userOrg.UserID.String())
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: generate key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		userOrg.DeveloperSecretKey = developerKey + uuid.NewString() + "-horizon"
		if err := c.modelCore.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Refresh developer key failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update developer key: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Refreshed developer key for user organization " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})
		return ctx.JSON(http.StatusOK, modelCore.DeveloperSecretKeyResponse{
			DeveloperSecretKey: userOrg.DeveloperSecretKey,
		})
	})

	// Join organization and branch using an invitation code
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/invitation-code/:code/join",
		Method:       "POST",
		Note:         "Joins an organization and branch using an invitation code.",
		ResponseType: modelCore.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		invitationCode, err := c.modelCore.VerifyInvitationCodeByCode(context, code)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: verify invitation code error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to verify invitation code: " + err.Error()})
		}
		if invitationCode == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invitation code not found",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		switch invitationCode.UserType {
		case modelCore.UserOrganizationTypeMember:
			if !c.modelCore.UserOrganizationMemberCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as member",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
			}
		case modelCore.UserOrganizationTypeEmployee:
			if !c.modelCore.UserOrganizationEmployeeCanJoin(context, user.ID, invitationCode.OrganizationID, invitationCode.BranchID) {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as employee",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
			}
		default:
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: cannot join as employee (default)",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: generate developer key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &modelCore.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              user.ID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              user.ID,
			OrganizationID:           invitationCode.OrganizationID,
			BranchID:                 &invitationCode.BranchID,
			UserID:                   user.ID,
			UserType:                 modelCore.UserOrganizationType(invitationCode.UserType),
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
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		if err := c.modelCore.RedeemInvitationCode(context, tx, invitationCode.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: redeem invitation code error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to redeem invitation code: " + err.Error()})
		}
		if err := c.modelCore.UserOrganizationManager.CreateWithTx(context, tx, userOrg); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: create user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to create user organization: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: commit tx error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Joined organization and branch using invitation code " + code,
			Module:      "UserOrganization",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrg))
	})

	// Join an organization and branch that is already created
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/join",
		Method:       "POST",
		Note:         "Joins an existing organization and branch.",
		ResponseType: modelCore.UserOrganizationResponse{},
		RequestType:  modelCore.UserOrganizationRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invalid organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: invalid branch_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		req, err := c.modelCore.UserOrganizationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		if req.UserType == modelCore.UserOrganizationTypeMember {
			if !c.modelCore.UserOrganizationMemberCanJoin(context, user.ID, *orgId, *branchId) {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as member",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
			}
		}
		if req.UserType == modelCore.UserOrganizationTypeEmployee {
			if !c.modelCore.UserOrganizationEmployeeCanJoin(context, user.ID, *orgId, *branchId) {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Join organization failed: cannot join as employee",
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
			}
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: generate developer key error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		developerKey = developerKey + uuid.NewString() + "-horizon"
		userOrg := &modelCore.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              user.ID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              user.ID,
			OrganizationID:           *orgId,
			BranchID:                 branchId,
			UserID:                   user.ID,
			UserType:                 modelCore.UserOrganizationTypeMember,
			Description:              req.Description,
			ApplicationDescription:   "",
			ApplicationStatus:        "pending",
			DeveloperSecretKey:       developerKey,
			PermissionName:           string(modelCore.UserOrganizationTypeMember),
			PermissionDescription:    "just a member",
			Permissions:              []string{},
			UserSettingDescription:   "",
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}

		if err := c.modelCore.UserOrganizationManager.Create(context, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Join organization failed: create user org error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": "Failed to create user organization: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Joined organization and branch " + orgId.String() + " - " + branchId.String() + " as member",
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrg))
	})

	// Leave a specific organization and branch (must have current organization)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/leave",
		Method: "POST",
		Note:   "Leaves the current organization and branch (must have current organization token set).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		switch userOrg.UserType {
		case modelCore.UserOrganizationTypeOwner, modelCore.UserOrganizationTypeEmployee:
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: forbidden for owner or employee",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Owners and employees cannot leave an organization"})
		}

		if err := c.modelCore.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Leave organization failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotAcceptable, map[string]string{"error": "Failed to leave organization: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "User left organization and branch: " + userOrg.OrganizationID.String(),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Check if the user can join as a member
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/can-join-member",
		Method: "GET",
		Note:   "Checks if the user can join as a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		if !c.modelCore.UserOrganizationMemberCanJoin(context, user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as member"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	// Check if the user can join as an employee
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/organization/:organization_id/branch/:branch_id/can-join-employee",
		Method: "GET",
		Note:   "Checks if the user can join as an employee.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch_id: " + err.Error()})
		}
		if !c.modelCore.UserOrganizationEmployeeCanJoin(context, user.ID, *orgId, *branchId) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot join as employee"})
		}
		return ctx.NoContent(http.StatusOK)
	})

	// Retrieve a specific user organization by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/:user_organization_id",
		Method:       "GET",
		Note:         "Returns a specific user organization by ID.",
		ResponseType: modelCore.UserOrganizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrg, err := c.modelCore.UserOrganizationManager.GetByIDRaw(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, userOrg)
	})

	// Accept an employee or member application by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id/accept",
		Method: "POST",
		Note:   "Accepts an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}

		userOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: user organization not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		if user.UserType != modelCore.UserOrganizationTypeOwner && user.UserType != "admin" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: not owner or admin",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners or admins can accept applications"})
		}

		if user.UserID == userOrg.UserID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: cannot accept own application",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot accept your own application"})
		}

		userOrg.ApplicationStatus = "accepted"
		if err := c.modelCore.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "approve-error",
				Description: "Accept application failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to accept user organization application: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "approve-success",
			Description: "Accepted user organization application for user " + userOrg.UserID.String(),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Reject an employee or member application by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id/reject",
		Method: "DELETE",
		Note:   "Rejects an employee or member application by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: unauthorized: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: " + err.Error()})
		}

		userOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: user organization not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		if user.UserType != modelCore.UserOrganizationTypeOwner && user.UserType != "admin" && user.UserType != modelCore.UserOrganizationTypeEmployee {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: not allowed for user type " + string(user.UserType),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners, admins, or employees can reject applications"})
		}

		if user.UserID == userOrg.UserID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: cannot reject own application",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You cannot reject your own application"})
		}

		if err := c.modelCore.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Reject application failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to reject user organization application: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Rejected user organization application for user " + userOrg.UserID.String(),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Delete a user organization by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/user-organization/:user_organization_id",
		Method: "DELETE",
		Note:   "Deletes a user organization by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: invalid user_organization_id: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}
		userOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: not found: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if err := c.modelCore.UserOrganizationManager.DeleteByID(context, userOrg.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete user organization failed: delete error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete user organizations by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/user-organization/bulk-delete",
		Method:      "DELETE",
		RequestType: modelCore.IDSRequest{},
		Note:        "Deletes multiple user organizations by providing an array of IDs in the request body.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelCore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bulk delete failed: invalid request body: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bulk delete failed: no IDs provided",
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion"})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bulk delete failed: begin tx error: " + tx.Error.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			userOrgId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: fmt.Sprintf("Bulk delete failed: invalid UUID: %s - %v", rawID, err),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			if _, err := c.modelCore.UserOrganizationManager.GetByID(context, userOrgId); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: fmt.Sprintf("Bulk delete failed: user org with ID %s not found: %v", rawID, err),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("User organization with ID %s not found: %v", rawID, err)})
			}

			if err := c.modelCore.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrgId); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: fmt.Sprintf("Bulk delete failed: delete error for ID %s: %v", rawID, err),
					Module:      "UserOrganization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete user organization with ID %s: %v", rawID, err)})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Bulk delete failed: commit tx error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: fmt.Sprintf("Bulk deleted user organizations: %v", reqBody.IDs),
			Module:      "UserOrganization",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Retrieve all employees of the current user's organization
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/employee",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all employees of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		employees, err := c.modelCore.Employees(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve employees: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, employees))
	})

	// Retrieve all members of the current user's organization
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/members",
		Method:       "GET",
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Returns all members of the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		members, err := c.modelCore.Members(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve members: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.Filtered(context, ctx, members))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/settings/:user_organization_id",
		Method:       "PUT",
		RequestType:  modelCore.UserOrganizationSettingsRequest{},
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Updates the user organization settings.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id: " + err.Error()})
		}

		var req modelCore.UserOrganizationSettingsRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid settings payload: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		// Update fields
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

		if err := c.modelCore.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization settings: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated settings for user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrg))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/user-organization/settings/current",
		Method:       "PUT",
		RequestType:  modelCore.UserOrganizationSelfSettingsRequest{},
		ResponseType: modelCore.UserOrganizationResponse{},
		Note:         "Updates the user organization settings.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req modelCore.UserOrganizationSelfSettingsRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: invalid payload: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid settings payload: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: validation error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
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

		if err := c.modelCore.UserOrganizationManager.UpdateFields(context, userOrg.ID, userOrg); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update settings failed: update error: " + err.Error(),
				Module:      "UserOrganization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization settings: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated settings for user organization: " + userOrg.ID.String(),
			Module:      "UserOrganization",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.UserOrganizationManager.ToModel(userOrg))
	})
}
