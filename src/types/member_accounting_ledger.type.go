package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberAccountingLedger struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_accounting_ledger"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_accounting_ledger"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		AccountID       uuid.UUID      `gorm:"type:uuid;not null"`
		Account         *Account       `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Count               int        `gorm:"type:int"`
		Balance             float64    `gorm:"type:decimal"`
		Interest            float64    `gorm:"type:decimal"`
		Fines               float64    `gorm:"type:decimal"`
		Due                 float64    `gorm:"type:decimal"`
		CarriedForwardDue   float64    `gorm:"type:decimal"`
		StoredValueFacility float64    `gorm:"type:decimal"`
		PrincipalDue        float64    `gorm:"type:decimal"`
		LastPay             *time.Time `gorm:"type:timestamp"`
	}

	MemberAccountingLedgerResponse struct {
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
		AccountID           uuid.UUID              `json:"account_id"`
		Account             *AccountResponse       `json:"account,omitempty"`
		Count               int                    `json:"count"`
		Balance             float64                `json:"balance"`
		Interest            float64                `json:"interest"`
		Fines               float64                `json:"fines"`
		Due                 float64                `json:"due"`
		CarriedForwardDue   float64                `json:"carried_forward_due"`
		StoredValueFacility float64                `json:"stored_value_facility"`
		PrincipalDue        float64                `json:"principal_due"`
		LastPay             *string                `json:"last_pay,omitempty"`
	}

	MemberAccountingLedgerRequest struct {
		OrganizationID      uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID            uuid.UUID  `json:"branch_id" validate:"required"`
		MemberProfileID     uuid.UUID  `json:"member_profile_id" validate:"required"`
		AccountID           uuid.UUID  `json:"account_id" validate:"required"`
		Count               int        `json:"count,omitempty"`
		Balance             float64    `json:"balance,omitempty"`
		Interest            float64    `json:"interest,omitempty"`
		Fines               float64    `json:"fines,omitempty"`
		Due                 float64    `json:"due,omitempty"`
		CarriedForwardDue   float64    `json:"carried_forward_due,omitempty"`
		StoredValueFacility float64    `json:"stored_value_facility,omitempty"`
		PrincipalDue        float64    `json:"principal_due,omitempty"`
		LastPay             *time.Time `json:"last_pay,omitempty"`
	}

	MemberAccountingLedgerUpdateOrCreateParams struct {
		MemberProfileID uuid.UUID `validate:"required"`
		AccountID       uuid.UUID `validate:"required"`
		OrganizationID  uuid.UUID `validate:"required"`
		BranchID        uuid.UUID `validate:"required"`
		UserID          uuid.UUID `validate:"required"`
		DebitAmount     float64
		CreditAmount    float64
		LastPayTime     time.Time `validate:"required"`
	}

	MemberAccountingLedgerAccountSummary struct {
		Balance     float64 `json:"balance"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	MemberAccountingLedgerBrowseReference struct {
		MemberAccountingLedger *MemberAccountingLedger
		BrowseReference        *BrowseReference
	}
)
