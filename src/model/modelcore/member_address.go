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
	MemberAddress struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid" json:"created_by,omitempty"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by_user,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid" json:"updated_by,omitempty"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by_user,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by,omitempty"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by_user,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_organization_branch_member_address" json:"organization_id"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_organization_branch_member_address" json:"branch_id"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Label         string   `gorm:"type:varchar(255);not null;default:home"`
		City          string   `gorm:"type:varchar(255);not null"`
		CountryCode   string   `gorm:"type:varchar(5);not null"`
		PostalCode    string   `gorm:"type:varchar(255)"`
		ProvinceState string   `gorm:"type:varchar(255)"`
		Barangay      string   `gorm:"type:varchar(255)"`
		Landmark      string   `gorm:"type:varchar(255)"`
		Address       string   `gorm:"type:varchar(255);not null"`
		Latitude      *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude     *float64 `gorm:"type:double precision" json:"longitude,omitempty"`
	}

	MemberAddressReponse struct {
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

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		Label         string   `json:"label"`
		City          string   `json:"city"`
		CountryCode   string   `json:"country_code"`
		PostalCode    string   `json:"postal_code"`
		ProvinceState string   `json:"province_state"`
		Barangay      string   `json:"barangay"`
		Landmark      string   `json:"landmark"`
		Address       string   `json:"address"`
		Longitude     *float64 `json:"longitude,omitempty"`
		Latitude      *float64 `json:"latitude,omitempty"`
	}

	MemberAddressRequest struct {
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`

		Label         string   `json:"label" validate:"required,min=1,max=255"`
		City          string   `json:"city" validate:"required,min=1,max=255"`
		CountryCode   string   `json:"country_code" validate:"required,min=1,max=5"`
		PostalCode    string   `json:"postal_code,omitempty" validate:"omitempty,max=255"`
		ProvinceState string   `json:"province_state,omitempty" validate:"omitempty,max=255"`
		Barangay      string   `json:"barangay,omitempty" validate:"omitempty,max=255"`
		Landmark      string   `json:"landmark,omitempty" validate:"omitempty,max=255"`
		Address       string   `json:"address" validate:"required,min=1,max=255"`
		Longitude     *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
		Latitude      *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	}
)

func (m *modelcore) MemberAddress() {
	m.Migration = append(m.Migration, &MemberAddress{})
	m.MemberAddressManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberAddress, MemberAddressReponse, MemberAddressRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Service:  m.provider.Service,
		Resource: func(data *MemberAddress) *MemberAddressReponse {
			if data == nil {
				return nil
			}
			return &MemberAddressReponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				Label:           data.Label,
				City:            data.City,
				CountryCode:     data.CountryCode,
				PostalCode:      data.PostalCode,
				ProvinceState:   data.ProvinceState,
				Barangay:        data.Barangay,
				Landmark:        data.Landmark,
				Address:         data.Address,
				Longitude:       data.Longitude,
				Latitude:        data.Latitude,
			}
		},

		Created: func(data *MemberAddress) []string {
			return []string{
				"member_address.create",
				fmt.Sprintf("member_address.create.%s", data.ID),
				fmt.Sprintf("member_address.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberAddress) []string {
			return []string{
				"member_address.update",
				fmt.Sprintf("member_address.update.%s", data.ID),
				fmt.Sprintf("member_address.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberAddress) []string {
			return []string{
				"member_address.delete",
				fmt.Sprintf("member_address.delete.%s", data.ID),
				fmt.Sprintf("member_address.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_address.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *modelcore) MemberAddressCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberAddress, error) {
	return m.MemberAddressManager.Find(context, &MemberAddress{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
