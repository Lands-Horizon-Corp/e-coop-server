package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
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
)

func (m *Model) OrganizationDailyUsageValidate(ctx echo.Context) (*OrganizationDailyUsageRequest, error) {
	return Validate[OrganizationDailyUsageRequest](ctx, m.validator)
}

func (m *Model) OrganizationDailyUsageModel(data *OrganizationDailyUsage) *OrganizationDailyUsageResponse {
	return ToModel(data, func(data *OrganizationDailyUsage) *OrganizationDailyUsageResponse {
		return &OrganizationDailyUsageResponse{
			ID:                      data.ID,
			OrganizationID:          data.OrganizationID,
			Organization:            m.OrganizationModel(data.Organization),
			TotalMembers:            data.TotalMembers,
			TotalBranches:           data.TotalBranches,
			TotalEmployees:          data.TotalEmployees,
			CashTransactionCount:    data.CashTransactionCount,
			CheckTransactionCount:   data.CheckTransactionCount,
			OnlineTransactionCount:  data.OnlineTransactionCount,
			CashTransactionAmount:   data.CashTransactionAmount,
			CheckTransactionAmount:  data.CheckTransactionAmount,
			OnlineTransactionAmount: data.OnlineTransactionAmount,
			TotalEmailSend:          data.TotalEmailSend,
			TotalMessageSend:        data.TotalMessageSend,
			TotalUploadSize:         data.TotalUploadSize,
			TotalReportRenderTime:   data.TotalReportRenderTime,
			CreatedAt:               data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:               data.UpdatedAt.Format(time.RFC3339),
		}
	})
}

func (m *Model) OrganizationDailyUsageModels(data []*OrganizationDailyUsage) []*OrganizationDailyUsageResponse {
	return ToModels(data, m.OrganizationDailyUsageModel)
}
