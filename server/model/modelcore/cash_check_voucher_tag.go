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
	CashCheckVoucherTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CashCheckVoucherID *uuid.UUID        `gorm:"type:uuid"`
		CashCheckVoucher   *CashCheckVoucher `gorm:"foreignKey:CashCheckVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"cash_check_voucher,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
		Category    string `gorm:"type:varchar(50)"`
		Color       string `gorm:"type:varchar(20)"`
		Icon        string `gorm:"type:varchar(20)"`
	}

	CashCheckVoucherTagResponse struct {
		ID                 uuid.UUID             `json:"id"`
		CreatedAt          string                `json:"created_at"`
		CreatedByID        uuid.UUID             `json:"created_by_id"`
		CreatedBy          *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt          string                `json:"updated_at"`
		UpdatedByID        uuid.UUID             `json:"updated_by_id"`
		UpdatedBy          *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID             `json:"organization_id"`
		Organization       *OrganizationResponse `json:"organization,omitempty"`
		BranchID           uuid.UUID             `json:"branch_id"`
		Branch             *BranchResponse       `json:"branch,omitempty"`
		CashCheckVoucherID *uuid.UUID            `json:"cash_check_voucher_id,omitempty"`
		Name               string                `json:"name"`
		Description        string                `json:"description"`
		Category           string                `json:"category"`
		Color              string                `json:"color"`
		Icon               string                `json:"icon"`
	}

	CashCheckVoucherTagRequest struct {
		CashCheckVoucherID *uuid.UUID `json:"cash_check_voucher_id,omitempty"`
		Name               string     `json:"name,omitempty"`
		Description        string     `json:"description,omitempty"`
		Category           string     `json:"category,omitempty"`
		Color              string     `json:"color,omitempty"`
		Icon               string     `json:"icon,omitempty"`
	}
)

func (m *ModelCore) cashCheckVoucherTag() {
	m.Migration = append(m.Migration, &CashCheckVoucherTag{})
	m.CashCheckVoucherTagManager = services.NewRepository(services.RepositoryParams[
		CashCheckVoucherTag, CashCheckVoucherTagResponse, CashCheckVoucherTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Service: m.provider.Service,
		Resource: func(data *CashCheckVoucherTag) *CashCheckVoucherTagResponse {
			if data == nil {
				return nil
			}
			return &CashCheckVoucherTagResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager.ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager.ToModel(data.Branch),
				CashCheckVoucherID: data.CashCheckVoucherID,
				Name:               data.Name,
				Description:        data.Description,
				Category:           data.Category,
				Color:              data.Color,
				Icon:               data.Icon,
			}
		},
		Created: func(data *CashCheckVoucherTag) []string {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashCheckVoucherTag) []string {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashCheckVoucherTag) []string {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) cashCheckVoucherTagCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CashCheckVoucherTag, error) {
	return m.CashCheckVoucherTagManager.Find(context, &CashCheckVoucherTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
