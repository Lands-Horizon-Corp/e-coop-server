package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberBankCard struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_card"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_card"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BankID *uuid.UUID `gorm:"type:uuid"`
		Bank   *Bank      `gorm:"foreignKey:BankID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"bank,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		AccountNumber  string    `gorm:"type:varchar(50);not null"`
		CardName       string    `gorm:"type:varchar(50);not null"`
		ExpirationDate time.Time `gorm:"not null"`
		IsDefault      bool      `gorm:"not null;default:false"`
	}

	MemberBankCardResponse struct {
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

		BankID *uuid.UUID    `json:"bank_id,omitempty"`
		Bank   *BankResponse `json:"bank,omitempty"`

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		AccountNumber  string `json:"account_number"`
		CardName       string `json:"card_name"`
		ExpirationDate string `json:"expiration_date"`
		IsDefault      bool   `json:"is_default"`
	}

	MemberBankCardRequest struct {
		AccountNumber   string     `json:"account_number" validate:"required,min=1,max=50"`
		CardName        string     `json:"card_name" validate:"required,min=1,max=50"`
		ExpirationDate  time.Time  `json:"expiration_date" validate:"required"`
		IsDefault       bool       `json:"is_default"`
		BankID          *uuid.UUID `json:"bank_id,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
	}
)

func (m *Core) MemberBankCardManager() *registry.Registry[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest] {
	return registry.NewRegistry(registry.RegistryParams[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Bank", "MemberProfile"},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberBankCard) *MemberBankCardResponse {
			if data == nil {
				return nil
			}
			return &MemberBankCardResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager().ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager().ToModel(data.Branch),
				BankID:          data.BankID,
				Bank:            m.BankManager().ToModel(data.Bank),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager().ToModel(data.MemberProfile),
				AccountNumber:   data.AccountNumber,
				CardName:        data.CardName,
				ExpirationDate:  data.ExpirationDate.Format(time.RFC3339),
				IsDefault:       data.IsDefault,
			}
		},

		Created: func(data *MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.create",
				fmt.Sprintf("member_bank_card.create.%s", data.ID),
				fmt.Sprintf("member_bank_card.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.update",
				fmt.Sprintf("member_bank_card.update.%s", data.ID),
				fmt.Sprintf("member_bank_card.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberBankCard) registry.Topics {
			return []string{
				"member_bank_card.delete",
				fmt.Sprintf("member_bank_card.delete.%s", data.ID),
				fmt.Sprintf("member_bank_card.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_bank_card.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) MemberBankCardCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberBankCard, error) {
	return m.MemberBankCardManager().Find(context, &MemberBankCard{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
