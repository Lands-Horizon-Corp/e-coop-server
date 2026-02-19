package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func OrganizationManager(service *horizon.HorizonService) *registry.Registry[types.Organization, types.OrganizationResponse, types.OrganizationRequest] {
	return registry.GetRegistry(registry.RegistryParams[types.Organization, types.OrganizationResponse, types.OrganizationRequest]{
		Preloads: []string{"Media", "CoverMedia",
			"SubscriptionPlan", "Branches",
			"OrganizationCategories", "OrganizationMedias", "OrganizationMedias.Media",
			"OrganizationCategories.Category",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Organization) *types.OrganizationResponse {
			if data == nil {
				return nil
			}
			return &types.OrganizationResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				CreatedByID: data.CreatedByID,
				CreatedBy:   UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID: data.UpdatedByID,
				UpdatedBy:   UserManager(service).ToModel(data.UpdatedBy),

				Name:               data.Name,
				Address:            data.Address,
				Email:              data.Email,
				ContactNumber:      data.ContactNumber,
				Description:        data.Description,
				Color:              data.Color,
				Theme:              data.Theme,
				TermsAndConditions: data.TermsAndConditions,
				PrivacyPolicy:      data.PrivacyPolicy,
				CookiePolicy:       data.CookiePolicy,
				RefundPolicy:       data.RefundPolicy,
				UserAgreement:      data.UserAgreement,
				IsPrivate:          data.IsPrivate,

				InstagramLink:       data.InstagramLink,
				FacebookLink:        data.FacebookLink,
				YoutubeLink:         data.YoutubeLink,
				PersonalWebsiteLink: data.PersonalWebsiteLink,
				XLink:               data.XLink,

				MediaID:    data.MediaID,
				Media:      MediaManager(service).ToModel(data.Media),
				CoverMedia: MediaManager(service).ToModel(data.CoverMedia),

				SubscriptionPlanMaxBranches:         data.SubscriptionPlanMaxBranches,
				SubscriptionPlanMaxEmployees:        data.SubscriptionPlanMaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: data.SubscriptionPlanMaxMembersPerBranch,
				SubscriptionPlanID:                  data.SubscriptionPlanID,
				SubscriptionPlanIsYearly:            false,
				SubscriptionPlan:                    SubscriptionPlanManager(service).ToModel(data.SubscriptionPlan),
				SubscriptionStartDate:               data.SubscriptionStartDate.Format(time.RFC3339),
				SubscriptionEndDate:                 data.SubscriptionEndDate.Format(time.RFC3339),

				Branches:               BranchManager(service).ToModels(data.Branches),
				OrganizationCategories: OrganizationCategoryManager(service).ToModels(data.OrganizationCategories),
				OrganizationMedias:     OrganizationMediaManager(service).ToModels(data.OrganizationMedias),
				Footsteps:              FootstepManager(service).ToModels(data.Footsteps),
				GeneratedReports:       GeneratedReportManager(service).ToModels(data.GeneratedReports),
				InvitationCodes:        InvitationCodeManager(service).ToModels(data.InvitationCodes),
				PermissionTemplates:    PermissionTemplateManager(service).ToModels(data.PermissionTemplates),
				UserOrganizations:      UserOrganizationManager(service).ToModels(data.UserOrganizations),
			}
		},

		Created: func(data *types.Organization) registry.Topics {
			return []string{
				"organization.create",
				fmt.Sprintf("organization.create.%s", data.ID),
			}
		},
		Updated: func(data *types.Organization) registry.Topics {
			return []string{
				"organization.update",
				fmt.Sprintf("organization.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.Organization) registry.Topics {
			return []string{
				"organization.delete",
				fmt.Sprintf("organization.delete.%s", data.ID),
			}
		},
	})
}

func GetPublicOrganization(ctx context.Context, service *horizon.HorizonService) ([]*types.Organization, error) {
	filters := []query.ArrFilterSQL{
		{Field: "is_private", Op: query.ModeEqual, Value: false},
	}
	return OrganizationManager(service).ArrFind(ctx, filters, nil)
}

func GetFeaturedOrganization(ctx context.Context, service *horizon.HorizonService) ([]*types.Organization, error) {
	filters := []query.ArrFilterSQL{
		{Field: "is_private", Op: query.ModeEqual, Value: false},
	}

	organizations, err := OrganizationManager(service).ArrFind(ctx, filters, nil, "Media", "CoverMedia",
		"SubscriptionPlan", "Branches",
		"OrganizationCategories", "OrganizationMedias", "OrganizationMedias.Media",
		"OrganizationCategories.Category")
	if err != nil {
		return nil, err
	}

	var featuredOrganizations []*types.Organization
	for _, org := range organizations {
		if len(org.Branches) >= 2 {
			featuredOrganizations = append(featuredOrganizations, org)
		}
	}

	if len(featuredOrganizations) > 10 {
		featuredOrganizations = featuredOrganizations[:10]
	}

	return featuredOrganizations, nil
}

func GetOrganizationsByCategoryID(ctx context.Context, service *horizon.HorizonService, categoryID uuid.UUID) ([]*types.Organization, error) {
	filters := []query.ArrFilterSQL{
		{Field: "category_id", Op: query.ModeEqual, Value: categoryID},
	}

	orgCategories, err := OrganizationCategoryManager(service).ArrFind(ctx, filters, nil, "Organization")
	if err != nil {
		return nil, err
	}

	var organizations []*types.Organization
	seen := make(map[uuid.UUID]bool)
	for _, orgCat := range orgCategories {
		if orgCat.Organization != nil && !orgCat.Organization.IsPrivate {
			if !seen[orgCat.Organization.ID] {
				organizations = append(organizations, orgCat.Organization)
				seen[orgCat.Organization.ID] = true
			}
		}
	}

	return organizations, nil
}

func GetRecentlyAddedOrganization(ctx context.Context, service *horizon.HorizonService) ([]*types.Organization, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	filters := []query.ArrFilterSQL{
		{Field: "is_private", Op: query.ModeEqual, Value: false},
		{Field: "created_at", Op: query.ModeGTE, Value: thirtyDaysAgo},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: "DESC"},
	}

	organizations, err := OrganizationManager(service).ArrFind(ctx, filters, sorts)
	if err != nil {
		return nil, err
	}

	if len(organizations) > 15 {
		organizations = organizations[:15]
	}

	return organizations, nil
}

