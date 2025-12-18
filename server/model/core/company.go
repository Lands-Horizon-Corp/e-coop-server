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

	CompanyRequest struct {
		Name        string     `json:"name" validate:"required,min=1,max=255"`
		Description string     `json:"description,omitempty"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
	}
)

func (m *Core) company() {
	m.Migration = append(m.Migration, &Company{})
	m.CompanyManager = *registry.NewRegistry(registry.RegistryParams[Company, CompanyResponse, CompanyRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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
		Created: func(data *Company) registry.Topics {
			return []string{
				"company.create",
				fmt.Sprintf("company.create.%s", data.ID),
				fmt.Sprintf("company.create.branch.%s", data.BranchID),
				fmt.Sprintf("company.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Company) registry.Topics {
			return []string{
				"company.update",
				fmt.Sprintf("company.update.%s", data.ID),
				fmt.Sprintf("company.update.branch.%s", data.BranchID),
				fmt.Sprintf("company.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Company) registry.Topics {
			return []string{
				"company.delete",
				fmt.Sprintf("company.delete.%s", data.ID),
				fmt.Sprintf("company.delete.branch.%s", data.BranchID),
				fmt.Sprintf("company.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) companySeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	branch, err := m.BranchManager.GetByID(context, branchID)
	if err != nil {
		return eris.Wrapf(err, "failed to get branch by ID: %s", branchID)
	}
	organization, err := m.OrganizationManager.GetByID(context, organizationID)
	if err != nil {
		return eris.Wrapf(err, "failed to get organization by ID: %s", organizationID)
	}

	companies := []*Company{
		{

			Name:        fmt.Sprintf("%s - %s", organization.Name, branch.Name),
			Description: fmt.Sprintf("The main company of %s located at %s, %s", organization.Name, branch.Address, branch.City),
		},
		{
			Name:        "Apple Inc.",
			Description: "American multinational technology company known for iPhone, Mac, and other consumer electronics.",
		},
		{
			Name:        "Microsoft Corporation",
			Description: "Global leader in software, cloud computing, and technology services.",
		},
		{
			Name:        "Google LLC (Alphabet Inc.)",
			Description: "Multinational conglomerate specializing in internet-related products and services.",
		},
		{
			Name:        "Amazon.com, Inc.",
			Description: "Global e-commerce, cloud computing, and AI company headquartered in Seattle.",
		},
		{
			Name:        "Meta Platforms, Inc.",
			Description: "Parent company of Facebook, Instagram, and WhatsApp.",
		},

		{
			Name:        "Toyota Motor Corporation",
			Description: "Japanese multinational automotive manufacturer and world leader in hybrid vehicles.",
		},
		{
			Name:        "Tesla, Inc.",
			Description: "American company specializing in electric vehicles and clean energy products.",
		},
		{
			Name:        "Volkswagen Group",
			Description: "German multinational automotive manufacturer owning Audi, Porsche, and Lamborghini.",
		},

		{
			Name:        "JPMorgan Chase & Co.",
			Description: "Largest bank in the United States by assets, offering global financial services.",
		},
		{
			Name:        "HSBC Holdings plc",
			Description: "British multinational bank serving customers in over 60 countries.",
		},
		{
			Name:        "Mastercard Incorporated",
			Description: "Global payment technology company connecting consumers, financial institutions, and merchants.",
		},
		{
			Name:        "Visa Inc.",
			Description: "Global leader in digital payments and financial technology.",
		},

		{
			Name:        "AT&T Inc.",
			Description: "American multinational telecommunications and media company.",
		},
		{
			Name:        "Verizon Communications Inc.",
			Description: "One of the largest telecommunications companies in the world.",
		},

		{
			Name:        "ExxonMobil Corporation",
			Description: "American multinational oil and gas corporation.",
		},
		{
			Name:        "Shell plc",
			Description: "Global energy and petrochemical company headquartered in London.",
		},
		{
			BranchID:    branchID,
			Name:        "BP (British Petroleum)",
			Description: "Multinational oil and gas company based in the United Kingdom.",
		},

		{
			Name:        "Nestlé S.A.",
			Description: "Swiss multinational food and beverage company, the largest in the world by revenue.",
		},
		{
			Name:        "The Coca-Cola Company",
			Description: "American beverage corporation known for its flagship soft drink brand Coca-Cola.",
		},
		{
			Name:        "PepsiCo, Inc.",
			Description: "Multinational food, snack, and beverage corporation.",
		},
		{
			Name:        "Unilever PLC",
			Description: "British-Dutch multinational consumer goods company known for Dove, Lifebuoy, and Knorr.",
		},

		{
			Name:        "Harvard University",
			Description: "Private Ivy League research university in Cambridge, Massachusetts.",
		},
		{
			Name:        "Massachusetts Institute of Technology (MIT)",
			Description: "World-renowned research university focused on science and technology.",
		},
		{
			Name:        "Stanford University",
			Description: "Private research university in Stanford, California, known for innovation and entrepreneurship.",
		},

		{
			Name:        "SpaceX",
			Description: "Aerospace manufacturer and space transport services company founded by Elon Musk.",
		},
		{
			Name:        "Starlink",
			Description: "Satellite internet constellation being constructed by SpaceX.",
		},
	}

	switch branch.Currency.ISO3166Alpha3 {
	case "USA": // United States
		companies = append(companies,
			&Company{

				Name:        "Pacific Gas and Electric Company (PG&E)",
				Description: "One of the largest combined natural gas and electric utilities in the United States, serving Northern and Central California.",
			},
			&Company{

				Name:        "Duke Energy Corporation",
				Description: "Major electric power holding company serving customers in the Southeast and Midwest United States.",
			},
			&Company{

				Name:        "Consolidated Edison, Inc. (Con Edison)",
				Description: "Provides electric, gas, and steam service in New York City and Westchester County.",
			},
			&Company{

				Name:        "Florida Power & Light Company (FPL)",
				Description: "The largest electric utility in Florida, providing power to over 5 million customer accounts.",
			},

			&Company{

				Name:        "American Water Works Company, Inc.",
				Description: "Largest publicly traded U.S. water and wastewater utility company, serving 14 million people across 24 states.",
			},
			&Company{

				Name:        "Aqua America (Essential Utilities)",
				Description: "Provides water and wastewater services to communities in eight U.S. states.",
			},
			&Company{

				Name:        "California Water Service (Cal Water)",
				Description: "Provides regulated and reliable water services to California communities.",
			},

			&Company{

				Name:        "Comcast Xfinity",
				Description: "One of the largest cable television and internet service providers in the U.S.",
			},
			&Company{

				Name:        "AT&T Internet",
				Description: "Major internet and telecommunications service provider in the United States.",
			},
			&Company{

				Name:        "Verizon Fios",
				Description: "Fiber-optic internet, TV, and phone service operated by Verizon Communications.",
			},
			&Company{

				Name:        "Spectrum (Charter Communications)",
				Description: "Cable television, internet, and phone provider serving millions across the U.S.",
			},

			&Company{

				Name:        "Southern California Gas Company (SoCalGas)",
				Description: "The largest natural gas distribution utility in the United States.",
			},
			&Company{

				Name:        "National Grid USA",
				Description: "Provides natural gas and electricity distribution services in the Northeastern United States.",
			},

			&Company{

				Name:        "Waste Management, Inc.",
				Description: "Leading provider of waste collection, disposal, and recycling services across the U.S.",
			},
			&Company{

				Name:        "Republic Services, Inc.",
				Description: "Environmental services company providing waste collection and recycling solutions nationwide.",
			},
		)

	case "DEU": // Germany
		companies = append(companies,
			&Company{

				Name:        "E.ON SE",
				Description: "One of Europe's largest electric utility service providers headquartered in Essen, Germany.",
			},
			&Company{

				Name:        "Deutsche Telekom AG",
				Description: "Major telecommunications company providing internet, mobile, and landline services across Europe.",
			},
			&Company{

				Name:        "Berliner Wasserbetriebe",
				Description: "The largest water supply and wastewater disposal company in Germany, serving Berlin and surrounding areas.",
			},
			&Company{

				Name:        "Vodafone GmbH",
				Description: "Leading broadband, cable TV, and mobile communications provider based in Düsseldorf, Germany.",
			},
			&Company{

				Name:        "RWE AG",
				Description: "Energy company focused on electricity generation, renewable energy, and trading headquartered in Essen.",
			},
		)

	case "JPN": // Japan
		companies = append(companies,
			&Company{

				Name:        "Tokyo Electric Power Company (TEPCO)",
				Description: "Japan's largest electric utility company providing electricity to the Greater Tokyo Area.",
			},
			&Company{

				Name:        "Tokyo Gas Co., Ltd.",
				Description: "Japan's largest natural gas utility company supplying energy and related services to households and industries in Tokyo.",
			},
			&Company{

				Name:        "NTT Communications Corporation",
				Description: "Major telecommunications and internet service provider under Nippon Telegraph and Telephone Corporation.",
			},
			&Company{

				Name:        "SoftBank Corp.",
				Description: "Leading Japanese telecom and internet company providing mobile, broadband, and enterprise network services.",
			},
			&Company{

				Name:        "Tokyo Metropolitan Waterworks Bureau",
				Description: "Official public utility providing clean water supply and wastewater management services in Tokyo.",
			},
		)

	case "GBR": // United Kingdom
		companies = append(companies,
			&Company{

				Name:        "British Gas",
				Description: "The UK's leading energy and home services provider, supplying gas and electricity to millions of households.",
			},
			&Company{

				Name:        "Thames Water",
				Description: "The largest water and wastewater services company in the UK, serving London and surrounding areas.",
			},
			&Company{

				Name:        "BT Group plc",
				Description: "Formerly British Telecom, BT is one of the UK's main broadband, landline, and TV service providers.",
			},
			&Company{

				Name:        "Virgin Media O2",
				Description: "A major telecom and internet provider offering broadband, mobile, and digital TV services across the UK.",
			},
			&Company{

				Name:        "Scottish Power",
				Description: "A leading UK energy supplier focusing on renewable electricity generation and green energy solutions.",
			},
		)

	case "AUS": // Australia
		companies = append(companies,
			&Company{

				Name:        "Origin Energy",
				Description: "One of Australia's leading energy companies providing electricity, natural gas, and solar solutions to homes and businesses.",
			},
			&Company{

				Name:        "Sydney Water",
				Description: "Australia’s largest water utility supplying high-quality drinking water, wastewater, and stormwater services across Sydney.",
			},
			&Company{

				Name:        "Telstra Corporation Limited",
				Description: "Australia’s biggest telecommunications and internet service provider offering broadband, mobile, and digital TV services.",
			},
			&Company{

				Name:        "AGL Energy",
				Description: "Leading Australian electricity and gas retailer generating power through both traditional and renewable energy sources.",
			},
			&Company{

				Name:        "Jemena",
				Description: "Australian energy infrastructure company managing electricity and gas distribution networks across multiple states.",
			},
			&Company{

				Name:        "Spotless Group Holdings",
				Description: "Property management and maintenance service provider offering cleaning, repairs, and facility support for residential and commercial clients.",
			},
		)

	case "CAN": // Canada
		companies = append(companies,
			&Company{

				Name:        "Hydro One",
				Description: "Ontario-based electricity transmission and distribution utility providing power to millions of homes and businesses across Canada.",
			},
			&Company{

				Name:        "Enbridge Gas Inc.",
				Description: "One of Canada’s largest natural gas distributors delivering energy to residential, commercial, and industrial customers nationwide.",
			},
			&Company{

				Name:        "Bell Canada",
				Description: "Leading telecommunications and internet provider offering mobile, broadband, and digital television services throughout Canada.",
			},
			&Company{

				Name:        "Rogers Communications",
				Description: "Major Canadian communications and media company providing internet, cable TV, and mobile services to consumers and businesses.",
			},
			&Company{

				Name:        "Toronto Water",
				Description: "Municipal water service providing clean water supply and wastewater treatment to residents of Toronto and nearby areas.",
			},
			&Company{
				Name:        "FirstService Corporation",
				Description: "North American property services and maintenance company providing cleaning, building repair, and residential management solutions.",
			},
		)

	case "CHE": // Switzerland
		companies = append(companies,
			&Company{

				Name:        "Swissgrid AG",
				Description: "The national electricity transmission system operator responsible for maintaining and managing Switzerland’s power grid.",
			},
			&Company{

				Name:        "BKW Energie AG",
				Description: "Major Swiss energy company providing electricity, renewable energy solutions, and infrastructure services.",
			},
			&Company{

				Name:        "Gaznat SA",
				Description: "Swiss natural gas supplier managing transportation, storage, and distribution for western Switzerland.",
			},
			&Company{

				Name:        "SIG (Services Industriels de Genève)",
				Description: "Public utility of Geneva providing water, gas, electricity, and renewable energy services to households and businesses.",
			},
			&Company{

				Name:        "Swisscom AG",
				Description: "Switzerland’s leading telecommunications and internet service provider offering broadband, mobile, and TV services.",
			},
			&Company{

				Name:        "Sunrise GmbH",
				Description: "Major telecom company providing mobile, internet, and digital TV services across Switzerland.",
			},
			&Company{

				Name:        "ISS Facility Services AG",
				Description: "Swiss-based provider of integrated facility management, cleaning, and building maintenance services.",
			},
			&Company{

				Name:        "Bouygues Energies & Services Schweiz AG",
				Description: "Leading Swiss company specializing in energy, building maintenance, and facility management solutions.",
			},
		)

	case "CHN": // China
		companies = append(companies,
			&Company{

				Name:        "State Grid Corporation of China",
				Description: "The world’s largest electric utility company providing electricity transmission and distribution services across China.",
			},
			&Company{

				Name:        "China Southern Power Grid",
				Description: "Electric utility supplying and managing power networks in southern provinces such as Guangdong, Guangxi, and Yunnan.",
			},
			&Company{

				Name:        "China Gas Holdings Limited",
				Description: "Leading natural gas distributor providing piped gas, LPG, and related energy services across Chinese cities.",
			},
			&Company{

				Name:        "Beijing Gas Group Co., Ltd.",
				Description: "Major urban gas supplier providing clean energy and heating services for residential and industrial users in Beijing.",
			},
			&Company{

				Name:        "Beijing Waterworks Group",
				Description: "Public utility responsible for water supply, sewage treatment, and pipeline management in Beijing.",
			},
			&Company{

				Name:        "Shanghai Municipal Waterworks",
				Description: "Major water utility company providing clean water distribution and wastewater treatment in Shanghai.",
			},
			&Company{

				Name:        "China Telecom",
				Description: "One of the largest telecommunications companies in China offering broadband, mobile, and digital TV services.",
			},
			&Company{
				Name:        "China Unicom",
				Description: "Telecom operator providing mobile, fixed-line, and internet services to consumers and enterprises across China.",
			},
			&Company{
				Name:        "China Mobile",
				Description: "The largest mobile and broadband network provider in China offering nationwide communication services.",
			},
			&Company{
				Name:        "Country Garden Services Holdings",
				Description: "Top property management and maintenance company offering residential cleaning, repair, and facility services.",
			},
			&Company{
				Name:        "China Overseas Property Holdings Limited",
				Description: "Leading facility management company providing residential and commercial property services throughout China.",
			},
		)

	case "SWE": // Sweden
		companies = append(companies,
			&Company{
				Name:        "Vattenfall AB",
				Description: "State-owned energy company supplying electricity, heat, and energy solutions across Sweden and Europe.",
			},
			&Company{
				Name:        "E.ON Sverige AB",
				Description: "Major energy provider in Sweden offering electricity, heating, and renewable energy services.",
			},
			&Company{
				Name:        "Göteborg Energi AB",
				Description: "Energy company providing electricity, district heating, cooling, and natural gas distribution in western Sweden.",
			},
			&Company{
				Name:        "Stockholm Vatten och Avfall",
				Description: "Public utility managing water supply and waste services for the Stockholm region.",
			},
			&Company{
				Name:        "Telia Company AB",
				Description: "Sweden’s largest telecommunications and broadband provider offering internet, mobile, and TV services.",
			},
			&Company{
				Name:        "Com Hem (Tele2 AB)",
				Description: "Leading broadband and cable TV provider serving homes across Sweden.",
			},
			&Company{
				Name:        "Coor Service Management AB",
				Description: "Integrated facilities management company providing cleaning, maintenance, and workplace services.",
			},
		)

	case "NZL": // New Zealand
		companies = append(companies,
			&Company{
				Name:        "Genesis Energy Limited",
				Description: "One of New Zealand's largest electricity and gas suppliers serving homes and businesses nationwide.",
			},
			&Company{
				Name:        "Contact Energy Limited",
				Description: "Major provider of electricity, natural gas, and broadband services across New Zealand.",
			},
			&Company{
				Name:        "Watercare Services Limited",
				Description: "Auckland-based public utility responsible for water supply and wastewater treatment.",
			},
			&Company{
				Name:        "Chorus Limited",
				Description: "New Zealand’s primary telecommunications infrastructure company providing broadband and fiber connections.",
			},
			&Company{
				Name:        "Spark New Zealand",
				Description: "Leading telecommunications company offering internet, mobile, and digital services.",
			},
			&Company{
				Name:        "Downer New Zealand",
				Description: "Infrastructure and facilities management company providing maintenance, utilities, and asset management services.",
			},
		)

	case "PHL": // Philippines
		companies = append(companies,
			&Company{
				Name:        "Manila Electric Company (Meralco)",
				Description: "Largest electric distribution utility serving Metro Manila and surrounding provinces.",
			},
			&Company{
				Name:        "Maynilad Water Services, Inc.",
				Description: "Provides water and wastewater services to the West Zone of Metro Manila.",
			},
			&Company{
				Name:        "Manila Water Company, Inc.",
				Description: "Provides water distribution and sanitation services to the East Zone of Metro Manila.",
			},
			&Company{
				Name:        "PLDT Inc.",
				Description: "Major telecommunications and internet service provider in the Philippines.",
			},
			&Company{
				Name:        "Globe Telecom, Inc.",
				Description: "Leading mobile and broadband provider offering telecommunications and data services.",
			},
			&Company{
				Name:        "Converge ICT Solutions, Inc.",
				Description: "Fiber internet provider known for high-speed residential and business broadband services.",
			},
			&Company{
				Name:        "DMCI Power Corporation",
				Description: "Independent power producer and energy distribution company serving off-grid areas.",
			},
			&Company{
				Name:        "Ayala Property Management Corporation (APMC)",
				Description: "Provides property and facilities management services for residential and commercial buildings.",
			},
		)

	case "IND": // India
		companies = append(companies,
			&Company{
				Name:        "Tata Power Company Limited",
				Description: "One of India's largest integrated power companies providing electricity generation and distribution.",
			},
			&Company{
				Name:        "BSES Rajdhani Power Limited",
				Description: "Delhi-based power distribution company serving residential and commercial customers.",
			},
			&Company{
				Name:        "Indraprastha Gas Limited (IGL)",
				Description: "Leading natural gas distribution company supplying CNG and PNG in Delhi NCR.",
			},
			&Company{
				Name:        "Bharti Airtel Limited",
				Description: "Telecommunications giant offering mobile, broadband, and digital TV services.",
			},
			&Company{
				Name:        "Reliance Jio Infocomm Limited",
				Description: "Major telecom provider offering 4G/5G mobile and fiber internet services.",
			},
			&Company{
				Name:        "Hindustan Unilever Facility Services",
				Description: "Facility and maintenance management company catering to industrial and commercial clients.",
			},
			&Company{
				Name:        "Delhi Jal Board",
				Description: "Government agency responsible for water supply and wastewater treatment in Delhi.",
			},
		)

	case "KOR": // South Korea
		companies = append(companies,
			&Company{
				Name:        "Korea Electric Power Corporation (KEPCO)",
				Description: "South Korea’s main electric utility responsible for power generation and distribution nationwide.",
			},
			&Company{
				Name:        "Seoul Waterworks Authority",
				Description: "Public water utility providing clean water and wastewater management in Seoul.",
			},
			&Company{
				Name:        "SK Broadband Co., Ltd.",
				Description: "Major internet and IPTV provider under SK Group.",
			},
			&Company{
				Name:        "KT Corporation",
				Description: "Leading telecommunications company offering mobile, internet, and digital services.",
			},
			&Company{
				Name:        "LG Uplus Corp.",
				Description: "Telecommunications company providing mobile, broadband, and enterprise network services.",
			},
			&Company{
				Name:        "GS Caltex Corporation",
				Description: "Oil and gas company also involved in energy supply and petrochemical services.",
			},
			&Company{
				Name:        "Hanmi Global Inc.",
				Description: "Facility and project management firm offering maintenance and engineering services.",
			},
		)

	case "THA": // Thailand
		companies = append(companies,
			&Company{
				Name:        "Metropolitan Electricity Authority (MEA)",
				Description: "Government agency providing electricity to Bangkok and nearby provinces.",
			},
			&Company{
				Name:        "Provincial Electricity Authority (PEA)",
				Description: "Electric utility distributing power across Thailand’s provinces.",
			},
			&Company{
				Name:        "Metropolitan Waterworks Authority (MWA)",
				Description: "Public utility responsible for clean water supply in Bangkok and surrounding areas.",
			},
			&Company{
				Name:        "TOT Public Company Limited (National Telecom)",
				Description: "State-owned telecommunications provider offering internet and mobile services.",
			},
			&Company{
				Name:        "True Corporation",
				Description: "Major telecom and broadband provider offering internet, TV, and mobile services.",
			},
			&Company{
				Name:        "PTT Public Company Limited",
				Description: "Thailand’s national oil and gas company supplying natural gas and energy solutions.",
			},
			&Company{
				Name:        "Jones Lang LaSalle (Thailand)",
				Description: "Facilities and property management company providing maintenance and cleaning services.",
			},
		)
	case "SGP": // Singapore
		companies = append(companies,
			&Company{
				Name:        "SP Group",
				Description: "Singapore’s national utilities group providing electricity, gas, and sustainable energy solutions.",
			},
			&Company{
				Name:        "PUB, Singapore’s National Water Agency",
				Description: "Government agency managing water supply, drainage, and wastewater treatment.",
			},
			&Company{
				Name:        "Singtel",
				Description: "Singapore’s largest telecommunications company offering mobile, internet, and TV services.",
			},
			&Company{
				Name:        "StarHub",
				Description: "Telecom provider offering broadband, mobile, and entertainment services.",
			},
			&Company{
				Name:        "M1 Limited",
				Description: "Integrated communications provider delivering mobile and fiber broadband services.",
			},
			&Company{
				Name:        "Keppel Infrastructure",
				Description: "Company providing energy, utilities, and environmental infrastructure services.",
			},
			&Company{
				Name:        "CBM Pte Ltd",
				Description: "Facilities management company offering maintenance, cleaning, and building services.",
			},
		)

	case "HKG": // Hong Kong
		companies = append(companies,
			&Company{
				Name:        "CLP Power Hong Kong Limited",
				Description: "Major electricity utility supplying power to Kowloon, New Territories, and Lantau.",
			},
			&Company{
				Name:        "The Hongkong Electric Company Limited (HK Electric)",
				Description: "Electricity provider serving Hong Kong Island and Lamma Island.",
			},
			&Company{
				Name:        "Towngas (The Hong Kong and China Gas Company Limited)",
				Description: "Provides town gas and energy solutions across Hong Kong.",
			},
			&Company{
				Name:        "Hong Kong Water Supplies Department",
				Description: "Government department responsible for water supply and management.",
			},
			&Company{
				Name:        "PCCW-HKT",
				Description: "Integrated telecommunications provider offering broadband, mobile, and media services.",
			},
			&Company{
				Name:        "SmarTone",
				Description: "Mobile network operator providing 4G/5G and internet services.",
			},
			&Company{
				Name:        "ISS Facility Services Hong Kong",
				Description: "Company offering property, cleaning, and maintenance services for commercial clients.",
			},
		)

	case "MYS": // Malaysia
		companies = append(companies,
			&Company{
				Name:        "Tenaga Nasional Berhad (TNB)",
				Description: "Malaysia’s largest electricity utility company providing power generation and distribution.",
			},
			&Company{
				Name:        "Syarikat Air Selangor Sdn Bhd",
				Description: "State-owned water company responsible for water supply in Selangor and Kuala Lumpur.",
			},
			&Company{
				Name:        "Petronas Gas Berhad",
				Description: "Subsidiary of Petronas providing gas processing and utilities services.",
			},
			&Company{
				Name:        "TM (Telekom Malaysia Berhad)",
				Description: "Malaysia’s leading broadband and telecommunications provider.",
			},
			&Company{
				Name:        "Maxis Berhad",
				Description: "Telecommunications company offering mobile and internet services.",
			},
			&Company{
				Name:        "UEM Edgenta Berhad",
				Description: "Facilities management and infrastructure maintenance service provider.",
			},
		)

	case "IDN": // Indonesia
		companies = append(companies,
			&Company{
				Name:        "Perusahaan Listrik Negara (PLN)",
				Description: "State-owned electricity company responsible for power generation and distribution.",
			},
			&Company{
				Name:        "Perusahaan Daerah Air Minum (PDAM)",
				Description: "Regional government-owned companies supplying water across Indonesia.",
			},
			&Company{
				Name:        "Pertamina Gas (Pertagas)",
				Description: "Subsidiary of Pertamina providing gas distribution and infrastructure services.",
			},
			&Company{
				Name:        "Telkom Indonesia",
				Description: "Indonesia’s largest telecommunications and broadband provider.",
			},
			&Company{
				Name:        "Indosat Ooredoo Hutchison",
				Description: "Telecommunications company providing mobile and internet services.",
			},
			&Company{
				Name:        "ISS Indonesia",
				Description: "Facilities and maintenance management services provider for industrial and commercial clients.",
			},
		)

	case "VNM": // Vietnam
		companies = append(companies,
			&Company{
				Name:        "Vietnam Electricity (EVN)",
				Description: "State-owned power company managing electricity generation and distribution.",
			},
			&Company{
				Name:        "Saigon Water Corporation (SAWACO)",
				Description: "Major water supply company serving Ho Chi Minh City.",
			},
			&Company{
				Name:        "PetroVietnam Gas (PV Gas)",
				Description: "Vietnam’s leading natural gas and energy provider.",
			},
			&Company{
				Name:        "Viettel Group",
				Description: "Telecommunications and internet service provider owned by the Ministry of Defense.",
			},
			&Company{
				Name:        "VNPT (Vietnam Posts and Telecommunications Group)",
				Description: "Government-owned telecom operator offering internet and communication services.",
			},
			&Company{
				Name:        "CBRE Vietnam",
				Description: "Facilities and property management company offering building maintenance services.",
			},
		)

	case "TWN": // Taiwan
		companies = append(companies,
			&Company{
				Name:        "Taiwan Power Company (Taipower)",
				Description: "State-owned company providing electricity generation and distribution.",
			},
			&Company{
				Name:        "Taiwan Water Corporation",
				Description: "National water utility responsible for water supply across Taiwan.",
			},
			&Company{
				Name:        "CPC Corporation",
				Description: "State-owned petroleum and gas company providing fuel and natural gas services.",
			},
			&Company{
				Name:        "Chunghwa Telecom Co., Ltd.",
				Description: "Taiwan’s largest telecom company offering internet, mobile, and data services.",
			},
			&Company{
				Name:        "Taiwan Mobile Co., Ltd.",
				Description: "Leading telecom provider offering broadband and mobile services.",
			},
			&Company{
				Name:        "Shin Kong Property Management Co., Ltd.",
				Description: "Facilities management and maintenance service company.",
			},
		)

	case "BRN": // Brunei
		companies = append(companies,
			&Company{
				Name:        "Department of Electrical Services (DES)",
				Description: "Government agency providing electricity services across Brunei.",
			},
			&Company{
				Name:        "Public Works Department (Jabatan Kerja Raya)",
				Description: "Responsible for water supply and infrastructure maintenance in Brunei.",
			},
			&Company{
				Name:        "Brunei Gas Carrier Sdn Bhd (BGC)",
				Description: "Provides gas transport and related energy services.",
			},
			&Company{
				Name:        "Imagine Sdn Bhd",
				Description: "Telecommunications company offering internet and mobile services.",
			},
			&Company{
				Name:        "Datastream Digital (DST)",
				Description: "Brunei’s major telecom provider for mobile and broadband services.",
			},
			&Company{
				Name:        "Armada Properties Sdn Bhd",
				Description: "Property and facilities management company offering maintenance and building services.",
			},
		)

	case "SAU": // Saudi Arabia
		companies = append(companies,
			&Company{
				Name:        "Saudi Electricity Company (SEC)",
				Description: "Kingdom’s main electric utility providing generation and distribution services.",
			},
			&Company{
				Name:        "National Water Company (NWC)",
				Description: "Government-owned company managing water supply and wastewater services.",
			},
			&Company{
				Name:        "Saudi Aramco Gas Operations",
				Description: "Division of Saudi Aramco responsible for natural gas distribution and processing.",
			},
			&Company{
				Name:        "Saudi Telecom Company (stc)",
				Description: "Leading telecom provider offering mobile, internet, and enterprise solutions.",
			},
			&Company{
				Name:        "Mobily (Etihad Etisalat)",
				Description: "Telecom and broadband service provider serving residential and business customers.",
			},
			&Company{
				Name:        "Initial Saudi Group",
				Description: "Facilities management and cleaning services provider across Saudi Arabia.",
			},
		)

	case "ARE": // United Arab Emirates
		companies = append(companies,
			&Company{
				Name:        "Dubai Electricity and Water Authority (DEWA)",
				Description: "Provides electricity, water, and sustainable energy solutions for Dubai.",
			},
			&Company{
				Name:        "Abu Dhabi Distribution Company (ADDC)",
				Description: "Distributes water and electricity in Abu Dhabi and nearby regions.",
			},
			&Company{
				Name:        "ENOC Group",
				Description: "Energy company involved in oil, gas, and fuel distribution.",
			},
			&Company{
				Name:        "Etisalat by e&",
				Description: "Major telecom company offering mobile, internet, and digital services.",
			},
			&Company{
				Name:        "du (Emirates Integrated Telecommunications Company)",
				Description: "Telecom operator providing mobile, broadband, and home services.",
			},
			&Company{
				Name:        "Farnek Services LLC",
				Description: "Facilities management and building maintenance provider in the UAE.",
			},
		)

	case "ISR": // Israel
		companies = append(companies,
			&Company{
				Name:        "Israel Electric Corporation (IEC)",
				Description: "Government-owned electric utility responsible for generation and supply.",
			},
			&Company{
				Name:        "Mekorot Water Company",
				Description: "National water company managing water supply and desalination systems.",
			},
			&Company{
				Name:        "Tamar Petroleum Ltd.",
				Description: "Natural gas supplier serving power and industrial sectors.",
			},
			&Company{
				Name:        "Bezeq Telecommunications Company Ltd.",
				Description: "Israel’s leading telecom and internet provider.",
			},
			&Company{
				Name:        "Cellcom Israel Ltd.",
				Description: "Telecommunications provider offering mobile, internet, and TV services.",
			},
			&Company{
				Name:        "CBRE Israel",
				Description: "Facilities and property management company providing maintenance services.",
			},
		)

	case "ZAF": // South Africa
		companies = append(companies,
			&Company{
				Name:        "Eskom Holdings SOC Ltd",
				Description: "South Africa’s state-owned electricity utility responsible for generation and distribution.",
			},
			&Company{
				Name:        "Johannesburg Water (SOC) Ltd",
				Description: "Municipal-owned company providing water and sanitation services in Johannesburg.",
			},
			&Company{
				Name:        "Sasol Gas (Pty) Ltd",
				Description: "Natural gas and energy solutions provider for industrial and domestic customers.",
			},
			&Company{
				Name:        "Telkom SA SOC Ltd",
				Description: "Telecommunications company offering broadband, fixed-line, and mobile services.",
			},
			&Company{
				Name:        "Vodacom Group Ltd",
				Description: "Leading mobile and internet service provider in South Africa.",
			},
			&Company{
				Name:        "Servest Group (Pty) Ltd",
				Description: "Facilities management company offering cleaning, maintenance, and landscaping services.",
			},
		)

	case "EGY": // Egypt
		companies = append(companies,
			&Company{
				Name:        "Egyptian Electricity Holding Company (EEHC)",
				Description: "National company managing electricity generation, transmission, and distribution.",
			},
			&Company{
				Name:        "Holding Company for Water and Wastewater (HCWW)",
				Description: "State-owned company responsible for water supply and sanitation services.",
			},
			&Company{
				Name:        "Town Gas Company",
				Description: "Provides natural gas distribution for residential and commercial use in Egypt.",
			},
			&Company{
				Name:        "Telecom Egypt (WE)",
				Description: "Main telecommunications and internet service provider in Egypt.",
			},
			&Company{
				Name:        "Orange Egypt",
				Description: "Mobile and broadband company offering telecom and digital services.",
			},
			&Company{
				Name:        "Arab Contractors (Osman Ahmed Osman & Co.)",
				Description: "Construction and facilities maintenance company providing building and infrastructure services.",
			},
		)

	case "TUR": // Turkey
		companies = append(companies,
			&Company{
				Name:        "Turkish Electricity Distribution Corporation (TEDAŞ)",
				Description: "Government-owned electricity distribution company serving Turkey.",
			},
			&Company{
				Name:        "İSKİ (Istanbul Water and Sewerage Administration)",
				Description: "Provides water supply and wastewater management for Istanbul.",
			},
			&Company{
				Name:        "BOTAŞ Petroleum Pipeline Corporation",
				Description: "State-owned natural gas transmission and distribution company.",
			},
			&Company{
				Name:        "Türk Telekom",
				Description: "National telecommunications and internet services provider.",
			},
			&Company{
				Name:        "Vodafone Turkey",
				Description: "Mobile and internet provider serving millions across Turkey.",
			},
			&Company{
				Name:        "ISS Turkey",
				Description: "Facilities management company offering maintenance, cleaning, and property services.",
			},
		)

	case "SEN": // Senegal (West African CFA)
		companies = append(companies,
			&Company{
				Name:        "Compagnie Ivoirienne d'Électricité (CIE)",
				Description: "Electricity company responsible for power generation and distribution in Côte d'Ivoire.",
			},
			&Company{
				Name:        "Société Nationale des Eaux du Sénégal (SONES)",
				Description: "Manages water production and distribution infrastructure in Senegal.",
			},
			&Company{
				Name:        "Senelec",
				Description: "State-owned electricity provider for Senegal.",
			},
			&Company{
				Name:        "Orange Côte d’Ivoire",
				Description: "Telecommunications provider offering mobile, internet, and payment services.",
			},
			&Company{
				Name:        "MTN Côte d’Ivoire",
				Description: "Mobile and broadband network operator in West Africa.",
			},
			&Company{
				Name:        "ENGIE Services Afrique de l’Ouest",
				Description: "Provides maintenance, energy efficiency, and facility management solutions.",
			},
		)

	case "CMR": // Cameroon (Central African CFA)
		companies = append(companies,
			&Company{
				Name:        "Eneo Cameroon S.A.",
				Description: "Cameroon’s primary electricity supplier responsible for power generation and distribution.",
			},
			&Company{
				Name:        "Camwater (Cameroon Water Utilities Corporation)",
				Description: "Manages water supply and infrastructure across Cameroon.",
			},
			&Company{
				Name:        "Société d’Énergie et d’Eau du Gabon (SEEG)",
				Description: "Provides water and electricity services throughout Gabon.",
			},
			&Company{
				Name:        "MTN Cameroon",
				Description: "Mobile and internet service provider across Central Africa.",
			},
			&Company{
				Name:        "Orange Cameroun",
				Description: "Telecom company offering mobile and data services.",
			},
			&Company{
				Name:        "Veolia Africa",
				Description: "International company providing water, waste, and energy management services in Africa.",
			},
		)

	case "MUS": // Mauritius
		companies = append(companies,
			&Company{
				Name:        "Central Electricity Board (CEB)",
				Description: "National electricity provider managing generation and distribution in Mauritius.",
			},
			&Company{
				Name:        "Central Water Authority (CWA)",
				Description: "Responsible for water supply and distribution across Mauritius.",
			},
			&Company{
				Name:        "Mauritius Telecom Ltd",
				Description: "Leading telecommunications company offering mobile and broadband services.",
			},
			&Company{
				Name:        "Emtel Ltd",
				Description: "Mobile network operator providing internet and 4G/5G services.",
			},
			&Company{
				Name:        "Gamma Civic Ltd",
				Description: "Facilities, construction, and maintenance services provider in Mauritius.",
			},
		)

	case "MDV": // Maldives
		companies = append(companies,
			&Company{
				Name:        "State Electric Company Limited (STELCO)",
				Description: "Provides electricity generation and distribution services across the Maldives.",
			},
			&Company{
				Name:        "Male' Water and Sewerage Company (MWSC)",
				Description: "Responsible for water supply and wastewater treatment in the Maldives.",
			},
			&Company{
				Name:        "Ooredoo Maldives",
				Description: "Telecommunications company offering mobile, internet, and data services.",
			},
			&Company{
				Name:        "Dhiraagu Plc",
				Description: "Maldives’ leading telecom and broadband service provider.",
			},
			&Company{
				Name:        "Urban Development and Construction Pvt Ltd",
				Description: "Company providing facilities management and maintenance services.",
			},
		)

	case "NOR": // Norway
		companies = append(companies,
			&Company{
				Name:        "Statkraft AS",
				Description: "State-owned hydropower company and leading electricity provider in Norway.",
			},
			&Company{
				Name:        "Hafslund Eco AS",
				Description: "Renewable energy company providing electricity and district heating in Oslo region.",
			},
			&Company{
				Name:        "Oslo Vann og Avløpsetaten",
				Description: "Municipal agency responsible for water supply and wastewater treatment in Oslo.",
			},
			&Company{
				Name:        "Telenor ASA",
				Description: "Norway’s largest telecommunications and broadband provider.",
			},
			&Company{
				Name:        "Altibox AS",
				Description: "Internet and TV service provider offering fiber broadband solutions.",
			},
			&Company{
				Name:        "ISS Facility Services Norway",
				Description: "Facilities management company providing cleaning and maintenance services nationwide.",
			},
		)

	case "DNK": // Denmark
		companies = append(companies,
			&Company{
				Name:        "Ørsted A/S",
				Description: "Global energy company based in Denmark, focusing on renewable electricity and heating.",
			},
			&Company{
				Name:        "HOFOR A/S",
				Description: "Greater Copenhagen’s utility company providing water, district heating, and waste management.",
			},
			&Company{
				Name:        "Evida A/S",
				Description: "National natural gas distribution operator in Denmark.",
			},
			&Company{
				Name:        "TDC Net A/S",
				Description: "Denmark’s primary telecom infrastructure and internet provider.",
			},
			&Company{
				Name:        "YouSee",
				Description: "Telecommunications and broadband provider serving households across Denmark.",
			},
			&Company{
				Name:        "ISS Danmark A/S",
				Description: "Facility management and cleaning services provider.",
			},
		)

	case "POL": // Poland
		companies = append(companies,
			&Company{
				Name:        "PGE Polska Grupa Energetyczna",
				Description: "Poland’s largest energy company providing electricity generation and distribution.",
			},
			&Company{
				Name:        "MPWiK Warszawa",
				Description: "Municipal Water and Sewerage Company serving Warsaw and nearby regions.",
			},
			&Company{
				Name:        "PGNiG (Polskie Górnictwo Naftowe i Gazownictwo)",
				Description: "National natural gas company handling supply and distribution.",
			},
			&Company{
				Name:        "Orange Polska",
				Description: "Leading telecommunications and internet provider.",
			},
			&Company{
				Name:        "Play (P4 Sp. z o.o.)",
				Description: "Telecom operator offering mobile and broadband services.",
			},
			&Company{
				Name:        "Veolia Energia Polska",
				Description: "Provides district heating, energy efficiency, and facility management services.",
			},
		)

	case "CZE": // Czech Republic
		companies = append(companies,
			&Company{
				Name:        "ČEZ Group",
				Description: "Leading energy company generating and distributing electricity and heat in the Czech Republic.",
			},
			&Company{
				Name:        "Pražská plynárenská",
				Description: "Major gas supplier providing natural gas distribution and energy services.",
			},
			&Company{
				Name:        "Pražská vodohospodářská společnost (PVS)",
				Description: "Company managing Prague’s water supply and sewage systems.",
			},
			&Company{
				Name:        "O2 Czech Republic",
				Description: "Telecommunication company providing mobile, internet, and broadband services.",
			},
			&Company{
				Name:        "Veolia Česká republika",
				Description: "Facilities management and environmental services company handling water, waste, and energy.",
			},
		)

	case "HUN": // Hungary
		companies = append(companies,
			&Company{
				Name:        "MVM Group",
				Description: "Hungary’s main state-owned energy company supplying electricity and gas.",
			},
			&Company{
				Name:        "Budapesti Elektromos Művek (ELMŰ)",
				Description: "Electric utility company serving Budapest and surrounding regions.",
			},
			&Company{
				Name:        "Budapest Waterworks (Fővárosi Vízművek)",
				Description: "Provides potable water and wastewater services in Budapest.",
			},
			&Company{
				Name:        "Magyar Telekom",
				Description: "Hungary’s leading telecommunications provider offering internet, mobile, and TV services.",
			},
			&Company{
				Name:        "B+N Referencia Zrt.",
				Description: "Facilities management and cleaning services company operating nationwide.",
			},
		)

	case "RUS": // Russia
		companies = append(companies,
			&Company{
				Name:        "Gazprom",
				Description: "Global energy giant supplying natural gas, electricity, and heat.",
			},
			&Company{
				Name:        "Rosseti",
				Description: "Russian power grid operator managing electricity transmission and distribution.",
			},
			&Company{
				Name:        "Mosvodokanal",
				Description: "Major water supply and wastewater management company in Moscow.",
			},
			&Company{
				Name:        "Rostelecom",
				Description: "National telecommunications provider offering internet, mobile, and digital services.",
			},
			&Company{
				Name:        "ISS Facility Services Russia",
				Description: "Facilities management company providing cleaning, maintenance, and energy services.",
			},
		)

	case "HRV": // Croatia
		companies = append(companies,
			&Company{
				Name:        "HEP Group (Hrvatska elektroprivreda)",
				Description: "National energy company providing electricity generation, distribution, and gas supply.",
			},
			&Company{
				Name:        "Hrvatske vode",
				Description: "Manages water resources, flood protection, and water supply infrastructure across Croatia.",
			},
			&Company{
				Name:        "Hrvatski Telekom",
				Description: "Leading telecommunications and internet service provider in Croatia.",
			},
			&Company{
				Name:        "Zagrebački Holding d.o.o.",
				Description: "Municipal services company managing waste, water, transport, and maintenance services in Zagreb.",
			},
		)

	case "BRA": // Brazil
		companies = append(companies,
			&Company{
				Name:        "Eletrobras",
				Description: "Brazil’s largest power utility company responsible for electricity generation and transmission.",
			},
			&Company{
				Name:        "Sabesp",
				Description: "Provides water supply and sewage treatment in São Paulo and nearby areas.",
			},
			&Company{
				Name:        "Petrobras",
				Description: "Integrated energy company dealing in oil, gas, and fuel distribution.",
			},
			&Company{
				Name:        "Vivo (Telefônica Brasil)",
				Description: "Telecom company providing mobile, internet, and TV services nationwide.",
			},
			&Company{
				Name:        "Grupo Enel Brasil",
				Description: "Operates electricity distribution, generation, and maintenance services across multiple states.",
			},
		)

	case "MEX": // Mexico
		companies = append(companies,
			&Company{
				Name:        "Comisión Federal de Electricidad (CFE)",
				Description: "Government-owned utility responsible for electricity generation and distribution.",
			},
			&Company{
				Name:        "Pemex",
				Description: "National oil and gas company supplying fuel and energy products.",
			},
			&Company{
				Name:        "Agua de México",
				Description: "Water supply and wastewater services provider across multiple municipalities.",
			},
			&Company{
				Name:        "Telmex",
				Description: "Mexico’s largest internet and telecommunications service provider.",
			},
			&Company{
				Name:        "ISS Facility Services México",
				Description: "Facilities management firm offering maintenance, cleaning, and building services.",
			},
		)

	case "ARG": // Argentina
		companies = append(companies,
			&Company{
				Name:        "Edesur",
				Description: "Electricity distribution company serving Buenos Aires and surrounding areas.",
			},
			&Company{
				Name:        "AySA (Agua y Saneamientos Argentinos)",
				Description: "Public company providing water and sanitation services in Greater Buenos Aires.",
			},
			&Company{
				Name:        "Metrogas",
				Description: "Main natural gas distributor in Buenos Aires.",
			},
			&Company{
				Name:        "Telecom Argentina",
				Description: "Telecommunications provider offering mobile, internet, and cable services.",
			},
			&Company{
				Name:        "Grupo Roggio",
				Description: "Infrastructure and services firm providing maintenance, transport, and utilities management.",
			},
		)

	case "CHL": // Chile
		companies = append(companies,
			&Company{
				Name:        "Enel Distribución Chile",
				Description: "Major electricity distribution company serving Santiago and nearby regions.",
			},
			&Company{
				Name:        "Aguas Andinas",
				Description: "Water supply and wastewater treatment company for the Santiago metropolitan area.",
			},
			&Company{
				Name:        "Metrogas Chile",
				Description: "Natural gas distribution company for residential and commercial clients.",
			},
			&Company{
				Name:        "Movistar Chile",
				Description: "Telecommunications company offering internet, mobile, and TV services.",
			},
			&Company{
				Name:        "Sodexo Chile",
				Description: "Facilities management and maintenance services provider.",
			},
		)

	case "COL": // Colombia
		companies = append(companies,
			&Company{
				Name:        "Grupo Energía Bogotá",
				Description: "Energy company managing electricity and natural gas infrastructure in Colombia and Latin America.",
			},
			&Company{
				Name:        "Empresas Públicas de Medellín (EPM)",
				Description: "Provides water, electricity, gas, and waste management services across Colombia.",
			},
			&Company{
				Name:        "Claro Colombia",
				Description: "Telecommunications provider offering broadband, mobile, and TV services.",
			},
			&Company{
				Name:        "Veolia Colombia",
				Description: "Environmental and facilities management services company.",
			},
		)

	case "PER": // Peru
		companies = append(companies,
			&Company{
				Name:        "Enel Distribución Perú",
				Description: "Electricity distribution and energy services provider for Lima and surrounding regions.",
			},
			&Company{
				Name:        "Sedapal",
				Description: "Public water and sanitation company serving Lima and Callao.",
			},
			&Company{
				Name:        "Petroperú",
				Description: "State-owned oil and gas company managing fuel distribution and refining.",
			},
			&Company{
				Name:        "Movistar Perú",
				Description: "Leading telecommunications provider offering mobile, internet, and TV services.",
			},
			&Company{
				Name:        "Graña y Montero (AENZA)",
				Description: "Engineering and facilities management company providing maintenance and infrastructure services.",
			},
		)

	case "URY": // Uruguay
		companies = append(companies,
			&Company{
				Name:        "UTE (Administración Nacional de Usinas y Trasmisiones Eléctricas)",
				Description: "State-owned utility responsible for electricity generation and distribution across Uruguay.",
			},
			&Company{
				Name:        "OSE (Obras Sanitarias del Estado)",
				Description: "National water and sanitation company providing potable water and wastewater treatment.",
			},
			&Company{
				Name:        "ANTEL",
				Description: "Government-owned telecommunications provider offering internet, phone, and mobile services.",
			},
			&Company{
				Name:        "MontevideoGas",
				Description: "Natural gas distributor for residential and industrial clients in Montevideo.",
			},
			&Company{
				Name:        "ISS Uruguay",
				Description: "Facilities management and maintenance services company operating nationwide.",
			},
		)

	case "DOM": // Dominican Republic
		companies = append(companies,
			&Company{
				Name:        "Edesur Dominicana",
				Description: "Electricity distribution company serving the southern region of the Dominican Republic.",
			},
			&Company{
				Name:        "CAASD (Corporación del Acueducto y Alcantarillado de Santo Domingo)",
				Description: "Public company managing water supply and wastewater services in Santo Domingo.",
			},
			&Company{
				Name:        "Claro Dominicana",
				Description: "Leading telecommunications provider offering internet, mobile, and cable TV services.",
			},
			&Company{
				Name:        "Propagas",
				Description: "Major distributor of liquefied petroleum gas for homes and businesses.",
			},
			&Company{
				Name:        "Grupo SID Facilities",
				Description: "Integrated maintenance and facility management services provider.",
			},
		)

	case "PRY": // Paraguay
		companies = append(companies,
			&Company{
				Name:        "ANDE (Administración Nacional de Electricidad)",
				Description: "State-owned company responsible for electricity generation and distribution in Paraguay.",
			},
			&Company{
				Name:        "ESSAP S.A.",
				Description: "National water and sanitation service provider.",
			},
			&Company{
				Name:        "Tigo Paraguay",
				Description: "Telecom company offering internet, mobile, and cable TV services.",
			},
			&Company{
				Name:        "Copetrol",
				Description: "Energy distributor providing fuel, gas, and energy-related services.",
			},
			&Company{
				Name:        "Sodexo Paraguay",
				Description: "Facilities and maintenance services company for business and industrial sectors.",
			},
		)

	case "BOL": // Bolivia
		companies = append(companies,
			&Company{
				Name:        "ENDE (Empresa Nacional de Electricidad)",
				Description: "Bolivia’s national electricity company managing generation, transmission, and distribution.",
			},
			&Company{
				Name:        "EPSAS",
				Description: "Water and sanitation provider serving La Paz and El Alto.",
			},
			&Company{
				Name:        "YPFB (Yacimientos Petrolíferos Fiscales Bolivianos)",
				Description: "State-owned oil and gas company managing energy production and distribution.",
			},
			&Company{
				Name:        "VIVA Bolivia",
				Description: "Telecommunications company providing internet, mobile, and digital services.",
			},
			&Company{
				Name:        "AESA Bolivia",
				Description: "Facilities management company offering cleaning, maintenance, and infrastructure support.",
			},
		)

	case "VEN": // Venezuela
		companies = append(companies,
			&Company{
				Name:        "CORPOELEC",
				Description: "National electric power corporation overseeing generation and distribution across Venezuela.",
			},
			&Company{
				Name:        "HIDROCAPITAL",
				Description: "Water utility responsible for supplying water and wastewater treatment in Caracas and nearby areas.",
			},
			&Company{
				Name:        "CANTV",
				Description: "National telecommunications provider offering fixed line, internet, and mobile services.",
			},
			&Company{
				Name:        "PDVSA Gas",
				Description: "Subsidiary of PDVSA managing natural gas production and distribution.",
			},
			&Company{
				Name:        "SENIAT Facilities",
				Description: "Public services and maintenance provider for infrastructure and office spaces.",
			},
		)

	case "PAK": // Pakistan
		companies = append(companies,
			&Company{
				Name:        "K-Electric",
				Description: "Private utility company responsible for electricity generation and distribution in Karachi.",
			},
			&Company{
				Name:        "WAPDA (Water and Power Development Authority)",
				Description: "Government agency managing water and electricity projects across Pakistan.",
			},
			&Company{
				Name:        "SNGPL (Sui Northern Gas Pipelines Limited)",
				Description: "Major natural gas supplier in northern Pakistan.",
			},
			&Company{
				Name:        "PTCL (Pakistan Telecommunication Company Limited)",
				Description: "National telecom company providing internet, broadband, and mobile services.",
			},
			&Company{
				Name:        "Servest Pakistan",
				Description: "Facilities management and maintenance services company operating in major cities.",
			},
		)

	case "BGD": // Bangladesh
		companies = append(companies,
			&Company{
				Name:        "Dhaka Electric Supply Company (DESCO)",
				Description: "Electric distribution company serving Dhaka and nearby regions.",
			},
			&Company{
				Name:        "WASA (Dhaka Water Supply and Sewerage Authority)",
				Description: "Public utility managing water and sanitation in the capital region.",
			},
			&Company{
				Name:        "Titas Gas Transmission and Distribution Company Limited",
				Description: "Largest gas distribution company in Bangladesh.",
			},
			&Company{
				Name:        "Grameenphone",
				Description: "Leading telecom operator offering mobile and internet services.",
			},
			&Company{
				Name:        "Sodexo Bangladesh",
				Description: "Facilities management and maintenance services provider for business and institutions.",
			},
		)

	case "LKA": // Sri Lanka
		companies = append(companies,
			&Company{
				Name:        "Ceylon Electricity Board (CEB)",
				Description: "Government-owned corporation responsible for electricity generation and distribution.",
			},
			&Company{
				Name:        "National Water Supply and Drainage Board (NWSDB)",
				Description: "Provides water supply and sanitation services throughout Sri Lanka.",
			},
			&Company{
				Name:        "Lanka IOC",
				Description: "Energy company involved in fuel distribution and oil services.",
			},
			&Company{
				Name:        "Dialog Axiata",
				Description: "Leading telecommunications provider offering internet, mobile, and digital TV.",
			},
			&Company{
				Name:        "Jones Lang LaSalle Sri Lanka",
				Description: "Facilities and property management services company.",
			},
		)

	case "NPL": // Nepal
		companies = append(companies,
			&Company{
				Name:        "Nepal Electricity Authority (NEA)",
				Description: "State-owned utility managing electricity generation and distribution across Nepal.",
			},
			&Company{
				Name:        "Kathmandu Upatyaka Khanepani Limited (KUKL)",
				Description: "Water supply and sanitation provider for the Kathmandu Valley.",
			},
			&Company{
				Name:        "Nepal Oil Corporation",
				Description: "Government-owned company distributing fuel and petroleum products.",
			},
			&Company{
				Name:        "Ncell Axiata",
				Description: "Telecom company providing mobile and internet services nationwide.",
			},
			&Company{
				Name:        "Nepal Facilities Management Services",
				Description: "Company offering building maintenance and support services.",
			},
		)

	case "MMR": // Myanmar
		companies = append(companies,
			&Company{
				Name:        "Yangon Electricity Supply Corporation (YESC)",
				Description: "Provides electricity distribution and billing services in Yangon.",
			},
			&Company{
				Name:        "Yangon City Development Committee (YCDC)",
				Description: "Municipal body managing water, sanitation, and waste services.",
			},
			&Company{
				Name:        "Myanma Oil and Gas Enterprise (MOGE)",
				Description: "Government enterprise managing oil and gas exploration and supply.",
			},
			&Company{
				Name:        "MPT (Myanmar Posts and Telecommunications)",
				Description: "Main telecom operator offering internet, mobile, and broadband services.",
			},
			&Company{
				Name:        "CBM Facilities Services",
				Description: "Private company providing maintenance, janitorial, and technical support services.",
			},
		)

	case "KHM": // Cambodia
		companies = append(companies,
			&Company{
				Name:        "Electricité du Cambodge (EDC)",
				Description: "State-owned company responsible for electricity generation, transmission, and distribution across Cambodia.",
			},
			&Company{
				Name:        "Phnom Penh Water Supply Authority (PPWSA)",
				Description: "Public utility providing clean water supply and sanitation services in Phnom Penh.",
			},
			&Company{
				Name:        "TotalEnergies Cambodia",
				Description: "Energy company supplying fuel, lubricants, and gas products.",
			},
			&Company{
				Name:        "Metfone",
				Description: "Major telecommunications operator offering internet, mobile, and broadband services.",
			},
			&Company{
				Name:        "Sodexo Cambodia",
				Description: "Facilities management and maintenance services provider for commercial and industrial clients.",
			},
		)

	case "LAO": // Laos
		companies = append(companies,
			&Company{
				Name:        "Électricité du Laos (EDL)",
				Description: "National power utility responsible for electricity generation and distribution in Laos.",
			},
			&Company{
				Name:        "Vientiane Water Supply State Enterprise",
				Description: "Provides water supply and sanitation services in Vientiane and nearby regions.",
			},
			&Company{
				Name:        "PetroTrade Lao Public Company",
				Description: "Major distributor of fuel and gas across Laos.",
			},
			&Company{
				Name:        "Lao Telecom",
				Description: "Leading telecommunications operator providing internet, mobile, and enterprise network services.",
			},
			&Company{
				Name:        "EDL-Gen Services",
				Description: "Subsidiary offering technical maintenance and energy infrastructure support.",
			},
		)

	case "NGA": // Nigeria
		companies = append(companies,
			&Company{
				Name:        "Ikeja Electric",
				Description: "Nigeria’s largest electricity distribution company serving the Lagos area.",
			},
			&Company{
				Name:        "Lagos Water Corporation",
				Description: "Public water supply agency providing treated water to Lagos and nearby areas.",
			},
			&Company{
				Name:        "TotalEnergies Nigeria",
				Description: "Oil and gas company supplying fuel, lubricants, and LPG services nationwide.",
			},
			&Company{
				Name:        "MTN Nigeria",
				Description: "Largest telecom operator offering mobile, data, and broadband services.",
			},
			&Company{
				Name:        "Alpha Mead Facilities",
				Description: "Integrated facilities management and maintenance services provider.",
			},
		)

	case "KEN": // Kenya
		companies = append(companies,
			&Company{
				Name:        "Kenya Power and Lighting Company (KPLC)",
				Description: "State-owned company managing electricity distribution and billing in Kenya.",
			},
			&Company{
				Name:        "Nairobi City Water and Sewerage Company",
				Description: "Provides water supply and wastewater services in the capital region.",
			},
			&Company{
				Name:        "Kenya Pipeline Company (KPC)",
				Description: "Handles transportation and storage of petroleum products across Kenya.",
			},
			&Company{
				Name:        "Safaricom PLC",
				Description: "Telecom giant offering internet, mobile, and M-Pesa financial services.",
			},
			&Company{
				Name:        "Sodexo Kenya",
				Description: "Facilities management and maintenance company serving corporate and industrial sectors.",
			},
		)

	case "GHA": // Ghana
		companies = append(companies,
			&Company{
				Name:        "Electricity Company of Ghana (ECG)",
				Description: "Main utility responsible for electricity distribution and customer billing.",
			},
			&Company{
				Name:        "Ghana Water Company Limited (GWCL)",
				Description: "Public utility managing water supply and sanitation services across Ghana.",
			},
			&Company{
				Name:        "Ghana National Gas Company",
				Description: "State-owned company managing gas processing and distribution infrastructure.",
			},
			&Company{
				Name:        "MTN Ghana",
				Description: "Leading telecom provider offering internet, mobile, and digital services.",
			},
			&Company{
				Name:        "Broll Ghana",
				Description: "Facilities and property management company providing cleaning and maintenance services.",
			},
		)

	case "MAR": // Morocco
		companies = append(companies,
			&Company{
				Name:        "ONEE (Office National de l'Électricité et de l'Eau Potable)",
				Description: "National company providing electricity and potable water across Morocco.",
			},
			&Company{
				Name:        "Lydec",
				Description: "Private company managing water, electricity, and sanitation services in Casablanca.",
			},
			&Company{
				Name:        "Afriquia Gaz",
				Description: "Leading gas distributor providing LPG and energy solutions.",
			},
			&Company{
				Name:        "Maroc Telecom",
				Description: "Major telecommunications company offering mobile, internet, and landline services.",
			},
			&Company{
				Name:        "Derichebourg Maroc",
				Description: "Facilities management and industrial maintenance company operating nationwide.",
			},
		)

	case "TUN": // Tunisia
		companies = append(companies,
			&Company{
				Name:        "STEG (Société Tunisienne de l'Électricité et du Gaz)",
				Description: "Public utility providing electricity and natural gas services across Tunisia.",
			},
			&Company{
				Name:        "SONEDE (Société Nationale d’Exploitation et de Distribution des Eaux)",
				Description: "National water company managing supply and sanitation.",
			},
			&Company{
				Name:        "Tunisie Telecom",
				Description: "Leading telecom operator offering internet, mobile, and broadband services.",
			},
			&Company{
				Name:        "Shell Tunisia",
				Description: "Energy company engaged in gas and petroleum distribution.",
			},
			&Company{
				Name:        "ISS Tunisie",
				Description: "Facilities and workplace management company providing cleaning and maintenance solutions.",
			},
		)

	case "ETH": // Ethiopia
		companies = append(companies,
			&Company{
				Name:        "Ethiopian Electric Utility (EEU)",
				Description: "Government utility company responsible for electricity distribution and customer services.",
			},
			&Company{
				Name:        "Addis Ababa Water and Sewerage Authority (AAWSA)",
				Description: "Provides water and wastewater services in the capital city.",
			},
			&Company{
				Name:        "Ethiopian Petroleum Supply Enterprise",
				Description: "State-owned distributor of petroleum and gas products.",
			},
			&Company{
				Name:        "Ethio Telecom",
				Description: "National telecommunications provider offering internet, mobile, and ICT solutions.",
			},
			&Company{
				Name:        "Sodexo Ethiopia",
				Description: "Facilities management and building maintenance services company.",
			},
		)

	case "DZA": // Algeria
		companies = append(companies,
			&Company{
				Name:        "Sonelgaz",
				Description: "State-owned company overseeing electricity and gas distribution throughout Algeria.",
			},
			&Company{
				Name:        "SEAAL (Société des Eaux et de l’Assainissement d’Alger)",
				Description: "Water and sanitation company serving Algiers and nearby regions.",
			},
			&Company{
				Name:        "Naftal",
				Description: "National petroleum and gas distributor providing LPG and fuel products.",
			},
			&Company{
				Name:        "Algérie Télécom",
				Description: "Telecom operator offering internet, broadband, and telephony services.",
			},
			&Company{
				Name:        "ENGIE Services Algérie",
				Description: "Facilities and energy management company offering maintenance and engineering services.",
			},
		)

	case "UKR": // Ukraine
		companies = append(companies,
			&Company{
				Name:        "DTEK",
				Description: "Largest private energy company generating and distributing electricity and gas across Ukraine.",
			},
			&Company{
				Name:        "Kyivvodokanal",
				Description: "Municipal water supply and wastewater treatment company serving Kyiv.",
			},
			&Company{
				Name:        "Naftogaz of Ukraine",
				Description: "National oil and gas company managing exploration, transport, and supply.",
			},
			&Company{
				Name:        "Kyivstar",
				Description: "Leading telecom company providing mobile, broadband, and digital services.",
			},
			&Company{
				Name:        "ISS Ukraine",
				Description: "Facilities management and maintenance service provider operating nationwide.",
			},
		)

	case "ROU": // Romania
		companies = append(companies,
			&Company{
				Name:        "Electrica SA",
				Description: "Major electricity distribution company serving multiple regions across Romania.",
			},
			&Company{
				Name:        "Engie Romania",
				Description: "Gas and energy supplier providing natural gas, electricity, and energy services nationwide.",
			},
			&Company{
				Name:        "Apa Nova București",
				Description: "Water and sewage utility company serving Bucharest and surrounding areas.",
			},
			&Company{
				Name:        "Digi Romania (RCS & RDS)",
				Description: "Leading internet, mobile, and cable TV service provider in Romania.",
			},
			&Company{
				Name:        "Compania Națională de Administrare a Infrastructurii Rutiere (CNAIR)",
				Description: "Responsible for road maintenance and infrastructure services.",
			},
		)

	case "BGR": // Bulgaria
		companies = append(companies,
			&Company{
				Name:        "CEZ Bulgaria",
				Description: "Electricity distribution and supply company operating mainly in western Bulgaria.",
			},
			&Company{
				Name:        "Overgas Inc.",
				Description: "Leading private natural gas supplier and distributor in Bulgaria.",
			},
			&Company{
				Name:        "Sofiyska Voda AD",
				Description: "Water and wastewater utility for Sofia.",
			},
			&Company{
				Name:        "Vivacom",
				Description: "Telecommunications provider offering broadband internet, mobile, and TV services.",
			},
			&Company{
				Name:        "TITAN Zlatna Panega",
				Description: "Company offering waste management and industrial maintenance services.",
			},
		)

	case "SRB": // Serbia
		companies = append(companies,
			&Company{
				Name:        "EPS (Elektroprivreda Srbije)",
				Description: "State-owned electricity generation and distribution company.",
			},
			&Company{
				Name:        "Srbijagas",
				Description: "Main natural gas distributor and supplier across Serbia.",
			},
			&Company{
				Name:        "JKP Beogradski Vodovod i Kanalizacija",
				Description: "Public utility managing water supply and sewage in Belgrade.",
			},
			&Company{
				Name:        "Telekom Srbija (MTS)",
				Description: "Leading telecommunications and internet provider in Serbia.",
			},
			&Company{
				Name:        "City Service Belgrade",
				Description: "Facilities and property maintenance management company.",
			},
		)

	case "ISL": // Iceland
		companies = append(companies,
			&Company{
				Name:        "Landsvirkjun",
				Description: "National power company generating renewable electricity from hydro, geothermal, and wind.",
			},
			&Company{
				Name:        "Veitur Utilities",
				Description: "Provides electricity, hot water, and cold water services in Reykjavík and surrounding areas.",
			},
			&Company{
				Name:        "Orkuveita Reykjavíkur (Reykjavik Energy)",
				Description: "Utility company providing geothermal heating, electricity, and water services.",
			},
			&Company{
				Name:        "Síminn",
				Description: "Iceland’s main telecom and internet provider.",
			},
			&Company{
				Name:        "Íslandsbanki Facility Services",
				Description: "Building and facility maintenance services across Iceland.",
			},
		)

	case "BLR": // Belarus
		companies = append(companies,
			&Company{
				Name:        "Belenergo",
				Description: "National electric power company managing generation and distribution across Belarus.",
			},
			&Company{
				Name:        "Beltopgaz",
				Description: "Main natural gas provider in Belarus.",
			},
			&Company{
				Name:        "Minskvodokanal",
				Description: "Municipal water supply and sewage company for Minsk.",
			},
			&Company{
				Name:        "A1 Telekom Belarus",
				Description: "Leading mobile and broadband internet provider in Belarus.",
			},
			&Company{
				Name:        "Belkommunservice",
				Description: "Municipal maintenance and waste management company.",
			},
		)

	case "FJI": // Fiji
		companies = append(companies,
			&Company{
				Name:        "Energy Fiji Limited (EFL)",
				Description: "Government-owned electricity generation and distribution company.",
			},
			&Company{
				Name:        "Water Authority of Fiji",
				Description: "National provider of water and wastewater services.",
			},
			&Company{
				Name:        "Fiji Gas",
				Description: "Main liquefied petroleum gas (LPG) supplier for residential and commercial use.",
			},
			&Company{
				Name:        "Vodafone Fiji",
				Description: "Leading telecommunications and broadband service provider.",
			},
			&Company{
				Name:        "Rentokil Initial Fiji",
				Description: "Provides cleaning, hygiene, and maintenance services across Fiji.",
			},
		)

	case "PNG": // Papua New Guinea
		companies = append(companies,
			&Company{
				Name:        "PNG Power Limited",
				Description: "State-owned utility responsible for electricity generation, transmission, and distribution.",
			},
			&Company{
				Name:        "Water PNG Limited",
				Description: "Public water supply and sanitation company.",
			},
			&Company{
				Name:        "Digicel PNG",
				Description: "Leading telecom provider offering internet and mobile services.",
			},
			&Company{
				Name:        "Ela Motors Maintenance Services",
				Description: "Facilities and maintenance services provider across Papua New Guinea.",
			},
		)

	case "JAM": // Jamaica
		companies = append(companies,
			&Company{
				Name:        "Jamaica Public Service Company (JPS)",
				Description: "Main electricity generation and distribution company in Jamaica.",
			},
			&Company{
				Name:        "National Water Commission (NWC)",
				Description: "Public utility providing potable water and wastewater services.",
			},
			&Company{
				Name:        "GasPro Jamaica",
				Description: "Distributor of liquefied petroleum gas and energy solutions.",
			},
			&Company{
				Name:        "FLOW Jamaica",
				Description: "Major telecommunications and broadband provider.",
			},
			&Company{
				Name:        "KLEANMIX Jamaica",
				Description: "Building and facility maintenance service provider.",
			},
		)

	case "CRI": // Costa Rica
		companies = append(companies,
			&Company{
				Name:        "Instituto Costarricense de Electricidad (ICE)",
				Description: "National provider of electricity and telecommunications.",
			},
			&Company{
				Name:        "A y A (Acueductos y Alcantarillados)",
				Description: "Public utility managing water supply and sewer systems.",
			},
			&Company{
				Name:        "Gas Tomza Costa Rica",
				Description: "Major distributor of LPG gas in Costa Rica.",
			},
			&Company{
				Name:        "Liberty Costa Rica",
				Description: "Telecom company offering internet, cable, and mobile services.",
			},
			&Company{
				Name:        "EULEN Costa Rica",
				Description: "Facilities management and cleaning services provider.",
			},
		)

	case "GTM": // Guatemala
		companies = append(companies,
			&Company{
				Name:        "Energuate",
				Description: "Electricity distribution company serving most of Guatemala.",
			},
			&Company{
				Name:        "EEGSA (Empresa Eléctrica de Guatemala)",
				Description: "Electric utility serving the capital and nearby regions.",
			},
			&Company{
				Name:        "EMPAGUA",
				Description: "Public water supply and sanitation company in Guatemala City.",
			},
			&Company{
				Name:        "Claro Guatemala",
				Description: "Leading telecommunications provider offering internet and mobile services.",
			},
			&Company{
				Name:        "Multiservicios GT",
				Description: "Maintenance and cleaning services provider for businesses and households.",
			},
		)
	case "KWT": // Kuwait
	case "QAT": // Qatar
	case "OMN": // Oman
	case "BHR": // Bahrain
	case "JOR": // Jordan
	case "KAZ": // Kazakhstan
	}

	for _, data := range companies {
		data.CreatedAt = now
		data.UpdatedAt = now
		data.CreatedByID = userID
		data.UpdatedByID = userID
		data.OrganizationID = organizationID
		data.BranchID = branchID
		if err := m.CompanyManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed company %s", data.Name)
		}
	}
	return nil
}

func (m *Core) CompanyCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*Company, error) {
	return m.CompanyManager.Find(context, &Company{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
