package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	UnbalancedAccount struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		BranchSettingsID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_unique_currency_per_branch_settings,priority:2" json:"branch_settings_id"`
		BranchSettings   *BranchSetting `gorm:"foreignKey:BranchSettingsID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch_settings,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_currency_per_branch_settings,priority:1" json:"currency_id"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		AccountForShortageID uuid.UUID `gorm:"type:uuid;not null" json:"account_for_shortage_id"`
		AccountForShortage   *Account  `gorm:"foreignKey:AccountForShortageID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account_for_shortage,omitempty"`

		AccountForOverageID uuid.UUID `gorm:"type:uuid;not null" json:"account_for_overage_id"`
		AccountForOverage   *Account  `gorm:"foreignKey:AccountForOverageID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account_for_overage,omitempty"`

		MemberProfileIDForShortage *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id_for_shortage"`
		MemberProfileForShortage   *MemberProfile `gorm:"foreignKey:MemberProfileIDForShortage;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"member_profile_for_shortage,omitempty"`

		MemberProfileIDForOverage *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id_for_overage"`
		MemberProfileForOverage   *MemberProfile `gorm:"foreignKey:MemberProfileIDForOverage;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"member_profile_for_overage,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	UnbalancedAccountResponse struct {
		ID               uuid.UUID              `json:"id"`
		CreatedAt        string                 `json:"created_at"`
		CreatedByID      uuid.UUID              `json:"created_by_id"`
		CreatedBy        *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt        string                 `json:"updated_at"`
		UpdatedByID      uuid.UUID              `json:"updated_by_id"`
		UpdatedBy        *UserResponse          `json:"updated_by,omitempty"`
		BranchSettingsID uuid.UUID              `json:"branch_settings_id"`
		BranchSettings   *BranchSettingResponse `json:"branch_settings,omitempty"`
		CurrencyID       uuid.UUID              `json:"currency_id"`
		Currency         *CurrencyResponse      `json:"currency,omitempty"`

		AccountForShortageID uuid.UUID        `json:"account_for_shortage_id"`
		AccountForShortage   *AccountResponse `json:"account_for_shortage,omitempty"`

		AccountForOverageID uuid.UUID        `json:"account_for_overage_id"`
		AccountForOverage   *AccountResponse `json:"account_for_overage,omitempty"`

		MemberProfileIDForShortage *uuid.UUID             `json:"member_profile_id_for_shortage,omitempty"`
		MemberProfileForShortage   *MemberProfileResponse `json:"member_profile_for_shortage,omitempty"`

		MemberProfileIDForOverage *uuid.UUID             `json:"member_profile_id_for_overage,omitempty"`
		MemberProfileForOverage   *MemberProfileResponse `json:"member_profile_for_overage,omitempty"`

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	UnbalancedAccountRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name        string    `json:"name" validate:"omitempty,min=1,max=255"`
		Description string    `json:"description,omitempty"`
		CurrencyID  uuid.UUID `json:"currency_id" validate:"required"`

		AccountForShortageID uuid.UUID `json:"account_for_shortage_id" validate:"required"`
		AccountForOverageID  uuid.UUID `json:"account_for_overage_id" validate:"required"`

		MemberProfileIDForShortage *uuid.UUID `json:"member_profile_id_for_shortage,omitempty"`
		MemberProfileIDForOverage  *uuid.UUID `json:"member_profile_id_for_overage,omitempty"`
	}
)

func (m *Core) UnbalancedAccountManager() *registry.Registry[UnbalancedAccount, UnbalancedAccountResponse, UnbalancedAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[UnbalancedAccount, UnbalancedAccountResponse, UnbalancedAccountRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "BranchSettings", "Currency",
			"AccountForShortage", "AccountForOverage",
			"MemberProfileForShortage", "MemberProfileForOverage",
		},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *UnbalancedAccount) *UnbalancedAccountResponse {
			if data == nil {
				return nil
			}
			return &UnbalancedAccountResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        m.UserManager().ToModel(data.UpdatedBy),
				BranchSettingsID: data.BranchSettingsID,
				BranchSettings:   m.BranchSettingManager().ToModel(data.BranchSettings),
				CurrencyID:       data.CurrencyID,
				Currency:         m.CurrencyManager().ToModel(data.Currency),

				AccountForShortageID: data.AccountForShortageID,
				AccountForShortage:   m.AccountManager().ToModel(data.AccountForShortage),

				AccountForOverageID: data.AccountForOverageID,
				AccountForOverage:   m.AccountManager().ToModel(data.AccountForOverage),

				MemberProfileIDForShortage: data.MemberProfileIDForShortage,
				MemberProfileForShortage:   m.MemberProfileManager().ToModel(data.MemberProfileForShortage),

				MemberProfileIDForOverage: data.MemberProfileIDForOverage,
				MemberProfileForOverage:   m.MemberProfileManager().ToModel(data.MemberProfileForOverage),

				Name:        data.Name,
				Description: data.Description,
			}
		},
		Created: func(data *UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.create",
				fmt.Sprintf("unbalanced_account.create.%s", data.ID),
				fmt.Sprintf("unbalanced_account.create.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.create.currency.%s", data.CurrencyID),
			}
		},
		Updated: func(data *UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.update",
				fmt.Sprintf("unbalanced_account.update.%s", data.ID),
				fmt.Sprintf("unbalanced_account.update.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.update.currency.%s", data.CurrencyID),
			}
		},
		Deleted: func(data *UnbalancedAccount) registry.Topics {
			return []string{
				"unbalanced_account.delete",
				fmt.Sprintf("unbalanced_account.delete.%s", data.ID),
				fmt.Sprintf("unbalanced_account.delete.branch_settings.%s", data.BranchSettingsID),
				fmt.Sprintf("unbalanced_account.delete.currency.%s", data.CurrencyID),
			}
		},
	})
}
