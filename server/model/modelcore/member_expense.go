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
	MemberExpense struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_expense"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_expense"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name        string  `gorm:"type:varchar(255)"`
		Amount      float64 `gorm:"type:decimal(20,6)"`
		Description string  `gorm:"type:text"`
	}

	MemberExpenseResponse struct {
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
		Name            string                 `json:"name"`
		Amount          float64                `json:"amount"`
		Description     string                 `json:"description"`
	}

	MemberExpenseRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		Name            string    `json:"name" validate:"required,min=1,max=255"`
		Amount          float64   `json:"amount" validate:"required"`
		Description     string    `json:"description,omitempty"`
	}
)

func (m *ModelCore) MemberExpense() {
	m.Migration = append(m.Migration, &MemberExpense{})
	m.MemberExpenseManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberExpense, MemberExpenseResponse, MemberExpenseRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberExpense) *MemberExpenseResponse {
			if data == nil {
				return nil
			}
			return &MemberExpenseResponse{
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
				Name:            data.Name,
				Amount:          data.Amount,
				Description:     data.Description,
			}
		},

		Created: func(data *MemberExpense) []string {
			return []string{
				"member_expense.create",
				fmt.Sprintf("member_expense.create.%s", data.ID),
				fmt.Sprintf("member_expense.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberExpense) []string {
			return []string{
				"member_expense.update",
				fmt.Sprintf("member_expense.update.%s", data.ID),
				fmt.Sprintf("member_expense.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberExpense) []string {
			return []string{
				"member_expense.delete",
				fmt.Sprintf("member_expense.delete.%s", data.ID),
				fmt.Sprintf("member_expense.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_expense.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberExpenseCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberExpense, error) {
	return m.MemberExpenseManager.Find(context, &MemberExpense{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
