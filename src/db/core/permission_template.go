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
		// ==========================================
		// 1. SYSTEM ADMINISTRATION
		// ==========================================
		{
			Name:        "Super Admin",
			Description: "Full access to all core resources and actions in the system. Bypasses standard restrictions.",
			Permissions: pq.StringArray{
				// Dashboard
				"Dashboard:Create", "Dashboard:Read", "Dashboard:Update", "Dashboard:Delete", "Dashboard:Export", "Dashboard:OwnRead", "Dashboard:OwnUpdate", "Dashboard:OwnDelete", "Dashboard:OwnExport",
				// Accounts
				"Account:Create", "Account:Read", "Account:Update", "Account:Delete", "Account:Export", "Account:OwnRead", "Account:OwnUpdate", "Account:OwnDelete", "Account:OwnExport",
				"AccountClassification:Create", "AccountClassification:Read", "AccountClassification:Update", "AccountClassification:Delete", "AccountClassification:Export", "AccountClassification:OwnRead", "AccountClassification:OwnUpdate", "AccountClassification:OwnDelete", "AccountClassification:OwnExport",
				"AccountCategory:Create", "AccountCategory:Read", "AccountCategory:Update", "AccountCategory:Delete", "AccountCategory:Export", "AccountCategory:OwnRead", "AccountCategory:OwnUpdate", "AccountCategory:OwnDelete", "AccountCategory:OwnExport",
				"AccountTag:Read", "AccountTransaction:Read", "AccountTransaction:Create",
				// Transactions & Adjustments
				"Transaction:Read", "Transaction:Create", "Transaction:Update",
				"QuickWithdraw:Read", "QuickWithdraw:Create", "QuickWithdraw:Update",
				"QuickDeposit:Read", "QuickDeposit:Create", "QuickDeposit:Update",
				"PaymentType:Create", "PaymentType:Read", "PaymentType:Update", "PaymentType:Delete", "PaymentType:Export", "PaymentType:OwnRead", "PaymentType:OwnUpdate", "PaymentType:OwnDelete", "PaymentType:OwnExport",
				"AdjustmentEntry:Read", "AdjustmentEntry:Create", "AdjustmentEntry:Export",
				// Approvals
				"Approvals:Read", "ApprovalsEndBatch:Read", "ApprovalsEndBatch:Update",
				"ApprovalsBlotterView:Read", "ApprovalsBlotterView:Update", "ApprovalsUser:Read", "ApprovalsUser:Update", "ApprovalsMemberProfile:Read", "ApprovalsMemberProfile:Update",
				"ApprovalsJV:Read", "ApprovalsJVDraft:Read", "ApprovalsJVPrinted:Read", "ApprovalsJVPrinted:Update", "ApprovalsJVApproved:Read", "ApprovalsJVApproved:Update", "ApprovalsJVReleased:Read", "ApprovalsJVReleased:Update",
				"ApprovalsCashVoucher:Read", "ApprovalsCashVoucherDraft:Read", "ApprovalsCashVoucherPrinted:Read", "ApprovalsCashVoucherPrinted:Update", "ApprovalsCashVoucherApproved:Read", "ApprovalsCashVoucherApproved:Update", "ApprovalsCashVoucherReleased:Read", "ApprovalsCashVoucherReleased:Update",
				"ApprovalsLoan:Read", "ApprovalsLoanDraft:Read", "ApprovalsLoanPrinted:Read", "ApprovalsLoanPrinted:Update", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update", "ApprovalsLoanReleased:Read", "ApprovalsLoanReleased:Update",
				// Vouchers & Ledgers (Strictly mapped to TS exclusions)
				"JournalVoucher:Create", "JournalVoucher:Read", "JournalVoucher:Update", "JournalVoucher:Export", "JournalVoucher:OwnRead", "JournalVoucher:OwnUpdate", "JournalVoucher:OwnExport",
				"CashCheckVoucher:Create", "CashCheckVoucher:Read", "CashCheckVoucher:Update", "CashCheckVoucher:Export", "CashCheckVoucher:OwnUpdate", "CashCheckVoucher:OwnExport",
				"GeneralLedger:Read", "DisbursementTransaction:Create",
				"FSDefinition:Create", "FSDefinition:Read", "FSDefinition:Update", "FSDefinition:Delete", "FSDefinition:Export",
				"GLDefinition:Create", "GLDefinition:Read", "GLDefinition:Update", "GLDefinition:Delete", "GLDefinition:Export",
				// Loans
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Delete", "Loan:Export", "Loan:OwnRead", "Loan:OwnUpdate", "Loan:OwnDelete", "Loan:OwnExport",
				"LoanStatus:Create", "LoanStatus:Read", "LoanStatus:Update", "LoanStatus:Delete", "LoanStatus:Export", "LoanStatus:OwnRead", "LoanStatus:OwnUpdate", "LoanStatus:OwnDelete", "LoanStatus:OwnExport",
				"LoanPurpose:Create", "LoanPurpose:Read", "LoanPurpose:Update", "LoanPurpose:Delete", "LoanPurpose:Export", "LoanPurpose:OwnRead", "LoanPurpose:OwnUpdate", "LoanPurpose:OwnDelete", "LoanPurpose:OwnExport",
				"LoanScheme:Create", "LoanScheme:Read", "LoanScheme:Update", "LoanScheme:Delete", "LoanScheme:Export", "LoanScheme:OwnRead", "LoanScheme:OwnUpdate", "LoanScheme:OwnDelete", "LoanScheme:OwnExport",
				"LoanChargeScheme:Create", "LoanChargeScheme:Read", "LoanChargeScheme:Update", "LoanChargeScheme:Delete", "LoanChargeScheme:Export", "LoanChargeScheme:OwnRead", "LoanChargeScheme:OwnUpdate", "LoanChargeScheme:OwnDelete", "LoanChargeScheme:OwnExport",
				"Collateral:Create", "Collateral:Read", "Collateral:Update", "Collateral:Delete", "Collateral:Export", "Collateral:OwnRead", "Collateral:OwnUpdate", "Collateral:OwnDelete", "Collateral:OwnExport",
				// Members (Strictly mapped to TS exclusions)
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export", "MemberProfile:OwnUpdate", "MemberProfile:OwnExport",
				"MemberProfileClose:Create", "MemberAccountingLedger:Read",
				"MemberType:Create", "MemberType:Read", "MemberType:Update", "MemberType:Delete", "MemberType:Export", "MemberType:OwnRead", "MemberType:OwnUpdate", "MemberType:OwnDelete", "MemberType:OwnExport",
				"MemberGroup:Create", "MemberGroup:Read", "MemberGroup:Update", "MemberGroup:Delete", "MemberGroup:Export", "MemberGroup:OwnRead", "MemberGroup:OwnUpdate", "MemberGroup:OwnDelete", "MemberGroup:OwnExport",
				// Employees & Settings
				"Employee:Create", "Employee:Read", "Employee:Update", "Employee:Delete", "Employee:Export", "Employee:OwnRead", "Employee:OwnUpdate", "Employee:OwnDelete", "Employee:OwnExport",
				"EmployeePermission:Update", "EmployeePermission:OwnUpdate",
				"Company:Create", "Company:Read", "Company:Update", "Company:Delete", "Company:Export", "Company:OwnRead", "Company:OwnUpdate", "Company:OwnDelete", "Company:OwnExport",
				"Branch:Create", "Branch:Read", "Branch:Update", "Branch:Delete", "Branch:Export", "Branch:OwnRead", "Branch:OwnUpdate", "Branch:OwnDelete", "Branch:OwnExport",
				"BranchSettings:Create", "BranchSettings:Read", "BranchSettings:Update", "BranchSettings:Delete", "BranchSettings:Export",
				"PermissionTemplate:Read", "User:Read", "ApiDoc:Read", "ApiKeyGen:Create",
				"Feed:Create", "Feed:Read", "Feed:Update", "Feed:Delete", "Feed:Export", "Feed:OwnRead", "Feed:OwnUpdate", "Feed:OwnDelete", "Feed:OwnExport",
				"FeedComment:Create", "FeedComment:Delete", "FeedComment:OwnDelete",
			},
		},
		{
			Name:        "IT / System Admin",
			Description: "Technical maintenance, API management, and developer documentation.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "Company:Read", "Company:Update", "Branch:Read", "BranchSettings:Read", "BranchSettings:Update",
				"ApiDoc:Read", "ApiKeyGen:Create", "User:Read", "PermissionTemplate:Read",
				"TagTemplate:Create", "TagTemplate:Read", "TagTemplate:Update",
				"Feed:Create", "Feed:Read", "Feed:Update", "Feed:Delete", "FeedComment:Delete",
			},
		},

		// ==========================================
		// 2. MANAGEMENT
		// ==========================================
		{
			Name:        "Branch Manager",
			Description: "Manage branch operations, oversee approvals and member profiles.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "Branch:Read", "BranchSettings:Read", "BranchSettings:Update",
				"MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update", "MemberProfile:Export",
				"Account:Read", "AccountTransaction:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "Loan:Update", "Loan:Export",
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				"Approvals:Read", "ApprovalsEndBatch:Read", "ApprovalsEndBatch:Update",
				"ApprovalsLoan:Read", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update",
				"ApprovalsCashVoucher:Read", "ApprovalsCashVoucherApproved:Read", "ApprovalsCashVoucherApproved:Update",
				"Employee:Read", "EmployeeFootstep:Read",
				"Bank:Read", "BillsAndCoins:Read", "CashCount:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "HR Manager",
			Description: "Manages employee records, timesheets, and holidays.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "Employee:Create", "Employee:Read", "Employee:Update", "Employee:Export",
				"EmployeeSettings:Read", "EmployeePermission:Update",
				"Timesheet:Read", "Timesheet:Update", "Timesheet:Export",
				"Holiday:Create", "Holiday:Read", "Holiday:Update", "Holiday:Delete",
				"MemberGender:Read", "MemberOccupation:Read",
				"Company:Read", "Branch:Read", "Area:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},

		// ==========================================
		// 3. FINANCE & ACCOUNTING
		// ==========================================
		{
			Name:        "Chief Accountant",
			Description: "Full access to the General Ledger, FS Definitions, and adjustments.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "JournalVoucher:Read", "JournalVoucher:Create", "JournalVoucher:Update", "JournalVoucher:Export",
				"GeneralLedger:Read", "FSDefinition:Read", "FSDefinition:Create", "FSDefinition:Update", "FSDefinition:Delete",
				"GLDefinition:Read", "GLDefinition:Create", "GLDefinition:Update", "GLDefinition:Delete",
				"AdjustmentEntry:Read", "AdjustmentEntry:Create", "AdjustmentEntry:Export",
				"AccountClassification:Create", "AccountClassification:Read", "AccountClassification:Update",
				"ApprovalsJV:Read", "ApprovalsJVDraft:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read", "MyGeneralLedger:Read",
			},
		},
		{
			Name:        "Treasury / Cashier",
			Description: "Handles payouts, cash positions, check remittances, and online transfers.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "CashCheckVoucher:Create", "CashCheckVoucher:Read", "CashCheckVoucher:Update", "CashCheckVoucher:Export",
				"DisburesmentType:Read", "DisbursementTransaction:Create",
				"CheckRemittance:Create", "CheckRemittance:Read", "CheckRemittance:Update",
				"OnlineRemittance:Create", "OnlineRemittance:Read", "OnlineRemittance:Update",
				"Bank:Read", "Bank:Export", "BillsAndCoins:Read", "CashCount:Read",
				"TransactionBatch:Read", "TransactionBatch:Export",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read", "MyDisbursements:Read",
			},
		},

		// ==========================================
		// 4. LOANS & CREDIT
		// ==========================================
		{
			Name:        "Credit Approver",
			Description: "Evaluate and approve loan requests and disbursements.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "LoanStatus:Read", "LoanPurpose:Read", "Collateral:Read",
				"Approvals:Read", "ApprovalsLoan:Read", "ApprovalsLoanApproved:Read", "ApprovalsLoanApproved:Update",
				"ApprovalsLoanReleased:Read", "ApprovalsLoanReleased:Update",
				"DisbursementTransaction:Create",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Loan Officer",
			Description: "Process loan applications, collateral, and schemes.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Create", "Loan:Read", "Loan:Update", "Loan:Export",
				"LoanStatus:Read", "LoanPurpose:Read", "Collateral:Create", "Collateral:Read", "Collateral:Update",
				"LoanScheme:Read", "LoanChargeScheme:Read", "Account:Read", "AccountTransaction:Read",
				"ApprovalsLoanDraft:Read", "ApprovalsLoanPrinted:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "Collections Officer",
			Description: "Focused on past-due accounts, repayment tracking, and collection batches.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "MemberAccountingLedger:Read",
				"Loan:Read", "Loan:Update", "Loan:Export", "LoanStatus:Read", "LoanTag:Read", "LoanTag:Update",
				"Account:Read", "AccountTransaction:Read",
				"TransactionBatch:Create", "TransactionBatch:Read", "Transactions:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},

		// ==========================================
		// 5. TELLER & OPERATIONS
		// ==========================================
		{
			Name:        "Head Teller",
			Description: "Oversees tellers, manages cash vaults, and handles end-of-batch.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "Account:Read", "AccountTransaction:Read",
				"Transaction:Read", "Transaction:Create", "Transaction:Update",
				"QuickDeposit:Read", "QuickDeposit:Create", "QuickDeposit:Update",
				"QuickWithdraw:Read", "QuickWithdraw:Create", "QuickWithdraw:Update",
				"CashCount:Read", "BillsAndCoins:Read", "BillsAndCoins:Update",
				"TransactionBatch:Create", "TransactionBatch:Read", "TransactionBatch:Update", "TransactionBatch:Export",
				"ApprovalsEndBatch:Read", "ApprovalsEndBatch:Update",
				"TimeInOut:Create", "TimeInOut:Read", "MyTimesheet:Read", "MySettings:Read", "MySettings:Update",
			},
		},
		{
			Name:        "Teller",
			Description: "Standard day-to-day deposits, withdrawals, and timesheets.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "Account:Read",
				"Transaction:Create", "Transaction:Read",
				"QuickDeposit:Create", "QuickDeposit:Read",
				"QuickWithdraw:Create", "QuickWithdraw:Read",
				"BillsAndCoins:Read", "CashCount:Read",
				"TimeInOut:Create", "TimeInOut:Read", "TimeInOut:Update", "TimeInOut:OwnUpdate",
				"MyTimesheet:Read", "MySettings:Read", "MySettings:Update",
			},
		},
		{
			Name:        "Member Services (CSR)",
			Description: "Handle member onboarding, group management, and basic inquiries.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Create", "MemberProfile:Read", "MemberProfile:Update",
				"MemberProfileFileMediaUpload:Create", "MemberProfileFileMediaUpload:Read",
				"MemberType:Read", "MemberGroup:Read", "MemberGroup:Update",
				"MemberCenter:Read", "MemberCenter:Update", "MemberDepartment:Read", "MemberDepartment:Update",
				"MemberGender:Read", "MemberOccupation:Read", "MemberClassification:Read",
				"Account:Create", "Account:Read", "AccountClassification:Read", "AccountCategory:Read", "AccountTransaction:Read",
				"InvitationCode:Create", "InvitationCode:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},

		// ==========================================
		// 6. AUDIT & COMPLIANCE
		// ==========================================
		{
			Name:        "Internal Auditor",
			Description: "Broad read-only access to all financial, member, and employee logs.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "MemberProfile:Read", "MemberProfile:Export", "MemberProfileFileArchives:Read", "MemberAccountingLedger:Read",
				"Account:Read", "Account:Export", "AccountTransaction:Read", "AccountTag:Read",
				"Loan:Read", "Loan:Export", "Collateral:Read", "LoanTag:Read",
				"JournalVoucher:Read", "GeneralLedger:Read", "AdjustmentEntry:Read",
				"TransactionBatch:Read", "TransactionBatchHistory:Read", "Transactions:Read",
				"Approvals:Read", "ApprovalsEndBatch:Read",
				"EmployeeFootstep:Read", "Footstep:Read", "Company:Read", "BranchSettings:Read",
				"MySettings:Read", "MySettings:Update", "MyTimesheet:Read",
			},
		},
		{
			Name:        "External Auditor (Viewer)",
			Description: "Strictly read-only access to core financial data for 3rd party auditing firms.",
			Permissions: pq.StringArray{
				"Dashboard:Read", "GeneralLedger:Read", "JournalVoucher:Read", "AdjustmentEntry:Read",
				"Loan:Read", "MemberProfile:Read", "TransactionBatch:Read", "TransactionBatchHistory:Read",
				"Bank:Read", "FSDefinition:Read", "GLDefinition:Read",
			},
		},

		// ==========================================
		// 7. GENERAL STAFF
		// ==========================================
		{
			Name:        "Basic Employee",
			Description: "Limited access for general staff (Timesheets, feeds, personal settings).",
			Permissions: pq.StringArray{
				"Dashboard:Read",
				"TimeInOut:Create", "TimeInOut:Read", "TimeInOut:Update", "TimeInOut:OwnUpdate",
				"MyTimesheet:Read", "MySettings:Read", "MySettings:Update", "MyGeneralLedger:Read",
				"AllMyFootsteps:Read", "MyBranchFootsteps:Read", "Holiday:Read",
				"Feed:Read", "FeedComment:Create", "FeedComment:OwnDelete",
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
