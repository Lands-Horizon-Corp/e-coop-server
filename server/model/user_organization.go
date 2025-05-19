package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	UserOrganization struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_org_branch"`

		Organization *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID     *uuid.UUID    `gorm:"type:uuid;index:idx_user_org_branch"`
		Branch       *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`
		UserID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_user_org_branch"`

		User                   *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
		UserType               string         `gorm:"type:varchar(50);not null"`
		Description            string         `gorm:"type:text" json:"description,omitempty"`
		ApplicationDescription string         `gorm:"type:text" json:"application_description,omitempty"`
		ApplicationStatus      string         `gorm:"type:varchar(50);not null;default:'pending'" json:"application_status"`
		DeveloperSecretKey     string         `gorm:"type:varchar(255);not null;unique" json:"developer_secret_key"`
		PermissionName         string         `gorm:"type:varchar(255);not null" json:"permission_name"`
		PermissionDescription  string         `gorm:"type:varchar(255);not null" json:"permission_description"`
		Permissions            pq.StringArray `gorm:"type:varchar(255)[]" json:"permissions"`
		IsSeeded               bool           `gorm:"not null;default:false" json:"is_seeded"`

		UserSettingDescription string `gorm:"type:text" json:"user_setting_description"`

		UserSettingStartOR int64 `gorm:"unsigned" json:"start_or"`
		UserSettingEndOR   int64 `gorm:"unsigned" json:"end_or"`
		UserSettingUsedOR  int64 `gorm:"unsigned" json:"used_or"`

		UserSettingStartVoucher int64 `gorm:"unsigned" json:"start_voucher"`
		UserSettingEndVoucher   int64 `gorm:"unsigned" json:"end_voucher"`
		UserSettingUsedVoucher  int64 `gorm:"unsigned" json:"used_voucher"`
	}

	UserOrganizationRequest struct {
		ID       *uuid.UUID `json:"id,omitempty"`
		UserType string     `json:"user_type,omitempty" validate:"omitempty,oneof=employee member"`

		Description            string   `json:"description,omitempty"`
		ApplicationDescription string   `json:"application_description,omitempty"`
		ApplicationStatus      string   `json:"application_status" validate:"omitempty,oneof=pending reported accepted ban not-allowed"`
		PermissionName         string   `json:"permission_name,omitempty"`
		PermissionDescription  string   `json:"permission_description,omitempty"`
		Permissions            []string `json:"permissions,omitempty" validate:"dive"`

		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR   int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR  int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`
	}

	UserOrganizationResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       *uuid.UUID            `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		UserID                 uuid.UUID     `json:"user_id"`
		User                   *UserResponse `json:"user,omitempty"`
		UserType               string        `json:"user_type"`
		Description            string        `json:"description,omitempty"`
		ApplicationDescription string        `json:"application_description,omitempty"`
		ApplicationStatus      string        `json:"application_status"`
		DeveloperSecretKey     string        `json:"developer_secret_key"`
		PermissionName         string        `json:"permission_name"`
		PermissionDescription  string        `json:"permission_description"`
		Permissions            []string      `json:"permissions"`

		UserSettingDescription string `json:"user_setting_description"`

		UserSettingStartOR int64 `json:"user_setting_start_or"`
		UserSettingEndOR   int64 `json:"user_setting_end_or"`
		UserSettingUsedOR  int64 `json:"user_setting_used_or"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher"`
	}
	UserOrganizationCollection struct {
		Manager horizon_manager.CollectionManager[UserOrganization]
	}
)

func (m *Model) UserOrganizationValidate(ctx echo.Context) (*UserOrganizationRequest, error) {
	return horizon_manager.Validate[UserOrganizationRequest](ctx, m.validator)
}

