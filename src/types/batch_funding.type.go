package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BatchFunding struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_batch_funding" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_batch_funding" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TransactionBatchID uuid.UUID         `gorm:"type:uuid;not null" json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		ProvidedByUserID uuid.UUID `gorm:"type:uuid;not null" json:"provided_by_user_id"`
		ProvidedByUser   *User     `gorm:"foreignKey:ProvidedByUserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"provided_by_user,omitempty"`

		SignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"signature_media_id"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null" json:"currency_id"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name        string  `gorm:"type:varchar(50)" json:"name"`
		Amount      float64 `gorm:"type:decimal" json:"amount"`
		Description string  `gorm:"type:text" json:"description"`
	}

	BatchFundingResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		TransactionBatchID uuid.UUID                 `json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		ProvidedByUserID   uuid.UUID                 `json:"provided_by_user_id"`
		ProvidedByUser     *UserResponse             `json:"provided_by_user,omitempty"`
		SignatureMediaID   *uuid.UUID                `json:"signature_media_id,omitempty"`
		SignatureMedia     *MediaResponse            `json:"signature_media,omitempty"`
		CurrencyID         uuid.UUID                 `json:"currency_id"`
		Currency           *CurrencyResponse         `json:"currency,omitempty"`
		Name               string                    `json:"name"`
		Amount             float64                   `json:"amount"`
		Description        string                    `json:"description"`
	}

	BatchFundingRequest struct {
		ProvidedByUserID uuid.UUID  `json:"provided_by_user_id" validate:"required"`
		SignatureMediaID *uuid.UUID `json:"signature_media_id,omitempty"`
		CurrencyID       uuid.UUID  `json:"currency_id" validate:"required"`
		Name             string     `json:"name" validate:"required,min=1,max=50"`
		Amount           float64    `json:"amount,omitempty"`
		Description      string     `json:"description,omitempty"`
	}
)
