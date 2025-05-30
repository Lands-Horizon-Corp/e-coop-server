package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberAddress struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_collectors_member_address"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_collectors_member_address"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID *uuid.UUID `gorm:"type:uuid"`
		User   *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`

		Label         string `gorm:"type:varchar(255);not null;default:home"`
		City          string `gorm:"type:varchar(255);not null"`
		CountryCode   string `gorm:"type:varchar(5);not null"`
		PostalCode    string `gorm:"type:varchar(255)"`
		ProvinceState string `gorm:"type:varchar(255)"`
		Barangay      string `gorm:"type:varchar(255)"`
		Landmark      string `gorm:"type:varchar(255)"`
		Address       string `gorm:"type:varchar(255);not null"`
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

		UserID *uuid.UUID    `json:"user_id,omitempty"`
		User   *UserResponse `json:"user,omitempty"`

		Label         string `json:"label"`
		City          string `json:"city"`
		CountryCode   string `json:"country_code"`
		PostalCode    string `json:"postal_code"`
		ProvinceState string `json:"province_state"`
		Barangay      string `json:"barangay"`
		Landmark      string `json:"landmark"`
		Address       string `json:"address"`
	}

	MemberAddressRequest struct {
		Label         string `json:"label" validate:"required,min=1,max=255"`
		City          string `json:"city" validate:"required,min=1,max=255"`
		CountryCode   string `json:"country_code" validate:"required,min=1,max=5"`
		PostalCode    string `json:"postal_code,omitempty" validate:"omitempty,max=255"`
		ProvinceState string `json:"province_state,omitempty" validate:"omitempty,max=255"`
		Barangay      string `json:"barangay,omitempty" validate:"omitempty,max=255"`
		Landmark      string `json:"landmark,omitempty" validate:"omitempty,max=255"`
		Address       string `json:"address" validate:"required,min=1,max=255"`
	}
)

func (m *Model) MemberAddress() {
	m.Migration = append(m.Migration, &MemberAddress{})
	m.MemberAddressManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberAddress, MemberAddressReponse, MemberAddressRequest]{
		Preloads: []string{"User", "Branch", "Organization", "CreatedBy", "UpdatedBy"},
		Service:  m.provider.Service,
		Resource: func(data *MemberAddress) *MemberAddressReponse {
			if data == nil {
				return nil
			}
			return &MemberAddressReponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				UserID:         data.UserID,
				User:           m.UserManager.ToModel(data.User),
				Label:          data.Label,
				City:           data.City,
				CountryCode:    data.CountryCode,
				PostalCode:     data.PostalCode,
				ProvinceState:  data.ProvinceState,
				Barangay:       data.Barangay,
				Landmark:       data.Landmark,
				Address:        data.Address,
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
