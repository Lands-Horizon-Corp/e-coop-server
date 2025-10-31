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
	// PostDatedCheck represents post-dated checks received from members as collateral or payment
	PostDatedCheck struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_post_dated_check"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_post_dated_check"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		FullName       string `gorm:"type:varchar(255)"`
		PassbookNumber string `gorm:"type:varchar(255)"`

		// Check info
		CheckNumber string    `gorm:"type:varchar(255)"`
		CheckDate   time.Time `gorm:"type:timestamp"`
		ClearDays   int       `gorm:"type:int"`
		DateCleared time.Time `gorm:"type:timestamp"`
		BankID      uuid.UUID `gorm:"type:uuid"`
		Bank        *Bank     `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		Amount      float64   `gorm:"type:decimal;default:0"`

		// OR info
		ReferenceNumber     string    `gorm:"type:varchar(255)"`
		OfficialReceiptDate time.Time `gorm:"type:timestamp"`
		CollateralUserID    uuid.UUID `gorm:"type:uuid"`

		Description string `gorm:"type:text"`
	}

	// PostDatedCheckResponse represents the response structure for post-dated check data
	PostDatedCheckResponse struct {
		ID                  uuid.UUID              `json:"id"`
		CreatedAt           string                 `json:"created_at"`
		CreatedByID         uuid.UUID              `json:"created_by_id"`
		CreatedBy           *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt           string                 `json:"updated_at"`
		UpdatedByID         uuid.UUID              `json:"updated_by_id"`
		UpdatedBy           *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID              `json:"organization_id"`
		Organization        *OrganizationResponse  `json:"organization,omitempty"`
		BranchID            uuid.UUID              `json:"branch_id"`
		Branch              *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID     uuid.UUID              `json:"member_profile_id"`
		MemberProfile       *MemberProfileResponse `json:"member_profile,omitempty"`
		FullName            string                 `json:"full_name"`
		PassbookNumber      string                 `json:"passbook_number"`
		CheckNumber         string                 `json:"check_number"`
		CheckDate           string                 `json:"check_date"`
		ClearDays           int                    `json:"clear_days"`
		DateCleared         string                 `json:"date_cleared"`
		BankID              uuid.UUID              `json:"bank_id"`
		Bank                *BankResponse          `json:"bank,omitempty"`
		Amount              float64                `json:"amount"`
		ReferenceNumber     string                 `json:"reference_number"`
		OfficialReceiptDate string                 `json:"official_receipt_date"`
		CollateralUserID    uuid.UUID              `json:"collateral_user_id"`
		Description         string                 `json:"description"`
	}

	// PostDatedCheckRequest represents the request structure for creating or updating post-dated check records
	PostDatedCheckRequest struct {
		MemberProfileID     uuid.UUID `json:"member_profile_id,omitempty"`
		FullName            string    `json:"full_name,omitempty"`
		PassbookNumber      string    `json:"passbook_number,omitempty"`
		CheckNumber         string    `json:"check_number,omitempty"`
		CheckDate           time.Time `json:"check_date,omitempty"`
		ClearDays           int       `json:"clear_days,omitempty"`
		DateCleared         time.Time `json:"date_cleared,omitempty"`
		BankID              uuid.UUID `json:"bank_id,omitempty"`
		Amount              float64   `json:"amount,omitempty"`
		ReferenceNumber     string    `json:"reference_number,omitempty"`
		OfficialReceiptDate time.Time `json:"official_receipt_date,omitempty"`
		CollateralUserID    uuid.UUID `json:"collateral_user_id,omitempty"`
		Description         string    `json:"description,omitempty"`
	}
)

// PostDatedCheck initializes the PostDatedCheck model and its repository manager
func (m *ModelCore) postDatedCheck() {
	m.Migration = append(m.Migration, &PostDatedCheck{})
	m.PostDatedCheckManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		PostDatedCheck, PostDatedCheckResponse, PostDatedCheckRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberProfile", "Bank",
		},
		Service: m.provider.Service,
		Resource: func(data *PostDatedCheck) *PostDatedCheckResponse {
			if data == nil {
				return nil
			}
			return &PostDatedCheckResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        m.OrganizationManager.ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              m.BranchManager.ToModel(data.Branch),
				MemberProfileID:     data.MemberProfileID,
				MemberProfile:       m.MemberProfileManager.ToModel(data.MemberProfile),
				FullName:            data.FullName,
				PassbookNumber:      data.PassbookNumber,
				CheckNumber:         data.CheckNumber,
				CheckDate:           data.CheckDate.Format(time.RFC3339),
				ClearDays:           data.ClearDays,
				DateCleared:         data.DateCleared.Format(time.RFC3339),
				BankID:              data.BankID,
				Bank:                m.BankManager.ToModel(data.Bank),
				Amount:              data.Amount,
				ReferenceNumber:     data.ReferenceNumber,
				OfficialReceiptDate: data.OfficialReceiptDate.Format(time.RFC3339),
				CollateralUserID:    data.CollateralUserID,
				Description:         data.Description,
			}
		},

		Created: func(data *PostDatedCheck) []string {
			return []string{
				"post_dated_check.create",
				fmt.Sprintf("post_dated_check.create.%s", data.ID),
				fmt.Sprintf("post_dated_check.create.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *PostDatedCheck) []string {
			return []string{
				"post_dated_check.update",
				fmt.Sprintf("post_dated_check.update.%s", data.ID),
				fmt.Sprintf("post_dated_check.update.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *PostDatedCheck) []string {
			return []string{
				"post_dated_check.delete",
				fmt.Sprintf("post_dated_check.delete.%s", data.ID),
				fmt.Sprintf("post_dated_check.delete.branch.%s", data.BranchID),
				fmt.Sprintf("post_dated_check.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// PostDatedCheckCurrentBranch retrieves all postdatedcheck records for the specified organization and branch
func (m *ModelCore) postDatedCheckCurrentbranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*PostDatedCheck, error) {
	return m.PostDatedCheckManager.Find(context, &PostDatedCheck{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
