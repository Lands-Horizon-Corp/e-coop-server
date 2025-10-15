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
	LoanStatus struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_status"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_status"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Icon        string `gorm:"type:varchar(255)"`
		Color       string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	LoanStatusResponse struct {
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
		Icon           string                `json:"icon"`
		Color          string                `json:"color"`
		Description    string                `json:"description"`
	}

	LoanStatusRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Icon        string `json:"icon,omitempty"`
		Color       string `json:"color,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func (m *ModelCore) LoanStatus() {
	m.Migration = append(m.Migration, &LoanStatus{})
	m.LoanStatusManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanStatus, LoanStatusResponse, LoanStatusRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanStatus) *LoanStatusResponse {
			if data == nil {
				return nil
			}
			return &LoanStatusResponse{
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
				Icon:           data.Icon,
				Color:          data.Color,
				Description:    data.Description,
			}
		},

		Created: func(data *LoanStatus) []string {
			return []string{
				"loan_status.create",
				fmt.Sprintf("loan_status.create.%s", data.ID),
				fmt.Sprintf("loan_status.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanStatus) []string {
			return []string{
				"loan_status.update",
				fmt.Sprintf("loan_status.update.%s", data.ID),
				fmt.Sprintf("loan_status.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanStatus) []string {
			return []string{
				"loan_status.delete",
				fmt.Sprintf("loan_status.delete.%s", data.ID),
				fmt.Sprintf("loan_status.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) LoanStatusSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	loanStatuses := []*LoanStatus{
		// Application Phase Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Application Submitted",
			Description:    "Loan application has been submitted and is awaiting review",
			Color:          "#3B82F6", // Blue
			Icon:           "File Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Under Review",
			Description:    "Loan application is currently being reviewed by loan officers",
			Color:          "#F59E0B", // Orange
			Icon:           "Eye View",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Documentation Required",
			Description:    "Additional documentation is required to process the loan application",
			Color:          "#EF4444", // Red
			Icon:           "Document File Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Credit Check",
			Description:    "Credit verification and background check is in progress",
			Color:          "#8B5CF6", // Purple
			Icon:           "Shield Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Collateral Verification",
			Description:    "Collateral assessment and verification is being conducted",
			Color:          "#06B6D4", // Cyan
			Icon:           "Shield Lock",
		},
		// Approval Phase Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Pending Approval",
			Description:    "Loan application is pending final approval from loan committee",
			Color:          "#F59E0B", // Orange
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
			Description:    "Loan application has been approved and ready for disbursement",
			Color:          "#10B981", // Green
			Icon:           "Badge Check Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Conditionally Approved",
			Description:    "Loan approved with specific conditions that must be met",
			Color:          "#F59E0B", // Orange
			Icon:           "Badge Question Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Rejected",
			Description:    "Loan application has been rejected",
			Color:          "#EF4444", // Red
			Icon:           "Badge Minus Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Withdrawn",
			Description:    "Loan application has been withdrawn by the applicant",
			Color:          "#6B7280", // Gray
			Icon:           "Exit Door",
		},
		// Disbursement Phase Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Ready for Disbursement",
			Description:    "Approved loan is ready for fund disbursement",
			Color:          "#059669", // Dark Green
			Icon:           "Money Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Partially Disbursed",
			Description:    "Loan funds have been partially disbursed",
			Color:          "#0891B2", // Cyan
			Icon:           "Coins Stack",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Fully Disbursed",
			Description:    "All loan funds have been disbursed to the borrower",
			Color:          "#10B981", // Green
			Icon:           "Money Stack",
		},
		// Active Loan Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Active",
			Description:    "Loan is active and payments are current",
			Color:          "#10B981", // Green
			Icon:           "Check Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Grace Period",
			Description:    "Loan is in grace period before repayment starts",
			Color:          "#06B6D4", // Cyan
			Icon:           "Calendar",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Current",
			Description:    "Loan payments are up to date and current",
			Color:          "#10B981", // Green
			Icon:           "Trend Up",
		},
		// Delinquency Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Past Due",
			Description:    "Loan payment is past due but within acceptable limits",
			Color:          "#F59E0B", // Orange
			Icon:           "Warning",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "30 Days Delinquent",
			Description:    "Loan payment is 30 days past due",
			Color:          "#F59E0B", // Orange
			Icon:           "Warning Circle",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "60 Days Delinquent",
			Description:    "Loan payment is 60 days past due",
			Color:          "#DC2626", // Red
			Icon:           "Error Exclamation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "90 Days Delinquent",
			Description:    "Loan payment is 90 days past due - serious delinquency",
			Color:          "#991B1B", // Dark Red
			Icon:           "Badge Exclamation Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Default",
			Description:    "Loan is in default status due to non-payment",
			Color:          "#7F1D1D", // Very Dark Red
			Icon:           "Error",
		},
		// Special Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Restructured",
			Description:    "Loan terms have been restructured due to borrower circumstances",
			Color:          "#7C3AED", // Violet
			Icon:           "Refresh",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Rescheduled",
			Description:    "Loan payment schedule has been rescheduled",
			Color:          "#8B5CF6", // Purple
			Icon:           "Calendar Check",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Under Legal Action",
			Description:    "Legal proceedings have been initiated for loan recovery",
			Color:          "#450A0A", // Very Dark Red
			Icon:           "Shield Exclamation",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Foreclosure",
			Description:    "Collateral foreclosure process has been initiated",
			Color:          "#450A0A", // Very Dark Red
			Icon:           "House Lock",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Written Off",
			Description:    "Loan has been written off as bad debt",
			Color:          "#374151", // Dark Gray
			Icon:           "Trash",
		},
		// Completion Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Paid in Full",
			Description:    "Loan has been completely paid off",
			Color:          "#059669", // Dark Green
			Icon:           "Badge Check Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Early Settlement",
			Description:    "Loan has been settled early before maturity",
			Color:          "#10B981", // Green
			Icon:           "Thumbs Up",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Closed",
			Description:    "Loan account has been closed and finalized",
			Color:          "#6B7280", // Gray
			Icon:           "Archive",
		},
		// Cooperative Specific Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Member Priority",
			Description:    "High priority loan for cooperative member in good standing",
			Color:          "#0F766E", // Teal
			Icon:           "Crown",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Emergency Assistance",
			Description:    "Emergency loan assistance for member in crisis",
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
			Name:           "Agricultural Season",
			Description:    "Seasonal agricultural loan for farming activities",
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
			Name:           "Livelihood Support",
			Description:    "Loan for livelihood and income generation projects",
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
			Name:           "Educational Financing",
			Description:    "Educational loan for member's children or family",
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
			Name:           "Medical Emergency",
			Description:    "Emergency medical loan for health-related expenses",
			Color:          "#DC2626", // Red
			Icon:           "Heart Break Fill",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cooperative Development",
			Description:    "Loan for cooperative business development and expansion",
			Color:          "#7C3AED", // Violet
			Icon:           "Building Gear",
		},
		// Role and Process Related Statuses
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Department Approval",
			Description:    "Awaiting approval from specific department or division",
			Color:          "#6366F1", // Indigo
			Icon:           "Building Cog",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Comaker Required",
			Description:    "Loan requires comaker or guarantor before approval",
			Color:          "#7C3AED", // Violet
			Icon:           "Users Add",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Comaker Verified",
			Description:    "Comaker has been verified and approved",
			Color:          "#10B981", // Green
			Icon:           "User Shield",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Collector Assigned",
			Description:    "Loan collector has been assigned for follow-up",
			Color:          "#F59E0B", // Orange
			Icon:           "User Cog",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Collection in Progress",
			Description:    "Loan collection activities are in progress",
			Color:          "#DC2626", // Red
			Icon:           "Running",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Payee Verification",
			Description:    "Verification of loan payee or beneficiary details",
			Color:          "#8B5CF6", // Purple
			Icon:           "ID Card",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Payee Confirmed",
			Description:    "Loan payee has been confirmed and verified",
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
			Name:           "Payer Setup",
			Description:    "Setting up payer information and payment details",
			Color:          "#06B6D4", // Cyan
			Icon:           "Account Setup",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Payer Verified",
			Description:    "Loan payer details have been verified and confirmed",
			Color:          "#059669", // Dark Green
			Icon:           "Verified Patch",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Department Review",
			Description:    "Under review by designated department or unit",
			Color:          "#7C2D12", // Brown
			Icon:           "Building Branch",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Comaker Default",
			Description:    "Comaker has defaulted on guarantor obligations",
			Color:          "#991B1B", // Dark Red
			Icon:           "User Lock",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Collection Hold",
			Description:    "Collection activities temporarily on hold",
			Color:          "#6B7280", // Gray
			Icon:           "Hand",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Payee Change Request",
			Description:    "Request to change loan payee or beneficiary",
			Color:          "#F59E0B", // Orange
			Icon:           "Swap Arrow",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Payer Change Request",
			Description:    "Request to change loan payer or payment source",
			Color:          "#F59E0B", // Orange
			Icon:           "Replace",
		},
	}

	for _, data := range loanStatuses {
		if err := m.LoanStatusManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed loan status %s", data.Name)
		}
	}
	return nil
}

func (m *ModelCore) LoanStatusCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanStatus, error) {
	return m.LoanStatusManager.Find(context, &LoanStatus{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
