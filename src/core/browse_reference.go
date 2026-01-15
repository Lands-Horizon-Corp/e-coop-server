package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InterestType string

const (
	InterestTypeYear   InterestType = "year"
	InterestTypeDate   InterestType = "date"
	InterestTypeAmount InterestType = "amount"
	InterestTypeNone   InterestType = "none"
)

type (
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

		InterestRatesByYear   []*InterestRateByYear   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDate   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmount `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_amount,omitempty"`
	}

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

		InterestRatesByYear   []*InterestRateByYearRequest   `json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDateRequest   `json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmountRequest `json:"interest_rates_by_amount,omitempty"`

		InterestRatesByYearDeleted   uuid.UUIDs `json:"interest_rates_by_year_deleted,omitempty"`
		InterestRatesByDateDeleted   uuid.UUIDs `json:"interest_rates_by_date_deleted,omitempty"`
		InterestRatesByAmountDeleted uuid.UUIDs `json:"interest_rates_by_amount_deleted,omitempty"`
	}
)

func BrowseReferenceManager(service *horizon.HorizonService) *registry.Registry[BrowseReference, BrowseReferenceResponse, BrowseReferenceRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		BrowseReference, BrowseReferenceResponse, BrowseReferenceRequest,
	]{
		Preloads: []string{
			"Account", "Account.Currency",
			"MemberType", "InterestRatesByYear", "InterestRatesByDate", "InterestRatesByAmount",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *BrowseReference) *BrowseReferenceResponse {
			if data == nil {
				return nil
			}
			return &BrowseReferenceResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				Name:                  data.Name,
				Description:           data.Description,
				InterestRate:          data.InterestRate,
				MinimumBalance:        data.MinimumBalance,
				Charges:               data.Charges,
				AccountID:             data.AccountID,
				Account:               AccountManager(service).ToModel(data.Account),
				MemberTypeID:          data.MemberTypeID,
				MemberType:            MemberTypeManager(service).ToModel(data.MemberType),
				InterestType:          data.InterestType,
				DefaultMinimumBalance: data.DefaultMinimumBalance,
				DefaultInterestRate:   data.DefaultInterestRate,
				InterestRatesByYear:   InterestRateByYearManager(service).ToModels(data.InterestRatesByYear),
				InterestRatesByDate:   InterestRateByDateManager(service).ToModels(data.InterestRatesByDate),
				InterestRatesByAmount: InterestRateByAmountManager(service).ToModels(data.InterestRatesByAmount),
			}
		},

		Created: func(data *BrowseReference) registry.Topics {
			return []string{
				"browse_reference.create",
				fmt.Sprintf("browse_reference.create.%s", data.ID),
				fmt.Sprintf("browse_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *BrowseReference) registry.Topics {
			return []string{
				"browse_reference.update",
				fmt.Sprintf("browse_reference.update.%s", data.ID),
				fmt.Sprintf("browse_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *BrowseReference) registry.Topics {
			return []string{
				"browse_reference.delete",
				fmt.Sprintf("browse_reference.delete.%s", data.ID),
				fmt.Sprintf("browse_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("browse_reference.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func BrowseReferenceCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByMemberType(context context.Context, service *horizon.HorizonService, memberTypeID, organizationID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "member_type_id", Op: query.ModeEqual, Value: memberTypeID},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByInterestType(context context.Context, service *horizon.HorizonService, interestType InterestType, organizationID, branchID uuid.UUID) ([]*BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "interest_type", Op: query.ModeEqual, Value: string(interestType)},
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}

func BrowseReferenceByField(
	context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID, accountID, memberTypeID *uuid.UUID,
) ([]*BrowseReference, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	if memberTypeID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "member_type_id", Op: query.ModeEqual, Value: *memberTypeID,
		})
	}

	if accountID != nil {
		filters = append(filters, query.ArrFilterSQL{
			Field: "account_id", Op: query.ModeEqual, Value: *accountID,
		})
	}

	return BrowseReferenceManager(service).ArrFind(context, filters, nil)
}
