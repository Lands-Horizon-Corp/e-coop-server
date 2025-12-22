package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
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

	MemberClassificationInterestRateResponse struct {
		ID                     uuid.UUID                     `json:"id"`
		CreatedAt              string                        `json:"created_at"`
		CreatedByID            uuid.UUID                     `json:"created_by_id"`
		CreatedBy              *UserResponse                 `json:"created_by,omitempty"`
		UpdatedAt              string                        `json:"updated_at"`
		UpdatedByID            uuid.UUID                     `json:"updated_by_id"`
		UpdatedBy              *UserResponse                 `json:"updated_by,omitempty"`
		OrganizationID         uuid.UUID                     `json:"organization_id"`
		Organization           *OrganizationResponse         `json:"organization,omitempty"`
		BranchID               uuid.UUID                     `json:"branch_id"`
		Branch                 *BranchResponse               `json:"branch,omitempty"`
		Name                   string                        `json:"name"`
		Description            string                        `json:"description"`
		InterestRateSchemeID   *uuid.UUID                    `json:"interest_rate_scheme_id,omitempty"`
		InterestRateScheme     *InterestRateSchemeResponse   `json:"interest_rate_scheme,omitempty"`
		MemberClassificationID *uuid.UUID                    `json:"member_classification_id,omitempty"`
		MemberClassification   *MemberClassificationResponse `json:"member_classification,omitempty"`

		Header1  int `json:"header1"`
		Header2  int `json:"header2"`
		Header3  int `json:"header3"`
		Header4  int `json:"header4"`
		Header5  int `json:"header5"`
		Header6  int `json:"header6"`
		Header7  int `json:"header7"`
		Header8  int `json:"header8"`
		Header9  int `json:"header9"`
		Header10 int `json:"header10"`
		Header11 int `json:"header11"`
		Header12 int `json:"header12"`
		Header13 int `json:"header13"`
		Header14 int `json:"header14"`
		Header15 int `json:"header15"`
		Header16 int `json:"header16"`
		Header17 int `json:"header17"`
		Header18 int `json:"header18"`
		Header19 int `json:"header19"`
		Header20 int `json:"header20"`
		Header21 int `json:"header21"`
		Header22 int `json:"header22"`
	}

	MemberClassificationInterestRateRequest struct {
		Name                        string     `json:"name" validate:"required,min=1,max=255"`
		Description                 string     `json:"description,omitempty"`
		InterestRateSchemeID        *uuid.UUID `json:"interest_rate_scheme_id,omitempty"`
		MemberClassificationID      *uuid.UUID `json:"member_classification_id,omitempty"`
		InterestRateByTermsHeaderID *uuid.UUID `json:"interest_rate_by_terms_header_id,omitempty"`

		Header1  int `json:"header1,omitempty"`
		Header2  int `json:"header2,omitempty"`
		Header3  int `json:"header3,omitempty"`
		Header4  int `json:"header4,omitempty"`
		Header5  int `json:"header5,omitempty"`
		Header6  int `json:"header6,omitempty"`
		Header7  int `json:"header7,omitempty"`
		Header8  int `json:"header8,omitempty"`
		Header9  int `json:"header9,omitempty"`
		Header10 int `json:"header10,omitempty"`
		Header11 int `json:"header11,omitempty"`
		Header12 int `json:"header12,omitempty"`
		Header13 int `json:"header13,omitempty"`
		Header14 int `json:"header14,omitempty"`
		Header15 int `json:"header15,omitempty"`
		Header16 int `json:"header16,omitempty"`
		Header17 int `json:"header17,omitempty"`
		Header18 int `json:"header18,omitempty"`
		Header19 int `json:"header19,omitempty"`
		Header20 int `json:"header20,omitempty"`
		Header21 int `json:"header21,omitempty"`
		Header22 int `json:"header22,omitempty"`
	}
)

func (m *Core) memberClassificationInterestRate() {
	m.Migration = append(m.Migration, &MemberClassificationInterestRate{})
	m.MemberClassificationInterestRateManager().= registry.NewRegistry(registry.RegistryParams[
		MemberClassificationInterestRate, MemberClassificationInterestRateResponse, MemberClassificationInterestRateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"InterestRateScheme", "MemberClassification", "InterestRateByTermsHeader",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberClassificationInterestRate) *MemberClassificationInterestRateResponse {
			if data == nil {
				return nil
			}
			return &MemberClassificationInterestRateResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           m.OrganizationManager().ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 m.BranchManager().ToModel(data.Branch),
				Name:                   data.Name,
				Description:            data.Description,
				InterestRateSchemeID:   data.InterestRateSchemeID,
				InterestRateScheme:     m.InterestRateSchemeManager().ToModel(data.InterestRateScheme),
				MemberClassificationID: data.MemberClassificationID,
				MemberClassification:   m.MemberClassificationManager().ToModel(data.MemberClassification),

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
		Created: func(data *MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.create",
				fmt.Sprintf("member_classification_interest_rate.create.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.update",
				fmt.Sprintf("member_classification_interest_rate.update.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberClassificationInterestRate) registry.Topics {
			return []string{
				"member_classification_interest_rate.delete",
				fmt.Sprintf("member_classification_interest_rate.delete.%s", data.ID),
				fmt.Sprintf("member_classification_interest_rate.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_interest_rate.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) MemberClassificationInterestRateCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberClassificationInterestRate, error) {
	return m.MemberClassificationInterestRateManager().Find(context, &MemberClassificationInterestRate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
