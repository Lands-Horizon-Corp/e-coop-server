package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	MemberOccupation struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`

		MemberProfiles []*MemberProfile `gorm:"foreignKey:MemberOccupationID;references:ID" json:"member_profiles,omitempty"`
	}

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

		Name           string                   `json:"name"`
		Description    string                   `json:"description"`
		MemberProfiles []*MemberProfileResponse `json:"member_profiles"`
	}

	MemberOccupationRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberOccupationCollection struct {
		Manager horizon_manager.CollectionManager[MemberOccupation]
	}
)

func (m *Model) MemberOccupationValidate(ctx echo.Context) (*MemberOccupationRequest, error) {
	return horizon_manager.Validate[MemberOccupationRequest](ctx, m.validator)
}

func (m *Model) MemberOccupationModel(data *MemberOccupation) *MemberOccupationResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberOccupation) *MemberOccupationResponse {
		return &MemberOccupationResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.UpdatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),
			Name:           data.Name,
			Description:    data.Description,
			MemberProfiles: m.MemberProfileModels(data.MemberProfiles),
		}
	})
}

func NewMemberOccupationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberOccupationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				fmt.Sprintf("member_occupation.create.%s", data.ID),
				fmt.Sprintf("member_occupation.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.create.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				"member_occupation.update",
				fmt.Sprintf("member_occupation.update.%s", data.ID),
				fmt.Sprintf("member_occupation.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.update.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				"member_occupation.delete",
				fmt.Sprintf("member_occupation.delete.%s", data.ID),
				fmt.Sprintf("member_occupation.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.delete.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberOccupationCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberOccupationModels(data []*MemberOccupation) []*MemberOccupationResponse {
	return horizon_manager.ToModels(data, m.MemberOccupationModel)
}

// member-occupation/branch/:branch_id
func (fc *MemberOccupationCollection) ListByBranch(branchID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		BranchID: branchID,
	})
}

// member-occupation/organization/:organization_id
func (fc *MemberOccupationCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		OrganizationID: organizationID,
	})
}

// member-occupation/organization/:organization_id/branch/:branch_id
func (fc *MemberOccupationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (fc *MemberOccupationCollection) Seeder(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOccupation, error) {
	now := time.Now()

	occupations := []*MemberOccupation{
		{
			ID:          uuid.New(),
			Name:        "Farmer",
			Description: "Engaged in agriculture or crop cultivation.",
		},
		{
			ID:          uuid.New(),
			Name:        "Fisherfolk",
			Description: "Involved in fishing and aquaculture activities.",
		},
		{
			ID:          uuid.New(),
			Name:        "Agricultural Technician",
			Description: "Specializes in modern agricultural practices and tools.",
		},
		{
			ID:          uuid.New(),
			Name:        "Software Developer",
			Description: "Develops and maintains software systems.",
		},
		{
			ID:          uuid.New(),
			Name:        "IT Specialist",
			Description: "Manages information technology infrastructure.",
		},
		{
			ID:          uuid.New(),
			Name:        "Accountant",
			Description: "Handles financial records and audits.",
		},
		{
			ID:          uuid.New(),
			Name:        "Teacher",
			Description: "Educates students in academic institutions.",
		},
		{
			ID:          uuid.New(),
			Name:        "Nurse",
			Description: "Provides healthcare and medical support.",
		},
		{
			ID:          uuid.New(),
			Name:        "Doctor",
			Description: "Licensed medical professional for diagnosing and treating patients.",
		},
		{
			ID:          uuid.New(),
			Name:        "Engineer",
			Description: "Designs and builds infrastructure or systems.",
		},
		{
			ID:          uuid.New(),
			Name:        "Construction Worker",
			Description: "Works on building and construction projects.",
		},
		{
			ID:          uuid.New(),
			Name:        "Driver",
			Description: "Professional vehicle operator (e.g., jeepney, tricycle, delivery).",
		},
		{
			ID:          uuid.New(),
			Name:        "Vendor",
			Description: "Operates a small retail business or market stall.",
		},
		{
			ID:          uuid.New(),
			Name:        "Self-Employed",
			Description: "Independent worker managing their own business or services.",
		},
		{
			ID:          uuid.New(),
			Name:        "Housewife",
			Description: "Manages household responsibilities full-time.",
		},
		{
			ID:          uuid.New(),
			Name:        "Househusband",
			Description: "Male homemaker managing family and household duties.",
		},
		{
			ID:          uuid.New(),
			Name:        "Artist",
			Description: "Engaged in creative fields like painting, sculpture, or multimedia arts.",
		},
		{
			ID:          uuid.New(),
			Name:        "Graphic Designer",
			Description: "Creates visual content using design software.",
		},
		{
			ID:          uuid.New(),
			Name:        "Call Center Agent",
			Description: "Provides customer service through phone or chat support.",
		},
		{
			ID:          uuid.New(),
			Name:        "Unemployed",
			Description: "Currently without formal occupation.",
		},
		{
			ID:          uuid.New(),
			Name:        "Physicist",
			Description: "Studies the properties and interactions of matter and energy.",
		},
		{
			ID:          uuid.New(),
			Name:        "Pharmacist",
			Description: "Dispenses medications and advises on their safe use.",
		},
		{
			ID:          uuid.New(),
			Name:        "Chef",
			Description: "Creates recipes and prepares meals in restaurants or catering.",
		},
		{
			ID:          uuid.New(),
			Name:        "Mechanic",
			Description: "Repairs and maintains vehicles and machinery.",
		},
		{
			ID:          uuid.New(),
			Name:        "Electrician",
			Description: "Installs and repairs electrical systems and wiring.",
		},
		{
			ID:          uuid.New(),
			Name:        "Plumber",
			Description: "Installs and repairs piping systems for water and waste.",
		},
		{
			ID:          uuid.New(),
			Name:        "Architect",
			Description: "Designs buildings and ensures structural soundness.",
		},
		{
			ID:          uuid.New(),
			Name:        "Banker",
			Description: "Manages financial transactions and client relationships.",
		},
		{
			ID:          uuid.New(),
			Name:        "Lawyer",
			Description: "Provides legal advice and represents clients in court.",
		},
		{
			ID:          uuid.New(),
			Name:        "Journalist",
			Description: "Researches and reports news for print, online, or broadcast.",
		},
		{
			ID:          uuid.New(),
			Name:        "Social Worker",
			Description: "Supports individuals and families through counseling and services.",
		},
		{
			ID:          uuid.New(),
			Name:        "Caregiver",
			Description: "Provides in-home care and assistance to the elderly or disabled.",
		},
		{
			ID:          uuid.New(),
			Name:        "Security Guard",
			Description: "Protects property and enforces safety protocols.",
		},
		{
			ID:          uuid.New(),
			Name:        "Teacher’s Aide",
			Description: "Assists teachers in classroom management and lesson prep.",
		},
		{
			ID:          uuid.New(),
			Name:        "Student",
			Description: "Currently enrolled in an educational institution.",
		},
		{
			ID:          uuid.New(),
			Name:        "Retiree",
			Description: "Previously employed, now retired from active work.",
		},
		{
			ID:          uuid.New(),
			Name:        "Entrepreneur",
			Description: "Owns and operates one or more business ventures.",
		},
		{
			ID:          uuid.New(),
			Name:        "Musician",
			Description: "Performs, composes, or teaches music.",
		},
		{
			ID:          uuid.New(),
			Name:        "Writer",
			Description: "Crafts written content—books, articles, or scripts.",
		},
		{
			ID:          uuid.New(),
			Name:        "Pilot",
			Description: "Operates aircraft for commercial or private flights.",
		},
		{
			ID:          uuid.New(),
			Name:        "Scientist",
			Description: "Conducts research in natural or social sciences.",
		},
		{
			ID:          uuid.New(),
			Name:        "Lab Technician",
			Description: "Performs tests and experiments in scientific labs.",
		},
		{
			ID:          uuid.New(),
			Name:        "Receptionist",
			Description: "Manages front-desk operations and customer inquiries.",
		},
		{
			ID:          uuid.New(),
			Name:        "Janitor",
			Description: "Keeps buildings clean and well-maintained.",
		},
	}

	// Set timestamps and foreign keys
	for _, o := range occupations {
		o.CreatedAt = now
		o.CreatedByID = userID
		o.UpdatedAt = now
		o.UpdatedByID = userID
		o.OrganizationID = organizationID
		o.BranchID = branchID
	}

	if err := fc.Manager.CreateMany(occupations); err != nil {
		return nil, err
	}

	return occupations, nil
}
