package model

import (
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	UserRating struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_user_rating"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_user_rating"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		RateeUserID uuid.UUID `gorm:"type:uuid;not null"`
		RateeUser   User      `gorm:"foreignKey:RateeUserID;constraint:OnDelete:SET NULL;" json:"ratee_user"`

		RaterUserID uuid.UUID `gorm:"type:uuid;not null"`
		RaterUser   User      `gorm:"foreignKey:RaterUserID;constraint:OnDelete:SET NULL;" json:"rater_user"`

		Rate   int    `gorm:"not null;check:rate >= 1 AND rate <= 5"`
		Remark string `gorm:"type:text"`
	}

	UserRatingRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		RateeUserID uuid.UUID `json:"ratee_user_id" validate:"required"`
		RaterUserID uuid.UUID `json:"rater_user_id" validate:"required"`
		Rate        int       `json:"rate" validate:"required,min=1,max=5"`
		Remark      string    `json:"remark" validate:"max=2000"`
	}

	UserRatingResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		DeletedByID    *uuid.UUID            `json:"deleted_by_id,omitempty"`
		DeletedBy      *UserResponse         `json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		RateeUserID uuid.UUID     `json:"ratee_user_id"`
		RateeUser   *UserResponse `json:"ratee_user,omitempty"`
		RaterUserID uuid.UUID     `json:"rater_user_id"`
		RaterUser   *UserResponse `json:"rater_user,omitempty"`
		Rate        int           `json:"rate"`
		Remark      string        `json:"remark"`
	}
)

func (m *Model) UserRating() {
	m.Migration = append(m.Migration, &UserRating{})
	m.UserRatingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[UserRating, UserRatingResponse, UserRatingRequest]{
		Preloads: []string{"Organization", "Branch", "RateeUser", "RaterUser"},
		Service:  m.provider.Service,
		Resource: func(data *UserRating) *UserRatingResponse {
			if data == nil {
				return nil
			}
			return &UserRatingResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				DeletedByID:    data.DeletedByID,
				DeletedBy:      m.UserManager.ToModel(data.DeletedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),

				RateeUserID: data.RateeUserID,
				RateeUser:   m.UserManager.ToModel(&data.RateeUser),
				RaterUserID: data.RaterUserID,
				RaterUser:   m.UserManager.ToModel(&data.RaterUser),
				Rate:        data.Rate,
				Remark:      data.Remark,
			}
		},
	})
}
