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
	MemberAsset struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_organization_branch_member_asset"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_organization_branch_member_asset"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"media,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name        string    `gorm:"type:varchar(255);not null"`
		EntryDate   time.Time `gorm:"not null"`
		Description string    `gorm:"type:text;"`
		Cost        float64   `gorm:"type:decimal(20,6);default:0"`
	}

	MemberAssetResponse struct {
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

		MediaID *uuid.UUID     `json:"media_id,omitempty"`
		Media   *MediaResponse `json:"media,omitempty"`

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		Name        string  `json:"name"`
		EntryDate   string  `json:"entry_date"`
		Description string  `json:"description"`
		Cost        float64 `json:"cost"`
	}

	MemberAssetRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		EntryDate   time.Time  `json:"entry_date" validate:"required"`
		Description string     `json:"description" validate:"required"`
		Cost        float64    `json:"cost,omitempty"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *ModelCore) MemberAsset() {
	m.Migration = append(m.Migration, &MemberAsset{})
	m.MemberAssetManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberAsset, MemberAssetResponse, MemberAssetRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberAsset) *MemberAssetResponse {
			if data == nil {
				return nil
			}
			return &MemberAssetResponse{
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
				MediaID:         data.MediaID,
				Media:           m.MediaManager.ToModel(data.Media),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				Name:            data.Name,
				EntryDate:       data.EntryDate.Format(time.RFC3339),
				Description:     data.Description,
				Cost:            data.Cost,
			}
		},

		Created: func(data *MemberAsset) []string {
			return []string{
				"member_asset.create",
				fmt.Sprintf("member_asset.create.%s", data.ID),
				fmt.Sprintf("member_asset.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberAsset) []string {
			return []string{
				"member_asset.update",
				fmt.Sprintf("member_asset.update.%s", data.ID),
				fmt.Sprintf("member_asset.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberAsset) []string {
			return []string{
				"member_asset.delete",
				fmt.Sprintf("member_asset.delete.%s", data.ID),
				fmt.Sprintf("member_asset.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_asset.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberAssetCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberAsset, error) {
	return m.MemberAssetManager.Find(context, &MemberAsset{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
