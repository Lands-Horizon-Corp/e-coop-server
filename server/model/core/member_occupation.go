package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// MemberOccupation represents a member's occupation information in the database
	MemberOccupation struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	// MemberOccupationResponse represents the response structure for member occupation data
	MemberOccupationResponse struct {
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
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	// MemberOccupationRequest represents the request structure for member occupation data
	MemberOccupationRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Core) memberOccupation() {
	m.Migration = append(m.Migration, &MemberOccupation{})
	m.MemberOccupationManager = *registry.NewRegistry(registry.RegistryParams[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberOccupation) *MemberOccupationResponse {
			if data == nil {
				return nil
			}
			return &MemberOccupationResponse{
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
				Name:           data.Name,
				Description:    data.Description,
			}
		},

		Created: func(data *MemberOccupation) registry.Topics {
			return []string{
				"member_occupation.create",
				fmt.Sprintf("member_occupation.create.%s", data.ID),
				fmt.Sprintf("member_occupation.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberOccupation) registry.Topics {
			return []string{
				"member_occupation.update",
				fmt.Sprintf("member_occupation.update.%s", data.ID),
				fmt.Sprintf("member_occupation.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberOccupation) registry.Topics {
			return []string{
				"member_occupation.delete",
				fmt.Sprintf("member_occupation.delete.%s", data.ID),
				fmt.Sprintf("member_occupation.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) memberOccupationSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {

	now := time.Now().UTC()
	memberOccupations := []*MemberOccupation{
		{Name: "Farmer", Description: "Engaged in agriculture or crop cultivation.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Fisherfolk", Description: "Involved in fishing and aquaculture activities.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Agricultural Technician", Description: "Specializes in modern agricultural practices and tools.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Software Developer", Description: "Develops and maintains software systems.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "IT Specialist", Description: "Manages information technology infrastructure.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Accountant", Description: "Handles financial records and audits.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Teacher", Description: "Educates students in academic institutions.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Nurse", Description: "Provides healthcare and medical support.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Doctor", Description: "Licensed medical professional for diagnosing and treating patients.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Engineer", Description: "Designs and builds infrastructure or systems.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Construction Worker", Description: "Works on building and construction projects.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Driver", Description: "Professional vehicle operator (e.g., jeepney, tricycle, delivery).", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Vendor", Description: "Operates a small retail business or market stall.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Self-Employed", Description: "Independent worker managing their own business or services.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Housewife", Description: "Manages household responsibilities full-time.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Househusband", Description: "Male homemaker managing family and household duties.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Artist", Description: "Engaged in creative fields like painting, sculpture, or multimedia arts.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Graphic Designer", Description: "Creates visual content using design software.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Call Center Agent", Description: "Provides customer service through phone or chat support.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Unemployed", Description: "Currently without formal occupation.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Physicist", Description: "Studies the properties and interactions of matter and energy.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Pharmacist", Description: "Dispenses medications and advises on their safe use.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Chef", Description: "Creates recipes and prepares meals in restaurants or catering.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Mechanic", Description: "Repairs and maintains vehicles and machinery.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Electrician", Description: "Installs and repairs electrical systems and wiring.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Plumber", Description: "Installs and repairs piping systems for water and waste.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Architect", Description: "Designs buildings and ensures structural soundness.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Banker", Description: "Manages financial transactions and client relationships.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Lawyer", Description: "Provides legal advice and represents clients in court.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Journalist", Description: "Researches and reports news for print, online, or broadcast.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Social Worker", Description: "Supports individuals and families through counseling and services.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Caregiver", Description: "Provides in-home care and assistance to the elderly or disabled.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Security Guard", Description: "Protects property and enforces safety protocols.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Teacher’s Aide", Description: "Assists teachers in classroom management and lesson prep.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Student", Description: "Currently enrolled in an educational institution.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Retiree", Description: "Previously employed, now retired from active work.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Entrepreneur", Description: "Owns and operates one or more business ventures.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Musician", Description: "Performs, composes, or teaches music.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Writer", Description: "Crafts written content—books, articles, or scripts.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Pilot", Description: "Operates aircraft for commercial or private flights.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Scientist", Description: "Conducts research in natural or social sciences.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Lab Technician", Description: "Performs tests and experiments in scientific labs.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Receptionist", Description: "Manages front-desk operations and customer inquiries.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
		{Name: "Janitor", Description: "Keeps buildings clean and well-maintained.", CreatedAt: now, CreatedByID: userID, UpdatedAt: now, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID},
	}
	for _, data := range memberOccupations {
		if err := m.MemberOccupationManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member ooccupation %s", data.Name)
		}
	}
	return nil
}

// MemberOccupationCurrentBranch retrieves member occupations for the current branch
func (m *Core) MemberOccupationCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOccupation, error) {
	return m.MemberOccupationManager.Find(context, &MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
