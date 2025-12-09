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
	// ChargesRateSchemeAccount represents the ChargesRateSchemeAccount model.
	ChargesRateSchemeAccount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ChargesRateSchemeID uuid.UUID          `gorm:"type:uuid;not null"`
		ChargesRateScheme   *ChargesRateScheme `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"charges_rate_scheme,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
	}

	// ChargesRateSchemeAccountResponse represents the response structure for chargesrateschemeaccount data

	// ChargesRateSchemeAccountResponse represents the response structure for ChargesRateSchemeAccount.
	ChargesRateSchemeAccountResponse struct {
		ID                  uuid.UUID                  `json:"id"`
		CreatedAt           string                     `json:"created_at"`
		CreatedByID         uuid.UUID                  `json:"created_by_id"`
		CreatedBy           *UserResponse              `json:"created_by,omitempty"`
		UpdatedAt           string                     `json:"updated_at"`
		UpdatedByID         uuid.UUID                  `json:"updated_by_id"`
		UpdatedBy           *UserResponse              `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID                  `json:"organization_id"`
		Organization        *OrganizationResponse      `json:"organization,omitempty"`
		BranchID            uuid.UUID                  `json:"branch_id"`
		Branch              *BranchResponse            `json:"branch,omitempty"`
		ChargesRateSchemeID uuid.UUID                  `json:"charges_rate_scheme_id"`
		ChargesRateScheme   *ChargesRateSchemeResponse `json:"charges_rate_scheme,omitempty"`
		AccountID           uuid.UUID                  `json:"account_id"`
		Account             *AccountResponse           `json:"account,omitempty"`
	}

	// ChargesRateSchemeAccountRequest represents the request structure for creating/updating chargesrateschemeaccount

	// ChargesRateSchemeAccountRequest represents the request structure for ChargesRateSchemeAccount.
	ChargesRateSchemeAccountRequest struct {
		ID        *uuid.UUID `json:"id,omitempty"`
		AccountID uuid.UUID  `json:"account_id" validate:"required"`
	}
)

func (m *Core) chargesRateSchemeAccount() {
	m.Migration = append(m.Migration, &ChargesRateSchemeAccount{})
	m.ChargesRateSchemeAccountManager = *registry.NewRegistry(registry.RegistryParams[
		ChargesRateSchemeAccount, ChargesRateSchemeAccountResponse, ChargesRateSchemeAccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "ChargesRateScheme", "Account",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateSchemeAccount) *ChargesRateSchemeAccountResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateSchemeAccountResponse{
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
				ChargesRateSchemeID: data.ChargesRateSchemeID,
				ChargesRateScheme:   m.ChargesRateSchemeManager.ToModel(data.ChargesRateScheme),
				AccountID:           data.AccountID,
				Account:             m.AccountManager.ToModel(data.Account),
			}
		},
		Created: func(data *ChargesRateSchemeAccount) []string {
			return []string{
				"charges_rate_scheme_account.create",
				fmt.Sprintf("charges_rate_scheme_account.create.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateSchemeAccount) []string {
			return []string{
				"charges_rate_scheme_account.update",
				fmt.Sprintf("charges_rate_scheme_account.update.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateSchemeAccount) []string {
			return []string{
				"charges_rate_scheme_account.delete",
				fmt.Sprintf("charges_rate_scheme_account.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_scheme_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_scheme_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// ChargesRateSchemeAccountCurrentBranch retrieves all charges rate scheme accounts for the specified organization and branch
func (m *Core) ChargesRateSchemeAccountCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*ChargesRateSchemeAccount, error) {
	return m.ChargesRateSchemeAccountManager.Find(context, &ChargesRateSchemeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
