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
		case "PHP": // Philippines
			holidays = []*Holiday{
				// Regular Holidays
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 9, 0, 0, 0, 0, time.UTC), Name: "Araw ng Kagitingan", Description: "Day of Valor"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labor Day", Description: "Celebration of workers and laborers"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC), Name: "Independence Day", Description: "Commemorates Philippine independence from Spain"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 26, 0, 0, 0, 0, time.UTC), Name: "National Heroes Day", Description: "Honoring Philippine national heroes"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 30, 0, 0, 0, 0, time.UTC), Name: "Bonifacio Day", Description: "Commemorates the birth of Andres Bonifacio"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 30, 0, 0, 0, 0, time.UTC), Name: "Rizal Day", Description: "Commemorates the life of Dr. Jose Rizal"},

				// Special (Non-Working) Holidays
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 25, 0, 0, 0, 0, time.UTC), Name: "EDSA People Power Revolution", Description: "Commemorates the 1986 EDSA Revolution"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 21, 0, 0, 0, 0, time.UTC), Name: "Ninoy Aquino Day", Description: "Commemorates the assassination of Benigno Aquino Jr"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC), Name: "All Saints' Day", Description: "Honoring all the saints"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 8, 0, 0, 0, 0, time.UTC), Name: "Feast of the Immaculate Conception", Description: "Catholic feast day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC), Name: "New Year's Eve", Description: "Last day of the year"},
			}

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
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 2, 0, 0, 0, 0, time.UTC), Name: "New Year's Holiday", Description: "Second day of the year holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 17, 0, 0, 0, 0, time.UTC), Name: "St. Patrick's Day", Description: "Northern Ireland only"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 3, 29, 0, 0, 0, 0, time.UTC), Name: "Good Friday", Description: "Christian holiday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC), Name: "Easter Monday", Description: "Day after Easter Sunday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 6, 0, 0, 0, 0, time.UTC), Name: "Early May Bank Holiday", Description: "First Monday in May"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 27, 0, 0, 0, 0, time.UTC), Name: "Spring Bank Holiday", Description: "Last Monday in May"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 7, 12, 0, 0, 0, 0, time.UTC), Name: "Battle of the Boyne", Description: "Northern Ireland only"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 5, 0, 0, 0, 0, time.UTC), Name: "Summer Bank Holiday", Description: "Scotland only"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 26, 0, 0, 0, 0, time.UTC), Name: "Summer Bank Holiday", Description: "England, Wales, Northern Ireland"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 30, 0, 0, 0, 0, time.UTC), Name: "St. Andrew's Day", Description: "Scotland only"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "EUR": // European Union (major countries)
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 9, 0, 0, 0, 0, time.UTC), Name: "Europe Day", Description: "Schuman Declaration anniversary"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "St. Stephen's Day", Description: "Second day of Christmas"},
			}

		case "AUD": // Australia
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 26, 0, 0, 0, 0, time.UTC), Name: "Australia Day", Description: "National day of Australia"},
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
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 5, 0, 0, 0, 0, time.UTC), Name: "Civic Holiday", Description: "First Monday in August"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 2, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "First Monday in September"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 9, 30, 0, 0, 0, 0, time.UTC), Name: "Truth and Reconciliation Day", Description: "National Day for Truth and Reconciliation"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 14, 0, 0, 0, 0, time.UTC), Name: "Thanksgiving", Description: "Second Monday in October"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 11, 11, 0, 0, 0, 0, time.UTC), Name: "Remembrance Day", Description: "Honors military veterans"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC), Name: "Boxing Day", Description: "Day after Christmas"},
			}

		case "SGD": // Singapore
			holidays = []*Holiday{
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), Name: "New Year's Day", Description: "First day of the year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 10, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "First day of Chinese New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 2, 11, 0, 0, 0, 0, time.UTC), Name: "Chinese New Year", Description: "Second day of Chinese New Year"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 4, 10, 0, 0, 0, 0, time.UTC), Name: "Hari Raya Puasa", Description: "End of Ramadan"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC), Name: "Labour Day", Description: "International Workers' Day"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 5, 22, 0, 0, 0, 0, time.UTC), Name: "Vesak Day", Description: "Buddha's birthday"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 6, 17, 0, 0, 0, 0, time.UTC), Name: "Hari Raya Haji", Description: "Feast of Sacrifice"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 8, 9, 0, 0, 0, 0, time.UTC), Name: "National Day", Description: "Singapore independence"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC), Name: "Deepavali", Description: "Festival of Lights"},
				{CreatedAt: now, UpdatedAt: now, CreatedByID: userID, UpdatedByID: userID, OrganizationID: organizationID, BranchID: branchID, EntryDate: time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Christmas Day", Description: "Celebration of the birth of Jesus Christ"},
			}

		// Add more currency cases as needed...

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
