package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TypeOfPaymentType string

const (
	PaymentTypeCash       TypeOfPaymentType = "cash"
	PaymentTypeCheck      TypeOfPaymentType = "check"
	PaymentTypeOnline     TypeOfPaymentType = "online"
	PaymentTypeAdjustment TypeOfPaymentType = "adjustment"
	PaymentTypeSystem     TypeOfPaymentType = "system"
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

		Name         string            `gorm:"type:varchar(255);not null"`
		Description  string            `gorm:"type:text"`
		NumberOfDays int               `gorm:"type:int"`
		Type         TypeOfPaymentType `gorm:"type:varchar(20)"`
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
		Type           TypeOfPaymentType     `json:"type"`
	}

	PaymentTypeRequest struct {
		Name         string            `json:"name" validate:"required,min=1,max=255"`
		Description  string            `json:"description,omitempty"`
		NumberOfDays int               `json:"number_of_days,omitempty"`
		Type         TypeOfPaymentType `json:"type" validate:"required,oneof=cash check online adjustment"`
	}
)

func (m *Core) paymentType() {
	m.Migration = append(m.Migration, &PaymentType{})
	m.PaymentTypeManager = registry.NewRegistry(registry.RegistryParams[
		PaymentType, PaymentTypeResponse, PaymentTypeRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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
		Created: func(data *PaymentType) registry.Topics {
			return []string{
				"payment_type.create",
				fmt.Sprintf("payment_type.create.%s", data.ID),
				fmt.Sprintf("payment_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *PaymentType) registry.Topics {
			return []string{
				"payment_type.update",
				fmt.Sprintf("payment_type.update.%s", data.ID),
				fmt.Sprintf("payment_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *PaymentType) registry.Topics {
			return []string{
				"payment_type.delete",
				fmt.Sprintf("payment_type.delete.%s", data.ID),
				fmt.Sprintf("payment_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) PaymentTypeCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*PaymentType, error) {
	return m.PaymentTypeManager.Find(context, &PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
