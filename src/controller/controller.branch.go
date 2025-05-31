package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) BranchController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/branch",
		Method:   "GET",
		Response: "TBranch[]",
		Note:     "If there's no user organization (e.g., unauthenticated), return all branches. If a user organization exists (from JWT), filter branches by that organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil || userOrg == nil {
			branches, err := c.model.BranchManager.ListRaw(context)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			return ctx.JSON(http.StatusOK, branches)
		}
		branches, err := c.model.GetBranchesByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModels(branches))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/organization/:organization_id",
		Method:   "GET",
		Response: "TBranch[]",
		Note:     "Returns branches filtered by a specific organization ID provided in the URL path parameter.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		branches, err := c.model.GetBranchesByOrganization(context, *orgId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.BranchManager.ToModels(branches))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/user-organization/:user_organization_id",
		Method:   "POST",
		Request:  "TBranch[]",
		Response: "{branch: TBranch, user_organization: TUserOrganization}",
		Note:     "Creates a new branch under a user organization. If the user organization doesn't have a branch yet, it will be updated. Otherwise, a new user organization record is created with the new branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request payload
		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		userOrganizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID"})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
		}

		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.UserType != "owner" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User must be an owner of this organization"})
		}

		organization, err := c.model.OrganizationManager.GetByID(context, userOrganization.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Associated organization not found"})
		}

		branchCount, err := c.model.GetBranchesByOrganizationCount(context, organization.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count organization branches"})
		}

		if branchCount >= int64(organization.SubscriptionPlanMaxBranches) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Branch limit reached for current subscription plan"})
		}

		branch := &model.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: userOrganization.OrganizationID,

			MediaID:       req.MediaID,
			Type:          req.Type,
			Name:          req.Name,
			Email:         req.Email,
			Description:   req.Description,
			CountryCode:   req.CountryCode,
			ContactNumber: req.ContactNumber,
			Address:       req.Address,
			Province:      req.Province,
			City:          req.City,
			Region:        req.Region,
			Barangay:      req.Barangay,
			PostalCode:    req.PostalCode,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			IsMainBranch:  req.IsMainBranch,
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		if err := c.model.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}

		if userOrganization.BranchID == nil {
			// Update existing userOrganization
			userOrganization.BranchID = &branch.ID
			userOrganization.UpdatedAt = time.Now().UTC()
			userOrganization.UpdatedByID = user.ID

			if err := c.model.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization: " + err.Error()})
			}
		} else {
			// Create new userOrganization with new branch
			developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key"})
			}

			newUserOrg := &model.UserOrganization{
				CreatedAt:          time.Now().UTC(),
				CreatedByID:        user.ID,
				UpdatedAt:          time.Now().UTC(),
				UpdatedByID:        user.ID,
				OrganizationID:     userOrganization.OrganizationID,
				BranchID:           &branch.ID,
				UserID:             user.ID,
				UserType:           "owner",
				ApplicationStatus:  "accepted",
				DeveloperSecretKey: developerKey + uuid.NewString() + "-horizon",
				PermissionName:     "owner",
				Permissions:        []string{},
			}

			if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, newUserOrg); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new user organization: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, map[string]any{
			"branch":            c.model.BranchManager.ToModel(branch),
			"user_organization": c.model.UserOrganizationManager.ToModel(userOrganization),
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/branch/user-organization/:user_organization_id",
		Method:   "PUT",
		Request:  "TBranch[]",
		Response: "{branch: TBranch, user_organization: TUserOrganization}",
		Note:     "Updates the branch information under the specified user organization. Only allowed if the user is an 'owner' and the user organization already has an existing branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		// Validate request body
		req, err := c.model.BranchManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch data: " + err.Error()})
		}

		// Get currently authenticated user
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required: " + err.Error()})
		}

		// Parse and validate user organization ID
		userOrganizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}

		// Fetch user organization
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}

		// Ensure user is an 'owner'
		if userOrganization.UserType != "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owners can update branches"})
		}

		// Ensure user organization has an associated branch
		if userOrganization.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "This user organization does not have an associated branch"})
		}

		// Retrieve the branch
		branch, err := c.model.BranchManager.GetByID(context, *userOrganization.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Associated branch not found: " + err.Error()})
		}

		// Update branch fields
		branch.UpdatedAt = time.Now().UTC()
		branch.UpdatedByID = user.ID
		branch.OrganizationID = userOrganization.OrganizationID
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

		// Save changes to the branch
		if err := c.model.BranchManager.UpdateByID(context, userOrganization.OrganizationID, branch); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update branch: " + err.Error()})
		}

		// Respond with updated data
		return ctx.JSON(http.StatusOK, map[string]any{
			"branch":            c.model.BranchManager.ToModel(branch),
			"user_organization": c.model.UserOrganizationManager.ToModel(userOrganization),
		})
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/branch/user-organization/:user_organization_id",
		Method: "DELETE",
		Note:   "Deletes a branch and the associated user organization if the user is the owner and fewer than 3 members exist under that branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrganizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user organization ID: " + err.Error()})
		}
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
		}
		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Branch not found"})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrganizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}
		if userOrganization.UserType != "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Permission denied: Only owners can delete branches"})
		}
		count, err := c.model.CountUserOrganizationPerBranch(context, userOrganization.UserID, *userOrganization.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count user organizations: " + err.Error()})
		}
		if count > 2 {
			return ctx.JSON(http.StatusForbidden, map[string]string{
				"error": "Cannot delete branch with more than 2 members",
			})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
		}
		if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction commit failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Branch and associated user organization deleted successfully.",
		})
	})

}
