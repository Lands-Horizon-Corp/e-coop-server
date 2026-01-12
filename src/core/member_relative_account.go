package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberRelativeAccount struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_relative_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_relative_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		RelativeMemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		RelativeMemberProfile   *MemberProfile `gorm:"foreignKey:RelativeMemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"relative_member_profile,omitempty"`

		FamilyRelationship string `gorm:"type:varchar(255);not null"` // Enum handled in frontend/validation
		Description        string `gorm:"type:text"`
	}

	MemberRelativeAccountResponse struct {
		ID                      uuid.UUID              `json:"id"`
		CreatedAt               string                 `json:"created_at"`
		CreatedByID             uuid.UUID              `json:"created_by_id"`
		CreatedBy               *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt               string                 `json:"updated_at"`
		UpdatedByID             uuid.UUID              `json:"updated_by_id"`
		UpdatedBy               *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID          uuid.UUID              `json:"organization_id"`
		Organization            *OrganizationResponse  `json:"organization,omitempty"`
		BranchID                uuid.UUID              `json:"branch_id"`
		Branch                  *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID         uuid.UUID              `json:"member_profile_id"`
		MemberProfile           *MemberProfileResponse `json:"member_profile,omitempty"`
		RelativeMemberProfileID uuid.UUID              `json:"relative_member_profile_id"`
		RelativeMemberProfile   *MemberProfileResponse `json:"relative_member_profile,omitempty"`
		FamilyRelationship      string                 `json:"family_relationship"`
		Description             string                 `json:"description"`
	}

	MemberRelativeAccountRequest struct {
		MemberProfileID         uuid.UUID `json:"member_profile_id" validate:"required"`
		RelativeMemberProfileID uuid.UUID `json:"relative_member_profile_id" validate:"required"`
		FamilyRelationship      string    `json:"family_relationship" validate:"required,min=1,max=255"`
		Description             string    `json:"description,omitempty"`
	}
)

func MemberRelativeAccountManager(service *horizon.HorizonService) *registry.Registry[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "RelativeMemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberRelativeAccount) *MemberRelativeAccountResponse {
			if data == nil {
				return nil
			}
			return &MemberRelativeAccountResponse{
				ID:                      data.ID,
				CreatedAt:               data.CreatedAt.Format(time.RFC3339),
				CreatedByID:             data.CreatedByID,
				CreatedBy:               UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:               data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:             data.UpdatedByID,
				UpdatedBy:               UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:          data.OrganizationID,
				Organization:            OrganizationManager(service).ToModel(data.Organization),
				BranchID:                data.BranchID,
				Branch:                  BranchManager(service).ToModel(data.Branch),
				MemberProfileID:         data.MemberProfileID,
				MemberProfile:           MemberProfileManager(service).ToModel(data.MemberProfile),
				RelativeMemberProfileID: data.RelativeMemberProfileID,
				RelativeMemberProfile:   MemberProfileManager(service).ToModel(data.RelativeMemberProfile),
				FamilyRelationship:      data.FamilyRelationship,
				Description:             data.Description,
			}
		},

		Created: func(data *MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.create",
				fmt.Sprintf("member_relative_account.create.%s", data.ID),
				fmt.Sprintf("member_relative_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.update",
				fmt.Sprintf("member_relative_account.update.%s", data.ID),
				fmt.Sprintf("member_relative_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberRelativeAccount) registry.Topics {
			return []string{
				"member_relative_account.delete",
				fmt.Sprintf("member_relative_account.delete.%s", data.ID),
				fmt.Sprintf("member_relative_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberRelativeAccountCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberRelativeAccount, error) {
	return MemberRelativeAccountManager(service).Find(context, &MemberRelativeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
