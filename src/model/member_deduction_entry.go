package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberDeductionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_deduction_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_deduction_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		Name           string    `gorm:"type:varchar(255)"`
		Description    string    `gorm:"type:text"`
		MembershipDate time.Time `gorm:"type:timestamp"`
	}

	MemberDeductionEntryResponse struct {
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
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		AccountID       uuid.UUID              `json:"account_id"`
		Account         *AccountResponse       `json:"account,omitempty"`
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		MembershipDate  string                 `json:"membership_date"`
	}

	MemberDeductionEntryRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		AccountID       uuid.UUID `json:"account_id" validate:"required"`
		Name            string    `json:"name,omitempty"`
		Description     string    `json:"description,omitempty"`
		MembershipDate  time.Time `json:"membership_date,omitempty"`
	}
)

func (m *Model) MemberDeductionEntry() {
	m.Migration = append(m.Migration, &MemberDeductionEntry{})
	m.MemberDeductionEntryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberDeductionEntry, MemberDeductionEntryResponse, MemberDeductionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "MemberProfile", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberDeductionEntry) *MemberDeductionEntryResponse {
			if data == nil {
				return nil
			}
			return &MemberDeductionEntryResponse{
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
				AccountID:       data.AccountID,
				Account:         m.AccountManager.ToModel(data.Account),
				Name:            data.Name,
				Description:     data.Description,
				MembershipDate:  data.MembershipDate.Format(time.RFC3339),
			}
		},
		Created: func(data *MemberDeductionEntry) []string {
			return []string{
				"member_deduction_entry.create",
				fmt.Sprintf("member_deduction_entry.create.%s", data.ID),
			}
		},
		Updated: func(data *MemberDeductionEntry) []string {
			return []string{
				"member_deduction_entry.update",
				fmt.Sprintf("member_deduction_entry.update.%s", data.ID),
			}
		},
		Deleted: func(data *MemberDeductionEntry) []string {
			return []string{
				"member_deduction_entry.delete",
				fmt.Sprintf("member_deduction_entry.delete.%s", data.ID),
			}
		},
	})
}
