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
	MemberTypeReferenceByAmount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_by_amount"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_reference_by_amount"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberTypeReferenceID uuid.UUID            `gorm:"type:uuid;not null"`
		MemberTypeReference   *MemberTypeReference `gorm:"foreignKey:MemberTypeReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_type_reference,omitempty"`

		From float64 `gorm:"type:decimal;default:0"`
		To   float64 `gorm:"type:decimal;default:0"`
		Rate float64 `gorm:"type:decimal;default:0"`
	}

	MemberTypeReferenceByAmountResponse struct {
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
		From                  float64                      `json:"from"`
		To                    float64                      `json:"to"`
		Rate                  float64                      `json:"rate"`
	}

	MemberTypeReferenceByAmountRequest struct {
		MemberTypeReferenceID uuid.UUID `json:"member_type_reference_id" validate:"required"`
		From                  float64   `json:"from,omitempty"`
		To                    float64   `json:"to,omitempty"`
		Rate                  float64   `json:"rate,omitempty"`
	}
)

func (m *Model) MemberTypeReferenceByAmount() {
	m.Migration = append(m.Migration, &MemberTypeReferenceByAmount{})
	m.MemberTypeReferenceByAmountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberTypeReferenceByAmount, MemberTypeReferenceByAmountResponse, MemberTypeReferenceByAmountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberTypeReference",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberTypeReferenceByAmount) *MemberTypeReferenceByAmountResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeReferenceByAmountResponse{
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
				From:                  data.From,
				To:                    data.To,
				Rate:                  data.Rate,
			}
		},

		Created: func(data *MemberTypeReferenceByAmount) []string {
			return []string{
				"member_type_reference_by_amount.create",
				fmt.Sprintf("member_type_reference_by_amount.create.%s", data.ID),
				fmt.Sprintf("member_type_reference_by_amount.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_by_amount.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberTypeReferenceByAmount) []string {
			return []string{
				"member_type_reference_by_amount.update",
				fmt.Sprintf("member_type_reference_by_amount.update.%s", data.ID),
				fmt.Sprintf("member_type_reference_by_amount.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_by_amount.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberTypeReferenceByAmount) []string {
			return []string{
				"member_type_reference_by_amount.delete",
				fmt.Sprintf("member_type_reference_by_amount.delete.%s", data.ID),
				fmt.Sprintf("member_type_reference_by_amount.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_reference_by_amount.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberTypeReferenceByAmountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberTypeReferenceByAmount, error) {
	return m.MemberTypeReferenceByAmountManager.Find(context, &MemberTypeReferenceByAmount{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
