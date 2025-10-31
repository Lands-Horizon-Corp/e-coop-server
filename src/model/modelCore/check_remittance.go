package modelCore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	CheckRemittance struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_check_remittance"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_check_remittance"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BankID             uuid.UUID         `gorm:"type:uuid;not null"`
		Bank               *Bank             `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		MediaID            *uuid.UUID        `gorm:"type:uuid"`
		Media              *Media            `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		CurrencyID      uuid.UUID  `gorm:"type:uuid;not null"`
		Currency        *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`
		ReferenceNumber string     `gorm:"type:varchar(255)"`
		AccountName     string     `gorm:"type:varchar(255)"`
		Amount          float64    `gorm:"type:decimal;not null"`
		DateEntry       *time.Time `gorm:"type:timestamp"`
		Description     string     `gorm:"type:text"`
	}

	CheckRemittanceResponse struct {
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
		BankID             uuid.UUID                 `json:"bank_id"`
		Bank               *BankResponse             `json:"bank,omitempty"`
		MediaID            *uuid.UUID                `json:"media_id,omitempty"`
		Media              *MediaResponse            `json:"media,omitempty"`
		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		CurrencyID         uuid.UUID                 `json:"currency_id"`
		Currency           *CurrencyResponse         `json:"currency,omitempty"`
		ReferenceNumber    string                    `json:"reference_number"`
		AccountName        string                    `json:"account_name"`
		Amount             float64                   `json:"amount"`
		DateEntry          *string                   `json:"date_entry,omitempty"`
		Description        string                    `json:"description"`
	}

	CheckRemittanceRequest struct {
		BankID             uuid.UUID  `json:"bank_id" validate:"required"`
		MediaID            *uuid.UUID `json:"media_id,omitempty"`
		EmployeeUserID     *uuid.UUID `json:"employee_user_id,omitempty"`
		TransactionBatchID *uuid.UUID `json:"transaction_batch_id,omitempty"`
		CurrencyID         uuid.UUID  `json:"currency_id" validate:"required"`
		ReferenceNumber    string     `json:"reference_number,omitempty"`
		AccountName        string     `json:"account_name,omitempty"`
		Amount             float64    `json:"amount" validate:"required"`
		DateEntry          *time.Time `json:"date_entry,omitempty"`
		Description        string     `json:"description,omitempty"`
	}
)

func (m *ModelCore) CheckRemittance() {
	m.Migration = append(m.Migration, &CheckRemittance{})
	m.CheckRemittanceManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		CheckRemittance, CheckRemittanceResponse, CheckRemittanceRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Bank", "Media", "EmployeeUser", "TransactionBatch", "Currency",
			"Bank.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *CheckRemittance) *CheckRemittanceResponse {
			if data == nil {
				return nil
			}
			var dateEntry *string
			if data.DateEntry != nil {
				s := data.DateEntry.Format(time.RFC3339)
				dateEntry = &s
			}
			return &CheckRemittanceResponse{
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
				BankID:             data.BankID,
				Bank:               m.BankManager.ToModel(data.Bank),
				MediaID:            data.MediaID,
				Media:              m.MediaManager.ToModel(data.Media),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       m.UserManager.ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   m.TransactionBatchManager.ToModel(data.TransactionBatch),
				CurrencyID:         data.CurrencyID,
				Currency:           m.CurrencyManager.ToModel(data.Currency),
				ReferenceNumber:    data.ReferenceNumber,
				AccountName:        data.AccountName,
				Amount:             data.Amount,
				DateEntry:          dateEntry,
				Description:        data.Description,
			}
		},
		Created: func(data *CheckRemittance) []string {
			return []string{
				"check_remittance.create",
				fmt.Sprintf("check_remittance.create.%s", data.ID),
				fmt.Sprintf("check_remittance.create.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CheckRemittance) []string {
			return []string{
				"check_remittance.update",
				fmt.Sprintf("check_remittance.update.%s", data.ID),
				fmt.Sprintf("check_remittance.update.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CheckRemittance) []string {
			return []string{
				"check_remittance.delete",
				fmt.Sprintf("check_remittance.delete.%s", data.ID),
				fmt.Sprintf("check_remittance.delete.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) CheckRemittanceCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*CheckRemittance, error) {
	return m.CheckRemittanceManager.Find(context, &CheckRemittance{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
