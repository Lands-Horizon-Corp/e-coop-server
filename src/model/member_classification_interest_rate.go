package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberClassificationInterestRate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification_interest_rate"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification_interest_rate"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`

		InterestRateSchemeID *uuid.UUID          `gorm:"type:uuid"`
		InterestRateScheme   *InterestRateScheme `gorm:"foreignKey:InterestRateSchemeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"interest_rate_scheme,omitempty"`

		MemberClassificationID *uuid.UUID            `gorm:"type:uuid"`
		MemberClassification   *MemberClassification `gorm:"foreignKey:MemberClassificationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_classification,omitempty"`

		InterestRateByTermsHeaderID *uuid.UUID                 `gorm:"type:uuid"`
		InterestRateByTermsHeader   *InterestRateByTermsHeader `gorm:"foreignKey:InterestRateByTermsHeaderID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"interest_rate_by_terms_header,omitempty"`
	}

	MemberClassificationInterestRateResponse struct {
		ID                          uuid.UUID                          `json:"id"`
		CreatedAt                   string                             `json:"created_at"`
		CreatedByID                 uuid.UUID                          `json:"created_by_id"`
		CreatedBy                   *UserResponse                      `json:"created_by,omitempty"`
		UpdatedAt                   string                             `json:"updated_at"`
		UpdatedByID                 uuid.UUID                          `json:"updated_by_id"`
		UpdatedBy                   *UserResponse                      `json:"updated_by,omitempty"`
		OrganizationID              uuid.UUID                          `json:"organization_id"`
		Organization                *OrganizationResponse              `json:"organization,omitempty"`
		BranchID                    uuid.UUID                          `json:"branch_id"`
		Branch                      *BranchResponse                    `json:"branch,omitempty"`
		Name                        string                             `json:"name"`
		Description                 string                             `json:"description"`
		InterestRateSchemeID        *uuid.UUID                         `json:"interest_rate_scheme_id,omitempty"`
		InterestRateScheme          *InterestRateSchemeResponse        `json:"interest_rate_scheme,omitempty"`
		MemberClassificationID      *uuid.UUID                         `json:"member_classification_id,omitempty"`
		MemberClassification        *MemberClassificationResponse      `json:"member_classification,omitempty"`
		InterestRateByTermsHeaderID *uuid.UUID                         `json:"interest_rate_by_terms_header_id,omitempty"`
		InterestRateByTermsHeader   *InterestRateByTermsHeaderResponse `json:"interest_rate_by_terms_header,omitempty"`
	}

	MemberClassificationInterestRateRequest struct {
		Name                        string     `json:"name" validate:"required,min=1,max=255"`
		Description                 string     `json:"description,omitempty"`
		InterestRateSchemeID        *uuid.UUID `json:"interest_rate_scheme_id,omitempty"`
		MemberClassificationID      *uuid.UUID `json:"member_classification_id,omitempty"`
		InterestRateByTermsHeaderID *uuid.UUID `json:"interest_rate_by_terms_header_id,omitempty"`
	}
)

func (m *Model) MemberClassificationInterestRate() {
	m.Migration = append(m.Migration, &MemberClassificationInterestRate{})
	m.MemberClassificationInterestRateManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberClassificationInterestRate, MemberClassificationInterestRateResponse, MemberClassificationInterestRateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"InterestRateScheme", "MemberClassification", "InterestRateByTermsHeader",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberClassificationInterestRate) *MemberClassificationInterestRateResponse {
			if data == nil {
				return nil
			}
			return &MemberClassificationInterestRateResponse{
				ID:                          data.ID,
				CreatedAt:                   data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                 data.CreatedByID,
				CreatedBy:                   m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                 data.UpdatedByID,
				UpdatedBy:                   m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:              data.OrganizationID,
				Organization:                m.OrganizationManager.ToModel(data.Organization),
				BranchID:                    data.BranchID,
				Branch:                      m.BranchManager.ToModel(data.Branch),
				Name:                        data.Name,
				Description:                 data.Description,
				InterestRateSchemeID:        data.InterestRateSchemeID,
				InterestRateScheme:          m.InterestRateSchemeManager.ToModel(data.InterestRateScheme),
				MemberClassificationID:      data.MemberClassificationID,
				MemberClassification:        m.MemberClassificationManager.ToModel(data.MemberClassification),
				InterestRateByTermsHeaderID: data.InterestRateByTermsHeaderID,
				InterestRateByTermsHeader:   m.InterestRateByTermsHeaderManager.ToModel(data.InterestRateByTermsHeader),
			}
		},
		Created: func(data *MemberClassificationInterestRate) []string {
			return []string{
				"member_classification_interest_rate.create",
				fmt.Sprintf("member_classification_interest_rate.create.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberClassificationInterestRate) []string {
			return []string{
				"member_classification_interest_rate.update",
				fmt.Sprintf("member_classification_interest_rate.update.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberClassificationInterestRate) []string {
			return []string{
				"member_classification_interest_rate.delete",
				fmt.Sprintf("member_classification_interest_rate.delete.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
