package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChargesRateSchemeType string

const (
	ChargesRateSchemeTypeByRange ChargesRateSchemeType = "by_range"
	ChargesRateSchemeTypeByType  ChargesRateSchemeType = "by_type"
	ChargesRateSchemeTypeByTerm  ChargesRateSchemeType = "by_term"
)

type (
	ChargesRateScheme struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name        string                `gorm:"type:varchar(255);not null;unique"`
		Description string                `gorm:"type:varchar(255);default:''"`
		Icon        string                `gorm:"type:varchar(255)"`
		Type        ChargesRateSchemeType `gorm:"type:varchar(50);not null"`

		MemberTypeID  *uuid.UUID         `gorm:"type:uuid"`
		MemberType    *MemberType        `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`
		ModeOfPayment *LoanModeOfPayment `gorm:"type:varchar(20)"`

		ModeOfPaymentHeader1  int `gorm:"default:0"`
		ModeOfPaymentHeader2  int `gorm:"default:0"`
		ModeOfPaymentHeader3  int `gorm:"default:0"`
		ModeOfPaymentHeader4  int `gorm:"default:0"`
		ModeOfPaymentHeader5  int `gorm:"default:0"`
		ModeOfPaymentHeader6  int `gorm:"default:0"`
		ModeOfPaymentHeader7  int `gorm:"default:0"`
		ModeOfPaymentHeader8  int `gorm:"default:0"`
		ModeOfPaymentHeader9  int `gorm:"default:0"`
		ModeOfPaymentHeader10 int `gorm:"default:0"`
		ModeOfPaymentHeader11 int `gorm:"default:0"`
		ModeOfPaymentHeader12 int `gorm:"default:0"`
		ModeOfPaymentHeader13 int `gorm:"default:0"`
		ModeOfPaymentHeader14 int `gorm:"default:0"`
		ModeOfPaymentHeader15 int `gorm:"default:0"`
		ModeOfPaymentHeader16 int `gorm:"default:0"`
		ModeOfPaymentHeader17 int `gorm:"default:0"`
		ModeOfPaymentHeader18 int `gorm:"default:0"`
		ModeOfPaymentHeader19 int `gorm:"default:0"`
		ModeOfPaymentHeader20 int `gorm:"default:0"`
		ModeOfPaymentHeader21 int `gorm:"default:0"`
		ModeOfPaymentHeader22 int `gorm:"default:0"`

		ByTermHeader1  int `gorm:"default:0"`
		ByTermHeader2  int `gorm:"default:0"`
		ByTermHeader3  int `gorm:"default:0"`
		ByTermHeader4  int `gorm:"default:0"`
		ByTermHeader5  int `gorm:"default:0"`
		ByTermHeader6  int `gorm:"default:0"`
		ByTermHeader7  int `gorm:"default:0"`
		ByTermHeader8  int `gorm:"default:0"`
		ByTermHeader9  int `gorm:"default:0"`
		ByTermHeader10 int `gorm:"default:0"`
		ByTermHeader11 int `gorm:"default:0"`
		ByTermHeader12 int `gorm:"default:0"`
		ByTermHeader13 int `gorm:"default:0"`
		ByTermHeader14 int `gorm:"default:0"`
		ByTermHeader15 int `gorm:"default:0"`
		ByTermHeader16 int `gorm:"default:0"`
		ByTermHeader17 int `gorm:"default:0"`
		ByTermHeader18 int `gorm:"default:0"`
		ByTermHeader19 int `gorm:"default:0"`
		ByTermHeader20 int `gorm:"default:0"`
		ByTermHeader21 int `gorm:"default:0"`
		ByTermHeader22 int `gorm:"default:0"`

		ChargesRateSchemeAccounts          []*ChargesRateSchemeAccount          `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_scheme_accounts,omitempty"`
		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmount `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_by_range_or_minimum_amounts,omitempty"`
		ChargesRateSchemeModeOfPayments    []*ChargesRateSchemeModeOfPayment    `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_scheme_model_of_payments,omitempty"`
		ChargesRateByTerms                 []*ChargesRateByTerm                 `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_by_terms,omitempty"`
	}

	ChargesRateSchemeResponse struct {
		ID                        uuid.UUID             `json:"id"`
		CreatedAt                 string                `json:"created_at"`
		CreatedByID               uuid.UUID             `json:"created_by_id"`
		CreatedBy                 *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt                 string                `json:"updated_at"`
		UpdatedByID               uuid.UUID             `json:"updated_by_id"`
		UpdatedBy                 *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID            uuid.UUID             `json:"organization_id"`
		Organization              *OrganizationResponse `json:"organization,omitempty"`
		BranchID                  uuid.UUID             `json:"branch_id"`
		Branch                    *BranchResponse       `json:"branch,omitempty"`
		CurrencyID                uuid.UUID             `json:"currency_id"`
		Currency                  *CurrencyResponse     `json:"currency,omitempty"`
		ChargesRateByTermHeaderID uuid.UUID             `json:"charges_rate_by_term_header_id"`

		Name        string                `json:"name"`
		Description string                `json:"description"`
		Icon        string                `json:"icon"`
		Type        ChargesRateSchemeType `json:"type"`

		MemberTypeID          *uuid.UUID          `json:"member_type_id,omitempty"`
		MemberType            *MemberTypeResponse `json:"member_type,omitempty"`
		ModeOfPayment         *LoanModeOfPayment  `json:"mode_of_payment,omitempty"`
		ModeOfPaymentHeader1  int                 `json:"mode_of_payment_header_1"`
		ModeOfPaymentHeader2  int                 `json:"mode_of_payment_header_2"`
		ModeOfPaymentHeader3  int                 `json:"mode_of_payment_header_3"`
		ModeOfPaymentHeader4  int                 `json:"mode_of_payment_header_4"`
		ModeOfPaymentHeader5  int                 `json:"mode_of_payment_header_5"`
		ModeOfPaymentHeader6  int                 `json:"mode_of_payment_header_6"`
		ModeOfPaymentHeader7  int                 `json:"mode_of_payment_header_7"`
		ModeOfPaymentHeader8  int                 `json:"mode_of_payment_header_8"`
		ModeOfPaymentHeader9  int                 `json:"mode_of_payment_header_9"`
		ModeOfPaymentHeader10 int                 `json:"mode_of_payment_header_10"`
		ModeOfPaymentHeader11 int                 `json:"mode_of_payment_header_11"`
		ModeOfPaymentHeader12 int                 `json:"mode_of_payment_header_12"`
		ModeOfPaymentHeader13 int                 `json:"mode_of_payment_header_13"`
		ModeOfPaymentHeader14 int                 `json:"mode_of_payment_header_14"`
		ModeOfPaymentHeader15 int                 `json:"mode_of_payment_header_15"`
		ModeOfPaymentHeader16 int                 `json:"mode_of_payment_header_16"`
		ModeOfPaymentHeader17 int                 `json:"mode_of_payment_header_17"`
		ModeOfPaymentHeader18 int                 `json:"mode_of_payment_header_18"`
		ModeOfPaymentHeader19 int                 `json:"mode_of_payment_header_19"`
		ModeOfPaymentHeader20 int                 `json:"mode_of_payment_header_20"`
		ModeOfPaymentHeader21 int                 `json:"mode_of_payment_header_21"`
		ModeOfPaymentHeader22 int                 `json:"mode_of_payment_header_22"`

		ByTermHeader1  int `json:"by_term_header_1"`
		ByTermHeader2  int `json:"by_term_header_2"`
		ByTermHeader3  int `json:"by_term_header_3"`
		ByTermHeader4  int `json:"by_term_header_4"`
		ByTermHeader5  int `json:"by_term_header_5"`
		ByTermHeader6  int `json:"by_term_header_6"`
		ByTermHeader7  int `json:"by_term_header_7"`
		ByTermHeader8  int `json:"by_term_header_8"`
		ByTermHeader9  int `json:"by_term_header_9"`
		ByTermHeader10 int `json:"by_term_header_10"`
		ByTermHeader11 int `json:"by_term_header_11"`
		ByTermHeader12 int `json:"by_term_header_12"`
		ByTermHeader13 int `json:"by_term_header_13"`
		ByTermHeader14 int `json:"by_term_header_14"`
		ByTermHeader15 int `json:"by_term_header_15"`
		ByTermHeader16 int `json:"by_term_header_16"`
		ByTermHeader17 int `json:"by_term_header_17"`
		ByTermHeader18 int `json:"by_term_header_18"`
		ByTermHeader19 int `json:"by_term_header_19"`
		ByTermHeader20 int `json:"by_term_header_20"`
		ByTermHeader21 int `json:"by_term_header_21"`
		ByTermHeader22 int `json:"by_term_header_22"`

		ChargesRateSchemeAccounts          []*ChargesRateSchemeAccountResponse          `json:"charges_rate_scheme_accounts,omitempty"`
		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmountResponse `json:"charges_rate_by_range_or_minimum_amounts,omitempty"`
		ChargesRateSchemeModeOfPayments    []*ChargesRateSchemeModeOfPaymentResponse    `json:"charges_rate_scheme_model_of_payments,omitempty"`
		ChargesRateByTerms                 []*ChargesRateByTermResponse                 `json:"charges_rate_by_terms,omitempty"`
	}

	ChargesRateSchemeRequest struct {
		ChargesRateByTermHeaderID uuid.UUID             `json:"charges_rate_by_term_header_id,omitempty"`
		Name                      string                `json:"name" validate:"required,min=1,max=255"`
		Description               string                `json:"description,omitempty" validate:"omitempty,max=255"`
		Icon                      string                `json:"icon,omitempty"`
		Type                      ChargesRateSchemeType `json:"type" validate:"required,oneof=by_range by_type by_term"`
		CurrencyID                uuid.UUID             `json:"currency_id" validate:"required"`
		AccountIDs                uuid.UUIDs            `json:"account_ids,omitempty"`

		MemberTypeID  *uuid.UUID         `json:"member_type_id,omitempty"`
		ModeOfPayment *LoanModeOfPayment `json:"mode_of_payment,omitempty" validate:"omitempty,oneof=all day daily weekly monthly semi-monthly quarterly semi-annual lumpsum"`

		ModeOfPaymentHeader1  int `json:"mode_of_payment_header_1,omitempty"`
		ModeOfPaymentHeader2  int `json:"mode_of_payment_header_2,omitempty"`
		ModeOfPaymentHeader3  int `json:"mode_of_payment_header_3,omitempty"`
		ModeOfPaymentHeader4  int `json:"mode_of_payment_header_4,omitempty"`
		ModeOfPaymentHeader5  int `json:"mode_of_payment_header_5,omitempty"`
		ModeOfPaymentHeader6  int `json:"mode_of_payment_header_6,omitempty"`
		ModeOfPaymentHeader7  int `json:"mode_of_payment_header_7,omitempty"`
		ModeOfPaymentHeader8  int `json:"mode_of_payment_header_8,omitempty"`
		ModeOfPaymentHeader9  int `json:"mode_of_payment_header_9,omitempty"`
		ModeOfPaymentHeader10 int `json:"mode_of_payment_header_10,omitempty"`
		ModeOfPaymentHeader11 int `json:"mode_of_payment_header_11,omitempty"`
		ModeOfPaymentHeader12 int `json:"mode_of_payment_header_12,omitempty"`
		ModeOfPaymentHeader13 int `json:"mode_of_payment_header_13,omitempty"`
		ModeOfPaymentHeader14 int `json:"mode_of_payment_header_14,omitempty"`
		ModeOfPaymentHeader15 int `json:"mode_of_payment_header_15,omitempty"`
		ModeOfPaymentHeader16 int `json:"mode_of_payment_header_16,omitempty"`
		ModeOfPaymentHeader17 int `json:"mode_of_payment_header_17,omitempty"`
		ModeOfPaymentHeader18 int `json:"mode_of_payment_header_18,omitempty"`
		ModeOfPaymentHeader19 int `json:"mode_of_payment_header_19,omitempty"`
		ModeOfPaymentHeader20 int `json:"mode_of_payment_header_20,omitempty"`
		ModeOfPaymentHeader21 int `json:"mode_of_payment_header_21,omitempty"`
		ModeOfPaymentHeader22 int `json:"mode_of_payment_header_22,omitempty"`

		ByTermHeader1  int `json:"by_term_header_1,omitempty"`
		ByTermHeader2  int `json:"by_term_header_2,omitempty"`
		ByTermHeader3  int `json:"by_term_header_3,omitempty"`
		ByTermHeader4  int `json:"by_term_header_4,omitempty"`
		ByTermHeader5  int `json:"by_term_header_5,omitempty"`
		ByTermHeader6  int `json:"by_term_header_6,omitempty"`
		ByTermHeader7  int `json:"by_term_header_7,omitempty"`
		ByTermHeader8  int `json:"by_term_header_8,omitempty"`
		ByTermHeader9  int `json:"by_term_header_9,omitempty"`
		ByTermHeader10 int `json:"by_term_header_10,omitempty"`
		ByTermHeader11 int `json:"by_term_header_11,omitempty"`
		ByTermHeader12 int `json:"by_term_header_12,omitempty"`
		ByTermHeader13 int `json:"by_term_header_13,omitempty"`
		ByTermHeader14 int `json:"by_term_header_14,omitempty"`
		ByTermHeader15 int `json:"by_term_header_15,omitempty"`
		ByTermHeader16 int `json:"by_term_header_16,omitempty"`
		ByTermHeader17 int `json:"by_term_header_17,omitempty"`
		ByTermHeader18 int `json:"by_term_header_18,omitempty"`
		ByTermHeader19 int `json:"by_term_header_19,omitempty"`
		ByTermHeader20 int `json:"by_term_header_20,omitempty"`
		ByTermHeader21 int `json:"by_term_header_21,omitempty"`
		ByTermHeader22 int `json:"by_term_header_22,omitempty"`

		ChargesRateSchemeAccounts          []*ChargesRateSchemeAccountRequest          `json:"charges_rate_scheme_accounts,omitempty"`
		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmountRequest `json:"charges_rate_by_range_or_minimum_amounts,omitempty"`
		ChargesRateSchemeModeOfPayments    []*ChargesRateSchemeModeOfPaymentRequest    `json:"charges_rate_scheme_model_of_payments,omitempty"`
		ChargesRateByTerms                 []*ChargesRateByTermRequest                 `json:"charges_rate_by_terms,omitempty"`

		ChargesRateSchemeAccountsDeleted          uuid.UUIDs `json:"charges_rate_scheme_accounts_deleted,omitempty"`
		ChargesRateByRangeOrMinimumAmountsDeleted uuid.UUIDs `json:"charges_rate_by_range_or_minimum_amounts_deleted,omitempty"`
		ChargesRateSchemeModeOfPaymentsDeleted    uuid.UUIDs `json:"charges_rate_scheme_model_of_payments_deleted,omitempty"`
		ChargesRateByTermsDeleted                 uuid.UUIDs `json:"charges_rate_by_terms_deleted,omitempty"`
	}
)
