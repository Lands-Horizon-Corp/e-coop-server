package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Organization struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		Name               string  `gorm:"type:varchar(255);not null" json:"name"`
		Address            *string `gorm:"type:varchar(500)" json:"address,omitempty"`
		Email              *string `gorm:"type:varchar(255)" json:"email,omitempty"`
		ContactNumber      *string `gorm:"type:varchar(20)" json:"contact_number,omitempty"`
		Description        *string `gorm:"type:text" json:"description,omitempty"`
		Color              *string `gorm:"type:varchar(50)" json:"color,omitempty"`
		TermsAndConditions *string `gorm:"type:text" json:"terms_and_conditions,omitempty"`
		PrivacyPolicy      *string `gorm:"type:text" json:"privacy_policy,omitempty"`
		CookiePolicy       *string `gorm:"type:text" json:"cookie_policy,omitempty"`
		RefundPolicy       *string `gorm:"type:text" json:"refund_policy,omitempty"`
		UserAgreement      *string `gorm:"type:text" json:"user_agreement,omitempty"`
		IsPrivate          bool    `gorm:"default:false" json:"is_private"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id,omitempty"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		CoverMediaID *uuid.UUID `gorm:"type:uuid" json:"cover_media_id,omitempty"`
		CoverMedia   *Media     `gorm:"foreignKey:CoverMediaID;constraint:OnDelete:SET NULL;" json:"cover_media,omitempty"`

		SubscriptionPlanMaxBranches         int               `gorm:"not null"`
		SubscriptionPlanMaxEmployees        int               `gorm:"not null"`
		SubscriptionPlanMaxMembersPerBranch int               `gorm:"not null"`
		SubscriptionPlanID                  *uuid.UUID        `gorm:"type:uuid" json:"subscription_plan_id,omitempty"`
		SubscriptionPlan                    *SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;constraint:OnDelete:SET NULL;" json:"subscription_plan,omitempty"`
		SubscriptionStartDate               time.Time         `json:"subscription_start_date"`
		SubscriptionEndDate                 time.Time         `json:"subscription_end_date"`

		Branches               []*Branch                 `gorm:"foreignKey:OrganizationID" json:"branches,omitempty"`                 // branches
		OrganizationCategories []*OrganizationCategory   `gorm:"foreignKey:OrganizationID" json:"organization_categories,omitempty"`  // organization categories
		OrganizationMedias     []*OrganizationMedia      `gorm:"foreignKey:OrganizationID" json:"organization_medias,omitempty"`      // organization media
		Footsteps              []*Footstep               `gorm:"foreignKey:OrganizationID" json:"footsteps,omitempty"`                // footstep
		GeneratedReports       []*GeneratedReport        `gorm:"foreignKey:OrganizationID" json:"generated_reports,omitempty"`        // generated report
		InvitationCodes        []*InvitationCode         `gorm:"foreignKey:OrganizationID" json:"invitation_codes,omitempty"`         // invitation code
		OrganizationDailyUsage []*OrganizationDailyUsage `gorm:"foreignKey:OrganizationID" json:"organization_daily_usage,omitempty"` // organization daily usage
		PermissionTemplates    []*PermissionTemplate     `gorm:"foreignKey:OrganizationID" json:"permission_templates,omitempty"`     // permission template
		UserOrganizations      []*UserOrganization       `gorm:"foreignKey:OrganizationID" json:"user_organizations,omitempty"`       // user organization
	}

	OrganizationResponse struct {
		ID          uuid.UUID     `json:"id"`
		CreatedAt   string        `json:"created_at"`
		CreatedByID uuid.UUID     `json:"created_by_id"`
		CreatedBy   *UserResponse `json:"created_by,omitempty"`
		UpdatedAt   string        `json:"updated_at"`
		UpdatedByID uuid.UUID     `json:"updated_by_id"`
		UpdatedBy   *UserResponse `json:"updated_by,omitempty"`

		Name               string  `json:"name"`
		Address            *string `json:"address,omitempty"`
		Email              *string `json:"email,omitempty"`
		ContactNumber      *string `json:"contact_number,omitempty"`
		Description        *string `json:"description,omitempty"`
		Color              *string `json:"color,omitempty"`
		TermsAndConditions *string `json:"terms_and_conditions,omitempty"`
		PrivacyPolicy      *string `json:"privacy_policy,omitempty"`
		CookiePolicy       *string `json:"cookie_policy,omitempty"`
		RefundPolicy       *string `json:"refund_policy,omitempty"`
		UserAgreement      *string `json:"user_agreement,omitempty"`
		IsPrivate          bool    `json:"is_private,omitempty"`

		MediaID      *uuid.UUID     `json:"media_id,omitempty"`
		Media        *MediaResponse `json:"media,omitempty"`
		CoverMediaID *uuid.UUID     `json:"cover_media_id,omitempty"`
		CoverMedia   *MediaResponse `json:"cover_media,omitempty"`

		SubscriptionPlanMaxBranches         int                       `json:"subscription_plan_max_branches"`
		SubscriptionPlanMaxEmployees        int                       `json:"subscription_plan_max_employees"`
		SubscriptionPlanMaxMembersPerBranch int                       `json:"subscription_plan_max_member_per_branch"`
		SubscriptionPlanID                  *uuid.UUID                `json:"subscription_plan_id,omitempty"`
		SubscriptionPlan                    *SubscriptionPlanResponse `json:"subscription_plan,omitempty"`
		SubscriptionStartDate               string                    `json:"subscription_start_date"`
		SubscriptionEndDate                 string                    `json:"subscription_end_date"`
		SubscriptionPlanIsYearly            bool                      `json:"subscription_plan_is_yearly,omitempty"`

		Branches               []*BranchResponse                 `json:"branches,omitempty"`
		OrganizationCategories []*OrganizationCategoryResponse   `json:"organization_categories,omitempty"`
		OrganizationMedias     []*OrganizationMediaResponse      `json:"organization_medias,omitempty"`
		Footsteps              []*FootstepResponse               `json:"footsteps,omitempty"`
		GeneratedReports       []*GeneratedReportResponse        `json:"generated_reports,omitempty"`
		InvitationCodes        []*InvitationCodeResponse         `json:"invitation_codes,omitempty"`
		OrganizationDailyUsage []*OrganizationDailyUsageResponse `json:"organization_daily_usage,omitempty"`
		PermissionTemplates    []*PermissionTemplateResponse     `json:"permission_templates,omitempty"`
		UserOrganizations      []*UserOrganizationResponse       `json:"user_organizations,omitempty"`
	}

	OrganizationRequest struct {
		ID *string `json:"id,omitempty"`

		Name               string  `json:"name" validate:"required,min=1,max=255"`
		Address            *string `json:"address,omitempty" validate:"omitempty,max=500"`
		Email              *string `json:"email,omitempty" validate:"omitempty,email"`
		ContactNumber      *string `json:"contact_number,omitempty" validate:"omitempty,max=20"`
		Description        *string `json:"description,omitempty"`
		Color              *string `json:"color,omitempty"`
		TermsAndConditions *string `json:"terms_and_conditions,omitempty"`
		PrivacyPolicy      *string `json:"privacy_policy,omitempty"`
		CookiePolicy       *string `json:"cookie_policy,omitempty"`
		RefundPolicy       *string `json:"refund_policy,omitempty"`
		UserAgreement      *string `json:"user_agreement,omitempty"`
		IsPrivate          bool    `json:"is_private,omitempty"`

		MediaID      *uuid.UUID `json:"media_id,omitempty"`
		CoverMediaID *uuid.UUID `json:"cover_media_id,omitempty"`

		SubscriptionPlanID       *uuid.UUID `json:"subscription_plan_id,omitempty"`
		SubscriptionPlanIsYearly bool       `json:"subscription_plan_is_yearly,omitempty"`

		OrganizationCategories []*OrganizationCategoryRequest `json:"organization_categories,omitempty"`

		CurrencyID *uuid.UUID `json:"currency_id,omitempty"`
	}

	OrganizationSubscriptionRequest struct {
		OrganizationID           uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		SubscriptionPlanID       uuid.UUID `json:"subscription_plan_id" validate:"required,uuid4"`
		SubscriptionPlanIsYearly *bool     `json:"subscription_plan_is_yearly,omitempty"`
	}

	CreateOrganizationResponse struct {
		Organization     *OrganizationResponse     `json:"organization"`
		UserOrganization *UserOrganizationResponse `json:"user_organization"`
	}

	OrganizationPerCategoryResponse struct {
		Category      *CategoryResponse       `json:"category"`
		Organizations []*OrganizationResponse `json:"organizations"`
	}
)

