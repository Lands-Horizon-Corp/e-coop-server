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

		Name               string  `gorm:"type:varchar(255);not null"`
		Address            *string `gorm:"type:varchar(500)"`
		Email              *string `gorm:"type:varchar(255)"`
		ContactNumber      *string `gorm:"type:varchar(20)"`
		Description        *string `gorm:"type:text"`
		Color              *string `gorm:"type:varchar(50)"`
		TermsAndConditions *string `gorm:"type:text"`
		PrivacyPolicy      *string `gorm:"type:text"`
		CookiePolicy       *string `gorm:"type:text"`
		RefundPolicy       *string `gorm:"type:text"`
		UserAgreement      *string `gorm:"type:text"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		CoverMediaID *uuid.UUID `gorm:"type:uuid"`
		CoverMedia   *Media     `gorm:"foreignKey:CoverMediaID;constraint:OnDelete:SET NULL;" json:"cover_media,omitempty"`

		OrganizationKey string `gorm:"type:varchar(255);not null;unique"`

		SubscriptionPlanID *uuid.UUID        `gorm:"type:uuid"`
		SubscriptionPlan   *SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;constraint:OnDelete:SET NULL"`

		SubscriptionStartDate time.Time
		SubscriptionEndDate   time.Time

		OrganizationCategories []*OrganizationCategory `gorm:"foreignKey:OrganizationID"`
		Branches               []*Branch               `gorm:"foreignKey:OrganizationID"`
	}

	OrganizationRequest struct {
		Name               string  `json:"name" validate:"required,min=1,max=255"`
		Address            *string `json:"address,omitempty"`
		Email              *string `json:"email,omitempty" validate:"omitempty,email"`
		ContactNumber      *string `json:"contact_number,omitempty"`
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

		DatabaseHost            *string `json:"database_host,omitempty"`
		DatabasePort            *string `json:"database_port,omitempty"`
		DatabaseName            *string `json:"database_name,omitempty"`
		DatabasePassword        *string `json:"database_password,omitempty"`
		DatabaseMigrationStatus string  `json:"database_migration_status" validate:"required"`
		DatabaseRemark          *string `json:"database_remark,omitempty"`
	}

	OrganizationResponse struct {
		ID                 uuid.UUID `json:"id"`
		Name               string    `json:"name"`
		Address            *string   `json:"address,omitempty"`
		Email              *string   `json:"email,omitempty"`
		ContactNumber      *string   `json:"contact_number,omitempty"`
		Description        *string   `json:"description,omitempty"`
		Color              *string   `json:"color,omitempty"`
		TermsAndConditions *string   `json:"terms_and_conditions,omitempty"`
		PrivacyPolicy      *string   `json:"privacy_policy,omitempty"`
		CookiePolicy       *string   `json:"cookie_policy,omitempty"`
		RefundPolicy       *string   `json:"refund_policy,omitempty"`
		UserAgreement      *string   `json:"user_agreement,omitempty"`

		MediaID      *uuid.UUID     `json:"media_id,omitempty"`
		Media        *MediaResponse `json:"media,omitempty"`
		CoverMediaID *uuid.UUID     `json:"cover_media_id,omitempty"`
		CoverMedia   *MediaResponse `json:"cover_media,omitempty"`

		OrganizationKey       string                    `json:"organization_key"`
		SubscriptionPlanID    *uuid.UUID                `json:"subscription_plan_id,omitempty"`
		SubscriptionPlan      *SubscriptionPlanResponse `json:"subscription_plan,omitempty"`
		SubscriptionStartDate string                    `json:"subscription_start_date"`
		SubscriptionEndDate   string                    `json:"subscription_end_date"`

		DatabaseHost            *string `json:"database_host,omitempty"`
		DatabasePort            *string `json:"database_port,omitempty"`
		DatabaseName            *string `json:"database_name,omitempty"`
		DatabasePassword        *string `json:"database_password,omitempty"`
		DatabaseMigrationStatus string  `json:"database_migration_status"`
		DatabaseRemark          *string `json:"database_remark,omitempty"`

		Branches               []*BranchResponse               `json:"branches,omitempty"`
		OrganizationCategories []*OrganizationCategoryResponse `json:"organizaton_categories"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	OrganizationCollection struct {
		validator      *validator.Validate
		mediaCol       *MediaCollection
		subPlanCol     *SubscriptionPlanCollection
		branchCol      *BranchCollection
		orgCategoryCol *OrganizationCategoryCollection
	}
)

func NewOrganizationCollection(
	mediaCol *MediaCollection,
	subPlanCol *SubscriptionPlanCollection,
	branchCol *BranchCollection,
	orgCategoryCol *OrganizationCategoryCollection,
) (*OrganizationCollection, error) {
	return &OrganizationCollection{
		validator:      validator.New(),
		mediaCol:       mediaCol,
		subPlanCol:     subPlanCol,
		branchCol:      branchCol,
		orgCategoryCol: orgCategoryCol,
	}, nil
}

// ValidateCreate binds and validates the request payload
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

// ToModel maps a DB Organization to an OrganizationResponse
func (oc *OrganizationCollection) ToModel(o *Organization) *OrganizationResponse {
	if o == nil {
		return nil
	}

	resp := &OrganizationResponse{
		ID:                 o.ID,
		Name:               o.Name,
		Address:            o.Address,
		Email:              o.Email,
		ContactNumber:      o.ContactNumber,
		Description:        o.Description,
		Color:              o.Color,
		TermsAndConditions: o.TermsAndConditions,
		PrivacyPolicy:      o.PrivacyPolicy,
		CookiePolicy:       o.CookiePolicy,
		RefundPolicy:       o.RefundPolicy,
		UserAgreement:      o.UserAgreement,

		MediaID:      o.MediaID,
		Media:        oc.mediaCol.ToModel(o.Media),
		CoverMediaID: o.CoverMediaID,
		CoverMedia:   oc.mediaCol.ToModel(o.CoverMedia),

		OrganizationKey:       o.OrganizationKey,
		SubscriptionPlanID:    o.SubscriptionPlanID,
		SubscriptionPlan:      oc.subPlanCol.ToModel(o.SubscriptionPlan),
		SubscriptionStartDate: o.SubscriptionStartDate.Format(time.RFC3339),
		SubscriptionEndDate:   o.SubscriptionEndDate.Format(time.RFC3339),

		Branches:               oc.branchCol.ToModels(o.Branches),
		OrganizationCategories: oc.orgCategoryCol.ToModels(o.OrganizationCategories),

		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
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
