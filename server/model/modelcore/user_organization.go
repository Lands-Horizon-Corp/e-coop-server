package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// UserOrganizationStatus represents the online status of a user within an organization
type UserOrganizationStatus string

const (
	// UserOrganizationStatusOnline indicates the user is currently online
	UserOrganizationStatusOnline UserOrganizationStatus = "online"
	// UserOrganizationStatusOffline indicates the user is currently offline
	UserOrganizationStatusOffline UserOrganizationStatus = "offline"
	// UserOrganizationStatusBusy indicates the user is currently busy
	UserOrganizationStatusBusy UserOrganizationStatus = "busy"
	// UserOrganizationStatusVacation indicates the user is on vacation
	UserOrganizationStatusVacation UserOrganizationStatus = "vacation"
	// UserOrganizationStatusCommuting indicates the user is commuting
	UserOrganizationStatusCommuting UserOrganizationStatus = "commuting"
)

// UserOrganizationType represents the role type of a user within an organization
type UserOrganizationType string

const (
	// UserOrganizationTypeOwner indicates the user is an owner of the organization
	UserOrganizationTypeOwner UserOrganizationType = "owner"
	// UserOrganizationTypeEmployee indicates the user is an employee of the organization
	UserOrganizationTypeEmployee UserOrganizationType = "employee"
	// UserOrganizationTypeMember indicates the user is a member of the organization
	UserOrganizationTypeMember UserOrganizationType = "member"
)

