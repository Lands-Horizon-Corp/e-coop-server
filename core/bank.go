package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
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

func (m *Core) BankManager() *registry.Registry[Bank, BankResponse, BankRequest] {
	return registry.GetRegistry(registry.RegistryParams[Bank, BankResponse, BankRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media"},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Bank) *BankResponse {
			if data == nil {
				return nil
			}
			return &BankResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),
				MediaID:        data.MediaID,
				Media:          m.MediaManager().ToModel(data.Media),
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *Bank) registry.Topics {
			return []string{
				"bank.create",
				fmt.Sprintf("bank.create.%s", data.ID),
				fmt.Sprintf("bank.create.branch.%s", data.BranchID),
				fmt.Sprintf("bank.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Bank) registry.Topics {
			return []string{
				"bank.update",
				fmt.Sprintf("bank.update.%s", data.ID),
				fmt.Sprintf("bank.update.branch.%s", data.BranchID),
				fmt.Sprintf("bank.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Bank) registry.Topics {
			return []string{
				"bank.delete",
				fmt.Sprintf("bank.delete.%s", data.ID),
				fmt.Sprintf("bank.delete.branch.%s", data.BranchID),
				fmt.Sprintf("bank.delete.organization.%s", data.OrganizationID),
			}
		},
	})

}

func (m *Core) bankSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	banks := []*Bank{
		{Name: "HSBC", Description: "One of the world’s largest multinational banks, serving customers globally with retail, commercial, and investment banking."},
		{Name: "Citibank", Description: "A major global bank headquartered in the United States, known for strong international consumer and corporate banking services."},
		{Name: "JPMorgan Chase", Description: "The largest bank in the United States offering worldwide financial services including investment, retail, and commercial banking."},
		{Name: "Bank of America", Description: "A leading US-based multinational bank providing global banking, investing, and financial risk management services."},
		{Name: "Wells Fargo", Description: "A major American financial services company offering banking, investment, and mortgage products with international reach."},
		{Name: "Standard Chartered", Description: "A British multinational bank operating across Asia, Africa, Europe, and the Middle East with strong global trade presence."},
		{Name: "Barclays", Description: "A British universal bank with global operations including retail, corporate, and investment banking."},
		{Name: "Deutsche Bank", Description: "Germany’s largest bank providing global investment banking, corporate solutions, and financial services."},
		{Name: "BNP Paribas", Description: "A major French international banking group offering retail and corporate financial services worldwide."},
		{Name: "UBS", Description: "A Swiss multinational investment bank known globally for wealth management and financial advisory."},
		{Name: "Credit Suisse", Description: "A Swiss global bank recognized for investment banking, wealth management, and finance services."},
		{Name: "Santander", Description: "A Spanish multinational bank offering retail and commercial banking services across Europe and the Americas."},
		{Name: "ING", Description: "A Dutch multinational banking group focused on retail, direct banking, and international financial services."},
		{Name: "Scotiabank", Description: "A Canadian global bank with strong international banking operations across Latin America and other regions."},
		{Name: "Royal Bank of Canada (RBC)", Description: "Canada’s largest bank, offering extensive global banking, wealth management, and investment services."},
		{Name: "Credit Agricole", Description: "A leading French international banking group specializing in retail and corporate banking."},
		{Name: "Mizuho Bank", Description: "A major Japanese multinational bank providing global corporate and investment banking services."},
		{Name: "MUFG Bank", Description: "Japan’s largest bank with a strong international footprint in corporate and retail banking."},
		{Name: "Sumitomo Mitsui Banking Corporation (SMBC)", Description: "A top Japanese global financial institution offering corporate and investment banking worldwide."},

		{Name: "Revolut", Description: "A global financial super-app offering international banking, money transfers, cards, and digital finance services."},
		{Name: "Wise", Description: "A global fintech company specializing in low-cost international money transfers and digital accounts."},
		{Name: "PayPal", Description: "A widely used global online payment platform offering digital wallets and financial services."},
		{Name: "N26", Description: "A fully digital European bank providing international mobile banking services."},
		{Name: "Monzo", Description: "A UK-based digital bank known for seamless international banking through mobile-first features."},

		{Name: "Google Pay", Description: "A widely used digital wallet and online payment system available globally for online and in-store purchases."},
		{Name: "Apple Pay", Description: "A secure mobile payment and digital wallet service allowing payments via Apple devices worldwide."},
		{Name: "Samsung Pay", Description: "A global mobile payment platform allowing contactless payments and online transactions via Samsung devices."},
		{Name: "PayU", Description: "A global fintech company offering online payment solutions and e-wallet services in multiple countries."},
		{Name: "Stripe", Description: "A leading online payment processing platform supporting global e-commerce, subscriptions, and financial services."},
		{Name: "Alipay", Description: "China’s largest mobile and online payment platform widely used internationally for cross-border payments."},
		{Name: "WeChat Pay", Description: "A Chinese mobile payment and digital wallet service integrated into the WeChat app for global transactions."},
		{Name: "Venmo", Description: "A US-based mobile payment service allowing peer-to-peer transfers and online payments."},
		{Name: "Cash App", Description: "A US-based mobile wallet enabling peer-to-peer money transfers, Bitcoin transactions, and investing."},
	}

	branch, err := m.BranchManager().GetByID(context, branchID)
	if err != nil {
		return eris.Wrapf(err, "failed to get branch %s for bank seeding", branchID)
	}
	switch branch.Currency.ISO3166Alpha3 {
	case "USA": // United States
		banks = append(banks,
			&Bank{Name: "Chase Bank", Description: "One of the largest US banks offering retail, commercial, and investment banking services."},
			&Bank{Name: "Bank of America", Description: "A leading national bank providing consumer, business, and private banking services."},
			&Bank{Name: "Wells Fargo", Description: "A major US bank offering nationwide banking, loans, and financial services."},
			&Bank{Name: "CitiBank", Description: "A global US bank with strong international and domestic banking operations."},
			&Bank{Name: "U.S. Bank", Description: "A large national bank offering a full range of financial services."},
			&Bank{Name: "PNC Bank", Description: "A major US bank offering retail and corporate banking, widely used across the East Coast."},
			&Bank{Name: "Capital One", Description: "A popular US bank known for credit cards, auto loans, and online banking."},
			&Bank{Name: "Truist Bank", Description: "A major regional bank formed from BB&T and SunTrust, offering full financial services."},

			&Bank{Name: "Bank of New York Mellon", Description: "A global investments and asset management company."},
			&Bank{Name: "Goldman Sachs", Description: "A top US investment bank offering financial advisory and asset management services."},
			&Bank{Name: "Morgan Stanley", Description: "A leading global investment bank based in the United States."},

			&Bank{Name: "Fifth Third Bank", Description: "A well-known regional bank serving the Midwest and Southeast US."},
			&Bank{Name: "KeyBank", Description: "A regional bank offering personal and business banking across multiple US states."},
			&Bank{Name: "Regions Bank", Description: "A Southeastern regional bank providing retail and corporate banking services."},
			&Bank{Name: "Huntington Bank", Description: "A Midwest regional bank known for consumer and small business banking services."},
			&Bank{Name: "TD Bank USA", Description: "The US arm of TD Bank Group, offering retail banking across the East Coast."},

			&Bank{Name: "Navy Federal Credit Union", Description: "The largest credit union in the US serving military members and families."},
			&Bank{Name: "Alliant Credit Union", Description: "A large US digital credit union offering nationwide banking services."},
			&Bank{Name: "Pentagon Federal Credit Union (PenFed)", Description: "A major US credit union providing a variety of financial services."},

			&Bank{Name: "Ally Bank", Description: "A popular online-only bank in the US offering savings, checking, and loans."},
			&Bank{Name: "Discover Bank", Description: "An online bank known for its credit cards and high-yield deposit accounts."},
			&Bank{Name: "Axos Bank", Description: "A digital bank providing online checking, savings, and loan services."},
			&Bank{Name: "Chime", Description: "A leading US neobank offering fee-free digital banking through a mobile app."},

			&Bank{Name: "Varo Bank", Description: "A fully digital US bank with no-fee checking and savings accounts."},
			&Bank{Name: "SoFi Bank", Description: "A digital bank offering loans, investing, and mobile banking services."},
			&Bank{Name: "Current", Description: "A mobile-first US banking platform offering instant notifications and budgeting tools."},
			&Bank{Name: "Green Dot Bank", Description: "A branchless US bank providing prepaid cards and online financial services."},
		)
	case "DEU": // Germany (Euro representative)
		banks = append(banks,
			&Bank{Name: "Deutsche Bank", Description: "Germany’s largest bank with global investment and corporate banking operations."},
			&Bank{Name: "Commerzbank", Description: "A major German bank offering retail, corporate, and international banking services."},
			&Bank{Name: "BNP Paribas", Description: "A major French banking group providing global financial and retail banking services."},
			&Bank{Name: "Crédit Agricole", Description: "A leading French retail and commercial bank with strong EU presence."},
			&Bank{Name: "Société Générale", Description: "A major French bank offering corporate, investment, and retail banking."},

			&Bank{Name: "ING Group", Description: "A Dutch multinational bank offering retail and direct banking services across Europe."},
			&Bank{Name: "UniCredit", Description: "A leading Italian banking group with operations across Europe."},
			&Bank{Name: "Santander", Description: "A Spanish multinational bank providing retail and commercial banking across Europe."},
			&Bank{Name: "ABN AMRO", Description: "A Dutch bank known for retail, commercial, and international banking."},

			&Bank{Name: "Revolut", Description: "A global financial super app offering international banking and card services."},
			&Bank{Name: "N26", Description: "A German mobile-only bank offering modern digital banking services in the EU."},
			&Bank{Name: "Bunq", Description: "A Dutch fully digital bank providing innovative mobile banking services across Europe."},
		)

	case "JPN": // Japan
		banks = append(banks,
			&Bank{Name: "MUFG Bank", Description: "Japan’s largest bank offering global corporate, retail, and investment banking."},
			&Bank{Name: "Mizuho Bank", Description: "A leading Japanese bank providing financial services across Asia and worldwide."},
			&Bank{Name: "SMBC (Sumitomo Mitsui Banking Corporation)", Description: "A top Japanese bank known for corporate and investment banking."},

			&Bank{Name: "Japan Post Bank", Description: "One of Japan’s largest retail banks with nationwide branches."},
			&Bank{Name: "Resona Bank", Description: "A major regional banking group in Japan offering retail financial services."},
			&Bank{Name: "Shinsei Bank", Description: "A Japanese commercial bank providing retail and institutional financial services."},

			&Bank{Name: "Sony Bank", Description: "A Japanese online bank offering digital savings, loans, and FX services."},
			&Bank{Name: "Rakuten Bank", Description: "Japan’s largest online bank offering digital accounts and payment services."},
		)

	case "GBR": // United Kingdom
		banks = append(banks,
			&Bank{Name: "HSBC UK", Description: "One of the largest UK banks offering full retail and corporate banking services."},
			&Bank{Name: "Barclays", Description: "A major British multinational bank with strong retail and investment banking."},
			&Bank{Name: "Lloyds Bank", Description: "A leading UK retail and commercial bank with nationwide presence."},
			&Bank{Name: "NatWest", Description: "A major retail and commercial bank serving individuals and businesses across the UK."},
			&Bank{Name: "Standard Chartered", Description: "A British multinational bank with strong global operations."},

			&Bank{Name: "Royal Bank of Scotland (RBS)", Description: "A well-known UK bank offering retail and commercial banking."},
			&Bank{Name: "TSB Bank", Description: "A popular UK retail bank with community-focused banking services."},
			&Bank{Name: "Halifax", Description: "A major UK retail bank specializing in mortgages and savings."},

			&Bank{Name: "Revolut", Description: "A global fintech bank offering multi-currency accounts and digital banking."},
			&Bank{Name: "Monzo", Description: "A popular UK mobile-only bank offering modern digital financial services."},
			&Bank{Name: "Starling Bank", Description: "A fully digital UK bank offering personal, business, and joint accounts."},
			&Bank{Name: "Atom Bank", Description: "A UK-based online bank specializing in savings and mortgage products."},
		)

	case "AUS": // Australia
		banks = append(banks,
			&Bank{Name: "Commonwealth Bank", Description: "Australia’s largest bank offering retail, business, and institutional banking."},
			&Bank{Name: "Westpac", Description: "A major Australian bank providing nationwide and international financial services."},
			&Bank{Name: "ANZ (Australia and New Zealand Banking Group)", Description: "A leading bank serving Australia, NZ, and the Asia-Pacific region."},
			&Bank{Name: "NAB (National Australia Bank)", Description: "One of Australia's largest banks offering comprehensive financial services."},

			&Bank{Name: "Bank of Queensland", Description: "A retail bank operating across Queensland and other states."},
			&Bank{Name: "Bendigo Bank", Description: "A community-focused Australian bank providing retail financial services."},
			&Bank{Name: "Macquarie Bank", Description: "A global Australian financial services group specializing in investment banking."},

			&Bank{Name: "UP Bank", Description: "A digital Australian neobank offering mobile-first banking."},
			&Bank{Name: "86 400", Description: "A digital neobank in Australia offering smart mobile banking solutions."},
		)

	case "CAN": // Canada
		banks = append(banks,
			&Bank{Name: "Royal Bank of Canada (RBC)", Description: "Canada’s largest bank offering comprehensive financial services worldwide."},
			&Bank{Name: "TD Canada Trust", Description: "A major Canadian bank providing retail and commercial banking across Canada and the US."},
			&Bank{Name: "Scotiabank", Description: "A global Canadian bank with strong operations in the Americas."},
			&Bank{Name: "Bank of Montreal (BMO)", Description: "One of Canada’s oldest banks providing retail and commercial banking."},
			&Bank{Name: "CIBC (Canadian Imperial Bank of Commerce)", Description: "A major Canadian financial institution offering retail and corporate banking."},

			&Bank{Name: "National Bank of Canada", Description: "A leading bank in Quebec offering nationwide financial services."},
			&Bank{Name: "HSBC Canada", Description: "The Canadian branch of HSBC offering international banking services."},

			&Bank{Name: "Tangerine Bank", Description: "A Canadian online bank offering no-fee digital banking services."},
			&Bank{Name: "Simplii Financial", Description: "A fully digital bank operated by CIBC offering no-fee banking."},
			&Bank{Name: "EQ Bank", Description: "A Canadian online bank providing high-interest savings and digital banking."},
		)

	case "CHE": // Switzerland
		banks = append(banks,
			&Bank{Name: "UBS", Description: "Switzerland’s largest bank offering global wealth, retail, and investment banking services."},
			&Bank{Name: "Credit Suisse", Description: "A major Swiss bank known for wealth management and global investment banking."},

			&Bank{Name: "Julius Baer", Description: "A leading Swiss private bank focused on wealth and asset management."},
			&Bank{Name: "Raiffeisen Switzerland", Description: "A large cooperative bank group offering retail banking services."},
			&Bank{Name: "Zurich Cantonal Bank (ZKB)", Description: "A major Swiss cantonal bank offering retail and corporate banking."},

			&Bank{Name: "Neon Bank", Description: "A Swiss digital bank offering mobile-first personal banking."},
			&Bank{Name: "Yapeal", Description: "A digital banking platform in Switzerland providing modern mobile banking services."},
		)

	case "CHN": // China
		banks = append(banks,
			&Bank{Name: "Industrial and Commercial Bank of China (ICBC)", Description: "The largest bank in China and one of the biggest globally, offering retail and corporate banking."},
			&Bank{Name: "China Construction Bank (CCB)", Description: "A major state-owned bank specializing in infrastructure and housing finance."},
			&Bank{Name: "Agricultural Bank of China (ABC)", Description: "One of China’s biggest banks, serving rural and urban banking needs."},
			&Bank{Name: "Bank of China (BOC)", Description: "A leading Chinese bank with a strong global presence and international banking services."},

			&Bank{Name: "Bank of Communications", Description: "One of China’s oldest banks, offering commercial and retail banking services."},
			&Bank{Name: "China Merchants Bank (CMB)", Description: "A major commercial bank known for innovative retail and corporate banking."},
			&Bank{Name: "Shanghai Pudong Development Bank (SPDB)", Description: "A large commercial bank providing corporate finance and digital banking."},
			&Bank{Name: "China CITIC Bank", Description: "A mid-tier Chinese bank offering retail and corporate financial services."},

			&Bank{Name: "Alipay", Description: "China’s largest digital wallet and payment platform operated by Ant Group."},
			&Bank{Name: "WeChat Pay", Description: "A dominant mobile payment service integrated with the WeChat ecosystem."},
			&Bank{Name: "JD Finance", Description: "A fintech platform offering digital payments and consumer finance services."},
			&Bank{Name: "MyBank", Description: "An online-only bank backed by Ant Group offering digital financial services."},
			&Bank{Name: "WeBank", Description: "China’s first fully digital bank, owned by Tencent, offering mobile-first banking."},
		)
	case "SWE": // Sweden
		banks = append(banks,
			&Bank{Name: "Swedbank", Description: "One of Sweden’s largest banks offering retail, private, and corporate banking."},
			&Bank{Name: "SEB (Skandinaviska Enskilda Banken)", Description: "A major Nordic financial group providing business and private banking."},
			&Bank{Name: "Handelsbanken", Description: "A well-known Swedish bank offering retail and corporate banking across Europe."},
			&Bank{Name: "Nordea", Description: "The largest Nordic bank offering financial services across Sweden and Europe."},

			&Bank{Name: "Länsförsäkringar Bank", Description: "A Swedish retail bank offering consumer loans and savings."},
			&Bank{Name: "Ikano Bank", Description: "A consumer finance bank originally founded by the IKEA family."},

			&Bank{Name: "Klarna", Description: "A Swedish fintech giant known for buy-now-pay-later, digital payments, and e-wallet services."},
			&Bank{Name: "Revolut", Description: "A popular EU digital banking platform used widely in Sweden."},
			&Bank{Name: "P.F.C. (Personal Finance Co.)", Description: "A Swedish neobank offering mobile-first banking and budgeting tools."},
			&Bank{Name: "Rocker", Description: "A Swedish fintech offering digital financial services and payments."},
			&Bank{Name: "Swish", Description: "Sweden’s widely used mobile payment system linked to Swedish banks."},
		)

	case "NZL": // New Zealand
		banks = append(banks,
			&Bank{Name: "ANZ New Zealand", Description: "New Zealand’s largest bank providing retail, business, and wealth banking."},
			&Bank{Name: "ASB Bank", Description: "A major NZ bank offering consumer, business, and rural banking services."},
			&Bank{Name: "BNZ (Bank of New Zealand)", Description: "One of the oldest NZ banks offering full retail and commercial banking."},
			&Bank{Name: "Westpac New Zealand", Description: "A leading bank offering retail, business, and community banking services."},

			&Bank{Name: "Kiwibank", Description: "A New Zealand state-owned bank offering competitive retail banking services."},
			&Bank{Name: "TSB Bank", Description: "A New Zealand-owned bank providing personal and business banking."},
			&Bank{Name: "Heartland Bank", Description: "A New Zealand bank known for mortgages, small business lending, and digital services."},

			&Bank{Name: "Wise", Description: "A global digital bank widely used in New Zealand for international transfers."},
			&Bank{Name: "Revolut", Description: "A digital financial super-app offering multi-currency accounts and payments in NZ."},
			&Bank{Name: "Jersey", Description: "An upcoming NZ-based digital wallet and financial services provider."},
			&Bank{Name: "Apple Pay", Description: "A widely used mobile wallet supporting NZ contactless payments."},
			&Bank{Name: "Google Pay", Description: "A digital wallet for online and contactless payments in New Zealand."},
		)
	case "PHL": // Philippines
		banks = append(banks,

			&Bank{Name: "BDO Unibank", Description: "The largest bank in the Philippines offering retail, corporate, and international banking services."},
			&Bank{Name: "BPI (Bank of the Philippine Islands)", Description: "One of the oldest and largest universal banks in the Philippines."},
			&Bank{Name: "Metrobank", Description: "A major Philippine bank offering corporate and consumer financial services."},
			&Bank{Name: "Landbank of the Philippines", Description: "A government-owned bank focused on agriculture and public-sector banking."},
			&Bank{Name: "PNB (Philippine National Bank)", Description: "A leading universal bank with wide domestic and international operations."},
			&Bank{Name: "Security Bank", Description: "A top Philippine bank offering retail and business banking solutions."},
			&Bank{Name: "Chinabank", Description: "A Philippine bank known for SME banking and commercial services."},
			&Bank{Name: "RCBC (Rizal Commercial Banking Corporation)", Description: "A major commercial bank offering personal and corporate financial services."},
			&Bank{Name: "UnionBank of the Philippines", Description: "A leading digital bank in the Philippines offering innovative online banking."},

			&Bank{Name: "EastWest Bank", Description: "A universal bank offering consumer loans, credit cards, and retail banking."},
			&Bank{Name: "Asia United Bank (AUB)", Description: "A Philippine commercial bank known for strong business banking services."},
			&Bank{Name: "UCPB (United Coconut Planters Bank)", Description: "A commercial bank merged with Landbank providing retail and corporate banking."},
			&Bank{Name: "Sterling Bank of Asia", Description: "A thrift bank offering consumer and SME banking."},
			&Bank{Name: "Philtrust Bank", Description: "One of the oldest banks specializing in trust and retail banking."},
			&Bank{Name: "Robinsons Bank", Description: "A commercial bank serving retail and SME sectors."},

			&Bank{Name: "HSBC Philippines", Description: "The Philippine branch of HSBC offering retail and global banking services."},
			&Bank{Name: "Citibank Philippines", Description: "Offers credit cards and wealth banking; retail portfolio now merged with UnionBank."},
			&Bank{Name: "Standard Chartered Philippines", Description: "A multinational bank offering corporate and institutional banking."},
			&Bank{Name: "Maybank Philippines", Description: "A Malaysian bank providing retail and commercial banking in the Philippines."},
			&Bank{Name: "Bank of Tokyo-Mitsubishi UFJ Manila", Description: "A major Japanese bank offering corporate financing services."},

			&Bank{Name: "Maya Bank", Description: "A fully digital bank offering high-yield savings, wallets, and virtual cards."},
			&Bank{Name: "SeaBank", Description: "A digital bank under Shopee focusing on high-interest savings and digital services."},
			&Bank{Name: "Tonik Bank", Description: "The first neobank in the Philippines offering full digital banking."},
			&Bank{Name: "GoTyme Bank", Description: "A digital bank offering kiosks, debit cards, and mobile-first banking."},
			&Bank{Name: "UNObank", Description: "A digital bank providing fully online savings, loans, and financial services."},
			&Bank{Name: "Overseas Filipino Bank (OFBank)", Description: "The first government-owned digital bank focused on OFWs."},

			&Bank{Name: "GCash", Description: "The largest e-wallet in the Philippines offering payments, savings, and QR transactions."},
			&Bank{Name: "Maya", Description: "A widely used wallet and digital bank offering bills payment and remittances."},
			&Bank{Name: "ShopeePay", Description: "A popular wallet for online payments, transfers, and bills in Shopee ecosystem."},
			&Bank{Name: "Lazada Wallet", Description: "An e-wallet for online purchases and refunds on Lazada."},
			&Bank{Name: "GrabPay", Description: "A mobile wallet for Grab services, payments, and online shopping."},
			&Bank{Name: "Coins.ph", Description: "A digital wallet offering crypto, remittances, and bill payments."},
			&Bank{Name: "StarPay", Description: "A government-accredited e-wallet used for social aid distribution."},
			&Bank{Name: "Bayad Wallet", Description: "A digital wallet for bills payment and government services."},
			&Bank{Name: "Komo by EastWest", Description: "A digital-only banking service offering online savings with high interest."},

			&Bank{Name: "Card Bank", Description: "A microfinance-oriented rural bank serving low-income communities."},
			&Bank{Name: "BDO Network Bank", Description: "BDO’s rural bank focused on micro and small loans."},
			&Bank{Name: "One Network Bank", Description: "A rural bank (merged with BDO Network Bank) providing financial services in Mindanao."},
			&Bank{Name: "Cantilan Bank", Description: "A rural bank known for technological innovation and digital finance."},
			&Bank{Name: "Producers Bank", Description: "A growing rural bank offering retail loans and microfinance."},
			&Bank{Name: "Rizal Rural Bank", Description: "A rural bank offering traditional deposit and loan services."},
		)
	case "IND": // India
		banks = append(banks,
			&Bank{Name: "State Bank of India (SBI)", Description: "India’s largest public sector bank offering extensive nationwide and international banking services."},
			&Bank{Name: "HDFC Bank", Description: "A top private bank in India known for retail banking, loans, and digital banking services."},
			&Bank{Name: "ICICI Bank", Description: "A major private sector bank offering a wide range of banking and financial services across India."},
			&Bank{Name: "Punjab National Bank (PNB)", Description: "One of India's largest public sector banks with a strong branch network nationwide."},
			&Bank{Name: "Axis Bank", Description: "A well-known private bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "Kotak Mahindra Bank", Description: "A private bank in India known for innovative banking and wealth management services."},

			&Bank{Name: "Paytm", Description: "India’s leading digital wallet and payments platform offering e-money and banking services."},
			&Bank{Name: "PhonePe", Description: "A widely used UPI-based mobile wallet for digital payments across India."},
			&Bank{Name: "Google Pay India", Description: "UPI-based digital payment service widely used for instant transfers and payments."},
		)

	case "KOR": // South Korea
		banks = append(banks,
			&Bank{Name: "KB Kookmin Bank", Description: "South Korea’s largest bank, offering retail, corporate, and global banking services."},
			&Bank{Name: "Shinhan Bank", Description: "A major bank in Korea providing comprehensive financial and global banking services."},
			&Bank{Name: "Woori Bank", Description: "One of South Korea’s oldest banks offering extensive domestic and international services."},
			&Bank{Name: "Hana Bank", Description: "A leading Korean bank known for corporate, retail, and global finance solutions."},
			&Bank{Name: "Industrial Bank of Korea (IBK)", Description: "A government-owned bank serving small and medium enterprises in Korea."},

			&Bank{Name: "KakaoPay", Description: "South Korea’s leading mobile wallet integrated with Kakao ecosystem for payments and transfers."},
			&Bank{Name: "Naver Pay", Description: "A major e-wallet linked to the Naver platform for e-commerce and digital payments."},
			&Bank{Name: "Samsung Pay", Description: "A mobile payment system widely used in Korea for contactless and online payments."},
		)

	case "THA": // Thailand
		banks = append(banks,
			&Bank{Name: "Bangkok Bank", Description: "Thailand’s largest bank offering retail, commercial, and international banking."},
			&Bank{Name: "Kasikornbank (KBank)", Description: "A major Thai bank known for digital banking and SME support."},
			&Bank{Name: "Siam Commercial Bank (SCB)", Description: "One of the oldest and largest Thai banks offering full financial services."},
			&Bank{Name: "Krungthai Bank (KTB)", Description: "A state-owned commercial bank serving nationwide financial needs."},
			&Bank{Name: "TMBThanachart Bank (TTB)", Description: "A leading bank formed from the TMB and Thanachart merger offering retail banking services."},

			&Bank{Name: "TrueMoney Wallet", Description: "Thailand’s most popular e-wallet offering payments, transfers, and mobile banking features."},
			&Bank{Name: "PromptPay", Description: "A government-backed digital payment system widely used for QR and mobile transfers."},
			&Bank{Name: "AirPay / ShopeePay", Description: "A widely used digital wallet integrated with Shopee for payments and promotions."},
		)

	case "SGP": // Singapore
		banks = append(banks,
			&Bank{Name: "DBS Bank", Description: "Singapore’s largest bank known for strong digital banking and global financial operations."},
			&Bank{Name: "OCBC Bank", Description: "A major Singaporean multinational bank offering corporate, retail, and investment services."},
			&Bank{Name: "United Overseas Bank (UOB)", Description: "A leading bank in Southeast Asia offering full financial solutions."},
			&Bank{Name: "Standard Chartered Singapore", Description: "A global bank with strong presence in Singapore for personal and business banking."},
			&Bank{Name: "HSBC Singapore", Description: "A multinational bank offering extensive corporate and wealth management services in Singapore."},

			&Bank{Name: "GrabPay", Description: "A widely adopted digital wallet in Singapore used for payments, transfers, and ride-hailing services."},
			&Bank{Name: "PayNow", Description: "Singapore’s national QR and mobile transfer system supported by major banks."},
			&Bank{Name: "NETS", Description: "Singapore’s popular electronic payment system used nationwide."},
		)

	case "HKG": // Hong Kong
		banks = append(banks,
			&Bank{Name: "HSBC Hong Kong", Description: "One of Hong Kong’s most dominant banks offering personal and commercial banking services."},
			&Bank{Name: "Bank of China (Hong Kong)", Description: "A major state-backed bank offering corporate, personal, and cross-border banking services."},
			&Bank{Name: "Standard Chartered Hong Kong", Description: "A major multinational bank with strong presence in Hong Kong’s financial sector."},
			&Bank{Name: "Hang Seng Bank", Description: "A leading local bank known for retail, commercial, and wealth management services."},
			&Bank{Name: "Citibank Hong Kong", Description: "A major international bank offering global banking and wealth services in Hong Kong."},

			&Bank{Name: "AlipayHK", Description: "A popular Hong Kong e-wallet for QR payments and transfers."},
			&Bank{Name: "WeChat Pay HK", Description: "A major digital wallet supporting QR payments, transfers, and cross-border transactions."},
			&Bank{Name: "Octopus Wallet", Description: "Hong Kong’s iconic stored-value e-payment card and digital wallet used in transport and retail."},
		)

	case "MYS": // Malaysia
		banks = append(banks,
			&Bank{Name: "Maybank", Description: "Malaysia’s largest bank offering retail, corporate, and international banking services."},
			&Bank{Name: "CIMB Bank", Description: "A major Malaysian universal bank with extensive operations in ASEAN."},
			&Bank{Name: "Public Bank Berhad", Description: "One of Malaysia’s most stable banks known for customer-focused retail banking."},
			&Bank{Name: "RHB Bank", Description: "A leading Malaysian bank providing retail, commercial, and investment services."},
			&Bank{Name: "Hong Leong Bank", Description: "A major financial institution offering digital and traditional banking services across Malaysia."},

			&Bank{Name: "Touch 'n Go eWallet", Description: "Malaysia’s most popular e-wallet used for transport, retail, and online transactions."},
			&Bank{Name: "Boost", Description: "A leading Malaysian e-wallet used for QR payments, bills, and mobile reloads."},
			&Bank{Name: "GrabPay Malaysia", Description: "A widely adopted mobile wallet integrated with Grab's ecosystem."},
		)

	case "IDN": // Indonesia
		banks = append(banks,
			&Bank{Name: "Bank Rakyat Indonesia (BRI)", Description: "Indonesia’s largest bank serving retail, microfinance, and corporate sectors."},
			&Bank{Name: "Bank Mandiri", Description: "A major Indonesian bank offering comprehensive banking and financial services."},
			&Bank{Name: "Bank Central Asia (BCA)", Description: "Indonesia’s leading private bank known for strong digital banking services."},
			&Bank{Name: "Bank Negara Indonesia (BNI)", Description: "A state-owned bank offering retail, business, and global banking services."},
			&Bank{Name: "Permata Bank", Description: "A fast-growing private bank providing modern digital and retail banking services."},

			&Bank{Name: "GoPay", Description: "Indonesia’s most popular e-wallet used in Gojek and general QR payments."},
			&Bank{Name: "OVO", Description: "A major digital wallet used for shopping, bills, and rewards."},
			&Bank{Name: "DANA", Description: "A widely used e-wallet offering QRIS payments, transfers, and online transactions."},
			&Bank{Name: "ShopeePay Indonesia", Description: "An e-wallet integrated with Shopee and used widely for QRIS payments."},
		)

	case "VNM": // Vietnam
		banks = append(banks,
			&Bank{Name: "Vietcombank", Description: "Vietnam’s largest commercial bank offering a wide range of retail and corporate services."},
			&Bank{Name: "VietinBank", Description: "A major state-owned bank providing financial services nationwide."},
			&Bank{Name: "BIDV", Description: "One of the biggest Vietnamese banks offering personal and business banking."},
			&Bank{Name: "Techcombank", Description: "A fast-growing private bank known for strong digital banking services."},
			&Bank{Name: "MB Bank", Description: "A major Vietnamese bank offering modern digital and military-associated financial services."},

			&Bank{Name: "MoMo", Description: "Vietnam’s leading e-wallet offering payments, transfers, and mobile financial services."},
			&Bank{Name: "ZaloPay", Description: "A popular mobile wallet integrated with Zalo, Vietnam’s biggest messaging app."},
			&Bank{Name: "ShopeePay Vietnam", Description: "A digital wallet used for shopping and QR payments through Shopee."},
			&Bank{Name: "VNPay", Description: "A major Vietnamese QR payment platform used across shops nationwide."},
		)

	case "TWN": // Taiwan
		banks = append(banks,
			&Bank{Name: "Bank of Taiwan", Description: "Taiwan’s largest state-owned bank offering comprehensive financial services."},
			&Bank{Name: "CTBC Bank", Description: "A major Taiwanese bank known for digital innovations and overseas presence."},
			&Bank{Name: "Taipei Fubon Bank", Description: "A leading financial institution offering retail and corporate banking solutions."},
			&Bank{Name: "E.SUN Bank", Description: "A well-known Taiwanese bank focused on customer service and digital banking."},
			&Bank{Name: "Mega International Commercial Bank", Description: "A major bank offering global finance and corporate services."},

			&Bank{Name: "JKoPay", Description: "Taiwan’s most popular e-wallet for QR payments and online transactions."},
			&Bank{Name: "Line Pay Taiwan", Description: "A widely used mobile wallet integrated with LINE messaging."},
			&Bank{Name: "Apple Pay Taiwan", Description: "A major contactless payment service used across retail outlets."},
			&Bank{Name: "Taiwan Pay", Description: "A government-backed QR and mobile payment solution."},
		)

	case "BRN": // Brunei
		banks = append(banks,
			&Bank{Name: "Bank Islam Brunei Darussalam (BIBD)", Description: "Brunei’s largest bank offering Sharia-compliant retail and corporate banking."},
			&Bank{Name: "Baidhuri Bank", Description: "A major bank in Brunei offering personal, corporate, and digital banking services."},
			&Bank{Name: "Standard Chartered Brunei", Description: "An international bank with significant operations in Brunei."},

			&Bank{Name: "BIBD NEXGEN Wallet", Description: "A mobile wallet by BIBD supporting online payments and transfers."},
			&Bank{Name: "Progresif Pay", Description: "A Brunei digital wallet used for payments, mobile top-ups, and remittances."},
			&Bank{Name: "Beep Digital Wallet", Description: "A local e-wallet used for QR payments and small retail transactions."},
		)

	case "SAU": // Saudi Arabia
		banks = append(banks,
			&Bank{Name: "National Commercial Bank (NCB)", Description: "Saudi Arabia’s largest bank, also known as AlAhli Bank, offering retail and corporate banking."},
			&Bank{Name: "Al Rajhi Bank", Description: "One of the world’s largest Islamic banks offering personal, business, and digital banking services."},
			&Bank{Name: "Saudi British Bank (SABB)", Description: "A major Saudi bank partnered with HSBC offering international financial services."},
			&Bank{Name: "Riyad Bank", Description: "A leading bank in Saudi Arabia providing retail and corporate financial services."},
			&Bank{Name: "Arab National Bank (ANB)", Description: "A large Saudi bank offering extensive commercial and personal banking."},
			&Bank{Name: "Bank AlJazira", Description: "An Islamic bank offering Sharia-compliant financial products."},

			&Bank{Name: "STC Pay", Description: "Saudi Arabia’s largest digital wallet for QR payments, transfers, and international remittances."},
			&Bank{Name: "Mada Pay", Description: "A national payment system supporting mobile and contactless transactions."},
			&Bank{Name: "UrPay", Description: "A rising Saudi digital wallet for payments and bill transactions."},
		)

	case "ARE": // United Arab Emirates
		banks = append(banks,
			&Bank{Name: "Emirates NBD", Description: "Dubai’s largest banking group offering retail, corporate, and international banking."},
			&Bank{Name: "First Abu Dhabi Bank (FAB)", Description: "The UAE’s biggest bank offering personal, business, and investment services."},
			&Bank{Name: "Dubai Islamic Bank (DIB)", Description: "The world’s first Islamic bank offering Sharia-compliant solutions."},
			&Bank{Name: "Abu Dhabi Commercial Bank (ADCB)", Description: "A major UAE bank offering retail, commercial, and digital banking."},
			&Bank{Name: "Mashreq Bank", Description: "One of the oldest private banks in the UAE with strong digital banking."},

			&Bank{Name: "eWallet UAE", Description: "A popular UAE digital wallet used for payments and transfers."},
			&Bank{Name: "Apple Pay UAE", Description: "A widely used mobile wallet for contactless payments."},
			&Bank{Name: "Google Pay UAE", Description: "A mobile payment system used across UAE retailers and services."},
			&Bank{Name: "Careem Pay", Description: "A growing wallet integrated with Careem services and ride-hailing."},
		)

	case "ISR": // Israel
		banks = append(banks,
			&Bank{Name: "Bank Hapoalim", Description: "Israel’s largest bank offering comprehensive retail and corporate banking."},
			&Bank{Name: "Bank Leumi", Description: "One of Israel’s oldest and largest banks with strong digital banking services."},
			&Bank{Name: "Israel Discount Bank", Description: "A major bank providing personal, corporate, and investment banking."},
			&Bank{Name: "Mizrahi-Tefahot Bank", Description: "Israel’s third-largest bank known for mortgage lending."},
			&Bank{Name: "First International Bank of Israel (FIBI)", Description: "A large Israeli bank offering retail and commercial services."},

			&Bank{Name: "Bit", Description: "Israel’s most popular digital wallet allowing fast person-to-person and merchant payments."},
			&Bank{Name: "Pepper Pay", Description: "A digital banking app from Bank Leumi offering modern payment services."},
			&Bank{Name: "PayBox", Description: "A widely used mobile wallet for group payments and transfers."},
		)

	case "ZAF": // South Africa
		banks = append(banks,
			&Bank{Name: "Standard Bank", Description: "Africa’s largest bank offering retail, corporate, and investment banking services."},
			&Bank{Name: "First National Bank (FNB)", Description: "A major South African bank known for digital innovation and mobile banking."},
			&Bank{Name: "ABSA Bank", Description: "A leading financial services provider operating across Africa."},
			&Bank{Name: "Nedbank", Description: "A top-tier bank offering personal, business, and investment banking."},
			&Bank{Name: "Capitec Bank", Description: "One of South Africa’s biggest retail banks known for affordable digital banking."},

			&Bank{Name: "SnapScan", Description: "A popular South African QR payment app for fast retail transactions."},
			&Bank{Name: "Zapper", Description: "A widely used digital wallet using QR payments across shops and restaurants."},
			&Bank{Name: "Vodapay", Description: "A mobile wallet by Vodacom for online and in-store payments."},
		)

	case "EGY": // Egypt
		banks = append(banks,
			&Bank{Name: "National Bank of Egypt (NBE)", Description: "Egypt’s oldest and largest bank offering comprehensive financial services."},
			&Bank{Name: "Banque Misr", Description: "A major Egyptian state-owned bank providing retail and business banking."},
			&Bank{Name: "Commercial International Bank (CIB)", Description: "Egypt’s largest private bank known for digital banking."},
			&Bank{Name: "Banque du Caire", Description: "One of Egypt’s top banks offering nationwide retail and commercial services."},
			&Bank{Name: "QNB Alahli", Description: "A subsidiary of Qatar National Bank offering corporate and personal banking."},

			&Bank{Name: "Vodafone Cash", Description: "Egypt’s largest mobile wallet used for payments, transfers, and bill payments."},
			&Bank{Name: "Etisalat Cash", Description: "A popular e-wallet service in Egypt for transfers and PMT services."},
			&Bank{Name: "Orange Money Egypt", Description: "A mobile wallet offering payments and remittance services."},
			&Bank{Name: "Meeza Digital Wallet", Description: "Egypt’s national payment wallet supporting QR and online payments."},
		)

	case "TUR": // Turkey
		banks = append(banks,
			&Bank{Name: "Ziraat Bankası", Description: "Turkey’s largest state-owned bank offering comprehensive financial services nationwide."},
			&Bank{Name: "Türkiye İş Bankası", Description: "Turkey’s oldest and largest private bank known for retail and corporate banking."},
			&Bank{Name: "Garanti BBVA", Description: "A major Turkish bank known for digital banking and corporate services."},
			&Bank{Name: "Akbank", Description: "One of Turkey’s leading banks offering modern digital and financial services."},
			&Bank{Name: "Halkbank", Description: "A state-owned bank focused on SMEs and retail banking."},

			&Bank{Name: "Papara", Description: "Turkey’s most popular digital wallet offering payments, cards, and money transfers."},
			&Bank{Name: "FastPay", Description: "A digital wallet by DenizBank allowing free money transfers and bill payments."},
			&Bank{Name: "Paycell", Description: "A mobile wallet by Turkcell allowing payments, bills, and QR transactions."},
			&Bank{Name: "Tosla", Description: "A modern e-wallet by Akbank for online payments and transfers."},
		)

	case "SEN": // Senegal (West African CFA Franc representative)
		banks = append(banks,
			&Bank{Name: "Ecobank", Description: "A major pan-African bank with strong presence in West Africa offering retail and corporate banking."},
			&Bank{Name: "Bank of Africa (BOA)", Description: "A leading banking group operating across many West African countries."},
			&Bank{Name: "Société Générale de Banques", Description: "A regional arm of Société Générale offering banking in multiple West African states."},
			&Bank{Name: "Coris Bank International", Description: "A growing West African bank offering personal and business banking."},
			&Bank{Name: "UBA (United Bank for Africa)", Description: "A pan-African bank offering retail banking, business banking, and digital services."},

			&Bank{Name: "Orange Money", Description: "One of the most widely used mobile wallets in West Africa for payments and transfers."},
			&Bank{Name: "MTN Mobile Money (MoMo)", Description: "A major mobile wallet enabling payments, transfers, and merchant transactions."},
			&Bank{Name: "Wave Mobile Money", Description: "A fast-growing mobile money platform focused on low-cost transfers."},
		)

	case "CMR": // Cameroon (Central African CFA Franc representative)
		banks = append(banks,
			&Bank{Name: "Afriland First Bank", Description: "A leading bank in Central Africa offering retail and business banking."},
			&Bank{Name: "BGFI Bank", Description: "One of the largest regional banking groups operating widely across Central Africa."},
			&Bank{Name: "Ecobank Cameroon", Description: "A major bank offering retail, corporate, and digital services in the region."},
			&Bank{Name: "UBA Cameroon", Description: "A regional branch of United Bank for Africa offering modern banking services."},
			&Bank{Name: "Société Générale Cameroun", Description: "A major French-backed bank operating in Central Africa."},

			&Bank{Name: "MTN Mobile Money", Description: "A dominant mobile wallet for payments and remittances across Central Africa."},
			&Bank{Name: "Airtel Money", Description: "A leading digital wallet used for transfers, bills, and merchant payments."},
			&Bank{Name: "Orange Money Central Africa", Description: "A widely used mobile money service for payments and everyday transactions."},
		)

	case "MUS": // Mauritius
		banks = append(banks,
			&Bank{Name: "Mauritius Commercial Bank (MCB)", Description: "The largest bank in Mauritius offering retail, corporate, and international services."},
			&Bank{Name: "State Bank of Mauritius (SBM)", Description: "A major Mauritian bank offering retail and business banking."},
			&Bank{Name: "Bank One", Description: "A leading bank in Mauritius providing modern banking and digital solutions."},
			&Bank{Name: "AfrAsia Bank", Description: "A private bank offering wealth management and business banking."},
			&Bank{Name: "ABC Banking Corporation", Description: "A dynamic bank focused on personal and SME banking services."},

			&Bank{Name: "Juice by MCB", Description: "Mauritius’ most popular digital wallet used for payments, bills, and transfers."},
			&Bank{Name: "my.t Money", Description: "A mobile wallet by Mauritius Telecom offering digital payments and top-ups."},
			&Bank{Name: "SBM EasyPay", Description: "A mobile payment service by SBM Bank."},
		)

	case "MDV": // Maldives
		banks = append(banks,
			&Bank{Name: "Bank of Maldives (BML)", Description: "The largest bank in the Maldives offering extensive financial services."},
			&Bank{Name: "Maldives Islamic Bank (MIB)", Description: "A major Islamic bank offering Sharia-compliant financial services."},
			&Bank{Name: "The Mauritius Commercial Bank (Maldives)", Description: "A foreign branch offering corporate and retail banking in the Maldives."},
			&Bank{Name: "State Bank of India (Maldives Branch)", Description: "A major foreign bank offering retail services in the Maldives."},

			&Bank{Name: "BML MobilePay", Description: "A leading digital wallet by Bank of Maldives for payments and transfers."},
			&Bank{Name: "Ooredoo m-Faisaa", Description: "A popular mobile wallet offering payments, transfers, and utility services."},
			&Bank{Name: "DhiraaguPay", Description: "A mobile payment service for QR transactions and online payments."},
		)

	case "NOR": // Norway
		banks = append(banks,
			&Bank{Name: "DNB ASA", Description: "Norway’s largest financial services group offering retail, corporate, and investment banking."},
			&Bank{Name: "Nordea Norway", Description: "A major bank providing retail and corporate services in Norway."},
			&Bank{Name: "SpareBank 1", Description: "A group of savings banks offering retail banking and mortgage services."},
			&Bank{Name: "Handelsbanken Norway", Description: "A Swedish bank with strong operations in Norway for corporate and private banking."},
			&Bank{Name: "Sbanken", Description: "A fully digital bank in Norway offering modern online banking solutions."},

			&Bank{Name: "Vipps", Description: "Norway’s most popular mobile payment app for transfers, payments, and QR transactions."},
			&Bank{Name: "Apple Pay Norway", Description: "Widely used contactless payment service linked to Norwegian banks."},
			&Bank{Name: "Google Pay Norway", Description: "Mobile payment solution for online and in-store payments."},
		)

	case "DNK": // Denmark
		banks = append(banks,
			&Bank{Name: "Danske Bank", Description: "Denmark’s largest bank offering retail, corporate, and international banking."},
			&Bank{Name: "Nordea Denmark", Description: "A major bank with full digital and corporate banking services in Denmark."},
			&Bank{Name: "Jyske Bank", Description: "A Danish bank offering personal, corporate, and investment banking solutions."},
			&Bank{Name: "Sydbank", Description: "A large Danish bank providing retail and business banking services."},
			&Bank{Name: "Nykredit Bank", Description: "Denmark’s top mortgage lender and full-service financial institution."},

			&Bank{Name: "MobilePay", Description: "Denmark’s most widely used mobile payment app for P2P and merchant payments."},
			&Bank{Name: "Apple Pay Denmark", Description: "Contactless payment solution used in retail and online shopping."},
			&Bank{Name: "Google Pay Denmark", Description: "Mobile wallet used for in-store and online payments."},
		)

	case "POL": // Poland
		banks = append(banks,
			&Bank{Name: "PKO Bank Polski", Description: "Poland’s largest bank offering retail, corporate, and investment banking."},
			&Bank{Name: "Bank Pekao", Description: "A major Polish bank providing comprehensive financial services nationwide."},
			&Bank{Name: "mBank", Description: "One of Poland’s leading digital banks offering innovative online banking services."},
			&Bank{Name: "Santander Bank Polska", Description: "A major international bank offering personal and corporate banking in Poland."},
			&Bank{Name: "ING Bank Śląski", Description: "A top bank in Poland known for retail, digital, and corporate banking solutions."},

			&Bank{Name: "Blik", Description: "Poland’s most popular mobile payment system supporting transfers and QR payments."},
			&Bank{Name: "PayPal Poland", Description: "Widely used for online payments and transfers."},
			&Bank{Name: "Google Pay Poland", Description: "Mobile wallet used for payments and digital purchases."},
		)

	case "CZE": // Czech Republic
		banks = append(banks,
			&Bank{Name: "Česká spořitelna", Description: "One of the largest Czech banks offering retail, corporate, and investment banking."},
			&Bank{Name: "ČSOB", Description: "A major bank in the Czech Republic providing personal and business financial services."},
			&Bank{Name: "Komerční banka", Description: "A top Czech bank offering retail and corporate banking, part of Société Générale."},
			&Bank{Name: "Raiffeisenbank Czech Republic", Description: "An international bank providing retail and corporate banking services."},
			&Bank{Name: "Moneta Money Bank", Description: "A growing bank in the Czech Republic focused on retail and digital banking."},

			&Bank{Name: "Twisto", Description: "A Czech e-wallet and payment app for online purchases and contactless payments."},
			&Bank{Name: "mBank CZ Wallet", Description: "Digital banking wallet for online and mobile transactions."},
			&Bank{Name: "Google Pay Czech Republic", Description: "Mobile wallet for payments, transfers, and online purchases."},
		)

	case "HUN": // Hungary
		banks = append(banks,
			&Bank{Name: "OTP Bank", Description: "Hungary’s largest bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "K&H Bank", Description: "A major Hungarian bank providing personal and business financial services."},
			&Bank{Name: "Erste Bank Hungary", Description: "Part of Erste Group, offering retail and corporate banking solutions."},
			&Bank{Name: "Raiffeisen Bank Hungary", Description: "A leading bank in Hungary offering modern digital and traditional banking."},
			&Bank{Name: "CIB Bank", Description: "A top Hungarian bank providing full-service retail and corporate banking."},

			&Bank{Name: "Simple by OTP", Description: "Hungary’s leading digital banking app with wallet and payment services."},
			&Bank{Name: "Revolut Hungary", Description: "A digital banking and e-wallet solution popular in Hungary."},
			&Bank{Name: "Google Pay Hungary", Description: "Mobile wallet for in-store and online payments."},
		)

	case "RUS": // Russia
		banks = append(banks,
			&Bank{Name: "Sberbank", Description: "Russia’s largest bank offering retail, corporate, and investment banking nationwide."},
			&Bank{Name: "VTB Bank", Description: "A major state-owned bank providing retail, corporate, and international financial services."},
			&Bank{Name: "Gazprombank", Description: "A leading Russian bank providing corporate and retail banking services."},
			&Bank{Name: "Alfa-Bank", Description: "One of Russia’s largest private banks offering modern digital and retail banking services."},
			&Bank{Name: "Tinkoff Bank", Description: "A fully digital bank in Russia known for online banking and e-wallet services."},

			&Bank{Name: "Yandex Money (now YooMoney)", Description: "A popular Russian e-wallet for online payments and transfers."},
			&Bank{Name: "Qiwi Wallet", Description: "Widely used digital wallet and payment service in Russia."},
			&Bank{Name: "SberPay", Description: "Digital wallet service offered by Sberbank for QR and online payments."},
		)

	case "HRV": // Croatia
		banks = append(banks,
			&Bank{Name: "Zagrebačka banka", Description: "Croatia’s largest bank offering retail, corporate, and investment services."},
			&Bank{Name: "Privredna banka Zagreb (PBZ)", Description: "A major Croatian bank providing comprehensive banking services."},
			&Bank{Name: "Raiffeisenbank Croatia", Description: "Part of Raiffeisen Group, offering retail and corporate banking in Croatia."},
			&Bank{Name: "Erste Bank Croatia", Description: "A leading bank in Croatia with digital and traditional banking solutions."},
			&Bank{Name: "OTP Bank Croatia", Description: "A significant regional bank offering personal and business banking."},

			&Bank{Name: "Settle Croatia", Description: "A digital wallet used for payments, QR transactions, and online purchases."},
			&Bank{Name: "Apple Pay Croatia", Description: "Contactless payment service supported by Croatian banks."},
			&Bank{Name: "Google Pay Croatia", Description: "Mobile wallet for online and in-store payments."},
		)

	case "BRA": // Brazil
		banks = append(banks,
			&Bank{Name: "Banco do Brasil", Description: "Brazil’s largest bank providing retail, corporate, and investment banking services."},
			&Bank{Name: "Itaú Unibanco", Description: "A major private Brazilian bank offering full-service banking and digital solutions."},
			&Bank{Name: "Bradesco", Description: "One of Brazil’s largest banks, offering retail and corporate banking nationwide."},
			&Bank{Name: "Santander Brazil", Description: "The Brazilian subsidiary of Santander, offering comprehensive financial services."},
			&Bank{Name: "Caixa Econômica Federal", Description: "A state-owned bank in Brazil known for retail banking and social programs."},

			&Bank{Name: "PicPay", Description: "Brazil’s most popular digital wallet offering payments, transfers, and QR transactions."},
			&Bank{Name: "Mercado Pago", Description: "Digital wallet by Mercado Libre for payments, QR, and online purchases."},
			&Bank{Name: "Nubank Wallet", Description: "Digital bank and wallet providing online payments and transfers in Brazil."},
		)

	case "MEX": // Mexico
		banks = append(banks,
			&Bank{Name: "BBVA México", Description: "A major bank in Mexico providing retail, corporate, and digital banking services."},
			&Bank{Name: "Banorte", Description: "One of Mexico’s largest banks offering retail, corporate, and investment banking."},
			&Bank{Name: "Santander Mexico", Description: "A large international bank offering personal and corporate banking solutions in Mexico."},
			&Bank{Name: "HSBC Mexico", Description: "A multinational bank providing full banking services in Mexico."},
			&Bank{Name: "Scotiabank Mexico", Description: "A leading Canadian bank offering banking services across Mexico."},

			&Bank{Name: "Mercado Pago Mexico", Description: "Digital wallet used for payments, transfers, and e-commerce in Mexico."},
			&Bank{Name: "Clip Wallet", Description: "A mobile wallet and card reader solution widely used for retail payments in Mexico."},
			&Bank{Name: "BBVA Wallet Mexico", Description: "Digital wallet provided by BBVA for online and contactless payments."},
		)

	case "ARG": // Argentina
		banks = append(banks,
			&Bank{Name: "Banco de la Nación Argentina", Description: "Argentina’s largest state-owned bank offering retail and corporate banking."},
			&Bank{Name: "Banco Galicia", Description: "A major private bank in Argentina offering digital and traditional banking services."},
			&Bank{Name: "Banco Santander Río", Description: "A leading bank in Argentina part of Santander Group providing full-service banking."},
			&Bank{Name: "BBVA Argentina", Description: "A major private bank offering retail and corporate banking solutions in Argentina."},
			&Bank{Name: "Banco Macro", Description: "A large Argentine bank providing services to individuals and businesses nationwide."},

			&Bank{Name: "Mercado Pago Argentina", Description: "One of Argentina’s most popular digital wallets for payments and transfers."},
			&Bank{Name: "Ualá", Description: "A mobile wallet and digital bank used widely for online payments and card services."},
			&Bank{Name: "Todo Pago", Description: "A digital payment platform enabling QR, online, and in-store transactions in Argentina."},
		)

	case "CHL": // Chile
		banks = append(banks,
			&Bank{Name: "Banco de Chile", Description: "One of Chile’s largest banks offering retail, corporate, and investment banking services."},
			&Bank{Name: "Banco Santander Chile", Description: "A leading international bank providing comprehensive banking services in Chile."},
			&Bank{Name: "BancoEstado", Description: "The state-owned bank offering financial services to individuals and businesses nationwide."},
			&Bank{Name: "Banco BCI", Description: "A major private bank in Chile offering retail, corporate, and digital banking."},
			&Bank{Name: "Scotiabank Chile", Description: "A branch of the Canadian bank providing retail and commercial banking in Chile."},

			&Bank{Name: "Mercado Pago Chile", Description: "A popular digital wallet for online and in-store payments."},
			&Bank{Name: "Mach", Description: "A Chilean digital wallet offering transfers, payments, and savings features."},
			&Bank{Name: "Flow Chile", Description: "A widely used e-wallet for online payments and QR transactions."},
		)

	case "COL": // Colombia
		banks = append(banks,
			&Bank{Name: "Bancolombia", Description: "Colombia’s largest bank offering retail, corporate, and investment banking."},
			&Bank{Name: "Banco de Bogotá", Description: "A major Colombian bank providing personal, corporate, and international banking services."},
			&Bank{Name: "Davivienda", Description: "A leading bank in Colombia known for retail banking and digital services."},
			&Bank{Name: "BBVA Colombia", Description: "A major international bank offering full-service banking in Colombia."},
			&Bank{Name: "Banco Popular", Description: "A Colombian bank providing retail and SME banking services."},

			&Bank{Name: "Nequi", Description: "A widely used Colombian mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "Daviplata", Description: "A digital wallet by Davivienda for P2P transfers and online payments."},
			&Bank{Name: "Movii", Description: "A Colombian e-wallet offering payments, transfers, and online transactions."},
		)

	case "PER": // Peru
		banks = append(banks,
			&Bank{Name: "Banco de Crédito del Perú (BCP)", Description: "The largest bank in Peru offering retail, corporate, and investment banking services."},
			&Bank{Name: "BBVA Perú", Description: "A major private bank in Peru offering digital, retail, and corporate banking solutions."},
			&Bank{Name: "Scotiabank Perú", Description: "Canadian bank providing full banking services in Peru."},
			&Bank{Name: "Interbank", Description: "A leading Peruvian bank known for retail and digital banking solutions."},
			&Bank{Name: "Banco Pichincha Perú", Description: "A commercial bank offering personal and business banking services in Peru."},

			&Bank{Name: "Yape", Description: "Peru’s most popular e-wallet for QR payments, transfers, and bill payments."},
			&Bank{Name: "Plin", Description: "A digital wallet used for instant transfers and payments in Peru."},
			&Bank{Name: "Tunki", Description: "A mobile banking and wallet app offered by BBVA Peru for payments and money transfers."},
		)

	case "URY": // Uruguay
		banks = append(banks,
			&Bank{Name: "Banco República (BROU)", Description: "Uruguay’s largest state-owned bank offering retail and corporate banking."},
			&Bank{Name: "Banco Santander Uruguay", Description: "A major international bank providing full banking services in Uruguay."},
			&Bank{Name: "BBVA Uruguay", Description: "A leading private bank in Uruguay offering digital and traditional banking."},
			&Bank{Name: "Itaú Uruguay", Description: "A top private bank providing retail and corporate banking services."},
			&Bank{Name: "Scotiabank Uruguay", Description: "A Canadian bank branch providing personal and business banking in Uruguay."},

			&Bank{Name: "Banred Wallet", Description: "A popular Uruguayan e-wallet for online payments and transfers."},
			&Bank{Name: "Redpagos Mobile", Description: "Digital wallet for payments and QR-based transactions."},
			&Bank{Name: "Mercado Pago Uruguay", Description: "A widely used digital wallet integrated with Mercado Libre for payments."},
		)

	case "DOM": // Dominican Republic
		banks = append(banks,
			&Bank{Name: "Banco Popular Dominicano", Description: "The largest bank in the Dominican Republic offering retail and corporate banking."},
			&Bank{Name: "Banco BHD León", Description: "A leading bank providing personal, business, and digital banking services."},
			&Bank{Name: "Scotiabank Dominican Republic", Description: "A branch of the Canadian bank offering full banking services in the DR."},
			&Bank{Name: "Banco del Progreso", Description: "A local bank offering retail and SME banking services."},
			&Bank{Name: "Banreservas", Description: "State-owned bank offering nationwide banking services."},

			&Bank{Name: "BHD Wallet", Description: "Digital wallet by BHD León for online payments and transfers."},
			&Bank{Name: "Teke Wallet", Description: "A mobile wallet used in the Dominican Republic for payments and transfers."},
			&Bank{Name: "Mercado Pago DR", Description: "Digital wallet integrated with Mercado Libre for payments and purchases."},
		)

	case "PRY": // Paraguay
		banks = append(banks,
			&Bank{Name: "Banco Nacional de Fomento (BNF)", Description: "Paraguay’s state-owned bank providing retail, agricultural, and corporate banking."},
			&Bank{Name: "Banco Familiar", Description: "A major private bank in Paraguay offering personal and business banking services."},
			&Bank{Name: "Banco Itaú Paraguay", Description: "A regional branch of Itaú offering full-service banking in Paraguay."},
			&Bank{Name: "Banco Continental", Description: "A leading Paraguayan bank providing retail and corporate banking solutions."},
			&Bank{Name: "Banco Regional", Description: "A bank offering commercial and personal banking services in Paraguay."},

			&Bank{Name: "Billetera Personal BNF", Description: "Digital wallet service offered by BNF for payments and transfers."},
			&Bank{Name: "Tigo Money", Description: "A mobile wallet used in Paraguay for transfers and payments."},
			&Bank{Name: "Bancard Wallet", Description: "Digital wallet for online and merchant payments in Paraguay."},
		)

	case "BOL": // Bolivia
		banks = append(banks,
			&Bank{Name: "Banco de Crédito de Bolivia", Description: "A major Bolivian bank offering retail and corporate banking services."},
			&Bank{Name: "Banco Nacional de Bolivia (BNB)", Description: "One of Bolivia’s largest banks providing full-service banking nationwide."},
			&Bank{Name: "Banco Mercantil Santa Cruz", Description: "A top private bank offering digital, corporate, and retail banking."},
			&Bank{Name: "Banco BISA", Description: "A major Bolivian bank providing personal and business banking services."},
			&Bank{Name: "Banco Fortaleza", Description: "A growing private bank in Bolivia focusing on SMEs and retail clients."},

			&Bank{Name: "Tigo Money Bolivia", Description: "A mobile wallet widely used for payments, top-ups, and transfers."},
			&Bank{Name: "Billetera BNB", Description: "Digital wallet service provided by Banco Nacional de Bolivia for payments and transfers."},
			&Bank{Name: "PagoMovil Bolivia", Description: "A mobile wallet for digital transactions and QR payments."},
		)

	case "VEN": // Venezuela
		banks = append(banks,
			&Bank{Name: "Banco de Venezuela", Description: "State-owned bank providing retail and corporate banking services across Venezuela."},
			&Bank{Name: "Banesco Banco Universal", Description: "One of Venezuela’s largest private banks offering full banking services."},
			&Bank{Name: "Banco Mercantil", Description: "A major Venezuelan bank providing personal, business, and online banking."},
			&Bank{Name: "Banco Provincial", Description: "A leading private bank in Venezuela focusing on retail and digital services."},
			&Bank{Name: "Banco Exterior", Description: "A commercial bank offering corporate and personal banking solutions."},

			&Bank{Name: "Pago Móvil", Description: "Venezuela’s most widely used mobile payment system for transfers and bills."},
			&Bank{Name: "Mercado Pago Venezuela", Description: "Digital wallet integrated with Mercado Libre for online purchases."},
			&Bank{Name: "Zelle Venezuela", Description: "Popular international digital wallet for P2P transfers in USD via local banks."},
		)

	case "PAK": // Pakistan
		banks = append(banks,
			&Bank{Name: "Habib Bank Limited (HBL)", Description: "Pakistan’s largest bank offering retail, corporate, and international banking."},
			&Bank{Name: "United Bank Limited (UBL)", Description: "A major Pakistani bank providing personal, business, and digital banking services."},
			&Bank{Name: "MCB Bank", Description: "A leading bank in Pakistan offering comprehensive financial solutions."},
			&Bank{Name: "Allied Bank", Description: "A large bank providing retail, corporate, and SME banking services."},
			&Bank{Name: "Bank Alfalah", Description: "A major private bank offering modern digital banking services."},

			&Bank{Name: "Easypaisa", Description: "Pakistan’s largest mobile wallet and branchless banking service."},
			&Bank{Name: "JazzCash", Description: "A widely used mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "UPaisa", Description: "Digital wallet offering online payments and P2P transfers in Pakistan."},
		)

	case "BGD": // Bangladesh
		banks = append(banks,
			&Bank{Name: "BRAC Bank", Description: "A leading private bank in Bangladesh offering retail, SME, and corporate banking."},
			&Bank{Name: "Dutch-Bangla Bank", Description: "A top bank in Bangladesh known for digital banking and widespread ATM network."},
			&Bank{Name: "Standard Chartered Bank Bangladesh", Description: "A multinational bank providing full-service banking in Bangladesh."},
			&Bank{Name: "Islami Bank Bangladesh", Description: "The largest Islamic bank in Bangladesh offering Sharia-compliant financial services."},
			&Bank{Name: "Prime Bank", Description: "A private commercial bank offering personal, corporate, and digital banking."},

			&Bank{Name: "bKash", Description: "Bangladesh’s leading mobile wallet used for payments, transfers, and merchant services."},
			&Bank{Name: "Nagad", Description: "A digital wallet and mobile banking service for money transfers and payments."},
			&Bank{Name: "Rocket by Dutch-Bangla Bank", Description: "A popular mobile wallet and branchless banking service in Bangladesh."},
		)

	case "LKA": // Sri Lanka
		banks = append(banks,
			&Bank{Name: "Bank of Ceylon", Description: "Sri Lanka’s largest state-owned bank providing retail and corporate banking services."},
			&Bank{Name: "Commercial Bank of Ceylon", Description: "A leading private bank offering personal, corporate, and digital banking services."},
			&Bank{Name: "Hatton National Bank (HNB)", Description: "A top private bank providing retail, corporate, and online banking solutions."},
			&Bank{Name: "Sampath Bank", Description: "A major bank in Sri Lanka known for modern banking and digital services."},
			&Bank{Name: "People’s Bank", Description: "State-owned bank offering financial services to individuals and businesses nationwide."},

			&Bank{Name: "ezCash", Description: "Sri Lanka’s most popular mobile wallet for payments, transfers, and top-ups."},
			&Bank{Name: "mCash", Description: "A digital wallet enabling mobile payments and online transactions."},
			&Bank{Name: "FriMi", Description: "Digital banking and wallet app offering payments and transfers in Sri Lanka."},
		)

	case "NPL": // Nepal
		banks = append(banks,
			&Bank{Name: "Nepal Investment Bank", Description: "One of Nepal’s largest banks offering retail and corporate banking services."},
			&Bank{Name: "Standard Chartered Bank Nepal", Description: "A multinational bank providing full-service banking in Nepal."},
			&Bank{Name: "Nabil Bank", Description: "A leading private bank in Nepal offering modern digital and retail banking."},
			&Bank{Name: "Himalayan Bank", Description: "A top bank providing personal, corporate, and SME banking in Nepal."},
			&Bank{Name: "Everest Bank", Description: "A private bank offering comprehensive banking and digital services in Nepal."},

			&Bank{Name: "eSewa", Description: "Nepal’s most popular digital wallet for payments, money transfers, and bills."},
			&Bank{Name: "IME Pay", Description: "A widely used mobile wallet offering online payments and remittances."},
			&Bank{Name: "Khalti", Description: "Digital payment platform for QR payments, bills, and top-ups in Nepal."},
		)

	case "MMR": // Myanmar
		banks = append(banks,
			&Bank{Name: "Kanbawza Bank (KBZ)", Description: "Myanmar’s largest private bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "Ayeyarwady Bank (AYA Bank)", Description: "A major private bank providing personal and business banking solutions."},
			&Bank{Name: "CB Bank", Description: "A leading bank in Myanmar offering modern banking and digital services."},
			&Bank{Name: "Myanmar Economic Bank (MEB)", Description: "State-owned bank providing financial services nationwide."},
			&Bank{Name: "Yoma Bank", Description: "A private bank offering corporate, retail, and SME banking in Myanmar."},

			&Bank{Name: "Wave Money", Description: "The most widely used mobile wallet in Myanmar for transfers, payments, and bills."},
			&Bank{Name: "KBZPay", Description: "Digital wallet offered by KBZ Bank for payments and remittances."},
			&Bank{Name: "AYA Pay", Description: "Mobile wallet service provided by AYA Bank for digital payments."},
		)

	case "KHM": // Cambodia
		banks = append(banks,
			&Bank{Name: "ACLEDA Bank", Description: "Cambodia’s largest commercial bank offering retail, SME, and corporate banking services."},
			&Bank{Name: "Canadia Bank", Description: "A major bank in Cambodia providing personal, corporate, and digital banking."},
			&Bank{Name: "Foreign Trade Bank of Cambodia (FTB)", Description: "A leading bank focused on international trade and commercial banking."},
			&Bank{Name: "ABA Bank", Description: "A fast-growing bank in Cambodia known for digital banking and e-wallet services."},
			&Bank{Name: "Cambodia Commercial Bank", Description: "Provides retail, corporate, and SME banking services across Cambodia."},

			&Bank{Name: "Pi Pay", Description: "One of Cambodia’s most popular digital wallets for payments and transfers."},
			&Bank{Name: "ABA Mobile", Description: "Digital wallet offered by ABA Bank for online and QR payments."},
			&Bank{Name: "Wing Money", Description: "Mobile wallet and payment service widely used in Cambodia for transfers and bills."},
		)

	case "LAO": // Laos
		banks = append(banks,
			&Bank{Name: "Banque pour le Commerce Extérieur Lao (BCEL)", Description: "The largest bank in Laos providing retail, corporate, and trade banking."},
			&Bank{Name: "ACLEDA Bank Laos", Description: "A major private bank offering retail, SME, and digital banking services."},
			&Bank{Name: "Lao Development Bank", Description: "State-owned bank providing commercial banking services across Laos."},
			&Bank{Name: "Sathapana Bank Laos", Description: "Bank offering retail and business banking solutions in Laos."},
			&Bank{Name: "Phongsavanh Bank", Description: "A growing private bank offering personal and corporate financial services."},

			&Bank{Name: "BCEL One", Description: "Digital wallet by BCEL for online payments, transfers, and bill payments."},
			&Bank{Name: "Wing Laos", Description: "Mobile wallet and payment service widely used in Laos."},
			&Bank{Name: "ACLEDA Unity Mobile", Description: "Mobile banking app for payments and digital transactions in Laos."},
		)

	case "NGA": // Nigeria
		banks = append(banks,
			&Bank{Name: "Access Bank", Description: "One of Nigeria’s largest banks offering retail, corporate, and digital banking services."},
			&Bank{Name: "Zenith Bank", Description: "A leading Nigerian bank providing personal, corporate, and investment banking."},
			&Bank{Name: "Guaranty Trust Bank (GTB)", Description: "A top Nigerian bank known for retail and digital banking services."},
			&Bank{Name: "First Bank of Nigeria", Description: "Nigeria’s oldest bank offering comprehensive banking solutions nationwide."},
			&Bank{Name: "United Bank for Africa (UBA)", Description: "A pan-African bank providing services across multiple countries including Nigeria."},

			&Bank{Name: "Paga", Description: "Nigeria’s most popular mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "Opay", Description: "Digital wallet offering payments, money transfers, and ride-hailing services in Nigeria."},
			&Bank{Name: "Kuda Bank Wallet", Description: "A fully digital bank and wallet service providing easy online payments."},
		)

	case "KEN": // Kenya
		banks = append(banks,
			&Bank{Name: "Equity Bank", Description: "Kenya’s largest bank providing retail, corporate, and digital banking services."},
			&Bank{Name: "KCB Bank", Description: "A leading bank in Kenya offering full-service banking and financial solutions."},
			&Bank{Name: "Co-operative Bank of Kenya", Description: "A major Kenyan bank serving retail, corporate, and SME clients."},
			&Bank{Name: "Standard Chartered Kenya", Description: "International bank providing personal, business, and digital banking."},
			&Bank{Name: "NIC Bank", Description: "A mid-sized bank offering retail and corporate banking solutions."},

			&Bank{Name: "M-Pesa", Description: "Kenya’s most popular mobile wallet for payments, transfers, and micro-financing."},
			&Bank{Name: "Equitel", Description: "Mobile banking and wallet service by Equity Bank for seamless transactions."},
			&Bank{Name: "KCB Mobi", Description: "Digital wallet provided by KCB for payments, transfers, and online banking."},
		)

	case "GHA": // Ghana
		banks = append(banks,
			&Bank{Name: "GCB Bank", Description: "One of Ghana’s largest state-owned banks providing retail and corporate banking."},
			&Bank{Name: "Ecobank Ghana", Description: "A pan-African bank offering personal, corporate, and digital banking services."},
			&Bank{Name: "Stanbic Bank Ghana", Description: "Part of Standard Bank Group, providing full-service banking in Ghana."},
			&Bank{Name: "Access Bank Ghana", Description: "Private bank offering modern banking and digital services."},
			&Bank{Name: "UBA Ghana", Description: "Part of United Bank for Africa, offering retail and corporate banking in Ghana."},

			&Bank{Name: "MTN Mobile Money (MoMo)", Description: "Ghana’s leading mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "AirtelTigo Money", Description: "Mobile wallet service for payments, top-ups, and transfers in Ghana."},
			&Bank{Name: "Zeepay Wallet", Description: "Digital wallet for remittances, payments, and QR-based transactions in Ghana."},
		)

	case "MAR": // Morocco
		banks = append(banks,
			&Bank{Name: "Attijariwafa Bank", Description: "Morocco’s largest bank offering retail, corporate, and international banking services."},
			&Bank{Name: "Banque Populaire", Description: "A major Moroccan bank providing personal and business banking solutions."},
			&Bank{Name: "BMCE Bank", Description: "One of Morocco’s top banks offering retail and corporate banking services."},
			&Bank{Name: "Société Générale Maroc", Description: "A branch of Société Générale providing full banking services in Morocco."},
			&Bank{Name: "Crédit du Maroc", Description: "Major bank offering personal, business, and investment banking services."},

			&Bank{Name: "M-Wallet Maroc", Description: "Digital wallet for mobile payments, transfers, and bill payments in Morocco."},
			&Bank{Name: "Inwi Money", Description: "Mobile wallet solution for payments and transfers in Morocco."},
			&Bank{Name: "Orange Money Morocco", Description: "Mobile wallet service for sending money, paying bills, and online payments."},
		)

	case "TUN": // Tunisia
		banks = append(banks,
			&Bank{Name: "Banque de Tunisie", Description: "One of Tunisia’s oldest banks providing retail and corporate banking services."},
			&Bank{Name: "Banque Internationale Arabe de Tunisie (BIAT)", Description: "The largest private bank in Tunisia offering full banking services."},
			&Bank{Name: "Attijari Bank Tunisia", Description: "A leading bank providing personal, corporate, and digital banking solutions."},
			&Bank{Name: "Société Tunisienne de Banque (STB)", Description: "State-owned bank offering nationwide retail and corporate banking."},
			&Bank{Name: "Amen Bank", Description: "Private Tunisian bank offering retail and business banking services."},

			&Bank{Name: "eDinar", Description: "Digital wallet in Tunisia for mobile payments, transfers, and bills."},
			&Bank{Name: "D17 Wallet", Description: "A mobile wallet widely used in Tunisia for online and in-store payments."},
			&Bank{Name: "Orange Money Tunisia", Description: "Mobile wallet solution for payments and money transfers."},
		)

	case "ETH": // Ethiopia
		banks = append(banks,
			&Bank{Name: "Commercial Bank of Ethiopia (CBE)", Description: "Ethiopia’s largest state-owned bank providing retail, corporate, and trade banking services."},
			&Bank{Name: "Dashen Bank", Description: "A leading private bank in Ethiopia offering retail, corporate, and digital banking services."},
			&Bank{Name: "Awash Bank", Description: "A major private bank providing personal, SME, and corporate banking in Ethiopia."},
			&Bank{Name: "Bank of Abyssinia", Description: "Private bank offering modern banking and digital services across Ethiopia."},
			&Bank{Name: "NIB International Bank", Description: "A growing bank providing commercial and personal banking solutions in Ethiopia."},

			&Bank{Name: "HelloCash", Description: "Ethiopia’s most popular mobile wallet for payments, transfers, and bill payments."},
			&Bank{Name: "M-BIRR", Description: "Mobile money service for payments, remittances, and transfers in Ethiopia."},
			&Bank{Name: "Amole", Description: "Digital wallet offered by Commercial Bank of Ethiopia for online payments and transfers."},
		)

	case "DZA": // Algeria
		banks = append(banks,
			&Bank{Name: "Banque Nationale d’Algérie (BNA)", Description: "State-owned bank offering retail and corporate banking services across Algeria."},
			&Bank{Name: "Banque Extérieure d’Algérie (BEA)", Description: "Major Algerian bank providing personal, business, and international banking services."},
			&Bank{Name: "Banque de l’Agriculture et du Développement Rural (BADR)", Description: "Bank focused on agriculture and rural development in Algeria."},
			&Bank{Name: "Société Générale Algérie", Description: "Branch of Société Générale providing retail and corporate banking in Algeria."},
			&Bank{Name: "CNEP Banque", Description: "Algerian bank offering personal and business banking solutions nationwide."},

			&Bank{Name: "BaridiMob", Description: "Mobile wallet service by Algeria Post for payments, transfers, and bills."},
			&Bank{Name: "E-Dinar Algeria", Description: "Digital wallet platform for online payments and QR transactions in Algeria."},
			&Bank{Name: "Mobilis Wallet", Description: "Telecom-based mobile wallet widely used for payments and transfers in Algeria."},
		)

	case "UKR": // Ukraine
		banks = append(banks,
			&Bank{Name: "PrivatBank", Description: "Ukraine’s largest bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "Oschadbank", Description: "State-owned bank providing personal and business banking across Ukraine."},
			&Bank{Name: "Raiffeisen Bank Aval", Description: "A major international bank providing retail and corporate banking services in Ukraine."},
			&Bank{Name: "Ukrsibbank", Description: "A top Ukrainian bank offering full banking services and digital solutions."},
			&Bank{Name: "Alfa-Bank Ukraine", Description: "Private bank providing personal, corporate, and online banking in Ukraine."},

			&Bank{Name: "Privat24", Description: "Digital wallet and mobile banking app by PrivatBank for payments, transfers, and QR payments."},
			&Bank{Name: "Monobank", Description: "Ukraine’s first fully digital bank providing mobile wallet and banking services."},
			&Bank{Name: "Portmone", Description: "Digital wallet used for online payments, bills, and transfers in Ukraine."},
		)

	case "ROU": // Romania
		banks = append(banks,
			&Bank{Name: "Banca Transilvania", Description: "Romania’s largest bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "BRD – Groupe Société Générale", Description: "A major bank providing full-service banking in Romania."},
			&Bank{Name: "BCR (Banca Comercială Română)", Description: "A leading Romanian bank offering retail, corporate, and investment services."},
			&Bank{Name: "Raiffeisen Bank Romania", Description: "International bank providing retail and corporate banking solutions in Romania."},
			&Bank{Name: "ING Bank Romania", Description: "Digital and traditional banking services provided by ING in Romania."},

			&Bank{Name: "Revolut Romania", Description: "Digital wallet and banking app offering payments and transfers in Romania."},
			&Bank{Name: "Orange Money Romania", Description: "Mobile wallet for payments, QR transactions, and transfers."},
			&Bank{Name: "PayU Wallet", Description: "Digital payment platform for online purchases and mobile payments in Romania."},
		)

	case "BGR": // Bulgaria
		banks = append(banks,
			&Bank{Name: "UniCredit Bulbank", Description: "The largest bank in Bulgaria providing retail, corporate, and investment banking services."},
			&Bank{Name: "DSK Bank", Description: "Major Bulgarian bank offering personal, SME, and corporate banking."},
			&Bank{Name: "First Investment Bank (Fibank)", Description: "Private bank providing retail and business banking services in Bulgaria."},
			&Bank{Name: "Raiffeisenbank Bulgaria", Description: "International bank offering full-service banking in Bulgaria."},
			&Bank{Name: "Postbank (Eurobank Bulgaria)", Description: "A major bank in Bulgaria providing retail and corporate banking solutions."},

			&Bank{Name: "Pay by Vivacom", Description: "Digital wallet and mobile payment solution in Bulgaria."},
			&Bank{Name: "ePay.bg", Description: "Widely used electronic wallet for payments, transfers, and online transactions."},
			&Bank{Name: "Revolut Bulgaria", Description: "Digital wallet app offering payments, transfers, and multi-currency support."},
		)

	case "SRB": // Serbia
		banks = append(banks,
			&Bank{Name: "Banca Intesa Beograd", Description: "One of Serbia’s largest banks offering retail, corporate, and digital banking services."},
			&Bank{Name: "Komercijalna Banka", Description: "Leading Serbian bank providing personal, SME, and corporate banking."},
			&Bank{Name: "UniCredit Bank Serbia", Description: "International bank offering retail, corporate, and investment banking solutions in Serbia."},
			&Bank{Name: "Raiffeisen Bank Serbia", Description: "Part of the Raiffeisen Group, providing full banking services in Serbia."},
			&Bank{Name: "OTP Bank Serbia", Description: "A private bank offering retail, corporate, and digital banking in Serbia."},

			&Bank{Name: "mCash Serbia", Description: "A mobile wallet and payment platform widely used in Serbia."},
			&Bank{Name: "PayPal", Description: "Global digital wallet supporting online payments for Serbian users."},
			&Bank{Name: "Revolut", Description: "Digital bank and wallet offering international banking and money transfers."},
		)

	case "ISL": // Iceland
		banks = append(banks,
			&Bank{Name: "Landsbankinn", Description: "One of Iceland’s largest banks offering retail, corporate, and digital banking."},
			&Bank{Name: "Arion Bank", Description: "A major Icelandic bank providing personal and business banking solutions."},
			&Bank{Name: "Íslandsbanki", Description: "Leading bank in Iceland offering full banking services and online banking."},
			&Bank{Name: "Kvika Bank", Description: "A bank specializing in investment, corporate, and digital banking in Iceland."},

			&Bank{Name: "Valitor", Description: "Icelandic payment solutions provider and digital wallet for online and in-store payments."},
			&Bank{Name: "Revolut", Description: "Digital bank and wallet offering international banking and money transfers for Icelandic users."},
			&Bank{Name: "Apple Pay", Description: "Mobile payment and digital wallet available in Iceland for iOS users."},
		)

	case "BLR": // Belarus
		banks = append(banks,
			&Bank{Name: "Belinvestbank", Description: "One of Belarus’s largest banks providing retail, corporate, and investment banking services."},
			&Bank{Name: "Belarusbank", Description: "State-owned bank offering full banking services across Belarus."},
			&Bank{Name: "Priorbank", Description: "A leading private bank in Belarus offering retail, corporate, and digital banking."},
			&Bank{Name: "BPS-Sberbank", Description: "Belarusian branch of Sberbank providing financial services nationwide."},

			&Bank{Name: "Yandex Money (YooMoney)", Description: "Digital wallet and payment platform accessible in Belarus."},
			&Bank{Name: "WebMoney", Description: "International digital payment system and wallet widely used in Belarus."},
			&Bank{Name: "Revolut", Description: "Digital banking and wallet service offering international transfers for Belarusian users."},
		)

	case "FJI": // Fiji
		banks = append(banks,
			&Bank{Name: "Bank of Fiji", Description: "The central bank of Fiji providing regulatory and limited banking services."},
			&Bank{Name: "ANZ Fiji", Description: "Part of ANZ Group offering personal, corporate, and digital banking services in Fiji."},
			&Bank{Name: "Westpac Fiji", Description: "A major bank providing retail, business, and online banking across Fiji."},
			&Bank{Name: "BSP Fiji", Description: "Bank of South Pacific branch offering full-service banking in Fiji."},

			&Bank{Name: "Fijipay", Description: "Local mobile wallet for payments, transfers, and bills in Fiji."},
			&Bank{Name: "PayPal", Description: "Global digital wallet for online payments accessible in Fiji."},
			&Bank{Name: "Revolut", Description: "Digital bank and wallet for international money transfers and payments."},
		)

	case "PNG": // Papua New Guinea
		banks = append(banks,
			&Bank{Name: "Bank of Papua New Guinea", Description: "Central bank of Papua New Guinea providing regulatory and limited banking services."},
			&Bank{Name: "ANZ PNG", Description: "Branch of ANZ Group offering retail, corporate, and digital banking in Papua New Guinea."},
			&Bank{Name: "Westpac PNG", Description: "A major bank providing personal and business banking services in PNG."},
			&Bank{Name: "Bank South Pacific (BSP)", Description: "Largest bank in PNG offering comprehensive banking and digital services."},

			&Bank{Name: "BSP Mobile Banking", Description: "Digital wallet and banking app offered by Bank South Pacific."},
			&Bank{Name: "PayPal", Description: "Global online payment platform accessible in Papua New Guinea."},
			&Bank{Name: "Revolut", Description: "Digital banking and wallet service offering international transfers and payments for PNG users."},
		)

	case "JAM": // Jamaica
		banks = append(banks,
			&Bank{Name: "National Commercial Bank (NCB)", Description: "Jamaica’s largest bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "Scotiabank Jamaica", Description: "A major international bank providing full-service banking and digital solutions in Jamaica."},
			&Bank{Name: "First Global Bank", Description: "Private bank offering retail, corporate, and wealth management services."},
			&Bank{Name: "Bank of Nova Scotia Jamaica", Description: "International bank branch offering personal, business, and online banking services."},

			&Bank{Name: "JMMB Money", Description: "Digital wallet and mobile payment service available in Jamaica."},
			&Bank{Name: "PayPal", Description: "Global digital wallet accessible for online payments in Jamaica."},
			&Bank{Name: "Revolut", Description: "Digital banking and wallet service offering international transfers and payments."},
		)

	case "CRI": // Costa Rica
		banks = append(banks,
			&Bank{Name: "Banco Nacional de Costa Rica", Description: "State-owned bank providing retail, corporate, and international banking services."},
			&Bank{Name: "Banco de Costa Rica", Description: "One of Costa Rica’s oldest banks offering full banking services."},
			&Bank{Name: "Scotiabank Costa Rica", Description: "International bank providing retail, corporate, and digital banking in Costa Rica."},
			&Bank{Name: "BAC Credomatic", Description: "Central American bank offering personal, business, and online banking solutions."},

			&Bank{Name: "SINPE Móvil", Description: "Costa Rica’s mobile wallet platform for payments, transfers, and bill payments."},
			&Bank{Name: "PayPal", Description: "Global digital wallet for online payments accessible in Costa Rica."},
			&Bank{Name: "Revolut", Description: "Digital banking and wallet service offering international transfers and payments."},
		)

	case "GTM": // Guatemala
		banks = append(banks,
			&Bank{Name: "Banco Industrial", Description: "Guatemala’s largest private bank offering retail, corporate, and digital banking services."},
			&Bank{Name: "Banco G&T Continental", Description: "Major Guatemalan bank providing full-service banking solutions."},
			&Bank{Name: "Banrural", Description: "A top bank in Guatemala offering personal, corporate, and rural banking services."},
			&Bank{Name: "BAC Credomatic Guatemala", Description: "Central American bank providing retail and corporate banking and digital services."},

			&Bank{Name: "Tigo Money", Description: "Guatemala’s mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "PayPal", Description: "Global online payment platform accessible in Guatemala."},
			&Bank{Name: "Revolut", Description: "Digital banking and wallet service for international transfers and payments."},
		)

	case "IMF": // Special Drawing Rights (IMF)
		banks = append(banks,
			&Bank{Name: "International Monetary Fund (IMF)", Description: "An international organization that issues SDRs and provides financial support and policy advice globally."},
			&Bank{Name: "World Bank", Description: "Provides global financial support and development assistance; participates in SDR allocations."},

			&Bank{Name: "No direct e-wallets", Description: "SDRs are a reserve asset, not used in consumer e-wallets; they are managed through IMF accounts."},
		)

	case "KWT": // Kuwait
		banks = append(banks,
			&Bank{Name: "National Bank of Kuwait (NBK)", Description: "Kuwait’s largest bank offering retail, corporate, and international banking services."},
			&Bank{Name: "Gulf Bank", Description: "Major Kuwaiti bank providing personal, business, and online banking solutions."},
			&Bank{Name: "Kuwait Finance House (KFH)", Description: "Islamic bank offering retail and corporate banking services in Kuwait."},
			&Bank{Name: "Boubyan Bank", Description: "Islamic bank providing personal, SME, and corporate banking solutions in Kuwait."},
			&Bank{Name: "Commercial Bank of Kuwait (CBK)", Description: "Private bank offering full-service banking and digital solutions."},

			&Bank{Name: "K-Net Wallet", Description: "Kuwait’s mobile and digital wallet for payments, transfers, and bills."},
			&Bank{Name: "Google Pay", Description: "Mobile wallet and payment service accessible in Kuwait for online and in-store payments."},
			&Bank{Name: "Apple Pay", Description: "Mobile payment and digital wallet service available in Kuwait for iOS users."},
			&Bank{Name: "PayPal", Description: "Global digital wallet for international and online payments."},
		)

	case "QAT": // Qatar
		banks = append(banks,
			&Bank{Name: "Qatar National Bank (QNB)", Description: "Qatar’s largest bank offering retail, corporate, and international banking services."},
			&Bank{Name: "Doha Bank", Description: "Major Qatari bank providing personal, business, and digital banking solutions."},
			&Bank{Name: "Commercial Bank of Qatar", Description: "Private bank offering full-service banking and online banking in Qatar."},
			&Bank{Name: "Masraf Al Rayan", Description: "Islamic bank providing Sharia-compliant retail and corporate banking services in Qatar."},
			&Bank{Name: "Qatar Islamic Bank (QIB)", Description: "Leading Islamic bank offering personal, corporate, and investment banking in Qatar."},

			&Bank{Name: "Ooredoo Money", Description: "Mobile wallet and digital payment solution widely used in Qatar."},
			&Bank{Name: "QPay", Description: "Qatar-based digital wallet for payments, transfers, and online transactions."},
			&Bank{Name: "Google Pay", Description: "Mobile wallet and payment service available in Qatar."},
			&Bank{Name: "Apple Pay", Description: "Digital wallet service accessible for iOS users in Qatar."},
		)

	case "OMN": // Oman
		banks = append(banks,
			&Bank{Name: "Bank Muscat", Description: "Oman’s largest bank offering retail, corporate, and investment banking services."},
			&Bank{Name: "National Bank of Oman (NBO)", Description: "Major Omani bank providing full-service banking and digital solutions."},
			&Bank{Name: "Bank Dhofar", Description: "Private bank offering personal, corporate, and online banking services in Oman."},
			&Bank{Name: "HSBC Oman", Description: "International bank providing retail and corporate banking solutions in Oman."},

			&Bank{Name: "OmanPay", Description: "National digital payment and e-wallet platform for mobile and online transactions."},
			&Bank{Name: "Google Pay", Description: "Mobile wallet and online payment solution available in Oman."},
			&Bank{Name: "Apple Pay", Description: "Digital wallet for iOS users in Oman for payments and transfers."},
		)

	case "BHR": // Bahrain
		banks = append(banks,
			&Bank{Name: "National Bank of Bahrain (NBB)", Description: "One of Bahrain’s largest banks offering retail, corporate, and investment banking services."},
			&Bank{Name: "Ahli United Bank", Description: "Leading Bahraini bank providing personal, corporate, and digital banking services."},
			&Bank{Name: "Bahrain Islamic Bank", Description: "Islamic bank offering Sharia-compliant personal and business banking in Bahrain."},
			&Bank{Name: "Gulf International Bank (GIB)", Description: "Major regional bank providing corporate and investment banking solutions."},

			&Bank{Name: "BenefitPay", Description: "Bahrain’s national mobile wallet for payments, transfers, and bills."},
			&Bank{Name: "Google Pay", Description: "Mobile wallet and payment platform available in Bahrain."},
			&Bank{Name: "Apple Pay", Description: "Digital wallet for iOS users in Bahrain for online and in-store payments."},
		)

	case "JOR": // Jordan
		banks = append(banks,
			&Bank{Name: "Arab Bank", Description: "Jordan’s largest bank and one of the biggest in the Middle East, providing retail, corporate, and digital banking services."},
			&Bank{Name: "Bank of Jordan", Description: "Major private bank offering personal, business, and online banking solutions."},
			&Bank{Name: "Jordan Ahli Bank", Description: "Private bank providing retail, corporate, and investment banking services."},
			&Bank{Name: "Cairo Amman Bank", Description: "Bank offering personal, SME, and corporate banking solutions in Jordan."},

			&Bank{Name: "eFAWATEERcom", Description: "Jordanian digital payment and bill payment platform widely used in the country."},
			&Bank{Name: "Google Pay", Description: "Mobile wallet and payment service accessible in Jordan."},
			&Bank{Name: "Apple Pay", Description: "Digital wallet for iOS users in Jordan for payments and transfers."},
		)

	case "KAZ": // Kazakhstan
		banks = append(banks,
			&Bank{Name: "Halyk Bank", Description: "Largest bank in Kazakhstan providing retail, corporate, and investment banking services."},
			&Bank{Name: "Kazkommertsbank (KKB)", Description: "Major bank offering personal, business, and online banking solutions."},
			&Bank{Name: "Sberbank Kazakhstan", Description: "Part of Sberbank Group, providing retail and corporate banking in Kazakhstan."},
			&Bank{Name: "ATF Bank", Description: "Private bank offering comprehensive banking and digital services across Kazakhstan."},

			&Bank{Name: "Kaspi.kz", Description: "Leading Kazakh digital wallet for payments, transfers, and online shopping."},
			&Bank{Name: "Halyk Wallet", Description: "Mobile wallet app by Halyk Bank for payments and transfers."},
			&Bank{Name: "Google Pay", Description: "Mobile payment solution available in Kazakhstan."},
			&Bank{Name: "Apple Pay", Description: "Digital wallet for iOS users in Kazakhstan."},
		)

	}

	finalBanks := &[]*Bank{}
	uniqueBanks := make(map[string]*Bank)

	for _, data := range banks {
		normalizedName := strings.ToLower(strings.TrimSpace(data.Name))
		normalizedName = strings.ReplaceAll(normalizedName, " ", "")
		normalizedName = strings.ReplaceAll(normalizedName, "-", "")
		normalizedName = strings.ReplaceAll(normalizedName, ".", "")
		normalizedName = strings.ReplaceAll(normalizedName, "'", "")
		normalizedName = strings.ReplaceAll(normalizedName, "(", "")
		normalizedName = strings.ReplaceAll(normalizedName, ")", "")

		uniqueBanks[normalizedName] = data
	}
	for _, bank := range uniqueBanks {
		*finalBanks = append(*finalBanks, bank)
		bank.OrganizationID = organizationID
		bank.BranchID = branchID
		bank.CreatedAt = now
		bank.UpdatedAt = now
		bank.CreatedByID = userID
		bank.UpdatedByID = userID
		if err := m.BankManager().CreateWithTx(context, tx, bank); err != nil {
			return eris.Wrapf(err, "failed to seed bank %s", bank.Name)
		}
	}

	return nil
}

func (m *Core) BankCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*Bank, error) {
	return m.BankManager().Find(context, &Bank{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
