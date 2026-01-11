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
	InterestRateByYear struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_year"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_year"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BrowseReferenceID uuid.UUID        `gorm:"type:uuid;not null;index:idx_browse_reference_year_range"`
		BrowseReference   *BrowseReference `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_reference,omitempty"`

		FromYear     int     `gorm:"not null;index:idx_browse_reference_year_range" json:"from_year" validate:"required,min=1"`
		ToYear       int     `gorm:"not null;index:idx_browse_reference_year_range" json:"to_year" validate:"required,min=1"`
		InterestRate float64 `gorm:"type:decimal(15,6);not null" json:"interest_rate" validate:"required,min=0"`
	}

	InterestRateByYearResponse struct {
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
		FromYear          int                      `json:"from_year"`
		ToYear            int                      `json:"to_year"`
		InterestRate      float64                  `json:"interest_rate"`
	}

	InterestRateByYearRequest struct {
		ID                *uuid.UUID `json:"id"`
		BrowseReferenceID uuid.UUID  `json:"browse_reference_id" validate:"required"`
		FromYear          int        `json:"from_year" validate:"required,min=1"`
		ToYear            int        `json:"to_year" validate:"required,min=1,gtefield=FromYear"`
		InterestRate      float64    `json:"interest_rate" validate:"required,min=0"`
	}
)

func (m *Core) InterestRateByYearManager() *registry.Registry[InterestRateByYear, InterestRateByYearResponse, InterestRateByYearRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		InterestRateByYear, InterestRateByYearResponse, InterestRateByYearRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *InterestRateByYear) *InterestRateByYearResponse {
			if data == nil {
				return nil
			}
			return &InterestRateByYearResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),

				BrowseReferenceID: data.BrowseReferenceID,
				BrowseReference:   m.BrowseReferenceManager().ToModel(data.BrowseReference),
				FromYear:          data.FromYear,
				ToYear:            data.ToYear,
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.create",
				fmt.Sprintf("interest_rate_by_year.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.update",
				fmt.Sprintf("interest_rate_by_year.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRateByYear) registry.Topics {
			return []string{
				"interest_rate_by_year.delete",
				fmt.Sprintf("interest_rate_by_year.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_year.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_year.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_year.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) InterestRateByYearForBrowseReference(context context.Context, browseReferenceID uuid.UUID) ([]*InterestRateByYear, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return m.InterestRateByYearManager().ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByYearForRange(context context.Context, browseReferenceID uuid.UUID, year int) ([]*InterestRateByYear, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_year", Op: query.ModeLTE, Value: year},
		{Field: "to_year", Op: query.ModeGTE, Value: year},
	}

	return m.InterestRateByYearManager().ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByYearCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*InterestRateByYear, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.InterestRateByYearManager().ArrFind(context, filters, nil)
}

func (m *Core) GetInterestRateForYear(context context.Context, browseReferenceID uuid.UUID, year int) (*InterestRateByYear, error) {
	rates, err := m.InterestRateByYearForRange(context, browseReferenceID, year)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
