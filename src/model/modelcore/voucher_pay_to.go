package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// VoucherPayTo represents a payee entity for voucher transactions within an organization
	VoucherPayTo struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_voucher_pay_to"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_voucher_pay_to"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string     `gorm:"type:varchar(255)"`
		MediaID     *uuid.UUID `gorm:"type:uuid"`
		Media       *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"media,omitempty"`
		Description string     `gorm:"type:varchar(255)"`
	}

	// VoucherPayToResponse represents the response structure for voucher payee data
	VoucherPayToResponse struct {
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
		MediaID        *uuid.UUID            `json:"media_id,omitempty"`
		Media          *MediaResponse        `json:"media,omitempty"`
		Description    string                `json:"description"`
	}

	// VoucherPayToRequest represents the request structure for creating or updating a voucher payee
	VoucherPayToRequest struct {
		Name        string     `json:"name,omitempty"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
		Description string     `json:"description,omitempty"`
	}
)

// VoucherPayTo initializes the voucher pay to repository and sets up migration
func (m *modelcore) VoucherPayTo() {
	m.Migration = append(m.Migration, &VoucherPayTo{})
	m.VoucherPayToManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		VoucherPayTo, VoucherPayToResponse, VoucherPayToRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Media",
		},
		Service: m.provider.Service,
		Resource: func(data *VoucherPayTo) *VoucherPayToResponse {
			if data == nil {
				return nil
			}
			return &VoucherPayToResponse{
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
				MediaID:        data.MediaID,
				Media:          m.MediaManager.ToModel(data.Media),
				Description:    data.Description,
			}
		},
		Created: func(data *VoucherPayTo) []string {
			return []string{
				"voucher_pay_to.create",
				fmt.Sprintf("voucher_pay_to.create.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.create.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *VoucherPayTo) []string {
			return []string{
				"voucher_pay_to.update",
				fmt.Sprintf("voucher_pay_to.update.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.update.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *VoucherPayTo) []string {
			return []string{
				"voucher_pay_to.delete",
				fmt.Sprintf("voucher_pay_to.delete.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.delete.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// VoucherPayToCurrentBranch retrieves all voucher payees for the specified organization and branch
func (m *modelcore) VoucherPayToCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*VoucherPayTo, error) {
	return m.VoucherPayToManager.Find(context, &VoucherPayTo{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
