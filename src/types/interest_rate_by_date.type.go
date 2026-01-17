package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	InterestRateByDate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_date"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_date"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BrowseReferenceID uuid.UUID        `gorm:"type:uuid;not null;index:idx_browse_reference_date_range"`
		BrowseReference   *BrowseReference `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_reference,omitempty"`

		FromDate     time.Time `gorm:"not null;index:idx_browse_reference_date_range" json:"from_date" validate:"required"`
		ToDate       time.Time `gorm:"not null;index:idx_browse_reference_date_range" json:"to_date" validate:"required"`
		InterestRate float64   `gorm:"type:decimal(15,6);not null" json:"interest_rate" validate:"required,min=0"`
	}

	InterestRateByDateResponse struct {
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

		BrowseReferenceID uuid.UUID                `json:"browse_reference_id"`
		BrowseReference   *BrowseReferenceResponse `json:"browse_reference,omitempty"`
		FromDate          string                   `json:"from_date"`
		ToDate            string                   `json:"to_date"`
		InterestRate      float64                  `json:"interest_rate"`
	}

	InterestRateByDateRequest struct {
		ID                *uuid.UUID `json:"id"`
		BrowseReferenceID uuid.UUID  `json:"browse_reference_id" validate:"required"`
		FromDate          time.Time  `json:"from_date" validate:"required"`
		ToDate            time.Time  `json:"to_date" validate:"required,gtefield=FromDate"`
		InterestRate      float64    `json:"interest_rate" validate:"required,min=0"`
	}
)
