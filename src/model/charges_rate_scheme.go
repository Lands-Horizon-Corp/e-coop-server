package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type ChargesRateMemberTypeEnum string

const (
	ChargesMemberTypeAll         ChargesRateMemberTypeEnum = "all"
	ChargesMemberTypeDaily       ChargesRateMemberTypeEnum = "daily"
	ChargesMemberTypeWeekly      ChargesRateMemberTypeEnum = "weekly"
	ChargesMemberTypeMonthly     ChargesRateMemberTypeEnum = "monthly"
	ChargesMemberTypeSemiMonthly ChargesRateMemberTypeEnum = "semi-monthly"
	ChargesMemberTypeQuarterly   ChargesRateMemberTypeEnum = "quarterly"
	ChargesMemberTypeSemiAnnual  ChargesRateMemberTypeEnum = "semi-annual"
	ChargesMemberTypeLumpsum     ChargesRateMemberTypeEnum = "lumpsum"
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

		ChargesRateByTermHeaderID uuid.UUID                `gorm:"type:uuid"`
		ChargesRateByTermHeader   *ChargesRateByTermHeader `gorm:"foreignKey:ChargesRateByTermHeaderID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"charges_rate_by_term_header,omitempty"`

		Name        string `gorm:"type:varchar(255);not null;unique"`
		Description string `gorm:"type:text;not null;unique"`
		Icon        string `gorm:"type:varchar(255)"`

		// One-to-many relationship with ChargesRateSchemeAccount
		ChargesRateSchemeAccounts []*ChargesRateSchemeAccount `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_scheme_accounts,omitempty"`

		// One-to-many relationship with ChargesRateByRangeOrMinimumAmount
		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmount `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_by_range_or_minimum_amounts,omitempty"`

		// One-to-many relationship with ChargesRateSchemeModelOfPayment
		ChargesRateSchemeModelOfPayments []*ChargesRateSchemeModelOfPayment `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"charges_rate_scheme_model_of_payments,omitempty"`

		MemberTypeID  uuid.UUID                 `gorm:"type:uuid;not null"`
		MemberType    *MemberType               `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`
		ModeOfPayment ChargesRateMemberTypeEnum `gorm:"type:varchar(20);default:'all'"`

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
	}

	ChargesRateSchemeResponse struct {
		ID                        uuid.UUID                        `json:"id"`
		CreatedAt                 string                           `json:"created_at"`
		CreatedByID               uuid.UUID                        `json:"created_by_id"`
		CreatedBy                 *UserResponse                    `json:"created_by,omitempty"`
		UpdatedAt                 string                           `json:"updated_at"`
		UpdatedByID               uuid.UUID                        `json:"updated_by_id"`
		UpdatedBy                 *UserResponse                    `json:"updated_by,omitempty"`
		OrganizationID            uuid.UUID                        `json:"organization_id"`
		Organization              *OrganizationResponse            `json:"organization,omitempty"`
		BranchID                  uuid.UUID                        `json:"branch_id"`
		Branch                    *BranchResponse                  `json:"branch,omitempty"`
		ChargesRateByTermHeaderID uuid.UUID                        `json:"charges_rate_by_term_header_id"`
		ChargesRateByTermHeader   *ChargesRateByTermHeaderResponse `json:"charges_rate_by_term_header,omitempty"`

		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`

		ChargesRateSchemeAccounts []*ChargesRateSchemeAccountResponse `json:"charges_rate_scheme_accounts,omitempty"`

		ChargesRateByRangeOrMinimumAmounts []*ChargesRateByRangeOrMinimumAmountResponse `json:"charges_rate_by_range_or_minimum_amounts,omitempty"`

		ChargesRateSchemeModelOfPayments []*ChargesRateSchemeModelOfPaymentResponse `json:"charges_rate_scheme_model_of_payments,omitempty"`

		MemberTypeID          uuid.UUID                 `json:"member_type_id"`
		MemberType            *MemberTypeResponse       `json:"member_type,omitempty"`
		ModeOfPayment         ChargesRateMemberTypeEnum `json:"mode_of_payment"`
		ModeOfPaymentHeader1  int                       `json:"mode_of_payment_header_1"`
		ModeOfPaymentHeader2  int                       `json:"mode_of_payment_header_2"`
		ModeOfPaymentHeader3  int                       `json:"mode_of_payment_header_3"`
		ModeOfPaymentHeader4  int                       `json:"mode_of_payment_header_4"`
		ModeOfPaymentHeader5  int                       `json:"mode_of_payment_header_5"`
		ModeOfPaymentHeader6  int                       `json:"mode_of_payment_header_6"`
		ModeOfPaymentHeader7  int                       `json:"mode_of_payment_header_7"`
		ModeOfPaymentHeader8  int                       `json:"mode_of_payment_header_8"`
		ModeOfPaymentHeader9  int                       `json:"mode_of_payment_header_9"`
		ModeOfPaymentHeader10 int                       `json:"mode_of_payment_header_10"`
		ModeOfPaymentHeader11 int                       `json:"mode_of_payment_header_11"`
		ModeOfPaymentHeader12 int                       `json:"mode_of_payment_header_12"`
		ModeOfPaymentHeader13 int                       `json:"mode_of_payment_header_13"`
		ModeOfPaymentHeader14 int                       `json:"mode_of_payment_header_14"`
		ModeOfPaymentHeader15 int                       `json:"mode_of_payment_header_15"`
		ModeOfPaymentHeader16 int                       `json:"mode_of_payment_header_16"`
		ModeOfPaymentHeader17 int                       `json:"mode_of_payment_header_17"`
		ModeOfPaymentHeader18 int                       `json:"mode_of_payment_header_18"`
		ModeOfPaymentHeader19 int                       `json:"mode_of_payment_header_19"`
		ModeOfPaymentHeader20 int                       `json:"mode_of_payment_header_20"`
		ModeOfPaymentHeader21 int                       `json:"mode_of_payment_header_21"`
		ModeOfPaymentHeader22 int                       `json:"mode_of_payment_header_22"`
	}

	ChargesRateSchemeRequest struct {
		ChargesRateByTermHeaderID uuid.UUID   `json:"charges_rate_by_term_header_id,omitempty"`
		Name                      string      `json:"name" validate:"required,min=1,max=255"`
		Description               string      `json:"description" validate:"required"`
		Icon                      string      `json:"icon,omitempty"`
		AccountIDs                []uuid.UUID `json:"account_ids,omitempty"`

		MemberTypeID  uuid.UUID                 `json:"member_type_id" validate:"required"`
		ModeOfPayment ChargesRateMemberTypeEnum `json:"mode_of_payment,omitempty" validate:"omitempty,oneof=all daily weekly monthly semi-monthly quarterly semi-annual lumpsum"`
	}
)

func (m *Model) ChargesRateScheme() {
	m.Migration = append(m.Migration, &ChargesRateScheme{})
	m.ChargesRateSchemeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"ChargesRateByTermHeader",
			"MemberType",
			"ChargesRateSchemeAccounts",
			"ChargesRateByRangeOrMinimumAmounts",
			"ChargesRateSchemeModelOfPayments",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateScheme) *ChargesRateSchemeResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateSchemeResponse{
				ID:                        data.ID,
				CreatedAt:                 data.CreatedAt.Format(time.RFC3339),
				CreatedByID:               data.CreatedByID,
				CreatedBy:                 m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                 data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:               data.UpdatedByID,
				UpdatedBy:                 m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:            data.OrganizationID,
				Organization:              m.OrganizationManager.ToModel(data.Organization),
				BranchID:                  data.BranchID,
				Branch:                    m.BranchManager.ToModel(data.Branch),
				ChargesRateByTermHeaderID: data.ChargesRateByTermHeaderID,
				ChargesRateByTermHeader:   m.ChargesRateByTermHeaderManager.ToModel(data.ChargesRateByTermHeader),
				Name:                      data.Name,
				Description:               data.Description,
				Icon:                      data.Icon,

				ChargesRateSchemeAccounts: m.ChargesRateSchemeAccountManager.ToModels(data.ChargesRateSchemeAccounts),

				ChargesRateByRangeOrMinimumAmounts: m.ChargesRateByRangeOrMinimumAmountManager.ToModels(data.ChargesRateByRangeOrMinimumAmounts),

				ChargesRateSchemeModelOfPayments: m.ChargesRateSchemeModelOfPaymentManager.ToModels(data.ChargesRateSchemeModelOfPayments),

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
			}
		},
		Created: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.create",
				fmt.Sprintf("charges_rate_scheme.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.update",
				fmt.Sprintf("charges_rate_scheme.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.delete",
				fmt.Sprintf("charges_rate_scheme.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) ChargesRateSchemeCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*ChargesRateScheme, error) {
	return m.ChargesRateSchemeManager.Find(context, &ChargesRateScheme{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
