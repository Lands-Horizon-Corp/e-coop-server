package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// PermissionTemplate represents predefined sets of permissions for users within organizations and branches
	PermissionTemplate struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string         `gorm:"type:varchar(255);not null"`
		Description string         `gorm:"type:text"`
		Permissions pq.StringArray `gorm:"type:varchar[];default:'{}'"`
	}

	// PermissionTemplateRequest represents the request structure for creating or updating permission templates
	PermissionTemplateRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name        string   `json:"name" validate:"required,min=1,max=255"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions,omitempty"`
	}

	// PermissionTemplateResponse represents the response structure for permission template data
	PermissionTemplateResponse struct {
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

		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions"`
	}
)

func (m *Core) permissionTemplateSeed(
	context context.Context,
	tx *gorm.DB, userID uuid.UUID,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) error {
	now := time.Now().UTC()
	permissionsTemplates := []*PermissionTemplate{
		{
			Name:        "Super Admin",
			Description: "Full access to all resources and actions in the system",
			Permissions: pq.StringArray{
				// Member Management
				"MemberType:Create", "MemberType:Read", "MemberType:Update", "MemberType:Delete", "MemberType:Export",
				"MemberType:OwnRead", "MemberType:OwnUpdate", "MemberType:OwnDelete", "MemberType:OwnExport",
				"MemberGroup:Create", "MemberGroup:Read", "MemberGroup:Update", "MemberGroup:Delete", "MemberGroup:Export",
				"MemberGroup:OwnRead", "MemberGroup:OwnUpdate", "MemberGroup:OwnDelete", "MemberGroup:OwnExport",
				"MemberCenter:Create", "MemberCenter:Read", "MemberCenter:Update", "MemberCenter:Delete", "MemberCenter:Export",
				"MemberCenter:OwnRead", "MemberCenter:OwnUpdate", "MemberCenter:OwnDelete", "MemberCenter:OwnExport",
				"MemberGender:Create", "MemberGender:Read", "MemberGender:Update", "MemberGender:Delete", "MemberGender:Export",
				"MemberGender:OwnRead", "MemberGender:OwnUpdate", "MemberGender:OwnDelete", "MemberGender:OwnExport",
				"MemberOccupation:Create", "MemberOccupation:Read", "MemberOccupation:Update", "MemberOccupation:Delete", "MemberOccupation:Export",
				"MemberOccupation:OwnRead", "MemberOccupation:OwnUpdate", "MemberOccupation:OwnDelete", "MemberOccupation:OwnExport",
				"MemberClassification:Create", "MemberClassification:Read", "MemberClassification:Update", "MemberClassification:Delete", "MemberClassification:Export",
				"MemberClassification:OwnRead", "MemberClassification:OwnUpdate", "MemberClassification:OwnDelete", "MemberClassification:OwnExport",
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Delete", "MemberProfile:Export",
				"MemberProfile:OwnRead", "MemberProfile:OwnUpdate", "MemberProfile:OwnDelete", "MemberProfile:OwnExport",
				// Banking & Finance
				"Banks:Create", "Banks:Read", "Banks:Update", "Banks:Delete", "Banks:Export",
				"Banks:OwnRead", "Banks:OwnUpdate", "Banks:OwnDelete", "Banks:OwnExport",
				"BillsAndCoin:Create", "BillsAndCoin:Read", "BillsAndCoin:Update", "BillsAndCoin:Delete", "BillsAndCoin:Export",
				"BillsAndCoin:OwnRead", "BillsAndCoin:OwnUpdate", "BillsAndCoin:OwnDelete", "BillsAndCoin:OwnExport",
				// Loan Management
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Delete", "Loan:Export",
				"Loan:OwnRead", "Loan:OwnUpdate", "Loan:OwnDelete", "Loan:OwnExport",
				"LoanStatus:Create", "LoanStatus:Read", "LoanStatus:Update", "LoanStatus:Delete", "LoanStatus:Export",
				"LoanStatus:OwnRead", "LoanStatus:OwnUpdate", "LoanStatus:OwnDelete", "LoanStatus:OwnExport",
				"LoanPurpose:Create", "LoanPurpose:Read", "LoanPurpose:Update", "LoanPurpose:Delete", "LoanPurpose:Export",
				"LoanPurpose:OwnRead", "LoanPurpose:OwnUpdate", "LoanPurpose:OwnDelete", "LoanPurpose:OwnExport",
				// System & Administration
				"Approvals:Read", "Approvals:Approve",
				"Holidays:Create", "Holidays:Read", "Holidays:Update", "Holidays:Delete", "Holidays:Export",
				"Holidays:OwnRead", "Holidays:OwnUpdate", "Holidays:OwnDelete", "Holidays:OwnExport",
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Delete", "TransactionBatch:Export",
				"TransactionBatch:OwnRead", "TransactionBatch:OwnUpdate", "TransactionBatch:OwnDelete", "TransactionBatch:OwnExport",
				"InvitationCode:Create", "InvitationCode:Read", "InvitationCode:Update", "InvitationCode:Delete", "InvitationCode:Export",
				"InvitationCode:OwnRead", "InvitationCode:OwnUpdate", "InvitationCode:OwnDelete", "InvitationCode:OwnExport",
				"Timesheet:Create", "Timesheet:Read", "Timesheet:Update", "Timesheet:Delete", "Timesheet:Export",
				"Timesheet:OwnRead", "Timesheet:OwnUpdate", "Timesheet:OwnDelete", "Timesheet:OwnExport",
				"Footstep:Create", "Footstep:Read", "Footstep:Update", "Footstep:Delete", "Footstep:Export",
				"Footstep:OwnRead", "Footstep:OwnUpdate", "Footstep:OwnDelete", "Footstep:OwnExport",
			},
		},
		{
			Name:        "Branch Manager",
			Description: "Manage branch operations, approve loans and transactions",
			Permissions: pq.StringArray{
				// Member Management - Full access
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				"MemberGender:Read", "MemberOccupation:Read", "MemberClassification:Read",
				// Loan Management - Full access
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read",
				// Approvals
				"Approvals:Read", "Approvals:Approve",
				// Banking
				"Banks:Read", "BillsAndCoin:Read",
				// Transactions
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				// System
				"Holidays:Read", "Timesheet:Read", "Footstep:Read",
			},
		},
		{
			Name:        "Loan Officer",
			Description: "Process and manage loan applications and payments",
			Permissions: pq.StringArray{
				// Member - Read only
				"MemberProfile:Read", "MemberProfile:Export",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				// Loan Management - Full access except delete
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read",
				// Transactions
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Export",
				// Banking
				"Banks:Read", "BillsAndCoin:Read",
				// Own records
				"Loan:OwnRead", "Loan:OwnUpdate", "Loan:OwnExport",
			},
		},
		{
			Name:        "Teller",
			Description: "Handle daily transactions and member services",
			Permissions: pq.StringArray{
				// Member - Read and basic updates
				"MemberProfile:Read", "MemberProfile:Update",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				// Transactions
				"TransactionBatch:Create", "TransactionBatch:Read",
				// Loan - Read only
				"Loan:Read", "LoanStatus:Read",
				// Banking
				"Banks:Read", "BillsAndCoin:Read",
				// Own records
				"TransactionBatch:OwnRead", "TransactionBatch:OwnUpdate",
				"Timesheet:Create", "Timesheet:OwnRead", "Timesheet:OwnUpdate",
			},
		},
		{
			Name:        "Accountant",
			Description: "Manage financial records and generate reports",
			Permissions: pq.StringArray{
				// Member - Read only
				"MemberProfile:Read", "MemberProfile:Export",
				// Loan - Read and export
				"Loan:Read", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read",
				// Transactions - Full access
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				// Banking
				"Banks:Read", "Banks:Export",
				"BillsAndCoin:Read", "BillsAndCoin:Export",
				// System
				"Holidays:Read",
			},
		},
		{
			Name:        "Member Services",
			Description: "Handle member inquiries and basic account management",
			Permissions: pq.StringArray{
				// Member Management
				"MemberProfile:Read", "MemberProfile:Update",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				"MemberGender:Read", "MemberOccupation:Read", "MemberClassification:Read",
				// Loan - Read only
				"Loan:Read", "LoanStatus:Read",
				// Own records
				"MemberProfile:OwnRead", "MemberProfile:OwnUpdate",
				"Timesheet:Create", "Timesheet:OwnRead",
			},
		},
		{
			Name:        "Auditor",
			Description: "Read-only access for audit and compliance purposes",
			Permissions: pq.StringArray{
				// Read and export all
				"MemberProfile:Read", "MemberProfile:Export",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				"Loan:Read", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read",
				"TransactionBatch:Read", "TransactionBatch:Export",
				"Banks:Read", "Banks:Export",
				"BillsAndCoin:Read",
				"Approvals:Read",
				"Footstep:Read", "Footstep:Export",
				"Timesheet:Read", "Timesheet:Export",
			},
		},
		{
			Name:        "Basic User",
			Description: "Limited access for general staff members",
			Permissions: pq.StringArray{
				// Own records only
				"MemberProfile:OwnRead",
				"Timesheet:Create", "Timesheet:OwnRead", "Timesheet:OwnUpdate",
				"Footstep:OwnRead",
				// System - Read only
				"Holidays:Read",
			},
		},

		// Additional cooperative-specific roles
		{
			Name:        "Collections Officer",
			Description: "Manage loan collections, track repayments and follow-up overdue accounts",
			Permissions: pq.StringArray{
				"Loan:Read", "Loan:OwnUpdate", "Loan:OwnRead", "Loan:Export",
				"LoanStatus:Read", "LoanStatus:OwnRead",
				"TransactionBatch:Create", "TransactionBatch:Read",
				"Approvals:Read",
			},
		},
		{
			Name:        "Cashier",
			Description: "Handle cash transactions and daily cash reconciliation",
			Permissions: pq.StringArray{
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Export",
				"Banks:Read", "BillsAndCoin:Read",
				"Timesheet:OwnRead", "Timesheet:OwnUpdate",
			},
		},
		{
			Name:        "Credit Analyst",
			Description: "Evaluate loan applications and recommend credit decisions",
			Permissions: pq.StringArray{
				"Loan:Read", "Loan:Update", "LoanStatus:Read", "LoanPurpose:Read",
				"MemberProfile:Read",
				"TransactionBatch:Read",
			},
		},
		{
			Name:        "Risk & Compliance",
			Description: "Monitor risk, compliance and perform audits",
			Permissions: pq.StringArray{
				"Loan:Read", "Loan:Export", "TransactionBatch:Read", "TransactionBatch:Export",
				"Approvals:Read", "MemberProfile:Read", "Timesheet:Read",
			},
		},
		{
			Name:        "Treasurer",
			Description: "Oversee treasury operations, cash positions and settlements",
			Permissions: pq.StringArray{
				"Banks:Read", "Banks:Export", "BillsAndCoin:Read", "TransactionBatch:Read", "TransactionBatch:Export",
				"Loan:Read",
			},
		},
		{
			Name:        "Cooperative Admin",
			Description: "Administrative role for cooperative-specific configuration and management",
			Permissions: pq.StringArray{
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Delete", "MemberProfile:Export",
				"Banks:Create", "Banks:Read", "Banks:Update", "Banks:Delete", "Banks:Export",
				"Holidays:Create", "Holidays:Read", "Holidays:Update", "Holidays:Delete", "Holidays:Export",
				"InvitationCode:Create", "InvitationCode:Read", "InvitationCode:Update", "InvitationCode:Delete", "InvitationCode:Export",
			},
		},
	}
	for _, permission := range permissionsTemplates {
		permission.CreatedAt = now
		permission.UpdatedAt = now
		permission.OrganizationID = organizationID
		permission.BranchID = branchID
		permission.UpdatedByID = userID
		permission.CreatedByID = userID
		if err := m.PermissionTemplateManager.CreateWithTx(context, tx, permission); err != nil {
			return eris.Wrapf(err, "failed to seed permission template %s", permission.Name)
		}
	}
	return nil
}

// PermissionTemplate initializes the permission template model and its repository manager
func (m *Core) permissionTemplate() {
	m.Migration = append(m.Migration, &PermissionTemplate{})
	m.PermissionTemplateManager = *registry.NewRegistry(registry.RegistryParams[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Branch.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Organization.OrganizationCategories",
			"Organization.OrganizationCategories.Category",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *PermissionTemplate) *PermissionTemplateResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}
			return &PermissionTemplateResponse{
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

				Name:        data.Name,
				Description: data.Description,
				Permissions: data.Permissions,
			}
		},

		Created: func(data *PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.create",
				fmt.Sprintf("permission_template.create.%s", data.ID),
				fmt.Sprintf("permission_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.update",
				fmt.Sprintf("permission_template.update.%s", data.ID),
				fmt.Sprintf("permission_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.delete",
				fmt.Sprintf("permission_template.delete.%s", data.ID),
				fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GetPermissionTemplateBybranch retrieves permission templates for a specific branch within an organization
func (m *Core) GetPermissionTemplateBybranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*PermissionTemplate, error) {
	return m.PermissionTemplateManager.Find(context, &PermissionTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