func (m *ModelCore) Organization() {
	m.Migration = append(m.Migration, &Organization{})
	m.OrganizationManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Organization, OrganizationResponse, OrganizationRequest]{
		Preloads: []string{"CreatedBy",
			"UpdatedBy", "Media", "CoverMedia",
			"SubscriptionPlan", "Branches",
			"OrganizationCategories", "OrganizationMedias", "OrganizationMedias.Media",
			"OrganizationCategories.Category",
			"Footsteps", "GeneratedReports", "InvitationCodes",
			"PermissionTemplates"},
		Service: m.provider.Service,
		Resource: func(data *Organization) *OrganizationResponse {
			if data == nil {
				return nil
			}
			return &OrganizationResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				CreatedByID: data.CreatedByID,
				CreatedBy:   m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID: data.UpdatedByID,
				UpdatedBy:   m.UserManager.ToModel(data.UpdatedBy),

				Name:               data.Name,
				Address:            data.Address,
				Email:              data.Email,
				ContactNumber:      data.ContactNumber,
				Description:        data.Description,
				Color:              data.Color,
				TermsAndConditions: data.TermsAndConditions,
				PrivacyPolicy:      data.PrivacyPolicy,
				CookiePolicy:       data.CookiePolicy,
				RefundPolicy:       data.RefundPolicy,
				UserAgreement:      data.UserAgreement,
				IsPrivate:          data.IsPrivate,

				MediaID:    data.MediaID,
				Media:      m.MediaManager.ToModel(data.Media),
				CoverMedia: m.MediaManager.ToModel(data.CoverMedia),

				SubscriptionPlanMaxBranches:         data.SubscriptionPlanMaxBranches,
				SubscriptionPlanMaxEmployees:        data.SubscriptionPlanMaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: data.SubscriptionPlanMaxMembersPerBranch,
				SubscriptionPlanID:                  data.SubscriptionPlanID,
				SubscriptionPlanIsYearly:            false, // TODO
				SubscriptionPlan:                    m.SubscriptionPlanManager.ToModel(data.SubscriptionPlan),
				SubscriptionStartDate:               data.SubscriptionStartDate.Format(time.RFC3339),
				SubscriptionEndDate:                 data.SubscriptionEndDate.Format(time.RFC3339),

				Branches:               m.BranchManager.ToModels(data.Branches),
				OrganizationCategories: m.OrganizationCategoryManager.ToModels(data.OrganizationCategories),
				OrganizationMedias:     m.OrganizationMediaManager.ToModels(data.OrganizationMedias),
				Footsteps:              m.FootstepManager.ToModels(data.Footsteps),
				GeneratedReports:       m.GeneratedReportManager.ToModels(data.GeneratedReports),
				InvitationCodes:        m.InvitationCodeManager.ToModels(data.InvitationCodes),
				PermissionTemplates:    m.PermissionTemplateManager.ToModels(data.PermissionTemplates),
				UserOrganizations:      m.UserOrganizationManager.ToModels(data.UserOrganizations),
			}
		},

		Created: func(data *Organization) []string {
			return []string{
				"organization.create",
				fmt.Sprintf("organization.create.%s", data.ID),
			}
		},
		Updated: func(data *Organization) []string {
			return []string{
				"organization.update",
				fmt.Sprintf("organization.update.%s", data.ID),
			}
		},
		Deleted: func(data *Organization) []string {
			return []string{
				"organization.delete",
				fmt.Sprintf("organization.delete.%s", data.ID),
			}
		},
	})
}

