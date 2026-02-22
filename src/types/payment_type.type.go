package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PaymentTypeCash       TypeOfPaymentType = "cash"
	PaymentTypeCheck      TypeOfPaymentType = "check"
	PaymentTypeOnline     TypeOfPaymentType = "online"
	PaymentTypeAdjustment TypeOfPaymentType = "adjustment"
	PaymentTypeSystem     TypeOfPaymentType = "system"
)

type (
	TypeOfPaymentType string
	PaymentType       struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_payment_type" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_peyment_type" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name         string            `gorm:"type:varchar(255);not null"`
		Description  string            `gorm:"type:text"`
		NumberOfDays int               `gorm:"type:int"`
		Type         TypeOfPaymentType `gorm:"type:varchar(20)"`

		AccountID uuid.UUID `gorm:"type:uuid;not null" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
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

		AccountID uuid.UUID        `json:"account_id"`
		Account   *AccountResponse `json:"account,omitempty"`
	}

	PaymentTypeRequest struct {
		Name         string            `json:"name" validate:"required,min=1,max=255"`
		AccountID    uuid.UUID         `json:"account_id" validate:"required"`
		Description  string            `json:"description,omitempty"`
		NumberOfDays int               `json:"number_of_days,omitempty"`
		Type         TypeOfPaymentType `json:"type" validate:"required,oneof=cash check online adjustment"`
	}
)
