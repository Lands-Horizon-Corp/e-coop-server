package controller

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) OrganizationController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/organization",
		Method:   "GET",
		Response: "TOrganization[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organization, err := c.model.GetPublicOrganization(context)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModels(organization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization/:organization_id",
		Method:   "GET",
		Response: "TCategory",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationID, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid organization ID")
		}
		organization, err := c.model.OrganizationManager.GetByIDRaw(context, *organizationID)
		if err != nil {
			return c.NotFound(ctx, "Organization")
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization",
		Method:   "POST",
		Request:  "TOrganization",
		Response: "{organization: TOrganization, user_organization: TUserOrganization}",
		Note:     "(User must be logged in) This will be use to create an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			return err
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		subscription, err := c.model.SubscriptionPlanManager.GetByID(context, *req.SubscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Subscription plan not found"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		var subscriptionEndDate time.Time
		if req.SubscriptionPlanIsYearly {
			subscriptionEndDate = time.Now().UTC().AddDate(1, 0, 0)
		} else {
			subscriptionEndDate = time.Now().UTC().Add(30 * 24 * time.Hour)
		}

		organization := &model.Organization{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.ID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.ID,
			Name:               req.Name,
			Address:            req.Address,
			Email:              req.Email,
			ContactNumber:      req.ContactNumber,
			Description:        req.Description,
			Color:              req.Color,
			TermsAndConditions: req.TermsAndConditions,
			PrivacyPolicy:      req.PrivacyPolicy,
			CookiePolicy:       req.CookiePolicy,
			RefundPolicy:       req.RefundPolicy,
			UserAgreement:      req.UserAgreement,
			IsPrivate:          req.IsPrivate,
			MediaID:            req.MediaID,
			CoverMediaID:       req.CoverMediaID,

			SubscriptionPlanMaxBranches:         subscription.MaxBranches,
			SubscriptionPlanMaxEmployees:        subscription.MaxEmployees,
			SubscriptionPlanMaxMembersPerBranch: subscription.MaxMembersPerBranch,

			SubscriptionPlanID:    &subscription.ID,
			SubscriptionStartDate: time.Now().UTC(),
			SubscriptionEndDate:   subscriptionEndDate,
		}

		if err := c.model.OrganizationManager.CreateWithTx(context, tx, organization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		developerKey, err := c.provider.Service.Security.GenerateUUIDv5(context, user.ID.String())
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "something wrong generting developer key"})
		}
		userOrganization := &model.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            user.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            user.ID,
			OrganizationID:         organization.ID,
			BranchID:               nil,
			UserID:                 user.ID,
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, category := range req.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.CreateWithTx(context, tx, &model.OrganizationCategory{
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				OrganizationID: &organization.ID,
				CategoryID:     &category.CategoryID,
			}); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, organization)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization/:organization_id",
		Method:   "PUT",
		Request:  "TOrganization",
		Response: "{organization: TOrganization, user_organization: TUserOrganization}",
		Note:     "(User must be logged in) This will be use to update an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		req, err := c.model.OrganizationManager.Validate(ctx)
		if err != nil {
			return err
		}

		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}

		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}
		if organization.CreatedByID != user.ID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this organization"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
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
		if err := c.model.OrganizationManager.UpdateByIDWithTx(context, tx, organization.ID, organization); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		organizationsFromCategory, err := c.model.GetOrganizationCategoryByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		for _, category := range organizationsFromCategory {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModel(organization))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/organization/:organization_id",
		Method: "DELETE",
		Note:   "(User must be logged in) This will be use to DELETE an organization",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		if err != nil {
			return err
		}
		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Organization not found"})
		}
		currentTime := time.Now().UTC()
		if organization.SubscriptionEndDate.After(currentTime) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Subscription plan is still active"})
		}
		userOrganizations, err := c.model.GetUserOrganizationByOrganization(context, organization.ID, nil)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if len(userOrganizations) >= 3 {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot delete organization with more than 2 user organizations"})
		}
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, category := range organization.OrganizationCategories {
			if err := c.model.OrganizationCategoryManager.DeleteByIDWithTx(context, tx, category.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		branches, err := c.model.GetBranchesByOrganization(context, organization.ID)
		if err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, branch := range branches {
			if err := c.model.OrganizationDestroyer(context, tx, user.ID, *organizationId, branch.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			if err := c.model.BranchManager.DeleteByIDWithTx(context, tx, branch.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := c.model.OrganizationManager.DeleteByIDWithTx(context, tx, *organizationId); err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, userOrganization := range userOrganizations {

			if err := c.model.UserOrganizationManager.DeleteByIDWithTx(context, tx, userOrganization.ID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
