package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// CancelledCashCheckVoucher represents the CancelledCashCheckVoucher model.
	CancelledCashCheckVoucher struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cancelled_cash_check_voucher" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cancelled_cash_check_voucher" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CheckNumber string    `gorm:"type:varchar(255);not null" json:"check_number"`
		EntryDate   time.Time `gorm:"not null" json:"entry_date"`
		Description string    `gorm:"type:text" json:"description"`
	}

	// CancelledCashCheckVoucherResponse represents the response structure for cancelledcashcheckvoucher data

	// CancelledCashCheckVoucherResponse represents the response structure for CancelledCashCheckVoucher.
	CancelledCashCheckVoucherResponse struct {
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
		CheckNumber    string                `json:"check_number"`
		EntryDate      string                `json:"entry_date"`
		Description    string                `json:"description"`
	}

	// CancelledCashCheckVoucherRequest represents the request structure for creating/updating cancelledcashcheckvoucher

	// CancelledCashCheckVoucherRequest represents the request structure for CancelledCashCheckVoucher.
	CancelledCashCheckVoucherRequest struct {
		CheckNumber string    `json:"check_number" validate:"required,min=1,max=255"`
		EntryDate   time.Time `json:"entry_date" validate:"required"`
		Description string    `json:"description,omitempty"`
	}
)

func (m *ModelCore) cancelledCashCheckVoucher() {
	m.Migration = append(m.Migration, &CancelledCashCheckVoucher{})
	m.CancelledCashCheckVoucherManager = services.NewRepository(services.RepositoryParams[
		CancelledCashCheckVoucher, CancelledCashCheckVoucherResponse, CancelledCashCheckVoucherRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *CancelledCashCheckVoucher) *CancelledCashCheckVoucherResponse {
			if data == nil {
				return nil
			}
			return &CancelledCashCheckVoucherResponse{
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
				CheckNumber:    data.CheckNumber,
				EntryDate:      data.EntryDate.Format(time.RFC3339),
				Description:    data.Description,
			}
		},
		Created: func(data *CancelledCashCheckVoucher) []string {
			return []string{
				"cancelled_cash_check_voucher.create",
				fmt.Sprintf("cancelled_cash_check_voucher.create.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CancelledCashCheckVoucher) []string {
			return []string{
				"cancelled_cash_check_voucher.update",
				fmt.Sprintf("cancelled_cash_check_voucher.update.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CancelledCashCheckVoucher) []string {
			return []string{
				"cancelled_cash_check_voucher.delete",
				fmt.Sprintf("cancelled_cash_check_voucher.delete.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// CancelledCashCheckVoucherCurrentBranch retrieves all cancelled cash check vouchers for the specified organization and branch
func (m *ModelCore) CancelledCashCheckVoucherCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*CancelledCashCheckVoucher, error) {
	return m.CancelledCashCheckVoucherManager.Find(context, &CancelledCashCheckVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
