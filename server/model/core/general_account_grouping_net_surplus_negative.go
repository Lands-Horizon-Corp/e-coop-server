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
	GeneralAccountGroupingNetSurplusNegative struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_account_grouping_net_surplus_negative"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_account_grouping_net_surplus_negative"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Name        string  `gorm:"type:varchar(255)"`
		Description string  `gorm:"type:text"`
		Percentage1 float64 `gorm:"type:decimal;default:0"`
		Percentage2 float64 `gorm:"type:decimal;default:0"`
	}

	GeneralAccountGroupingNetSurplusNegativeResponse struct {
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

	GeneralAccountGroupingNetSurplusNegativeRequest struct {
		Name        string    `json:"name" validate:"required,min=1,max=255"`
		Description string    `json:"description,omitempty"`
		AccountID   uuid.UUID `json:"account_id" validate:"required"`
		Percentage1 float64   `json:"percentage_1,omitempty"`
		Percentage2 float64   `json:"percentage_2,omitempty"`
	}
)

func (m *Core) generalAccountGroupingNetSurplusNegative() {
	m.Migration = append(m.Migration, &GeneralAccountGroupingNetSurplusNegative{})
	m.GeneralaccountgroupingnetsurplusnegativeManager().= registry.NewRegistry(registry.RegistryParams[
		GeneralAccountGroupingNetSurplusNegative,
		GeneralAccountGroupingNetSurplusNegativeResponse,
		GeneralAccountGroupingNetSurplusNegativeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneralAccountGroupingNetSurplusNegative) *GeneralAccountGroupingNetSurplusNegativeResponse {
			if data == nil {
				return nil
			}
			return &GeneralAccountGroupingNetSurplusNegativeResponse{
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
		Created: func(data *GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.create",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.update",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.delete",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) GeneralAccountGroupingNetSurplusNegativeCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralAccountGroupingNetSurplusNegative, error) {
	return m.GeneralaccountgroupingnetsurplusnegativeManager().Find(context, &GeneralAccountGroupingNetSurplusNegative{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
