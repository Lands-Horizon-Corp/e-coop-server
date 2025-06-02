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
	JournalVoucher struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EmployeeUserID  *uuid.UUID     `gorm:"type:uuid"`
		EmployeeUser    *User          `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		PrintedByUserID  *uuid.UUID `gorm:"type:uuid"`
		PrintedByUser    *User      `gorm:"foreignKey:PrintedByUserID;constraint:OnDelete:SET NULL;" json:"printed_by_user,omitempty"`
		ApprovedByUserID *uuid.UUID `gorm:"type:uuid"`
		ApprovedByUser   *User      `gorm:"foreignKey:ApprovedByUserID;constraint:OnDelete:SET NULL;" json:"approved_by_user,omitempty"`
		ReleasedByUserID *uuid.UUID `gorm:"type:uuid"`
		ReleasedByUser   *User      `gorm:"foreignKey:ReleasedByUserID;constraint:OnDelete:SET NULL;" json:"released_by_user,omitempty"`

		TotalDebit  float64 `gorm:"type:decimal"`
		TotalCredit float64 `gorm:"type:decimal"`

		PrintCount  int    `gorm:"default:0"`
		Description string `gorm:"type:text"`

		PrintedDate  *time.Time `gorm:"type:timestamp"`
		ApprovedDate *time.Time `gorm:"type:timestamp"`
		ReleasedDate *time.Time `gorm:"type:timestamp"`

		// SIGNATURES
		ApprovedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBySignatureMedia   *Media     `gorm:"foreignKey:ApprovedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"approved_by_signature_media,omitempty"`
		ApprovedByName             string     `gorm:"type:varchar(255)"`
		ApprovedByPosition         string     `gorm:"type:varchar(255)"`

		PreparedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PreparedBySignatureMedia   *Media     `gorm:"foreignKey:PreparedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"prepared_by_signature_media,omitempty"`
		PreparedByName             string     `gorm:"type:varchar(255)"`
		PreparedByPosition         string     `gorm:"type:varchar(255)"`

		CertifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CertifiedBySignatureMedia   *Media     `gorm:"foreignKey:CertifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"certified_by_signature_media,omitempty"`
		CertifiedByName             string     `gorm:"type:varchar(255)"`
		CertifiedByPosition         string     `gorm:"type:varchar(255)"`

		VerifiedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		VerifiedBySignatureMedia   *Media     `gorm:"foreignKey:VerifiedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"verified_by_signature_media,omitempty"`
		VerifiedByName             string     `gorm:"type:varchar(255)"`
		VerifiedByPosition         string     `gorm:"type:varchar(255)"`

		CheckBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		CheckBySignatureMedia   *Media     `gorm:"foreignKey:CheckBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"check_by_signature_media,omitempty"`
		CheckByName             string     `gorm:"type:varchar(255)"`
		CheckByPosition         string     `gorm:"type:varchar(255)"`

		AcknowledgeBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		AcknowledgeBySignatureMedia   *Media     `gorm:"foreignKey:AcknowledgeBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"acknowledge_by_signature_media,omitempty"`
		AcknowledgeByName             string     `gorm:"type:varchar(255)"`
		AcknowledgeByPosition         string     `gorm:"type:varchar(255)"`

		NotedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		NotedBySignatureMedia   *Media     `gorm:"foreignKey:NotedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"noted_by_signature_media,omitempty"`
		NotedByName             string     `gorm:"type:varchar(255)"`
		NotedByPosition         string     `gorm:"type:varchar(255)"`

		PostedBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PostedBySignatureMedia   *Media     `gorm:"foreignKey:PostedBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"posted_by_signature_media,omitempty"`
		PostedByName             string     `gorm:"type:varchar(255)"`
		PostedByPosition         string     `gorm:"type:varchar(255)"`

		PaidBySignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		PaidBySignatureMedia   *Media     `gorm:"foreignKey:PaidBySignatureMediaID;constraint:OnDelete:SET NULL;" json:"paid_by_signature_media,omitempty"`
		PaidByName             string     `gorm:"type:varchar(255)"`
		PaidByPosition         string     `gorm:"type:varchar(255)"`
	}

	JournalVoucherResponse struct {
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

		EmployeeUserID  *uuid.UUID             `json:"employee_user_id,omitempty"`
		EmployeeUser    *UserResponse          `json:"employee_user,omitempty"`
		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		PrintedByUserID  *uuid.UUID    `json:"printed_by_user_id,omitempty"`
		PrintedByUser    *UserResponse `json:"printed_by_user,omitempty"`
		ApprovedByUserID *uuid.UUID    `json:"approved_by_user_id,omitempty"`
		ApprovedByUser   *UserResponse `json:"approved_by_user,omitempty"`
		ReleasedByUserID *uuid.UUID    `json:"released_by_user_id,omitempty"`
		ReleasedByUser   *UserResponse `json:"released_by_user,omitempty"`

		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`

		PrintCount  int    `json:"print_count"`
		Description string `json:"description"`

		PrintedDate  *string `json:"printed_date,omitempty"`
		ApprovedDate *string `json:"approved_date,omitempty"`
		ReleasedDate *string `json:"released_date,omitempty"`

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
	}

	JournalVoucherRequest struct {
		EmployeeUserID  *uuid.UUID `json:"employee_user_id,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`

		PrintedByUserID  *uuid.UUID `json:"printed_by_user_id,omitempty"`
		ApprovedByUserID *uuid.UUID `json:"approved_by_user_id,omitempty"`
		ReleasedByUserID *uuid.UUID `json:"released_by_user_id,omitempty"`

		TotalDebit  float64 `json:"total_debit,omitempty"`
		TotalCredit float64 `json:"total_credit,omitempty"`

		PrintCount  int    `json:"print_count,omitempty"`
		Description string `json:"description,omitempty"`

		PrintedDate  *time.Time `json:"printed_date,omitempty"`
		ApprovedDate *time.Time `json:"approved_date,omitempty"`
		ReleasedDate *time.Time `json:"released_date,omitempty"`

		ApprovedBySignatureMediaID    *uuid.UUID `json:"approved_by_signature_media_id,omitempty"`
		ApprovedByName                string     `json:"approved_by_name,omitempty"`
		ApprovedByPosition            string     `json:"approved_by_position,omitempty"`
		PreparedBySignatureMediaID    *uuid.UUID `json:"prepared_by_signature_media_id,omitempty"`
		PreparedByName                string     `json:"prepared_by_name,omitempty"`
		PreparedByPosition            string     `json:"prepared_by_position,omitempty"`
		CertifiedBySignatureMediaID   *uuid.UUID `json:"certified_by_signature_media_id,omitempty"`
		CertifiedByName               string     `json:"certified_by_name,omitempty"`
		CertifiedByPosition           string     `json:"certified_by_position,omitempty"`
		VerifiedBySignatureMediaID    *uuid.UUID `json:"verified_by_signature_media_id,omitempty"`
		VerifiedByName                string     `json:"verified_by_name,omitempty"`
		VerifiedByPosition            string     `json:"verified_by_position,omitempty"`
		CheckBySignatureMediaID       *uuid.UUID `json:"check_by_signature_media_id,omitempty"`
		CheckByName                   string     `json:"check_by_name,omitempty"`
		CheckByPosition               string     `json:"check_by_position,omitempty"`
		AcknowledgeBySignatureMediaID *uuid.UUID `json:"acknowledge_by_signature_media_id,omitempty"`
		AcknowledgeByName             string     `json:"acknowledge_by_name,omitempty"`
		AcknowledgeByPosition         string     `json:"acknowledge_by_position,omitempty"`
		NotedBySignatureMediaID       *uuid.UUID `json:"noted_by_signature_media_id,omitempty"`
		NotedByName                   string     `json:"noted_by_name,omitempty"`
		NotedByPosition               string     `json:"noted_by_position,omitempty"`
		PostedBySignatureMediaID      *uuid.UUID `json:"posted_by_signature_media_id,omitempty"`
		PostedByName                  string     `json:"posted_by_name,omitempty"`
		PostedByPosition              string     `json:"posted_by_position,omitempty"`
		PaidBySignatureMediaID        *uuid.UUID `json:"paid_by_signature_media_id,omitempty"`
		PaidByName                    string     `json:"paid_by_name,omitempty"`
		PaidByPosition                string     `json:"paid_by_position,omitempty"`
	}
)

