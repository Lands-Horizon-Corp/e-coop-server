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
	DisbursementTransaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement_transaction"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement_transaction"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Disbursement       *Disbursement     `gorm:"foreignKey:DisbursementID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"disbursement,omitempty"`
		TransactionBatchID uuid.UUID         `gorm:"type:uuid;not null"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		EmployeeUserID     uuid.UUID         `gorm:"type:uuid;not null"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		DisbursementID  uuid.UUID `gorm:"type:uuid;not null"`
		ReferenceNumber string    `gorm:"type:varchar(50)"`
		Amount          float64   `gorm:"type:decimal"`
		Description     string    `gorm:"type:text"`
		EmployeeName    string    `gorm:"type:varchar(100)"`
	}

	DisbursementTransactionResponse struct {
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

		DisbursementID     uuid.UUID                 `json:"disbursement_id"`
		Disbursement       *DisbursementResponse     `json:"disbursement,omitempty"`
		TransactionBatchID uuid.UUID                 `json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		EmployeeUserID     uuid.UUID                 `json:"employee_user_id"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`

		ReferenceNumber string  `json:"reference_number"`
		Amount          float64 `json:"amount"`
	}

	DisbursementTransactionRequest struct {
		DisbursementID *uuid.UUID `json:"disbursement_id" validate:"required"`

		Description              string `json:"description,omitempty"`
		IsReferenceNumberChecked bool   `json:"is_reference_number_checked,omitempty"`

		ReferenceNumber string  `json:"reference_number,omitempty"`
		Amount          float64 `json:"amount,omitempty"`
	}
)

func (m *ModelCore) disbursementTransaction() {
	m.Migration = append(m.Migration, &DisbursementTransaction{})
	m.DisbursementTransactionManager = services.NewRepository(services.RepositoryParams[
		DisbursementTransaction, DisbursementTransactionResponse, DisbursementTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Disbursement", "TransactionBatch", "EmployeeUser",
			"Disbursement.Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *DisbursementTransaction) *DisbursementTransactionResponse {
			if data == nil {
				return nil
			}
			return &DisbursementTransactionResponse{
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
				DisbursementID:     data.DisbursementID,
				Disbursement:       m.DisbursementManager.ToModel(data.Disbursement),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   m.TransactionBatchManager.ToModel(data.TransactionBatch),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       m.UserManager.ToModel(data.EmployeeUser),
				ReferenceNumber:    data.ReferenceNumber,
				Amount:             data.Amount,
			}
		},
		Created: func(data *DisbursementTransaction) []string {
			return []string{
				"disbursement_transaction.create",
				fmt.Sprintf("disbursement_transaction.create.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *DisbursementTransaction) []string {
			return []string{
				"disbursement_transaction.update",
				fmt.Sprintf("disbursement_transaction.update.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *DisbursementTransaction) []string {
			return []string{
				"disbursement_transaction.delete",
				fmt.Sprintf("disbursement_transaction.delete.%s", data.ID),
				fmt.Sprintf("disbursement_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) disbursementTransactionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*DisbursementTransaction, error) {
	return m.DisbursementTransactionManager.Find(context, &DisbursementTransaction{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
