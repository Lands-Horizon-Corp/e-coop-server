package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// MutualAidContribution represents the MutualAidContribution model.
	MutualAidContribution struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_aid_contribution" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_mutual_aid_contribution" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MonthsFrom int     `gorm:"not null" json:"months_from"`
		MonthsTo   int     `gorm:"not null" json:"months_to"`
		Amount     float64 `gorm:"type:decimal(15,2);not null" json:"amount"`
	}

	// MutualAidContributionResponse represents the response structure for MutualAidContribution.
	MutualAidContributionResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`
		MonthsFrom     int                   `json:"months_from"`
		MonthsTo       int                   `json:"months_to"`
		Amount         float64               `json:"amount"`
	}

	// MutualAidContributionRequest represents the request structure for MutualAidContribution.
	MutualAidContributionRequest struct {
		MonthsFrom int     `json:"months_from" validate:"required,min=0"`
		MonthsTo   int     `json:"months_to" validate:"required,min=0"`
		Amount     float64 `json:"amount" validate:"required,min=0"`
	}
)

func (m *Core) mutualAidContribution() {
	m.Migration = append(m.Migration, &MutualAidContribution{})
	m.MutualAidContributionManager = *registry.NewRegistry(registry.RegistryParams[MutualAidContribution, MutualAidContributionResponse, MutualAidContributionRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Organization", "Branch"},
		Service:  m.provider.Service,
		Resource: func(data *MutualAidContribution) *MutualAidContributionResponse {
			if data == nil {
				return nil
			}
			return &MutualAidContributionResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				MonthsFrom:     data.MonthsFrom,
				MonthsTo:       data.MonthsTo,
				Amount:         data.Amount,
			}
		},
		Created: func(data *MutualAidContribution) []string {
			return []string{
				"mutual_aid_contribution.create",
				fmt.Sprintf("mutual_aid_contribution.create.%s", data.ID),
				fmt.Sprintf("mutual_aid_contribution.create.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_aid_contribution.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MutualAidContribution) []string {
			return []string{
				"mutual_aid_contribution.update",
				fmt.Sprintf("mutual_aid_contribution.update.%s", data.ID),
				fmt.Sprintf("mutual_aid_contribution.update.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_aid_contribution.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MutualAidContribution) []string {
			return []string{
				"mutual_aid_contribution.delete",
				fmt.Sprintf("mutual_aid_contribution.delete.%s", data.ID),
				fmt.Sprintf("mutual_aid_contribution.delete.branch.%s", data.BranchID),
				fmt.Sprintf("mutual_aid_contribution.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// MutualAidContributionCurrentBranch retrieves all mutual aid contributions associated with the specified organization and branch.
func (m *Core) MutualAidContributionCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MutualAidContribution, error) {
	return m.MutualAidContributionManager.Find(context, &MutualAidContribution{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
