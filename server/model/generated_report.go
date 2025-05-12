package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	GeneratedReport struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_generated_report"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID      *uuid.UUID `gorm:"type:uuid"`
		User        *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID     *uuid.UUID `gorm:"type:uuid;not null"`
		Media       *Media     `gorm:"foreignKey:MediaID"`
		Name        string     `gorm:"type:varchar(255);not null"`
		Description string     `gorm:"type:text;not null"`
		Status      string     `gorm:"type:varchar(50);not null"`
		Progress    float64    `gorm:"not null"`
	}
	GeneratedReportResponse struct {
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

		UserID      *uuid.UUID     `json:"user_id"`
		User        *UserResponse  `json:"user"`
		MediaID     *uuid.UUID     `json:"media_id"`
		Media       *MediaResponse `json:"media,omitempty"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Status      string         `json:"status"`
		Progress    float64        `json:"progress"`
	}

	GeneratedReportRequest struct {
		Name        string `json:"firstName" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1"`
	}
	GeneratedReportCollection struct {
		Manager CollectionManager[GeneratedReport]
	}
)

func (m *Model) GeneratedReportModel(data *GeneratedReport) *GeneratedReportResponse {
	return ToModel(data, func(cu *GeneratedReport) *GeneratedReportResponse {
		return &GeneratedReportResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.UpdatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),
			UserID:         data.UserID,
			User:           m.UserModel(data.User),
			MediaID:        data.MediaID,
			Media:          m.MediaModel(data.Media),
			Name:           data.Name,
			Description:    data.Description,
			Status:         data.Status,
			Progress:       data.Progress,
		}
	})
}

func (m *Model) GeneratedReportModels(data []*GeneratedReport) []*GeneratedReportResponse {
	return ToModels(data, m.GeneratedReportModel)
}

func NewGeneratedReportCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*GeneratedReportCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *GeneratedReport) ([]string, any) {
			return []string{
				"generated_report.create",
				fmt.Sprintf("generated_report.create.%s", data.ID),
				fmt.Sprintf("generated_report.create.user.%s", data.UserID),
			}, model.GeneratedReportModel(data)
		},
		func(data *GeneratedReport) ([]string, any) {
			return []string{
				"generated_report.update",
				fmt.Sprintf("generated_report.update.%s", data.ID),
				fmt.Sprintf("generated_report.update.user.%s", data.UserID),
			}, model.GeneratedReportModel(data)
		},
		func(data *GeneratedReport) ([]string, any) {
			return []string{
				"generated_report.delete",
				fmt.Sprintf("generated_report.delete.%s", data.ID),
				fmt.Sprintf("generated_report.delete.user.%s", data.UserID),
			}, model.GeneratedReportModel(data)
		},
	)
	return &GeneratedReportCollection{
		Manager: manager,
	}, nil
}
