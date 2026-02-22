package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func permissionTemplateSeed(
	context context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB, userID uuid.UUID,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) error {
	now := time.Now().UTC()
	permissionsTemplates := []*types.PermissionTemplate{
		{
			Name:        "Super Admin",
			Description: "Full access to all core resources and actions in the system",
			Permissions: pq.StringArray{
				// Members
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export",
				"MemberType:Create", "MemberType:Read", "MemberType:Update", "MemberType:Delete",
				"MemberGroup:Create", "MemberGroup:Read", "MemberGroup:Update", "MemberGroup:Delete",
				"MemberCenter:Create", "MemberCenter:Read", "MemberCenter:Update", "MemberCenter:Delete",
				// Accounts & Transactions
				"Account:Create", "Account:Read", "Account:Update", "Account:Delete", "Account:Export",
				"Transaction:Create", "Transaction:Read", "Transaction:Update",
				"QuickDeposit:Create", "QuickDeposit:Read", "QuickDeposit:Update",
				"QuickWithdraw:Create", "QuickWithdraw:Read", "QuickWithdraw:Update",
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Delete", "TransactionBatch:Export",
				// Loans
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Delete", "Loan:Export",
				"LoanScheme:Create", "LoanScheme:Read", "LoanScheme:Update", "LoanScheme:Delete",
				"Collateral:Create", "Collateral:Read", "Collateral:Update", "Collateral:Delete",
				// Accounting
				"JournalVoucher:Create", "JournalVoucher:Read", "JournalVoucher:Update", "JournalVoucher:Export",
				"GeneralLedger:Read", "FSDefinition:Create", "FSDefinition:Read", "GLDefinition:Create", "GLDefinition:Read",
				"AdjustmentEntry:Create", "AdjustmentEntry:Read", "AdjustmentEntry:Export",
				// Approvals (Using proper modules)
				"Approvals:Read", "ApprovalsJV:Read", "ApprovalsJVApproved:Read", "ApprovalsJVApproved:Update",
				"ApprovalsLoan:Read", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update",
				"ApprovalsCashVoucher:Read", "ApprovalsCashVoucherApproved:Read", "ApprovalsCashVoucherApproved:Update",
				// Setup & Config
				"Bank:Create", "Bank:Read", "Bank:Update", "Bank:Delete", "Bank:Export",
				"BillsAndCoins:Create", "BillsAndCoins:Read", "BillsAndCoins:Update", "BillsAndCoins:Delete", "BillsAndCoins:Export",
				"Holiday:Create", "Holiday:Read", "Holiday:Update", "Holiday:Delete", "Holiday:Export",
				"Company:Create", "Company:Read", "Company:Update", "Branch:Create", "Branch:Read", "Branch:Update",
				// Employees
				"Employee:Create", "Employee:Read", "Employee:Update", "Employee:Delete",
				"PermissionTemplate:Read", "Area:Create", "Area:Read", "Area:Update",
			},
		},
		{
			Name:        "Branch Manager",
			Description: "Manage branch operations, oversee approvals and member profiles",
			Permissions: pq.StringArray{
				"Dashboard:Read", "Branch:Read", "BranchSettings:Read", "BranchSettings:Update",
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export",
				"Account:Read", "AccountTransaction:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "Loan:Update", "Loan:Export",
				"TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				"Approvals:Read", "ApprovalsLoan:Read", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update",
				"ApprovalsCashVoucher:Read", "ApprovalsCashVoucherApproved:Read", "ApprovalsCashVoucherApproved:Update",
				"Employee:Read", "EmployeeFootstep:Read",
				"Bank:Read", "BillsAndCoins:Read", "CashCount:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Loan Officer",
			Description: "Process loan applications, collateral, and schemes",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read", "Collateral:Create", "Collateral:Read", "Collateral:Update",
				"LoanScheme:Read", "LoanChargeScheme:Read",
				"Account:Read", "AccountTransaction:Read",
				"ApprovalsLoanDraft:Read", "ApprovalsLoanPrinted:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Teller",
			Description: "Handle daily transactions, deposits, and withdrawals",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Read", "Account:Read", "AccountTransaction:Read",
				"Transaction:Create", "Transaction:Read", "Transaction:Update",
				"QuickDeposit:Create", "QuickDeposit:Read", "QuickDeposit:Update",
				"QuickWithdraw:Create", "QuickWithdraw:Read", "QuickWithdraw:Update",
				"TransactionBatch:Create", "TransactionBatch:Read",
				"BillsAndCoins:Read", "CashCount:Read",
				"TimeInOut:Create", "TimeInOut:Read", "MyTimesheet:Read", "MySettings:Read", "MySettings:Update",
			},
		},
		{
			Name:        "Accountant / Bookkeeper",
			Description: "Manage financial records, ledger definitions, and adjustments",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"JournalVoucher:Create", "JournalVoucher:Read", "JournalVoucher:Update", "JournalVoucher:Export",
				"GeneralLedger:Read", "AdjustmentEntry:Create", "AdjustmentEntry:Read", "AdjustmentEntry:Export",
				"FSDefinition:Create", "FSDefinition:Read", "FSDefinition:Update",
				"GLDefinition:Create", "GLDefinition:Read", "GLDefinition:Update",
				"ApprovalsJV:Read", "ApprovalsJVDraft:Read",
				"Bank:Read", "Holiday:Read", "Account:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read", "MyGeneralLedger:Read",
			},
		},
		{
			Name:        "Treasury / Cashier",
			Description: "Handle cash positions, disbursements, and remittances",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"CashCheckVoucher:Create", "CashCheckVoucher:Read", "CashCheckVoucher:Update", "CashCheckVoucher:Export",
				"DisburesmentType:Read", "DisbursementTransaction:Create",
				"CheckRemittance:Create", "CheckRemittance:Read", "CheckRemittance:Update",
				"OnlineRemittance:Create", "OnlineRemittance:Read", "OnlineRemittance:Update",
				"Bank:Read", "BillsAndCoins:Read", "CashCount:Read",
				"TransactionBatch:Read", "TransactionBatch:Export",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read", "MyDisbursements:Read",
			},
		},
		{
			Name:        "Member Services (CSR)",
			Description: "Handle member onboarding and basic inquiries",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update",
				"MemberProfileFileMediaUpload:Create", "MemberProfileFileMediaUpload:Read",
				"MemberType:Read", "MemberGroup:Read", "MemberCenter:Read",
				"MemberGender:Read", "MemberOccupation:Read", "MemberClassification:Read",
				"Account:Create", "Account:Read", "AccountClassification:Read", "AccountCategory:Read",
				"InvitationCode:Create", "InvitationCode:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Auditor / Compliance",
			Description: "Strict read-only access for auditing and risk assessment",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Read", "MemberProfile:Export", "MemberAccountingLedger:Read",
				"Account:Read", "Account:Export", "AccountTransaction:Read",
				"Loan:Read", "Loan:Export", "Collateral:Read",
				"JournalVoucher:Read", "GeneralLedger:Read", "AdjustmentEntry:Read",
				"TransactionBatch:Read", "TransactionBatchHistory:Read", "Transactions:Read",
				"Approvals:Read", "ApprovalsEndBatch:Read", "EmployeeFootstep:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Collections Officer",
			Description: "Monitor loans, follow-up overdue accounts, and view ledgers",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "Loan:Export", "LoanStatus:Read", "LoanTag:Read",
				"Account:Read", "AccountTransaction:Read",
				"TransactionBatch:Read", "Transactions:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Credit Approver",
			Description: "Evaluate and approve loan requests",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "LoanStatus:Read", "LoanPurpose:Read", "Collateral:Read",
				"Approvals:Read", "ApprovalsLoan:Read", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update",
				"ApprovalsLoanReleased:Read", "ApprovalsLoanReleased:Update",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "HR / Admin",
			Description: "Manage employees, permissions, schedules, and holidays",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"Employee:Create", "Employee:Read", "Employee:Update", "Employee:Export",
				"EmployeeSettings:Read", "EmployeePermission:Update",
				"PermissionTemplate:Read", "Timesheet:Read", "Timesheet:Update", "Timesheet:Export",
				"Holiday:Create", "Holiday:Read", "Holiday:Update", "Holiday:Delete",
				"Company:Read", "Branch:Read", "Area:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "System IT Support",
			Description: "Manage technical configs, API keys, and system diagnostics",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"Company:Read", "Company:Update", "Branch:Read", "BranchSettings:Read", "BranchSettings:Update",
				"ApiDoc:Read", "ApiKeyGen:Create", "User:Read",
				"Feed:Create", "Feed:Read", "Feed:Update", "Feed:Delete", "FeedComment:Delete",
				"TagTemplate:Create", "TagTemplate:Read", "TagTemplate:Update",
				"EmployeeFootstep:Read", "Footstep:Read",
			},
		},
		{
			Name:        "Credit Analyst",
			Description: "In-depth risk analysis of loan applications and scheme performance.",
			Permissions: pq.StringArray{
				"Loan:Read", "LoanStatus:Read", "LoanScheme:Read", "LoanChargeScheme:Read",
				"MemberProfile:Read", "MemberClassification:Read", "MemberGroup:Read",
				"AccountTransaction:Read", "GeneralLedger:Read",
			},
		},
		{
			Name:        "Collateral Specialist",
			Description: "Dedicated role for managing and valuing property/assets used as loan security.",
			Permissions: pq.StringArray{
				"Collateral:Create", "Collateral:Read", "Collateral:Update", "Collateral:Delete", "Collateral:Export",
				"Loan:Read", "MemberProfile:Read", "MemberProfileFileMediaUpload:Read",
			},
		},
		{
			Name:        "Collection / Remedial Officer",
			Description: "Focused on past-due accounts, repayment tracking, and collection batches.",
			Permissions: pq.StringArray{
				"Loan:Read", "Loan:Update", "LoanTag:Update", "LoanTag:Read",
				"AccountTransaction:Read", "TransactionBatch:Create", "TransactionBatch:Read",
				"MemberProfile:Read", "MemberAccountingLedger:Read",
			},
		},

		// --- TELLER & CASH OPERATIONS ---
		{
			Name:        "Head Teller",
			Description: "Oversees tellers, manages cash vaults, and handles end-of-batch.",
			Permissions: pq.StringArray{
				"Transaction:Read", "Transaction:Create", "Transaction:Update",
				"QuickDeposit:Read", "QuickDeposit:Create", "QuickDeposit:Update",
				"QuickWithdraw:Read", "QuickWithdraw:Create", "QuickWithdraw:Update",
				"CashCount:Read", "BillsAndCoins:Read", "BillsAndCoins:Update",
				"TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				"ApprovalsEndBatch:Read", "ApprovalsEndBatch:Update",
			},
		},
		{
			Name:        "General Teller",
			Description: "Standard day-to-day deposits, withdrawals, and timesheets.",
			Permissions: pq.StringArray{
				"Transaction:Create", "Transaction:Read",
				"QuickDeposit:Create", "QuickDeposit:Read",
				"QuickWithdraw:Create", "QuickWithdraw:Read",
				"BillsAndCoins:Read", "MemberProfile:Read", "Account:Read",
				"TimeInOut:Create", "TimeInOut:Read", "TimeInOut:Update", "TimeInOut:OwnUpdate",
			},
		},

		// --- ACCOUNTING & FINANCE ---
		{
			Name:        "Chief Accountant",
			Description: "Full access to the General Ledger, FS Definitions, and adjustments.",
			Permissions: pq.StringArray{
				"JournalVoucher:Read", "JournalVoucher:Create", "JournalVoucher:Update", "JournalVoucher:Export",
				"GeneralLedger:Read", "FSDefinition:Read", "FSDefinition:Create", "FSDefinition:Update", "FSDefinition:Delete",
				"GLDefinition:Read", "GLDefinition:Create", "GLDefinition:Update", "GLDefinition:Delete",
				"AdjustmentEntry:Read", "AdjustmentEntry:Create", "AdjustmentEntry:Export",
				"AccountClassification:Create", "AccountClassification:Read", "AccountClassification:Update",
			},
		},
		{
			Name:        "Disbursement Officer",
			Description: "Handles payouts, check remittances, and online transfers.",
			Permissions: pq.StringArray{
				"CashCheckVoucher:Create", "CashCheckVoucher:Read", "CashCheckVoucher:Update", "CashCheckVoucher:Export",
				"DisburesmentType:Read", "DisbursementTransaction:Create",
				"CheckRemittance:Create", "CheckRemittance:Read", "CheckRemittance:Update",
				"OnlineRemittance:Create", "OnlineRemittance:Read", "OnlineRemittance:Update",
				"Bank:Read", "Bank:Export",
			},
		},

		// --- MEMBERSHIP & CUSTOMER SERVICE ---
		{
			Name:        "Membership Coordinator",
			Description: "Specialist for onboarding members, managing groups, and departments.",
			Permissions: pq.StringArray{
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export",
				"MemberType:Read", "MemberGroup:Read", "MemberGroup:Update", "MemberCenter:Read", "MemberCenter:Update",
				"MemberDepartment:Read", "MemberDepartment:Update", "MemberOccupation:Read",
				"MemberProfileFileMediaUpload:Create", "MemberProfileFileMediaUpload:Read",
				"InvitationCode:Create", "InvitationCode:Read",
			},
		},
		{
			Name:        "Member Support Analyst",
			Description: "Inquiry desk role: views ledgers, histories, and profile details.",
			Permissions: pq.StringArray{
				"MemberProfile:Read", "MemberAccountingLedger:Read", "AccountTransaction:Read",
				"Loan:Read", "LoanStatus:Read", "Transaction:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},

		// --- APPROVALS & WORKFLOW ---
		{
			Name:        "JV Approver",
			Description: "Dedicated role for authorizing Journal Vouchers.",
			Permissions: pq.StringArray{
				"ApprovalsJV:Read", "ApprovalsJVApproved:Read", "ApprovalsJVApproved:Update",
				"ApprovalsJVReleased:Read", "ApprovalsJVReleased:Update",
				"JournalVoucher:Read",
			},
		},
		{
			Name:        "Loan Disbursement Manager",
			Description: "Final authority to release loan funds to members.",
			Permissions: pq.StringArray{
				"ApprovalsLoanApproved:Read", "ApprovalsLoanReleased:Read", "ApprovalsLoanReleased:Update",
				"Loan:Read", "Loan:Update", "DisbursementTransaction:Create",
			},
		},

		// --- COMPLIANCE & AUDIT ---
		{
			Name:        "Internal Auditor",
			Description: "Broad read-only access to all financial and employee logs.",
			Permissions: pq.StringArray{
				"GeneralLedger:Read", "JournalVoucher:Read", "AccountTransaction:Read",
				"TransactionBatchHistory:Read", "EmployeeFootstep:Read", "Footstep:Read",
				"MemberProfile:Read", "MemberProfileFileArchives:Read",
				"MemberAccountingLedger:Read", "Audit:Read", "Dashboard:Read",
			},
		},
		{
			Name:        "Risk & Compliance Officer",
			Description: "Monitors suspicious activity, account tags, and system settings.",
			Permissions: pq.StringArray{
				"AccountTag:Read", "LoanTag:Read", "MemberClassification:Read",
				"BranchSettings:Read", "Company:Read", "Audit:Read",
				"EmployeeFootstep:Read", "Footstep:Read", "AllMyFootsteps:Read",
			},
		},

		// --- HUMAN RESOURCES & IT ---
		{
			Name:        "HR Manager",
			Description: "Manages employee records, timesheets, and holidays.",
			Permissions: pq.StringArray{
				"Employee:Create", "Employee:Read", "Employee:Update", "Employee:Export",
				"EmployeeSettings:Read", "EmployeePermission:Update",
				"Timesheet:Read", "Timesheet:Update", "Timesheet:Export",
				"Holiday:Create", "Holiday:Read", "Holiday:Update", "Holiday:Delete",
				"MemberGender:Read", "MemberOccupation:Read",
			},
		},
		{
			Name:        "IT / System Admin",
			Description: "Technical maintenance, API management, and developer documentation.",
			Permissions: pq.StringArray{
				"ApiDoc:Read", "ApiKeyGen:Create", "User:Read", "PermissionTemplate:Read",
				"TagTemplate:Create", "TagTemplate:Read", "TagTemplate:Update",
				"Feed:Create", "Feed:Read", "Feed:Update", "Feed:Delete",
				"Company:Update", "BranchSettings:Update",
			},
		},

		// --- SPECIALIZED ROLES ---
		{
			Name:        "Field Agent / Collector",
			Description: "Mobile-focused role for field work and quick deposits.",
			Permissions: pq.StringArray{
				"MemberProfile:Read", "MemberProfile:OwnRead",
				"QuickDeposit:Create", "QuickDeposit:Read",
				"Loan:Read", "Loan:OwnRead",
				"TimeInOut:Create", "TimeInOut:Read",
				"Footstep:Read", "MyBranchFootsteps:Read",
			},
		},
		{
			Name:        "Savings Interest Specialist",
			Description: "Handles the generation of interest and mutual fund aids.",
			Permissions: pq.StringArray{
				"GenerateSavingsInterest:Read", "GenerateSavingsInterest:Create", "GenerateSavingsInterest:Update", "GenerateSavingsInterest:Export",
				"GenerateMutualFundAid:Read", "GenerateMutualFundAid:Create", "GenerateMutualFundAid:Update",
				"Account:Read", "AccountTransaction:Read",
			},
		},
		{
			Name:        "External Auditor (Read-Only)",
			Description: "Strictly read-only access for 3rd party auditing firms.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "GeneralLedger:Read", "JournalVoucher:Read", "Loan:Read", "MemberProfile:Read",
				"TransactionBatch:Read", "Bank:Read", "FSDefinition:Read", "GLDefinition:Read",
			},
		},
		{
			Name:        "Basic Staff",
			Description: "Standard employee with access to their own files and feed.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "Feed:Read", "FeedComment:Create", "FeedComment:OwnDelete",
				"TimeInOut:Create", "TimeInOut:Read", "MyTimesheet:Read",
				"MySettings:Read", "MySettings:Update", "MyGeneralLedger:Read",
				"AllMyFootsteps:Read", "Holiday:Read",
			},
		},
		{
			Name:        "Basic Employee",
			Description: "Limited access for general staff (Timesheets, feeds, settings)",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"TimeInOut:Create", "TimeInOut:Read", "TimeInOut:Update",
				"MyTimesheet:Read", "MySettings:Read", "MySettings:Update",
				"AllMyFootsteps:Read", "MyBranchFootsteps:Read",
				"Feed:Read", "FeedComment:Create", "FeedComment:OwnDelete",
				"Holiday:Read",
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
		if err := PermissionTemplateManager(service).CreateWithTx(context, tx, permission); err != nil {
			return eris.Wrapf(err, "failed to seed permission template %s", permission.Name)
		}
	}
	return nil
}

func PermissionTemplateManager(service *horizon.HorizonService) *registry.Registry[types.PermissionTemplate, types.PermissionTemplateResponse, types.PermissionTemplateRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.PermissionTemplate, types.PermissionTemplateResponse, types.PermissionTemplateRequest]{
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
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.PermissionTemplate) *types.PermissionTemplateResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}
			return &types.PermissionTemplateResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				Name:        data.Name,
				Description: data.Description,
				Permissions: data.Permissions,
			}
		},

		Created: func(data *types.PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.create",
				fmt.Sprintf("permission_template.create.%s", data.ID),
				fmt.Sprintf("permission_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.update",
				fmt.Sprintf("permission_template.update.%s", data.ID),
				fmt.Sprintf("permission_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.PermissionTemplate) registry.Topics {
			return []string{
				"permission_template.delete",
				fmt.Sprintf("permission_template.delete.%s", data.ID),
				fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetPermissionTemplateBybranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.PermissionTemplate, error) {
	return PermissionTemplateManager(service).Find(context, &types.PermissionTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