func (m *Model) UserOrganizationModel(data *UserOrganization) *UserOrganizationResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *UserOrganization) *UserOrganizationResponse {
		return &UserOrganizationResponse{
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

			UserType:               data.UserType,
			UserID:                 data.UserID,
			User:                   m.UserModel(data.User),
			Description:            data.Description,
			ApplicationDescription: data.ApplicationDescription,
			ApplicationStatus:      data.ApplicationStatus,
			DeveloperSecretKey:     data.DeveloperSecretKey,
			PermissionName:         data.PermissionName,
			PermissionDescription:  data.PermissionDescription,
			Permissions:            data.Permissions,

			UserSettingDescription:  data.UserSettingDescription,
			UserSettingStartOR:      data.UserSettingStartOR,
			UserSettingEndOR:        data.UserSettingEndOR,
			UserSettingUsedOR:       data.UserSettingUsedOR,
			UserSettingStartVoucher: data.UserSettingStartVoucher,
			UserSettingEndVoucher:   data.UserSettingEndVoucher,
			UserSettingUsedVoucher:  data.UserSettingUsedVoucher,
		}
	})
}

func (m *Model) UserOrganizationModels(data []*UserOrganization) []*UserOrganizationResponse {
	return horizon_manager.ToModels(data, m.UserOrganizationModel)
}

func NewUserOrganizationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*UserOrganizationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.create",
				fmt.Sprintf("user_organization.create.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.update",
				fmt.Sprintf("user_organization.update.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		func(data *UserOrganization) ([]string, any) {
			return []string{
				"user_organization.delete",
				fmt.Sprintf("user_organization.delete.%s", data.ID),
			}, model.UserOrganizationModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Branch.Media",
			"User",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Organization.OrganizationCategory.Category",
		},
	)
	return &UserOrganizationCollection{
		Manager: manager,
	}, nil
}

// user-organization/user/:user_id
func (fc *UserOrganizationCollection) ListByUser(userID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		UserID: userID,
	})
}

// user-organization/branch/:branch_id
func (fc *UserOrganizationCollection) ListByBranch(branchID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		BranchID: &branchID,
	})
}

// user-organization/organization/:organization_id
func (fc *UserOrganizationCollection) ListByOrganization(organizationID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		OrganizationID: organizationID,
	})
}

// user-organization/organization/:organization_id/branch/:branch_id
func (fc *UserOrganizationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) (*UserOrganization, error) {
	return fc.Manager.FindOne(&UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}

// user-organization/user/:user_id/branch/:branch_id
func (fc *UserOrganizationCollection) ListByUserBranch(userID uuid.UUID, branchID uuid.UUID) (*UserOrganization, error) {
	return fc.Manager.FindOne(&UserOrganization{
		UserID:   userID,
		BranchID: &branchID,
	})
}

// user-organization/user/:user_id/organization/:organization_id
func (fc *UserOrganizationCollection) ListByUserOrganization(userID uuid.UUID, organizationID uuid.UUID) ([]*UserOrganization, error) {
	return fc.Manager.Find(&UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
}

// user-organization/user/:user_id/organization/:organization_id/branch/:branch_id
func (fc *UserOrganizationCollection) ByUserOrganizationBranch(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (*UserOrganization, error) {
	return fc.Manager.FindOne(&UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}

func (fc *UserOrganizationCollection) CountUserOrganizationBranch(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return fc.Manager.Count(&UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
}

func (fc *UserOrganizationCollection) CountByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return fc.Manager.Count(&UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}
func (fc *UserOrganizationCollection) CountByOrganization(organizationID uuid.UUID) (int64, error) {
	return fc.Manager.Count(&UserOrganization{
		OrganizationID: organizationID,
	})
}

func (fc *UserOrganizationCollection) EmployeeCanJoin(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existingCount, err := fc.CountUserOrganizationBranch(userID, organizationID, branchID)
	return err == nil && existingCount == 0
}

func (fc *UserOrganizationCollection) MemberCanJoin(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existingBranchCount, err := fc.CountUserOrganizationBranch(userID, organizationID, branchID)
	if err != nil || existingBranchCount > 0 {
		return false
	}
	existingOrgCount, err := fc.Manager.Count(&UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	return err == nil && existingOrgCount == 0
}
