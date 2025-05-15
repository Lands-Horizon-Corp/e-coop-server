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
	MemberAddress struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_address"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_address"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Label         string `gorm:"type:varchar(255);not null;default:''"`
		City          string `gorm:"type:varchar(255);not null;default:''"`
		CountryCode   string `gorm:"type:varchar(255);not null;default:''"`
		PostalCode    string `gorm:"type:varchar(255);not null;default:''"`
		ProvinceState string `gorm:"type:varchar(255);not null;default:''"`
		Parangay      string `gorm:"type:varchar(255);not null;default:''"`
		Landmark      string `gorm:"type:varchar(255);not null;default:''"`
		Address       string `gorm:"type:varchar(255);not null;default:''"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`
	}

	MemberAddressResponse struct {
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

		Label         string `json:"label,omitempty"`
		City          string `json:"city,omitempty"`
		CountryCode   string `json:"country_code,omitempty"`
		PostalCode    string `json:"postal_code,omitempty"`
		ProvinceState string `json:"province_state,omitempty"`
		Parangay      string `json:"baranggy,omitempty"`
		Landmark      string `json:"landmark,omitempty"`
		Address       string `json:"address,omitempty"`
	}

	MemberAddressRequest struct {
		Label         string `json:"label,omitempty" validate:"max=100"`
		City          string `json:"city,omitempty" validate:"max=100"`
		CountryCode   string `json:"country_code,omitempty" validate:"max=5,uppercase"`
		PostalCode    string `json:"postal_code,omitempty" validate:"max=20"`
		ProvinceState string `json:"province_state,omitempty" validate:"max=100"`
		Parangay      string `json:"barangay,omitempty" validate:"max=100"`
		Landmark      string `json:"landmark,omitempty" validate:"max=255"`
		Address       string `json:"address,omitempty" validate:"required,max=255"`
	}

	MemberAddressCollection struct {
		Manager horizon_manager.CollectionManager[MemberAddress]
	}
)

func (m *Model) MemberAddressValidate(ctx echo.Context) (*MemberAddressRequest, error) {
	return horizon_manager.Validate[MemberAddressRequest](ctx, m.validator)
}

func (m *Model) MemberAddressModel(data *MemberAddress) *MemberAddressResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberAddress) *MemberAddressResponse {
		return &MemberAddressResponse{
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

			Label:         data.Label,
			City:          data.City,
			CountryCode:   data.CountryCode,
			PostalCode:    data.PostalCode,
			ProvinceState: data.ProvinceState,
			Parangay:      data.Parangay,
			Landmark:      data.Landmark,
			Address:       data.Address,
		}
	})
}

func (m *Model) MemberAddressModels(data []*MemberAddress) []*MemberAddressResponse {
	return horizon_manager.ToModels(data, m.MemberAddressModel)
}

func NewMemberAddressCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberAddressCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberAddress) ([]string, any) {
			return []string{
				fmt.Sprintf("member_address.create.%s", data.ID),
				fmt.Sprintf("member_address.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_address.create.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_address.create.organization.%s", data.OrganizationID),
			}, model.MemberAddressModel(data)
		},
		func(data *MemberAddress) ([]string, any) {
			return []string{
				"member_address.update",
				fmt.Sprintf("member_address.update.%s", data.ID),
				fmt.Sprintf("member_address.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_address.update.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_address.update.organization.%s", data.OrganizationID),
			}, model.MemberAddressModel(data)
		},
		func(data *MemberAddress) ([]string, any) {
			return []string{
				"member_address.delete",
				fmt.Sprintf("member_address.delete.%s", data.ID),
				fmt.Sprintf("member_address.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_address.delete.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_address.delete.organization.%s", data.OrganizationID),
			}, model.MemberAddressModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberAddressCollection{
		Manager: manager,
	}, nil
}

// member-address/member_profile_id
func (fc *MemberAddressCollection) ListByMemberProfile(memberProfileId uuid.UUID) ([]*MemberAddress, error) {
	return fc.Manager.Find(&MemberAddress{
		MemberProfileID: &memberProfileId,
	})
}

// member-address/branch/:branch_id
func (fc *MemberAddressCollection) ListByBranch(branchID uuid.UUID) ([]*MemberAddress, error) {
	return fc.Manager.Find(&MemberAddress{
		BranchID: branchID,
	})
}

// member-address/organization/:organization_id
func (fc *MemberAddressCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberAddress, error) {
	return fc.Manager.Find(&MemberAddress{
		OrganizationID: organizationID,
	})
}

// member-address/organization/:organization_id/branch/:branch_id
func (fc *MemberAddressCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberAddress, error) {
	return fc.Manager.Find(&MemberAddress{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
