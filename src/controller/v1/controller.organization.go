package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) OrganizationController() {
	req := c.provider.Service.Request

	// Get all public organizations
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization",
		Method:       "GET",
		ResponseType: modelCore.OrganizationResponse{},
		Note:         "Returns all public organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organization, err := c.modelCore.GetPublicOrganization(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.OrganizationManager.Filtered(context, ctx, organization))
	})

	// Get an organization by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization/:organization_id",
		Method:       "GET",
		ResponseType: modelCore.OrganizationResponse{},

		Note: "Returns a specific organization by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := c.modelCore.OrganizationManager.GetByIDRaw(context, *organizationID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	// Create a new organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization",
		Method:       "POST",
		RequestType:  modelCore.OrganizationRequest{},
		ResponseType: modelCore.CreateOrganizationResponse{},
		Note:         "Creates a new organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.OrganizationManager.Validate(ctx)
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
		subscription, err := c.modelCore.SubscriptionPlanManager.GetByID(context, *req.SubscriptionPlanID)
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

		organization := &modelCore.Organization{
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

		if err := c.modelCore.OrganizationManager.CreateWithTx(context, tx, organization); err != nil {
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

		branch := &modelCore.Branch{
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
		if err := c.modelCore.BranchManager.CreateWithTx(context, tx, branch); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create branch error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch: " + err.Error()})
		}

		// Create default branch settings for the new branch
		branchSetting := &modelCore.BranchSetting{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			BranchID:  branch.ID,

			CurrencyID: *req.CurrencyID,

			// Withdraw Settings
			WithdrawAllowUserInput: true,
			WithdrawPrefix:         "WD",
			WithdrawORStart:        1,
			WithdrawORCurrent:      1,
			WithdrawOREnd:          999999,
			WithdrawORIteration:    1,
			WithdrawORUnique:       true,
			WithdrawUseDateOR:      false,

			// Deposit Settings
			DepositAllowUserInput: true,
			DepositPrefix:         "DP",
			DepositORStart:        1,
			DepositORCurrent:      1,
			DepositOREnd:          999999,
			DepositORIteration:    1,
			DepositORUnique:       true,
			DepositUseDateOR:      false,

			// Loan Settings
			LoanAllowUserInput: true,
			LoanPrefix:         "LN",
			LoanORStart:        1,
			LoanORCurrent:      1,
			LoanOREnd:          999999,
			LoanORIteration:    1,
			LoanORUnique:       true,
			LoanUseDateOR:      false,

			// Check Voucher Settings
			CheckVoucherAllowUserInput: true,
			CheckVoucherPrefix:         "CV",
			CheckVoucherORStart:        1,
			CheckVoucherORCurrent:      1,
			CheckVoucherOREnd:          999999,
			CheckVoucherORIteration:    1,
			CheckVoucherORUnique:       true,
			CheckVoucherUseDateOR:      false,

			// Default Member Type - can be set later
			DefaultMemberTypeID:       nil,
			LoanAppliedEqualToBalance: true,
		}

		if err := c.modelCore.BranchSettingManager.CreateWithTx(context, tx, branchSetting); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create branch settings error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create branch settings: " + err.Error()})
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
		userOrganization := &modelCore.UserOrganization{
			CreatedAt:                time.Now().UTC(),
			CreatedByID:              user.ID,
			UpdatedAt:                time.Now().UTC(),
			UpdatedByID:              user.ID,
			OrganizationID:           organization.ID,
			UserID:                   user.ID,
			BranchID:                 &branch.ID,
			UserType:                 modelCore.UserOrganizationTypeOwner,
			Description:              "",
			ApplicationDescription:   "",
			ApplicationStatus:        "accepted",
			DeveloperSecretKey:       developerKey + uuid.NewString() + "-horizon",
			PermissionName:           string(modelCore.UserOrganizationTypeOwner),
			PermissionDescription:    "",
			Permissions:              []string{},
			UserSettingStartOR:       0,
			UserSettingEndOR:         1000,
			UserSettingUsedOR:        0,
			UserSettingStartVoucher:  0,
			UserSettingEndVoucher:    0,
			UserSettingUsedVoucher:   0,
			UserSettingNumberPadding: 7,
		}
		if err := c.modelCore.UserOrganizationManager.CreateWithTx(context, tx, userOrganization); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create organization failed: create user org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user organization: " + err.Error()})
		}
		for _, category := range req.OrganizationCategories {
			if err := c.modelCore.OrganizationCategoryManager.CreateWithTx(context, tx, &modelCore.OrganizationCategory{
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

		organizationMedia := &[]modelCore.OrganizationMedia{
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
			if err := c.modelCore.OrganizationMediaManager.CreateWithTx(context, tx, &orgMedia); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Create organization failed: create org media error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization media: " + err.Error()})
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
		return ctx.JSON(http.StatusOK, modelCore.CreateOrganizationResponse{
			Organization:     c.modelCore.OrganizationManager.ToModel(organization),
			UserOrganization: c.modelCore.UserOrganizationManager.ToModel(userOrganization),
		})
	})

	// Update an organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization/:organization_id",
		Method:       "PUT",
		RequestType:  modelCore.OrganizationRequest{},
		ResponseType: modelCore.OrganizationResponse{},
		Note:         "Updates an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		req, err := c.modelCore.OrganizationManager.Validate(ctx)
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

		organization, err := c.modelCore.OrganizationManager.GetByID(context, *organizationId)
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
		organization.IsPrivate = req.IsPrivate
		if err := c.modelCore.OrganizationManager.UpdateFieldsWithTx(context, tx, organization.ID, organization); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update organization failed: update org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update organization: " + err.Error()})
		}
		organizationsFromCategory, err := c.modelCore.GetOrganizationCategoryByOrganization(context, organization.ID)
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
			if err := c.modelCore.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
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
			if err := c.modelCore.OrganizationCategoryManager.CreateWithTx(context, tx, &modelCore.OrganizationCategory{
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
		return ctx.JSON(http.StatusOK, c.modelCore.OrganizationManager.ToModel(organization))
	})

	// Delete an organization (user must be logged in)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/organization/:organization_id",
		Method: "DELETE",
		Note:   "Deletes an organization. User must be logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := handlers.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: invalid organization_id: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_id: " + err.Error()})
		}
		organization, err := c.modelCore.OrganizationManager.GetByID(context, *organizationId)
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
		userOrganizations, err := c.modelCore.GetUserOrganizationByOrganization(context, organization.ID, nil)
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
			if err := c.modelCore.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete org category error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization category: " + err.Error()})
			}
		}
		branches, err := c.modelCore.GetBranchesByOrganization(context, organization.ID)
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
			if err := c.modelCore.OrganizationDestroyer(context, tx, user.ID, *organizationId, branch.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: destroy branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to destroy organization branch: " + err.Error()})
			}
			if err := c.modelCore.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Delete organization failed: delete branch error: " + err.Error(),
					Module:      "Organization",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete branch: " + err.Error()})
			}
		}
		if err := c.modelCore.OrganizationManager.DeleteByIDWithTx(context, tx, *organizationId); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete organization failed: delete org error: " + err.Error(),
				Module:      "Organization",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organization: " + err.Error()})
		}
		for _, userOrganization := range userOrganizations {
			if err := c.modelCore.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
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

	// GET /api/v1/organization/featured
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization/featured",
		Method:       "GET",
		ResponseType: modelCore.OrganizationResponse{},
		Note:         "Returns featured organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizations, err := c.modelCore.GetFeaturedOrganization(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve featured organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.OrganizationManager.ToModels(organizations))
	})

	// GET /api/v1/organization/recently
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization/recently",
		Method:       "GET",
		ResponseType: modelCore.OrganizationResponse{},
		Note:         "Returns recently added organizations.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizations, err := c.modelCore.GetRecentlyAddedOrganization(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve recently added organizations: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.OrganizationManager.ToModels(organizations))
	})

	// GET /api/v1/organization/category
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization/category",
		Method:       "GET",
		ResponseType: modelCore.OrganizationPerCategoryResponse{},
		Note:         "Returns all organization categories.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.modelCore.CategoryManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization categories: " + err.Error()})
		}
		organizations, err := c.modelCore.GetPublicOrganization(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organizations: " + err.Error()})
		}
		result := []modelCore.OrganizationPerCategoryResponse{}
		for _, category := range categories {
			orgs := []*modelCore.Organization{}
			for _, org := range organizations {
				hasCategory := false
				for _, orgCategory := range org.OrganizationCategories {
					if handlers.UuidPtrEqual(orgCategory.CategoryID, &category.ID) {
						hasCategory = true
						break
					}
				}
				if hasCategory {
					orgs = append(orgs, org)
				}
			}
			result = append(result, modelCore.OrganizationPerCategoryResponse{
				Category:      c.modelCore.CategoryManager.ToModel(category),
				Organizations: c.modelCore.OrganizationManager.ToModels(orgs),
			})
		}

		return ctx.JSON(http.StatusOK, result)
	})
}
