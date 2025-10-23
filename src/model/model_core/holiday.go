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
