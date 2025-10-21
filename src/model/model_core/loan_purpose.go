package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	LoanPurpose struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_purpose"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_purpose"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Description string `gorm:"type:text"`
		Icon        string `gorm:"type:varchar(255)"`
	}

	LoanPurposeResponse struct {
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
		Description    string                `json:"description"`
		Icon           string                `json:"icon"`
	}

	LoanPurposeRequest struct {
		Description string `json:"description,omitempty"`
		Icon        string `json:"icon,omitempty"`
	}
)

func (m *ModelCore) LoanPurpose() {
	m.Migration = append(m.Migration, &LoanPurpose{})
	m.LoanPurposeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanPurpose, LoanPurposeResponse, LoanPurposeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanPurpose) *LoanPurposeResponse {
			if data == nil {
				return nil
			}
			return &LoanPurposeResponse{
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
				Description:    data.Description,
				Icon:           data.Icon,
			}
		},

		Created: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.create",
				fmt.Sprintf("loan_purpose.create.%s", data.ID),
				fmt.Sprintf("loan_purpose.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.update",
				fmt.Sprintf("loan_purpose.update.%s", data.ID),
				fmt.Sprintf("loan_purpose.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.delete",
				fmt.Sprintf("loan_purpose.delete.%s", data.ID),
				fmt.Sprintf("loan_purpose.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) LoanPurposeSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	loanPurposes := []*LoanPurpose{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Home Purchase/Construction",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Vehicle Purchase",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Business Capital",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Education/Tuition Fee",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Medical/Healthcare",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Emergency/Personal",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Agricultural/Farming",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Debt Consolidation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Home Improvement/Renovation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Appliance/Electronics",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Wedding/Special Occasion",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Equipment/Machinery",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Investment/Securities",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Description:    "Other/Miscellaneous",
		},
	}

	for _, data := range loanPurposes {
		if err := m.LoanPurposeManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed loan purpose %s", data.Description)
		}
	}

	return nil
}

func (m *ModelCore) LoanPurposeCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanPurpose, error) {
	return m.LoanPurposeManager.Find(context, &LoanPurpose{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
