package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserOrganizationStatus string

const (
	UserOrganizationStatusOnline    UserOrganizationStatus = "online"
	UserOrganizationStatusOffline   UserOrganizationStatus = "offline"
	UserOrganizationStatusBusy      UserOrganizationStatus = "busy"
	UserOrganizationStatusVacation  UserOrganizationStatus = "vacation"
	UserOrganizationStatusCommuting UserOrganizationStatus = "commuting"
)

type (
	UserOrganization struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_user_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`

		BranchID *uuid.UUID `gorm:"type:uuid;index:idx_user_org_branch" json:"branch_id,omitempty"`
		Branch   *Branch    `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_org_branch" json:"user_id"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`

		UserType               string         `gorm:"type:varchar(50);not null" json:"user_type"`
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

		// Override settings for branch
		SettingsAllowWithdrawNegativeBalance bool `gorm:"not null;default:false" json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `gorm:"not null;default:false" json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `gorm:"not null;default:false" json:"maintaining_balance"`

		Status       UserOrganizationStatus `gorm:"type:varchar(50);not null;default:'offline'" json:"status"`
		LastOnlineAt time.Time              `gorm:"default:now()" json:"last_online_at"`
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
	UserOrganizationSettingsRequest struct {
		UserType    string `json:"user_type,omitempty" validate:"omitempty,oneof=employee member"`
		Description string `json:"description,omitempty"`

		ApplicationDescription string `json:"application_description,omitempty"`
		ApplicationStatus      string `json:"application_status" validate:"omitempty,oneof=pending reported accepted ban not-allowed"`

		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR   int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR  int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`
	}

	UserOrganizationSelfSettingsRequest struct {
		Description            string `json:"description,omitempty"`
		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR   int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR  int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`
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

		// Override settings for branch
		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		Status       UserOrganizationStatus `json:"status"`
		LastOnlineAt time.Time              `json:"last_online_at"`
	}

	UserOrganizationPermissionPayload struct {
		PermissionName        string   `json:"permission_name" validate:"required"`
		PermissionDescription string   `json:"permission_description" validate:"required"`
		Permissions           []string `json:"permissions" validate:"required,min=1,dive,required"`
	}

	DeveloperSecretKeyResponse struct {
		DeveloperSecretKey string `json:"developer_secret_key"`
	}

	UserOrganizationStatusRequest struct {
		UserOrganizationStatus UserOrganizationStatus `json:"user_organization_status" validate:"required,oneof=online offline busy vacation commuting"`
	}
	UserOrganizationStatusResponse struct {
		OfflineUsers   []*UserOrganizationResponse `json:"user_organizations,omitempty"`
		OnlineUsers    []*UserOrganizationResponse `json:"online_user_organizations,omitempty"`
		CommutingUsers []*UserOrganizationResponse `json:"commuting_user_organizations,omitempty"`
		BusyUsers      []*UserOrganizationResponse `json:"busy_user_organizations,omitempty"`
		VacationUsers  []*UserOrganizationResponse `json:"vacation_user_organizations,omitempty"`
	}
)

func (m *Model) UserOrganization() {
	m.Migration = append(m.Migration, &UserOrganization{})
	m.UserOrganizationManager = horizon_services.NewRepository(horizon_services.RepositoryParams[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Branch.Media",
			"User",
			"User.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Organization.OrganizationCategories",
			"Organization.OrganizationCategories.Category",
		},
		Service: m.provider.Service,
		Resource: func(data *UserOrganization) *UserOrganizationResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}
			return &UserOrganizationResponse{
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

				UserID:                 data.UserID,
				User:                   m.UserManager.ToModel(data.User),
				UserType:               data.UserType,
				Description:            data.Description,
				ApplicationDescription: data.ApplicationDescription,
				ApplicationStatus:      data.ApplicationStatus,
				DeveloperSecretKey:     "",
				PermissionName:         data.PermissionName,
				PermissionDescription:  data.PermissionDescription,
				Permissions:            data.Permissions,

				UserSettingDescription: data.UserSettingDescription,

				UserSettingStartOR:                   data.UserSettingStartOR,
				UserSettingEndOR:                     data.UserSettingEndOR,
				UserSettingUsedOR:                    data.UserSettingUsedOR,
				UserSettingStartVoucher:              data.UserSettingStartVoucher,
				UserSettingEndVoucher:                data.UserSettingEndVoucher,
				UserSettingUsedVoucher:               data.UserSettingUsedVoucher,
				SettingsAllowWithdrawNegativeBalance: data.SettingsAllowWithdrawNegativeBalance,
				SettingsAllowWithdrawExactBalance:    data.SettingsAllowWithdrawExactBalance,
				SettingsMaintainingBalance:           data.SettingsMaintainingBalance,
				Status:                               data.Status,
				LastOnlineAt:                         data.LastOnlineAt,
			}
		},
		Created: func(data *UserOrganization) []string {
			return []string{
				"user_organization.create",
				fmt.Sprintf("user_organization.create.%s", data.ID),
				fmt.Sprintf("user_organization.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *UserOrganization) []string {
			return []string{
				"user_organization.update",
				fmt.Sprintf("user_organization.update.%s", data.ID),
				fmt.Sprintf("user_organization.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *UserOrganization) []string {
			return []string{
				"user_organization.delete",
				fmt.Sprintf("user_organization.delete.%s", data.ID),
				fmt.Sprintf("user_organization.delete.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.delete.user.%s", data.UserID),
			}
		},
	})
}

func (m *Model) GetUserOrganizationByUser(context context.Context, userId uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		UserID: userId,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}

func (m *Model) GetUserOrganizationByOrganization(context context.Context, organizationId uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationId,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}

func (m *Model) GetUserOrganizationByBranch(context context.Context, organizationId uuid.UUID, branchId uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationId,
		BranchID:       &branchId,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}
func (m *Model) CountUserOrganizationPerBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return m.UserOrganizationManager.Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}
func (m *Model) CountUserOrganizationBranch(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return m.UserOrganizationManager.Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
}
func (m *Model) UserOrganizationEmployeeCanJoin(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := m.CountUserOrganizationBranch(context, userID, organizationID, branchID)
	return err == nil && existing == 0
}

func (m *Model) UserOrganizationMemberCanJoin(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := m.CountUserOrganizationBranch(context, userID, organizationID, branchID)
	if err != nil || existing > 0 {
		return false
	}
	existingOrgCount, err := m.UserOrganizationManager.Count(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	return err == nil && existingOrgCount == 0
}

func (m *Model) Employees(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return m.UserOrganizationManager.Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       "employee",
	})
}

func (m *Model) Members(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return m.UserOrganizationManager.Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       "members",
	})
}
