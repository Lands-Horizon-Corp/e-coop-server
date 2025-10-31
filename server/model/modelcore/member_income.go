package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// MemberIncome represents a member's income information in the database
	MemberIncome struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_income"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_income"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID         *uuid.UUID     `gorm:"type:uuid" json:"media_id,omitempty"`
		Media           *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"media,omitempty"`
		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name        string     `gorm:"type:varchar(255)"`
		Source      string     `gorm:"type:varchar(255)"`
		Amount      float64    `gorm:"type:decimal(20,6)"`
		ReleaseDate *time.Time `gorm:"type:timestamp"`
	}

	// MemberIncomeResponse represents the response structure for member income data
	MemberIncomeResponse struct {
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
		MediaID         *uuid.UUID             `json:"media_id,omitempty"`
		Media           *MediaResponse         `json:"media,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		Name            string                 `json:"name"`
		Source          string                 `json:"source"`
		Amount          float64                `json:"amount"`
		ReleaseDate     *string                `json:"release_date,omitempty"`
	}

	// MemberIncomeRequest represents the request structure for creating/updating member income
	MemberIncomeRequest struct {
		MediaID     *uuid.UUID `json:"media_id"`
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Source      string     `json:"source" validate:"required,min=1,max=255"`
		Amount      float64    `json:"amount" validate:"required"`
		ReleaseDate *time.Time `json:"release_date,omitempty"`
	}
)

func (m *ModelCore) memberIncome() {
	m.Migration = append(m.Migration, &MemberIncome{})
	m.MemberIncomeManager = services.NewRepository(services.RepositoryParams[MemberIncome, MemberIncomeResponse, MemberIncomeRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberIncome) *MemberIncomeResponse {
			if data == nil {
				return nil
			}
			var releaseDateStr *string
			if data.ReleaseDate != nil {
				s := data.ReleaseDate.Format(time.RFC3339)
				releaseDateStr = &s
			}
			return &MemberIncomeResponse{
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
				Source:          data.Source,
				Amount:          data.Amount,
				ReleaseDate:     releaseDateStr,
			}
		},

		Created: func(data *MemberIncome) []string {
			return []string{
				"member_income.create",
				fmt.Sprintf("member_income.create.%s", data.ID),
				fmt.Sprintf("member_income.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberIncome) []string {
			return []string{
				"member_income.update",
				fmt.Sprintf("member_income.update.%s", data.ID),
				fmt.Sprintf("member_income.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberIncome) []string {
			return []string{
				"member_income.delete",
				fmt.Sprintf("member_income.delete.%s", data.ID),
				fmt.Sprintf("member_income.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_income.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// MemberIncomeCurrentBranch retrieves member income records for a specific organization branch
func (m *ModelCore) MemberIncomeCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberIncome, error) {
	return m.MemberIncomeManager.Find(context, &MemberIncome{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