type (
	// UserOrganization represents the relationship between a user and an organization/branch
	UserOrganization struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id,omitempty"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_user_org_branch" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`

		BranchID *uuid.UUID `gorm:"type:uuid;index:idx_user_org_branch" json:"branch_id,omitempty"`
		Branch   *Branch    `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_org_branch" json:"user_id"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`

		UserType               UserOrganizationType `gorm:"type:varchar(50);not null" json:"user_type"`
		Description            string               `gorm:"type:text" json:"description,omitempty"`
		ApplicationDescription string               `gorm:"type:text" json:"application_description,omitempty"`
		ApplicationStatus      string               `gorm:"type:varchar(50);not null;default:'pending'" json:"application_status"`
		DeveloperSecretKey     string               `gorm:"type:varchar(255);not null;unique" json:"developer_secret_key"`
		PermissionName         string               `gorm:"type:varchar(255);not null" json:"permission_name"`
		PermissionDescription  string               `gorm:"type:varchar(255);not null" json:"permission_description"`
		Permissions            pq.StringArray       `gorm:"type:varchar(255)[]" json:"permissions"`
		IsSeeded               bool                 `gorm:"not null;default:false" json:"is_seeded"`

		UserSettingDescription   string `gorm:"type:text" json:"user_setting_description"`
		UserSettingNumberPadding int    `gorm:"not null;default:0" json:"user_setting_number_padding"`

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

		SettingsAccountingPaymentDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingPaymentDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingPaymentDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_payment_default_value,omitempty"`

		SettingsAccountingDepositDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingDepositDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_deposit_default_value,omitempty"`

		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingWithdrawDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_withdraw_default_value,omitempty"`

		SettingsPaymentTypeDefaultValueID *uuid.UUID   `gorm:"type:uuid;index" json:"settings_payment_type_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValue   *PaymentType `gorm:"foreignKey:SettingsPaymentTypeDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_payment_type_default_value,omitempty"`

		BranchSettingDefaultMemberTypeID *uuid.UUID  `gorm:"type:uuid;index" json:"branch_setting_default_member_type_id,omitempty"`
		BranchSettingDefaultMemberType   *MemberType `gorm:"foreignKey:BranchSettingDefaultMemberTypeID;constraint:OnDelete:SET NULL;"`
	}

	// UserOrganizationRequest represents the request payload for creating or updating user organization data
	UserOrganizationRequest struct {
		ID       *uuid.UUID           `json:"id,omitempty"`
		UserType UserOrganizationType `json:"user_type,omitempty" validate:"omitempty,oneof=employee member owner"`

		Description            string   `json:"description,omitempty"`
		ApplicationDescription string   `json:"application_description,omitempty"`
		ApplicationStatus      string   `json:"application_status" validate:"omitempty,oneof=pending reported accepted ban not-allowed"`
		PermissionName         string   `json:"permission_name,omitempty"`
		PermissionDescription  string   `json:"permission_description,omitempty"`
		Permissions            []string `json:"permissions,omitempty" validate:"dive"`

		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR       int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR         int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR        int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`
		UserSettingNumberPadding int   `json:"user_setting_number_padding,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`
	}

	// UserOrganizationSettingsRequest represents the request payload for updating user organization settings
	UserOrganizationSettingsRequest struct {
		UserType    UserOrganizationType `json:"user_type,omitempty" validate:"omitempty,oneof=employee member"`
		Description string               `json:"description,omitempty"`

		ApplicationDescription string `json:"application_description,omitempty"`
		ApplicationStatus      string `json:"application_status" validate:"omitempty,oneof=pending reported accepted ban not-allowed"`

		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR       int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR         int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR        int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`
		UserSettingNumberPadding int   `json:"user_setting_number_padding,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`
	}

	// UserOrganizationSelfSettingsRequest represents the request payload for users updating their own organization settings
	UserOrganizationSelfSettingsRequest struct {
		Description            string `json:"description,omitempty"`
		UserSettingDescription string `json:"user_setting_description,omitempty"`

		UserSettingStartOR       int64 `json:"user_setting_start_or,omitempty" validate:"min=0"`
		UserSettingEndOR         int64 `json:"user_setting_end_or,omitempty" validate:"min=0"`
		UserSettingUsedOR        int64 `json:"user_setting_used_or,omitempty" validate:"min=0"`
		UserSettingNumberPadding int   `json:"user_setting_number_padding,omitempty" validate:"min=0"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher,omitempty" validate:"min=0"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher,omitempty" validate:"min=0"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`
	}

	// UserOrganizationResponse represents the JSON response structure for user organization data
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

		UserID                 uuid.UUID            `json:"user_id"`
		User                   *UserResponse        `json:"user,omitempty"`
		UserType               UserOrganizationType `json:"user_type"`
		Description            string               `json:"description,omitempty"`
		ApplicationDescription string               `json:"application_description,omitempty"`
		ApplicationStatus      string               `json:"application_status"`
		DeveloperSecretKey     string               `json:"developer_secret_key"`
		PermissionName         string               `json:"permission_name"`
		PermissionDescription  string               `json:"permission_description"`
		Permissions            []string             `json:"permissions"`

		UserSettingDescription string `json:"user_setting_description"`

		UserSettingNumberPadding int   `json:"user_setting_number_padding"`
		UserSettingStartOR       int64 `json:"user_setting_start_or"`
		UserSettingEndOR         int64 `json:"user_setting_end_or"`
		UserSettingUsedOR        int64 `json:"user_setting_used_or"`

		UserSettingStartVoucher int64 `json:"user_setting_start_voucher"`
		UserSettingEndVoucher   int64 `json:"user_setting_end_voucher"`
		UserSettingUsedVoucher  int64 `json:"user_setting_used_voucher"`

		// Override settings for branch
		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		Status       UserOrganizationStatus `json:"status"`
		LastOnlineAt time.Time              `json:"last_online_at"`

		SettingsAccountingPaymentDefaultValueID *uuid.UUID       `json:"settings_accounting_payment_default_value_id"`
		SettingsAccountingPaymentDefaultValue   *AccountResponse `json:"settings_accounting_payment_default_value,omitempty"`

		SettingsAccountingDepositDefaultValueID *uuid.UUID       `json:"settings_accounting_deposit_default_value_id"`
		SettingsAccountingDepositDefaultValue   *AccountResponse `json:"settings_accounting_deposit_default_value,omitempty"`

		SettingsAccountingWithdrawDefaultValueID *uuid.UUID       `json:"settings_accounting_withdraw_default_value_id"`
		SettingsAccountingWithdrawDefaultValue   *AccountResponse `json:"settings_accounting_withdraw_default_value,omitempty"`

		SettingsPaymentTypeDefaultValueID *uuid.UUID           `json:"settings_payment_type_default_value_id"`
		SettingsPaymentTypeDefaultValue   *PaymentTypeResponse `json:"settings_payment_type_default_value,omitempty"`
	}

	// UserOrganizationPermissionPayload represents the payload for managing user organization permissions
	UserOrganizationPermissionPayload struct {
		PermissionName        string   `json:"permission_name" validate:"required"`
		PermissionDescription string   `json:"permission_description" validate:"required"`
		Permissions           []string `json:"permissions" validate:"required,min=1,dive,required"`
	}

	// DeveloperSecretKeyResponse represents the response containing a developer secret key
	DeveloperSecretKeyResponse struct {
		DeveloperSecretKey string `json:"developer_secret_key"`
	}

	// UserOrganizationStatusRequest represents the request payload for updating user organization status
	UserOrganizationStatusRequest struct {
		UserOrganizationStatus UserOrganizationStatus `json:"user_organization_status" validate:"required,oneof=online offline busy vacation commuting"`
	}

	// UserOrganizationStatusResponse represents the response containing user organization status information
	UserOrganizationStatusResponse struct {
		OfflineUsers   []*UserOrganizationResponse `json:"user_organizations,omitempty"`
		OnlineUsers    []*UserOrganizationResponse `json:"online_user_organizations,omitempty"`
		CommutingUsers []*UserOrganizationResponse `json:"commuting_user_organizations,omitempty"`
		BusyUsers      []*UserOrganizationResponse `json:"busy_user_organizations,omitempty"`
		VacationUsers  []*UserOrganizationResponse `json:"vacation_user_organizations,omitempty"`

		OnlineUsersCount int `json:"online_users_count"`
		OnlineMembers    int `json:"online_members"`
		TotalMembers     int `json:"total_members"`
		OnlineEmployees  int `json:"online_employees"`
		TotalEmployees   int `json:"total_employees"`

		TotalActiveEmployees int                  `json:"total_active_employees"`
		ActiveEmployees      []*TimesheetResponse `json:"active_employees,omitempty"`
	}
)

