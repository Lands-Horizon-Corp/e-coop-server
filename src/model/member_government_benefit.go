package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberGovernmentBenefit struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_government_benefits"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_government_benefits"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		FrontMediaID *uuid.UUID `gorm:"type:uuid"`
		FrontMedia   *Media     `gorm:"foreignKey:FrontMediaID;constraint:OnDelete:SET NULL;" json:"front_media,omitempty"`

		BackMediaID *uuid.UUID `gorm:"type:uuid"`
		BackMedia   *Media     `gorm:"foreignKey:BackMediaID;constraint:OnDelete:SET NULL;" json:"back_media,omitempty"`

		CountryCode string     `gorm:"type:varchar(4)"`
		Description string     `gorm:"type:text"`
		Name        string     `gorm:"type:varchar(254)"`
		Value       string     `gorm:"type:varchar(254);unique;not null"`
		ExpiryDate  *time.Time `gorm:"type:date"`
	}

	MemberGovernmentBenefitResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		FrontMediaID    *uuid.UUID             `json:"front_media_id,omitempty"`
		FrontMedia      *MediaResponse         `json:"front_media,omitempty"`
		BackMediaID     *uuid.UUID             `json:"back_media_id,omitempty"`
		BackMedia       *MediaResponse         `json:"back_media,omitempty"`
		CountryCode     string                 `json:"country_code"`
		Description     string                 `json:"description"`
		Name            string                 `json:"name"`
		Value           string                 `json:"value"`
		ExpiryDate      *string                `json:"expiry_date,omitempty"`
	}

	MemberGovernmentBenefitRequest struct {
		MemberProfileID uuid.UUID  `json:"member_profile_id" validate:"required"`
		FrontMediaID    *uuid.UUID `json:"front_media_id,omitempty"`
		BackMediaID     *uuid.UUID `json:"back_media_id,omitempty"`
		CountryCode     string     `json:"country_code,omitempty"`
		Description     string     `json:"description,omitempty"`
		Name            string     `json:"name,omitempty"`
		Value           string     `json:"value" validate:"required,min=1,max=254"`
		ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	}
)

func (m *Model) MemberGovernmentBenefit() {
	m.Migration = append(m.Migration, &MemberGovernmentBenefit{})
	m.MemberGovernmentBenefitManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberGovernmentBenefit, MemberGovernmentBenefitResponse, MemberGovernmentBenefitRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberProfile", "FrontMedia", "BackMedia"},
		Service:  m.provider.Service,
		Resource: func(data *MemberGovernmentBenefit) *MemberGovernmentBenefitResponse {
			if data == nil {
				return nil
			}
			var expiryDateStr *string
			if data.ExpiryDate != nil {
				s := data.ExpiryDate.Format("2006-01-02")
				expiryDateStr = &s
			}
			return &MemberGovernmentBenefitResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				FrontMediaID:    data.FrontMediaID,
				FrontMedia:      m.MediaManager.ToModel(data.FrontMedia),
				BackMediaID:     data.BackMediaID,
				BackMedia:       m.MediaManager.ToModel(data.BackMedia),
				CountryCode:     data.CountryCode,
				Description:     data.Description,
				Name:            data.Name,
				Value:           data.Value,
				ExpiryDate:      expiryDateStr,
			}
		},

		Created: func(data *MemberGovernmentBenefit) []string {
			return []string{
				"member_government_benefit.create",
				fmt.Sprintf("member_government_benefit.create.%s", data.ID),
				fmt.Sprintf("member_government_benefit.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberGovernmentBenefit) []string {
			return []string{
				"member_government_benefit.update",
				fmt.Sprintf("member_government_benefit.update.%s", data.ID),
				fmt.Sprintf("member_government_benefit.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberGovernmentBenefit) []string {
			return []string{
				"member_government_benefit.delete",
				fmt.Sprintf("member_government_benefit.delete.%s", data.ID),
				fmt.Sprintf("member_government_benefit.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_government_benefit.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberGovernmentBenefitCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberGovernmentBenefit, error) {
	return m.MemberGovernmentBenefitManager.Find(context, &MemberGovernmentBenefit{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
