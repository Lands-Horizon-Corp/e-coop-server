package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	Bank struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_bank" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_bank" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	BankResponse struct {
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
		MediaID        *uuid.UUID            `json:"media_id,omitempty"`
		Media          *MediaResponse        `json:"media,omitempty"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	BankRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Description string     `json:"description,omitempty"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *ModelCore) bank() {
	m.Migration = append(m.Migration, &Bank{})
	m.BankManager = services.NewRepository(services.RepositoryParams[Bank, BankResponse, BankRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media"},
		Service:  m.provider.Service,
		Resource: func(data *Bank) *BankResponse {
			if data == nil {
				return nil
			}
			return &BankResponse{
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
				MediaID:        data.MediaID,
				Media:          m.MediaManager.ToModel(data.Media),
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *Bank) []string {
			return []string{
				"bank.create",
				fmt.Sprintf("bank.create.%s", data.ID),
				fmt.Sprintf("bank.create.branch.%s", data.BranchID),
				fmt.Sprintf("bank.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Bank) []string {
			return []string{
				"bank.update",
				fmt.Sprintf("bank.update.%s", data.ID),
				fmt.Sprintf("bank.update.branch.%s", data.BranchID),
				fmt.Sprintf("bank.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Bank) []string {
			return []string{
				"bank.delete",
				fmt.Sprintf("bank.delete.%s", data.ID),
				fmt.Sprintf("bank.delete.branch.%s", data.BranchID),
				fmt.Sprintf("bank.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) bankSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	banks := []*Bank{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "BDO Unibank, Inc.",
			Description:    "The largest bank in the Philippines by assets, BDO offers a wide range of financial services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bank of the Philippine Islands (BPI)",
			Description:    "One of the oldest banks in Southeast Asia, BPI provides banking and financial solutions.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Metropolitan Bank & Trust Company (Metrobank)",
			Description:    "A major universal bank in the Philippines, known for its extensive branch network.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Land Bank of the Philippines (Landbank)",
			Description:    "A government-owned bank focused on serving farmers and fishermen.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Philippine National Bank (PNB)",
			Description:    "One of the country’s largest banks, offering a full range of banking services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "China Banking Corporation (Chinabank)",
			Description:    "A leading private universal bank in the Philippines.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Security Bank Corporation",
			Description:    "A universal bank known for its innovative banking products.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Union Bank of the Philippines (UnionBank)",
			Description:    "A universal bank recognized for its digital banking leadership.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Development Bank of the Philippines (DBP)",
			Description:    "A government-owned development bank supporting infrastructure and social projects.",
		},
	}
	for _, data := range banks {
		if err := m.BankManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed bank %s", data.Name)
		}
	}
	return nil
}

// BankCurrentbranch retrieves all banks associated with the specified organization and branch.
func (m *ModelCore) BankCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Bank, error) {
	return m.BankManager.Find(context, &Bank{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
