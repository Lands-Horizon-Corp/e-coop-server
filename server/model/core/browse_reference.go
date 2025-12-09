package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InterestType represents the type of interest calculation
type InterestType string

// Interest type constants
const (
	InterestTypeYear   InterestType = "year"
	InterestTypeDate   InterestType = "date"
	InterestTypeAmount InterestType = "amount"
	InterestTypeNone   InterestType = "none"
)

type (
	// BrowseReference represents a reference configuration for browsing accounts
	BrowseReference struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_reference"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_reference"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name" validate:"required,max=255"`
		Description string `gorm:"type:text" json:"description"`

		InterestRate   float64 `gorm:"type:decimal(15,6);default:0" json:"interest_rate"`
		MinimumBalance float64 `gorm:"type:decimal(15,2);default:0" json:"minimum_balance"`
		Charges        float64 `gorm:"type:decimal(15,2);default:0" json:"charges"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		MemberTypeID *uuid.UUID  `gorm:"type:uuid"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		InterestType InterestType `gorm:"type:varchar(20);not null;default:'year'" json:"interest_type" validate:"required,oneof=year date amount"`

		DefaultMinimumBalance float64 `gorm:"type:decimal(15,2);default:0" json:"default_minimum_balance"`
		DefaultInterestRate   float64 `gorm:"type:decimal(15,6);default:0" json:"default_interest_rate"`

		// Relationships
		InterestRatesByYear   []*InterestRateByYear   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDate   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmount `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_amount,omitempty"`
	}

	// BrowseReferenceResponse represents the response structure for browse reference data
	BrowseReferenceResponse struct {
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

		Name                  string              `json:"name"`
		Description           string              `json:"description"`
		InterestRate          float64             `json:"interest_rate"`
		MinimumBalance        float64             `json:"minimum_balance"`
		Charges               float64             `json:"charges"`
		AccountID             *uuid.UUID          `json:"account_id,omitempty"`
		Account               *AccountResponse    `json:"account,omitempty"`
		MemberTypeID          *uuid.UUID          `json:"member_type_id,omitempty"`
		MemberType            *MemberTypeResponse `json:"member_type,omitempty"`
		InterestType          InterestType        `json:"interest_type"`
		DefaultMinimumBalance float64             `json:"default_minimum_balance"`
		DefaultInterestRate   float64             `json:"default_interest_rate"`

		InterestRatesByYear   []*InterestRateByYearResponse   `json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDateResponse   `json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmountResponse `json:"interest_rates_by_amount,omitempty"`
	}

	// BrowseReferenceRequest represents the request structure for creating/updating browse references
	BrowseReferenceRequest struct {
		ID          *uuid.UUID `json:"id"`
		Name        string     `json:"name" validate:"required,max=255"`
		Description string     `json:"description"`

		InterestRate   float64 `json:"interest_rate"`
		MinimumBalance float64 `json:"minimum_balance"`
		Charges        float64 `json:"charges"`

		AccountID    *uuid.UUID   `json:"account_id"`
		MemberTypeID *uuid.UUID   `json:"member_type_id"`
		InterestType InterestType `json:"interest_type" validate:"required,oneof=year date amount none"`

		DefaultMinimumBalance float64 `json:"default_minimum_balance"`
		DefaultInterestRate   float64 `json:"default_interest_rate"`

		// Nested relationships for creation/update
		InterestRatesByYear   []*InterestRateByYearRequest   `json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDateRequest   `json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmountRequest `json:"interest_rates_by_amount,omitempty"`

		// Delete IDs for nested relationships
		InterestRatesByYearDeleted   uuid.UUIDs `json:"interest_rates_by_year_deleted,omitempty"`
		InterestRatesByDateDeleted   uuid.UUIDs `json:"interest_rates_by_date_deleted,omitempty"`
		InterestRatesByAmountDeleted uuid.UUIDs `json:"interest_rates_by_amount_deleted,omitempty"`
	}
)

func (m *Core) browseReference() {
	m.Migration = append(m.Migration, &BrowseReference{})
	m.BrowseReferenceManager = *registry.NewRegistry(registry.RegistryParams[
		BrowseReference, BrowseReferenceResponse, BrowseReferenceRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "Account", "Account.Currency",
			"MemberType", "InterestRatesByYear", "InterestRatesByDate", "InterestRatesByAmount",
		},
		Service: m.provider.Service,
		Resource: func(data *BrowseReference) *BrowseReferenceResponse {
			if data == nil {
				return nil
			}
			return &BrowseReferenceResponse{
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

				Name:                  data.Name,
				Description:           data.Description,
				InterestRate:          data.InterestRate,
				MinimumBalance:        data.MinimumBalance,
				Charges:               data.Charges,
				AccountID:             data.AccountID,
				Account:               m.AccountManager.ToModel(data.Account),
				MemberTypeID:          data.MemberTypeID,
				MemberType:            m.MemberTypeManager.ToModel(data.MemberType),
				InterestType:          data.InterestType,
				DefaultMinimumBalance: data.DefaultMinimumBalance,
				DefaultInterestRate:   data.DefaultInterestRate,
				InterestRatesByYear:   m.InterestRateByYearManager.ToModels(data.InterestRatesByYear),
				InterestRatesByDate:   m.InterestRateByDateManager.ToModels(data.InterestRatesByDate),
				InterestRatesByAmount: m.InterestRateByAmountManager.ToModels(data.InterestRatesByAmount),
			}
		},

		Created: func(data *BrowseReference) []string {
			return []string{
				"browse_reference.create",
				fmt.Sprintf("browse_reference.create.%s", data.ID),
				fmt.Sprintf("browse_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *BrowseReference) []string {
			return []string{
				"browse_reference.update",
				fmt.Sprintf("browse_reference.update.%s", data.ID),
				fmt.Sprintf("browse_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *BrowseReference) []string {
			return []string{
				"browse_reference.delete",
				fmt.Sprintf("browse_reference.delete.%s", data.ID),
				fmt.Sprintf("browse_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// BrowseReferenceCurrentBranch retrieves browse references for the specified branch and organization
func (m *Core) BrowseReferenceCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	return m.BrowseReferenceManager.ArrFind(context, filters, nil)
}

// BrowseReferenceByMemberType retrieves browse references for a specific member type
func (m *Core) BrowseReferenceByMemberType(context context.Context, memberTypeID, organizationID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "member_type_id", Op: registry.OpEq, Value: memberTypeID},
	}

	return m.BrowseReferenceManager.ArrFind(context, filters, nil)
}

// BrowseReferenceByInterestType retrieves browse references by interest type
func (m *Core) BrowseReferenceByInterestType(context context.Context, interestType InterestType, organizationID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
		{Field: "interest_type", Op: registry.OpEq, Value: string(interestType)},
	}

	return m.BrowseReferenceManager.ArrFind(context, filters, nil)
}

func (m *Core) BrowseReferenceByField(
	context context.Context, organizationID, branchID uuid.UUID, accountID, memberTypeID *uuid.UUID,
) ([]*BrowseReference, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: registry.OpEq, Value: organizationID},
		{Field: "branch_id", Op: registry.OpEq, Value: branchID},
	}

	// Add filters based on provided parameters
	if memberTypeID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "member_type_id", Op: registry.OpEq, Value: *memberTypeID,
		})
	}

	if accountID != nil {
		filters = append(filters, registry.FilterSQL{
			Field: "account_id", Op: registry.OpEq, Value: *accountID,
		})
	}

	return m.BrowseReferenceManager.ArrFind(context, filters, nil)
}
