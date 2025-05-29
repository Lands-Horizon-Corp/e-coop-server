package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	TimeDepositType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TimeDepositComputationHeaderID uuid.UUID                     `gorm:"type:uuid"`
		TimeDepositComputationHeader   *TimeDepositComputationHeader `gorm:"foreignKey:TimeDepositComputationHeaderID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"time_deposit_computation_header,omitempty"`

		Name        string `gorm:"type:varchar(255);not null;unique"`
		Description string `gorm:"type:text"`

		PreMature     int     `gorm:"default:0"`
		PreMatureRate float64 `gorm:"type:decimal;default:0"`
		Excess        float64 `gorm:"type:decimal;default:0"`
	}

	TimeDepositTypeResponse struct {
		ID                             uuid.UUID                             `json:"id"`
		CreatedAt                      string                                `json:"created_at"`
		CreatedByID                    uuid.UUID                             `json:"created_by_id"`
		CreatedBy                      *UserResponse                         `json:"created_by,omitempty"`
		UpdatedAt                      string                                `json:"updated_at"`
		UpdatedByID                    uuid.UUID                             `json:"updated_by_id"`
		UpdatedBy                      *UserResponse                         `json:"updated_by,omitempty"`
		OrganizationID                 uuid.UUID                             `json:"organization_id"`
		Organization                   *OrganizationResponse                 `json:"organization,omitempty"`
		BranchID                       uuid.UUID                             `json:"branch_id"`
		Branch                         *BranchResponse                       `json:"branch,omitempty"`
		TimeDepositComputationHeaderID uuid.UUID                             `json:"time_deposit_computation_header_id"`
		TimeDepositComputationHeader   *TimeDepositComputationHeaderResponse `json:"time_deposit_computation_header,omitempty"`
		Name                           string                                `json:"name"`
		Description                    string                                `json:"description"`
		PreMature                      int                                   `json:"pre_mature"`
		PreMatureRate                  float64                               `json:"pre_mature_rate"`
		Excess                         float64                               `json:"excess"`
	}

	TimeDepositTypeRequest struct {
		TimeDepositComputationHeaderID uuid.UUID `json:"time_deposit_computation_header_id,omitempty"`
		Name                           string    `json:"name" validate:"required,min=1,max=255"`
		Description                    string    `json:"description,omitempty"`
		PreMature                      int       `json:"pre_mature,omitempty"`
		PreMatureRate                  float64   `json:"pre_mature_rate,omitempty"`
		Excess                         float64   `json:"excess,omitempty"`
	}
)

func (m *Model) TimeDepositType() {
	m.Migration = append(m.Migration, &TimeDepositType{})
	m.TimeDepositTypeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		TimeDepositType, TimeDepositTypeResponse, TimeDepositTypeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "TimeDepositComputationHeader",
		},
		Service: m.provider.Service,
		Resource: func(data *TimeDepositType) *TimeDepositTypeResponse {
			if data == nil {
				return nil
			}
			return &TimeDepositTypeResponse{
				ID:                             data.ID,
				CreatedAt:                      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                    data.CreatedByID,
				CreatedBy:                      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                    data.UpdatedByID,
				UpdatedBy:                      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                 data.OrganizationID,
				Organization:                   m.OrganizationManager.ToModel(data.Organization),
				BranchID:                       data.BranchID,
				Branch:                         m.BranchManager.ToModel(data.Branch),
				TimeDepositComputationHeaderID: data.TimeDepositComputationHeaderID,
				TimeDepositComputationHeader:   m.TimeDepositComputationHeaderManager.ToModel(data.TimeDepositComputationHeader),
				Name:                           data.Name,
				Description:                    data.Description,
				PreMature:                      data.PreMature,
				PreMatureRate:                  data.PreMatureRate,
				Excess:                         data.Excess,
			}
		},
		Created: func(data *TimeDepositType) []string {
			return []string{
				"time_deposit_type.create",
				fmt.Sprintf("time_deposit_type.create.%s", data.ID),
			}
		},
		Updated: func(data *TimeDepositType) []string {
			return []string{
				"time_deposit_type.update",
				fmt.Sprintf("time_deposit_type.update.%s", data.ID),
			}
		},
		Deleted: func(data *TimeDepositType) []string {
			return []string{
				"time_deposit_type.delete",
				fmt.Sprintf("time_deposit_type.delete.%s", data.ID),
			}
		},
	})
}
