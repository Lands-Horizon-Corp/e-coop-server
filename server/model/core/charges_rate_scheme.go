package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChargesRateSchemeType represents the type of charges rate scheme calculation method
type ChargesRateSchemeType string

// Charges rate scheme type constants
const (
	ChargesRateSchemeTypeByRange ChargesRateSchemeType = "by_range"
	ChargesRateSchemeTypeByType  ChargesRateSchemeType = "by_type"
	ChargesRateSchemeTypeByTerm  ChargesRateSchemeType = "by_term"
)

type (
	// ChargesRateScheme represents the ChargesRateScheme model.
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

		// By type / MOP / Terms
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

		// By Terms
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

	// ChargesRateSchemeResponse represents the response structure for chargesratescheme data

	// ChargesRateSchemeResponse represents the response structure for ChargesRateScheme.
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

	// ChargesRateSchemeRequest represents the request structure for creating/updating chargesratescheme

	// ChargesRateSchemeRequest represents the request structure for ChargesRateScheme.
	ChargesRateSchemeRequest struct {
		ChargesRateByTermHeaderID uuid.UUID             `json:"charges_rate_by_term_header_id,omitempty"`
		Name                      string                `json:"name" validate:"required,min=1,max=255"`
		Description               string                `json:"description,omitempty" validate:"omitempty,max=255"`
		Icon                      string                `json:"icon,omitempty"`
		Type                      ChargesRateSchemeType `json:"type" validate:"required,oneof=by_range by_type by_term"`
		CurrencyID                uuid.UUID             `json:"currency_id" validate:"required"`
		AccountIDs                uuid.UUIDs            `json:"account_ids,omitempty"`

		MemberTypeID  *uuid.UUID         `json:"member_type_id,omitempty"`
		ModeOfPayment *LoanModeOfPayment `json:"mode_of_payment,omitempty" validate:"omitempty,oneof=all daily weekly monthly semi-monthly quarterly semi-annual lumpsum"`

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

		// Nested relationships for creation/update
		ChargesRateSchemeAccounts          []*ChargesRateSchemeAccountRequest          `json:"charges_rate_scheme_accounts,omitempty"`
		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmountRequest `json:"charges_rate_by_range_or_minimum_amounts,omitempty"`
		ChargesRateSchemeModeOfPayments    []*ChargesRateSchemeModeOfPaymentRequest    `json:"charges_rate_scheme_model_of_payments,omitempty"`
		ChargesRateByTerms                 []*ChargesRateByTermRequest                 `json:"charges_rate_by_terms,omitempty"`

		// Deletion arrays
		ChargesRateSchemeAccountsDeleted          uuid.UUIDs `json:"charges_rate_scheme_accounts_deleted,omitempty"`
		ChargesRateByRangeOrMinimumAmountsDeleted uuid.UUIDs `json:"charges_rate_by_range_or_minimum_amounts_deleted,omitempty"`
		ChargesRateSchemeModeOfPaymentsDeleted    uuid.UUIDs `json:"charges_rate_scheme_model_of_payments_deleted,omitempty"`
		ChargesRateByTermsDeleted                 uuid.UUIDs `json:"charges_rate_by_terms_deleted,omitempty"`
	}
)

