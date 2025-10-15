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
	MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_interest_rate_by_ultima_membership_date_per_year"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_interest_rate_by_ultima_membership_date_per_year"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberTypeReferenceID uuid.UUID            `gorm:"type:uuid;not null"`
		MemberTypeReference   *MemberTypeReference `gorm:"foreignKey:MemberTypeReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_type_reference,omitempty"`

		YearFrom int     `gorm:"type:int"`
		YearTo   int     `gorm:"type:int"`
		Rate     float64 `gorm:"type:decimal;default:0"`
	}

	MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse struct {
		ID                    uuid.UUID                    `json:"id"`
		CreatedAt             string                       `json:"created_at"`
		CreatedByID           uuid.UUID                    `json:"created_by_id"`
		CreatedBy             *UserResponse                `json:"created_by,omitempty"`
		UpdatedAt             string                       `json:"updated_at"`
		UpdatedByID           uuid.UUID                    `json:"updated_by_id"`
		UpdatedBy             *UserResponse                `json:"updated_by,omitempty"`
		OrganizationID        uuid.UUID                    `json:"organization_id"`
		Organization          *OrganizationResponse        `json:"organization,omitempty"`
		BranchID              uuid.UUID                    `json:"branch_id"`
		Branch                *BranchResponse              `json:"branch,omitempty"`
		MemberTypeReferenceID uuid.UUID                    `json:"member_type_reference_id"`
		MemberTypeReference   *MemberTypeReferenceResponse `json:"member_type_reference,omitempty"`
		YearFrom              int                          `json:"year_from"`
		YearTo                int                          `json:"year_to"`
		Rate                  float64                      `json:"rate"`
	}

	MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearRequest struct {
		MemberTypeReferenceID uuid.UUID `json:"member_type_reference_id" validate:"required"`
		YearFrom              int       `json:"year_from,omitempty"`
		YearTo                int       `json:"year_to,omitempty"`
		Rate                  float64   `json:"rate,omitempty"`
	}
)

func (m *ModelCore) MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear() {
	m.Migration = append(m.Migration, &MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear{})
	m.MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear,
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse,
		MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberTypeReference",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear) *MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearResponse{
				ID:                    data.ID,
				CreatedAt:             data.CreatedAt.Format(time.RFC3339),
				CreatedByID:           data.CreatedByID,
				CreatedBy:             m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:             data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:           data.UpdatedByID,
				UpdatedBy:             m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:        data.OrganizationID,
				Organization:          m.OrganizationManager.ToModel(data.Organization),
				BranchID:              data.BranchID,
				Branch:                m.BranchManager.ToModel(data.Branch),
				MemberTypeReferenceID: data.MemberTypeReferenceID,
				MemberTypeReference:   m.MemberTypeReferenceManager.ToModel(data.MemberTypeReference),
				YearFrom:              data.YearFrom,
				YearTo:                data.YearTo,
				Rate:                  data.Rate,
			}
		},

		Created: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date_per_year.create",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.create.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date_per_year.update",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.update.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date_per_year.delete",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.delete.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date_per_year.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear, error) {
	return m.MemberTypeReferenceInterestRateByUltimaMembershipDatePerYearManager.Find(context, &MemberTypeReferenceInterestRateByUltimaMembershipDatePerYear{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
