package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComputationType represents the type of interest computation
type SavingsComputationType string

// Computation type constants
const (
	// Daily Lowest Balance - Uses the lowest balance found during the computation period
	// Formula: Interest = Lowest_Balance × Interest_Rate × (Days_in_Period ÷ 365)
	// Notes:
	// - If period deposited is less than 30 days, NO INTEREST
	// - If lowest balance is below maintaining balance, NO INTEREST
	SavingsComputationTypeDailyLowestBalance SavingsComputationType = "daily_lowest_balance"

	// Average Daily Balance (ADB) - Calculates the average of all daily balances in the period
	// Formula:
	// Step 1: ADB = (Sum of all daily balances) ÷ Number_of_Days_in_Period
	// Step 2: Interest = ADB × Interest_Rate × (Days_in_Period ÷ 365)
	// Notes:
	// - Records balance every day and calculates average
	// - If ADB is below maintaining balance, NO INTEREST
	SavingsComputationTypeAverageDailyBalance SavingsComputationType = "average_daily_balance"

	// Monthly End Lowest Balance - Uses the lowest balance at month end for the period
	// Formula: Interest = Month_End_Lowest_Balance × Interest_Rate × (Days_in_Period ÷ 365)
	// Notes:
	// - If period deposited is less than 30 days, NO INTEREST
	// - If month end balance is below maintaining balance, NO INTEREST
	SavingsComputationTypeMonthlyEndLowestBalance SavingsComputationType = "monthly_end_lowest_balance"

	// ADB End Balance - Average Daily Balance calculated at the end of the period
	// Formula: Same as ADB but computed at period end
	// Interest = ADB_at_Period_End × Interest_Rate × (Days_in_Period ÷ 365)
	SavingsComputationTypeADBEndBalance SavingsComputationType = "adb_end_balance"

	// Monthly Lowest Balance Average - Average of the lowest balances for each month
	// Formula: Interest = (Sum of monthly lowest balances ÷ Number_of_Months) × Interest_Rate × (Days_in_Period ÷ 365)
	SavingsComputationTypeMonthlyLowestBalanceAverage SavingsComputationType = "monthly_lowest_balance_average"

	// Monthly End Balance Average - Average of month-end balances across the period
	// Formula: Interest = (Sum of month-end balances ÷ Number_of_Months) × Interest_Rate × (Days_in_Period ÷ 365)
	SavingsComputationTypeMonthlyEndBalanceAverage SavingsComputationType = "monthly_end_balance_average"

	// Monthly End Balance Total - Uses the final month-end balance for the entire period
	// Formula: Interest = Final_Month_End_Balance × Interest_Rate × (Days_in_Period ÷ 365)
	// Notes:
	// - Only the last day's balance of the final month matters
	// - If period deposited is less than 30 days, NO INTEREST
	// - If final month end balance is below maintaining balance, NO INTEREST
	SavingsComputationTypeMonthlyEndBalanceTotal SavingsComputationType = "monthly_end_balance_total"
)

