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
	Transaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		SignatureMediaID   *uuid.UUID        `gorm:"type:uuid"`
		SignatureMedia     *Media            `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"signature_media,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		MemberProfileID      *uuid.UUID          `gorm:"type:uuid"`
		MemberProfile        *MemberProfile      `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID          `gorm:"type:uuid"`
		MemberJointAccount   *MemberJointAccount `gorm:"foreignKey:MemberJointAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_joint_account,omitempty"`

		LoanBalance float64 `gorm:"type:decimal;default:0"`
		LoanDue     float64 `gorm:"type:decimal;default:0"`
		TotalDue    float64 `gorm:"type:decimal;default:0"`
		FinesDue    float64 `gorm:"type:decimal;default:0"`
		TotalLoan   float64 `gorm:"type:decimal;default:0"`
		InterestDue float64 `gorm:"type:decimal;default:0"`

		ReferenceNumber string              `gorm:"type:varchar(50)"`
		Source          GeneralLedgerSource `gorm:"type:varchar(50)"`

		Amount         float64         `gorm:"type:decimal"`
		Description    string          `gorm:"type:text"`
		GeneralLedgers []GeneralLedger `gorm:"foreignKey:TransactionID" json:"general_ledgers,omitempty"`
	}

	TransactionResponse struct {
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

		SignatureMediaID     *uuid.UUID                  `json:"signature_media_id,omitempty"`
		SignatureMedia       *MediaResponse              `json:"signature_media,omitempty"`
		TransactionBatchID   *uuid.UUID                  `json:"transaction_batch_id,omitempty"`
		TransactionBatch     *TransactionBatchResponse   `json:"transaction_batch,omitempty"`
		EmployeeUserID       *uuid.UUID                  `json:"employee_user_id,omitempty"`
		EmployeeUser         *UserResponse               `json:"employee_user,omitempty"`
		MemberProfileID      *uuid.UUID                  `json:"member_profile_id,omitempty"`
		MemberProfile        *MemberProfileResponse      `json:"member_profile,omitempty"`
		MemberJointAccountID *uuid.UUID                  `json:"member_joint_account_id,omitempty"`
		MemberJointAccount   *MemberJointAccountResponse `json:"member_joint_account,omitempty"`
		LoanBalance          float64                     `json:"loan_balance"`
		LoanDue              float64                     `json:"loan_due"`
		TotalDue             float64                     `json:"total_due"`
		FinesDue             float64                     `json:"fines_due"`
		TotalLoan            float64                     `json:"total_loan"`
		InterestDue          float64                     `json:"interest_due"`
		ReferenceNumber      string                      `json:"reference_number"`
		Source               GeneralLedgerSource         `json:"source"`
		Amount               float64                     `json:"amount"`
		Description          string                      `json:"description"`
	}

	TransactionRequest struct {
		SignatureMediaID         *uuid.UUID          `json:"signature_media_id,omitempty"`
		MemberProfileID          *uuid.UUID          `json:"member_profile_id" validate:"required"`
		MemberJointAccountID     *uuid.UUID          `json:"member_joint_account_id,omitempty"`
		ReferenceNumber          string              `json:"reference_number" validate:"required"`
		IsReferenceNumberChecked bool                `json:"is_reference_number_checked,omitempty"`
		Source                   GeneralLedgerSource `json:"source" validate:"required,oneof=withdraw deposit journal payment adjustment 'journal voucher' 'check voucher'"`
		Description              string              `json:"description,omitempty"`
	}
	TransactionRequestEdit struct {
		Description     string `json:"description,omitempty"`
		ReferenceNumber string `json:"reference_number,omitempty"`
	}
)

func (m *Model) Transaction() {
	m.Migration = append(m.Migration, &Transaction{})
	m.TransactionManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Transaction, TransactionResponse, TransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch",
			"Organization", "SignatureMedia", "TransactionBatch", "EmployeeUser",
			"MemberProfile",
			"MemberProfile.Media",
			"MemberJointAccount.PictureMedia",
			"MemberJointAccount.SignatureMedia",
		},
		Service: m.provider.Service,
		Resource: func(data *Transaction) *TransactionResponse {
			if data == nil {
				return nil
			}
			return &TransactionResponse{
				ID:                   data.ID,
				CreatedAt:            data.CreatedAt.Format(time.RFC3339),
				CreatedByID:          data.CreatedByID,
				CreatedBy:            m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:            data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:          data.UpdatedByID,
				UpdatedBy:            m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:       data.OrganizationID,
				Organization:         m.OrganizationManager.ToModel(data.Organization),
				BranchID:             data.BranchID,
				Branch:               m.BranchManager.ToModel(data.Branch),
				SignatureMediaID:     data.SignatureMediaID,
				SignatureMedia:       m.MediaManager.ToModel(data.SignatureMedia),
				TransactionBatchID:   data.TransactionBatchID,
				TransactionBatch:     m.TransactionBatchManager.ToModel(data.TransactionBatch),
				EmployeeUserID:       data.EmployeeUserID,
				EmployeeUser:         m.UserManager.ToModel(data.EmployeeUser),
				MemberProfileID:      data.MemberProfileID,
				MemberProfile:        m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberJointAccountID: data.MemberJointAccountID,
				MemberJointAccount:   m.MemberJointAccountManager.ToModel(data.MemberJointAccount),
				LoanBalance:          data.LoanBalance,
				LoanDue:              data.LoanDue,
				TotalDue:             data.TotalDue,
				FinesDue:             data.FinesDue,
				TotalLoan:            data.TotalLoan,
				InterestDue:          data.InterestDue,
				ReferenceNumber:      data.ReferenceNumber,
				Source:               data.Source,
				Amount:               data.Amount,
				Description:          data.Description,
			}
		},

		Created: func(data *Transaction) []string {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.create.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.create.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.create",
				fmt.Sprintf("transaction.create.%s", data.ID),
				fmt.Sprintf("transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.create.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
		Updated: func(data *Transaction) []string {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.update.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.update.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.update",
				fmt.Sprintf("transaction.update.%s", data.ID),
				fmt.Sprintf("transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.update.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
		Deleted: func(data *Transaction) []string {
			events := []string{}
			if data.MemberProfileID != nil {
				events = append(events, fmt.Sprintf("transaction.update.member_profile.%s", data.MemberProfileID))
			}
			if data.EmployeeUserID != nil {
				events = append(events, fmt.Sprintf("transaction.update.employee.%s", data.EmployeeUserID))
			}
			events = append(events,
				"transaction.delete",
				fmt.Sprintf("transaction.delete.%s", data.ID),
				fmt.Sprintf("transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("transaction.delete.transaction_batch.%s", data.TransactionBatchID),
			)
			return events
		},
	})
}

func (m *Model) TransactionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Transaction, error) {
	return m.TransactionManager.Find(context, &Transaction{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
