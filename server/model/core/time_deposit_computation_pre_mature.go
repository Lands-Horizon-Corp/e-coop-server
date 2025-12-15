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
	TimeDepositComputationPreMature struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_pre_mature"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_pre_mature"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TimeDepositTypeID uuid.UUID        `gorm:"type:uuid;not null"`
		TimeDepositType   *TimeDepositType `gorm:"foreignKey:TimeDepositTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"time_deposit_type,omitempty"`

		Terms int     `gorm:"default:0"`
		From  int     `gorm:"default:0"`
		To    int     `gorm:"default:0"`
		Rate  float64 `gorm:"type:decimal;default:0"`
	}


	TimeDepositComputationPreMatureResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		TimeDepositTypeID uuid.UUID                `json:"time_deposit_type_id"`
		TimeDepositType   *TimeDepositTypeResponse `json:"time_deposit_type,omitempty"`
		Terms             int                      `json:"terms"`
		From              int                      `json:"from"`
		To                int                      `json:"to"`
		Rate              float64                  `json:"rate"`
	}


	TimeDepositComputationPreMatureRequest struct {
		ID                *uuid.UUID `json:"id,omitempty"`
		TimeDepositTypeID uuid.UUID  `json:"time_deposit_type_id" validate:"required"`
		Terms             int        `json:"terms,omitempty"`
		From              int        `json:"from,omitempty"`
		To                int        `json:"to,omitempty"`
		Rate              float64    `json:"rate,omitempty"`
	}
)

func (m *Core) timeDepositComputationPreMature() {
	m.Migration = append(m.Migration, &TimeDepositComputationPreMature{})
	m.TimeDepositComputationPreMatureManager = *registry.NewRegistry(registry.RegistryParams[
		TimeDepositComputationPreMature, TimeDepositComputationPreMatureResponse, TimeDepositComputationPreMatureRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "TimeDepositType",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *TimeDepositComputationPreMature) *TimeDepositComputationPreMatureResponse {
			if data == nil {
				return nil
			}
			return &TimeDepositComputationPreMatureResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager.ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager.ToModel(data.Branch),
				TimeDepositTypeID: data.TimeDepositTypeID,
				TimeDepositType:   m.TimeDepositTypeManager.ToModel(data.TimeDepositType),
				Terms:             data.Terms,
				From:              data.From,
				To:                data.To,
				Rate:              data.Rate,
			}
		},

		Created: func(data *TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.create",
				fmt.Sprintf("time_deposit_computation_pre_mature.create.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.create.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.update",
				fmt.Sprintf("time_deposit_computation_pre_mature.update.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.update.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *TimeDepositComputationPreMature) registry.Topics {
			return []string{
				"time_deposit_computation_pre_mature.delete",
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.%s", data.ID),
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation_pre_mature.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) TimeDepositComputationPreMatureCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*TimeDepositComputationPreMature, error) {
	return m.TimeDepositComputationPreMatureManager.Find(context, &TimeDepositComputationPreMature{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
