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
	LoanGuaranteedFundPerMonth struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Month              int `gorm:"type:int;default:0"`
		LoanGuaranteedFund int `gorm:"type:int;default:0"`
	}

	LoanGuaranteedFundPerMonthResponse struct {
		ID                 uuid.UUID             `json:"id"`
		CreatedAt          string                `json:"created_at"`
		CreatedByID        uuid.UUID             `json:"created_by_id"`
		CreatedBy          *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt          string                `json:"updated_at"`
		UpdatedByID        uuid.UUID             `json:"updated_by_id"`
		UpdatedBy          *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID             `json:"organization_id"`
		Organization       *OrganizationResponse `json:"organization,omitempty"`
		BranchID           uuid.UUID             `json:"branch_id"`
		Branch             *BranchResponse       `json:"branch,omitempty"`
		Month              int                   `json:"month"`
		LoanGuaranteedFund int                   `json:"loan_guaranteed_fund"`
	}

	LoanGuaranteedFundPerMonthRequest struct {
		Month              int `json:"month,omitempty"`
		LoanGuaranteedFund int `json:"loan_guaranteed_fund,omitempty"`
	}
)

func LoanGuaranteedFundPerMonthManager(service *horizon.HorizonService) *registry.Registry[LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		LoanGuaranteedFundPerMonth, LoanGuaranteedFundPerMonthResponse, LoanGuaranteedFundPerMonthRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *LoanGuaranteedFundPerMonth) *LoanGuaranteedFundPerMonthResponse {
			if data == nil {
				return nil
			}
			return &LoanGuaranteedFundPerMonthResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				Month:              data.Month,
				LoanGuaranteedFund: data.LoanGuaranteedFund,
			}
		},

		Created: func(data *LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.create",
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.update",
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanGuaranteedFundPerMonth) registry.Topics {
			return []string{
				"loan_guaranteed_fund_per_month.delete",
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.%s", data.ID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_guaranteed_fund_per_month.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanGuaranteedFundPerMonthCurrentBranch(ctx context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanGuaranteedFundPerMonth, error) {
	return LoanGuaranteedFundPerMonthManager(service).Find(ctx, &LoanGuaranteedFundPerMonth{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
