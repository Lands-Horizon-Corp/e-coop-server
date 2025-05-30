package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	InterestRateByTerm struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_term"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_term"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name       string `gorm:"type:varchar(255)"`
		Descrition string `gorm:"type:varchar(4028)"` // Note: spelling matches schema

		MemberClassificationInterestRateID uuid.UUID                         `gorm:"type:uuid"`
		MemberClassificationInterestRate   *MemberClassificationInterestRate `gorm:"foreignKey:MemberClassificationInterestRateID;constraint:OnDelete:SET NULL;" json:"member_classification_interest_rate,omitempty"`
	}

	InterestRateByTermResponse struct {
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
		Descrition                         string                                    `json:"descrition"`
		MemberClassificationInterestRateID uuid.UUID                                 `json:"member_classification_interest_rate_id"`
		MemberClassificationInterestRate   *MemberClassificationInterestRateResponse `json:"member_classification_interest_rate,omitempty"`
	}

	InterestRateByTermRequest struct {
		Name                               string    `json:"name,omitempty"`
		Descrition                         string    `json:"descrition,omitempty"`
		MemberClassificationInterestRateID uuid.UUID `json:"member_classification_interest_rate_id,omitempty"`
	}
)

func (m *Model) InterestRateByTerm() {
	m.Migration = append(m.Migration, &InterestRateByTerm{})
	m.InterestRateByTermManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		InterestRateByTerm, InterestRateByTermResponse, InterestRateByTermRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "MemberClassificationInterestRate",
		},
		Service: m.provider.Service,
		Resource: func(data *InterestRateByTerm) *InterestRateByTermResponse {
			if data == nil {
				return nil
			}
			return &InterestRateByTermResponse{
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
				Descrition:                         data.Descrition,
				MemberClassificationInterestRateID: data.MemberClassificationInterestRateID,
				MemberClassificationInterestRate:   m.MemberClassificationInterestRateManager.ToModel(data.MemberClassificationInterestRate),
			}
		},
		Created: func(data *InterestRateByTerm) []string {
			return []string{
				"interest_rate_by_term.create",
				fmt.Sprintf("interest_rate_by_term.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRateByTerm) []string {
			return []string{
				"interest_rate_by_term.update",
				fmt.Sprintf("interest_rate_by_term.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRateByTerm) []string {
			return []string{
				"interest_rate_by_term.delete",
				fmt.Sprintf("interest_rate_by_term.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
