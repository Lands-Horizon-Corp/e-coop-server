package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CashCheckVoucherStatus string

const (
	CashCheckVoucherStatusPending  CashCheckVoucherStatus = "pending"
	CashCheckVoucherStatusPrinted  CashCheckVoucherStatus = "printed"
	CashCheckVoucherStatusApproved CashCheckVoucherStatus = "approved"
	CashCheckVoucherStatusReleased CashCheckVoucherStatus = "released"
)

type (
	CashCheckVoucher struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		Name string `gorm:"type:varchar(255)"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_check_voucher" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null" json:"currency_id"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid" json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		PrintedByID        *uuid.UUID        `gorm:"type:uuid" json:"printed_by_id,omitempty"`
		PrintedBy          *User             `gorm:"foreignKey:PrintedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"printed_by,omitempty"`
		ApprovedByID       *uuid.UUID        `gorm:"type:uuid" json:"approved_by_id,omitempty"`
		ApprovedBy         *User             `gorm:"foreignKey:ApprovedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"approved_by,omitempty"`
		ReleasedByID       *uuid.UUID        `gorm:"type:uuid" json:"released_by_id,omitempty"`
		ReleasedBy         *User             `gorm:"foreignKey:ReleasedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"released_by,omitempty"`

		PayTo string `gorm:"type:varchar(255)" json:"pay_to,omitempty"`

		Status                     CashCheckVoucherStatus `gorm:"type:varchar(20)" json:"status,omitempty"` // enum as string
		Description                string                 `gorm:"type:text" json:"description,omitempty"`
		CashVoucherNumber          string                 `gorm:"type:varchar(255)" json:"cash_voucher_number,omitempty"`
		TotalDebit                 float64                `gorm:"type:decimal" json:"total_debit,omitempty"`
		TotalCredit                float64                `gorm:"type:decimal" json:"total_credit,omitempty"`
		PrintCount                 int                    `gorm:"default:0" json:"print_count,omitempty"`
		EntryDate                  *time.Time             `gorm:"default:null" json:"entry_date,omitempty"`
		PrintedDate                *time.Time             `gorm:"default:null" json:"printed_date,omitempty"`
		ApprovedDate               *time.Time             `gorm:"default:null" json:"approved_date,omitempty"`
		ReleasedDate               *time.Time             `gorm:"default:null" json:"released_date,omitempty"` // SIGNATURES
		ApprovedBySignatureMediaID *uuid.UUID             `gorm:"type:uuid" json:"approved_by_signature_media_id,omitempty"`
		ApprovedBySignatureMedia   *Media                 `gorm:"foreignKey:ApprovedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string                 `gorm:"type:varchar(255)" json:"approved_by_name,omitempty"`
		ApprovedByPosition         string                 `gorm:"type:varchar(255)" json:"approved_by_position,omitempty"`

		PreparedBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"prepared_by_signature_media_id,omitempty"`
		PreparedBySignatureMedia   *Media     `gorm:"foreignKey:PreparedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string     `gorm:"type:varchar(255)" json:"prepared_by_name,omitempty"`
		PreparedByPosition         string     `gorm:"type:varchar(255)" json:"prepared_by_position,omitempty"`

		CertifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"certified_by_signature_media_id,omitempty"`
		CertifiedBySignatureMedia   *Media     `gorm:"foreignKey:CertifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string     `gorm:"type:varchar(255)" json:"certified_by_name,omitempty"`
		CertifiedByPosition         string     `gorm:"type:varchar(255)" json:"certified_by_position,omitempty"`

		VerifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"verified_by_signature_media_id,omitempty"`
		VerifiedBySignatureMedia   *Media     `gorm:"foreignKey:VerifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string     `gorm:"type:varchar(255)" json:"verified_by_name,omitempty"`
		VerifiedByPosition         string     `gorm:"type:varchar(255)" json:"verified_by_position,omitempty"`

		CheckBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"check_by_signature_media_id,omitempty"`
		CheckBySignatureMedia   *Media     `gorm:"foreignKey:CheckBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"check_by_signature_media,omitempty"`
		CheckByName             string     `gorm:"type:varchar(255)" json:"check_by_name,omitempty"`
		CheckByPosition         string     `gorm:"type:varchar(255)" json:"check_by_position,omitempty"`

		AcknowledgeBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeBySignatureMedia   *Media     `gorm:"foreignKey:AcknowledgeBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string     `gorm:"type:varchar(255)" json:"acknowledge_by_name,omitempty"`
		AcknowledgeByPosition         string     `gorm:"type:varchar(255)" json:"acknowledge_by_position,omitempty"`

		NotedBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"noted_by_signature_media_id,omitempty"`
		NotedBySignatureMedia   *Media     `gorm:"foreignKey:NotedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"noted_by_signature_media,omitempty"`
		NotedByName             string     `gorm:"type:varchar(255)" json:"noted_by_name,omitempty"`
		NotedByPosition         string     `gorm:"type:varchar(255)" json:"noted_by_position,omitempty"`

		PostedBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"posted_by_signature_media_id,omitempty"`
		PostedBySignatureMedia   *Media     `gorm:"foreignKey:PostedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"posted_by_signature_media,omitempty"`
		PostedByName             string     `gorm:"type:varchar(255)" json:"posted_by_name,omitempty"`
		PostedByPosition         string     `gorm:"type:varchar(255)" json:"posted_by_position,omitempty"`

		PaidBySignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"paid_by_signature_media_id,omitempty"`
		PaidBySignatureMedia   *Media     `gorm:"foreignKey:PaidBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"paid_by_signature_media,omitempty"`
		PaidByName             string     `gorm:"type:varchar(255)" json:"paid_by_name,omitempty"`
		PaidByPosition         string     `gorm:"type:varchar(255)" json:"paid_by_position,omitempty"`

		CashCheckVoucherTags    []*CashCheckVoucherTag   `gorm:"foreignKey:CashCheckVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"cash_check_voucher_tags,omitempty"`
		CashCheckVoucherEntries []*CashCheckVoucherEntry `gorm:"foreignKey:CashCheckVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"cash_check_voucher_entries,omitempty"`
	}

	CashCheckVoucherResponse struct {
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
		CurrencyID     uuid.UUID             `json:"currency_id"`
		Currency       *CurrencyResponse     `json:"currency,omitempty"`

		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		PrintedByID        *uuid.UUID                `json:"printed_by_id,omitempty"`
		PrintedBy          *UserResponse             `json:"printed_by,omitempty"`
		ApprovedByID       *uuid.UUID                `json:"approved_by_id,omitempty"`
		ApprovedBy         *UserResponse             `json:"approved_by,omitempty"`
		ReleasedByID       *uuid.UUID                `json:"released_by_id,omitempty"`
		ReleasedBy         *UserResponse             `json:"released_by,omitempty"`

		PayTo string `json:"pay_to"`

		Name              string                 `json:"name"`
		Status            CashCheckVoucherStatus `json:"status"`
		Description       string                 `json:"description"`
		CashVoucherNumber string                 `json:"cash_voucher_number"`
		TotalDebit        float64                `json:"total_debit"`
		TotalCredit       float64                `json:"total_credit"`
		PrintCount        int                    `json:"print_count"`
		EntryDate         *string                `json:"entry_date,omitempty"`
		PrintedDate       *string                `json:"printed_date,omitempty"`
		ApprovedDate      *string                `json:"approved_date,omitempty"`
		ReleasedDate      *string                `json:"released_date,omitempty"`

		ApprovedBySignatureMediaID *uuid.UUID     `json:"approved_by_signature_media_id,omitempty"`
		ApprovedBySignatureMedia   *MediaResponse `json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string         `json:"approved_by_name"`
		ApprovedByPosition         string         `json:"approved_by_position"`

		PreparedBySignatureMediaID *uuid.UUID     `json:"prepared_by_signature_media_id,omitempty"`
		PreparedBySignatureMedia   *MediaResponse `json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string         `json:"prepared_by_name"`
		PreparedByPosition         string         `json:"prepared_by_position"`

		CertifiedBySignatureMediaID *uuid.UUID     `json:"certified_by_signature_media_id,omitempty"`
		CertifiedBySignatureMedia   *MediaResponse `json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string         `json:"certified_by_name"`
		CertifiedByPosition         string         `json:"certified_by_position"`

		VerifiedBySignatureMediaID *uuid.UUID     `json:"verified_by_signature_media_id,omitempty"`
		VerifiedBySignatureMedia   *MediaResponse `json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string         `json:"verified_by_name"`
		VerifiedByPosition         string         `json:"verified_by_position"`

		CheckBySignatureMediaID *uuid.UUID     `json:"check_by_signature_media_id,omitempty"`
		CheckBySignatureMedia   *MediaResponse `json:"check_by_signature_media,omitempty"`
		CheckByName             string         `json:"check_by_name"`
		CheckByPosition         string         `json:"check_by_position"`

		AcknowledgeBySignatureMediaID *uuid.UUID     `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeBySignatureMedia   *MediaResponse `json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string         `json:"acknowledge_by_name"`
		AcknowledgeByPosition         string         `json:"acknowledge_by_position"`

		NotedBySignatureMediaID *uuid.UUID     `json:"noted_by_signature_media_id,omitempty"`
		NotedBySignatureMedia   *MediaResponse `json:"noted_by_signature_media,omitempty"`
		NotedByName             string         `json:"noted_by_name"`
		NotedByPosition         string         `json:"noted_by_position"`

		PostedBySignatureMediaID *uuid.UUID     `json:"posted_by_signature_media_id,omitempty"`
		PostedBySignatureMedia   *MediaResponse `json:"posted_by_signature_media,omitempty"`
		PostedByName             string         `json:"posted_by_name"`
		PostedByPosition         string         `json:"posted_by_position"`

		PaidBySignatureMediaID *uuid.UUID     `json:"paid_by_signature_media_id,omitempty"`
		PaidBySignatureMedia   *MediaResponse `json:"paid_by_signature_media,omitempty"`
		PaidByName             string         `json:"paid_by_name"`
		PaidByPosition         string         `json:"paid_by_position"`

		CashCheckVoucherTags    []*CashCheckVoucherTagResponse   `json:"cash_check_voucher_tags,omitempty"`
		CashCheckVoucherEntries []*CashCheckVoucherEntryResponse `json:"cash_check_voucher_entries,omitempty"`
	}

	CashCheckVoucherRequest struct {
		CurrencyID uuid.UUID `json:"currency_id" validate:"required"`

		PayTo             string                 `json:"pay_to,omitempty"`
		Status            CashCheckVoucherStatus `json:"status,omitempty"`
		Description       string                 `json:"description,omitempty"`
		CashVoucherNumber string                 `json:"cash_voucher_number,omitempty"`
		Name              string                 `json:"name" validate:"required"`
		PrintCount        int                    `json:"print_count,omitempty"`
		EntryDate         *time.Time             `json:"entry_date,omitempty"`
		PrintedDate       *time.Time             `json:"printed_date,omitempty"`
		ApprovedDate      *time.Time             `json:"approved_date,omitempty"`
		ReleasedDate      *time.Time             `json:"released_date,omitempty"`

		ApprovedBySignatureMediaID *uuid.UUID `json:"approved_by_signature_media_id,omitempty"`
		ApprovedByName             string     `json:"approved_by_name,omitempty"`
		ApprovedByPosition         string     `json:"approved_by_position,omitempty"`

		PreparedBySignatureMediaID *uuid.UUID `json:"prepared_by_signature_media_id,omitempty"`
		PreparedByName             string     `json:"prepared_by_name,omitempty"`
		PreparedByPosition         string     `json:"prepared_by_position,omitempty"`

		CertifiedBySignatureMediaID *uuid.UUID `json:"certified_by_signature_media_id,omitempty"`
		CertifiedByName             string     `json:"certified_by_name,omitempty"`
		CertifiedByPosition         string     `json:"certified_by_position,omitempty"`

		VerifiedBySignatureMediaID *uuid.UUID `json:"verified_by_signature_media_id,omitempty"`
		VerifiedByName             string     `json:"verified_by_name,omitempty"`
		VerifiedByPosition         string     `json:"verified_by_position,omitempty"`

		CheckBySignatureMediaID *uuid.UUID `json:"check_by_signature_media_id,omitempty"`
		CheckByName             string     `json:"check_by_name,omitempty"`
		CheckByPosition         string     `json:"check_by_position,omitempty"`

		AcknowledgeBySignatureMediaID *uuid.UUID `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeByName             string     `json:"acknowledge_by_name,omitempty"`
		AcknowledgeByPosition         string     `json:"acknowledge_by_position,omitempty"`

		NotedBySignatureMediaID *uuid.UUID `json:"noted_by_signature_media_id,omitempty"`
		NotedByName             string     `json:"noted_by_name,omitempty"`
		NotedByPosition         string     `json:"noted_by_position,omitempty"`

		PostedBySignatureMediaID *uuid.UUID `json:"posted_by_signature_media_id,omitempty"`
		PostedByName             string     `json:"posted_by_name,omitempty"`
		PostedByPosition         string     `json:"posted_by_position,omitempty"`

		PaidBySignatureMediaID *uuid.UUID `json:"paid_by_signature_media_id,omitempty"`
		PaidByName             string     `json:"paid_by_name,omitempty"`
		PaidByPosition         string     `json:"paid_by_position,omitempty"`

		CashCheckVoucherEntries        []*CashCheckVoucherEntryRequest `json:"cash_check_voucher_entries,omitempty"`
		CashCheckVoucherEntriesDeleted uuid.UUIDs                      `json:"cash_check_voucher_entries_deleted,omitempty"`
	}

	CashCheckVoucherPrintRequest struct {
		CashVoucherNumber string `json:"cash_voucher_number" validate:"required"`
	}
)

func CashCheckVoucherManager(service *horizon.HorizonService) *registry.Registry[CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		CashCheckVoucher, CashCheckVoucherResponse, CashCheckVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
			"EmployeeUser", "TransactionBatch", "EmployeeUser.Media",
			"PrintedBy", "ApprovedBy", "ReleasedBy",
			"PrintedBy.Media", "ApprovedBy.Media", "ReleasedBy.Media",
			"CashCheckVoucherTags", "CashCheckVoucherEntries",
			"CashCheckVoucherEntries.MemberProfile", "CashCheckVoucherEntries.Account", "CashCheckVoucherEntries.LoanTransaction", "CashCheckVoucherEntries.MemberProfile.Media",
			"CashCheckVoucherEntries.Account.Currency",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *CashCheckVoucher) *CashCheckVoucherResponse {
			if data == nil {
				return nil
			}
			var printedDate, approvedDate, releasedDate, entryDate *string
			if data.EntryDate != nil {
				str := data.EntryDate.Format(time.RFC3339)
				entryDate = &str
			}
			if data.PrintedDate != nil {
				str := data.PrintedDate.Format(time.RFC3339)
				printedDate = &str
			}
			if data.ApprovedDate != nil {
				str := data.ApprovedDate.Format(time.RFC3339)
				approvedDate = &str
			}
			if data.ReleasedDate != nil {
				str := data.ReleasedDate.Format(time.RFC3339)
				releasedDate = &str
			}
			return &CashCheckVoucherResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				CurrencyID:     data.CurrencyID,
				Currency:       CurrencyManager(service).ToModel(data.Currency),

				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				PrintedByID:        data.PrintedByID,
				PrintedBy:          UserManager(service).ToModel(data.PrintedBy),
				ApprovedByID:       data.ApprovedByID,
				ApprovedBy:         UserManager(service).ToModel(data.ApprovedBy),
				ReleasedByID:       data.ReleasedByID,
				ReleasedBy:         UserManager(service).ToModel(data.ReleasedBy),

				PayTo: data.PayTo,

				Status:            data.Status,
				Description:       data.Description,
				CashVoucherNumber: data.CashVoucherNumber,
				TotalDebit:        data.TotalDebit,
				TotalCredit:       data.TotalCredit,
				PrintCount:        data.PrintCount,
				EntryDate:         entryDate,
				PrintedDate:       printedDate,
				ApprovedDate:      approvedDate,
				ReleasedDate:      releasedDate,

				ApprovedBySignatureMediaID: data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:   MediaManager(service).ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:             data.ApprovedByName,
				ApprovedByPosition:         data.ApprovedByPosition,

				PreparedBySignatureMediaID: data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:   MediaManager(service).ToModel(data.PreparedBySignatureMedia),
				PreparedByName:             data.PreparedByName,
				PreparedByPosition:         data.PreparedByPosition,

				CertifiedBySignatureMediaID: data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:   MediaManager(service).ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:             data.CertifiedByName,
				CertifiedByPosition:         data.CertifiedByPosition,

				VerifiedBySignatureMediaID: data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:   MediaManager(service).ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:             data.VerifiedByName,
				VerifiedByPosition:         data.VerifiedByPosition,

				CheckBySignatureMediaID: data.CheckBySignatureMediaID,
				CheckBySignatureMedia:   MediaManager(service).ToModel(data.CheckBySignatureMedia),
				CheckByName:             data.CheckByName,
				CheckByPosition:         data.CheckByPosition,

				AcknowledgeBySignatureMediaID: data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:   MediaManager(service).ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:             data.AcknowledgeByName,
				AcknowledgeByPosition:         data.AcknowledgeByPosition,

				NotedBySignatureMediaID: data.NotedBySignatureMediaID,
				NotedBySignatureMedia:   MediaManager(service).ToModel(data.NotedBySignatureMedia),
				NotedByName:             data.NotedByName,
				NotedByPosition:         data.NotedByPosition,

				PostedBySignatureMediaID: data.PostedBySignatureMediaID,
				PostedBySignatureMedia:   MediaManager(service).ToModel(data.PostedBySignatureMedia),
				PostedByName:             data.PostedByName,
				PostedByPosition:         data.PostedByPosition,

				PaidBySignatureMediaID: data.PaidBySignatureMediaID,
				PaidBySignatureMedia:   MediaManager(service).ToModel(data.PaidBySignatureMedia),
				PaidByName:             data.PaidByName,
				PaidByPosition:         data.PaidByPosition,

				CashCheckVoucherTags:    CashCheckVoucherTagManager(service).ToModels(data.CashCheckVoucherTags),
				CashCheckVoucherEntries: CashCheckVoucherEntryManager(service).ToModels(data.CashCheckVoucherEntries),

				Name: data.Name,
			}
		},
		Created: func(data *CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.create",
				fmt.Sprintf("cash_check_voucher.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.update",
				fmt.Sprintf("cash_check_voucher.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *CashCheckVoucher) registry.Topics {
			return []string{
				"cash_check_voucher.delete",
				fmt.Sprintf("cash_check_voucher.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CashCheckVoucherCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*CashCheckVoucher, error) {
	return CashCheckVoucherManager(service).Find(context, &CashCheckVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func CashCheckVoucherDraft(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*CashCheckVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "printed_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherPrinted(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*CashCheckVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherApproved(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*CashCheckVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherReleased(ctx context.Context, service *horizon.HorizonService, branchID, organizationID uuid.UUID) ([]*CashCheckVoucher, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}

func CashCheckVoucherReleasedCurrentDay(ctx context.Context, service *horizon.HorizonService, branchID uuid.UUID, organizationID uuid.UUID) ([]*CashCheckVoucher, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "printed_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "approved_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeIsNotEmpty, Value: nil},
		{Field: "released_date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "released_date", Op: query.ModeLT, Value: endOfDay},
	}

	cashCheckVouchers, err := CashCheckVoucherManager(service).ArrFind(ctx, filters, []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	})
	if err != nil {
		return nil, err
	}
	return cashCheckVouchers, nil
}
