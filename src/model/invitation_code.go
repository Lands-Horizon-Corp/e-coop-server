package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	InvitationCode struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
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

		PermissionName        string         `gorm:"type:varchar(255);not null" json:"permission_name"`
		PermissionDescription string         `gorm:"type:varchar(255);not null" json:"permission_description"`
		Permissions           pq.StringArray `gorm:"type:varchar(255)[]" json:"permissions"`
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

		PermissionName        string   `json:"permission_name"`
		PermissionDescription string   `json:"permission_description"`
		Permissions           []string `json:"permissions"`
	}

	InvitationCodeRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		UserType       string    `json:"user_type" validate:"required,oneof=employee owner member"`
		Code           string    `json:"code" validate:"required,max=255"`
		ExpirationDate time.Time `json:"expiration_date" validate:"required"`
		MaxUse         int       `json:"max_use" validate:"required"`
		Description    string    `json:"description,omitempty"`

		PermissionName        string   `json:"permission_name,omitempty"`
		PermissionDescription string   `json:"permission_description,omitempty"`
		Permissions           []string `json:"permissions,omitempty" validate:"dive"`
	}
)

func (m *Model) InvitationCode() {
	m.Migration = append(m.Migration, &InvitationCode{})
	m.InvitationCodeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Organization.Media",
			"Branch.Media",
			"Branch",
		},
		Service: m.provider.Service,
		Resource: func(data *InvitationCode) *InvitationCodeResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}

			return &InvitationCodeResponse{
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

				UserType:       data.UserType,
				Code:           data.Code,
				ExpirationDate: data.ExpirationDate.Format(time.RFC3339),
				MaxUse:         data.MaxUse,
				CurrentUse:     data.CurrentUse,
				Description:    data.Description,

				PermissionName:        data.PermissionName,
				PermissionDescription: data.PermissionDescription,
				Permissions:           data.Permissions,
			}
		},
		Created: func(data *InvitationCode) []string {
			return []string{
				"invitation_code.create",
				fmt.Sprintf("invitation_code.create.%s", data.ID),
				fmt.Sprintf("invitation_code.create.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *InvitationCode) []string {
			return []string{
				"invitation_code.update",
				fmt.Sprintf("invitation_code.update.%s", data.ID),
				fmt.Sprintf("invitation_code.update.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *InvitationCode) []string {
			return []string{
				"invitation_code.delete",
				fmt.Sprintf("invitation_code.delete.%s", data.ID),
				fmt.Sprintf("invitation_code.delete.branch.%s", data.BranchID),
				fmt.Sprintf("invitation_code.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) GetInvitationCodeByBranch(context context.Context, organizationId uuid.UUID, branchId uuid.UUID) ([]*InvitationCode, error) {
	return m.InvitationCodeManager.Find(context, &InvitationCode{
		OrganizationID: organizationId,
		BranchID:       branchId,
	})
}

func (m *Model) GetInvitationCodeByCode(context context.Context, code string) (*InvitationCode, error) {
	return m.InvitationCodeManager.FindOne(context, &InvitationCode{
		Code: code,
	})
}

func (m *Model) InvitationCodeSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now()
	expiration := now.AddDate(0, 1, 0)

	invitationCodes := []*InvitationCode{
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       "employee",
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         5,
			CurrentUse:     0,
			Description:    "Invitation code for employees (max 5 uses)",
		},
		{
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			UserType:       "member",
			Code:           uuid.New().String(),
			ExpirationDate: expiration,
			MaxUse:         1000,
			CurrentUse:     0,
			Description:    "Invitation code for members (max 1000 uses)",
		},
	}
	for _, data := range invitationCodes {
		if err := m.InvitationCodeManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed invitation code for %s", data.UserType)
		}
	}
	return nil
}

func (m *Model) VerifyInvitationCodeByCode(context context.Context, code string) (*InvitationCode, error) {
	data, err := m.GetInvitationCodeByCode(context, code)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if now.After(data.ExpirationDate) {
		return nil, eris.Errorf("invitation code %q expired on %s", code, data.ExpirationDate.Format(time.RFC3339))
	}
	if data.CurrentUse >= data.MaxUse {
		return nil, eris.Errorf(
			"invitation code %q has already been used %d times (max %d)",
			code, data.CurrentUse, data.MaxUse,
		)
	}
	return data, nil
}

func (m *Model) RedeemInvitationCode(context context.Context, tx *gorm.DB, invitationCodeId uuid.UUID) error {
	data, err := m.InvitationCodeManager.GetByID(context, invitationCodeId)
	if err != nil {
		return err
	}
	data.CurrentUse++
	if err := m.InvitationCodeManager.UpdateWithTx(context, tx, data); err != nil {
		return eris.Wrapf(
			err,
			"failed to redeem invitation code %q (increment CurrentUse)",
			data.Code,
		)
	}
	return nil
}
