package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	TimeDepositComputationHeader struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_header"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_header"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Header1  int `gorm:"default:30"`
		Header2  int `gorm:"default:60"`
		Header3  int `gorm:"default:90"`
		Header4  int `gorm:"default:120"`
		Header5  int `gorm:"default:150"`
		Header6  int `gorm:"default:180"`
		Header7  int `gorm:"default:210"`
		Header8  int `gorm:"default:240"`
		Header9  int `gorm:"default:300"`
		Header10 int `gorm:"default:330"`
		Header11 int `gorm:"default:360"`
	}

	TimeDepositComputationHeaderResponse struct {
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
		Header1        int                   `json:"header_1"`
		Header2        int                   `json:"header_2"`
		Header3        int                   `json:"header_3"`
		Header4        int                   `json:"header_4"`
		Header5        int                   `json:"header_5"`
		Header6        int                   `json:"header_6"`
		Header7        int                   `json:"header_7"`
		Header8        int                   `json:"header_8"`
		Header9        int                   `json:"header_9"`
		Header10       int                   `json:"header_10"`
		Header11       int                   `json:"header_11"`
	}

	TimeDepositComputationHeaderRequest struct {
		Header1  int `json:"header_1,omitempty"`
		Header2  int `json:"header_2,omitempty"`
		Header3  int `json:"header_3,omitempty"`
		Header4  int `json:"header_4,omitempty"`
		Header5  int `json:"header_5,omitempty"`
		Header6  int `json:"header_6,omitempty"`
		Header7  int `json:"header_7,omitempty"`
		Header8  int `json:"header_8,omitempty"`
		Header9  int `json:"header_9,omitempty"`
		Header10 int `json:"header_10,omitempty"`
		Header11 int `json:"header_11,omitempty"`
	}
)

func (m *Model) TimeDepositComputationHeader() {
	m.Migration = append(m.Migration, &TimeDepositComputationHeader{})
	m.TimeDepositComputationHeaderManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		TimeDepositComputationHeader, TimeDepositComputationHeaderResponse, TimeDepositComputationHeaderRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *TimeDepositComputationHeader) *TimeDepositComputationHeaderResponse {
			if data == nil {
				return nil
			}
			return &TimeDepositComputationHeaderResponse{
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
				Header1:        data.Header1,
				Header2:        data.Header2,
				Header3:        data.Header3,
				Header4:        data.Header4,
				Header5:        data.Header5,
				Header6:        data.Header6,
				Header7:        data.Header7,
				Header8:        data.Header8,
				Header9:        data.Header9,
				Header10:       data.Header10,
				Header11:       data.Header11,
			}
		},
		Created: func(data *TimeDepositComputationHeader) []string {
			return []string{
				"time_deposit_computation_header.create",
				fmt.Sprintf("time_deposit_computation_header.create.%s", data.ID),
			}
		},
		Updated: func(data *TimeDepositComputationHeader) []string {
			return []string{
				"time_deposit_computation_header.update",
				fmt.Sprintf("time_deposit_computation_header.update.%s", data.ID),
			}
		},
		Deleted: func(data *TimeDepositComputationHeader) []string {
			return []string{
				"time_deposit_computation_header.delete",
				fmt.Sprintf("time_deposit_computation_header.delete.%s", data.ID),
			}
		},
	})
}
