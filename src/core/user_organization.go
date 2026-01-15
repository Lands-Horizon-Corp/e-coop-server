package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
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

type UserOrganizationType string

const (
	UserOrganizationTypeOwner    UserOrganizationType = "owner"
	UserOrganizationTypeEmployee UserOrganizationType = "employee"
	UserOrganizationTypeMember   UserOrganizationType = "member"
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

		UserSettingDescription string `gorm:"type:text" json:"user_setting_description"`

		PaymentORUnique         bool   `gorm:"not null;default:false" json:"payment_or_unique"`
		PaymentORAllowUserInput bool   `gorm:"not null;default:true" json:"payment_or_allow_user_input"`
		PaymentORCurrent        int64  `gorm:"not null;default:1" json:"payment_or_current"`
		PaymentOREnd            int64  `gorm:"not null;default:9999" json:"payment_or_end"`
		PaymentORStart          int64  `gorm:"not null;default:1" json:"payment_or_start"`
		PaymentORIteration      int64  `gorm:"not null;default:1" json:"payment_or_iteration"`
		PaymentORUseDateOR      bool   `gorm:"not null;default:false" json:"payment_or_use_date_or"`
		PaymentPrefix           string `gorm:"type:varchar(50);default:''" json:"payment_prefix"`
		PaymentPadding          int    `gorm:"not null;default:6" json:"payment_padding"`

		SettingsAllowWithdrawNegativeBalance bool `gorm:"not null;default:false" json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `gorm:"not null;default:false" json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `gorm:"not null;default:false" json:"maintaining_balance"`

		Status       UserOrganizationStatus `gorm:"type:varchar(50);not null;default:'offline'" json:"status"`
		LastOnlineAt time.Time              `gorm:"default:now()" json:"last_online_at"`

		TimeMachineTime *time.Time `gorm:"type:timestamp" json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingPaymentDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingPaymentDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_payment_default_value,omitempty"`

		SettingsAccountingDepositDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingDepositDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_deposit_default_value,omitempty"`

		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `gorm:"type:uuid;index" json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValue   *Account   `gorm:"foreignKey:SettingsAccountingWithdrawDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_accounting_withdraw_default_value,omitempty"`

		SettingsPaymentTypeDefaultValueID *uuid.UUID   `gorm:"type:uuid;index" json:"settings_payment_type_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValue   *PaymentType `gorm:"foreignKey:SettingsPaymentTypeDefaultValueID;constraint:OnDelete:SET NULL;" json:"settings_payment_type_default_value,omitempty"`
	}

	EmployeeCreateRequest struct {
		FirstName  string `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName string `json:"middle_name,omitempty" validate:"max=255"`
		LastName   string `json:"last_name" validate:"required,min=1,max=255"`
		FullName   string `json:"full_name,omitempty" validate:"max=255"`
		Suffix     string `json:"suffix,omitempty" validate:"max=50"`

		BirthDate     *time.Time `json:"birthdate" validate:"required"`
		ContactNumber string     `json:"contact_number,omitempty" validate:"max=255"`

		Username string `json:"user_name" validate:"required,min=1,max=255"`
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required,min=6,max=128"`

		MediaID *uuid.UUID `json:"media_id,omitempty"`

		ApplicationDescription string `json:"application_description,omitempty"`

		PermissionName        string   `json:"permission_name" validate:"required"`
		PermissionDescription string   `json:"permission_description" validate:"required"`
		Permissions           []string `json:"permissions" validate:"omitempty,dive,min=1"`
	}
)

func (uo *UserOrganization) UserOrgTime() time.Time {
	if uo.TimeMachineTime != nil && !uo.TimeMachineTime.IsZero() {
		if uo.Branch != nil && uo.Branch.Currency != nil && uo.Branch.Currency.Timezone != "" {
			loc, err := time.LoadLocation(uo.Branch.Currency.Timezone)
			if err != nil {
				return uo.TimeMachineTime.UTC()
			}
			localTime := time.Date(
				uo.TimeMachineTime.Year(),
				uo.TimeMachineTime.Month(),
				uo.TimeMachineTime.Day(),
				uo.TimeMachineTime.Hour(),
				uo.TimeMachineTime.Minute(),
				uo.TimeMachineTime.Second(),
				uo.TimeMachineTime.Nanosecond(),
				loc,
			)
			return localTime.UTC()
		}
		return uo.TimeMachineTime.UTC()
	}
	return time.Now().UTC()
}

