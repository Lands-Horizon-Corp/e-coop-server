package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"

	"gorm.io/gorm"
)

type (
	InvitationCode struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_invitation_code"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_invitation_code"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserType       string    `gorm:"type:varchar(255);not null"`
		Code           string    `gorm:"type:varchar(255);not null;unique"`
		ExpirationDate time.Time `gorm:"not null"`
		MaxUse         int       `gorm:"not null"`
		CurrentUse     int       `gorm:"default:0"`
		Description    string    `gorm:"type:text"`
	}

	InvitationCodeResponse struct {
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

		UserType       string            `json:"user_type"`
		Code           string            `json:"code"`
		ExpirationDate string            `json:"expiration_date"`
		MaxUse         int               `json:"max_use"`
		CurrentUse     int               `json:"current_use"`
		Description    string            `json:"description,omitempty"`
		QRCode         *horizon.QRResult `json:"qr_code,omitempty"`
	}

	InvitationCodeRequest struct {
		UserType       string    `json:"user_type" validate:"required,oneof=employee owner member"`
		ExpirationDate time.Time `json:"expiration_date" validate:"required"`
		MaxUse         int       `json:"max_use" validate:"required"`
		Description    string    `json:"description,omitempty"`
	}

	QRInvitationLInk struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"UserType"`
		Code           string `json:"Code"`
		CurrentUse     int    `json:"CurrentUse"`
		Description    string `json:"Description"`
	}
)

func (m *Model) InvitationCodeValidate(ctx echo.Context) (*InvitationCodeRequest, error) {
	return Validate[InvitationCodeRequest](ctx, m.validator)
}

func (m *Model) InvitationCodeModel(data *InvitationCode) *InvitationCodeResponse {

	return ToModel(data, func(data *InvitationCode) *InvitationCodeResponse {
		encoded, err := m.qr.Encode(&QRInvitationLInk{
			OrganizationID: data.OrganizationID.String(),
			BranchID:       data.BranchID.String(),
			UserType:       data.UserType,
			Code:           data.Code,
			CurrentUse:     data.CurrentUse,
			Description:    data.Description,
		})
		if err != nil {
			return nil
		}
		return &InvitationCodeResponse{
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
			UserType:       data.UserType,

			Code:           data.Code,
			ExpirationDate: data.ExpirationDate.Format(time.RFC3339),
			MaxUse:         data.MaxUse,
			CurrentUse:     data.CurrentUse,
			Description:    data.Description,
			QRCode:         encoded,
		}
	})
}

func (m *Model) InvitationCodeModels(data []*InvitationCode) []*InvitationCodeResponse {
	return ToModels(data, m.InvitationCodeModel)
}
