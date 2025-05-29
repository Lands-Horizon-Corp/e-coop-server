package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Enum for types_of_payment_type
type TypesOfPaymentType string

const (
	PaymentTypeCash   TypesOfPaymentType = "cash"
	PaymentTypeCheck  TypesOfPaymentType = "check"
	PaymentTypeOnline TypesOfPaymentType = "online"
)

type (
	PaymentType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_payment_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_payment_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name         string             `gorm:"type:varchar(255);not null"`
		Description  string             `gorm:"type:text"`
		NumberOfDays int                `gorm:"type:int"`
		Type         TypesOfPaymentType `gorm:"type:varchar(20)"`
	}

	PaymentTypeResponse struct {
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
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		NumberOfDays   int                   `json:"number_of_days"`
		Type           TypesOfPaymentType    `json:"type"`
	}

	PaymentTypeRequest struct {
		Name         string             `json:"name" validate:"required,min=1,max=255"`
		Description  string             `json:"description,omitempty"`
		NumberOfDays int                `json:"number_of_days,omitempty"`
		Type         TypesOfPaymentType `json:"type" validate:"required,oneof=cash check online"`
	}
)

func (m *Model) PaymentType() {
	m.Migration = append(m.Migration, &PaymentType{})
	m.PaymentTypeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		PaymentType, PaymentTypeResponse, PaymentTypeRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *PaymentType) *PaymentTypeResponse {
			if data == nil {
				return nil
			}
			return &PaymentTypeResponse{
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
				Name:           data.Name,
				Description:    data.Description,
				NumberOfDays:   data.NumberOfDays,
				Type:           data.Type,
			}
		},
		Created: func(data *PaymentType) []string {
			return []string{
				"payment_type.create",
				fmt.Sprintf("payment_type.create.%s", data.ID),
			}
		},
		Updated: func(data *PaymentType) []string {
			return []string{
				"payment_type.update",
				fmt.Sprintf("payment_type.update.%s", data.ID),
			}
		},
		Deleted: func(data *PaymentType) []string {
			return []string{
				"payment_type.delete",
				fmt.Sprintf("payment_type.delete.%s", data.ID),
			}
		},
	})
}
