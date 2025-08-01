package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberTypeReferenceInterestRateByUltimaMembershipDate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_interest_rate_by_ultima_membership_date"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_interest_rate_by_ultima_membership_date"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberTypeReferenceID uuid.UUID            `gorm:"type:uuid;not null"`
		MemberTypeReference   *MemberTypeReference `gorm:"foreignKey:MemberTypeReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_type_reference,omitempty"`

		DateFrom time.Time `gorm:"type:timestamp"`
		DateTo   time.Time `gorm:"type:timestamp"`
		Rate     float64   `gorm:"type:decimal;default:0"`
	}

	MemberTypeReferenceInterestRateByUltimaMembershipDateResponse struct {
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
		DateFrom              string                       `json:"date_from"`
		DateTo                string                       `json:"date_to"`
		Rate                  float64                      `json:"rate"`
	}

	MemberTypeReferenceInterestRateByUltimaMembershipDateRequest struct {
		MemberTypeReferenceID uuid.UUID `json:"member_type_reference_id" validate:"required"`
		DateFrom              time.Time `json:"date_from,omitempty"`
		DateTo                time.Time `json:"date_to,omitempty"`
		Rate                  float64   `json:"rate,omitempty"`
	}
)

func (m *Model) MemberTypeReferenceInterestRateByUltimaMembershipDate() {
	m.Migration = append(m.Migration, &MemberTypeReferenceInterestRateByUltimaMembershipDate{})
	m.MemberTypeReferenceInterestRateByUltimaMembershipDateManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberTypeReferenceInterestRateByUltimaMembershipDate,
		MemberTypeReferenceInterestRateByUltimaMembershipDateResponse,
		MemberTypeReferenceInterestRateByUltimaMembershipDateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberTypeReference",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDate) *MemberTypeReferenceInterestRateByUltimaMembershipDateResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeReferenceInterestRateByUltimaMembershipDateResponse{
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
				DateFrom:              data.DateFrom.Format(time.RFC3339),
				DateTo:                data.DateTo.Format(time.RFC3339),
				Rate:                  data.Rate,
			}
		},

		Created: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDate) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date.create",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.create.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDate) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date.update",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.update.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberTypeReferenceInterestRateByUltimaMembershipDate) []string {
			return []string{
				"member_type_reference_interest_rate_by_ultima_membership_date.delete",
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.delete.%s", data.ID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_interest_rate_by_ultima_membership_date.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberTypeReferenceInterestRateByUltimaMembershipDateCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberTypeReferenceInterestRateByUltimaMembershipDate, error) {
	return m.MemberTypeReferenceInterestRateByUltimaMembershipDateManager.Find(context, &MemberTypeReferenceInterestRateByUltimaMembershipDate{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
