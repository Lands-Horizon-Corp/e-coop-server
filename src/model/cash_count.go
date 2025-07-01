package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	CashCount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_count"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_count"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EmployeeUserID     uuid.UUID         `gorm:"type:uuid;not null"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID uuid.UUID         `gorm:"type:uuid;not null"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		CountryCode string  `gorm:"type:varchar(5);not null"`
		BillAmount  float64 `gorm:"type:decimal"`
		Quantity    int     `gorm:"type:int"`
		Amount      float64 `gorm:"type:decimal"`
	}

	CashCountResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		EmployeeUserID     uuid.UUID                 `json:"employee_user_id"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID uuid.UUID                 `json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		CountryCode        string                    `json:"country_code"`
		BillAmount         float64                   `json:"bill_amount"`
		Quantity           int                       `json:"quantity"`
		Amount             float64                   `json:"amount"`
	}

	CashCountRequest struct {
		ID                 *uuid.UUID `json:"id,omitempty"`
		OrganizationID     uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID           uuid.UUID  `json:"branch_id" validate:"required"`
		EmployeeUserID     uuid.UUID  `json:"employee_user_id" validate:"required"`
		TransactionBatchID uuid.UUID  `json:"transaction_batch_id" validate:"required"`
		CountryCode        string     `json:"country_code" validate:"required,min=1,max=5"`
		BillAmount         float64    `json:"bill_amount,omitempty"`
		Quantity           int        `json:"quantity,omitempty"`
		Amount             float64    `json:"amount,omitempty"`
	}
)

func (m *Model) CashCount() {
	m.Migration = append(m.Migration, &CashCount{})
	m.CashCountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		CashCount, CashCountResponse, CashCountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"EmployeeUser", "TransactionBatch",
		},
		Service: m.provider.Service,
		Resource: func(data *CashCount) *CashCountResponse {
			if data == nil {
				return nil
			}
			return &CashCountResponse{
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
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       m.UserManager.ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   m.TransactionBatchManager.ToModel(data.TransactionBatch),
				CountryCode:        data.CountryCode,
				BillAmount:         data.BillAmount,
				Quantity:           data.Quantity,
				Amount:             data.Amount,
			}
		},
		Created: func(data *CashCount) []string {
			return []string{
				"cash_count.create",
				fmt.Sprintf("cash_count.create.%s", data.ID),
				fmt.Sprintf("cash_count.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashCount) []string {
			return []string{
				"cash_count.update",
				fmt.Sprintf("cash_count.update.%s", data.ID),
				fmt.Sprintf("cash_count.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashCount) []string {
			return []string{
				"cash_count.delete",
				fmt.Sprintf("cash_count.delete.%s", data.ID),
				fmt.Sprintf("cash_count.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) CashCountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CashCount, error) {
	return m.CashCountManager.Find(context, &CashCount{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