func (m *Model) JournalVoucher() {
	m.Migration = append(m.Migration, &JournalVoucher{})
	m.JournalVoucherManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		JournalVoucher, JournalVoucherResponse, JournalVoucherRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"EmployeeUser", "MemberProfile", "PrintedByUser", "ApprovedByUser", "ReleasedByUser",
			"ApprovedBySignatureMedia", "PreparedBySignatureMedia", "CertifiedBySignatureMedia",
			"VerifiedBySignatureMedia", "CheckBySignatureMedia", "AcknowledgeBySignatureMedia",
			"NotedBySignatureMedia", "PostedBySignatureMedia", "PaidBySignatureMedia",
		},
		Service: m.provider.Service,
		Resource: func(data *JournalVoucher) *JournalVoucherResponse {
			if data == nil {
				return nil
			}
			var printedDate, approvedDate, releasedDate *string
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
			return &JournalVoucherResponse{
				ID:                            data.ID,
				CreatedAt:                     data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                   data.CreatedByID,
				CreatedBy:                     m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                     data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                   data.UpdatedByID,
				UpdatedBy:                     m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                data.OrganizationID,
				Organization:                  m.OrganizationManager.ToModel(data.Organization),
				BranchID:                      data.BranchID,
				Branch:                        m.BranchManager.ToModel(data.Branch),
				EmployeeUserID:                data.EmployeeUserID,
				EmployeeUser:                  m.UserManager.ToModel(data.EmployeeUser),
				MemberProfileID:               data.MemberProfileID,
				MemberProfile:                 m.MemberProfileManager.ToModel(data.MemberProfile),
				PrintedByUserID:               data.PrintedByUserID,
				PrintedByUser:                 m.UserManager.ToModel(data.PrintedByUser),
				ApprovedByUserID:              data.ApprovedByUserID,
				ApprovedByUser:                m.UserManager.ToModel(data.ApprovedByUser),
				ReleasedByUserID:              data.ReleasedByUserID,
				ReleasedByUser:                m.UserManager.ToModel(data.ReleasedByUser),
				TotalDebit:                    data.TotalDebit,
				TotalCredit:                   data.TotalCredit,
				PrintCount:                    data.PrintCount,
				Description:                   data.Description,
				PrintedDate:                   printedDate,
				ApprovedDate:                  approvedDate,
				ReleasedDate:                  releasedDate,
				ApprovedBySignatureMediaID:    data.ApprovedBySignatureMediaID,
				ApprovedBySignatureMedia:      m.MediaManager.ToModel(data.ApprovedBySignatureMedia),
				ApprovedByName:                data.ApprovedByName,
				ApprovedByPosition:            data.ApprovedByPosition,
				PreparedBySignatureMediaID:    data.PreparedBySignatureMediaID,
				PreparedBySignatureMedia:      m.MediaManager.ToModel(data.PreparedBySignatureMedia),
				PreparedByName:                data.PreparedByName,
				PreparedByPosition:            data.PreparedByPosition,
				CertifiedBySignatureMediaID:   data.CertifiedBySignatureMediaID,
				CertifiedBySignatureMedia:     m.MediaManager.ToModel(data.CertifiedBySignatureMedia),
				CertifiedByName:               data.CertifiedByName,
				CertifiedByPosition:           data.CertifiedByPosition,
				VerifiedBySignatureMediaID:    data.VerifiedBySignatureMediaID,
				VerifiedBySignatureMedia:      m.MediaManager.ToModel(data.VerifiedBySignatureMedia),
				VerifiedByName:                data.VerifiedByName,
				VerifiedByPosition:            data.VerifiedByPosition,
				CheckBySignatureMediaID:       data.CheckBySignatureMediaID,
				CheckBySignatureMedia:         m.MediaManager.ToModel(data.CheckBySignatureMedia),
				CheckByName:                   data.CheckByName,
				CheckByPosition:               data.CheckByPosition,
				AcknowledgeBySignatureMediaID: data.AcknowledgeBySignatureMediaID,
				AcknowledgeBySignatureMedia:   m.MediaManager.ToModel(data.AcknowledgeBySignatureMedia),
				AcknowledgeByName:             data.AcknowledgeByName,
				AcknowledgeByPosition:         data.AcknowledgeByPosition,
				NotedBySignatureMediaID:       data.NotedBySignatureMediaID,
				NotedBySignatureMedia:         m.MediaManager.ToModel(data.NotedBySignatureMedia),
				NotedByName:                   data.NotedByName,
				NotedByPosition:               data.NotedByPosition,
				PostedBySignatureMediaID:      data.PostedBySignatureMediaID,
				PostedBySignatureMedia:        m.MediaManager.ToModel(data.PostedBySignatureMedia),
				PostedByName:                  data.PostedByName,
				PostedByPosition:              data.PostedByPosition,
				PaidBySignatureMediaID:        data.PaidBySignatureMediaID,
				PaidBySignatureMedia:          m.MediaManager.ToModel(data.PaidBySignatureMedia),
				PaidByName:                    data.PaidByName,
				PaidByPosition:                data.PaidByPosition,
			}
		},

		Created: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.create",
				fmt.Sprintf("journal_voucher.create.%s", data.ID),
				fmt.Sprintf("journal_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.update",
				fmt.Sprintf("journal_voucher.update.%s", data.ID),
				fmt.Sprintf("journal_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *JournalVoucher) []string {
			return []string{
				"journal_voucher.delete",
				fmt.Sprintf("journal_voucher.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) JournalVoucherCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*JournalVoucher, error) {
	return m.JournalVoucherManager.Find(context, &JournalVoucher{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
