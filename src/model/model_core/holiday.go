package model_core

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
	Holiday struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_holidays"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_holidays"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		EntryDate   time.Time `gorm:"not null"`
		Name        string    `gorm:"type:varchar(255)"`
		Description string    `gorm:"type:text"`
	}

	HolidayResponse struct {
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
		EntryDate      string                `json:"entry_date"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	HolidayRequest struct {
		EntryDate   time.Time `json:"entry_date" validate:"required"`
		Name        string    `json:"name" validate:"required,min=1,max=255"`
		Description string    `json:"description,omitempty"`
		CurrencyID  uuid.UUID `json:"currency_id" validate:"required"`
	}
	HoldayYearAvaiable struct {
		Year  int `json:"year"`
		Count int `json:"count"`
	}
)

func (m *ModelCore) Holiday() {
	m.Migration = append(m.Migration, &Holiday{})
	m.HolidayManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Holiday, HolidayResponse, HolidayRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *Holiday) *HolidayResponse {
			if data == nil {
				return nil
			}
			return &HolidayResponse{
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
				EntryDate:      data.EntryDate.Format(time.RFC3339),
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *Holiday) []string {
			return []string{
				"holiday.create",
				fmt.Sprintf("holiday.create.%s", data.ID),
				fmt.Sprintf("holiday.create.branch.%s", data.BranchID),
				fmt.Sprintf("holiday.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Holiday) []string {
			return []string{
				"holiday.update",
				fmt.Sprintf("holiday.update.%s", data.ID),
				fmt.Sprintf("holiday.update.branch.%s", data.BranchID),
				fmt.Sprintf("holiday.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Holiday) []string {
			return []string{
				"holiday.delete",
				fmt.Sprintf("holiday.delete.%s", data.ID),
				fmt.Sprintf("holiday.delete.branch.%s", data.BranchID),
				fmt.Sprintf("holiday.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) HolidaySeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	year := now.Year()

	currencies, err := m.CurrencyManager.List(context)
	if err != nil {
		return eris.Wrap(err, "failed to list currencies for holiday seeding")
	}
	if len(currencies) == 0 {
		return eris.New("no currencies found for holiday seeding")
	}

	for _, currency := range currencies {
		var holidays []*Holiday

		switch currency.CurrencyCode {
		case "USD": // United States
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 15, 0, 0, 0, 0, time.UTC), Name: "Martin Luther King Jr. Day", Description: "Birthday of Martin Luther King Jr."},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 19, 0, 0, 0, 0, time.UTC), Name: "Presidents' Day", Description: "Washington's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 27, 0, 0, 0, 0, time.UTC), Name: "Memorial Day", Description: "Honors military personnel who died in service"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 19, 0, 0, 0, 0, time.UTC), Name: "Juneteenth", Description: "Emancipation of enslaved African Americans"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 4, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "American independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 2, 0, 0, 0, 0, time.UTC), Name: "Labor Day", Description: "Celebrates American workers"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 14, 0, 0, 0, 0, time.UTC), Name: "Columbus Day", Description: "Christopher Columbus's arrival in America"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Veterans Day", Description: "Honors military veterans"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 28, 0, 0, 0, 0, time.UTC), Name: "Thanksgiving Day", Description: "National day of thanksgiving"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "EUR": // European Union (Germany as representative)
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 3, 0, 0, 0, 0, time.UTC), Name: "German Unity Day", Description: "Reunification of Germany"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
			}

		case "JPY": // Japan
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "Ganjitsu - First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 8, 0, 0, 0, 0, time.UTC), Name: "Coming of Age Day", Description: "Seijin no Hi - Celebration of adulthood"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "National Foundation Day", Description: "Kenkoku Kinen no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 23, 0, 0, 0, 0, time.UTC), Name: "Emperor's Birthday", Description: "Tenno Tanjobi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 20, 0, 0, 0, 0, time.UTC), Name: "Spring Equinox", Description: "Shunbun no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 29, 0, 0, 0, 0, time.UTC), Name: "Showa Day", Description: "Showa no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 3, 0, 0, 0, 0, time.UTC), Name: "Constitution Memorial Day", Description: "Kenpo Kinenbi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 4, 0, 0, 0, 0, time.UTC), Name: "Greenery Day", Description: "Midori no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 5, 0, 0, 0, 0, time.UTC), Name: "Children's Day", Description: "Kodomo no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 15, 0, 0, 0, 0, time.UTC), Name: "Marine Day", Description: "Umi no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 11, 0, 0, 0, 0, time.UTC), Name: "Mountain Day", Description: "Yama no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Respect for the Aged Day", Description: "Keiro no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 23, 0, 0, 0, 0, time.UTC), Name: "Autumn Equinox", Description: "Shubun no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 14, 0, 0, 0, 0, time.UTC), Name: "Sports Day", Description: "Taiiku no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 3, 0, 0, 0, 0, time.UTC), Name: "Culture Day", Description: "Bunka no Hi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 23, 0, 0, 0, 0, time.UTC), Name: "Labor Thanksgiving Day", Description: "Kinro Kansha no Hi"},
			}

		case "GBP": // United Kingdom
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 6, 0, 0, 0, 0, time.UTC), Name: "Early May Bank Holiday", Description: "First Monday in May"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 27, 0, 0, 0, 0, time.UTC), Name: "Spring Bank Holiday", Description: "Last Monday in May"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 26, 0, 0, 0, 0, time.UTC), Name: "Summer Bank Holiday", Description: "Last Monday in August"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "AUD": // Australia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 26, 0, 0, 0, 0, time.UTC), Name: "Australia Day", Description: "National day of Australia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 25, 0, 0, 0, 0, time.UTC), Name: "ANZAC Day", Description: "Remembers all Australians who served and died"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "CAD": // Canada
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 19, 0, 0, 0, 0, time.UTC), Name: "Family Day", Description: "Third Monday in February"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Victoria Day", Description: "Monday before May 25"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC), Name: "Canada Day", Description: "Anniversary of Confederation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 2, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "First Monday in September"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 30, 0, 0, 0, 0, time.UTC), Name: "National Day for Truth and Reconciliation", Description: "Honors Indigenous communities"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 14, 0, 0, 0, 0, time.UTC), Name: "Thanksgiving", Description: "Second Monday in October"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Remembrance Day", Description: "Honors military veterans"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "CHF": // Switzerland
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC), Name: "Berchtold's Day", Description: "Swiss holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 1, 0, 0, 0, 0, time.UTC), Name: "Swiss National Day", Description: "Foundation of Swiss Confederacy"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
			}

		case "CNY": // China
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year's Eve", Description: "Spring Festival Eve"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Spring Festival - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Spring Festival - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Spring Festival - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 4, 0, 0, 0, 0, time.UTC), Name: "Qingming Festival", Description: "Tomb-Sweeping Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 10, 0, 0, 0, 0, time.UTC), Name: "Dragon Boat Festival", Description: "Traditional Chinese festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 17, 0, 0, 0, 0, time.UTC), Name: "Mid-Autumn Festival", Description: "Moon Festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 1, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Founding of People's Republic of China"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 2, 0, 0, 0, 0, time.UTC), Name: "National Day Holiday", Description: "Second day of National Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 3, 0, 0, 0, 0, time.UTC), Name: "National Day Holiday", Description: "Third day of National Day"},
			}

		case "SEK": // Sweden
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 6, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Swedish National Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 21, 0, 0, 0, 0, time.UTC), Name: "Midsummer Eve", Description: "Swedish tradition"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 22, 0, 0, 0, 0, time.UTC), Name: "Midsummer Day", Description: "Swedish tradition"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 24, 0, 0, 0, 0, time.UTC), Name: "Christmas Eve", Description: "Day before Christmas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC), Name: "New Year's Eve", Description: "Last day of the year"},
			}

		case "NZD": // New Zealand
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC), Name: "Day after New Year's Day", Description: "Second day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 6, 0, 0, 0, 0, time.UTC), Name: "Waitangi Day", Description: "National day of New Zealand"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 25, 0, 0, 0, 0, time.UTC), Name: "ANZAC Day", Description: "Remembers all New Zealanders who served and died"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 3, 0, 0, 0, 0, time.UTC), Name: "King's Birthday", Description: "Official birthday of the monarch"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 28, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "Celebrates workers' rights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "PHP": // Philippines
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 25, 0, 0, 0, 0, time.UTC), Name: "EDSA People Power Revolution Anniversary", Description: "Commemorates the peaceful revolution in 1986"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 9, 0, 0, 0, 0, time.UTC), Name: "Araw ng Kagitingan", Description: "Day of Valor - Bataan and Corregidor Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 30, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labor Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Philippine independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 21, 0, 0, 0, 0, time.UTC), Name: "Ninoy Aquino Day", Description: "Commemorates assassination of Benigno Aquino Jr."},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 26, 0, 0, 0, 0, time.UTC), Name: "National Heroes Day", Description: "Honors Filipino heroes"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 30, 0, 0, 0, 0, time.UTC), Name: "Bonifacio Day", Description: "Birthday of Andres Bonifacio"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 30, 0, 0, 0, 0, time.UTC), Name: "Rizal Day", Description: "Commemorates national hero José Rizal"},
			}

		case "INR": // India
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 26, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "Commemorates adoption of the Constitution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 13, 0, 0, 0, 0, time.UTC), Name: "Holi", Description: "Festival of colors"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Ram Navami", Description: "Birth of Lord Rama"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Independence from British rule"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 19, 0, 0, 0, 0, time.UTC), Name: "Janmashtami", Description: "Birth of Lord Krishna"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 2, 0, 0, 0, 0, time.UTC), Name: "Gandhi Jayanti", Description: "Birthday of Mahatma Gandhi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 24, 0, 0, 0, 0, time.UTC), Name: "Dussehra", Description: "Victory of good over evil"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 12, 0, 0, 0, 0, time.UTC), Name: "Diwali", Description: "Festival of lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "KRW": // South Korea
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Seollal", Description: "Korean New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Seollal", Description: "Korean New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Seollal", Description: "Korean New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 1, 0, 0, 0, 0, time.UTC), Name: "Independence Movement Day", Description: "Commemorates March 1st Movement in 1919"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 5, 0, 0, 0, 0, time.UTC), Name: "Children's Day", Description: "Celebrates children"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Buddha's Birthday", Description: "Celebrates birth of Buddha"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 6, 0, 0, 0, 0, time.UTC), Name: "Memorial Day", Description: "Honors those who died for the country"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Liberation Day", Description: "Liberation from Japanese rule"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Chuseok", Description: "Korean harvest festival - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 17, 0, 0, 0, 0, time.UTC), Name: "Chuseok", Description: "Korean harvest festival - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 18, 0, 0, 0, 0, time.UTC), Name: "Chuseok", Description: "Korean harvest festival - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 3, 0, 0, 0, 0, time.UTC), Name: "National Foundation Day", Description: "Founding of Korea"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 9, 0, 0, 0, 0, time.UTC), Name: "Hangeul Day", Description: "Korean alphabet creation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "THB": // Thailand
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 26, 0, 0, 0, 0, time.UTC), Name: "Makha Bucha Day", Description: "Buddhist holy day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 6, 0, 0, 0, 0, time.UTC), Name: "Chakri Day", Description: "Founding of Chakri Dynasty"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Songkran Festival", Description: "Traditional Thai New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Songkran Festival", Description: "Traditional Thai New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 15, 0, 0, 0, 0, time.UTC), Name: "Songkran Festival", Description: "Traditional Thai New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labor Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 4, 0, 0, 0, 0, time.UTC), Name: "Coronation Day", Description: "King Rama X's coronation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 24, 0, 0, 0, 0, time.UTC), Name: "Visakha Bucha Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 28, 0, 0, 0, 0, time.UTC), Name: "King's Birthday", Description: "Birthday of King Rama X"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 12, 0, 0, 0, 0, time.UTC), Name: "Queen Mother's Birthday", Description: "Birthday of Queen Sirikit"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 13, 0, 0, 0, 0, time.UTC), Name: "King Bhumibol Memorial Day", Description: "Death of King Rama IX"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 23, 0, 0, 0, 0, time.UTC), Name: "Chulalongkorn Day", Description: "Death of King Rama V"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 5, 0, 0, 0, 0, time.UTC), Name: "King Bhumibol's Birthday", Description: "Birthday of King Rama IX"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 10, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "First constitution of Thailand"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC), Name: "New Year's Eve", Description: "Last day of the year"},
			}

		case "SGD": // Singapore
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Vesak Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 9, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Independence of Singapore"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Deepavali", Description: "Hindu festival of lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "HKD": // Hong Kong
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year's Day", Description: "Lunar New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Second day of Chinese New Year", Description: "Lunar New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Third day of Chinese New Year", Description: "Lunar New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 4, 0, 0, 0, 0, time.UTC), Name: "Ching Ming Festival", Description: "Tomb-sweeping day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Buddha's Birthday", Description: "Birth of Buddha"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 10, 0, 0, 0, 0, time.UTC), Name: "Dragon Boat Festival", Description: "Traditional Chinese festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC), Name: "HKSAR Establishment Day", Description: "Hong Kong Special Administrative Region Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 17, 0, 0, 0, 0, time.UTC), Name: "Mid-Autumn Festival", Description: "Moon festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 1, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "People's Republic of China National Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 23, 0, 0, 0, 0, time.UTC), Name: "Chung Yeung Festival", Description: "Double Ninth Festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "MYR": // Malaysia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 1, 0, 0, 0, 0, time.UTC), Name: "Federal Territory Day", Description: "Formation of Federal Territory"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Wesak Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 3, 0, 0, 0, 0, time.UTC), Name: "Yang di-Pertuan Agong's Birthday", Description: "King's Official Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 31, 0, 0, 0, 0, time.UTC), Name: "Merdeka Day", Description: "Independence Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Malaysia Day", Description: "Formation of Malaysia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Deepavali", Description: "Hindu festival of lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "IDR": // Indonesia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 11, 0, 0, 0, 0, time.UTC), Name: "Nyepi", Description: "Balinese Day of Silence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension of Jesus Christ", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Vesak Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 1, 0, 0, 0, 0, time.UTC), Name: "Pancasila Day", Description: "State philosophy of Indonesia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 17, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Independence from Netherlands"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "VND": // Vietnam
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 9, 0, 0, 0, 0, time.UTC), Name: "Tet Holiday", Description: "Vietnamese New Year - Eve"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Tet Holiday", Description: "Vietnamese New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Tet Holiday", Description: "Vietnamese New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Tet Holiday", Description: "Vietnamese New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 18, 0, 0, 0, 0, time.UTC), Name: "Hung Kings Festival", Description: "Commemorates the Hung Kings"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 30, 0, 0, 0, 0, time.UTC), Name: "Liberation Day", Description: "Fall of Saigon"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "International Labor Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 2, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Independence from France"},
			}

		case "TWD": // Taiwan
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 9, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year's Eve", Description: "Lunar New Year's Eve"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Spring Festival", Description: "Chinese New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Spring Festival", Description: "Chinese New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Spring Festival", Description: "Chinese New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 28, 0, 0, 0, 0, time.UTC), Name: "Peace Memorial Day", Description: "228 Incident"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 4, 0, 0, 0, 0, time.UTC), Name: "Tomb Sweeping Day", Description: "Qingming Festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 10, 0, 0, 0, 0, time.UTC), Name: "Dragon Boat Festival", Description: "Traditional Chinese festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 17, 0, 0, 0, 0, time.UTC), Name: "Mid-Autumn Festival", Description: "Moon festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 10, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Double Ten Day - Republic of China"},
			}

		case "BND": // Brunei
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 23, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Birthday of the Prophet Muhammad", Description: "Islamic holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 31, 0, 0, 0, 0, time.UTC), Name: "Royal Brunei Armed Forces Day", Description: "Military commemoration"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 15, 0, 0, 0, 0, time.UTC), Name: "His Majesty's Birthday", Description: "Sultan's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "SAR": // Saudi Arabia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 22, 0, 0, 0, 0, time.UTC), Name: "Founding Day", Description: "Founding of the First Saudi State"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 12, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 18, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 23, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Unification of Saudi Arabia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Prophet Muhammad's Birthday", Description: "Mawlid al-Nabi"},
			}

		case "AED": // United Arab Emirates
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Lailat al Miraj", Description: "Night Journey of Prophet Muhammad"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 12, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 15, 0, 0, 0, 0, time.UTC), Name: "Arafat Day", Description: "Day of Arafat"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 18, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Prophet Muhammad's Birthday", Description: "Mawlid al-Nabi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 30, 0, 0, 0, 0, time.UTC), Name: "Commemoration Day", Description: "Martyrs' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 2, 0, 0, 0, 0, time.UTC), Name: "UAE National Day", Description: "Formation of the UAE"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 3, 0, 0, 0, 0, time.UTC), Name: "UAE National Day", Description: "Second day of National Day"},
			}

		case "ILS": // Israel
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Passover", Description: "Pesach - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 19, 0, 0, 0, 0, time.UTC), Name: "Passover", Description: "Pesach - Last day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 26, 0, 0, 0, 0, time.UTC), Name: "Holocaust Remembrance Day", Description: "Yom HaShoah"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 4, 0, 0, 0, 0, time.UTC), Name: "Memorial Day", Description: "Yom HaZikaron"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 5, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Yom HaAtzmaut"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 25, 0, 0, 0, 0, time.UTC), Name: "Lag BaOmer", Description: "Jewish holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 2, 0, 0, 0, 0, time.UTC), Name: "Shavuot", Description: "Pentecost - Festival of Weeks"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Rosh Hashanah", Description: "Jewish New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 17, 0, 0, 0, 0, time.UTC), Name: "Rosh Hashanah", Description: "Jewish New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 25, 0, 0, 0, 0, time.UTC), Name: "Yom Kippur", Description: "Day of Atonement"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 30, 0, 0, 0, 0, time.UTC), Name: "Sukkot", Description: "Festival of Tabernacles - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 7, 0, 0, 0, 0, time.UTC), Name: "Simchat Torah", Description: "Rejoicing with the Torah"},
			}

		case "ZAR": // South Africa
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 21, 0, 0, 0, 0, time.UTC), Name: "Human Rights Day", Description: "Commemorates Sharpeville massacre"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Family Day", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 27, 0, 0, 0, 0, time.UTC), Name: "Freedom Day", Description: "First democratic elections"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Workers' Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Youth Day", Description: "Soweto uprising"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 9, 0, 0, 0, 0, time.UTC), Name: "National Women's Day", Description: "Women's march to Union Buildings"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 24, 0, 0, 0, 0, time.UTC), Name: "Heritage Day", Description: "Celebrating South African culture"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 16, 0, 0, 0, 0, time.UTC), Name: "Day of Reconciliation", Description: "Promoting reconciliation and unity"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Day of Goodwill", Description: "Day after Christmas"},
			}

		case "EGP": // Egypt
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 25, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "January 25 Revolution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Lailat al Miraj", Description: "Night Journey of Prophet Muhammad"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 12, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 25, 0, 0, 0, 0, time.UTC), Name: "Sinai Liberation Day", Description: "Return of Sinai Peninsula"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 18, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 23, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "July 23 Revolution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Prophet Muhammad's Birthday", Description: "Mawlid al-Nabi"},
			}

		case "TRY": // Turkey
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "Ramazan Bayramı - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "Ramazan Bayramı - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 12, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "Ramazan Bayramı - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 23, 0, 0, 0, 0, time.UTC), Name: "National Sovereignty and Children's Day", Description: "Opening of Turkish Grand National Assembly"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour and Solidarity Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 19, 0, 0, 0, 0, time.UTC), Name: "Commemoration of Atatürk", Description: "Youth and Sports Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Kurban Bayramı - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Kurban Bayramı - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 18, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Kurban Bayramı - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 19, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Kurban Bayramı - Fourth day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 15, 0, 0, 0, 0, time.UTC), Name: "Democracy and National Unity Day", Description: "Commemorates July 15 coup attempt"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 30, 0, 0, 0, 0, time.UTC), Name: "Victory Day", Description: "Battle of Dumlupınar victory"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 29, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "Proclamation of the Republic"},
			}

		case "XOF": // West African CFA Franc
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 4, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Senegal Independence (representative)"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 25, 0, 0, 0, 0, time.UTC), Name: "Africa Day", Description: "Organization of African Unity"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "XAF": // Central African CFA Franc
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Cameroon Independence (representative)"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Cameroon National Day (representative)"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 25, 0, 0, 0, 0, time.UTC), Name: "Africa Day", Description: "Organization of African Unity"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "MUR": // Mauritius
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC), Name: "New Year Holiday", Description: "Second day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 1, 0, 0, 0, 0, time.UTC), Name: "Abolition of Slavery", Description: "End of slavery in Mauritius"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Lunar New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 12, 0, 0, 0, 0, time.UTC), Name: "Independence & Republic Day", Description: "Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption of Mary", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Diwali", Description: "Hindu festival of lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "MVR": // Maldives
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 26, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Prophet Muhammad's Birthday", Description: "Mawlid al-Nabi"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 3, 0, 0, 0, 0, time.UTC), Name: "Victory Day", Description: "Failed coup attempt in 1988"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "Establishment of the Republic"},
			}

		case "NOK": // Norway
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 17, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Norwegian National Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 19, 0, 0, 0, 0, time.UTC), Name: "Whit Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "DKK": // Denmark
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 25, 0, 0, 0, 0, time.UTC), Name: "Great Prayer Day", Description: "Store Bededag - Danish holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 19, 0, 0, 0, 0, time.UTC), Name: "Whit Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 5, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Danish Constitution Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 24, 0, 0, 0, 0, time.UTC), Name: "Christmas Eve", Description: "Day before Christmas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "PLN": // Poland
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 3, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Constitution of 3 May 1791"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 19, 0, 0, 0, 0, time.UTC), Name: "Pentecost", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 29, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption of Mary", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Restoration of independence in 1918"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "CZK": // Czech Republic
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 8, 0, 0, 0, 0, time.UTC), Name: "Liberation Day", Description: "End of World War II in Europe"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 5, 0, 0, 0, 0, time.UTC), Name: "Saints Cyril and Methodius Day", Description: "Apostles of the Slavs"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 6, 0, 0, 0, 0, time.UTC), Name: "Jan Hus Day", Description: "Death of Jan Hus"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 28, 0, 0, 0, 0, time.UTC), Name: "Czech Statehood Day", Description: "Day of Czech Statehood"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 28, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Creation of Czechoslovakia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 17, 0, 0, 0, 0, time.UTC), Name: "Freedom and Democracy Day", Description: "Velvet Revolution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 24, 0, 0, 0, 0, time.UTC), Name: "Christmas Eve", Description: "Day before Christmas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
			}

		case "HUF": // Hungary
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 15, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "1848 Revolution and War of Independence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 19, 0, 0, 0, 0, time.UTC), Name: "Whit Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 20, 0, 0, 0, 0, time.UTC), Name: "Whit Monday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 20, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Foundation of the Hungarian state"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 23, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "1956 Revolution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "RUB": // Russia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC), Name: "New Year Holidays", Description: "Second day of New Year holidays"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 7, 0, 0, 0, 0, time.UTC), Name: "Orthodox Christmas", Description: "Russian Orthodox Christmas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 23, 0, 0, 0, 0, time.UTC), Name: "Defender of the Fatherland Day", Description: "Armed Forces Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 8, 0, 0, 0, 0, time.UTC), Name: "International Women's Day", Description: "Women's Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "Spring and Labour Holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Victory Day", Description: "Victory in World War II"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Russia Day", Description: "National Day of Russia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 4, 0, 0, 0, 0, time.UTC), Name: "Unity Day", Description: "Day of People's Unity"},
			}

		case "EUR-HR": // Croatia (Euro)
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 30, 0, 0, 0, 0, time.UTC), Name: "Statehood Day", Description: "Croatian Parliament Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 19, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 22, 0, 0, 0, 0, time.UTC), Name: "Anti-Fascist Struggle Day", Description: "Resistance movement in WWII"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 5, 0, 0, 0, 0, time.UTC), Name: "Victory and Homeland Thanksgiving Day", Description: "Operation Storm"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption of Mary", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 8, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Croatian independence from Yugoslavia"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
			}

		case "BRL": // Brazil
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Carnival Monday", Description: "First day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Carnival Tuesday", Description: "Second day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 21, 0, 0, 0, 0, time.UTC), Name: "Tiradentes", Description: "Martyrdom of Tiradentes"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 30, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 7, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Brazilian Independence from Portugal"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Our Lady of Aparecida", Description: "Patron saint of Brazil"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 2, 0, 0, 0, 0, time.UTC), Name: "All Souls' Day", Description: "Day of the Dead"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 15, 0, 0, 0, 0, time.UTC), Name: "Proclamation of the Republic", Description: "End of the Empire of Brazil"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "MXN": // Mexico
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 5, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Mexican Constitution of 1917"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 21, 0, 0, 0, 0, time.UTC), Name: "Benito Juárez's Birthday", Description: "Birthday of Benito Juárez"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Mexican Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 2, 0, 0, 0, 0, time.UTC), Name: "Day of the Dead", Description: "Día de los Muertos"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 20, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "Mexican Revolution of 1910"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 12, 0, 0, 0, 0, time.UTC), Name: "Our Lady of Guadalupe", Description: "Patron saint of Mexico"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "ARS": // Argentina
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Carnival Monday", Description: "First day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Carnival Tuesday", Description: "Second day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 24, 0, 0, 0, 0, time.UTC), Name: "Day of Remembrance for Truth and Justice", Description: "Memory of disappeared during military dictatorship"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 2, 0, 0, 0, 0, time.UTC), Name: "Malvinas Day", Description: "Day of the Veterans and Fallen of the Malvinas War"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 25, 0, 0, 0, 0, time.UTC), Name: "May Revolution", Description: "First government assembly of 1810"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 20, 0, 0, 0, 0, time.UTC), Name: "Flag Day", Description: "Death of Manuel Belgrano"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 9, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Declaration of Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 17, 0, 0, 0, 0, time.UTC), Name: "San Martín Day", Description: "Death of José de San Martín"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Columbus Day", Description: "Día del Respeto a la Diversidad Cultural"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 20, 0, 0, 0, 0, time.UTC), Name: "National Sovereignty Day", Description: "Battle of Vuelta de Obligado"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Immaculate Conception", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "CLP": // Chile
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 30, 0, 0, 0, 0, time.UTC), Name: "Holy Saturday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 21, 0, 0, 0, 0, time.UTC), Name: "Navy Day", Description: "Naval Battle of Iquique"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC), Name: "Saints Peter and Paul", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 16, 0, 0, 0, 0, time.UTC), Name: "Our Lady of Mount Carmel", Description: "Patron saint of Chile"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Assumption of Mary", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 18, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Chilean Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 19, 0, 0, 0, 0, time.UTC), Name: "Army Day", Description: "Day of the Glories of the Army"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Columbus Day", Description: "Day of the Race"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Protestant Reformation Day", Description: "National Day of Evangelical Churches"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Immaculate Conception", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "COP": // Colombia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 8, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 25, 0, 0, 0, 0, time.UTC), Name: "Saint Joseph's Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 13, 0, 0, 0, 0, time.UTC), Name: "Ascension Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 3, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 10, 0, 0, 0, 0, time.UTC), Name: "Sacred Heart", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC), Name: "Saints Peter and Paul", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 20, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Colombian Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 7, 0, 0, 0, 0, time.UTC), Name: "Battle of Boyacá", Description: "Colombian Independence victory"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 19, 0, 0, 0, 0, time.UTC), Name: "Assumption of Mary", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 14, 0, 0, 0, 0, time.UTC), Name: "Columbus Day", Description: "Day of the Race"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 4, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Independence of Cartagena", Description: "Cartagena Independence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Immaculate Conception", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "PEN": // Peru
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 29, 0, 0, 0, 0, time.UTC), Name: "Saints Peter and Paul", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 28, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Peruvian Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 29, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Second day of Independence celebration"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 30, 0, 0, 0, 0, time.UTC), Name: "Santa Rosa de Lima", Description: "Patron saint of Peru and the Americas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 8, 0, 0, 0, 0, time.UTC), Name: "Battle of Angamos", Description: "Naval battle during War of the Pacific"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Immaculate Conception", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 9, 0, 0, 0, 0, time.UTC), Name: "Battle of Ayacucho", Description: "Final battle of Peruvian Independence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "UYU": // Uruguay
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Carnival Monday", Description: "First day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Carnival Tuesday", Description: "Second day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 19, 0, 0, 0, 0, time.UTC), Name: "Landing of the 33 Patriots", Description: "Beginning of independence movement"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 18, 0, 0, 0, 0, time.UTC), Name: "Battle of Las Piedras", Description: "First victory in independence war"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 19, 0, 0, 0, 0, time.UTC), Name: "Artigas' Birthday", Description: "Birthday of national hero José Artigas"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 18, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Uruguayan Constitution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 25, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Declaration of Independence from Brazil"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Columbus Day", Description: "Day of the Race"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 2, 0, 0, 0, 0, time.UTC), Name: "All Souls' Day", Description: "Day of the Dead"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "DOP": // Dominican Republic
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Epiphany", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 21, 0, 0, 0, 0, time.UTC), Name: "Our Lady of Altagracia", Description: "Patron saint of Dominican Republic"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 26, 0, 0, 0, 0, time.UTC), Name: "Juan Pablo Duarte Day", Description: "Birthday of founding father"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 27, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Dominican Independence from Haiti"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 30, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 16, 0, 0, 0, 0, time.UTC), Name: "Restoration Day", Description: "War of Restoration"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 24, 0, 0, 0, 0, time.UTC), Name: "Our Lady of Las Mercedes", Description: "Patron saint of Dominican Republic"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 6, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Dominican Constitution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "PYG": // Paraguay
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 1, 0, 0, 0, 0, time.UTC), Name: "Heroes' Day", Description: "Day of Heroes of the Fatherland"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 15, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Paraguayan Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Chaco Armistice", Description: "End of Chaco War"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "Founding of Asunción", Description: "Foundation of the capital city"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 29, 0, 0, 0, 0, time.UTC), Name: "Battle of Boquerón", Description: "Victory during Chaco War"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Immaculate Conception", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "BOB": // Bolivia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 22, 0, 0, 0, 0, time.UTC), Name: "Plurinational State Day", Description: "Foundation of the Plurinational State"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Carnival Monday", Description: "First day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Carnival Tuesday", Description: "Second day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 30, 0, 0, 0, 0, time.UTC), Name: "Corpus Christi", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 21, 0, 0, 0, 0, time.UTC), Name: "Andean New Year", Description: "Aymara New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 6, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Bolivian Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 2, 0, 0, 0, 0, time.UTC), Name: "All Souls' Day", Description: "Day of the Dead"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "VES": // Venezuela
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Carnival Monday", Description: "First day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 13, 0, 0, 0, 0, time.UTC), Name: "Carnival Tuesday", Description: "Second day of Carnival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 28, 0, 0, 0, 0, time.UTC), Name: "Maundy Thursday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 19, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Venezuelan Independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 24, 0, 0, 0, 0, time.UTC), Name: "Battle of Carabobo", Description: "Venezuelan Independence victory"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 5, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Venezuelan Independence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 24, 0, 0, 0, 0, time.UTC), Name: "Simón Bolívar's Birthday", Description: "Birthday of El Libertador"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Indigenous Resistance Day", Description: "Day of Indigenous Resistance"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 17, 0, 0, 0, 0, time.UTC), Name: "Death of Simón Bolívar", Description: "Death of El Libertador"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "PKR": // Pakistan
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 5, 0, 0, 0, 0, time.UTC), Name: "Kashmir Solidarity Day", Description: "Support for Kashmir"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 23, 0, 0, 0, 0, time.UTC), Name: "Pakistan Day", Description: "Pakistan Resolution Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Muharram", Description: "Islamic New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 14, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Pakistani Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Milad un-Nabi", Description: "Prophet Muhammad's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 9, 0, 0, 0, 0, time.UTC), Name: "Iqbal Day", Description: "Allama Iqbal's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Quaid-e-Azam's Birthday", Description: "Muhammad Ali Jinnah's Birthday"},
			}

		case "BDT": // Bangladesh
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 21, 0, 0, 0, 0, time.UTC), Name: "International Mother Language Day", Description: "Language Movement Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 17, 0, 0, 0, 0, time.UTC), Name: "Sheikh Mujibur Rahman's Birthday", Description: "Father of the Nation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 26, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Bangladeshi Independence from Pakistan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Pohela Boishakh", Description: "Bengali New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "May Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC), Name: "National Mourning Day", Description: "Sheikh Mujibur Rahman's assassination"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Milad un-Nabi", Description: "Prophet Muhammad's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 16, 0, 0, 0, 0, time.UTC), Name: "Victory Day", Description: "Liberation War victory"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "LKR": // Sri Lanka
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 4, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Sri Lankan Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Sinhala and Tamil New Year", Description: "Traditional New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Sinhala and Tamil New Year", Description: "Traditional New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "May Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 23, 0, 0, 0, 0, time.UTC), Name: "Vesak Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 5, 0, 0, 0, 0, time.UTC), Name: "Esala Perahera", Description: "Buddhist festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Deepavali", Description: "Hindu Festival of Lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "NPR": // Nepal
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 11, 0, 0, 0, 0, time.UTC), Name: "Prithvi Jayanti", Description: "National Unity Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 19, 0, 0, 0, 0, time.UTC), Name: "Democracy Day", Description: "End of autocracy"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 8, 0, 0, 0, 0, time.UTC), Name: "International Women's Day", Description: "Women's rights and achievements"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Nepali New Year", Description: "Bikram Sambat New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 26, 0, 0, 0, 0, time.UTC), Name: "Buddha Jayanti", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 28, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "End of monarchy"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 19, 0, 0, 0, 0, time.UTC), Name: "Janai Purnima", Description: "Sacred thread festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 20, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Adoption of new constitution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Tihar", Description: "Festival of Lights - Deepawali"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "MMK": // Myanmar
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Myanmar Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 12, 0, 0, 0, 0, time.UTC), Name: "Union Day", Description: "Panglong Agreement"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 2, 0, 0, 0, 0, time.UTC), Name: "Peasants' Day", Description: "Agricultural workers' day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 27, 0, 0, 0, 0, time.UTC), Name: "Armed Forces Day", Description: "Resistance against Japanese occupation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Thingyan", Description: "Water Festival - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Thingyan", Description: "Water Festival - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 15, 0, 0, 0, 0, time.UTC), Name: "Thingyan", Description: "Water Festival - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 16, 0, 0, 0, 0, time.UTC), Name: "Myanmar New Year", Description: "First day of Myanmar calendar"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 26, 0, 0, 0, 0, time.UTC), Name: "Buddha Day", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 19, 0, 0, 0, 0, time.UTC), Name: "Martyrs' Day", Description: "Assassination of General Aung San"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Deepavali", Description: "Festival of Lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		case "KHR": // Cambodia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 7, 0, 0, 0, 0, time.UTC), Name: "Victory Day", Description: "Victory over Khmer Rouge"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 8, 0, 0, 0, 0, time.UTC), Name: "International Women's Day", Description: "Women's rights and achievements"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Khmer New Year", Description: "Traditional New Year - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Khmer New Year", Description: "Traditional New Year - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 15, 0, 0, 0, 0, time.UTC), Name: "Khmer New Year", Description: "Traditional New Year - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 14, 0, 0, 0, 0, time.UTC), Name: "King's Birthday", Description: "Birthday of King Norodom Sihamoni"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 26, 0, 0, 0, 0, time.UTC), Name: "Visak Bochea", Description: "Buddha's birth, enlightenment, and death"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 24, 0, 0, 0, 0, time.UTC), Name: "Constitution Day", Description: "Adoption of Constitution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "King Father's Birthday", Description: "Birthday of former King Norodom Sihanouk"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 9, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Cambodian Independence from France"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 14, 0, 0, 0, 0, time.UTC), Name: "Water Festival", Description: "Bon Om Touk - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 15, 0, 0, 0, 0, time.UTC), Name: "Water Festival", Description: "Bon Om Touk - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 16, 0, 0, 0, 0, time.UTC), Name: "Water Festival", Description: "Bon Om Touk - Third day"},
			}

		case "LAK": // Laos
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 20, 0, 0, 0, 0, time.UTC), Name: "Army Day", Description: "Pathet Lao Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 8, 0, 0, 0, 0, time.UTC), Name: "International Women's Day", Description: "Women's rights and achievements"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 22, 0, 0, 0, 0, time.UTC), Name: "Lao People's Party Day", Description: "Foundation of Lao People's Revolutionary Party"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 13, 0, 0, 0, 0, time.UTC), Name: "Lao New Year", Description: "Pi Mai - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 14, 0, 0, 0, 0, time.UTC), Name: "Lao New Year", Description: "Pi Mai - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 15, 0, 0, 0, 0, time.UTC), Name: "Lao New Year", Description: "Pi Mai - Third day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 26, 0, 0, 0, 0, time.UTC), Name: "Buddha's Birthday", Description: "Visakha Bucha Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 12, 0, 0, 0, 0, time.UTC), Name: "Liberation Day", Description: "End of French rule"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 13, 0, 0, 0, 0, time.UTC), Name: "That Luang Festival", Description: "Most important Buddhist festival"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 2, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Lao People's Democratic Republic founding"},
			}

		case "NGN": // Nigeria
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC), Name: "Easter Sunday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Workers' Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 27, 0, 0, 0, 0, time.UTC), Name: "Children's Day", Description: "Celebration of children"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Democracy Day", Description: "Return to democracy"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 1, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Nigerian Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "KES": // Kenya
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 1, 0, 0, 0, 0, time.UTC), Name: "Madaraka Day", Description: "Internal self-rule"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 27, 0, 0, 0, 0, time.UTC), Name: "Utamaduni Day", Description: "Culture Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 10, 0, 0, 0, 0, time.UTC), Name: "Huduma Day", Description: "Service Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 20, 0, 0, 0, 0, time.UTC), Name: "Mashujaa Day", Description: "Heroes' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 12, 0, 0, 0, 0, time.UTC), Name: "Jamhuri Day", Description: "Independence and Republic Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "GHS": // Ghana
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 6, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Ghanaian Independence from Britain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "May Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "Establishment of the Republic"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 4, 0, 0, 0, 0, time.UTC), Name: "Founders' Day", Description: "Celebrating founding fathers"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 21, 0, 0, 0, 0, time.UTC), Name: "Kwame Nkrumah Memorial Day", Description: "Birth of first president"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Second day of Christmas"},
			}

		case "MAD": // Morocco
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 11, 0, 0, 0, 0, time.UTC), Name: "Independence Manifesto Day", Description: "Independence movement commemoration"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 30, 0, 0, 0, 0, time.UTC), Name: "Throne Day", Description: "King Mohammed VI's accession"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 14, 0, 0, 0, 0, time.UTC), Name: "Oued Ed-Dahab Day", Description: "Allegiance of Oued Ed-Dahab"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 20, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "Revolution of the King and the People"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 21, 0, 0, 0, 0, time.UTC), Name: "Youth Day", Description: "King Mohammed VI's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Mawlid", Description: "Prophet Muhammad's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 6, 0, 0, 0, 0, time.UTC), Name: "Green March Day", Description: "Green March into Western Sahara"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 18, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Moroccan Independence from France"},
			}

		case "TND": // Tunisia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 14, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "Tunisian Revolution anniversary"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 20, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Tunisian Independence from France"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 9, 0, 0, 0, 0, time.UTC), Name: "Martyrs' Day", Description: "Commemoration of martyrs"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 25, 0, 0, 0, 0, time.UTC), Name: "Republic Day", Description: "Establishment of the Republic"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 13, 0, 0, 0, 0, time.UTC), Name: "Women's Day", Description: "Tunisian Women's Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Mawlid", Description: "Prophet Muhammad's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 15, 0, 0, 0, 0, time.UTC), Name: "Evacuation Day", Description: "Evacuation of French troops"},
			}

		case "ETB": // Ethiopia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 7, 0, 0, 0, 0, time.UTC), Name: "Ethiopian Christmas", Description: "Orthodox Christmas - Genna"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 19, 0, 0, 0, 0, time.UTC), Name: "Timkat", Description: "Ethiopian Orthodox Epiphany"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 2, 0, 0, 0, 0, time.UTC), Name: "Battle of Adwa", Description: "Victory over Italian invasion"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 27, 0, 0, 0, 0, time.UTC), Name: "Ethiopian Good Friday", Description: "Orthodox Good Friday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 29, 0, 0, 0, 0, time.UTC), Name: "Ethiopian Easter", Description: "Orthodox Easter"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 5, 0, 0, 0, 0, time.UTC), Name: "Patriots' Victory Day", Description: "Liberation from Italian occupation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 28, 0, 0, 0, 0, time.UTC), Name: "Downfall of Derg", Description: "End of military regime"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 11, 0, 0, 0, 0, time.UTC), Name: "Ethiopian New Year", Description: "Enkutatash - New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 27, 0, 0, 0, 0, time.UTC), Name: "Meskel", Description: "Finding of the True Cross"},
			}

		case "DZD": // Algeria
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 12, 0, 0, 0, 0, time.UTC), Name: "Amazigh New Year", Description: "Berber New Year - Yennayer"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 11, 0, 0, 0, 0, time.UTC), Name: "Eid al-Fitr", Description: "End of Ramadan - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 16, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - First day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Eid al-Adha", Description: "Festival of Sacrifice - Second day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 5, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Algerian Independence from France"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 7, 0, 0, 0, 0, time.UTC), Name: "Islamic New Year", Description: "First day of Muharram"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 16, 0, 0, 0, 0, time.UTC), Name: "Mawlid", Description: "Prophet Muhammad's Birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "Revolution Day", Description: "Start of Algerian Revolution"},
			}

		default:
			// For currencies without specific holidays, add common international holidays
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}
		}

		for _, holiday := range holidays {
			if err := m.HolidayManager.CreateWithTx(context, tx, holiday); err != nil {
				return eris.Wrapf(err, "failed to seed holiday %s for currency %s", holiday.Name, currency.CurrencyCode)
			}
		}
	}

	return nil
}

func (m *ModelCore) HolidayCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*Holiday, error) {
	return m.HolidayManager.Find(context, &Holiday{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