type (
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

		PaymentORUnique         bool  `json:"payment_or_unique,omitempty"`
		PaymentORAllowUserInput bool  `json:"payment_or_allow_user_input,omitempty"`
		PaymentORCurrent        int64 `json:"payment_or_current,omitempty" validate:"min=1"`
		PaymentORStart          int64 `json:"payment_or_start"`
		PaymentOREnd            int64 `json:"payment_or_end,omitempty" validate:"min=1"`
		PaymentORIteration      int64 `json:"payment_or_iteration,omitempty" validate:"min=1"`
		PaymentORUseDateOR      bool  `json:"payment_or_use_date_or,omitempty"`
		PaymentPrefix           bool  `json:"payment_prefix,omitempty"`
		PaymentPadding          int   `json:"payment_padding,omitempty" validate:"min=0"`
	}

	UserOrganizationSettingsRequest struct {
		UserType    UserOrganizationType `json:"user_type,omitempty" validate:"omitempty,oneof=employee member"`
		Description string               `json:"description,omitempty"`

		ApplicationDescription string `json:"application_description,omitempty"`
		ApplicationStatus      string `json:"application_status" validate:"omitempty,oneof=pending reported accepted ban not-allowed"`

		UserSettingDescription string `json:"user_setting_description,omitempty"`

		PaymentORUnique         bool   `json:"payment_or_unique,omitempty"`
		PaymentORAllowUserInput bool   `json:"payment_or_allow_user_input,omitempty"`
		PaymentORCurrent        int64  `json:"payment_or_current,omitempty" validate:"min=1"`
		PaymentOREnd            int64  `json:"payment_or_end,omitempty" validate:"min=1"`
		PaymentORStart          int64  `json:"payment_or_start,omitempty" validate:"min=1"`
		PaymentORIteration      int64  `json:"payment_or_iteration,omitempty" validate:"min=1"`
		PaymentORUseDateOR      bool   `json:"payment_or_use_date_or,omitempty"`
		PaymentPrefix           string `json:"payment_prefix,omitempty"`
		PaymentPadding          int    `json:"payment_padding,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		TimeMachineTime *time.Time `json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`
	}

	UserOrganizationSelfSettingsRequest struct {
		Description            string `json:"description,omitempty"`
		UserSettingDescription string `json:"user_setting_description,omitempty"`

		PaymentORUnique         bool   `json:"payment_or_unique,omitempty"`
		PaymentORAllowUserInput bool   `json:"payment_or_allow_user_input,omitempty"`
		PaymentORCurrent        int64  `json:"payment_or_current,omitempty" validate:"min=1"`
		PaymentORStart          int64  `json:"payment_or_start,omitempty" validate:"min=1"`
		PaymentOREnd            int64  `json:"payment_or_end,omitempty" validate:"min=1"`
		PaymentORIteration      int64  `json:"payment_or_iteration,omitempty" validate:"min=1"`
		PaymentORUseDateOR      bool   `json:"payment_or_use_date_or,omitempty"`
		PaymentPrefix           string `json:"payment_prefix,omitempty"`
		PaymentPadding          int    `json:"payment_padding,omitempty" validate:"min=0"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		TimeMachineTime *time.Time `json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`
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

		PaymentORUnique         bool   `json:"payment_or_unique"`
		PaymentORAllowUserInput bool   `json:"payment_or_allow_user_input"`
		PaymentORCurrent        int64  `json:"payment_or_current"`
		PaymentORStart          int64  `json:"payment_or_start,omitempty" validate:"min=1"`
		PaymentOREnd            int64  `json:"payment_or_end"`
		PaymentORIteration      int64  `json:"payment_or_iteration"`
		PaymentORUseDateOR      bool   `json:"payment_or_use_date_or"`
		PaymentPrefix           string `json:"payment_prefix"`
		PaymentPadding          int    `json:"payment_padding"`

		SettingsAllowWithdrawNegativeBalance bool `json:"allow_withdraw_negative_balance"`
		SettingsAllowWithdrawExactBalance    bool `json:"allow_withdraw_exact_balance"`
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		Status       UserOrganizationStatus `json:"status"`
		LastOnlineAt time.Time              `json:"last_online_at"`

		TimeMachineTime *time.Time `json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID *uuid.UUID       `json:"settings_accounting_payment_default_value_id"`
		SettingsAccountingPaymentDefaultValue   *AccountResponse `json:"settings_accounting_payment_default_value,omitempty"`

		SettingsAccountingDepositDefaultValueID *uuid.UUID       `json:"settings_accounting_deposit_default_value_id"`
		SettingsAccountingDepositDefaultValue   *AccountResponse `json:"settings_accounting_deposit_default_value,omitempty"`

		SettingsAccountingWithdrawDefaultValueID *uuid.UUID       `json:"settings_accounting_withdraw_default_value_id"`
		SettingsAccountingWithdrawDefaultValue   *AccountResponse `json:"settings_accounting_withdraw_default_value,omitempty"`

		SettingsPaymentTypeDefaultValueID *uuid.UUID           `json:"settings_payment_type_default_value_id"`
		SettingsPaymentTypeDefaultValue   *PaymentTypeResponse `json:"settings_payment_type_default_value,omitempty"`
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

		OnlineUsersCount int `json:"online_users_count"`
		OnlineMembers    int `json:"online_members"`
		TotalMembers     int `json:"total_members"`
		OnlineEmployees  int `json:"online_employees"`
		TotalEmployees   int `json:"total_employees"`

		TotalActiveEmployees int                  `json:"total_active_employees"`
		ActiveEmployees      []*TimesheetResponse `json:"active_employees,omitempty"`
	}
)

func UserOrganizationManager(service *horizon.HorizonService) *registry.Registry[UserOrganization, UserOrganizationResponse, UserOrganizationRequest] {
	return registry.NewRegistry(registry.RegistryParams[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]{
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

			"Branch.BranchSetting.CompassionFundAccount",
			"Branch.BranchSetting.CompassionFundAccount.Currency",

			"Branch.BranchSetting.UnbalancedAccounts.Currency",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.AccountForOverage",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForShortage",
			"Branch.BranchSetting.UnbalancedAccounts.MemberProfileForOverage",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
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
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),

				UserID:                 data.UserID,
				User:                   UserManager(service).ToModel(data.User),
				UserType:               data.UserType,
				Description:            data.Description,
				ApplicationDescription: data.ApplicationDescription,
				ApplicationStatus:      data.ApplicationStatus,
				DeveloperSecretKey:     "",
				PermissionName:         data.PermissionName,
				PermissionDescription:  data.PermissionDescription,
				Permissions:            data.Permissions,

				UserSettingDescription: data.UserSettingDescription,

				PaymentORUnique:                      data.PaymentORUnique,
				PaymentORAllowUserInput:              data.PaymentORAllowUserInput,
				PaymentORCurrent:                     data.PaymentORCurrent,
				PaymentORStart:                       data.PaymentORStart,
				PaymentOREnd:                         data.PaymentOREnd,
				PaymentORIteration:                   data.PaymentORIteration,
				PaymentORUseDateOR:                   data.PaymentORUseDateOR,
				PaymentPrefix:                        data.PaymentPrefix,
				PaymentPadding:                       data.PaymentPadding,
				SettingsAllowWithdrawNegativeBalance: data.SettingsAllowWithdrawNegativeBalance,
				SettingsAllowWithdrawExactBalance:    data.SettingsAllowWithdrawExactBalance,
				SettingsMaintainingBalance:           data.SettingsMaintainingBalance,
				Status:                               data.Status,
				LastOnlineAt:                         data.LastOnlineAt,
				TimeMachineTime:                      data.TimeMachineTime,

				SettingsAccountingPaymentDefaultValueID:  data.SettingsAccountingPaymentDefaultValueID,
				SettingsAccountingPaymentDefaultValue:    AccountManager(service).ToModel(data.SettingsAccountingPaymentDefaultValue),
				SettingsAccountingDepositDefaultValueID:  data.SettingsAccountingDepositDefaultValueID,
				SettingsAccountingDepositDefaultValue:    AccountManager(service).ToModel(data.SettingsAccountingDepositDefaultValue),
				SettingsAccountingWithdrawDefaultValueID: data.SettingsAccountingWithdrawDefaultValueID,
				SettingsAccountingWithdrawDefaultValue:   AccountManager(service).ToModel(data.SettingsAccountingWithdrawDefaultValue),
				SettingsPaymentTypeDefaultValueID:        data.SettingsPaymentTypeDefaultValueID,
				SettingsPaymentTypeDefaultValue:          PaymentTypeManager(service).ToModel(data.SettingsPaymentTypeDefaultValue),
			}
		},
		Created: func(data *UserOrganization) registry.Topics {
			return []string{
				"user_organization.create",
				fmt.Sprintf("user_organization.create.%s", data.ID),
				fmt.Sprintf("user_organization.create.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *UserOrganization) registry.Topics {
			return []string{
				"user_organization.update",
				fmt.Sprintf("user_organization.update.%s", data.ID),
				fmt.Sprintf("user_organization.update.branch.%s", data.BranchID),
				fmt.Sprintf("user_organization.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("user_organization.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *UserOrganization) registry.Topics {
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

func GetUserOrganizationByUser(context context.Context, service *horizon.HorizonService, userID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		UserID: userID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func GetUserOrganizationByOrganization(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func GetUserOrganizationByBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID, pending *bool) ([]*UserOrganization, error) {
	filter := &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	}
	if pending != nil && *pending {
		filter.ApplicationStatus = "pending"
	}
	return UserOrganizationManager(service).Find(context, filter)
}

func CountUserOrganizationPerBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return UserOrganizationManager(service).Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}

func CountUserOrganizationbranch(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (int64, error) {
	return UserOrganizationManager(service).Count(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserID:         userID,
	})
}

func UserOrganizationEmployeeCanJoin(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := CountUserOrganizationbranch(context, service, userID, organizationID, branchID)
	return err == nil && existing == 0
}

func UserOrganizationMemberCanJoin(context context.Context, service *horizon.HorizonService, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) bool {
	existing, err := CountUserOrganizationbranch(context, service, userID, organizationID, branchID)
	if err != nil || existing > 0 {
		return false
	}
	existingOrgCount, err := UserOrganizationManager(service).Count(context, &UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	return err == nil && existingOrgCount == 0
}

func Employees(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return UserOrganizationManager(service).Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       UserOrganizationTypeEmployee,
	})
}

func Members(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*UserOrganization, error) {
	return UserOrganizationManager(service).Find(context, &UserOrganization{
		OrganizationID: organizationID,
		BranchID:       &branchID,
		UserType:       UserOrganizationTypeMember,
	})
}
