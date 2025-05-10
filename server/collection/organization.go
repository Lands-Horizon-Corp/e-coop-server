package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	Organization struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

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

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id,omitempty"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		CoverMediaID *uuid.UUID `gorm:"type:uuid" json:"cover_media_id,omitempty"`
		CoverMedia   *Media     `gorm:"foreignKey:CoverMediaID;constraint:OnDelete:SET NULL;" json:"cover_media,omitempty"`

		OrganizationKey string `gorm:"type:varchar(255);not null;unique" json:"organization_key"`

		SubscriptionPlanID    *uuid.UUID        `gorm:"type:uuid" json:"subscription_plan_id,omitempty"`
		SubscriptionPlan      *SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;constraint:OnDelete:SET NULL;" json:"subscription_plan,omitempty"`
		SubscriptionStartDate time.Time         `json:"subscription_start_date"`
		SubscriptionEndDate   time.Time         `json:"subscription_end_date"`

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
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`

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

		MediaID      *uuid.UUID     `json:"media_id,omitempty"`
		Media        *MediaResponse `json:"media,omitempty"`
		CoverMediaID *uuid.UUID     `json:"cover_media_id,omitempty"`
		CoverMedia   *MediaResponse `json:"cover_media,omitempty"`

		OrganizationKey       string                    `json:"organization_key"`
		SubscriptionPlanID    *uuid.UUID                `json:"subscription_plan_id,omitempty"`
		SubscriptionPlan      *SubscriptionPlanResponse `json:"subscription_plan,omitempty"`
		SubscriptionStartDate string                    `json:"subscription_start_date"`
		SubscriptionEndDate   string                    `json:"subscription_end_date"`

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

		MediaID      *uuid.UUID `json:"media_id,omitempty"`
		CoverMediaID *uuid.UUID `json:"cover_media_id,omitempty"`

		OrganizationKey       string     `json:"organization_key" validate:"required,min=1"`
		SubscriptionPlanID    *uuid.UUID `json:"subscription_plan_id,omitempty"`
		SubscriptionStartDate time.Time  `json:"subscription_start_date" validate:"required"`
		SubscriptionEndDate   time.Time  `json:"subscription_end_date" validate:"required"`
	}

	OrganizationCollection struct {
		validator              *validator.Validate
		media                  *MediaCollection
		subscriptionPlan       *SubscriptionPlanCollection
		branch                 *BranchCollection
		organzationCategory    *OrganizationCategoryCollection
		footstep               *FootstepCollection
		generatedReport        *GeneratedReportCollection
		invitationCode         *InvitationCodeCollection
		organizationDailyUsage *OrganizationDailyUsageCollection
		permissionTemplate     *PermissionTemplateCollection
		userOrganization       *UserOrganizationCollection
	}
)

func NewOrganizationCollection(
	media *MediaCollection,
	subscriptionPlan *SubscriptionPlanCollection,
	branch *BranchCollection,
	organzationCategory *OrganizationCategoryCollection,
	footstep *FootstepCollection,
	generatedReport *GeneratedReportCollection,
	invitationCode *InvitationCodeCollection,
	organizationDailyUsage *OrganizationDailyUsageCollection,
	permissionTemplate *PermissionTemplateCollection,
	userOrganization *UserOrganizationCollection,
) (*OrganizationCollection, error) {
	return &OrganizationCollection{
		validator:              validator.New(),
		media:                  media,
		subscriptionPlan:       subscriptionPlan,
		branch:                 branch,
		organzationCategory:    organzationCategory,
		footstep:               footstep,
		generatedReport:        generatedReport,
		invitationCode:         invitationCode,
		organizationDailyUsage: organizationDailyUsage,
		permissionTemplate:     permissionTemplate,
		userOrganization:       userOrganization,
	}, nil
}

func (oc *OrganizationCollection) ValidateCreate(c echo.Context) (*OrganizationRequest, error) {
	req := new(OrganizationRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := oc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (oc *OrganizationCollection) ToModel(o *Organization) *OrganizationResponse {
	if o == nil {
		return nil
	}

	resp := &OrganizationResponse{
		ID:                    o.ID,
		CreatedAt:             o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             o.UpdatedAt.Format(time.RFC3339),
		Name:                  o.Name,
		Address:               o.Address,
		Email:                 o.Email,
		ContactNumber:         o.ContactNumber,
		Description:           o.Description,
		Color:                 o.Color,
		TermsAndConditions:    o.TermsAndConditions,
		PrivacyPolicy:         o.PrivacyPolicy,
		CookiePolicy:          o.CookiePolicy,
		RefundPolicy:          o.RefundPolicy,
		UserAgreement:         o.UserAgreement,
		MediaID:               o.MediaID,
		CoverMediaID:          o.CoverMediaID,
		OrganizationKey:       o.OrganizationKey,
		SubscriptionPlanID:    o.SubscriptionPlanID,
		SubscriptionStartDate: o.SubscriptionStartDate.Format(time.RFC3339),
		SubscriptionEndDate:   o.SubscriptionEndDate.Format(time.RFC3339),

		Branches:               oc.branch.ToModels(o.Branches),
		OrganizationCategories: oc.organzationCategory.ToModels(o.OrganizationCategories),
		Footsteps:              oc.footstep.ToModels(o.Footsteps),
		GeneratedReports:       oc.generatedReport.ToModels(o.GeneratedReports),
		InvitationCodes:        oc.invitationCode.ToModels(o.InvitationCodes),
		OrganizationDailyUsage: oc.organizationDailyUsage.ToModels(o.OrganizationDailyUsage),
		PermissionTemplates:    oc.permissionTemplate.ToModels(o.PermissionTemplates),
		UserOrganizations:      oc.userOrganization.ToModels(o.UserOrganizations),
	}
	return resp
}

// ToModels maps multiple DB Organizations to responses
func (oc *OrganizationCollection) ToModels(data []*Organization) []*OrganizationResponse {
	if data == nil {
		return []*OrganizationResponse{}
	}
	var out []*OrganizationResponse
	for _, o := range data {
		if m := oc.ToModel(o); m != nil {
			out = append(out, m)
		}
	}
	return out
}
