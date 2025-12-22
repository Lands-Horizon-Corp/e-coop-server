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
	TimeDepositComputation struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TimeDepositTypeID uuid.UUID        `gorm:"type:uuid;not null"`
		TimeDepositType   *TimeDepositType `gorm:"foreignKey:TimeDepositTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"time_deposit_type,omitempty"`

		MinimumAmount float64 `gorm:"type:decimal;default:0"`
		MaximumAmount float64 `gorm:"type:decimal;default:0"`

		Header1  float64 `gorm:"type:decimal;default:0"`
		Header2  float64 `gorm:"type:decimal;default:0"`
		Header3  float64 `gorm:"type:decimal;default:0"`
		Header4  float64 `gorm:"type:decimal;default:0"`
		Header5  float64 `gorm:"type:decimal;default:0"`
		Header6  float64 `gorm:"type:decimal;default:0"`
		Header7  float64 `gorm:"type:decimal;default:0"`
		Header8  float64 `gorm:"type:decimal;default:0"`
		Header9  float64 `gorm:"type:decimal;default:0"`
		Header10 float64 `gorm:"type:decimal;default:0"`
		Header11 float64 `gorm:"type:decimal;default:0"`
	}

	TimeDepositComputationResponse struct {
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
		MinimumAmount     float64                  `json:"minimum_amount"`
		MaximumAmount     float64                  `json:"maximum_amount"`
		Header1           float64                  `json:"header_1"`
		Header2           float64                  `json:"header_2"`
		Header3           float64                  `json:"header_3"`
		Header4           float64                  `json:"header_4"`
		Header5           float64                  `json:"header_5"`
		Header6           float64                  `json:"header_6"`
		Header7           float64                  `json:"header_7"`
		Header8           float64                  `json:"header_8"`
		Header9           float64                  `json:"header_9"`
		Header10          float64                  `json:"header_10"`
		Header11          float64                  `json:"header_11"`
	}

	TimeDepositComputationRequest struct {
		ID            *uuid.UUID `json:"id,omitempty"`
		MinimumAmount float64    `json:"minimum_amount,omitempty"`
		MaximumAmount float64    `json:"maximum_amount,omitempty"`
		Header1       float64    `json:"header_1,omitempty"`
		Header2       float64    `json:"header_2,omitempty"`
		Header3       float64    `json:"header_3,omitempty"`
		Header4       float64    `json:"header_4,omitempty"`
		Header5       float64    `json:"header_5,omitempty"`
		Header6       float64    `json:"header_6,omitempty"`
		Header7       float64    `json:"header_7,omitempty"`
		Header8       float64    `json:"header_8,omitempty"`
		Header9       float64    `json:"header_9,omitempty"`
		Header10      float64    `json:"header_10,omitempty"`
		Header11      float64    `json:"header_11,omitempty"`
	}
)

func (m *Core) timeDepositComputation() {
	m.Migration = append(m.Migration, &TimeDepositComputation{})
	m.TimeDepositComputationManager() = registry.NewRegistry(registry.RegistryParams[
		TimeDepositComputation, TimeDepositComputationResponse, TimeDepositComputationRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "TimeDepositType",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *TimeDepositComputation) *TimeDepositComputationResponse {
			if data == nil {
				return nil
			}
			return &TimeDepositComputationResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager().ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager().ToModel(data.Branch),
				TimeDepositTypeID: data.TimeDepositTypeID,
				TimeDepositType:   m.TimeDepositTypeManager().ToModel(data.TimeDepositType),
				MinimumAmount:     data.MinimumAmount,
				MaximumAmount:     data.MaximumAmount,
				Header1:           data.Header1,
				Header2:           data.Header2,
				Header3:           data.Header3,
				Header4:           data.Header4,
				Header5:           data.Header5,
				Header6:           data.Header6,
				Header7:           data.Header7,
				Header8:           data.Header8,
				Header9:           data.Header9,
				Header10:          data.Header10,
				Header11:          data.Header11,
			}
		},

		Created: func(data *TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.create",
				fmt.Sprintf("time_deposit_computation.create.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.create.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.update",
				fmt.Sprintf("time_deposit_computation.update.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.update.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *TimeDepositComputation) registry.Topics {
			return []string{
				"time_deposit_computation.delete",
				fmt.Sprintf("time_deposit_computation.delete.%s", data.ID),
				fmt.Sprintf("time_deposit_computation.delete.branch.%s", data.BranchID),
				fmt.Sprintf("time_deposit_computation.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) TimeDepositComputationCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*TimeDepositComputation, error) {
	return m.TimeDepositComputationManager().Find(context, &TimeDepositComputation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
