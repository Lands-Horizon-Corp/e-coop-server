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
	// Company represents the Company model.
	Company struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_company" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_company" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	// CompanyResponse represents the response structure for company data

	// CompanyResponse represents the response structure for Company.
	CompanyResponse struct {
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

	// CompanyRequest represents the request structure for creating/updating company

	// CompanyRequest represents the request structure for Company.
	CompanyRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Description string     `json:"description,omitempty"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *ModelCore) company() {
	m.Migration = append(m.Migration, &Company{})
	m.CompanyManager = services.NewRepository(services.RepositoryParams[Company, CompanyResponse, CompanyRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media"},
		Service:  m.provider.Service,
		Resource: func(data *Company) *CompanyResponse {
			if data == nil {
				return nil
			}
			return &CompanyResponse{
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
		Created: func(data *Company) []string {
			return []string{
				"company.create",
				fmt.Sprintf("company.create.%s", data.ID),
				fmt.Sprintf("company.create.branch.%s", data.BranchID),
				fmt.Sprintf("company.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Company) []string {
			return []string{
				"company.update",
				fmt.Sprintf("company.update.%s", data.ID),
				fmt.Sprintf("company.update.branch.%s", data.BranchID),
				fmt.Sprintf("company.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Company) []string {
			return []string{
				"company.delete",
				fmt.Sprintf("company.delete.%s", data.ID),
				fmt.Sprintf("company.delete.branch.%s", data.BranchID),
				fmt.Sprintf("company.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) companySeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	branch, err := m.BranchManager.GetByID(context, branchID)
	if err != nil {
		return eris.Wrapf(err, "failed to get branch by ID: %s", branchID)
	}
	organization, err := m.OrganizationManager.GetByID(context, organizationID)
	if err != nil {
		return eris.Wrapf(err, "failed to get organization by ID: %s", organizationID)
	}
	currency, err := m.CurrencyFindByAlpha2(context, branch.CountryCode)
	if err != nil {
		return eris.Wrap(err, "failed to find currency for account seeding")
	}
	companies := []*Company{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           fmt.Sprintf("%s - %s", organization.Name, branch.Name),
			Description:    fmt.Sprintf("The main company of %s located at %s, %s", organization.Name, branch.Address, branch.City),
		},

		// Technology Companies
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Apple Inc.",
			Description:    "American multinational technology company known for iPhone, Mac, and other consumer electronics.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Microsoft Corporation",
			Description:    "Global leader in software, cloud computing, and technology services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Google LLC (Alphabet Inc.)",
			Description:    "Multinational conglomerate specializing in internet-related products and services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Amazon.com, Inc.",
			Description:    "Global e-commerce, cloud computing, and AI company headquartered in Seattle.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Meta Platforms, Inc.",
			Description:    "Parent company of Facebook, Instagram, and WhatsApp.",
		},

		// Automotive Companies
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Toyota Motor Corporation",
			Description:    "Japanese multinational automotive manufacturer and world leader in hybrid vehicles.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Tesla, Inc.",
			Description:    "American company specializing in electric vehicles and clean energy products.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Volkswagen Group",
			Description:    "German multinational automotive manufacturer owning Audi, Porsche, and Lamborghini.",
		},

		// Finance & Banking
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "JPMorgan Chase & Co.",
			Description:    "Largest bank in the United States by assets, offering global financial services.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "HSBC Holdings plc",
			Description:    "British multinational bank serving customers in over 60 countries.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Mastercard Incorporated",
			Description:    "Global payment technology company connecting consumers, financial institutions, and merchants.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Visa Inc.",
			Description:    "Global leader in digital payments and financial technology.",
		},

		// Telecommunications
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "AT&T Inc.",
			Description:    "American multinational telecommunications and media company.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Verizon Communications Inc.",
			Description:    "One of the largest telecommunications companies in the world.",
		},

		// Energy Companies
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "ExxonMobil Corporation",
			Description:    "American multinational oil and gas corporation.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Shell plc",
			Description:    "Global energy and petrochemical company headquartered in London.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "BP (British Petroleum)",
			Description:    "Multinational oil and gas company based in the United Kingdom.",
		},

		// Food & Consumer Goods
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Nestlé S.A.",
			Description:    "Swiss multinational food and beverage company, the largest in the world by revenue.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "The Coca-Cola Company",
			Description:    "American beverage corporation known for its flagship soft drink brand Coca-Cola.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "PepsiCo, Inc.",
			Description:    "Multinational food, snack, and beverage corporation.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Unilever PLC",
			Description:    "British-Dutch multinational consumer goods company known for Dove, Lifebuoy, and Knorr.",
		},

		// Education & Research
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Harvard University",
			Description:    "Private Ivy League research university in Cambridge, Massachusetts.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Massachusetts Institute of Technology (MIT)",
			Description:    "World-renowned research university focused on science and technology.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Stanford University",
			Description:    "Private research university in Stanford, California, known for innovation and entrepreneurship.",
		},

		// Starlink and cables
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "SpaceX",
			Description:    "Aerospace manufacturer and space transport services company founded by Elon Musk.",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Starlink",
			Description:    "Satellite internet constellation being constructed by SpaceX.",
		},
	}

	switch currency.CurrencyCode {
	case "USD": // United States
		companies = append(companies,
			// Electricity / Energy Companies
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Pacific Gas and Electric Company (PG&E)",
				Description:    "One of the largest combined natural gas and electric utilities in the United States, serving Northern and Central California.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Duke Energy Corporation",
				Description:    "Major electric power holding company serving customers in the Southeast and Midwest United States.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Consolidated Edison, Inc. (Con Edison)",
				Description:    "Provides electric, gas, and steam service in New York City and Westchester County.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Florida Power & Light Company (FPL)",
				Description:    "The largest electric utility in Florida, providing power to over 5 million customer accounts.",
			},

			// Water Utilities
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "American Water Works Company, Inc.",
				Description:    "Largest publicly traded U.S. water and wastewater utility company, serving 14 million people across 24 states.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Aqua America (Essential Utilities)",
				Description:    "Provides water and wastewater services to communities in eight U.S. states.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "California Water Service (Cal Water)",
				Description:    "Provides regulated and reliable water services to California communities.",
			},

			// Internet / Cable / Telecom
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Comcast Xfinity",
				Description:    "One of the largest cable television and internet service providers in the U.S.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "AT&T Internet",
				Description:    "Major internet and telecommunications service provider in the United States.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Verizon Fios",
				Description:    "Fiber-optic internet, TV, and phone service operated by Verizon Communications.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Spectrum (Charter Communications)",
				Description:    "Cable television, internet, and phone provider serving millions across the U.S.",
			},

			// Gas Companies
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Southern California Gas Company (SoCalGas)",
				Description:    "The largest natural gas distribution utility in the United States.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "National Grid USA",
				Description:    "Provides natural gas and electricity distribution services in the Northeastern United States.",
			},

			// Waste Management
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Waste Management, Inc.",
				Description:    "Leading provider of waste collection, disposal, and recycling services across the U.S.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Republic Services, Inc.",
				Description:    "Environmental services company providing waste collection and recycling solutions nationwide.",
			},
		)

	case "EUR": // European Union (Germany as representative)
		companies = append(companies,
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "E.ON SE",
				Description:    "One of Europe's largest electric utility service providers headquartered in Essen, Germany.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Deutsche Telekom AG",
				Description:    "Major telecommunications company providing internet, mobile, and landline services across Europe.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Berliner Wasserbetriebe",
				Description:    "The largest water supply and wastewater disposal company in Germany, serving Berlin and surrounding areas.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Vodafone GmbH",
				Description:    "Leading broadband, cable TV, and mobile communications provider based in Düsseldorf, Germany.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "RWE AG",
				Description:    "Energy company focused on electricity generation, renewable energy, and trading headquartered in Essen.",
			},
		)

	case "JPY": // Japan
		companies = append(companies,
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Tokyo Electric Power Company (TEPCO)",
				Description:    "Japan's largest electric utility company providing electricity to the Greater Tokyo Area.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Tokyo Gas Co., Ltd.",
				Description:    "Japan's largest natural gas utility company supplying energy and related services to households and industries in Tokyo.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "NTT Communications Corporation",
				Description:    "Major telecommunications and internet service provider under Nippon Telegraph and Telephone Corporation.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "SoftBank Corp.",
				Description:    "Leading Japanese telecom and internet company providing mobile, broadband, and enterprise network services.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Tokyo Metropolitan Waterworks Bureau",
				Description:    "Official public utility providing clean water supply and wastewater management services in Tokyo.",
			},
		)

	case "GBP": // United Kingdom
		companies = append(companies,
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "British Gas",
				Description:    "The UK's leading energy and home services provider, supplying gas and electricity to millions of households.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Thames Water",
				Description:    "The largest water and wastewater services company in the UK, serving London and surrounding areas.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "BT Group plc",
				Description:    "Formerly British Telecom, BT is one of the UK's main broadband, landline, and TV service providers.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Virgin Media O2",
				Description:    "A major telecom and internet provider offering broadband, mobile, and digital TV services across the UK.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Scottish Power",
				Description:    "A leading UK energy supplier focusing on renewable electricity generation and green energy solutions.",
			},
		)

	case "AUD": // Australia
		companies = append(companies,
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Origin Energy",
				Description:    "One of Australia's leading energy companies providing electricity, natural gas, and solar solutions to homes and businesses.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Sydney Water",
				Description:    "Australia’s largest water utility supplying high-quality drinking water, wastewater, and stormwater services across Sydney.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Telstra Corporation Limited",
				Description:    "Australia’s biggest telecommunications and internet service provider offering broadband, mobile, and digital TV services.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "AGL Energy",
				Description:    "Leading Australian electricity and gas retailer generating power through both traditional and renewable energy sources.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Jemena",
				Description:    "Australian energy infrastructure company managing electricity and gas distribution networks across multiple states.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Spotless Group Holdings",
				Description:    "Property management and maintenance service provider offering cleaning, repairs, and facility support for residential and commercial clients.",
			},
		)

	case "CAD": // Canada
		companies = append(companies,
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Hydro One",
				Description:    "Ontario-based electricity transmission and distribution utility providing power to millions of homes and businesses across Canada.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Enbridge Gas Inc.",
				Description:    "One of Canada’s largest natural gas distributors delivering energy to residential, commercial, and industrial customers nationwide.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Bell Canada",
				Description:    "Leading telecommunications and internet provider offering mobile, broadband, and digital television services throughout Canada.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Rogers Communications",
				Description:    "Major Canadian communications and media company providing internet, cable TV, and mobile services to consumers and businesses.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "Toronto Water",
				Description:    "Municipal water service providing clean water supply and wastewater treatment to residents of Toronto and nearby areas.",
			},
			&Company{
				CreatedAt:      now,
				UpdatedAt:      now,
				CreatedByID:    userID,
				UpdatedByID:    userID,
				OrganizationID: organizationID,
				BranchID:       branchID,
				Name:           "FirstService Corporation",
				Description:    "North American property services and maintenance company providing cleaning, building repair, and residential management solutions.",
			},
		)

	case "CHF": // Switzerland
	case "CNY": // China
	case "SEK": // Sweden
	case "NZD": // New Zealand
	case "PHP": // Philippines
	case "INR": // India
	case "KRW": // South Korea
	case "THB": // Thailand
	case "SGD": // Singapore
	case "HKD": // Hong Kong
	case "MYR": // Malaysia
	case "IDR": // Indonesia
	case "VND": // Vietnam
	case "TWD": // Taiwan
	case "BND": // Brunei
	case "SAR": // Saudi Arabia
	case "AED": // United Arab Emirates
	case "ILS": // Israel
	case "ZAR": // South Africa
	case "EGP": // Egypt
	case "TRY": // Turkey
	case "XOF": // West African CFA Franc
	case "XAF": // Central African CFA Franc
	case "MUR": // Mauritius
	case "MVR": // Maldives
	case "NOK": // Norway
	case "DKK": // Denmark
	case "PLN": // Poland
	case "CZK": // Czech Republic
	case "HUF": // Hungary
	case "RUB": // Russia
	case "EUR-HR": // Croatia (Euro)
	case "BRL": // Brazil
	case "MXN": // Mexico
	case "ARS": // Argentina
	case "CLP": // Chile
	case "COP": // Colombia
	case "PEN": // Peru
	case "UYU": // Uruguay
	case "DOP": // Dominican Republic
	case "PYG": // Paraguay
	case "BOB": // Bolivia
	case "VES": // Venezuela
	case "PKR": // Pakistan
	case "BDT": // Bangladesh
	case "LKR": // Sri Lanka
	case "NPR": // Nepal
	case "MMK": // Myanmar
	case "KHR": // Cambodia
	case "LAK": // Laos
	case "NGN": // Nigeria
	case "KES": // Kenya
	case "GHS": // Ghana
	case "MAD": // Morocco
	case "TND": // Tunisia
	case "ETB": // Ethiopia
	case "DZD": // Algeria
	case "UAH": // Ukraine
	case "RON": // Romania
	case "BGN": // Bulgaria
	case "RSD": // Serbia
	case "ISK": // Iceland
	case "BYN": // Belarus
	case "FJD": // Fiji
	case "PGK": // Papua New Guinea
	case "JMD": // Jamaica
	case "CRC": // Costa Rica
	case "GTQ": // Guatemala
	case "XDR": // Special Drawing Rights (IMF)
	case "KWD": // Kuwait
	case "QAR": // Qatar
	case "OMR": // Oman
	case "BHD": // Bahrain
	case "JOD": // Jordan
	case "KZT": // Kazakhstan
	}

	for _, data := range companies {
		if err := m.CompanyManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed company %s", data.Name)
		}
	}
	return nil
}

// CompanyCurrentBranch returns all companies for the given organization and branch.
func (m *ModelCore) CompanyCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*Company, error) {
	return m.CompanyManager.Find(context, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
