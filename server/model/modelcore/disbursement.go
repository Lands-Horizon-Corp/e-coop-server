package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	Disbursement struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Icon        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
	}

	DisbursementResponse struct {
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
		CurrencyID     uuid.UUID             `json:"currency_id"`
		Currency       *CurrencyResponse     `json:"currency,omitempty"`
		Name           string                `json:"name"`
		Icon           string                `json:"icon"`
		Description    string                `json:"description"`
	}

	DisbursementRequest struct {
		Name        string    `json:"name" validate:"required,min=1,max=50"`
		Icon        string    `json:"icon,omitempty"`
		Description string    `json:"description,omitempty"`
		CurrencyID  uuid.UUID `json:"currency_id" validate:"required"`
	}
)

func (m *ModelCore) disbursement() {
	m.Migration = append(m.Migration, &Disbursement{})
	m.DisbursementManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Disbursement, DisbursementResponse, DisbursementRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
			"Organization.Media", "Branch.Media",
		},
		Service: m.provider.Service,
		Resource: func(data *Disbursement) *DisbursementResponse {
			if data == nil {
				return nil
			}
			return &DisbursementResponse{
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
				CurrencyID:     data.CurrencyID,
				Currency:       m.CurrencyManager.ToModel(data.Currency),
				Name:           data.Name,
				Icon:           data.Icon,
				Description:    data.Description,
			}
		},
		Created: func(data *Disbursement) []string {
			return []string{
				"disbursement.create",
				fmt.Sprintf("disbursement.create.%s", data.ID),
				fmt.Sprintf("disbursement.create.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Disbursement) []string {
			return []string{
				"disbursement.update",
				fmt.Sprintf("disbursement.update.%s", data.ID),
				fmt.Sprintf("disbursement.update.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Disbursement) []string {
			return []string{
				"disbursement.delete",
				fmt.Sprintf("disbursement.delete.%s", data.ID),
				fmt.Sprintf("disbursement.delete.branch.%s", data.BranchID),
				fmt.Sprintf("disbursement.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) DisbursementCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Disbursement, error) {
	return m.DisbursementManager.Find(context, &Disbursement{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *ModelCore) DisbursementSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	branch, err := m.BranchManager.GetByID(context, branchID)
	if err != nil {
		return eris.Wrap(err, "failed to find branch for account seeding")
	}
	currency, err := m.CurrencyFindByAlpha2(context, branch.CountryCode)
	if err != nil {
		return eris.Wrap(err, "failed to find currency for account seeding")
	}

	disbursements := []*Disbursement{
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Petty Cash",

			Description: "Small cash disbursements for minor expenses and miscellaneous operational costs.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Office Supplies",

			Description: "Purchase of office materials, stationery, and administrative supplies.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Utilities",
			Description:    "Payment for electricity, water, internet, and other utility services.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Travel Expenses",
			Description:    "Transportation costs, accommodation, and meal allowances for official travels.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Meeting Expenses",
			Description:    "Costs associated with meetings, seminars, and training sessions.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Equipment Purchase",
			Description:    "Acquisition of office equipment, furniture, and technology devices.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Maintenance & Repairs",
			Description:    "Building maintenance, equipment repairs, and facility improvements.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Insurance Premium",
			Description:    "Payment for insurance coverage including property, liability, and employee insurance.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Professional Services",
			Description:    "Fees for legal, accounting, consulting, and other professional services.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Member Benefits",
			Description:    "Disbursements for member welfare, dividends, and cooperative benefits distribution.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Loan Disbursement",
			Description:    "Release of approved loans to cooperative members.",
		},
		{
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			CurrencyID:     currency.ID,
			Name:           "Emergency Fund",
			Description:    "Disbursements from emergency reserves for urgent organizational needs.",
		},
	}

	for _, data := range disbursements {
		if err := m.DisbursementManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed disbursement %s", data.Name)
		}
	}

	return nil
}
