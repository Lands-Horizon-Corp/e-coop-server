
def main():
    txt = """
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
	MemberCloseRemark struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_close_remark"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_close_remark"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Reason      string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`
	}

	MemberCloseRemarkResponse struct {
		ID                     uuid.UUID                     `json:"id"`
		CreatedAt              string                        `json:"created_at"`
		CreatedByID            uuid.UUID                     `json:"created_by_id"`
		CreatedBy              *UserResponse                 `json:"created_by,omitempty"`
		UpdatedAt              string                        `json:"updated_at"`
		UpdatedByID            uuid.UUID                     `json:"updated_by_id"`
		UpdatedBy              *UserResponse                 `json:"updated_by,omitempty"`
		OrganizationID         uuid.UUID                     `json:"organization_id"`
		Organization           *OrganizationResponse         `json:"organization,omitempty"`
		BranchID               uuid.UUID                     `json:"branch_id"`
		Branch                 *BranchResponse               `json:"branch,omitempty"`
		MemberProfileID        uuid.UUID                     `json:"member_profile_id,omitempty"`
		MemberProfile          *MemberProfileResponse        `json:"member_profile,omitempty"`
		MemberClassificationID uuid.UUID                     `json:"member_classification_id,omitempty"`
		MemberClassification   *MemberClassificationResponse `json:"member_classification,omitempty"`
		Reason                 string                        `json:"reason,omitempty"`
		Description            string                        `json:"description,omitempty"`
	}

	MemberCloseRemarkRequest struct {
		Reason      string `json:"reason,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberCloseRemarkCollection struct {
		Manager horizon_manager.CollectionManager[MemberCloseRemark]
	}
)

func (m *Model) MemberCloseRemarkValidate(ctx echo.Context) (*MemberCloseRemarkRequest, error) {
	return horizon_manager.Validate[MemberCloseRemarkRequest](ctx, m.validator)
}

func (m *Model) MemberCloseRemarkModel(data *MemberCloseRemark) *MemberCloseRemarkResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberCloseRemark) *MemberCloseRemarkResponse {
		return &MemberCloseRemarkResponse{
			ID:              data.ID,
			CreatedAt:       data.CreatedAt.Format(time.RFC3339),
			CreatedByID:     data.CreatedByID,
			CreatedBy:       m.UserModel(data.CreatedBy),
			UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:     data.UpdatedByID,
			UpdatedBy:       m.UserModel(data.UpdatedBy),
			OrganizationID:  data.OrganizationID,
			Organization:    m.OrganizationModel(data.Organization),
			BranchID:        data.BranchID,
			Branch:          m.BranchModel(data.Branch),
			MemberProfileID: *data.MemberProfileID,
			MemberProfile:   m.MemberProfileModel(data.MemberProfile),

			Reason:      data.Reason,
			Description: data.Description,
		}
	})
}

func (m *Model) MemberCloseRemarkModels(data []*MemberCloseRemark) []*MemberCloseRemarkResponse {
	return horizon_manager.ToModels(data, m.MemberCloseRemarkModel)
}

func NewMemberCloseRemarkCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberCloseRemarkCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberCloseRemark) ([]string, any) {
			return []string{
				fmt.Sprintf("member_close_remark.create.%s", data.ID),
				fmt.Sprintf("member_close_remark.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.create.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_close_remark.create.organization.%s", data.OrganizationID),
			}, model.MemberCloseRemarkModel(data)
		},
		func(data *MemberCloseRemark) ([]string, any) {
			return []string{
				"member_close_remark.update",
				fmt.Sprintf("member_close_remark.update.%s", data.ID),
				fmt.Sprintf("member_close_remark.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.update.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_close_remark.update.organization.%s", data.OrganizationID),
			}, model.MemberCloseRemarkModel(data)
		},
		func(data *MemberCloseRemark) ([]string, any) {
			return []string{
				"member_close_remark.delete",
				fmt.Sprintf("member_close_remark.delete.%s", data.ID),
				fmt.Sprintf("member_close_remark.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.delete.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_close_remark.delete.organization.%s", data.OrganizationID),
			}, model.MemberCloseRemarkModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberCloseRemarkCollection{
		Manager: manager,
	}, nil
}

// member-close-remark/member_profile_id
func (fc *MemberCloseRemarkCollection) ListByMemberProfile(memberProfileId uuid.UUID) ([]*MemberCloseRemark, error) {
	return fc.Manager.Find(&MemberCloseRemark{
		MemberProfileID: &memberProfileId,
	})
}

// member-close-remark/branch/:branch_id
func (fc *MemberCloseRemarkCollection) ListByBranch(branchID uuid.UUID) ([]*MemberCloseRemark, error) {
	return fc.Manager.Find(&MemberCloseRemark{
		BranchID: branchID,
	})
}

// member-close-remark/organization/:organization_id
func (fc *MemberCloseRemarkCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberCloseRemark, error) {
	return fc.Manager.Find(&MemberCloseRemark{
		OrganizationID: organizationID,
	})
}

// member-close-remark/organization/:organization_id/branch/:branch_id
func (fc *MemberCloseRemarkCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberCloseRemark, error) {
	return fc.Manager.Find(&MemberCloseRemark{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

"""
    replacements = [
        ("member_close_remark", "member_asset"),
        ("member-close-remark", "member-asset"),
        ("memberCloseRemark", "memberAsset"),
        ("MemberCloseRemark", "MemberAsset"),
        ("MemberCloseRemark", "MemberAsset"),
    ]




    for (from_change, to_change) in replacements:
        txt = txt.replace(from_change, to_change)

    print(txt)


if __name__ == "__main__":
    main()
# clear; uv run documentation/main.py