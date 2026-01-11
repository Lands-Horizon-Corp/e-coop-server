package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneralAccountGroupingNetSurplusPositive struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_account_grouping_net_surplus_positive"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_account_grouping_net_surplus_positive"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Name        string  `gorm:"type:varchar(255)"`
		Description string  `gorm:"type:text"`
		Percentage1 float64 `gorm:"type:decimal;default:0"`
		Percentage2 float64 `gorm:"type:decimal;default:0"`
	}

	GeneralAccountGroupingNetSurplusPositiveResponse struct {
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
		AccountID      uuid.UUID             `json:"account_id"`
		Account        *AccountResponse      `json:"account,omitempty"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		Percentage1    float64               `json:"percentage_1"`
		Percentage2    float64               `json:"percentage_2"`
	}

	GeneralAccountGroupingNetSurplusPositiveRequest struct {
		Name        string    `json:"name" validate:"required,min=1,max=255"`
		Description string    `json:"description,omitempty"`
		AccountID   uuid.UUID `json:"account_id" validate:"required"`
		Percentage1 float64   `json:"percentage_1,omitempty"`
		Percentage2 float64   `json:"percentage_2,omitempty"`
	}
)

func (m *Core) GeneralAccountGroupingNetSurplusPositiveManager() *registry.Registry[GeneralAccountGroupingNetSurplusPositive, GeneralAccountGroupingNetSurplusPositiveResponse, GeneralAccountGroupingNetSurplusPositiveRequest] {
	return registry.NewRegistry(registry.RegistryParams[GeneralAccountGroupingNetSurplusPositive, GeneralAccountGroupingNetSurplusPositiveResponse, GeneralAccountGroupingNetSurplusPositiveRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneralAccountGroupingNetSurplusPositive) *GeneralAccountGroupingNetSurplusPositiveResponse {
			if data == nil {
				return nil
			}
			return &GeneralAccountGroupingNetSurplusPositiveResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),
				AccountID:      data.AccountID,
				Account:        m.AccountManager().ToModel(data.Account),
				Name:           data.Name,
				Description:    data.Description,
				Percentage1:    data.Percentage1,
				Percentage2:    data.Percentage2,
			}
		},
		Created: func(data *GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.create",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.update",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.delete",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) GeneralAccountGroupingNetSurplusPositiveCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralAccountGroupingNetSurplusPositive, error) {
	return m.GeneralAccountGroupingNetSurplusPositiveManager().Find(context, &GeneralAccountGroupingNetSurplusPositive{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
