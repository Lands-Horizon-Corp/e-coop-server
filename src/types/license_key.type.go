package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type License struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	// Audit
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
	CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
	UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// License Core
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	LicenseKey  string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"license_key"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`

	// Usage
	IsUsed bool       `gorm:"default:false" json:"is_used"`
	UsedAt *time.Time `json:"used_at,omitempty"`

	// Status
	IsRevoked bool       `gorm:"default:false" json:"is_revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type LicenseRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type LicenseResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LicenseKey  string    `json:"license_key"`
	ExpiresAt   *string   `json:"expires_at,omitempty"`

	IsUsed    bool    `json:"is_used"`
	UsedAt    *string `json:"used_at,omitempty"`
	IsRevoked bool    `json:"is_revoked"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type LicenseActivateRequest struct {
	LicenseKey  string `json:"license_key" validate:"required,len=127"`
	Fingerprint string `json:"fingerprint" validate:"required,min=10"`
}

type LicenseActivateResponse struct {
	SecretKey string `json:"secret_key"`
}

type LicenseVerifyRequest struct {
	SecretKey   string `json:"secret_key" validate:"required"`
	Fingerprint string `json:"fingerprint" validate:"required,min=10"`
}

type LicenseDeactivateRequest struct {
	SecretKey   string `json:"secret_key" validate:"required"`
	Fingerprint string `json:"fingerprint" validate:"required,min=10"`
}