func GetOrganizationPerCategory(context context.Context, service *horizon.HorizonService) ([]types.OrganizationPerCategoryResponse, error) {
	categories, err := CategoryManager(service).List(context)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get categories")
	}
	organizations, err := GetPublicOrganization(context, service)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve organizations")
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
			Category:      CategoryManager(service).ToModel(category),
			Organizations: OrganizationManager(service).ToModels(orgs),
		})
	}
	return result, nil
}

func SeedOrganization(ctx context.Context, service *horizon.HorizonService, config types.OrganizationSeedConfig) error {
	orgs, err := OrganizationManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	if len(orgs) > 0 {
		return nil
	}
	hashedPassword, err := service.Security.HashPassword(config.AdminPassword)
	if err != nil {
		return eris.Wrap(err, "failed to hash password for admin user")
	}
	userMedia, err := createImageMedia(ctx, service, []string{config.AdminLogoPath}, "User Profile")
	if err != nil {
		return eris.Wrap(err, "failed to create admin user media")
	}
	owner := &types.User{
		MediaID:           &userMedia.ID,
		Email:             config.AdminEmail,
		Password:          hashedPassword,
		Birthdate:         &config.AdminBirthdate,
		Username:          config.AdminUsername,
		FullName:          config.AdminFullName,
		FirstName:         &config.AdminFirstName,
		MiddleName:        config.AdminMiddleName,
		LastName:          &config.AdminLastName,
		Suffix:            config.AdminSuffix,
		ContactNumber:     config.AdminContactNumber,
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := UserManager(service).Create(ctx, owner); err != nil {
		return eris.Wrap(err, "failed to create admin user")
	}
	subscriptions, err := SubscriptionPlanManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	currency, err := CurrencyFindByAlpha2(ctx, service, config.CurrencyAlpha2)
	if err != nil {
		return eris.Wrap(err, "failed to find currency")
	}
	sub := types.SubscriptionPlan{}
	for _, s := range subscriptions {
		if s.CurrencyID != nil && *s.CurrencyID == currency.ID {
			sub = *s
			break
		}
	}
	if len(subscriptions) == 0 {
		return eris.New("no subscription plan found")
	}
	categories, err := CategoryManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	logoMedia, err := createImageMedia(ctx, service, []string{config.OrgLogoPath}, "Organization Logo")
	if err != nil {
		return eris.Wrap(err, "failed to upload logo")
	}
	profileMedia, err := createImageMedia(ctx, service, []string{config.OrgProfilePath}, "Organization Profile")
	if err != nil {
		return eris.Wrap(err, "failed to upload profile image")
	}
	subscriptionEndDate := time.Now().AddDate(0, 0, config.SubscriptionDays)
	organization := &types.Organization{
		CreatedAt:                           time.Now().UTC(),
		CreatedByID:                         owner.ID,
		UpdatedAt:                           time.Now().UTC(),
		UpdatedByID:                         owner.ID,
		Name:                                config.OrgName,
		Address:                             config.OrgAddress,
		Email:                               config.OrgEmail,
		ContactNumber:                       config.OrgContactNumber,
		Description:                         config.OrgDescription,
		Color:                               config.OrgColor,
		TermsAndConditions:                  config.OrgTerms,
		PrivacyPolicy:                       config.OrgPrivacy,
		CookiePolicy:                        config.OrgCookie,
		RefundPolicy:                        config.OrgRefund,
		UserAgreement:                       config.OrgUserAgreement,
		IsPrivate:                           config.OrgIsPrivate,
		MediaID:                             &logoMedia.ID,
		CoverMediaID:                        &profileMedia.ID,
		SubscriptionPlanMaxBranches:         sub.MaxBranches,
		SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
		SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
		SubscriptionPlanID:                  &sub.ID,
		SubscriptionStartDate:               time.Now().UTC(),
		SubscriptionEndDate:                 subscriptionEndDate,
		InstagramLink:                       config.OrgInstagram,
		FacebookLink:                        config.OrgFacebook,
		YoutubeLink:                         config.OrgYoutube,
		PersonalWebsiteLink:                 config.OrgPersonalWebsite,
		XLink:                               config.OrgXLink,
	}
	if err := OrganizationManager(service).Create(ctx, organization); err != nil {
		return eris.Wrap(err, "failed to create organization")
	}
	for _, entry := range config.SeminarEntries {
		media, err := createImageMedia(ctx, service, []string{entry.MediaPath}, entry.Name)
		if err != nil {
			return eris.Wrapf(err, "failed to upload seminar image for %s", entry.Name)
		}
		orgMedia := &types.OrganizationMedia{
			MediaID:        media.ID,
			Name:           entry.Name,
			Description:    &entry.Description,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: organization.ID,
		}
		if err := OrganizationMediaManager(service).Create(ctx, orgMedia); err != nil {
			return eris.Wrapf(err, "failed to create organization media for seminar %s", entry.Name)
		}
	}
	for _, category := range categories {
		if err := OrganizationCategoryManager(service).Create(ctx, &types.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     &category.ID,
		}); err != nil {
			return eris.Wrap(err, "failed to link organization to category")
		}
	}
	for idx, br := range config.Branches {
		branchMedia, err := createImageMedia(ctx, service, []string{br.LogoPath}, "Branch Logo")
		if err != nil {
			return eris.Wrap(err, "failed to upload branch image")
		}
		branch := &types.Branch{
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             owner.ID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             owner.ID,
			OrganizationID:          organization.ID,
			Type:                    br.Type,
			Name:                    br.Name,
			Email:                   br.Email,
			Address:                 br.Address,
			Province:                br.Region,
			City:                    br.City,
			Region:                  br.Region,
			Barangay:                br.Barangay,
			PostalCode:              br.PostalCode,
			CurrencyID:              &currency.ID,
			ContactNumber:           &br.Contact,
			MediaID:                 &branchMedia.ID,
			Latitude:                &br.Latitude,
			Longitude:               &br.Longitude,
			TaxIdentificationNumber: &br.TaxID,
		}
		if err := BranchManager(service).Create(ctx, branch); err != nil {
			return eris.Wrapf(err, "failed to create branch %s", br.Name)
		}
		branchSetting := &types.BranchSetting{
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			BranchID:               branch.ID,
			WithdrawAllowUserInput: br.WithdrawAllowUserInput,
			WithdrawPrefix:         br.WithdrawPrefix,
			WithdrawORStart:        br.WithdrawORStart,
			WithdrawORCurrent:      br.WithdrawORCurrent,
			WithdrawOREnd:          br.WithdrawOREnd,
			WithdrawORIteration:    br.WithdrawORIteration,
			CurrencyID:             currency.ID,
		}
		if err := BranchSettingManager(service).Create(ctx, branchSetting); err != nil {
			return eris.Wrap(err, "failed to create branch settings")
		}
		developerKey, err := service.Security.GenerateUUIDv5(fmt.Sprintf("%s-%s-%s", owner.ID, organization.ID, branch.ID))
		if err != nil {
			return eris.Wrap(err, "failed to generate developer key")
		}
		ownerOrganization := &types.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            owner.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            owner.ID,
			BranchID:               &branch.ID,
			OrganizationID:         organization.ID,
			UserID:                 owner.ID,
			UserType:               types.UserOrganizationTypeOwner,
			Description:            "Founder and owner of the organization",
			ApplicationDescription: "Owner of the cooperative",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey + "-owner-horizon",
			PermissionName:         "Owner",
			PermissionDescription:  "Full administrative permissions over the cooperative",
			Permissions:            []string{"read", "write", "manage", "delete", "admin"},
			Status:                 types.UserOrganizationStatusOnline,
			LastOnlineAt:           time.Now().UTC(),
		}
		if err := UserOrganizationManager(service).Create(ctx, ownerOrganization); err != nil {
			return eris.Wrap(err, "failed to create owner association")
		}
		tx, endTx := service.Database.StartTransaction(ctx)
		if err := OrganizationSeeder(ctx, service, tx, owner.ID, organization.ID, branch.ID); err != nil {
			return endTx(err)
		}
		if err := endTx(nil); err != nil {
			return err
		}
		invitationCodes := []types.UserOrganizationType{
			types.UserOrganizationTypeMember,
			types.UserOrganizationTypeEmployee,
		}
		for _, userType := range invitationCodes {
			invitationCode := &types.InvitationCode{
				CreatedAt:      time.Now().UTC(),
				CreatedByID:    owner.ID,
				UpdatedAt:      time.Now().UTC(),
				UpdatedByID:    owner.ID,
				OrganizationID: organization.ID,
				BranchID:       branch.ID,
				UserType:       userType,
				Code:           fmt.Sprintf("%s-%s-%d", config.OrgName, userType, idx+1), // Dynamic code based on org name
				ExpirationDate: time.Now().UTC().Add(config.InvitationExpiration),
				MaxUse:         config.InvitationMaxUse,
				CurrentUse:     0,
				Description:    fmt.Sprintf("Invitation for %s of %s %s", userType, config.OrgName, br.Name),
			}
			if err := InvitationCodeManager(service).Create(ctx, invitationCode); err != nil {
				return eris.Wrap(err, "failed to create invitation code")
			}
		}
	}
	return nil
}
