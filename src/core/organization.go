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
