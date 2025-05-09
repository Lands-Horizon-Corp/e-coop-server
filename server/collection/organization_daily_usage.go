package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	// OrganizationDailyUsage represents the organization's daily usage data model
	OrganizationDailyUsage struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		TotalMembers   int           `gorm:"not null"`
		TotalBranches  int           `gorm:"not null"`
		TotalEmployees int           `gorm:"not null"`

		CashTransactionCount   int `gorm:"not null"`
		CheckTransactionCount  int `gorm:"not null"`
		OnlineTransactionCount int `gorm:"not null"`

		CashTransactionAmount   float64 `gorm:"not null"`
		CheckTransactionAmount  float64 `gorm:"not null"`
		OnlineTransactionAmount float64 `gorm:"not null"`

		TotalEmailSend        int     `gorm:"not null"`
		TotalMessageSend      int     `gorm:"not null"`
		TotalUploadSize       float64 `gorm:"not null"`
		TotalReportRenderTime float64 `gorm:"not null"`
	}

	// OrganizationDailyUsageRequest defines the payload for creating a new OrganizationDailyUsage
	OrganizationDailyUsageRequest struct {
		OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
		TotalMembers   int       `json:"total_members" validate:"required,min=0"`
		TotalBranches  int       `json:"total_branches" validate:"required,min=0"`
		TotalEmployees int       `json:"total_employees" validate:"required,min=0"`

		CashTransactionCount   int `json:"cash_transaction_count" validate:"required,min=0"`
		CheckTransactionCount  int `json:"check_transaction_count" validate:"required,min=0"`
		OnlineTransactionCount int `json:"online_transaction_count" validate:"required,min=0"`

		CashTransactionAmount   float64 `json:"cash_transaction_amount" validate:"required,min=0"`
		CheckTransactionAmount  float64 `json:"check_transaction_amount" validate:"required,min=0"`
		OnlineTransactionAmount float64 `json:"online_transaction_amount" validate:"required,min=0"`

		TotalEmailSend        int     `json:"total_email_send" validate:"required,min=0"`
		TotalMessageSend      int     `json:"total_message_send" validate:"required,min=0"`
		TotalUploadSize       float64 `json:"total_upload_size" validate:"required,min=0"`
		TotalReportRenderTime float64 `json:"total_report_render_time" validate:"required,min=0"`
	}

	// OrganizationDailyUsageResponse defines the HTTP response for OrganizationDailyUsage
	OrganizationDailyUsageResponse struct {
		ID             uuid.UUID             `json:"id"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		TotalMembers   int                   `json:"total_members"`
		TotalBranches  int                   `json:"total_branches"`
		TotalEmployees int                   `json:"total_employees"`

		CashTransactionCount   int `json:"cash_transaction_count"`
		CheckTransactionCount  int `json:"check_transaction_count"`
		OnlineTransactionCount int `json:"online_transaction_count"`

		CashTransactionAmount   float64 `json:"cash_transaction_amount"`
		CheckTransactionAmount  float64 `json:"check_transaction_amount"`
		OnlineTransactionAmount float64 `json:"online_transaction_amount"`

		TotalEmailSend        int     `json:"total_email_send"`
		TotalMessageSend      int     `json:"total_message_send"`
		TotalUploadSize       float64 `json:"total_upload_size"`
		TotalReportRenderTime float64 `json:"total_report_render_time"`
		CreatedAt             string  `json:"created_at"`
		UpdatedAt             string  `json:"updated_at"`
	}

	OrganizationDailyUsageCollection struct {
		validator       *validator.Validate
		organizationCol *OrganizationCollection
	}
)

func NewOrganizationDailyUsageCollection(organizationCol *OrganizationCollection) (*OrganizationDailyUsageCollection, error) {
	return &OrganizationDailyUsageCollection{
		validator:       validator.New(),
		organizationCol: organizationCol,
	}, nil
}

func (oduc *OrganizationDailyUsageCollection) ValidateCreate(c echo.Context) (*OrganizationDailyUsageRequest, error) {
	req := new(OrganizationDailyUsageRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := oduc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (oduc *OrganizationDailyUsageCollection) ToModel(odu *OrganizationDailyUsage) *OrganizationDailyUsageResponse {
	if odu == nil {
		return nil
	}
	resp := &OrganizationDailyUsageResponse{
		ID:                      odu.ID,
		OrganizationID:          odu.OrganizationID,
		Organization:            oduc.organizationCol.ToModel(odu.Organization),
		TotalMembers:            odu.TotalMembers,
		TotalBranches:           odu.TotalBranches,
		TotalEmployees:          odu.TotalEmployees,
		CashTransactionCount:    odu.CashTransactionCount,
		CheckTransactionCount:   odu.CheckTransactionCount,
		OnlineTransactionCount:  odu.OnlineTransactionCount,
		CashTransactionAmount:   odu.CashTransactionAmount,
		CheckTransactionAmount:  odu.CheckTransactionAmount,
		OnlineTransactionAmount: odu.OnlineTransactionAmount,
		TotalEmailSend:          odu.TotalEmailSend,
		TotalMessageSend:        odu.TotalMessageSend,
		TotalUploadSize:         odu.TotalUploadSize,
		TotalReportRenderTime:   odu.TotalReportRenderTime,
		CreatedAt:               odu.CreatedAt.Format(time.RFC3339),
		UpdatedAt:               odu.UpdatedAt.Format(time.RFC3339),
	}
	return resp
}

// ToModels maps a slice of OrganizationDailyUsage DB models to OrganizationDailyUsageResponse
func (oduc *OrganizationDailyUsageCollection) ToModels(data []*OrganizationDailyUsage) []*OrganizationDailyUsageResponse {
	if data == nil {
		return []*OrganizationDailyUsageResponse{}
	}
	var out []*OrganizationDailyUsageResponse
	for _, odu := range data {
		if m := oduc.ToModel(odu); m != nil {
			out = append(out, m)
		}
	}
	return out
}
