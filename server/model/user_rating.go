package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	UserRating struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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

	UserRatingCollection struct {
		Manager horizon_manager.CollectionManager[UserRating]
	}
)

func (m *Model) UserRatingValidate(ctx echo.Context) (*UserRatingRequest, error) {
	return horizon_manager.Validate[UserRatingRequest](ctx, m.validator)
}

func (m *Model) UserRatingModel(data *UserRating) *UserRatingResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *UserRating) *UserRatingResponse {
		return &UserRatingResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.UpdatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),

			RateeUserID: data.RateeUserID,
			RateeUser:   m.UserModel(&data.RateeUser),
			RaterUserID: data.RaterUserID,
			RaterUser:   m.UserModel(&data.RaterUser),
			Rate:        data.Rate,
			Remark:      data.Remark,
		}
	})
}

func (m *Model) UserRatingModels(data []*UserRating) []*UserRatingResponse {
	return horizon_manager.ToModels(data, m.UserRatingModel)
}

func NewUserRatingCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*UserRatingCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *UserRating) ([]string, any) {
			return []string{
				fmt.Sprintf("user_rating.create.%s", data.ID),
				fmt.Sprintf("user_rating.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.create.organization.%s", data.OrganizationID),
			}, model.UserRatingModel(data)
		},
		func(data *UserRating) ([]string, any) {
			return []string{
				"user_rating.update",
				fmt.Sprintf("user_rating.update.%s", data.ID),
				fmt.Sprintf("user_rating.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.update.organization.%s", data.OrganizationID),
			}, model.UserRatingModel(data)
		},
		func(data *UserRating) ([]string, any) {
			return []string{
				"user_rating.delete",
				fmt.Sprintf("user_rating.delete.%s", data.ID),
				fmt.Sprintf("user_rating.delete.branch.%s", data.BranchID),
				fmt.Sprintf("user_rating.delete.organization.%s", data.OrganizationID),
			}, model.UserRatingModel(data)
		},
		[]string{
			"RateeUser",
			"RaterUser",
		},
	)

	return &UserRatingCollection{
		Manager: manager,
	}, nil
}

// user-rating/user-ratee/:user_ratee_id
func (urc *UserRatingCollection) ListByUserRatee(rateeId uuid.UUID) ([]*UserRating, error) {
	return urc.Manager.Find(&UserRating{
		RateeUserID: rateeId,
	})
}

// user-rating/user-rater/user_rater_id
func (urc *UserRatingCollection) ListByUserRater(raterId uuid.UUID) ([]*UserRating, error) {
	return urc.Manager.Find(&UserRating{
		RaterUserID: raterId,
	})
}

// user_setting/branch/:branch_id
func (fc *UserRatingCollection) ListByBranch(branchID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		BranchID: branchID,
	})
}

// user_setting/organization/:organization_id
func (fc *UserRatingCollection) ListByOrganization(organizationID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: organizationID,
	})
}

// user_setting/organization/:organization_id/branch/:branch_id
func (fc *UserRatingCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// GET /user-rating/branch/:branch_id/ratee/:ratee_user_id
func (fc *UserRatingCollection) ListByBranchRatee(branchID, rateeID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		BranchID:    branchID,
		RateeUserID: rateeID,
	})
}

// GET /user-rating/branch/:branch_id/rater/:rater_user_id
func (fc *UserRatingCollection) ListByBranchRater(branchID, raterID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		BranchID:    branchID,
		RaterUserID: raterID,
	})
}

// GET /user-rating/organization/:organization_id/ratee/:ratee_user_id
func (fc *UserRatingCollection) ListByOrganizationRatee(organizationID, rateeID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: organizationID,
		RateeUserID:    rateeID,
	})
}

// GET /user-rating/organization/:organization_id/rater/:rater_user_id
func (fc *UserRatingCollection) ListByOrganizationRater(organizationID, raterID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: organizationID,
		RaterUserID:    raterID,
	})
}

// GET /user-rating/organization/:organization_id/branch/:branch_id/ratee/:ratee_user_id
func (fc *UserRatingCollection) ListByOrgBranchRatee(orgID, branchID, rateeID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: orgID,
		BranchID:       branchID,
		RateeUserID:    rateeID,
	})
}

// GET /user-rating/organization/:organization_id/branch/:branch_id/rater/:rater_user_id
func (fc *UserRatingCollection) ListByOrgBranchRater(orgID, branchID, raterID uuid.UUID) ([]*UserRating, error) {
	return fc.Manager.Find(&UserRating{
		OrganizationID: orgID,
		BranchID:       branchID,
		RaterUserID:    raterID,
	})
}
