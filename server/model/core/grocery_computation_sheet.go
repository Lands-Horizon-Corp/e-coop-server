package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// GroceryComputationSheet represents the GroceryComputationSheet model.
	GroceryComputationSheet struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_grocery_computation_sheet"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_grocery_computation_sheet"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		SchemeNumber int    `gorm:"not null;unique"`
		Description  string `gorm:"type:text"`
	}

	// GroceryComputationSheetResponse represents the response structure for grocerycomputationsheet data

	// GroceryComputationSheetResponse represents the response structure for GroceryComputationSheet.
	GroceryComputationSheetResponse struct {
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
		SchemeNumber   int                   `json:"scheme_number"`
		Description    string                `json:"description"`
	}

	// GroceryComputationSheetRequest represents the request structure for creating/updating grocerycomputationsheet

	// GroceryComputationSheetRequest represents the request structure for GroceryComputationSheet.
	GroceryComputationSheetRequest struct {
		SchemeNumber int    `json:"scheme_number" validate:"required"`
		Description  string `json:"description,omitempty"`
	}
)

func (m *Core) groceryComputationSheet() {
	m.Migration = append(m.Migration, &GroceryComputationSheet{})
	m.GroceryComputationSheetManager = *registry.NewRegistry(registry.RegistryParams[
		GroceryComputationSheet, GroceryComputationSheetResponse, GroceryComputationSheetRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Service: m.provider.Service,
		Resource: func(data *GroceryComputationSheet) *GroceryComputationSheetResponse {
			if data == nil {
				return nil
			}
			return &GroceryComputationSheetResponse{
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
				SchemeNumber:   data.SchemeNumber,
				Description:    data.Description,
			}
		},
		Created: func(data *GroceryComputationSheet) []string {
			return []string{
				"grocery_computation_sheet.create",
				fmt.Sprintf("grocery_computation_sheet.create.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GroceryComputationSheet) []string {
			return []string{
				"grocery_computation_sheet.update",
				fmt.Sprintf("grocery_computation_sheet.update.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GroceryComputationSheet) []string {
			return []string{
				"grocery_computation_sheet.delete",
				fmt.Sprintf("grocery_computation_sheet.delete.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.delete.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GroceryComputationSheetCurrentBranch returns GroceryComputationSheetCurrentBranch for the current branch or organization where applicable.
func (m *Core) GroceryComputationSheetCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GroceryComputationSheet, error) {
	return m.GroceryComputationSheetManager.Find(context, &GroceryComputationSheet{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