type (
	// GeneratedSavingsInterest represents a savings interest computation record
	GeneratedSavingsInterest struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generated_savings_interest"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generated_savings_interest"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		DocumentNo          string    `gorm:"type:varchar(255);default:''" json:"document_no"`
		LastComputationDate time.Time `gorm:"not null" json:"last_computation_date" validate:"required"`
		NewComputationDate  time.Time `gorm:"not null" json:"new_computation_date" validate:"required"`

		AccountID    *uuid.UUID  `gorm:"type:uuid"`
		Account      *Account    `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`
		MemberTypeID *uuid.UUID  `gorm:"type:uuid"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		SavingsComputationType          SavingsComputationType `gorm:"type:varchar(50);not null" json:"savings_computation_type" validate:"required,oneof=daily_lowest_balance average_daily_balance monthly_end_lowest_balance adb_end_balance monthly_lowest_balance_average monthly_end_balance_average monthly_end_balance_total"`
		IncludeClosedAccount            bool                   `gorm:"default:false" json:"include_closed_account"`
		IncludeExistingComputedInterest bool                   `gorm:"default:false" json:"include_existing_computed_interest"`

		InterestTaxRate float64 `gorm:"type:decimal(15,6);default:0" json:"interest_tax_rate"`
		TotalInterest   float64 `gorm:"type:decimal(15,2);default:0" json:"total_interest"`
		TotalTax        float64 `gorm:"type:decimal(15,2);default:0" json:"total_tax"`

		PrintedByUserID *uuid.UUID `gorm:"type:uuid"`
		PrintedByUser   *User      `gorm:"foreignKey:PrintedByUserID;constraint:OnDelete:SET NULL;" json:"printed_by_user,omitempty"`
		PrintedDate     *time.Time `gorm:"" json:"printed_date,omitempty"`

		PostedByUserID *uuid.UUID `gorm:"type:uuid"`
		PostedByUser   *User      `gorm:"foreignKey:PostedByUserID;constraint:OnDelete:SET NULL;" json:"posted_by_user,omitempty"`
		PostedDate     *time.Time `gorm:"" json:"posted_date,omitempty"`

		CheckVoucherNumber *string `gorm:"type:varchar(255)" json:"check_voucher_number,omitempty"`

		PostAccountID *uuid.UUID `json:"post_account_id,omitempty"`
		PostAccount   *Account   `json:"post_account,omitempty"`

		// Relationships
		Entries []*GeneratedSavingsInterestEntry `gorm:"foreignKey:GeneratedSavingsInterestID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"entries,omitempty"`
	}

	// GeneratedSavingsInterestResponse represents the response structure for generated savings interest data
	GeneratedSavingsInterestResponse struct {
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

		DocumentNo                      string                                   `json:"document_no"`
		LastComputationDate             string                                   `json:"last_computation_date"`
		NewComputationDate              string                                   `json:"new_computation_date"`
		AccountID                       *uuid.UUID                               `json:"account_id,omitempty"`
		Account                         *AccountResponse                         `json:"account,omitempty"`
		MemberTypeID                    *uuid.UUID                               `json:"member_type_id,omitempty"`
		MemberType                      *MemberTypeResponse                      `json:"member_type,omitempty"`
		SavingsComputationType          SavingsComputationType                   `json:"savings_computation_type"`
		IncludeClosedAccount            bool                                     `json:"include_closed_account"`
		IncludeExistingComputedInterest bool                                     `json:"include_existing_computed_interest"`
		InterestTaxRate                 float64                                  `json:"interest_tax_rate"`
		TotalInterest                   float64                                  `json:"total_interest"`
		TotalTax                        float64                                  `json:"total_tax"`
		PrintedByUserID                 *uuid.UUID                               `json:"printed_by_user_id,omitempty"`
		PrintedByUser                   *UserResponse                            `json:"printed_by_user,omitempty"`
		PrintedDate                     *string                                  `json:"printed_date,omitempty"`
		PostedByUserID                  *uuid.UUID                               `json:"posted_by_user_id,omitempty"`
		PostedByUser                    *UserResponse                            `json:"posted_by_user,omitempty"`
		PostedDate                      *string                                  `json:"posted_date,omitempty"`
		CheckVoucherNumber              *string                                  `json:"check_voucher_number,omitempty"`
		PostAccountID                   *uuid.UUID                               `json:"post_account_id"`
		PostAccount                     *AccountResponse                         `json:"post_account,omitempty"`
		Entries                         []*GeneratedSavingsInterestEntryResponse `json:"entries,omitempty"`
	}

	// GeneratedSavingsInterestRequest represents the request structure for creating/updating generated savings interest
	GeneratedSavingsInterestRequest struct {
		DocumentNo                      string                 `json:"document_no"`
		LastComputationDate             time.Time              `json:"last_computation_date" validate:"required"`
		NewComputationDate              time.Time              `json:"new_computation_date" validate:"required"`
		AccountID                       *uuid.UUID             `json:"account_id"`
		MemberTypeID                    *uuid.UUID             `json:"member_type_id"`
		SavingsComputationType          SavingsComputationType `json:"savings_computation_type" validate:"required,oneof=daily_lowest_balance average_daily_balance monthly_end_lowest_balance adb_end_balance monthly_lowest_balance_average monthly_end_balance_average monthly_end_balance_total"`
		IncludeClosedAccount            bool                   `json:"include_closed_account"`
		IncludeExistingComputedInterest bool                   `json:"include_existing_computed_interest"`
		InterestTaxRate                 float64                `json:"interest_tax_rate"`
	}

	GenerateSavingsInterestPostRequest struct {
		CheckVoucherNumber *string    `json:"check_voucher_number"`
		PostAccountID      *uuid.UUID `json:"post_account_id"`
		EntryDate          *time.Time `json:"entry_date"`
	}

	GeneratedSavingsInterestViewResponse struct {
		Entries       []*GeneratedSavingsInterestEntryResponse `json:"entries,omitempty"`
		TotalTax      float64                                  `json:"total_tax"`
		TotalInterest float64                                  `json:"total_interest"`
	}
)