func (m *Core) chargesRateScheme() {
	m.Migration = append(m.Migration, &ChargesRateScheme{})
	m.ChargesRateSchemeManager = *registry.NewRegistry(registry.RegistryParams[
		ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Currency",

			"MemberType",
			"ChargesRateSchemeAccounts",
			"ChargesRateByRangeOrMinimumAmounts",
			"ChargesRateSchemeModeOfPayments",
			"ChargesRateByTerms",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateScheme) *ChargesRateSchemeResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateSchemeResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				CurrencyID:     data.CurrencyID,
				Currency:       m.CurrencyManager.ToModel(data.Currency),
				Name:           data.Name,
				Description:    data.Description,
				Icon:           data.Icon,
				Type:           data.Type,

				ChargesRateSchemeAccounts: m.ChargesRateSchemeAccountManager.ToModels(data.ChargesRateSchemeAccounts),

				ChargesRateByRangeOrMinimumAmounts: m.ChargesRateByRangeOrMinimumAmountManager.ToModels(data.ChargesRateByRangeOrMinimumAmounts),

				ChargesRateSchemeModeOfPayments: m.ChargesRateSchemeModeOfPaymentManager.ToModels(data.ChargesRateSchemeModeOfPayments),

				ChargesRateByTerms: m.ChargesRateByTermManager.ToModels(data.ChargesRateByTerms),

				MemberTypeID:  data.MemberTypeID,
				MemberType:    m.MemberTypeManager.ToModel(data.MemberType),
				ModeOfPayment: data.ModeOfPayment,

				ModeOfPaymentHeader1:  data.ModeOfPaymentHeader1,
				ModeOfPaymentHeader2:  data.ModeOfPaymentHeader2,
				ModeOfPaymentHeader3:  data.ModeOfPaymentHeader3,
				ModeOfPaymentHeader4:  data.ModeOfPaymentHeader4,
				ModeOfPaymentHeader5:  data.ModeOfPaymentHeader5,
				ModeOfPaymentHeader6:  data.ModeOfPaymentHeader6,
				ModeOfPaymentHeader7:  data.ModeOfPaymentHeader7,
				ModeOfPaymentHeader8:  data.ModeOfPaymentHeader8,
				ModeOfPaymentHeader9:  data.ModeOfPaymentHeader9,
				ModeOfPaymentHeader10: data.ModeOfPaymentHeader10,
				ModeOfPaymentHeader11: data.ModeOfPaymentHeader11,
				ModeOfPaymentHeader12: data.ModeOfPaymentHeader12,
				ModeOfPaymentHeader13: data.ModeOfPaymentHeader13,
				ModeOfPaymentHeader14: data.ModeOfPaymentHeader14,
				ModeOfPaymentHeader15: data.ModeOfPaymentHeader15,
				ModeOfPaymentHeader16: data.ModeOfPaymentHeader16,
				ModeOfPaymentHeader17: data.ModeOfPaymentHeader17,
				ModeOfPaymentHeader18: data.ModeOfPaymentHeader18,
				ModeOfPaymentHeader19: data.ModeOfPaymentHeader19,
				ModeOfPaymentHeader20: data.ModeOfPaymentHeader20,
				ModeOfPaymentHeader21: data.ModeOfPaymentHeader21,
				ModeOfPaymentHeader22: data.ModeOfPaymentHeader22,
				ByTermHeader1:         data.ByTermHeader1,
				ByTermHeader2:         data.ByTermHeader2,
				ByTermHeader3:         data.ByTermHeader3,
				ByTermHeader4:         data.ByTermHeader4,
				ByTermHeader5:         data.ByTermHeader5,
				ByTermHeader6:         data.ByTermHeader6,
				ByTermHeader7:         data.ByTermHeader7,
				ByTermHeader8:         data.ByTermHeader8,
				ByTermHeader9:         data.ByTermHeader9,
				ByTermHeader10:        data.ByTermHeader10,
				ByTermHeader11:        data.ByTermHeader11,
				ByTermHeader12:        data.ByTermHeader12,
				ByTermHeader13:        data.ByTermHeader13,
				ByTermHeader14:        data.ByTermHeader14,
				ByTermHeader15:        data.ByTermHeader15,
				ByTermHeader16:        data.ByTermHeader16,
				ByTermHeader17:        data.ByTermHeader17,
				ByTermHeader18:        data.ByTermHeader18,
				ByTermHeader19:        data.ByTermHeader19,
				ByTermHeader20:        data.ByTermHeader20,
				ByTermHeader21:        data.ByTermHeader21,
				ByTermHeader22:        data.ByTermHeader22,
			}
		},
		Created: func(data *ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.create",
				fmt.Sprintf("charges_rate_scheme.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.update",
				fmt.Sprintf("charges_rate_scheme.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateScheme) registry.Topics {
			return []string{
				"charges_rate_scheme.delete",
				fmt.Sprintf("charges_rate_scheme.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// ChargesRateSchemeCurrentBranch retrieves all charges rate schemes for the current branch and organization
func (m *Core) ChargesRateSchemeCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*ChargesRateScheme, error) {
	return m.ChargesRateSchemeManager.Find(context, &ChargesRateScheme{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
