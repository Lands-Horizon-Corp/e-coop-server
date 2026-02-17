package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FeedLike struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_feed_like" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_feed_like" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		FeedID uuid.UUID `gorm:"type:uuid;not null;index:idx_feed_like_unique,unique" json:"feed_id"`
		Feed   *Feed     `gorm:"foreignKey:FeedID;constraint:OnDelete:CASCADE;" json:"feed,omitempty"`
		UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_feed_like_unique,unique" json:"user_id"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	}

	FeedLikeResponse struct {
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
		FeedID         uuid.UUID             `json:"feed_id"`
		UserID         uuid.UUID             `json:"user_id"`
		User           *UserResponse         `json:"user,omitempty"`
	}

	FeedLikeRequest struct {
		FeedID uuid.UUID `json:"feed_id" validate:"required"`
	}
)
