package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
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

func (m *ModelCore) BillAndCoins() {
	m.Migration = append(m.Migration, &BillAndCoins{})
	m.BillAndCoinsManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "Media", "Currency"},
		Service:  m.provider.Service,
		Resource: func(data *BillAndCoins) *BillAndCoinsResponse {
			if data == nil {
				return nil
			}
			return &BillAndCoinsResponse{
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
				MediaID:        data.MediaID,
				Media:          m.MediaManager.ToModel(data.Media),
				CurrencyID:     data.CurrencyID,
				Currency:       m.CurrencyManager.ToModel(data.Currency),
				Name:           data.Name,
				Value:          data.Value,
			}
		},
		Created: func(data *BillAndCoins) []string {
			return []string{
				"bill_and_coins.create",
				fmt.Sprintf("bill_and_coins.create.%s", data.ID),
				fmt.Sprintf("bill_and_coins.create.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *BillAndCoins) []string {
			return []string{
				"bill_and_coins.update",
				fmt.Sprintf("bill_and_coins.update.%s", data.ID),
				fmt.Sprintf("bill_and_coins.update.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *BillAndCoins) []string {
			return []string{
				"bill_and_coins.delete",
				fmt.Sprintf("bill_and_coins.delete.%s", data.ID),
				fmt.Sprintf("bill_and_coins.delete.branch.%s", data.BranchID),
				fmt.Sprintf("bill_and_coins.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) BillAndCoinsSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	// Find Philippine Peso currency
	phpCurrency, err := m.CurrencyManager.FindOne(context, &Currency{CurrencyCode: "PHP"})
	if err != nil {
		return eris.Wrap(err, "failed to find PHP currency for bill and coins seeding")
	}

	billAndCoins := []*BillAndCoins{
		// Banknotes (New Generation Currency Series)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 1000 Bill", Value: 1000.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 500 Bill", Value: 500.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 200 Bill", Value: 200.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 100 Bill", Value: 100.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 50 Bill", Value: 50.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 20 Bill", Value: 20.00, CurrencyID: phpCurrency.ID},

		// Coins (New Generation Currency Series)
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 20 Coin", Value: 20.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 10 Coin", Value: 10.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 5 Coin", Value: 5.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 1 Coin", Value: 1.00, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.25 Sentimo Coin", Value: 0.25, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.05 Sentimo Coin", Value: 0.05, CurrencyID: phpCurrency.ID},
		{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, Name: "₱ 0.1 Sentimo Coin", Value: 0.01, CurrencyID: phpCurrency.ID},
	}
	for _, data := range billAndCoins {
		if err := m.BillAndCoinsManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed bill or coin %s", data.Name)
		}
	}
	return nil
}
func (m *ModelCore) BillAndCoinsCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*BillAndCoins, error) {
	return m.BillAndCoinsManager.Find(context, &BillAndCoins{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
