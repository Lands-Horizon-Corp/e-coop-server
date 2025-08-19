package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	InterestRateByTermsHeader struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_terms_header"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_interest_rate_by_terms_header"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberClassificationInterestRateID uuid.UUID `gorm:"type:uuid"`
		// If you have a related model, add it here:
		// MemberClassificationInterestRate   *MemberClassificationInterestRate `gorm:"foreignKey:MemberClassificationInterestRateID;constraint:OnDelete:SET NULL;" json:"member_classification_interest_rate,omitempty"`

		Header1  int `gorm:"default:30"`
		Header2  int `gorm:"default:60"`
		Header3  int `gorm:"default:90"`
		Header4  int `gorm:"default:120"`
		Header5  int `gorm:"default:150"`
		Header6  int `gorm:"default:180"`
		Header7  int `gorm:"default:210"`
		Header8  int `gorm:"default:240"`
		Header9  int `gorm:"default:270"`
		Header10 int `gorm:"default:300"`
		Header11 int `gorm:"default:330"`
		Header12 int `gorm:"default:360"`
		Header13 int `gorm:"default:390"`
		Header14 int `gorm:"default:410"`
		Header15 int `gorm:"default:440"`
		Header16 int `gorm:"default:470"`
		Header17 int `gorm:"default:500"`
		Header18 int `gorm:"default:530"`
		Header19 int `gorm:"default:560"`
		Header20 int `gorm:"default:590"`
		Header21 int `gorm:"default:610"`
		Header22 int `gorm:"default:640"`
	}

	InterestRateByTermsHeaderResponse struct {
		ID                                 uuid.UUID             `json:"id"`
		CreatedAt                          string                `json:"created_at"`
		CreatedByID                        uuid.UUID             `json:"created_by_id"`
		CreatedBy                          *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt                          string                `json:"updated_at"`
		UpdatedByID                        uuid.UUID             `json:"updated_by_id"`
		UpdatedBy                          *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID                     uuid.UUID             `json:"organization_id"`
		Organization                       *OrganizationResponse `json:"organization,omitempty"`
		BranchID                           uuid.UUID             `json:"branch_id"`
		Branch                             *BranchResponse       `json:"branch,omitempty"`
		MemberClassificationInterestRateID uuid.UUID             `json:"member_classification_interest_rate_id"`
		// MemberClassificationInterestRate   *MemberClassificationInterestRateResponse `json:"member_classification_interest_rate,omitempty"`
		Header1  int `json:"header_1"`
		Header2  int `json:"header_2"`
		Header3  int `json:"header_3"`
		Header4  int `json:"header_4"`
		Header5  int `json:"header_5"`
		Header6  int `json:"header_6"`
		Header7  int `json:"header_7"`
		Header8  int `json:"header_8"`
		Header9  int `json:"header_9"`
		Header10 int `json:"header_10"`
		Header11 int `json:"header_11"`
		Header12 int `json:"header_12"`
		Header13 int `json:"header_13"`
		Header14 int `json:"header_14"`
		Header15 int `json:"header_15"`
		Header16 int `json:"header_16"`
		Header17 int `json:"header_17"`
		Header18 int `json:"header_18"`
		Header19 int `json:"header_19"`
		Header20 int `json:"header_20"`
		Header21 int `json:"header_21"`
		Header22 int `json:"header_22"`
	}

	InterestRateByTermsHeaderRequest struct {
		MemberClassificationInterestRateID uuid.UUID `json:"member_classification_interest_rate_id"`
		Header1                            int       `json:"header_1,omitempty"`
		Header2                            int       `json:"header_2,omitempty"`
		Header3                            int       `json:"header_3,omitempty"`
		Header4                            int       `json:"header_4,omitempty"`
		Header5                            int       `json:"header_5,omitempty"`
		Header6                            int       `json:"header_6,omitempty"`
		Header7                            int       `json:"header_7,omitempty"`
		Header8                            int       `json:"header_8,omitempty"`
		Header9                            int       `json:"header_9,omitempty"`
		Header10                           int       `json:"header_10,omitempty"`
		Header11                           int       `json:"header_11,omitempty"`
		Header12                           int       `json:"header_12,omitempty"`
		Header13                           int       `json:"header_13,omitempty"`
		Header14                           int       `json:"header_14,omitempty"`
		Header15                           int       `json:"header_15,omitempty"`
		Header16                           int       `json:"header_16,omitempty"`
		Header17                           int       `json:"header_17,omitempty"`
		Header18                           int       `json:"header_18,omitempty"`
		Header19                           int       `json:"header_19,omitempty"`
		Header20                           int       `json:"header_20,omitempty"`
		Header21                           int       `json:"header_21,omitempty"`
		Header22                           int       `json:"header_22,omitempty"`
	}
)

func (m *Model) InterestRateByTermsHeader() {
	m.Migration = append(m.Migration, &InterestRateByTermsHeader{})
	m.InterestRateByTermsHeaderManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		InterestRateByTermsHeader, InterestRateByTermsHeaderResponse, InterestRateByTermsHeaderRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			// "MemberClassificationInterestRate",
		},
		Service: m.provider.Service,
		Resource: func(data *InterestRateByTermsHeader) *InterestRateByTermsHeaderResponse {
			if data == nil {
				return nil
			}
			return &InterestRateByTermsHeaderResponse{
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
				MemberClassificationInterestRateID: data.MemberClassificationInterestRateID,
				// MemberClassificationInterestRate:   m.MemberClassificationInterestRateManager.ToModel(data.MemberClassificationInterestRate),
				Header1:  data.Header1,
				Header2:  data.Header2,
				Header3:  data.Header3,
				Header4:  data.Header4,
				Header5:  data.Header5,
				Header6:  data.Header6,
				Header7:  data.Header7,
				Header8:  data.Header8,
				Header9:  data.Header9,
				Header10: data.Header10,
				Header11: data.Header11,
				Header12: data.Header12,
				Header13: data.Header13,
				Header14: data.Header14,
				Header15: data.Header15,
				Header16: data.Header16,
				Header17: data.Header17,
				Header18: data.Header18,
				Header19: data.Header19,
				Header20: data.Header20,
				Header21: data.Header21,
				Header22: data.Header22,
			}
		},
		Created: func(data *InterestRateByTermsHeader) []string {
			return []string{
				"interest_rate_by_terms_header.create",
				fmt.Sprintf("interest_rate_by_terms_header.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_terms_header.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_terms_header.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InterestRateByTermsHeader) []string {
			return []string{
				"interest_rate_by_terms_header.update",
				fmt.Sprintf("interest_rate_by_terms_header.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_terms_header.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_terms_header.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InterestRateByTermsHeader) []string {
			return []string{
				"interest_rate_by_terms_header.delete",
				fmt.Sprintf("interest_rate_by_terms_header.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_terms_header.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_terms_header.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) InterestRateByTermsHeaderCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*InterestRateByTermsHeader, error) {
	return m.InterestRateByTermsHeaderManager.Find(context, &InterestRateByTermsHeader{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