func (m *ModelCore) GetPublicOrganization(ctx context.Context) ([]*Organization, error) {
	filters := []horizon_services.Filter{
		{Field: "is_private", Op: horizon_services.OpEq, Value: false},
	}
	organizations, err := m.OrganizationManager.FindWithFilters(ctx, filters, "OrganizationCategories", "OrganizationCategories.Category")
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (m *ModelCore) GetFeaturedOrganization(ctx context.Context) ([]*Organization, error) {
	// Featured organizations are:
	// 1. Public (not private)
	// 2. Have a cover media (more visually appealing)
	// 3. Have multiple branches (indicates established organization)
	// 4. Have a description (complete profile)
	filters := []horizon_services.Filter{
		{Field: "is_private", Op: horizon_services.OpEq, Value: false},
		{Field: "cover_media_id", Op: horizon_services.OpNotNull, Value: nil},
		{Field: "description", Op: horizon_services.OpNotNull, Value: nil},
	}

	// Use a custom query to include organizations with multiple branches
	organizations, err := m.OrganizationManager.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Filter organizations that have at least 2 branches for "featured" status
	var featuredOrganizations []*Organization
	for _, org := range organizations {
		if len(org.Branches) >= 2 {
			featuredOrganizations = append(featuredOrganizations, org)
		}
	}

	// Limit to top 10 featured organizations
	if len(featuredOrganizations) > 10 {
		featuredOrganizations = featuredOrganizations[:10]
	}

	return featuredOrganizations, nil
}

func (m *ModelCore) GetOrganizationsByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*Organization, error) {
	// Get organization categories that match the category ID
	orgCategories, err := m.OrganizationCategoryManager.Find(ctx, &OrganizationCategory{
		CategoryID: &categoryID,
	})
	if err != nil {
		return nil, err
	}

	var organizations []*Organization
	// Get all organizations for each organization category
	for _, orgCat := range orgCategories {
		if orgCat.OrganizationID != nil {
			org, err := m.OrganizationManager.GetByID(ctx, *orgCat.OrganizationID)
			if err == nil && org != nil {
				// Only include public organizations
				if !org.IsPrivate {
					organizations = append(organizations, org)
				}
			}
		}
	}

	// Convert to response models
	return organizations, nil
}

func (m *ModelCore) GetRecentlyAddedOrganization(ctx context.Context) ([]*Organization, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	filters := []horizon_services.Filter{
		{Field: "is_private", Op: horizon_services.OpEq, Value: false},
		{Field: "created_at", Op: horizon_services.OpGte, Value: thirtyDaysAgo},
	}
	organizations, err := m.OrganizationManager.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(organizations)-1; i++ {
		for j := i + 1; j < len(organizations); j++ {
			if organizations[i].CreatedAt.Before(organizations[j].CreatedAt) {
				organizations[i], organizations[j] = organizations[j], organizations[i]
			}
		}
	}
	if len(organizations) > 15 {
		organizations = organizations[:15]
	}
	return organizations, nil
}
