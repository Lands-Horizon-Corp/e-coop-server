package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
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
		Code           string    `json:"code" validate:"required,max=255"`
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
	InvitationCodeCollection struct {
		Manager CollectionManager[InvitationCode]
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

func NewInvitationCodeCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*InvitationCodeCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *InvitationCode) ([]string, any) {
			return []string{
				"invitation_code.create",
				fmt.Sprintf("invitation_code.create.%s", data.ID),
				fmt.Sprintf("invitation_code.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("invitation_code.create.organization.%s", data.BranchID),
				fmt.Sprintf("invitation_code.create.user.%s", data.CreatedByID),
			}, model.InvitationCodeModel(data)
		},
		func(data *InvitationCode) ([]string, any) {
			return []string{
				"invitation_code.update",
				fmt.Sprintf("invitation_code.update.%s", data.ID),
				fmt.Sprintf("invitation_code.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("invitation_code.update.organization.%s", data.BranchID),
				fmt.Sprintf("invitation_code.update.user.%s", data.CreatedByID),
			}, model.InvitationCodeModel(data)
		},
		func(data *InvitationCode) ([]string, any) {
			return []string{
				"invitation_code.delete",
				fmt.Sprintf("invitation_code.delete.%s", data.ID),
				fmt.Sprintf("invitation_code.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("invitation_code.delete.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.delete.user.%s", data.CreatedByID),
			}, model.InvitationCodeModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &InvitationCodeCollection{
		Manager: manager,
	}, nil
}

// invitation-code/branch/:branch_id
func (fc *InvitationCodeCollection) ListByBranch(branchID uuid.UUID) ([]*InvitationCode, error) {
	return fc.Manager.Find(&InvitationCode{
		BranchID: branchID,
	})
}

// invitation-code/organization/:organization_id
func (fc *InvitationCodeCollection) ListByOrganization(organizationID uuid.UUID) ([]*InvitationCode, error) {
	return fc.Manager.Find(&InvitationCode{
		OrganizationID: organizationID,
	})
}

// invitation-code/organization/:organization_id/branch/:branch_id
func (fc *InvitationCodeCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*InvitationCode, error) {
	return fc.Manager.Find(&InvitationCode{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

// invitation-code/exists/:code
func (ic *InvitationCodeCollection) Exists(code string) (bool, error) {
	_, err := ic.Manager.FindOne(&InvitationCode{Code: code})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// invitation-code/code/:code
func (ic *InvitationCodeCollection) ByCode(code string) (*InvitationCode, error) {
	return ic.Manager.FindOne(&InvitationCode{Code: code})
}

func (ic *InvitationCodeCollection) Redeem(tx *gorm.DB, code string) (*InvitationCode, error) {
	inv, err := ic.Manager.FindOne(&InvitationCode{Code: code})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "failed to lookup invitation code %q", code)
	}
	now := time.Now()
	if now.After(inv.ExpirationDate) {
		return nil, eris.Errorf("invitation code %q expired on %s", code, inv.ExpirationDate.Format(time.RFC3339))
	}
	if inv.CurrentUse >= inv.MaxUse {
		return nil, eris.Errorf(
			"invitation code %q has already been used %d times (max %d)",
			code, inv.CurrentUse, inv.MaxUse,
		)
	}
	inv.CurrentUse++
	if err := ic.Manager.UpdateWithTx(tx, inv); err != nil {
		return nil, eris.Wrapf(
			err,
			"failed to redeem invitation code %q (increment CurrentUse)",
			code,
		)
	}
	return inv, nil
}

// invitation-code/verfiy/:code
func (ic *InvitationCodeCollection) Verify(code string) (*InvitationCode, error) {
	inv, err := ic.Manager.FindOne(&InvitationCode{Code: code})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "failed to lookup invitation code %q", code)
	}
	now := time.Now()
	if now.After(inv.ExpirationDate) {
		return nil, eris.Errorf("invitation code %q expired on %s", code, inv.ExpirationDate.Format(time.RFC3339))
	}
	if inv.CurrentUse >= inv.MaxUse {
		return nil, eris.Errorf(
			"invitation code %q has already been used %d times (max %d)",
			code, inv.CurrentUse, inv.MaxUse,
		)
	}
	return inv, nil
}
