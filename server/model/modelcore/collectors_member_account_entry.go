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
	CollectorsMemberAccountEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collectors_member_account_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_collectors_member_account_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CollectorUserID *uuid.UUID     `gorm:"type:uuid"`
		CollectorUser   *User          `gorm:"foreignKey:CollectorUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"collector_user,omitempty"`
		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		AccountID       *uuid.UUID     `gorm:"type:uuid"`
		Account         *Account       `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		Description string `gorm:"type:text"`
	}

	CollectorsMemberAccountEntryResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		CollectorUserID *uuid.UUID             `json:"collector_user_id,omitempty"`
		CollectorUser   *UserResponse          `json:"collector_user,omitempty"`
		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		AccountID       *uuid.UUID             `json:"account_id,omitempty"`
		Account         *AccountResponse       `json:"account,omitempty"`
		Description     string                 `json:"description"`
	}

	CollectorsMemberAccountEntryRequest struct {
		CollectorUserID *uuid.UUID `json:"collector_user_id,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
		AccountID       *uuid.UUID `json:"account_id,omitempty"`
		Description     string     `json:"description,omitempty"`
	}
)

func (m *ModelCore) collectorsMemberAccountEntry() {
	m.Migration = append(m.Migration, &CollectorsMemberAccountEntry{})
	m.CollectorsMemberAccountEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		CollectorsMemberAccountEntry, CollectorsMemberAccountEntryResponse, CollectorsMemberAccountEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"CollectorUser", "MemberProfile", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *CollectorsMemberAccountEntry) *CollectorsMemberAccountEntryResponse {
			if data == nil {
				return nil
			}
			return &CollectorsMemberAccountEntryResponse{
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
				CollectorUserID: data.CollectorUserID,
				CollectorUser:   m.UserManager.ToModel(data.CollectorUser),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				AccountID:       data.AccountID,
				Account:         m.AccountManager.ToModel(data.Account),
				Description:     data.Description,
			}
		},
		Created: func(data *CollectorsMemberAccountEntry) []string {
			return []string{
				"collectors_member_account_entry.create",
				fmt.Sprintf("collectors_member_account_entry.create.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CollectorsMemberAccountEntry) []string {
			return []string{
				"collectors_member_account_entry.update",
				fmt.Sprintf("collectors_member_account_entry.update.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CollectorsMemberAccountEntry) []string {
			return []string{
				"collectors_member_account_entry.delete",
				fmt.Sprintf("collectors_member_account_entry.delete.%s", data.ID),
				fmt.Sprintf("collectors_member_account_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("collectors_member_account_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) CollectorsMemberAccountEntryCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CollectorsMemberAccountEntry, error) {
	return m.CollectorsMemberAccountEntryManager.Find(context, &CollectorsMemberAccountEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
