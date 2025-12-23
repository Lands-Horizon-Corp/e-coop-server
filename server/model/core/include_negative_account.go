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
	IncludeNegativeAccount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_include_negative_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_include_negative_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ComputationSheetID *uuid.UUID        `gorm:"type:uuid"`
		ComputationSheet   *ComputationSheet `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"computation_sheet,omitempty"`
		AccountID          *uuid.UUID        `gorm:"type:uuid"`
		Account            *Account          `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		Description string `gorm:"type:text"`
	}

	IncludeNegativeAccountResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		ComputationSheetID *uuid.UUID                `json:"computation_sheet_id,omitempty"`
		ComputationSheet   *ComputationSheetResponse `json:"computation_sheet,omitempty"`
		AccountID          *uuid.UUID                `json:"account_id,omitempty"`
		Account            *AccountResponse          `json:"account,omitempty"`
		Description        string                    `json:"description"`
	}

	IncludeNegativeAccountRequest struct {
		ComputationSheetID *uuid.UUID `json:"computation_sheet_id,omitempty"`
		AccountID          *uuid.UUID `json:"account_id,omitempty"`
		Description        string     `json:"description,omitempty"`
	}
)

func (m *Core) IncludeNegativeAccountManager() *registry.Registry[IncludeNegativeAccount, IncludeNegativeAccountResponse, IncludeNegativeAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		IncludeNegativeAccount, IncludeNegativeAccountResponse, IncludeNegativeAccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"ComputationSheet", "Account",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *IncludeNegativeAccount) *IncludeNegativeAccountResponse {
			if data == nil {
				return nil
			}
			return &IncludeNegativeAccountResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager().ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager().ToModel(data.Branch),
				ComputationSheetID: data.ComputationSheetID,
				ComputationSheet:   m.ComputationSheetManager().ToModel(data.ComputationSheet),
				AccountID:          data.AccountID,
				Account:            m.AccountManager().ToModel(data.Account),
				Description:        data.Description,
			}
		},
		Created: func(data *IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.create",
				fmt.Sprintf("include_negative_account.create.%s", data.ID),
				fmt.Sprintf("include_negative_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.update",
				fmt.Sprintf("include_negative_account.update.%s", data.ID),
				fmt.Sprintf("include_negative_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.delete",
				fmt.Sprintf("include_negative_account.delete.%s", data.ID),
				fmt.Sprintf("include_negative_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) IncludeNegativeAccountCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*IncludeNegativeAccount, error) {
	return m.IncludeNegativeAccountManager().Find(context, &IncludeNegativeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
