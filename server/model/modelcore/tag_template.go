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

// TagCategory represents different categorization types for tags in the cooperative system
type TagCategory string

const (
	TagCategoryStatus          TagCategory = "status"
	TagCategoryAlert           TagCategory = "alert"
	TagCategoryPriority        TagCategory = "priority"
	TagCategoryTransactionType TagCategory = "transaction type"
	TagCategoryAccountType     TagCategory = "account type"
	TagCategorySpecial         TagCategory = "special"
	TagCategoryCalculation     TagCategory = "calculation"
	TagCategoryCooperative     TagCategory = "cooperative"
	TagCategoryLoan            TagCategory = "loan"
	TagCategoryCommunity       TagCategory = "community"
	TagCategoryInsurance       TagCategory = "insurance"
	TagCategoryGovernance      TagCategory = "governance"
	TagCategoryReserves        TagCategory = "reserves"
	TagCategoryDigital         TagCategory = "digital"
	TagCategoryMembership      TagCategory = "membership"
	TagCategorySecurity        TagCategory = "security"
)

type (
	// TagTemplate represents reusable tag templates for categorizing and organizing cooperative data
	TagTemplate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_tag_template"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_tag_template"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string      `gorm:"type:varchar(50)"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	// TagTemplateResponse represents the response structure for tag template data
	TagTemplateResponse struct {
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
		Category       TagCategory           `json:"category"`
		Color          string                `json:"color"`
		Icon           string                `json:"icon"`
	}

	// TagTemplateRequest represents the request structure for creating or updating tag templates
	TagTemplateRequest struct {
		Name        string      `json:"name" validate:"required,min=1,max=50"`
		Description string      `json:"description,omitempty"`
		Category    TagCategory `json:"category,omitempty"`
		Color       string      `json:"color,omitempty"`
		Icon        string      `json:"icon,omitempty"`
	}
)

// TagTemplate initializes the tag template model and its repository manager
func (m *ModelCore) tagTemplate() {
	m.Migration = append(m.Migration, &TagTemplate{})
	m.TagTemplateManager = services.NewRepository(services.RepositoryParams[
		TagTemplate, TagTemplateResponse, TagTemplateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Service: m.provider.Service,
		Resource: func(data *TagTemplate) *TagTemplateResponse {
			if data == nil {
				return nil
			}
			return &TagTemplateResponse{
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
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},

		Created: func(data *TagTemplate) []string {
			return []string{
				"tag_template.create",
				fmt.Sprintf("tag_template.create.%s", data.ID),
				fmt.Sprintf("tag_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *TagTemplate) []string {
			return []string{
				"tag_template.update",
				fmt.Sprintf("tag_template.update.%s", data.ID),
				fmt.Sprintf("tag_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *TagTemplate) []string {
			return []string{
				"tag_template.delete",
				fmt.Sprintf("tag_template.delete.%s", data.ID),
				fmt.Sprintf("tag_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// TagTemplateSeed initializes the database with default tag templates for a branch
func (m *ModelCore) tagTemplateSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	tagTemplates := []*TagTemplate{
		// Status Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Pending",
			Description:    "Transaction or record is pending approval or processing",
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryStatus,
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
			Category:       TagCategoryAlert,
			Color:          "#F59E0B", // Yellow/Orange
			Icon:           "Warning",
		},
		// Priority Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "High Priority",
			Description:    "High priority transaction or record requiring urgent attention",
			Category:       TagCategoryPriority,
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
			Category:       TagCategoryPriority,
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
			Category:       TagCategoryPriority,
			Color:          "#6B7280", // Gray
			Icon:           "Info",
		},
		// Transaction Type Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cash Transaction",
			Description:    "Transaction involving cash payments or receipts",
			Category:       TagCategoryTransactionType,
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
			Category:       TagCategoryTransactionType,
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
			Category:       TagCategoryTransactionType,
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
			Category:       TagCategoryTransactionType,
			Color:          "#06B6D4", // Cyan
			Icon:           "Online Payment",
		},
		// Account Classification Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Asset",
			Description:    "Account classified as an asset",
			Category:       TagCategoryAccountType,
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
			Category:       TagCategoryAccountType,
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
			Category:       TagCategoryAccountType,
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
			Category:       TagCategoryAccountType,
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
			Category:       TagCategoryAccountType,
			Color:          "#DC2626", // Dark Red
			Icon:           "Trend Down",
		},
		// Special Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Recurring",
			Description:    "Recurring transaction or scheduled entry",
			Category:       TagCategorySpecial,
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
			Category:       TagCategorySpecial,
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
			Category:       TagCategorySpecial,
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
			Category:       TagCategorySpecial,
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
			Category:       TagCategorySpecial,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryCalculation,
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
			Category:       TagCategoryCalculation,
			Color:          "#BE185D", // Pink
			Icon:           "Price Tag",
		},
		// Cooperative Banking Specific Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Savings",
			Description:    "Savings account transactions for cooperative members",
			Category:       TagCategoryCooperative,
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
			Category:       TagCategoryCooperative,
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
			Category:       TagCategoryCooperative,
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
			Category:       TagCategoryCooperative,
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
			Category:       TagCategoryCooperative,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryLoan,
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
			Category:       TagCategoryLoan,
			Color:          "#7C3AED", // Violet
			Icon:           "House",
		},
		// Community and Social Services Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Community Fund",
			Description:    "Contributions to community development funds",
			Category:       TagCategoryCommunity,
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
			Category:       TagCategoryCommunity,
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
			Category:       TagCategoryCommunity,
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
			Category:       TagCategoryCommunity,
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
			Category:       TagCategoryInsurance,
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
			Category:       TagCategoryInsurance,
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
			Category:       TagCategoryInsurance,
			Color:          "#65A30D", // Lime
			Icon:           "Plant Growth",
		},
		// Operational Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Board Resolution",
			Description:    "Transactions requiring board resolution approval",
			Category:       TagCategoryGovernance,
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
			Category:       TagCategoryGovernance,
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
			Category:       TagCategoryGovernance,
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
			Category:       TagCategoryReserves,
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
			Category:       TagCategoryReserves,
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
			Category:       TagCategoryReserves,
			Color:          "#DC2626", // Red
			Icon:           "Warning Circle",
		},
		// Digital Banking Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Mobile Banking",
			Description:    "Transactions processed through mobile banking",
			Category:       TagCategoryDigital,
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
			Category:       TagCategoryDigital,
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
			Category:       TagCategoryDigital,
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
			Category:       TagCategoryDigital,
			Color:          "#059669", // Emerald
			Icon:           "QR Code",
		},
		// Member Services Tags
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "New Member",
			Description:    "New member registration and onboarding",
			Category:       TagCategoryMembership,
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
			Category:       TagCategoryMembership,
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
			Category:       TagCategoryInsurance,
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
			Category:       TagCategorySecurity,
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
			Category:       TagCategorySecurity,
			Color:          "#0F766E", // Teal
			Icon:           "User Shield",
		},
	}

	for _, data := range tagTemplates {
		if err := m.TagTemplateManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed tag template %s", data.Name)
		}
	}
	return nil
}

// TagTemplateCurrentBranch retrieves all tag templates for a specific branch within an organization
func (m *ModelCore) tagTemplateCurrentbranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*TagTemplate, error) {
	return m.TagTemplateManager.Find(context, &TagTemplate{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