func (m *Core) generatedSavingsInterest() {
	m.Migration = append(m.Migration, &GeneratedSavingsInterest{})
	m.GeneratedSavingsInterestManager = *registry.NewRegistry(registry.RegistryParams[
		GeneratedSavingsInterest, GeneratedSavingsInterestResponse, GeneratedSavingsInterestRequest,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"Account",
			"Account.Currency",
			"MemberType",
			"PrintedByUser", "PostedByUser", "PostAccount",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneratedSavingsInterest) *GeneratedSavingsInterestResponse {
			if data == nil {
				return nil
			}

			var postedDate *string
			if data.PostedDate != nil {
				formatted := data.PostedDate.Format(time.RFC3339)
				postedDate = &formatted
			}

			var printedDate *string
			if data.PrintedDate != nil {
				formatted := data.PrintedDate.Format(time.RFC3339)
				printedDate = &formatted
			}

			return &GeneratedSavingsInterestResponse{
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

				DocumentNo:                      data.DocumentNo,
				LastComputationDate:             data.LastComputationDate.Format(time.RFC3339),
				NewComputationDate:              data.NewComputationDate.Format(time.RFC3339),
				AccountID:                       data.AccountID,
				Account:                         m.AccountManager.ToModel(data.Account),
				MemberTypeID:                    data.MemberTypeID,
				MemberType:                      m.MemberTypeManager.ToModel(data.MemberType),
				SavingsComputationType:          data.SavingsComputationType,
				IncludeClosedAccount:            data.IncludeClosedAccount,
				IncludeExistingComputedInterest: data.IncludeExistingComputedInterest,
				InterestTaxRate:                 data.InterestTaxRate,
				TotalInterest:                   data.TotalInterest,
				TotalTax:                        data.TotalTax,
				PrintedByUserID:                 data.PrintedByUserID,
				PrintedByUser:                   m.UserManager.ToModel(data.PrintedByUser),
				PrintedDate:                     printedDate,
				PostedByUserID:                  data.PostedByUserID,
				PostedByUser:                    m.UserManager.ToModel(data.PostedByUser),
				PostedDate:                      postedDate,
				CheckVoucherNumber:              data.CheckVoucherNumber,
				PostAccountID:                   data.PostAccountID,
				PostAccount:                     m.AccountManager.ToModel(data.PostAccount),
				Entries:                         m.GeneratedSavingsInterestEntryManager.ToModels(data.Entries),
			}
		},

		Created: func(data *GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.create",
				fmt.Sprintf("generated_savings_interest.create.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.create.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.update",
				fmt.Sprintf("generated_savings_interest.update.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.update.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneratedSavingsInterest) registry.Topics {
			return []string{
				"generated_savings_interest.delete",
				fmt.Sprintf("generated_savings_interest.delete.%s", data.ID),
				fmt.Sprintf("generated_savings_interest.delete.branch.%s", data.BranchID),
				fmt.Sprintf("generated_savings_interest.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
