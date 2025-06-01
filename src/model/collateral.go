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
	Collateral struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collateral"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collateral"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Icon        string `gorm:"type:varchar(255)"` // from React Icons, e.g. "FaCar", "MdHome"
		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text"`
	}

	CollateralResponse struct {
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
		Icon           string                `json:"icon"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	CollateralRequest struct {
		Icon        string `json:"icon,omitempty"`
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Model) Collateral() {
	m.Migration = append(m.Migration, &Collateral{})
	m.CollateralManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Collateral, CollateralResponse, CollateralRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *Collateral) *CollateralResponse {
			if data == nil {
				return nil
			}
			return &CollateralResponse{
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
				Icon:           data.Icon,
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *Collateral) []string {
			return []string{
				"collateral.create",
				fmt.Sprintf("collateral.create.%s", data.ID),
				fmt.Sprintf("collateral.create.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Collateral) []string {
			return []string{
				"collateral.update",
				fmt.Sprintf("collateral.update.%s", data.ID),
				fmt.Sprintf("collateral.update.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Collateral) []string {
			return []string{
				"collateral.delete",
				fmt.Sprintf("collateral.delete.%s", data.ID),
				fmt.Sprintf("collateral.delete.branch.%s", data.BranchID),
				fmt.Sprintf("collateral.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) CollateralCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Collateral, error) {
	return m.CollateralManager.Find(context, &Collateral{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
