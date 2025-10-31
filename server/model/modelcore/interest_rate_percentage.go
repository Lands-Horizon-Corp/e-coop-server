package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	InterestRatePercentage struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_percentage"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_percentage"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name                               string                            `gorm:"type:varchar(255)"`
		Description                        string                            `gorm:"type:varchar(4028)"`
		Months                             int                               `gorm:"default:0"`
		InterestRate                       float64                           `gorm:"type:decimal;default:0"`
		MemberClassificationInterestRateID uuid.UUID                         `gorm:"type:uuid"`
		MemberClassificationInterestRate   *MemberClassificationInterestRate `gorm:"foreignKey:MemberClassificationInterestRateID;constraint:OnDelete:SET NULL;" json:"member_classification_interest_rate,omitempty"`
	}

	InterestRatePercentageResponse struct {
		ID                                 uuid.UUID                                 `json:"id"`
		CreatedAt                          string                                    `json:"created_at"`
		CreatedByID                        uuid.UUID                                 `json:"created_by_id"`
		CreatedBy                          *UserResponse                             `json:"created_by,omitempty"`
		UpdatedAt                          string                                    `json:"updated_at"`
		UpdatedByID                        uuid.UUID                                 `json:"updated_by_id"`
		UpdatedBy                          *UserResponse                             `json:"updated_by,omitempty"`
		OrganizationID                     uuid.UUID                                 `json:"organization_id"`
		Organization                       *OrganizationResponse                     `json:"organization,omitempty"`
		BranchID                           uuid.UUID                                 `json:"branch_id"`
		Branch                             *BranchResponse                           `json:"branch,omitempty"`
		Name                               string                                    `json:"name"`
		Description                        string                                    `json:"description"`
		Months                             int                                       `json:"months"`
		InterestRate                       float64                                   `json:"interest_rate"`
		MemberClassificationInterestRateID uuid.UUID                                 `json:"member_classification_interest_rate_id"`
		MemberClassificationInterestRate   *MemberClassificationInterestRateResponse `json:"member_classification_interest_rate,omitempty"`
	}

	InterestRatePercentageRequest struct {
		Name                               string    `json:"name,omitempty"`
		Description                        string    `json:"description,omitempty"`
		Months                             int       `json:"months,omitempty"`
		InterestRate                       float64   `json:"interest_rate,omitempty"`
		MemberClassificationInterestRateID uuid.UUID `json:"member_classification_interest_rate_id,omitempty"`
	}
)

func (m *ModelCore) interestRatePercentage() {
	m.Migration = append(m.Migration, &InterestRatePercentage{})
	m.InterestRatePercentageManager = services.NewRepository(services.RepositoryParams[
		InterestRatePercentage, InterestRatePercentageResponse, InterestRatePercentageRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberClassificationInterestRate",
		},
		Service: m.provider.Service,
		Resource: func(data *InterestRatePercentage) *InterestRatePercentageResponse {
			if data == nil {
				return nil
			}
			return &InterestRatePercentageResponse{
				ID:                                 data.ID,
				CreatedAt:                          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                        data.CreatedByID,
				CreatedBy:                          m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                        data.UpdatedByID,
				UpdatedBy:                          m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                     data.OrganizationID,
				Organization:                       m.OrganizationManager.ToModel(data.Organization),
				BranchID:                           data.BranchID,
				Branch:                             m.BranchManager.ToModel(data.Branch),
				Name:                               data.Name,
				Description:                        data.Description,
				Months:                             data.Months,
				InterestRate:                       data.InterestRate,
				MemberClassificationInterestRateID: data.MemberClassificationInterestRateID,
				MemberClassificationInterestRate:   m.MemberClassificationInterestRateManager.ToModel(data.MemberClassificationInterestRate),
			}
		},
		Created: func(data *InterestRatePercentage) []string {
			return []string{
				"interest_rate_percentage.create",
				fmt.Sprintf("interest_rate_percentage.create.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRatePercentage) []string {
			return []string{
				"interest_rate_percentage.update",
				fmt.Sprintf("interest_rate_percentage.update.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRatePercentage) []string {
			return []string{
				"interest_rate_percentage.delete",
				fmt.Sprintf("interest_rate_percentage.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) InterestRatePercentageCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*InterestRatePercentage, error) {
	return m.InterestRatePercentageManager.Find(context, &InterestRatePercentage{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
