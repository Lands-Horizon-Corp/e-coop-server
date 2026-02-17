package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Feed struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid;not null" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_feed" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_feed" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Description  string         `gorm:"type:varchar(255);not null" json:"description"`
		FeedMedias   []*FeedMedia   `gorm:"foreignKey:FeedID;constraint:OnDelete:CASCADE;" json:"feed_medias,omitempty"`
		FeedComments []*FeedComment `gorm:"foreignKey:FeedID;constraint:OnDelete:CASCADE;" json:"feed_comments,omitempty"`
		UserLikes    []*FeedLike    `gorm:"foreignKey:FeedID;constraint:OnDelete:CASCADE;" json:"user_likes,omitempty"`
	}

	FeedResponse struct {
		ID             uuid.UUID              `json:"id"`
		CreatedAt      string                 `json:"created_at"`
		CreatedByID    uuid.UUID              `json:"created_by_id"`
		CreatedBy      *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt      string                 `json:"updated_at"`
		UpdatedByID    uuid.UUID              `json:"updated_by_id"`
		UpdatedBy      *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID              `json:"organization_id"`
		Organization   *OrganizationResponse  `json:"organization,omitempty"`
		BranchID       uuid.UUID              `json:"branch_id"`
		Branch         *BranchResponse        `json:"branch,omitempty"`
		Description    string                 `json:"description"`
		FeedMedias     []*FeedMediaResponse   `json:"feed_medias,omitempty"`
		FeedComments   []*FeedCommentResponse `json:"feed_comments,omitempty"`
		UserLikes      []*FeedLikeResponse    `json:"user_likes,omitempty"`
		IsLiked        bool                   `json:"is_liked"`
	}

	FeedRequest struct {
		Description string      `json:"description" validate:"required,min=1,max=255"`
		MediaIDs    []uuid.UUID `json:"media_ids"`
	}
)
