package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// UserRating represents a rating given by one user to another user within an organization
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

	// UserRatingRequest represents the request structure for creating or updating a user rating
	UserRatingRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		RateeUserID uuid.UUID `json:"ratee_user_id" validate:"required"`
		RaterUserID uuid.UUID `json:"rater_user_id" validate:"required"`
		Rate        int       `json:"rate" validate:"required,min=1,max=5"`
		Remark      string    `json:"remark" validate:"max=2000"`
	}

	// UserRatingResponse represents the response structure for user rating data
	UserRatingResponse struct {
		ID          uuid.UUID     `json:"id"`
		CreatedAt   string        `json:"created_at"`
		CreatedByID uuid.UUID     `json:"created_by_id"`
		CreatedBy   *UserResponse `json:"created_by,omitempty"`
		UpdatedAt   string        `json:"updated_at"`
		UpdatedByID uuid.UUID     `json:"updated_by_id"`
		UpdatedBy   *UserResponse `json:"updated_by,omitempty"`

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

// UserRating initializes the user rating repository and sets up migration
func (m *ModelCore) UserRating() {
	m.Migration = append(m.Migration, &UserRating{})
	m.UserRatingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[UserRating, UserRatingResponse, UserRatingRequest]{
		Preloads: []string{"Organization", "Branch", "RateeUser", "RaterUser"},
		Service:  m.provider.Service,
		Resource: func(data *UserRating) *UserRatingResponse {
			if data == nil {
				return nil
			}
			return &UserRatingResponse{
				ID:          data.ID,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				CreatedByID: data.CreatedByID,
				CreatedBy:   m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID: data.UpdatedByID,
				UpdatedBy:   m.UserManager.ToModel(data.UpdatedBy),

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
		Created: func(data *UserRating) []string {
			return []string{
				"user_rating.create",
				fmt.Sprintf("user_rating.create.%s", data.ID),
				fmt.Sprintf("user_rating.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *UserRating) []string {
			return []string{
				"user_rating.update",
				fmt.Sprintf("user_rating.update.%s", data.ID),
				fmt.Sprintf("user_rating.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *UserRating) []string {
			return []string{
				"user_rating.delete",
				fmt.Sprintf("user_rating.delete.%s", data.ID),
				fmt.Sprintf("user_rating.delete.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GetUserRatee retrieves all ratings where the specified user is the ratee (being rated)
func (m *ModelCore) GetUserRatee(context context.Context, userID uuid.UUID) ([]*UserRating, error) {
	return m.UserRatingManager.Find(context, &UserRating{
		RateeUserID: userID,
	})
}

// GetUserRater retrieves all ratings where the specified user is the rater (giving ratings)
func (m *ModelCore) GetUserRater(context context.Context, userID uuid.UUID) ([]*UserRating, error) {
	return m.UserRatingManager.Find(context, &UserRating{
		RaterUserID: userID,
	})
}

// UserRatingCurrentBranch retrieves all user ratings for the specified organization and branch
func (m *ModelCore) UserRatingCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*UserRating, error) {
	return m.UserRatingManager.Find(context, &UserRating{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
