package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	OnlineRemittance struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_online_remittance"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_online_remittance"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BankID             uuid.UUID         `gorm:"type:uuid;not null"`
		Bank               *Bank             `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		MediaID            *uuid.UUID        `gorm:"type:uuid"`
		Media              *Media            `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		CountryCode     string     `gorm:"type:varchar(5)"`
		ReferenceNumber string     `gorm:"type:varchar(255)"`
		Amount          float64    `gorm:"type:decimal;not null"`
		AccountName     string     `gorm:"type:varchar(255)"`
		DateEntry       *time.Time `gorm:"type:timestamp"`
		Description     string     `gorm:"type:text"`
	}

	OnlineRemittanceResponse struct {
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
		CountryCode        string                    `json:"country_code"`
		ReferenceNumber    string                    `json:"reference_number"`
		Amount             float64                   `json:"amount"`
		AccountName        string                    `json:"account_name"`
		DateEntry          *string                   `json:"date_entry,omitempty"`
		Description        string                    `json:"description"`
	}

	OnlineRemittanceRequest struct {
		BankID             uuid.UUID  `json:"bank_id" validate:"required"`
		MediaID            *uuid.UUID `json:"media_id,omitempty"`
		EmployeeUserID     *uuid.UUID `json:"employee_user_id,omitempty"`
		TransactionBatchID *uuid.UUID `json:"transaction_batch_id,omitempty"`
		CountryCode        string     `json:"country_code,omitempty"`
		ReferenceNumber    string     `json:"reference_number,omitempty"`
		Amount             float64    `json:"amount" validate:"required"`
		AccountName        string     `json:"account_name,omitempty"`
		DateEntry          *time.Time `json:"date_entry,omitempty"`
		Description        string     `json:"description,omitempty"`
	}
)

func (m *ModelCore) OnlineRemittance() {
	m.Migration = append(m.Migration, &OnlineRemittance{})
	m.OnlineRemittanceManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		OnlineRemittance, OnlineRemittanceResponse, OnlineRemittanceRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"Bank", "Media", "EmployeeUser", "TransactionBatch",
			"Bank.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *OnlineRemittance) *OnlineRemittanceResponse {
			if data == nil {
				return nil
			}
			var dateEntry *string
			if data.DateEntry != nil {
				s := data.DateEntry.Format(time.RFC3339)
				dateEntry = &s
			}
			return &OnlineRemittanceResponse{
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
				CountryCode:        data.CountryCode,
				ReferenceNumber:    data.ReferenceNumber,
				Amount:             data.Amount,
				AccountName:        data.AccountName,
				DateEntry:          dateEntry,
				Description:        data.Description,
			}
		},
		Created: func(data *OnlineRemittance) []string {
			return []string{
				"online_remittance.create",
				fmt.Sprintf("online_remittance.create.%s", data.ID),
				fmt.Sprintf("online_remittance.create.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OnlineRemittance) []string {
			return []string{
				"online_remittance.update",
				fmt.Sprintf("online_remittance.update.%s", data.ID),
				fmt.Sprintf("online_remittance.update.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OnlineRemittance) []string {
			return []string{
				"online_remittance.delete",
				fmt.Sprintf("online_remittance.delete.%s", data.ID),
				fmt.Sprintf("online_remittance.delete.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) OnlineRemittanceCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*OnlineRemittance, error) {
	return m.OnlineRemittanceManager.Find(context, &OnlineRemittance{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
