package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) OrganizationController() {
	req := c.provider.Service.Request

	// Get all public organizations
	req.RegisterRoute(handlers.Route{
		Route:        "/organization",
		Method:       "GET",
		ResponseType: model.OrganizationResponse{},
		Note:         "Returns all public organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organization, err := c.model.GetPublicOrganization(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.Filtered(context, ctx, organization))
	})

	// Get an organization by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/organization/:organization_id",
		Method:       "GET",
		ResponseType: model.OrganizationResponse{},

		Note: "Returns a specific organization by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := c.model.OrganizationManager.GetByIDRaw(context, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	// Create a new organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:        "/organization",
		Method:       "POST",
		RequestType:  model.OrganizationRequest{},
		ResponseType: model.CreateOrganizationResponse{},
		Note:         "Creates a new organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: validation error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: user error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		subscription, err := c.model.SubscriptionPlanManager.GetByID(context, *req.SubscriptionPlanID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: subscription plan error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Subscription plan not found: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		var subscriptionEndDate time.Time
		if req.SubscriptionPlanIsYearly {
			subscriptionEndDate = time.Now().UTC().AddDate(1, 0, 0)
		} else {
			subscriptionEndDate = time.Now().UTC().Add(30 * 24 * time.Hour)
		}

		organization := &model.Organization{
			CreatedAt:                           time.Now().UTC(),
			CreatedByID:                         user.ID,
			UpdatedAt:                           time.Now().UTC(),
			UpdatedByID:                         user.ID,
			Name:                                req.Name,
			Address:                             req.Address,
			Email:                               req.Email,
			ContactNumber:                       req.ContactNumber,
			Description:                         req.Description,
			Color:                               req.Color,
			TermsAndConditions:                  req.TermsAndConditions,
			PrivacyPolicy:                       req.PrivacyPolicy,
			CookiePolicy:                        req.CookiePolicy,
			RefundPolicy:                        req.RefundPolicy,
			UserAgreement:                       req.UserAgreement,
			IsPrivate:                           req.IsPrivate,
			MediaID:                             req.MediaID,
			CoverMediaID:                        req.CoverMediaID,
			SubscriptionPlanMaxBranches:         subscription.MaxBranches,
			SubscriptionPlanMaxEmployees:        subscription.MaxEmployees,
			SubscriptionPlanMaxMembersPerBranch: subscription.MaxMembersPerBranch,
			SubscriptionPlanID:                  &subscription.ID,
			SubscriptionStartDate:               time.Now().UTC(),
			SubscriptionEndDate:                 subscriptionEndDate,
		}

		if err := c.model.OrganizationManager.CreateWithTx(context, tx, organization); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization: " + err.Error()})
		}

		var longitude float64 = 0
		var latitude float64 = 0

		branch := &model.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: organization.ID,
			MediaID:        req.MediaID,
			Name:           req.Name,
			Email:          *req.Email,
			Description:    req.Description,
			CountryCode:    "",
			ContactNumber:  req.ContactNumber,
			Latitude:       &latitude,
			Longitude:      &longitude,
		}
		if err := c.model.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create branch error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}
		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: generate dev key error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + err.Error()})
		}
		userOrganization := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         organization.ID,
			UserID:                 user.ID,
			BranchID:               &branch.ID,
			UserType:               "owner",
			Description:            "",
			ApplicationDescription: "",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey + uuid.NewString() + "-horizon",
			PermissionName:         "owner",
			PermissionDescription:  "",
			Permissions:            []string{},
		}
		if err := c.model.UserOrganizationManager.CreateWithTx(context, tx, userOrganization); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create user org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user organization: " + err.Error()})
		}
		for _, category := range req.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.CreateWithTx(context, tx, &model.OrganizationCategory{
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Create organization failed: create org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization category: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.JSON(http.StatusOK, model.CreateOrganizationResponse{
			Organization:     c.model.OrganizationManager.ToModel(organization),
			UserOrganization: c.model.UserOrganizationManager.ToModel(userOrganization),
		})
	})

	// Update an organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:        "/organization/:organization_id",
		Method:       "PUT",
		RequestType:  model.OrganizationRequest{},
		ResponseType: model.OrganizationResponse{},
		Note:         "Updates an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: validation error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: user error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}

		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: not found: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		if organization.CreatedByID != user.ID {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: not authorized",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		organization.Name = req.Name
		organization.Address = req.Address
		organization.Email = req.Email
		organization.ContactNumber = req.ContactNumber
		organization.Description = req.Description
		organization.Color = req.Color
		organization.TermsAndConditions = req.TermsAndConditions
		organization.PrivacyPolicy = req.PrivacyPolicy
		organization.CookiePolicy = req.CookiePolicy
		organization.RefundPolicy = req.RefundPolicy
		organization.UserAgreement = req.UserAgreement
		organization.IsPrivate = req.IsPrivate
		organization.MediaID = req.MediaID
		organization.CoverMediaID = req.CoverMediaID
		organization.UpdatedAt = time.Now().UTC()
		organization.UpdatedByID = user.ID
		if err := c.model.OrganizationManager.UpdateFieldsWithTx(context, tx, organization.ID, organization); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: update org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization: " + err.Error()})
		}
		organizationsFromCategory, err := c.model.GetOrganizationCategoryByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: get org categories error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get organization categories: " + err.Error()})
		}

		for _, category := range organizationsFromCategory {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update organization failed: delete org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization category: " + err.Error()})
			}
		}

		for _, category := range req.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.CreateWithTx(context, tx, &model.OrganizationCategory{
				ID:             *category.ID,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update organization failed: create org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization category: " + err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModel(organization))
	})

	// Delete an organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:  "/organization/:organization_id",
		Method: "DELETE",
		Note:   "Deletes an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: not found: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		currentTime := time.Now().UTC()
		if organization.SubscriptionEndDate.After(currentTime) {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: subscription still active",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
		}
		userOrganizations, err := c.model.GetUserOrganizationByOrganization(context, organization.ID, nil)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: get user org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user organizations: " + err.Error()})
		}
		if len(userOrganizations) >= 3 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: more than 2 user orgs",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: user error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		for _, category := range organization.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization category: " + err.Error()})
			}
		}
		branches, err := c.model.GetBranchesByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: get branches error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branches for organization: " + err.Error()})
		}
		for _, branch := range branches {
			if err := c.model.OrganizationDestroyer(context, tx, user.ID, *organizationId, branch.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: destroy branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to destroy organization branch: " + err.Error()})
			}
			if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
			}
		}
		if err := c.model.OrganizationManager.DeleteByIDWithTx(context, tx, *organizationId); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: delete org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization: " + err.Error()})
		}
		for _, userOrganization := range userOrganizations {
			if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete user org error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
