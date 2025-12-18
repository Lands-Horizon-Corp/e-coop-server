package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	InterestRateByAmount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_amount"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_amount"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BrowseReferenceID uuid.UUID        `gorm:"type:uuid;not null;index:idx_browse_reference_amount_range"`
		BrowseReference   *BrowseReference `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_reference,omitempty"`

		FromAmount   float64 `gorm:"type:decimal(15,2);not null;index:idx_browse_reference_amount_range" json:"from_amount" validate:"required,min=0"`
		ToAmount     float64 `gorm:"type:decimal(15,2);not null;index:idx_browse_reference_amount_range" json:"to_amount" validate:"required,min=0"`
		InterestRate float64 `gorm:"type:decimal(15,6);not null" json:"interest_rate" validate:"required,min=0"`
	}

	InterestRateByAmountResponse struct {
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

		BrowseReferenceID uuid.UUID                `json:"browse_reference_id"`
		BrowseReference   *BrowseReferenceResponse `json:"browse_reference,omitempty"`
		FromAmount        float64                  `json:"from_amount"`
		ToAmount          float64                  `json:"to_amount"`
		InterestRate      float64                  `json:"interest_rate"`
	}

	InterestRateByAmountRequest struct {
		ID                *uuid.UUID `json:"id"`
		BrowseReferenceID uuid.UUID  `json:"browse_reference_id" validate:"required"`
		FromAmount        float64    `json:"from_amount" validate:"required,min=0"`
		ToAmount          float64    `json:"to_amount" validate:"required,min=0,gtefield=FromAmount"`
		InterestRate      float64    `json:"interest_rate" validate:"required,min=0"`
	}
)

func (m *Core) interestRateByAmount() {
	m.Migration = append(m.Migration, &InterestRateByAmount{})
	m.InterestRateByAmountManager = *registry.NewRegistry(registry.RegistryParams[
		InterestRateByAmount, InterestRateByAmountResponse, InterestRateByAmountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *InterestRateByAmount) *InterestRateByAmountResponse {
			if data == nil {
				return nil
			}
			return &InterestRateByAmountResponse{
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

				BrowseReferenceID: data.BrowseReferenceID,
				BrowseReference:   m.BrowseReferenceManager.ToModel(data.BrowseReference),
				FromAmount:        data.FromAmount,
				ToAmount:          data.ToAmount,
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.create",
				fmt.Sprintf("interest_rate_by_amount.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.update",
				fmt.Sprintf("interest_rate_by_amount.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRateByAmount) registry.Topics {
			return []string{
				"interest_rate_by_amount.delete",
				fmt.Sprintf("interest_rate_by_amount.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_amount.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_amount.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_amount.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) InterestRateByAmountForBrowseReference(context context.Context, browseReferenceID uuid.UUID) ([]*InterestRateByAmount, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return m.InterestRateByAmountManager.ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByAmountForRange(context context.Context, browseReferenceID uuid.UUID, amount float64) ([]*InterestRateByAmount, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_amount", Op: query.ModeLTE, Value: amount},
		{Field: "to_amount", Op: query.ModeGTE, Value: amount},
	}

	return m.InterestRateByAmountManager.ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByAmountCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*InterestRateByAmount, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.InterestRateByAmountManager.ArrFind(context, filters, nil)
}

func (m *Core) GetInterestRateForAmount(context context.Context, browseReferenceID uuid.UUID, amount float64) (*InterestRateByAmount, error) {
	rates, err := m.InterestRateByAmountForRange(context, browseReferenceID, amount)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
