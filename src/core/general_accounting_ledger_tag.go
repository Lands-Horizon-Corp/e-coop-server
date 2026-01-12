package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneralAccountingLedgerTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneralLedgerID uuid.UUID      `gorm:"type:uuid;not null"`
		GeneralLedger   *GeneralLedger `gorm:"foreignKey:GeneralLedgerID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"general_ledger,omitempty"`

		Name        string      `gorm:"type:varchar(50)"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	GeneralAccountingLedgerTagResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		GeneralLedgerID uuid.UUID              `json:"general_ledger_id"`
		GeneralLedger   *GeneralLedgerResponse `json:"general_ledger,omitempty"`
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		Category        TagCategory            `json:"category"`
		Color           string                 `json:"color"`
		Icon            string                 `json:"icon"`
	}

	GeneralAccountingLedgerTagRequest struct {
		GeneralLedgerID uuid.UUID   `json:"general_ledger_id" validate:"required"`
		Name            string      `json:"name" validate:"required,min=1,max=50"`
		Description     string      `json:"description,omitempty"`
		Category        TagCategory `json:"category,omitempty"`
		Color           string      `json:"color,omitempty"`
		Icon            string      `json:"icon,omitempty"`
	}
)

func GeneralAccountingLedgerTagManager(service *horizon.HorizonService) *registry.Registry[GeneralAccountingLedgerTag, GeneralAccountingLedgerTagResponse, GeneralAccountingLedgerTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		GeneralAccountingLedgerTag, GeneralAccountingLedgerTagResponse, GeneralAccountingLedgerTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "GeneralLedger",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GeneralAccountingLedgerTag) *GeneralAccountingLedgerTagResponse {
			if data == nil {
				return nil
			}
			return &GeneralAccountingLedgerTagResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				GeneralLedgerID: data.GeneralLedgerID,
				GeneralLedger:   GeneralLedgerManager(service).ToModel(data.GeneralLedger),
				Name:            data.Name,
				Description:     data.Description,
				Category:        data.Category,
				Color:           data.Color,
				Icon:            data.Icon,
			}
		},
		Created: func(data *GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.create",
				fmt.Sprintf("general_ledger_tag.create.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.update",
				fmt.Sprintf("general_ledger_tag.update.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralAccountingLedgerTag) registry.Topics {
			return []string{
				"general_ledger_tag.delete",
				fmt.Sprintf("general_ledger_tag.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GeneralAccountingLedgerTagCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*GeneralAccountingLedgerTag, error) {
	return GeneralAccountingLedgerTagManager(service).Find(context, &GeneralAccountingLedgerTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