// UserOrganization initializes the UserOrganization model and its repository manager
func (m *ModelCore) userOrganization() {
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

			"SettingsAccountingPaymentDefaultValue",
			"SettingsAccountingDepositDefaultValue",
			"SettingsAccountingWithdrawDefaultValue",
			"SettingsPaymentTypeDefaultValue",

			"Branch.BranchSetting",
			"Branch.BranchSetting.Currency",

			"Branch.BranchSetting.CashOnHandAccount",
			"Branch.BranchSetting.CashOnHandAccount.Currency",
			"Branch.BranchSetting.PaidUpSharedCapitalAccount",
			"Branch.BranchSetting.PaidUpSharedCapitalAccount.Currency",

			"Branch.BranchSetting.UnbalancedAccounts.Currency",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForOverage",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForOverage",
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

				UserSettingNumberPadding:             data.UserSettingNumberPadding,
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

				SettingsAccountingPaymentDefaultValueID:  data.SettingsAccountingPaymentDefaultValueID,
				SettingsAccountingPaymentDefaultValue:    m.AccountManager.ToModel(data.SettingsAccountingPaymentDefaultValue),
				SettingsAccountingDepositDefaultValueID:  data.SettingsAccountingDepositDefaultValueID,
				SettingsAccountingDepositDefaultValue:    m.AccountManager.ToModel(data.SettingsAccountingDepositDefaultValue),
				SettingsAccountingWithdrawDefaultValueID: data.SettingsAccountingWithdrawDefaultValueID,
				SettingsAccountingWithdrawDefaultValue:   m.AccountManager.ToModel(data.SettingsAccountingWithdrawDefaultValue),
				SettingsPaymentTypeDefaultValueID:        data.SettingsPaymentTypeDefaultValueID,
				SettingsPaymentTypeDefaultValue:          m.PaymentTypeManager.ToModel(data.SettingsPaymentTypeDefaultValue),
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

// GetUserOrganizationByUser retrieves all user organizations for a specific user
func (m *ModelCore) GetUserOrganizationByUser(context context.Context, userID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		UserID: userID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}

// GetUserOrganizationByOrganization retrieves all user organizations for a specific organization
func (m *ModelCore) GetUserOrganizationByOrganization(context context.Context, organizationID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}

// GetUserOrganizationByBranch retrieves all user organizations for a specific organization branch
func (m *ModelCore) GetUserOrganizationBybranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return m.UserOrganizationManager.Find(context, filter)
}

// CountUserOrganizationPerBranch counts the number of user organizations for a specific branch
func (m *ModelCore) CountUserOrganizationPerbranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return m.UserOrganizationManager.Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}

// CountUserOrganizationBranch counts user organizations for a specific user in a branch
func (m *ModelCore) CountUserOrganizationbranch(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return m.UserOrganizationManager.Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
}

// UserOrganizationEmployeeCanJoin checks if a user can join an organization as an employee
func (m *ModelCore) UserOrganizationEmployeeCanJoin(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := m.CountUserOrganizationbranch(context, userID, organizationID, branchID)
	return err == nil && existing == 0
}

// UserOrganizationMemberCanJoin checks if a user can join an organization as a member
func (m *ModelCore) UserOrganizationMemberCanJoin(context context.Context, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := m.CountUserOrganizationbranch(context, userID, organizationID, branchID)
	if err != nil || existing > 0 {
		return false
	}
	existingOrgCount, err := m.UserOrganizationManager.Count(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	return err == nil && existingOrgCount == 0
}

// Employees retrieves all employee user organizations for the specified organization and branch
func (m *ModelCore) Employees(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return m.UserOrganizationManager.Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       UserOrganizationTypeEmployee,
	})
}

// Members retrieves all member user organizations for the specified organization and branch
func (m *ModelCore) Members(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return m.UserOrganizationManager.Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       UserOrganizationTypeMember,
	})
}
