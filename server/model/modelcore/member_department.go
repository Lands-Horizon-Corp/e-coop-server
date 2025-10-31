package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	MemberDepartment struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_department" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_department" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string  `gorm:"type:varchar(255);not null" json:"name"`
		Description string  `gorm:"type:text" json:"description"`
		Icon        *string `gorm:"type:varchar(255)" json:"icon,omitempty"`
	}

	MemberDepartmentResponse struct {
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
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		Icon           *string               `json:"icon,omitempty"`
	}

	MemberDepartmentRequest struct {
		Name        string  `json:"name" validate:"required,min=1,max=255"`
		Description string  `json:"description,omitempty"`
		Icon        *string `json:"icon,omitempty"`
	}
)

func (m *ModelCore) memberDepartment() {
	m.Migration = append(m.Migration, &MemberDepartment{})
	m.MemberDepartmentManager = services.NewRepository(services.RepositoryParams[MemberDepartment, MemberDepartmentResponse, MemberDepartmentRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *MemberDepartment) *MemberDepartmentResponse {
			if data == nil {
				return nil
			}
			return &MemberDepartmentResponse{
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
				Name:           data.Name,
				Description:    data.Description,
				Icon:           data.Icon,
			}
		},
		Created: func(data *MemberDepartment) []string {
			return []string{
				"member_department.create",
				fmt.Sprintf("member_department.create.%s", data.ID),
				fmt.Sprintf("member_department.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberDepartment) []string {
			return []string{
				"member_department.update",
				fmt.Sprintf("member_department.update.%s", data.ID),
				fmt.Sprintf("member_department.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberDepartment) []string {
			return []string{
				"member_department.delete",
				fmt.Sprintf("member_department.delete.%s", data.ID),
				fmt.Sprintf("member_department.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_department.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) memberDepartmentSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberDepartments := []*MemberDepartment{
		{
			Name:        "General Services",
			Description: "Handles general inquiries, member registration, and basic services.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Loans and Credit",
			Description: "Manages loan applications, approvals, and credit services.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Savings and Deposits",
			Description: "Handles savings accounts, deposits, and withdrawal transactions.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Accounting and Finance",
			Description: "Manages financial records, budgeting, and accounting operations.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Member Relations",
			Description: "Focuses on member engagement, complaints, and relationship management.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Information Technology",
			Description: "Manages IT systems, software, hardware, and technical support.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Marketing and Communications",
			Description: "Handles promotional activities, communications, and public relations.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Human Resources",
			Description: "Manages employee relations, recruitment, and organizational development.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Audit and Compliance",
			Description:    "Ensures regulatory compliance and conducts internal audits.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Legal Affairs",
			Description: "Handles legal matters, contracts, and regulatory compliance.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:        "Insurance Services",
			Description: "Manages insurance products and services for members.",

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Education and Training",
			Description:    "Provides financial literacy and cooperative education programs.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Security and Risk Management",
			Description:    "Handles security protocols and risk assessment procedures.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
		{
			Name:           "Operations Management",
			Description:    "Oversees daily operations and process improvement initiatives.",
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
		},
	}

	for _, data := range memberDepartments {
		if err := m.MemberDepartmentManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member department %s", data.Name)
		}
	}
	return nil
}

func (m *ModelCore) MemberDepartmentCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberDepartment, error) {
	return m.MemberDepartmentManager.Find(context, &MemberDepartment{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
