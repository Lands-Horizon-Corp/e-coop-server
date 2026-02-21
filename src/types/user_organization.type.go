package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	UserOrganizationTypeOwner    UserOrganizationType = "owner"
	UserOrganizationTypeEmployee UserOrganizationType = "employee"
	UserOrganizationTypeMember   UserOrganizationType = "member"

	UserOrganizationStatusOnline    UserOrganizationStatus = "online"
	UserOrganizationStatusOffline   UserOrganizationStatus = "offline"
	UserOrganizationStatusBusy      UserOrganizationStatus = "busy"
	UserOrganizationStatusVacation  UserOrganizationStatus = "vacation"
	UserOrganizationStatusCommuting UserOrganizationStatus = "commuting"
)

type (
	UserOrganizationStatus string
	UserOrganizationType   string
	UserOrganization       struct {
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

		LoanVoucherAutoIncrement      bool `gorm:"not null;default:false" json:"loan_voucher_auto_increment"`
		AdjustmentEntryAutoIncrement  bool `gorm:"not null;default:false" json:"adjustment_entry_auto_increment"`
		JournalVoucherAutoIncrement   bool `gorm:"not null;default:false" json:"journal_voucher_auto_increment"`
		CashCheckVoucherAutoIncrement bool `gorm:"not null;default:false" json:"cash_check_voucher_auto_increment"`
		DepositAutoIncrement          bool `gorm:"not null;default:false" json:"deposit_auto_increment"`
		WithdrawAutoIncrement         bool `gorm:"not null;default:false" json:"withdraw_auto_increment"`
		PaymentAutoIncrement          bool `gorm:"not null;default:false" json:"payment_auto_increment"`
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

func (uo *UserOrganization) TimeMachine() time.Time {
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
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		TimeMachineTime *time.Time `json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`

		LoanVoucherAutoIncrement      bool `json:"loan_voucher_auto_increment"`
		AdjustmentEntryAutoIncrement  bool `json:"adjustment_entry_auto_increment"`
		JournalVoucherAutoIncrement   bool `json:"journal_voucher_auto_increment"`
		CashCheckVoucherAutoIncrement bool `json:"cash_check_voucher_auto_increment"`
		DepositAutoIncrement          bool `json:"deposit_auto_increment"`
		WithdrawAutoIncrement         bool `json:"withdraw_auto_increment"`
		PaymentAutoIncrement          bool `json:"payment_auto_increment"`
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
		SettingsMaintainingBalance           bool `json:"maintaining_balance"`

		TimeMachineTime *time.Time `json:"time_machine_time,omitempty"`

		SettingsAccountingPaymentDefaultValueID  *uuid.UUID `json:"settings_accounting_payment_default_value_id,omitempty"`
		SettingsAccountingDepositDefaultValueID  *uuid.UUID `json:"settings_accounting_deposit_default_value_id,omitempty"`
		SettingsAccountingWithdrawDefaultValueID *uuid.UUID `json:"settings_accounting_withdraw_default_value_id,omitempty"`
		SettingsPaymentTypeDefaultValueID        *uuid.UUID `json:"settings_payment_type_default_value_id,omitempty"`

		LoanVoucherAutoIncrement      bool `json:"loan_voucher_auto_increment"`
		AdjustmentEntryAutoIncrement  bool `json:"adjustment_entry_auto_increment"`
		JournalVoucherAutoIncrement   bool `json:"journal_voucher_auto_increment"`
		CashCheckVoucherAutoIncrement bool `json:"cash_check_voucher_auto_increment"`
		DepositAutoIncrement          bool `json:"deposit_auto_increment"`
		WithdrawAutoIncrement         bool `json:"withdraw_auto_increment"`
		PaymentAutoIncrement          bool `json:"payment_auto_increment"`
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

		LoanVoucherAutoIncrement      bool `json:"loan_voucher_auto_increment"`
		AdjustmentEntryAutoIncrement  bool `json:"adjustment_entry_auto_increment"`
		JournalVoucherAutoIncrement   bool `json:"journal_voucher_auto_increment"`
		CashCheckVoucherAutoIncrement bool `json:"cash_check_voucher_auto_increment"`
		DepositAutoIncrement          bool `json:"deposit_auto_increment"`
		WithdrawAutoIncrement         bool `json:"withdraw_auto_increment"`
		PaymentAutoIncrement          bool `json:"payment_auto_increment"`
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
