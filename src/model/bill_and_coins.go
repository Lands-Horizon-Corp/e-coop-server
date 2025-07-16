package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
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

		Name        string  `gorm:"type:varchar(255);uniqueIndex:idx_unique_name_org_branch" json:"name"`
		Value       float64 `gorm:"type:decimal;not null" json:"value"`
		CountryCode string  `gorm:"type:varchar(5);not null" json:"country_code"`
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
		Name           string                `json:"name"`
		Value          float64               `json:"value"`
		CountryCode    string                `json:"country_code"`
	}

	BillAndCoinsRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Value       float64    `json:"value" validate:"required"`
		CountryCode string     `json:"country_code" validate:"required,min=1,max=5"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *Model) BillAndCoins() {
	m.Migration = append(m.Migration, &BillAndCoins{})
	m.BillAndCoinsManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		BillAndCoins, BillAndCoinsResponse, BillAndCoinsRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "Media"},
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
				Name:           data.Name,
				Value:          data.Value,
				CountryCode:    data.CountryCode,
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

func (m *Model) BillAndCoinsCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*BillAndCoins, error) {
	return m.BillAndCoinsManager.Find(context, &BillAndCoins{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
