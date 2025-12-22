package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BrowseExcludeIncludeAccounts struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_exclude_include_accounts"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_exclude_include_accounts"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ComputationSheetID *uuid.UUID        `gorm:"type:uuid"`
		ComputationSheet   *ComputationSheet `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"computation_sheet,omitempty"`

		FinesAccountID *uuid.UUID `gorm:"type:uuid"`
		FinesAccount   *Account   `gorm:"foreignKey:FinesAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"fines_account,omitempty"`

		ComakerAccountID *uuid.UUID `gorm:"type:uuid"`
		ComakerAccount   *Account   `gorm:"foreignKey:ComakerAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"comaker_account,omitempty"`

		InterestAccountID *uuid.UUID `gorm:"type:uuid"`
		InterestAccount   *Account   `gorm:"foreignKey:InterestAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"interest_account,omitempty"`

		DeliquentAccountID *uuid.UUID `gorm:"type:uuid"`
		DeliquentAccount   *Account   `gorm:"foreignKey:DeliquentAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"deliquent_account,omitempty"`

		IncludeExistingLoanAccountID *uuid.UUID `gorm:"type:uuid"`
		IncludeExistingLoanAccount   *Account   `gorm:"foreignKey:IncludeExistingLoanAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"include_existing_loan_account,omitempty"`
	}

	BrowseExcludeIncludeAccountsResponse struct {
		ID                           uuid.UUID                 `json:"id"`
		CreatedAt                    string                    `json:"created_at"`
		CreatedByID                  uuid.UUID                 `json:"created_by_id"`
		CreatedBy                    *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt                    string                    `json:"updated_at"`
		UpdatedByID                  uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy                    *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID               uuid.UUID                 `json:"organization_id"`
		Organization                 *OrganizationResponse     `json:"organization,omitempty"`
		BranchID                     uuid.UUID                 `json:"branch_id"`
		Branch                       *BranchResponse           `json:"branch,omitempty"`
		ComputationSheetID           *uuid.UUID                `json:"computation_sheet_id,omitempty"`
		ComputationSheet             *ComputationSheetResponse `json:"computation_sheet,omitempty"`
		FinesAccountID               *uuid.UUID                `json:"fines_account_id,omitempty"`
		FinesAccount                 *AccountResponse          `json:"fines_account,omitempty"`
		ComakerAccountID             *uuid.UUID                `json:"comaker_account_id,omitempty"`
		ComakerAccount               *AccountResponse          `json:"comaker_account,omitempty"`
		InterestAccountID            *uuid.UUID                `json:"interest_account_id,omitempty"`
		InterestAccount              *AccountResponse          `json:"interest_account,omitempty"`
		DeliquentAccountID           *uuid.UUID                `json:"deliquent_account_id,omitempty"`
		DeliquentAccount             *AccountResponse          `json:"deliquent_account,omitempty"`
		IncludeExistingLoanAccountID *uuid.UUID                `json:"include_existing_loan_account_id,omitempty"`
		IncludeExistingLoanAccount   *AccountResponse          `json:"include_existing_loan_account,omitempty"`
	}

	BrowseExcludeIncludeAccountsRequest struct {
		ComputationSheetID           *uuid.UUID `json:"computation_sheet_id,omitempty"`
		FinesAccountID               *uuid.UUID `json:"fines_account_id,omitempty"`
		ComakerAccountID             *uuid.UUID `json:"comaker_account_id,omitempty"`
		InterestAccountID            *uuid.UUID `json:"interest_account_id,omitempty"`
		DeliquentAccountID           *uuid.UUID `json:"deliquent_account_id,omitempty"`
		IncludeExistingLoanAccountID *uuid.UUID `json:"include_existing_loan_account_id,omitempty"`
	}
)

func (m *Core) BrowseExcludeIncludeAccountsManager() *registry.Registry[BrowseExcludeIncludeAccounts, BrowseExcludeIncludeAccountsResponse, BrowseExcludeIncludeAccountsRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		BrowseExcludeIncludeAccounts, BrowseExcludeIncludeAccountsResponse, BrowseExcludeIncludeAccountsRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"ComputationSheet",
			"FinesAccount", "ComakerAccount", "InterestAccount", "DeliquentAccount", "IncludeExistingLoanAccount",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *BrowseExcludeIncludeAccounts) *BrowseExcludeIncludeAccountsResponse {
			if data == nil {
				return nil
			}
			return &BrowseExcludeIncludeAccountsResponse{
				ID:                           data.ID,
				CreatedAt:                    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                  data.CreatedByID,
				CreatedBy:                    m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:                    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                  data.UpdatedByID,
				UpdatedBy:                    m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:               data.OrganizationID,
				Organization:                 m.OrganizationManager().ToModel(data.Organization),
				BranchID:                     data.BranchID,
				Branch:                       m.BranchManager().ToModel(data.Branch),
				ComputationSheetID:           data.ComputationSheetID,
				ComputationSheet:             m.ComputationSheetManager().ToModel(data.ComputationSheet),
				FinesAccountID:               data.FinesAccountID,
				FinesAccount:                 m.AccountManager().ToModel(data.FinesAccount),
				ComakerAccountID:             data.ComakerAccountID,
				ComakerAccount:               m.AccountManager().ToModel(data.ComakerAccount),
				InterestAccountID:            data.InterestAccountID,
				InterestAccount:              m.AccountManager().ToModel(data.InterestAccount),
				DeliquentAccountID:           data.DeliquentAccountID,
				DeliquentAccount:             m.AccountManager().ToModel(data.DeliquentAccount),
				IncludeExistingLoanAccountID: data.IncludeExistingLoanAccountID,
				IncludeExistingLoanAccount:   m.AccountManager().ToModel(data.IncludeExistingLoanAccount),
			}
		},
		Created: func(data *BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.create",
				fmt.Sprintf("browse_exclude_include_accounts.create.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.create.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.update",
				fmt.Sprintf("browse_exclude_include_accounts.update.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.update.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.delete",
				fmt.Sprintf("browse_exclude_include_accounts.delete.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.delete.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) BrowseExcludeIncludeAccountsCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*BrowseExcludeIncludeAccounts, error) {
	return m.BrowseExcludeIncludeAccountsManager().Find(context, &BrowseExcludeIncludeAccounts{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
