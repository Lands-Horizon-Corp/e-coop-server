package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	ChargesRateScheme struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ChargesRateByTermHeaderID uuid.UUID                `gorm:"type:uuid"`
		ChargesRateByTermHeader   *ChargesRateByTermHeader `gorm:"foreignKey:ChargesRateByTermHeaderID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"charges_rate_by_term_header,omitempty"`

		ChargesRateMemberTypeModeOfPaymentID uuid.UUID                           `gorm:"type:uuid"`
		ChargesRateMemberTypeModeOfPayment   *ChargesRateMemberTypeModeOfPayment `gorm:"foreignKey:ChargesRateMemberTypeModeOfPaymentID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"charges_rate_member_type_mode_of_payment,omitempty"`

		Name        string `gorm:"type:varchar(255);not null;unique"`
		Description string `gorm:"type:text;not null;unique"`
	}

	ChargesRateSchemeResponse struct {
		ID                                   uuid.UUID                                   `json:"id"`
		CreatedAt                            string                                      `json:"created_at"`
		CreatedByID                          uuid.UUID                                   `json:"created_by_id"`
		CreatedBy                            *UserResponse                               `json:"created_by,omitempty"`
		UpdatedAt                            string                                      `json:"updated_at"`
		UpdatedByID                          uuid.UUID                                   `json:"updated_by_id"`
		UpdatedBy                            *UserResponse                               `json:"updated_by,omitempty"`
		OrganizationID                       uuid.UUID                                   `json:"organization_id"`
		Organization                         *OrganizationResponse                       `json:"organization,omitempty"`
		BranchID                             uuid.UUID                                   `json:"branch_id"`
		Branch                               *BranchResponse                             `json:"branch,omitempty"`
		ChargesRateByTermHeaderID            uuid.UUID                                   `json:"charges_rate_by_term_header_id"`
		ChargesRateByTermHeader              *ChargesRateByTermHeaderResponse            `json:"charges_rate_by_term_header,omitempty"`
		ChargesRateMemberTypeModeOfPaymentID uuid.UUID                                   `json:"charges_rate_member_type_mode_of_payment_id"`
		ChargesRateMemberTypeModeOfPayment   *ChargesRateMemberTypeModeOfPaymentResponse `json:"charges_rate_member_type_mode_of_payment,omitempty"`
		Name                                 string                                      `json:"name"`
		Description                          string                                      `json:"description"`
	}

	ChargesRateSchemeRequest struct {
		ChargesRateByTermHeaderID            uuid.UUID `json:"charges_rate_by_term_header_id,omitempty"`
		ChargesRateMemberTypeModeOfPaymentID uuid.UUID `json:"charges_rate_member_type_mode_of_payment_id,omitempty"`
		Name                                 string    `json:"name" validate:"required,min=1,max=255"`
		Description                          string    `json:"description" validate:"required"`
	}
)

func (m *Model) ChargesRateScheme() {
	m.Migration = append(m.Migration, &ChargesRateScheme{})
	m.ChargesRateSchemeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ChargesRateScheme, ChargesRateSchemeResponse, ChargesRateSchemeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "ChargesRateByTermHeader", "ChargesRateMemberTypeModeOfPayment",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateScheme) *ChargesRateSchemeResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateSchemeResponse{
				ID:                                   data.ID,
				CreatedAt:                            data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                          data.CreatedByID,
				CreatedBy:                            m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                            data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                          data.UpdatedByID,
				UpdatedBy:                            m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                       data.OrganizationID,
				Organization:                         m.OrganizationManager.ToModel(data.Organization),
				BranchID:                             data.BranchID,
				Branch:                               m.BranchManager.ToModel(data.Branch),
				ChargesRateByTermHeaderID:            data.ChargesRateByTermHeaderID,
				ChargesRateByTermHeader:              m.ChargesRateByTermHeaderManager.ToModel(data.ChargesRateByTermHeader),
				ChargesRateMemberTypeModeOfPaymentID: data.ChargesRateMemberTypeModeOfPaymentID,
				ChargesRateMemberTypeModeOfPayment:   m.ChargesRateMemberTypeModeOfPaymentManager.ToModel(data.ChargesRateMemberTypeModeOfPayment),
				Name:                                 data.Name,
				Description:                          data.Description,
			}
		},
		Created: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.create",
				fmt.Sprintf("charges_rate_scheme.create.%s", data.ID),
			}
		},
		Updated: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.update",
				fmt.Sprintf("charges_rate_scheme.update.%s", data.ID),
			}
		},
		Deleted: func(data *ChargesRateScheme) []string {
			return []string{
				"charges_rate_scheme.delete",
				fmt.Sprintf("charges_rate_scheme.delete.%s", data.ID),
			}
		},
	})
}
