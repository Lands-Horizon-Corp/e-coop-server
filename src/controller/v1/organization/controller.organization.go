package organization

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func OrganizationController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization",
		Method:       "GET",
		ResponseType: types.OrganizationResponse{},
		Note:         "Returns all public organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organization, err := core.GetPublicOrganization(context, service)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.OrganizationManager(service).ToModels(organization))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization/:organization_id",
		Method:       "GET",
		ResponseType: types.OrganizationResponse{},

		Note: "Returns a specific organization by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := core.OrganizationManager(service).GetByIDRaw(context, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization",
		Method:       "POST",
		RequestType:  types.OrganizationRequest{},
		ResponseType: types.CreateOrganizationResponse{},
		Note:         "Creates a new organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.OrganizationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: validation error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: user error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}
		subscription, err := core.SubscriptionPlanManager(service).GetByID(context, *req.SubscriptionPlanID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: subscription plan error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Subscription plan not found: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + endTx(tx.Error).Error()})
		}
		var subscriptionEndDate time.Time
		if req.SubscriptionPlanIsYearly {
			subscriptionEndDate = time.Now().UTC().AddDate(1, 0, 0)
		} else {
			subscriptionEndDate = time.Now().UTC().Add(30 * 24 * time.Hour)
		}

		organization := &types.Organization{
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
			InstagramLink:                       req.InstagramLink,
			FacebookLink:                        req.FacebookLink,
			YoutubeLink:                         req.YoutubeLink,
			PersonalWebsiteLink:                 req.PersonalWebsiteLink,
			XLink:                               req.XLink,
			Theme:                               req.Theme,
		}

		if err := core.OrganizationManager(service).CreateWithTx(context, tx, organization); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization: " + endTx(err).Error()})
		}

		longitude := 0.0
		latitude := 0.0

		branch := &types.Branch{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.ID,
			OrganizationID: organization.ID,
			MediaID:        req.MediaID,
			Name:           req.Name,
			Email:          *req.Email,
			Description:    req.Description,
			CurrencyID:     req.CurrencyID,
			ContactNumber:  req.ContactNumber,
			Latitude:       &latitude,
			Longitude:      &longitude,
		}
		if err := core.BranchManager(service).CreateWithTx(context, tx, branch); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create branch error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + endTx(err).Error()})
		}

		branchSetting := &types.BranchSetting{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			BranchID:  branch.ID,

			CurrencyID: *req.CurrencyID,

			DefaultMemberTypeID:   nil,
			DefaultMemberGenderID: nil,
		}

		if err := core.BranchSettingManager(service).CreateWithTx(context, tx, branchSetting); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create branch settings error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + endTx(err).Error()})
		}

		developerKey, err := service.Security.GenerateUUIDv5(user.ID.String())
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: generate dev key error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate developer key: " + endTx(err).Error()})
		}
		userOrganization := &types.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         organization.ID,
			UserID:                 user.ID,
			BranchID:               &branch.ID,
			UserType:               types.UserOrganizationTypeOwner,
			Description:            "",
			ApplicationDescription: "",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey + uuid.NewString() + "-horizon",
			PermissionName:         string(types.UserOrganizationTypeOwner),
			PermissionDescription:  "",
			Permissions:            []string{},
		}
		if err := core.UserOrganizationManager(service).CreateWithTx(context, tx, userOrganization); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create user org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user organization: " + endTx(err).Error()})
		}
		for _, category := range req.OrganizationCategories {
			if err := core.OrganizationCategoryManager(service).CreateWithTx(context, tx, &types.OrganizationCategory{
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Create organization failed: create org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization category: " + endTx(err).Error()})
			}
		}

		organizationMedia := &[]types.OrganizationMedia{
			{
				Name:           "Cover Image",
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: organization.ID,
				MediaID:        *req.CoverMediaID,
			},
			{
				Name:           "Profile Image",
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: organization.ID,
				MediaID:        *req.MediaID,
			},
		}
		for _, orgMedia := range *organizationMedia {
			if err := core.OrganizationMediaManager(service).CreateWithTx(context, tx, &orgMedia); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Create organization failed: create org media error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + endTx(err).Error()})
			}
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.JSON(http.StatusOK, types.CreateOrganizationResponse{
			Organization:     core.OrganizationManager(service).ToModel(organization),
			UserOrganization: core.UserOrganizationManager(service).ToModel(userOrganization),
		})
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization/:organization_id",
		Method:       "PUT",
		RequestType:  types.OrganizationRequest{},
		ResponseType: types.OrganizationResponse{},
		Note:         "Updates an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		req, err := core.OrganizationManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: validation error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		user, err := event.CurrentUser(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: user error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get current user: " + err.Error()})
		}

		organization, err := core.OrganizationManager(service).GetByID(context, *organizationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: not found: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		if organization.CreatedByID != user.ID {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: not authorized",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + endTx(tx.Error).Error()})
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
		organization.IsPrivate = req.IsPrivate
		organization.InstagramLink = req.InstagramLink
		organization.FacebookLink = req.FacebookLink
		organization.YoutubeLink = req.YoutubeLink
		organization.PersonalWebsiteLink = req.PersonalWebsiteLink
		organization.XLink = req.XLink
		organization.Theme = req.Theme
		if err := core.OrganizationManager(service).UpdateByIDWithTx(context, tx, organization.ID, organization); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: update org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization: " + endTx(err).Error()})
		}
		organizationsFromCategory, err := core.GetOrganizationCategoryByOrganization(context, service, organization.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: get org categories error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get organization categories: " + endTx(err).Error()})
		}

		for _, category := range organizationsFromCategory {
			if err := core.OrganizationCategoryManager(service).DeleteWithTx(context, tx, category.ID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update organization failed: delete org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization category: " + endTx(err).Error()})
			}
		}

		for _, category := range req.OrganizationCategories {
			if err := core.OrganizationCategoryManager(service).CreateWithTx(context, tx, &types.OrganizationCategory{
				ID:             *category.ID,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "update-error",
					Description: "Update organization failed: create org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization category: " + endTx(err).Error()})
			}
		}

		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.JSON(http.StatusOK, core.OrganizationManager(service).ToModel(organization))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/organization/:organization_id",
		Method: "DELETE",
		Note:   "Deletes an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := helpers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := core.OrganizationManager(service).GetByID(context, *organizationID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: not found: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		currentTime := time.Now().UTC()
		if organization.SubscriptionEndDate.After(currentTime) {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: subscription still active",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
		}
		userOrganizations, err := core.GetUserOrganizationByOrganization(context, service, organization.ID, nil)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: get user org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user organizations: " + err.Error()})
		}
		if len(userOrganizations) >= 3 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: more than 2 user orgs",
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: begin tx error: " + tx.Error.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + endTx(tx.Error).Error()})
		}
		for _, category := range organization.OrganizationCategories {
			if err := core.OrganizationCategoryManager(service).DeleteWithTx(context, tx, category.ID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization category: " + endTx(err).Error()})
			}
		}
		branches, err := core.GetBranchesByOrganization(context, service, organization.ID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: get branches error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get branches for organization: " + endTx(err).Error()})
		}
		for _, branch := range branches {
			if err := core.OrganizationDestroyer(context, service, tx, *organizationID, branch.ID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: destroy branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to destroy organization branch: " + endTx(err).Error()})
			}
			if err := core.BranchManager(service).DeleteWithTx(context, tx, branch.ID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + endTx(err).Error()})
			}
		}
		if err := core.OrganizationManager(service).DeleteWithTx(context, tx, *organizationID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: delete org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization: " + endTx(err).Error()})
		}
		for _, userOrganization := range userOrganizations {
			if err := core.UserOrganizationManager(service).DeleteWithTx(context, tx, userOrganization.ID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete user org error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user organization: " + endTx(err).Error()})
			}
		}
		if err := endTx(nil); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: commit tx error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted organization: " + organization.Name,
			Module:      "Organization",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization/featured",
		Method:       "GET",
		ResponseType: types.OrganizationResponse{},
		Note:         "Returns featured organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizations, err := core.GetFeaturedOrganization(context, service)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve featured organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.OrganizationManager(service).ToModels(organizations))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization/recently",
		Method:       "GET",
		ResponseType: types.OrganizationResponse{},
		Note:         "Returns recently added organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizations, err := core.GetRecentlyAddedOrganization(context, service)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve recently added organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.OrganizationManager(service).ToModels(organizations))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization/category",
		Method:       "GET",
		ResponseType: types.OrganizationPerCategoryResponse{},
		Note:         "Returns all organization categories.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := core.CategoryManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization categories: " + err.Error()})
		}
		organizations, err := core.GetPublicOrganization(context, service)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organizations: " + err.Error()})
		}
		result := []types.OrganizationPerCategoryResponse{}
		for _, category := range categories {
			orgs := []*types.Organization{}
			for _, org := range organizations {
				hasCategory := false
				for _, orgCategory := range org.OrganizationCategories {
					if helpers.UUIDPtrEqual(orgCategory.CategoryID, &category.ID) {
						hasCategory = true
						break
					}
				}
				if hasCategory {
					orgs = append(orgs, org)
				}
			}
			result = append(result, types.OrganizationPerCategoryResponse{
				Category:      core.CategoryManager(service).ToModel(category),
				Organizations: core.OrganizationManager(service).ToModels(orgs),
			})
		}

		return ctx.JSON(http.StatusOK, result)
	})
}
