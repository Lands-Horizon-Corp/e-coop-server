package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	Organization struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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

		Branches               []*Branch                 `gorm:"foreignKey:OrganizationID" json:"branches,omitempty"`
		OrganizationCategories []*OrganizationCategory   `gorm:"foreignKey:OrganizationID" json:"organization_categories,omitempty"`
		Footsteps              []*Footstep               `gorm:"foreignKey:OrganizationID" json:"footsteps,omitempty"`
		GeneratedReports       []*GeneratedReport        `gorm:"foreignKey:OrganizationID" json:"generated_reports,omitempty"`
		InvitationCodes        []*InvitationCode         `gorm:"foreignKey:OrganizationID" json:"invitation_codes,omitempty"`
		OrganizationDailyUsage []*OrganizationDailyUsage `gorm:"foreignKey:OrganizationID" json:"organization_daily_usage,omitempty"`
		PermissionTemplates    []*PermissionTemplate     `gorm:"foreignKey:OrganizationID" json:"permission_templates,omitempty"`
		UserOrganizations      []*UserOrganization       `gorm:"foreignKey:OrganizationID" json:"user_organizations,omitempty"`
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

		Branches               []*BranchResponse                 `json:"branches,omitempty"`
		OrganizationCategories []*OrganizationCategoryResponse   `json:"organization_categories,omitempty"`
		Footsteps              []*FootstepResponse               `json:"footsteps,omitempty"`
		GeneratedReports       []*GeneratedReportResponse        `json:"generated_reports,omitempty"`
		InvitationCodes        []*InvitationCodeResponse         `json:"invitation_codes,omitempty"`
		OrganizationDailyUsage []*OrganizationDailyUsageResponse `json:"organization_daily_usage,omitempty"`
		PermissionTemplates    []*PermissionTemplateResponse     `json:"permission_templates,omitempty"`
		UserOrganizations      []*UserOrganizationResponse       `json:"user_organizations,omitempty"`
	}

	OrganizationRequest struct {
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

		OrganizationCategories []*uuid.UUID `json:"organization_categories,omitempty"`
	}

	OrganizationSubscriptionRequest struct {
		OrganizationID           uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		SubscriptionPlanID       uuid.UUID `json:"subscription_plan_id" validate:"required,uuid4"`
		SubscriptionPlanIsYearly *bool     `json:"subscription_plan_is_yearly,omitempty"`
	}

	OrganizationCollection struct {
		Manager CollectionManager[Organization]
	}
)

func (m *Model) OrganizationValidate(ctx echo.Context) (*OrganizationRequest, error) {
	return Validate[OrganizationRequest](ctx, m.validator)
}
func (m *Model) OrganizationSubscriptionValidate(ctx echo.Context) (*OrganizationSubscriptionRequest, error) {
	return Validate[OrganizationSubscriptionRequest](ctx, m.validator)
}

func (m *Model) OrganizationModel(data *Organization) *OrganizationResponse {
	return ToModel(data, func(data *Organization) *OrganizationResponse {
		return &OrganizationResponse{
			ID:          data.ID,
			CreatedAt:   data.CreatedAt.Format(time.RFC3339),
			CreatedByID: data.CreatedByID,
			CreatedBy:   m.UserModel(data.CreatedBy),
			UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID: data.UpdatedByID,
			UpdatedBy:   m.UserModel(data.UpdatedBy),

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

			MediaID:      data.MediaID,
			Media:        m.MediaModel(data.Media),
			CoverMediaID: data.CoverMediaID,
			CoverMedia:   m.MediaModel(data.CoverMedia),

			SubscriptionPlanMaxBranches:         data.SubscriptionPlanMaxBranches,
			SubscriptionPlanMaxEmployees:        data.SubscriptionPlanMaxEmployees,
			SubscriptionPlanMaxMembersPerBranch: data.SubscriptionPlanMaxMembersPerBranch,
			SubscriptionPlan:                    m.SubscriptionPlanModel(data.SubscriptionPlan),
			SubscriptionPlanID:                  data.SubscriptionPlanID,
			SubscriptionStartDate:               data.SubscriptionStartDate.Format(time.RFC3339),
			SubscriptionEndDate:                 data.SubscriptionEndDate.Format(time.RFC3339),

			Branches:               m.BranchModels(data.Branches),
			OrganizationCategories: m.OrganizationCategoryModels(data.OrganizationCategories),
			Footsteps:              m.FootstepModels(data.Footsteps),
			GeneratedReports:       m.GeneratedReportModels(data.GeneratedReports),
			InvitationCodes:        m.InvitationCodeModels(data.InvitationCodes),
			OrganizationDailyUsage: m.OrganizationDailyUsageModels(data.OrganizationDailyUsage),
			PermissionTemplates:    m.PermissionTemplateModels(data.PermissionTemplates),
			UserOrganizations:      m.UserOrganizationModels(data.UserOrganizations),
		}
	})
}

func (m *Model) OrganizationModels(data []*Organization) []*OrganizationResponse {
	return ToModels(data, m.OrganizationModel)
}
