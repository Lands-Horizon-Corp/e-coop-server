package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func TagTemplateManager(service *horizon.HorizonService) *registry.Registry[types.TagTemplate, types.TagTemplateResponse, types.TagTemplateRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.TagTemplate, types.TagTemplateResponse, types.TagTemplateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TagTemplate) *types.TagTemplateResponse {
			if data == nil {
				return nil
			}
			return &types.TagTemplateResponse{
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
				Name:           data.Name,
				Description:    data.Description,
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},

		Created: func(data *types.TagTemplate) registry.Topics {
			return []string{
				"tag_template.create",
				fmt.Sprintf("tag_template.create.%s", data.ID),
				fmt.Sprintf("tag_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.TagTemplate) registry.Topics {
			return []string{
				"tag_template.update",
				fmt.Sprintf("tag_template.update.%s", data.ID),
				fmt.Sprintf("tag_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.TagTemplate) registry.Topics {
			return []string{
				"tag_template.delete",
				fmt.Sprintf("tag_template.delete.%s", data.ID),
				fmt.Sprintf("tag_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func tagTemplateSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	tagTemplates := []*types.TagTemplate{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Pending",
			Description:    "Transaction or record is pending approval or processing",
			Category:       types.TagCategoryStatus,
			Color:          "#F59E0B", // Yellow
			Icon:           "Clock",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Approved",
			Description:    "Transaction or record has been approved",
			Category:       types.TagCategoryStatus,
			Color:          "#10B981", // Green
			Icon:           "Badge Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Rejected",
			Description:    "Transaction or record has been rejected",
			Category:       types.TagCategoryStatus,
			Color:          "#EF4444", // Red
			Icon:           "Badge Minus",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Aborted",
			Description:    "Transaction or record was aborted during processing",
			Category:       types.TagCategoryStatus,
			Color:          "#DC2626", // Dark Red
			Icon:           "Stop",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Completed",
			Description:    "Transaction or record has been completed successfully",
			Category:       types.TagCategoryStatus,
			Color:          "#059669", // Dark Green
			Icon:           "Check Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "In Progress",
			Description:    "Transaction or record is currently being processed",
			Category:       types.TagCategoryStatus,
			Color:          "#3B82F6", // Blue
			Icon:           "Loading Spinner",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Warning",
			Description:    "Transaction or record requires attention or has warnings",
			Category:       types.TagCategoryAlert,
			Color:          "#F59E0B", // Yellow/Orange
			Icon:           "Warning",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "High Priority",
			Description:    "High priority transaction or record requiring urgent attention",
			Category:       types.TagCategoryPriority,
			Color:          "#DC2626", // Red
			Icon:           "Badge Exclamation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Medium Priority",
			Description:    "Medium priority transaction or record",
			Category:       types.TagCategoryPriority,
			Color:          "#F59E0B", // Orange
			Icon:           "Badge Question",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Low Priority",
			Description:    "Low priority transaction or record",
			Category:       types.TagCategoryPriority,
			Color:          "#6B7280", // Gray
			Icon:           "Info",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cash Transaction",
			Description:    "Transaction involving cash payments or receipts",
			Category:       types.TagCategoryTransactionType,
			Color:          "#10B981", // Green
			Icon:           "Money Stack",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bank Transfer",
			Description:    "Transaction processed through bank transfer",
			Category:       types.TagCategoryTransactionType,
			Color:          "#3B82F6", // Blue
			Icon:           "Bank",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Credit Card",
			Description:    "Transaction processed via credit card",
			Category:       types.TagCategoryTransactionType,
			Color:          "#8B5CF6", // Purple
			Icon:           "Credit Card",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Online Payment",
			Description:    "Transaction processed through online payment systems",
			Category:       types.TagCategoryTransactionType,
			Color:          "#06B6D4", // Cyan
			Icon:           "Online Payment",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Asset",
			Description:    "Account classified as an asset",
			Category:       types.TagCategoryAccountType,
			Color:          "#10B981", // Green
			Icon:           "Money Bag",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Liability",
			Description:    "Account classified as a liability",
			Category:       types.TagCategoryAccountType,
			Color:          "#EF4444", // Red
			Icon:           "Credit Card",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Equity",
			Description:    "Account classified as equity",
			Category:       types.TagCategoryAccountType,
			Color:          "#8B5CF6", // Purple
			Icon:           "Pie Chart",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Revenue",
			Description:    "Account classified as revenue or income",
			Category:       types.TagCategoryAccountType,
			Color:          "#059669", // Dark Green
			Icon:           "Trend Up",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Expense",
			Description:    "Account classified as an expense",
			Category:       types.TagCategoryAccountType,
			Color:          "#DC2626", // Dark Red
			Icon:           "Trend Down",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Recurring",
			Description:    "Recurring transaction or scheduled entry",
			Category:       types.TagCategorySpecial,
			Color:          "#06B6D4", // Cyan
			Icon:           "Refresh",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Adjustment",
			Description:    "Adjustment entry for corrections or modifications",
			Category:       types.TagCategorySpecial,
			Color:          "#F59E0B", // Orange
			Icon:           "Adjust",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Year End",
			Description:    "Year-end closing or adjustment entries",
			Category:       types.TagCategorySpecial,
			Color:          "#6366F1", // Indigo
			Icon:           "Calendar",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Audit",
			Description:    "Entry related to audit requirements or adjustments",
			Category:       types.TagCategorySpecial,
			Color:          "#7C2D12", // Brown
			Icon:           "Shield Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Tax Related",
			Description:    "Transaction or entry related to tax calculations or payments",
			Category:       types.TagCategorySpecial,
			Color:          "#991B1B", // Dark Red
			Icon:           "Receipt",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Loan Related",
			Description:    "Transaction or entry related to loan processing or payments",
			Category:       types.TagCategoryLoan,
			Color:          "#7C2D12", // Brown
			Icon:           "Hand Coins",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Interest",
			Description:    "Transaction involving interest calculations or payments",
			Category:       types.TagCategoryCalculation,
			Color:          "#0891B2", // Dark Cyan
			Icon:           "Percent",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Fee",
			Description:    "Transaction involving service fees or charges",
			Category:       types.TagCategoryCalculation,
			Color:          "#BE185D", // Pink
			Icon:           "Price Tag",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Savings",
			Description:    "Savings account transactions for cooperative members",
			Category:       types.TagCategoryCooperative,
			Color:          "#16A34A", // Green
			Icon:           "Savings",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Share Capital",
			Description:    "Share capital contributions and transactions",
			Category:       types.TagCategoryCooperative,
			Color:          "#0F766E", // Teal
			Icon:           "Pie Chart",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Dividend Distribution",
			Description:    "Distribution of dividends to cooperative members",
			Category:       types.TagCategoryCooperative,
			Color:          "#059669", // Emerald
			Icon:           "Hand Coins",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Patronage Refund",
			Description:    "Patronage refunds based on member usage",
			Category:       types.TagCategoryCooperative,
			Color:          "#0D9488", // Teal
			Icon:           "Money Trend",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Membership Fee",
			Description:    "Membership fees and registration costs",
			Category:       types.TagCategoryCooperative,
			Color:          "#7C2D12", // Brown
			Icon:           "User Tag",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Loan",
			Description:    "Loans provided to cooperative members",
			Category:       types.TagCategoryLoan,
			Color:          "#B45309", // Amber
			Icon:           "Hand Deposit",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Agricultural Loan",
			Description:    "Specialized loans for agricultural purposes",
			Category:       types.TagCategoryLoan,
			Color:          "#65A30D", // Lime
			Icon:           "Plant Growth",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Microfinance",
			Description:    "Small loans for micro-enterprises and small businesses",
			Category:       types.TagCategoryLoan,
			Color:          "#CA8A04", // Yellow
			Icon:           "Money Bag",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Emergency Loan",
			Description:    "Emergency loans for urgent member needs",
			Category:       types.TagCategoryLoan,
			Color:          "#DC2626", // Red
			Icon:           "Shield",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Educational Loan",
			Description:    "Loans for educational expenses and tuition",
			Category:       types.TagCategoryLoan,
			Color:          "#2563EB", // Blue
			Icon:           "School",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Housing Loan",
			Description:    "Loans for housing and real estate purchases",
			Category:       types.TagCategoryLoan,
			Color:          "#7C3AED", // Violet
			Icon:           "House",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Community Fund",
			Description:    "Contributions to community development funds",
			Category:       types.TagCategoryCommunity,
			Color:          "#0891B2", // Cyan
			Icon:           "People Group",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Social Responsibility",
			Description:    "Corporate social responsibility initiatives",
			Category:       types.TagCategoryCommunity,
			Color:          "#059669", // Emerald
			Icon:           "Hand Shake Heart",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Charity Donation",
			Description:    "Charitable donations and community support",
			Category:       types.TagCategoryCommunity,
			Color:          "#DB2777", // Pink
			Icon:           "Hands Helping",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Educational Support",
			Description:    "Educational assistance and scholarship programs",
			Category:       types.TagCategoryCommunity,
			Color:          "#2563EB", // Blue
			Icon:           "Graduation Cap",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Health Insurance",
			Description:    "Health insurance contributions and claims",
			Category:       types.TagCategoryInsurance,
			Color:          "#DC2626", // Red
			Icon:           "Shield Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Life Insurance",
			Description:    "Life insurance premiums and benefits",
			Category:       types.TagCategoryInsurance,
			Color:          "#7C2D12", // Brown
			Icon:           "Shield",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Crop Insurance",
			Description:    "Agricultural crop insurance for farmers",
			Category:       types.TagCategoryInsurance,
			Color:          "#65A30D", // Lime
			Icon:           "Plant Growth",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Board Resolution",
			Description:    "Transactions requiring board resolution approval",
			Category:       types.TagCategoryGovernance,
			Color:          "#6366F1", // Indigo
			Icon:           "Users 3",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "AGM Related",
			Description:    "Annual General Meeting related transactions",
			Category:       types.TagCategoryGovernance,
			Color:          "#7C3AED", // Violet
			Icon:           "Calendar Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Regulatory Compliance",
			Description:    "Compliance with regulatory requirements",
			Category:       types.TagCategoryGovernance,
			Color:          "#991B1B", // Red
			Icon:           "Shield Exclamation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Reserve Fund",
			Description:    "Transactions related to reserve fund allocations",
			Category:       types.TagCategoryReserves,
			Color:          "#0F766E", // Teal
			Icon:           "Wallet",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Capital Reserve",
			Description:    "Capital reserve fund transactions",
			Category:       types.TagCategoryReserves,
			Color:          "#059669", // Emerald
			Icon:           "Money Stack",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bad Debt Provision",
			Description:    "Provision for bad debts and loan losses",
			Category:       types.TagCategoryReserves,
			Color:          "#DC2626", // Red
			Icon:           "Warning Circle",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Mobile Banking",
			Description:    "Transactions processed through mobile banking",
			Category:       types.TagCategoryDigital,
			Color:          "#0891B2", // Cyan
			Icon:           "Smartphone",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Online Banking",
			Description:    "Internet banking transactions and services",
			Category:       types.TagCategoryDigital,
			Color:          "#2563EB", // Blue
			Icon:           "Globe",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "ATM Transaction",
			Description:    "Automated Teller Machine transactions",
			Category:       types.TagCategoryDigital,
			Color:          "#7C3AED", // Violet
			Icon:           "Monitor",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "QR Payment",
			Description:    "QR code based payment transactions",
			Category:       types.TagCategoryDigital,
			Color:          "#059669", // Emerald
			Icon:           "QR Code",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "New Member",
			Description:    "New member registration and onboarding",
			Category:       types.TagCategoryMembership,
			Color:          "#16A34A", // Green
			Icon:           "User Plus",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Withdrawal",
			Description:    "Member withdrawal from cooperative",
			Category:       types.TagCategoryMembership,
			Color:          "#DC2626", // Red
			Icon:           "Exit Door",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Death Benefit",
			Description:    "Death benefits and insurance claims",
			Category:       types.TagCategoryInsurance,
			Color:          "#374151", // Gray
			Icon:           "Shield Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Loan Collateral",
			Description:    "Collateral security for loan transactions",
			Category:       types.TagCategorySecurity,
			Color:          "#92400E", // Yellow
			Icon:           "Shield Lock",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Guarantee",
			Description:    "Guarantee and surety related transactions",
			Category:       types.TagCategorySecurity,
			Color:          "#0F766E", // Teal
			Icon:           "User Shield",
		},
	}

	for _, data := range tagTemplates {
		if err := TagTemplateManager(service).CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed tag template %s", data.Name)
		}
	}
	return nil
}

func TagTemplateCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.TagTemplate, error) {
	return TagTemplateManager(service).Find(context, &types.TagTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
