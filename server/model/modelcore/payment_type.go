package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// TypeOfPaymentType represents the different types of payment methods available
type TypeOfPaymentType string

const (
	// PaymentTypeCash represents cash payment type
	PaymentTypeCash TypeOfPaymentType = "cash"
	// PaymentTypeCheck represents check payment type
	PaymentTypeCheck TypeOfPaymentType = "check"
	// PaymentTypeOnline represents online payment type
	PaymentTypeOnline TypeOfPaymentType = "online"
	// PaymentTypeAdjustment represents adjustment payment type
	PaymentTypeAdjustment TypeOfPaymentType = "adjustment"
)

type (
	// PaymentType represents a payment method in the system
	PaymentType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_payment_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_payment_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name         string            `gorm:"type:varchar(255);not null"`
		Description  string            `gorm:"type:text"`
		NumberOfDays int               `gorm:"type:int"`
		Type         TypeOfPaymentType `gorm:"type:varchar(20)"`
	}

	// PaymentTypeResponse represents the JSON response structure for payment type data
	PaymentTypeResponse struct {
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
		NumberOfDays   int                   `json:"number_of_days"`
		Type           TypeOfPaymentType     `json:"type"`
	}

	// PaymentTypeRequest represents the request payload for creating or updating payment type data
	PaymentTypeRequest struct {
		Name         string            `json:"name" validate:"required,min=1,max=255"`
		Description  string            `json:"description,omitempty"`
		NumberOfDays int               `json:"number_of_days,omitempty"`
		Type         TypeOfPaymentType `json:"type" validate:"required,oneof=cash check online adjustment"`
	}
)

// PaymentType initializes the PaymentType model and its repository manager
func (m *ModelCore) PaymentType() {
	m.Migration = append(m.Migration, &PaymentType{})
	m.PaymentTypeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		PaymentType, PaymentTypeResponse, PaymentTypeRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *PaymentType) *PaymentTypeResponse {
			if data == nil {
				return nil
			}
			return &PaymentTypeResponse{
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
				NumberOfDays:   data.NumberOfDays,
				Type:           data.Type,
			}
		},
		Created: func(data *PaymentType) []string {
			return []string{
				"payment_type.create",
				fmt.Sprintf("payment_type.create.%s", data.ID),
				fmt.Sprintf("payment_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *PaymentType) []string {
			return []string{
				"payment_type.update",
				fmt.Sprintf("payment_type.update.%s", data.ID),
				fmt.Sprintf("payment_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *PaymentType) []string {
			return []string{
				"payment_type.delete",
				fmt.Sprintf("payment_type.delete.%s", data.ID),
				fmt.Sprintf("payment_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// PaymentTypeSeed seeds default payment types for a new organization branch
func (m *ModelCore) PaymentTypeSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	cashOnHandPayment := &PaymentType{
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedByID:    userID,
		UpdatedByID:    userID,
		OrganizationID: organizationID,
		BranchID:       branchID,
		Name:           "Cash On Hand",
		Description:    "Cash available at the branch for immediate use.",
		NumberOfDays:   0,
		Type:           PaymentTypeCash,
	}
	if err := m.PaymentTypeManager.CreateWithTx(context, tx, cashOnHandPayment); err != nil {
		return eris.Wrap(err, "failed to seed cash on hand payment type")
	}
	userOrganization, err := m.UserOrganizationManager.FindOne(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find user organization for seeding payment types")
	}
	userOrganization.SettingsPaymentTypeDefaultValueID = &cashOnHandPayment.ID
	if err := m.UserOrganizationManager.UpdateFieldsWithTx(context, tx, userOrganization.ID, userOrganization); err != nil {
		return eris.Wrap(err, "failed to update user organization with default payment type")
	}
	paymentTypes := []*PaymentType{
		// Cash types
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Forward Cash On Hand",
			Description:    "Physical cash received and forwarded for transactions.",
			NumberOfDays:   0,
			Type:           PaymentTypeCash,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Petty Cash",
			Description:    "Small amount of cash for minor expenses.",
			NumberOfDays:   0,
			Type:           PaymentTypeCash,
		},
		// Online types
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "E-Wallet",
			Description:    "Digital wallet for online payments.",
			NumberOfDays:   0,
			Type:           PaymentTypeOnline,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "E-Bank",
			Description:    "Online banking transfer.",
			NumberOfDays:   0,
			Type:           PaymentTypeOnline,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "GCash",
			Description:    "GCash mobile wallet payment.",
			NumberOfDays:   0,
			Type:           PaymentTypeOnline,
		},
		// Check/Bank types
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Cheque",
			Description:    "Payment via cheque/check.",
			NumberOfDays:   3,
			Type:           PaymentTypeCheck,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bank Transfer",
			Description:    "Direct bank-to-bank transfer.",
			NumberOfDays:   1,
			Type:           PaymentTypeCheck,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Manager's Check",
			Description:    "Bank-issued check for secure payments.",
			NumberOfDays:   2,
			Type:           PaymentTypeCheck,
		},
		// Adjustment types
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Manual Adjustment",
			Description:    "Manual adjustments for corrections and reconciliation.",
			NumberOfDays:   0,
			Type:           PaymentTypeAdjustment,
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Adjustment Entry",
			Description:    "Manual adjustments for corrections and reconciliation.",
			NumberOfDays:   0,
			Type:           PaymentTypeAdjustment,
		},
	}

	for _, data := range paymentTypes {
		if err := m.PaymentTypeManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed payment type %s", data.Name)
		}
	}
	return nil
}

// PaymentTypeCurrentBranch retrieves all payment types for the specified organization and branch
func (m *ModelCore) PaymentTypeCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*PaymentType, error) {
	return m.PaymentTypeManager.Find(context, &PaymentType{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
