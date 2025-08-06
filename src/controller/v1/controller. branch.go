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

// BranchController registers routes related to branch management.
func (c *Controller) BranchController() {
	req := c.provider.Service.Request

	// GET /branch: List all branches or filter by user's organization from JWT if available.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch",
		Method:       "GET",
		Note:         "Returns all branches if unauthenticated; otherwise, returns branches filtered by the user's organization from JWT.",
		ResponseType: model.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg == nil {
			branches, err := c.model.BranchManager.List(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve branches: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := c.model.GetBranchesByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.BranchManager.Filtered(context, ctx, branches))
	})

	// GET /branch/organization/:organization_id: List branches by organization ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "GET",
		Note:         "Returns all branches belonging to the specified organization.",
		ResponseType: model.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}
		branches, err := c.model.GetBranchesByOrganization(context, *orgId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not retrieve organization branches: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BranchManager.Filtered(context, ctx, branches))
	})

	// POST /branch/organization/:organization_id: Create a branch for an organization.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/organization/:organization_id",
		Method:       "POST",
		Note:         "Creates a new branch for the given organization. If the user already has a branch, a new user organization is created; otherwise, the user's current user organization is updated with the new branch.",
		Private:      true,
		RequestType:  model.BranchRequest{},
		ResponseType: model.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			// Footstep for create branch error
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to validate branch data for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		organizationId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Invalid organization ID for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization ID: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "User authentication required for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required"})
		}

		userOrganization, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:         user.ID,
			OrganizationID: *organizationId,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("User organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.UserType != "owner" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Only organization owners can create branches for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only organization owners can create branches"})
		}

		organization, err := c.model.OrganizationManager.GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Organization not found for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}

		branchCount, err := c.model.GetBranchesByOrganizationCount(context, organization.ID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed branch count for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not count organization branches: " + err.Error()})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: "Branch limit reached for POST /branch/organization/:organization_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for the current subscription plan"})
		}

		branch := &model.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: userOrganization.OrganizationID,
			MediaID:        req.MediaID,
			Type:           req.Type,
			Name:           req.Name,
			Email:          req.Email,
			Description:    req.Description,
			CountryCode:    req.CountryCode,
			ContactNumber:  req.ContactNumber,
			Address:        req.Address,
			Province:       req.Province,
			City:           req.City,
			Region:         req.Region,
			Barangay:       req.Barangay,
			PostalCode:     req.PostalCode,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			IsMainBranch:   req.IsMainBranch,
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to start DB transaction for POST /branch/organization/:organization_id: %v", tx.Error),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		if err := c.model.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to create branch for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}

		if userOrganization.BranchID == nil {
			// Assign branch to existing user organization
			userOrganization.BranchID = &branch.ID
			userOrganization.UpdatedAt = time.Now().UTC()
			userOrganization.UpdatedByID = user.ID

			if err := c.model.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to update user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + err.Error()})
			}
		} else {
			// Create new user organization for this branch
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to generate developer key for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
			}

			newUserOrg := &model.UserOrganization{
				CreatedAt:                time.Now().UTC(),
				CreatedByID:              user.ID,
				UpdatedAt:                time.Now().UTC(),
				UpdatedByID:              user.ID,
				OrganizationID:           userOrganization.OrganizationID,
				BranchID:                 &branch.ID,
				UserID:                   user.ID,
				UserType:                 "owner",
				ApplicationStatus:        "accepted",
				DeveloperSecretKey:       developerKey + uuid.NewString() + "-horizon",
				PermissionName:           "owner",
				Permissions:              []string{},
				UserSettingStartOR:       0,
				UserSettingEndOR:         5,
				UserSettingUsedOR:        0,
				UserSettingStartVoucher:  0,
				UserSettingEndVoucher:    5,
				UserSettingUsedVoucher:   0,
				UserSettingNumberPadding: 7,
			}

			if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create error",
					Description: fmt.Sprintf("Failed to create new user organization for POST /branch/organization/:organization_id: %v", err),
					Module:      "branch",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create error",
				Description: fmt.Sprintf("Failed to commit transaction for POST /branch/organization/:organization_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		// Event notification
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Create: %s", branch.Name),
			Description: fmt.Sprintf("Created a new branch: %s", branch.Name),
		})

		// Footstep for create branch success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create success",
			Description: fmt.Sprintf("Created branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModel(branch))
	})

	// PUT /branch/:branch_id: Update an existing branch (only by owner).
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/branch/:branch_id",
		Method:       "PUT",
		Note:         "Updates branch information for the specified branch. Only allowed for the owner of the branch.",
		Private:      true,
		RequestType:  model.BranchRequest{},
		ResponseType: model.BranchResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			// Footstep for update error
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to validate branch data for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "User authentication required for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required"})
		}

		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid branch id for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}

		userOrg, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:   user.ID,
			BranchID: branchId,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("User organization not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization for this branch not found: " + err.Error()})
		}
		if userOrg.UserType != "owner" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: "Only the branch owner can update branch for PUT /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can update branch information"})
		}

		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Branch not found for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		// Update branch fields
		branch.UpdatedAt = time.Now().UTC()
		branch.UpdatedByID = user.ID
		branch.MediaID = req.MediaID
		branch.Type = req.Type
		branch.Name = req.Name
		branch.Email = req.Email
		branch.Description = req.Description
		branch.CountryCode = req.CountryCode
		branch.ContactNumber = req.ContactNumber
		branch.Address = req.Address
		branch.Province = req.Province
		branch.City = req.City
		branch.Region = req.Region
		branch.Barangay = req.Barangay
		branch.PostalCode = req.PostalCode
		branch.Latitude = req.Latitude
		branch.Longitude = req.Longitude
		branch.IsMainBranch = req.IsMainBranch

		if err := c.model.BranchManager.UpdateFields(context, branch.ID, branch); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Failed to update branch for PUT /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Update: %s", branch.Name),
			Description: fmt.Sprintf("Updated branch: %s", branch.Name),
		})

		// Footstep for update branch success
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update success",
			Description: fmt.Sprintf("Updated branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})

		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModel(branch))
	})

	// DELETE /branch/:branch_id: Delete a branch (owner only, if fewer than 3 members).
	req.RegisterRoute(handlers.Route{
		Route:   "/api/v1/branch/:branch_id",
		Method:  "DELETE",
		Note:    "Deletes the specified branch if the user is the owner and there are less than 3 members in the branch.",
		Private: true,
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		branchId, err := handlers.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Invalid branch ID for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "User authentication required for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication required"})
		}
		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Branch not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found: " + err.Error()})
		}

		userOrganization, err := c.model.UserOrganizationManager.FindOne(context, &model.UserOrganization{
			UserID:         user.ID,
			BranchID:       branchId,
			OrganizationID: branch.OrganizationID,
		})
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("User organization not found for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}
		if userOrganization.UserType != "owner" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Only the branch owner can delete this branch for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only the branch owner can delete this branch"})
		}
		count, err := c.model.CountUserOrganizationPerBranch(context, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Could not check branch membership for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not check branch membership: " + err.Error()})
		}
		if count > 2 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: "Cannot delete branch with more than 2 members for DELETE /branch/:branch_id",
				Module:      "branch",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete branch with more than 2 members"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to start DB transaction for DELETE /branch/:branch_id: %v", tx.Error),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete branch for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
		}
		if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to delete user organization for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete error",
				Description: fmt.Sprintf("Failed to commit transaction for DELETE /branch/:branch_id: %v", err),
				Module:      "branch",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Notification(context, ctx, event.NotificationEvent{
			Title:       fmt.Sprintf("Delete: %s", branch.Name),
			Description: fmt.Sprintf("Deleted branch: %s", branch.Name),
		})
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete success",
			Description: fmt.Sprintf("Deleted branch: %s, ID: %s", branch.Name, branch.ID),
			Module:      "branch",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
