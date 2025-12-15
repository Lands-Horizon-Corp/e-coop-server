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
	InterestRateByDate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_date"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_date"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BrowseReferenceID uuid.UUID        `gorm:"type:uuid;not null;index:idx_browse_reference_date_range"`
		BrowseReference   *BrowseReference `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"browse_reference,omitempty"`

		FromDate     time.Time `gorm:"not null;index:idx_browse_reference_date_range" json:"from_date" validate:"required"`
		ToDate       time.Time `gorm:"not null;index:idx_browse_reference_date_range" json:"to_date" validate:"required"`
		InterestRate float64   `gorm:"type:decimal(15,6);not null" json:"interest_rate" validate:"required,min=0"`
	}

	InterestRateByDateResponse struct {
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
		FromDate          string                   `json:"from_date"`
		ToDate            string                   `json:"to_date"`
		InterestRate      float64                  `json:"interest_rate"`
	}

	InterestRateByDateRequest struct {
		ID                *uuid.UUID `json:"id"`
		BrowseReferenceID uuid.UUID  `json:"browse_reference_id" validate:"required"`
		FromDate          time.Time  `json:"from_date" validate:"required"`
		ToDate            time.Time  `json:"to_date" validate:"required,gtefield=FromDate"`
		InterestRate      float64    `json:"interest_rate" validate:"required,min=0"`
	}
)

func (m *Core) interestRateByDate() {
	m.Migration = append(m.Migration, &InterestRateByDate{})
	m.InterestRateByDateManager = *registry.NewRegistry(registry.RegistryParams[
		InterestRateByDate, InterestRateByDateResponse, InterestRateByDateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Organization", "Branch", "BrowseReference",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *InterestRateByDate) *InterestRateByDateResponse {
			if data == nil {
				return nil
			}
			return &InterestRateByDateResponse{
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
				FromDate:          data.FromDate.Format(time.RFC3339),
				ToDate:            data.ToDate.Format(time.RFC3339),
				InterestRate:      data.InterestRate,
			}
		},

		Created: func(data *InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.create",
				fmt.Sprintf("interest_rate_by_date.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.create.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.update",
				fmt.Sprintf("interest_rate_by_date.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.update.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRateByDate) registry.Topics {
			return []string{
				"interest_rate_by_date.delete",
				fmt.Sprintf("interest_rate_by_date.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_date.delete.browse_reference.%s", data.BrowseReferenceID),
				fmt.Sprintf("interest_rate_by_date.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_date.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) InterestRateByDateForBrowseReference(context context.Context, browseReferenceID uuid.UUID) ([]*InterestRateByDate, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
	}

	return m.InterestRateByDateManager.ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByDateForRange(context context.Context, browseReferenceID uuid.UUID, date time.Time) ([]*InterestRateByDate, error) {
	filters := []registry.FilterSQL{
		{Field: "browse_reference_id", Op: query.ModeEqual, Value: browseReferenceID},
		{Field: "from_date", Op: query.ModeLTE, Value: date},
		{Field: "to_date", Op: query.ModeGTE, Value: date},
	}

	return m.InterestRateByDateManager.ArrFind(context, filters, nil)
}

func (m *Core) InterestRateByDateCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*InterestRateByDate, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.InterestRateByDateManager.ArrFind(context, filters, nil)
}

func (m *Core) GetInterestRateForDate(context context.Context, browseReferenceID uuid.UUID, date time.Time) (*InterestRateByDate, error) {
	rates, err := m.InterestRateByDateForRange(context, browseReferenceID, date)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	return rates[0], nil
}
