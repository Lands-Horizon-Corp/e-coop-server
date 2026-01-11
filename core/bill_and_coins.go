package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	BillAndCoins struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_bill_and_coins;uniqueIndex:idx_unique_name_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_bill_and_coins;uniqueIndex:idx_unique_name_org_branch" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"media,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null" json:"currency_id"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name  string  `gorm:"type:varchar(255);uniqueIndex:idx_unique_name_org_branch" json:"name"`
		Value float64 `gorm:"type:decimal;not null" json:"value"`
	}

	BillAndCoinsResponse struct {
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
		MediaID        *uuid.UUID            `json:"media_id,omitempty"`
		Media          *MediaResponse        `json:"media,omitempty"`
		CurrencyID     uuid.UUID             `json:"currency_id"`
		Currency       *CurrencyResponse     `json:"currency,omitempty"`
		Name           string                `json:"name"`
		Value          float64               `json:"value"`
	}

	BillAndCoinsRequest struct {
		Name       string     `json:"name" validate:"required,min=1,max=255"`
		Value      float64    `json:"value" validate:"required"`
		CurrencyID uuid.UUID  `json:"currency_id" validate:"required"`
		MediaID    *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *Core) BillAndCoinsManager() *registry.Registry[BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "Currency"},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *BillAndCoins) *BillAndCoinsResponse {
			if data == nil {
				return nil
			}
			return &BillAndCoinsResponse{
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
				MediaID:        data.MediaID,
				Media:          m.MediaManager().ToModel(data.Media),
				CurrencyID:     data.CurrencyID,
				Currency:       m.CurrencyManager().ToModel(data.Currency),
				Name:           data.Name,
				Value:          data.Value,
			}
		},
		Created: func(data *BillAndCoins) registry.Topics {
			return []string{
				"bill_and_coins.create",
				fmt.Sprintf("bill_and_coins.create.%s", data.ID),
				fmt.Sprintf("bill_and_coins.create.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *BillAndCoins) registry.Topics {
			return []string{
				"bill_and_coins.update",
				fmt.Sprintf("bill_and_coins.update.%s", data.ID),
				fmt.Sprintf("bill_and_coins.update.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *BillAndCoins) registry.Topics {
			return []string{
				"bill_and_coins.delete",
				fmt.Sprintf("bill_and_coins.delete.%s", data.ID),
				fmt.Sprintf("bill_and_coins.delete.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) billAndCoinsSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	curency, err := m.CurrencyManager().List(context)
	if err != nil {
		return eris.Wrap(err, "failed to list currencies for bill and coins seeding")
	}
	if len(curency) == 0 {
		return eris.New("no currencies found for bill and coins seeding")
	}
	for _, currency := range curency {
		billAndCoins := []*BillAndCoins{}
		switch currency.ISO3166Alpha3 {
		case "PHL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 1000 Bill", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 500 Bill", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 200 Bill", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 100 Bill", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 50 Bill", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 20 Bill", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.25 Sentimo Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.05 Sentimo Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.1 Sentimo Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "USA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 100 Bill", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 50 Bill", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 20 Bill", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 10 Bill", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 5 Bill", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 2 Bill", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 1 Bill", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 0.50 Half Dollar Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 0.25 Quarter Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 0.10 Dime Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 0.05 Nickel Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "US$ 0.01 Penny Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "DEU":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.50 Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.20 Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.10 Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.05 Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.02 Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "€0.01 Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "HRV":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.50 Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.20 Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.10 Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.05 Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.02 Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "EUR-HR €0.01 Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "JPN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JPY ¥1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "GBR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "£1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50p Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20p Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10p Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5p Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2p Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1p Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "AUS":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "A$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AUD 50c Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AUD 20c Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AUD 10c Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AUD 5c Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "CAN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$2 Coin (Toonie)", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "C$1 Coin (Loonie)", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25¢ Coin (Quarter)", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10¢ Coin (Dime)", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5¢ Coin (Nickel)", Value: 0.05, CurrencyID: currency.ID},
			}
		case "CHE":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 50 Rappen Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CHF 20 Rappen Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Rappen Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Rappen Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "CHN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY ¥1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY 5 Jiao Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY 1 Jiao Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY 5 Fen Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CNY 1 Fen Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "SWE":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SEK 1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "NZL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZ$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZD 50c Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZD 20c Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NZD 10c Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "IND":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₹1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "KOR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₩10 Coin", Value: 10.00, CurrencyID: currency.ID},
			}
		case "THA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "฿1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "THB 50 Satang Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "THB 25 Satang Coin", Value: 0.25, CurrencyID: currency.ID},
			}
		case "SGP":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$2 Banknote", Value: 2.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SGD 50¢ Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SGD 20¢ Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SGD 10¢ Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "SGD 5¢ Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "HKG":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HK$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HKD 50¢ Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HKD 20¢ Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "HKD 10¢ Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "MYS":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RM1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MYR 50 Sen Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MYR 20 Sen Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MYR 10 Sen Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MYR 5 Sen Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "IDN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp1,000 Coin", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp200 Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rp100 Coin", Value: 100.00, CurrencyID: currency.ID},
			}
		case "VNM":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫500,000 Banknote", Value: 500000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫200,000 Banknote", Value: 200000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫5,000 Coin", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫2,000 Coin", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫1,000 Coin", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₫500 Coin", Value: 500.00, CurrencyID: currency.ID},
			}
		case "TWN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$100 Banknote", Value: 100.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NT$1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "BRN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$25 Banknote", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "B$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BND 50 Sen Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BND 20 Sen Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BND 10 Sen Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BND 5 Sen Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "SAU":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.س 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Halala Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Halala Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Halala Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Halala Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "ARE":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.إ 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AED 50 Fils Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AED 25 Fils Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AED 10 Fils Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "ISR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₪1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Agorot Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Agorot Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "ZAF":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "R1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ZAR 50c Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ZAR 20c Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ZAR 10c Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "EGY":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ج.م 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Piastres Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Piastres Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Piastres Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "TUR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₺1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Kuruş Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Kuruş Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Kuruş Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Kuruş Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "BFA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 500 Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 250 Coin", Value: 250.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 25 Coin", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XOF CFA 10 Coin", Value: 10.00, CurrencyID: currency.ID},
			}
		case "CMR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 500 Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 25 Coin", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "XAF CFA 5 Coin", Value: 5.00, CurrencyID: currency.ID},
			}
		case "MUS":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨25 Banknote", Value: 25.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR ₨1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR 50 Cents Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MUR 20 Cents Coin", Value: 0.20, CurrencyID: currency.ID},
			}
		case "MDV":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Rf1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Laari Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Laari Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Laari Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "NOR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 1000 kr Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 500 kr Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 200 kr Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 100 kr Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 50 kr Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 20 kr Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 10 kr Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 5 kr Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NOK 1 kr Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "DNK":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 1000 kr Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 500 kr Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 200 kr Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 100 kr Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 50 kr Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 20 kr Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 10 kr Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 5 kr Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 2 kr Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 1 kr Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DKK 50 øre Coin", Value: 0.50, CurrencyID: currency.ID},
			}
		case "POL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 zł Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "200 zł Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 zł Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 zł Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 zł Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 zł Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 zł Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 zł Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 zł Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 groszy Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 groszy Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 groszy Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 groszy Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 groszy Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 grosz Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "CZE":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5000 Kč Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2000 Kč Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1000 Kč Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 Kč Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "200 Kč Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 Kč Banknote", Value: 100.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Kč Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Kč Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Kč Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Kč Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 Kč Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Kč Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "HUN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20,000 Ft Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10,000 Ft Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5,000 Ft Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2,000 Ft Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1,000 Ft Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 Ft Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "200 Ft Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 Ft Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Ft Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Ft Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Ft Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Ft Coin", Value: 5.00, CurrencyID: currency.ID},
			}
		case "RUS":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽5000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽50 Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽25 Coin", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₽1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Kopecks Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Kopecks Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Kopecks Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Kopeck Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "BRA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$2 Banknote", Value: 2.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL R$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL 25 Centavos Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL 10 Centavos Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BRL 5 Centavos Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "MEX":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MX$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MXN 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MXN 20 Centavos Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "MXN 10 Centavos Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "ARG":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "AR$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ARS 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ARS 25 Centavos Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ARS 10 Centavos Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "CHL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CL$10 Coin", Value: 10.00, CurrencyID: currency.ID},
			}
		case "COL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$1,000 Coin", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$200 Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "CO$50 Coin", Value: 50.00, CurrencyID: currency.ID},
			}
		case "PER":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "S/1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PEN 50 Céntimos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PEN 20 Céntimos Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PEN 10 Céntimos Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "URY":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "$U1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Centésimos Coin", Value: 0.50, CurrencyID: currency.ID},
			}
		case "DOM":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$25 Coin", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "RD$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DOP 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "DOP 25 Centavos Coin", Value: 0.25, CurrencyID: currency.ID},
			}
		case "PRY":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲1,000 Coin", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₲50 Coin", Value: 50.00, CurrencyID: currency.ID},
			}
		case "BOL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BOB 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BOB 20 Centavos Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BOB 10 Centavos Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "VEN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 1,000,000 Banknote", Value: 1000000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 500,000 Banknote", Value: 500000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 200,000 Banknote", Value: 200000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bs.S 1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Céntimos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Céntimos Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Céntimos Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "PAK":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨5000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR ₨1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "PKR 50 Paisa Coin", Value: 0.50, CurrencyID: currency.ID},
			}
		case "BGD":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳2 Banknote", Value: 2.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT ৳1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT 50 Paisa Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BDT 25 Paisa Coin", Value: 0.25, CurrencyID: currency.ID},
			}
		case "LKA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs5000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR Rs1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR 50 Cents Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "LKR 25 Cents Coin", Value: 0.25, CurrencyID: currency.ID},
			}
		case "NPL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR Rs1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR 50 Paisa Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "NPR 25 Paisa Coin", Value: 0.25, CurrencyID: currency.ID},
			}
		case "MMR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K10,000 PG Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K5,000 PG Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K1,000 PG Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K500 PG Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K200 PG Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K100 PG Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K50 PG Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K100 PG Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K50 PG Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K10 PG Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K5 PG Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K1 PG Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "KHM":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛500 Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛200 Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "៛50 Coin", Value: 50.00, CurrencyID: currency.ID},
			}
		case "LAO":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭100,000 Banknote", Value: 100000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭50,000 Banknote", Value: 50000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭500 Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭5,000 Coin", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭1,000 Coin", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₭50 Coin", Value: 50.00, CurrencyID: currency.ID},
			}
		case "NGA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₦1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Kobo Coin", Value: 0.50, CurrencyID: currency.ID},
			}
		case "KEN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh50 Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh40 Coin", Value: 40.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KSh1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KES 50 Cents Coin", Value: 0.50, CurrencyID: currency.ID},
			}
		case "GHA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵2 Banknote", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₵1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Pesewas Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Pesewas Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Pesewas Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Pesewas Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "MAR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.20 Banknote", Value: 20.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.م.1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Centimes Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Centimes Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Centimes Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "TUN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت30 Banknote", Value: 30.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ت1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 Millimes Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 Millimes Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Millimes Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "ETH":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Bir1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Bir Santim Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Bir Santim Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Bir Santim Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Bir Santim Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Bir Santim Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "DZA":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج100 Banknote", Value: 100.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج200 Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ج5 Coin", Value: 5.00, CurrencyID: currency.ID},
			}
		case "UKR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₴1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Kopiyok Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Kopiyok Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Kopiyok Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Kopiyok Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 Kopiyky Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Kopiyky Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "ROU":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 lei Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "200 lei Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 lei Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 lei Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 lei Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 lei Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 leu Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 bani Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 bani Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 bani Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 ban Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "BGR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "лв1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 стотинки Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 стотинки Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 стотинки Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 стотинки Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 стотинки Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 стотинка Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "SRB":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин5000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин2000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин10 Banknote", Value: 10.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "дин1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "ISL":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10,000 kr Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5,000 kr Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2,000 kr Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1,000 kr Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 kr Banknote", Value: 500.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 kr Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 kr Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 kr Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 kr Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 kr Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		case "BLR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Br5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 Br2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 Br1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Br Kopecks Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Br Kopecks Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Br Kopecks Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Br Kopecks Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "2 Br Kopecks Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "1 Br Kopeck Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "FJI":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$5 Banknote", Value: 5.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJ$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJD 50 Cents Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJD 20 Cents Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJD 10 Cents Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "FJD 5 Cents Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "PNG":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K2 Banknote", Value: 2.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "K1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Toea Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "20 Toea Coin", Value: 0.20, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Toea Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Toea Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "JAM":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$5000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$1000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$50 Banknote", Value: 50.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "J$1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JMD 50 Cents Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JMD 25 Cents Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JMD 10 Cents Coin", Value: 0.10, CurrencyID: currency.ID},
			}
		case "CRI":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡500 Coin", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡25 Coin", Value: 25.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₡5 Coin", Value: 5.00, CurrencyID: currency.ID},
			}
		case "GTM":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q200 Banknote", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "Q1 Coin", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "GTQ 50 Centavos Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "GTQ 25 Centavos Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "GTQ 10 Centavos Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Centavos Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "KWT":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 1 Banknote", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 1/2 Banknote", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ك 1/4 Banknote", Value: 0.25, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KWD 100 Fils Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KWD 50 Fils Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KWD 20 Fils Coin", Value: 0.02, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KWD 10 Fils Coin", Value: 0.01, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "KWD 5 Fils Coin", Value: 0.005, CurrencyID: currency.ID},
			}
		case "QAT":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 100 Banknote", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ق 1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Dirhams Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Dirhams Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Dirhams Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "5 Dirhams Coin", Value: 0.05, CurrencyID: currency.ID},
			}
		case "OMN":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ع 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ع 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ع 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ع 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ر.ع 1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "500 Baisa Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "100 Baisa Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "50 Baisa Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "25 Baisa Coin", Value: 0.025, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "10 Baisa Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "BHR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ب.د 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ب.د 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ب.د 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ب.د 1 Banknote", Value: 1.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "ب.د 1/2 Banknote", Value: 0.50, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BHD 100 Fils Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BHD 50 Fils Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BHD 25 Fils Coin", Value: 0.025, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BHD 10 Fils Coin", Value: 0.01, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "BHD 5 Fils Coin", Value: 0.005, CurrencyID: currency.ID},
			}
		case "JOR":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 50 Banknote", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 20 Banknote", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 10 Banknote", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 5 Banknote", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 1 Banknote", Value: 1.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 1/2 Coin", Value: 0.50, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "د.ا 1/4 Coin", Value: 0.25, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JOD 100 Fils Coin", Value: 0.10, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JOD 50 Fils Coin", Value: 0.05, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JOD 25 Fils Coin", Value: 0.025, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "JOD 10 Fils Coin", Value: 0.01, CurrencyID: currency.ID},
			}
		case "KAZ":
			billAndCoins = []*BillAndCoins{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸20,000 Banknote", Value: 20000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸10,000 Banknote", Value: 10000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸5,000 Banknote", Value: 5000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸2,000 Banknote", Value: 2000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸1,000 Banknote", Value: 1000.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸500 Banknote", Value: 500.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸200 Banknote", Value: 200.00, CurrencyID: currency.ID},

				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸200 Coin", Value: 200.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸100 Coin", Value: 100.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸50 Coin", Value: 50.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸20 Coin", Value: 20.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸10 Coin", Value: 10.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸5 Coin", Value: 5.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸2 Coin", Value: 2.00, CurrencyID: currency.ID},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₸1 Coin", Value: 1.00, CurrencyID: currency.ID},
			}
		}
		for _, data := range billAndCoins {
			if err := m.BillAndCoinsManager().CreateWithTx(context, tx, data); err != nil {
				return eris.Wrapf(err, "failed to seed bill or coin %s", data.Name)
			}
		}
	}

	return nil
}

func (m *Core) BillAndCoinsCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*BillAndCoins, error) {
	return m.BillAndCoinsManager().Find(context, &BillAndCoins{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
