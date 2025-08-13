package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	BranchSetting struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		// Withdraw Settings
		WithdrawAllowUserInput bool   `gorm:"not null;default:true" json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `gorm:"type:varchar(50);not null;default:'WD'" json:"withdraw_prefix"`
		WithdrawORStart        int    `gorm:"not null;default:1" json:"withdraw_or_start"`
		WithdrawORCurrent      int    `gorm:"not null;default:1" json:"withdraw_or_current"`
		WithdrawOREnd          int    `gorm:"not null;default:999999" json:"withdraw_or_end"`
		WithdrawORIteration    int    `gorm:"not null;default:1" json:"withdraw_or_iteration"`
		WithdrawORUnique       bool   `gorm:"not null;default:true" json:"withdraw_or_unique"`
		WithdrawUseDateOR      bool   `gorm:"not null;default:false" json:"withdraw_use_date_or"`

		// Deposit Settings
		DepositAllowUserInput bool   `gorm:"not null;default:true" json:"deposit_allow_user_input"`
		DepositPrefix         string `gorm:"type:varchar(50);not null;default:'DP'" json:"deposit_prefix"`
		DepositORStart        int    `gorm:"not null;default:1" json:"deposit_or_start"`
		DepositORCurrent      int    `gorm:"not null;default:1" json:"deposit_or_current"`
		DepositOREnd          int    `gorm:"not null;default:999999" json:"deposit_or_end"`
		DepositORIteration    int    `gorm:"not null;default:1" json:"deposit_or_iteration"`
		DepositORUnique       bool   `gorm:"not null;default:true" json:"deposit_or_unique"`
		DepositUseDateOR      bool   `gorm:"not null;default:false" json:"deposit_use_date_or"`

		// Loan Settings
		LoanAllowUserInput bool   `gorm:"not null;default:true" json:"loan_allow_user_input"`
		LoanPrefix         string `gorm:"type:varchar(50);not null;default:'LN'" json:"loan_prefix"`
		LoanORStart        int    `gorm:"not null;default:1" json:"loan_or_start"`
		LoanORCurrent      int    `gorm:"not null;default:1" json:"loan_or_current"`
		LoanOREnd          int    `gorm:"not null;default:999999" json:"loan_or_end"`
		LoanORIteration    int    `gorm:"not null;default:1" json:"loan_or_iteration"`
		LoanORUnique       bool   `gorm:"not null;default:true" json:"loan_or_unique"`
		LoanUseDateOR      bool   `gorm:"not null;default:false" json:"loan_use_date_or"`

		// Check Voucher Settings
		CheckVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"check_voucher_allow_user_input"`
		CheckVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'CV'" json:"check_voucher_prefix"`
		CheckVoucherORStart        int    `gorm:"not null;default:1" json:"check_voucher_or_start"`
		CheckVoucherORCurrent      int    `gorm:"not null;default:1" json:"check_voucher_or_current"`
		CheckVoucherOREnd          int    `gorm:"not null;default:999999" json:"check_voucher_or_end"`
		CheckVoucherORIteration    int    `gorm:"not null;default:1" json:"check_voucher_or_iteration"`
		CheckVoucherORUnique       bool   `gorm:"not null;default:true" json:"check_voucher_or_unique"`
		CheckVoucherUseDateOR      bool   `gorm:"not null;default:false" json:"check_voucher_use_date_or"`

		// Default Member Type
		DefaultMemberTypeID *uuid.UUID  `gorm:"type:uuid" json:"default_member_type_id,omitempty"`
		DefaultMemberType   *MemberType `gorm:"foreignKey:DefaultMemberTypeID;constraint:OnDelete:SET NULL;" json:"default_member_type,omitempty"`
	}

	BranchSettingRequest struct {
		// Withdraw Settings
		WithdrawAllowUserInput bool   `json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `json:"withdraw_prefix" validate:"omitempty"`
		WithdrawORStart        int    `json:"withdraw_or_start" validate:"min=0"`
		WithdrawORCurrent      int    `json:"withdraw_or_current" validate:"min=0"`
		WithdrawOREnd          int    `json:"withdraw_or_end" validate:"min=0"`
		WithdrawORIteration    int    `json:"withdraw_or_iteration" validate:"min=0"`
		WithdrawORUnique       bool   `json:"withdraw_or_unique"`
		WithdrawUseDateOR      bool   `json:"withdraw_use_date_or"`

		// Deposit Settings
		DepositAllowUserInput bool   `json:"deposit_allow_user_input"`
		DepositPrefix         string `json:"deposit_prefix" validate:"omitempty"`
		DepositORStart        int    `json:"deposit_or_start" validate:"min=0"`
		DepositORCurrent      int    `json:"deposit_or_current" validate:"min=0"`
		DepositOREnd          int    `json:"deposit_or_end" validate:"min=0"`
		DepositORIteration    int    `json:"deposit_or_iteration" validate:"min=0"`
		DepositORUnique       bool   `json:"deposit_or_unique"`
		DepositUseDateOR      bool   `json:"deposit_use_date_or"`

		// Loan Settings
		LoanAllowUserInput bool   `json:"loan_allow_user_input"`
		LoanPrefix         string `json:"loan_prefix" validate:"omitempty"`
		LoanORStart        int    `json:"loan_or_start" validate:"min=0"`
		LoanORCurrent      int    `json:"loan_or_current" validate:"min=0"`
		LoanOREnd          int    `json:"loan_or_end" validate:"min=0"`
		LoanORIteration    int    `json:"loan_or_iteration" validate:"min=0"`
		LoanORUnique       bool   `json:"loan_or_unique"`
		LoanUseDateOR      bool   `json:"loan_use_date_or"`

		// Check Voucher Settings
		CheckVoucherAllowUserInput bool   `json:"check_voucher_allow_user_input"`
		CheckVoucherPrefix         string `json:"check_voucher_prefix" validate:"omitempty"`
		CheckVoucherORStart        int    `json:"check_voucher_or_start" validate:"min=0"`
		CheckVoucherORCurrent      int    `json:"check_voucher_or_current" validate:"min=0"`
		CheckVoucherOREnd          int    `json:"check_voucher_or_end" validate:"min=0"`
		CheckVoucherORIteration    int    `json:"check_voucher_or_iteration" validate:"min=0"`
		CheckVoucherORUnique       bool   `json:"check_voucher_or_unique"`
		CheckVoucherUseDateOR      bool   `json:"check_voucher_use_date_or"`

		// Default Member Type
		DefaultMemberTypeID *uuid.UUID `json:"default_member_type_id,omitempty"`
	}

	BranchSettingResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		BranchID  uuid.UUID `json:"branch_id"`

		// Withdraw Settings
		WithdrawAllowUserInput bool   `json:"withdraw_allow_user_input"`
		WithdrawPrefix         string `json:"withdraw_prefix"`
		WithdrawORStart        int    `json:"withdraw_or_start"`
		WithdrawORCurrent      int    `json:"withdraw_or_current"`
		WithdrawOREnd          int    `json:"withdraw_or_end"`
		WithdrawORIteration    int    `json:"withdraw_or_iteration"`
		WithdrawORUnique       bool   `json:"withdraw_or_unique"`
		WithdrawUseDateOR      bool   `json:"withdraw_use_date_or"`

		// Deposit Settings
		DepositAllowUserInput bool   `json:"deposit_allow_user_input"`
		DepositPrefix         string `json:"deposit_prefix"`
		DepositORStart        int    `json:"deposit_or_start"`
		DepositORCurrent      int    `json:"deposit_or_current"`
		DepositOREnd          int    `json:"deposit_or_end"`
		DepositORIteration    int    `json:"deposit_or_iteration"`
		DepositORUnique       bool   `json:"deposit_or_unique"`
		DepositUseDateOR      bool   `json:"deposit_use_date_or"`

		// Loan Settings
		LoanAllowUserInput bool   `json:"loan_allow_user_input"`
		LoanPrefix         string `json:"loan_prefix"`
		LoanORStart        int    `json:"loan_or_start"`
		LoanORCurrent      int    `json:"loan_or_current"`
		LoanOREnd          int    `json:"loan_or_end"`
		LoanORIteration    int    `json:"loan_or_iteration"`
		LoanORUnique       bool   `json:"loan_or_unique"`
		LoanUseDateOR      bool   `json:"loan_use_date_or"`

		// Check Voucher Settings
		CheckVoucherAllowUserInput bool   `json:"check_voucher_allow_user_input"`
		CheckVoucherPrefix         string `json:"check_voucher_prefix"`
		CheckVoucherORStart        int    `json:"check_voucher_or_start"`
		CheckVoucherORCurrent      int    `json:"check_voucher_or_current"`
		CheckVoucherOREnd          int    `json:"check_voucher_or_end"`
		CheckVoucherORIteration    int    `json:"check_voucher_or_iteration"`
		CheckVoucherORUnique       bool   `json:"check_voucher_or_unique"`
		CheckVoucherUseDateOR      bool   `json:"check_voucher_use_date_or"`

		// Default Member Type
		DefaultMemberTypeID *uuid.UUID          `json:"default_member_type_id,omitempty"`
		DefaultMemberType   *MemberTypeResponse `json:"default_member_type,omitempty"`
	}
)

func (m *Model) BranchSetting() {
	m.Migration = append(m.Migration, &BranchSetting{})
	m.BranchSettingManager = horizon_services.NewRepository(horizon_services.RepositoryParams[BranchSetting, BranchSettingResponse, BranchSettingRequest]{
		Preloads: []string{
			"Branch",
			"DefaultMemberType",
		},
		Service: m.provider.Service,
		Resource: func(data *BranchSetting) *BranchSettingResponse {
			if data == nil {
				return nil
			}
			return &BranchSettingResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),
				BranchID:  data.BranchID,

				WithdrawAllowUserInput: data.WithdrawAllowUserInput,
				WithdrawPrefix:         data.WithdrawPrefix,
				WithdrawORStart:        data.WithdrawORStart,
				WithdrawORCurrent:      data.WithdrawORCurrent,
				WithdrawOREnd:          data.WithdrawOREnd,
				WithdrawORIteration:    data.WithdrawORIteration,
				WithdrawORUnique:       data.WithdrawORUnique,
				WithdrawUseDateOR:      data.WithdrawUseDateOR,

				DepositAllowUserInput: data.DepositAllowUserInput,
				DepositPrefix:         data.DepositPrefix,
				DepositORStart:        data.DepositORStart,
				DepositORCurrent:      data.DepositORCurrent,
				DepositOREnd:          data.DepositOREnd,
				DepositORIteration:    data.DepositORIteration,
				DepositORUnique:       data.DepositORUnique,
				DepositUseDateOR:      data.DepositUseDateOR,

				LoanAllowUserInput: data.LoanAllowUserInput,
				LoanPrefix:         data.LoanPrefix,
				LoanORStart:        data.LoanORStart,
				LoanORCurrent:      data.LoanORCurrent,
				LoanOREnd:          data.LoanOREnd,
				LoanORIteration:    data.LoanORIteration,
				LoanORUnique:       data.LoanORUnique,
				LoanUseDateOR:      data.LoanUseDateOR,

				CheckVoucherAllowUserInput: data.CheckVoucherAllowUserInput,
				CheckVoucherPrefix:         data.CheckVoucherPrefix,
				CheckVoucherORStart:        data.CheckVoucherORStart,
				CheckVoucherORCurrent:      data.CheckVoucherORCurrent,
				CheckVoucherOREnd:          data.CheckVoucherOREnd,
				CheckVoucherORIteration:    data.CheckVoucherORIteration,
				CheckVoucherORUnique:       data.CheckVoucherORUnique,
				CheckVoucherUseDateOR:      data.CheckVoucherUseDateOR,

				DefaultMemberTypeID: data.DefaultMemberTypeID,
				DefaultMemberType:   m.MemberTypeManager.ToModel(data.DefaultMemberType),
			}
		},
		Created: func(data *BranchSetting) []string {
			return []string{
				"branch_setting.create",
				fmt.Sprintf("branch_setting.create.%s", data.ID),
				fmt.Sprintf("branch_setting.create.branch.%s", data.BranchID),
			}
		},
		Updated: func(data *BranchSetting) []string {
			return []string{
				"branch_setting.update",
				fmt.Sprintf("branch_setting.update.%s", data.ID),
				fmt.Sprintf("branch_setting.update.branch.%s", data.BranchID),
			}
		},
		Deleted: func(data *BranchSetting) []string {
			return []string{
				"branch_setting.delete",
				fmt.Sprintf("branch_setting.delete.%s", data.ID),
				fmt.Sprintf("branch_setting.delete.branch.%s", data.BranchID),
			}
		},
	})
}
