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
	ChargesRateByTermHeader struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_term_header"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_term_header"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Header1  int `gorm:"default:0"`
		Header2  int `gorm:"default:0"`
		Header3  int `gorm:"default:0"`
		Header4  int `gorm:"default:0"`
		Header5  int `gorm:"default:0"`
		Header6  int `gorm:"default:0"`
		Header7  int `gorm:"default:0"`
		Header8  int `gorm:"default:0"`
		Header9  int `gorm:"default:0"`
		Header10 int `gorm:"default:0"`
		Header11 int `gorm:"default:0"`
		Header12 int `gorm:"default:0"`
		Header13 int `gorm:"default:0"`
		Header14 int `gorm:"default:0"`
		Header15 int `gorm:"default:0"`
		Header16 int `gorm:"default:0"`
		Header17 int `gorm:"default:0"`
		Header18 int `gorm:"default:0"`
		Header19 int `gorm:"default:0"`
		Header20 int `gorm:"default:0"`
		Header21 int `gorm:"default:0"`
		Header22 int `gorm:"default:0"`
	}

	ChargesRateByTermHeaderResponse struct {
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
		Header1        int                   `json:"header_1"`
		Header2        int                   `json:"header_2"`
		Header3        int                   `json:"header_3"`
		Header4        int                   `json:"header_4"`
		Header5        int                   `json:"header_5"`
		Header6        int                   `json:"header_6"`
		Header7        int                   `json:"header_7"`
		Header8        int                   `json:"header_8"`
		Header9        int                   `json:"header_9"`
		Header10       int                   `json:"header_10"`
		Header11       int                   `json:"header_11"`
		Header12       int                   `json:"header_12"`
		Header13       int                   `json:"header_13"`
		Header14       int                   `json:"header_14"`
		Header15       int                   `json:"header_15"`
		Header16       int                   `json:"header_16"`
		Header17       int                   `json:"header_17"`
		Header18       int                   `json:"header_18"`
		Header19       int                   `json:"header_19"`
		Header20       int                   `json:"header_20"`
		Header21       int                   `json:"header_21"`
		Header22       int                   `json:"header_22"`
	}

	ChargesRateByTermHeaderRequest struct {
		Header1  int `json:"header_1,omitempty"`
		Header2  int `json:"header_2,omitempty"`
		Header3  int `json:"header_3,omitempty"`
		Header4  int `json:"header_4,omitempty"`
		Header5  int `json:"header_5,omitempty"`
		Header6  int `json:"header_6,omitempty"`
		Header7  int `json:"header_7,omitempty"`
		Header8  int `json:"header_8,omitempty"`
		Header9  int `json:"header_9,omitempty"`
		Header10 int `json:"header_10,omitempty"`
		Header11 int `json:"header_11,omitempty"`
		Header12 int `json:"header_12,omitempty"`
		Header13 int `json:"header_13,omitempty"`
		Header14 int `json:"header_14,omitempty"`
		Header15 int `json:"header_15,omitempty"`
		Header16 int `json:"header_16,omitempty"`
		Header17 int `json:"header_17,omitempty"`
		Header18 int `json:"header_18,omitempty"`
		Header19 int `json:"header_19,omitempty"`
		Header20 int `json:"header_20,omitempty"`
		Header21 int `json:"header_21,omitempty"`
		Header22 int `json:"header_22,omitempty"`
	}
)

func (m *Model) ChargesRateByTermHeader() {
	m.Migration = append(m.Migration, &ChargesRateByTermHeader{})
	m.ChargesRateByTermHeaderManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ChargesRateByTermHeader, ChargesRateByTermHeaderResponse, ChargesRateByTermHeaderRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateByTermHeader) *ChargesRateByTermHeaderResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateByTermHeaderResponse{
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
				Header1:        data.Header1,
				Header2:        data.Header2,
				Header3:        data.Header3,
				Header4:        data.Header4,
				Header5:        data.Header5,
				Header6:        data.Header6,
				Header7:        data.Header7,
				Header8:        data.Header8,
				Header9:        data.Header9,
				Header10:       data.Header10,
				Header11:       data.Header11,
				Header12:       data.Header12,
				Header13:       data.Header13,
				Header14:       data.Header14,
				Header15:       data.Header15,
				Header16:       data.Header16,
				Header17:       data.Header17,
				Header18:       data.Header18,
				Header19:       data.Header19,
				Header20:       data.Header20,
				Header21:       data.Header21,
				Header22:       data.Header22,
			}
		},
		Created: func(data *ChargesRateByTermHeader) []string {
			return []string{
				"charges_rate_by_term_header.create",
				fmt.Sprintf("charges_rate_by_term_header.create.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term_header.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term_header.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateByTermHeader) []string {
			return []string{
				"charges_rate_by_term_header.update",
				fmt.Sprintf("charges_rate_by_term_header.update.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term_header.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term_header.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateByTermHeader) []string {
			return []string{
				"charges_rate_by_term_header.delete",
				fmt.Sprintf("charges_rate_by_term_header.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term_header.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term_header.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) ChargesRateByTermHeaderCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*ChargesRateByTermHeader, error) {
	return m.ChargesRateByTermHeaderManager.Find(context, &ChargesRateByTermHeader{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
