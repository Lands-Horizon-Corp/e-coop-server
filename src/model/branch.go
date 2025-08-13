package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Branch struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		MediaID        *uuid.UUID     `gorm:"type:uuid" json:"media_id"`
		Media          *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Type          string   `gorm:"type:varchar(100);not null" json:"type"`
		Name          string   `gorm:"type:varchar(255);not null" json:"name"`
		Email         string   `gorm:"type:varchar(255);not null" json:"email"`
		Description   *string  `gorm:"type:text" json:"description,omitempty"`
		CountryCode   string   `gorm:"type:varchar(10);not null" json:"country_code"`
		ContactNumber *string  `gorm:"type:varchar(20)" json:"contact_number,omitempty"`
		Address       string   `gorm:"type:varchar(500);not null" json:"address"`
		Province      string   `gorm:"type:varchar(100);not null" json:"province"`
		City          string   `gorm:"type:varchar(100);not null" json:"city"`
		Region        string   `gorm:"type:varchar(100);not null" json:"region"`
		Barangay      string   `gorm:"type:varchar(100);not null" json:"barangay"`
		PostalCode    string   `gorm:"type:varchar(20);not null" json:"postal_code"`
		Latitude      *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude     *float64 `gorm:"type:double precision" json:"longitude,omitempty"`
		IsMainBranch  bool     `gorm:"not null;default:false" json:"is_main_branch"`

		Footsteps           []*Footstep           `gorm:"foreignKey:BranchID" json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReport    `gorm:"foreignKey:BranchID" json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCode     `gorm:"foreignKey:BranchID" json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplate `gorm:"foreignKey:BranchID" json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganization   `gorm:"foreignKey:BranchID" json:"user_organizations,omitempty"`

		BranchSettingWithdrawAllowUserInput bool   `gorm:"not null;default:true" json:"branch_setting_withdraw_allow_user_input"`
		BranchSettingWithdrawPrefix         string `gorm:"type:varchar(50);not null;default:'branch_setting_withdraw_or'"`
		BranchSettingWithdrawORStart        int    `gorm:"not null;default:0" json:"branch_setting_withdraw_or_start"`
		BranchSettingWithdrawORCurrent      int    `gorm:"not null;default:0" json:"branch_setting_withdraw_or_current"`
		BranchSettingWithdrawOREnd          int    `gorm:"not null;default:0" json:"branch_setting_withdraw_or_end"`
		BranchSettingWithdrawIteration      int    `gorm:"not null;default:0" json:"branch_setting_withdraw_or_iteration"`
		BranchSettingWithdrawORUnique       bool   `gorm:"not null;default:false" json:"branch_setting_withdraw_or_unique"`
		BranchSettingWithdrawUseDateOR      bool   `gorm:"not null;default:false" json:"branch_setting_withdraw_use_date_or"`

		BranchSettingDepositAllowUserInput bool   `gorm:"not null;default:true" json:"branch_setting_deposit_allow_user_input"`
		BranchSettingDepositPrefix         string `gorm:"type:varchar(50);not null;default:'branch_setting_deposit_or'"`
		BranchSettingDepositORStart        int    `gorm:"not null;default:0" json:"branch_setting_deposit_or_start"`
		BranchSettingDepositORCurrent      int    `gorm:"not null;default:0" json:"branch_setting_deposit_or_current"`
		BranchSettingDepositOREnd          int    `gorm:"not null;default:0" json:"branch_setting_deposit_or_end"`
		BranchSettingDepositORIteration    int    `gorm:"not null;default:0" json:"branch_setting_deposit_or_iteration"`
		BranchSettingDepositORUnique       bool   `gorm:"not null;default:false" json:"branch_setting_deposit_or_unique"`
		BranchSettingDepositUseDateOR      bool   `gorm:"not null;default:false" json:"branch_setting_deposit_use_date_or"`

		BranchSettingLoanAllowUserInput bool   `gorm:"not null;default:true" json:"branch_setting_loan_allow_user_input"`
		BranchSettingLoanPrefix         string `gorm:"type:varchar(50);not null;default:'branch_setting_loan_or'"`
		BranchSettingLoanORStart        int    `gorm:"not null;default:0" json:"branch_setting_loan_or_start"`
		BranchSettingLoanORCurrent      int    `gorm:"not null;default:0" json:"branch_setting_loan_or_current"`
		BranchSettingLoanOREnd          int    `gorm:"not null;default:0" json:"branch_setting_loan_or_end"`
		BranchSettingLoanORIteration    int    `gorm:"not null;default:0" json:"branch_setting_loan_or_iteration"`
		BranchSettingLoanORUnique       bool   `gorm:"not null;default:false" json:"branch_setting_loan_or_unique"`
		BranchSettingLoanUseDateOR      bool   `gorm:"not null;default:false" json:"branch_setting_loan_use_date_or"`

		BranchSettingCheckVoucherAllowUserInput bool   `gorm:"not null;default:true" json:"branch_setting_check_voucher_allow_user_input"`
		BranchSettingCheckVoucherPrefix         string `gorm:"type:varchar(50);not null;default:'branch_setting_check_voucher_or'"`
		BranchSettingCheckVoucherORStart        int    `gorm:"not null;default:0" json:"branch_setting_check_voucher_or_start"`
		BranchSettingCheckVoucherORCurrent      int    `gorm:"not null;default:0" json:"branch_setting_check_voucher_or_current"`
		BranchSettingCheckVoucherOREnd          int    `gorm:"not null;default:0" json:"branch_setting_check_voucher_or_end"`
		BranchSettingCheckVoucherORIteration    int    `gorm:"not null;default:0" json:"branch_setting_check_voucher_or_iteration"`
		BranchSettingCheckVoucherORUnique       bool   `gorm:"not null;default:false" json:"branch_setting_check_voucher_or_unique"`
		BranchSettingCheckVoucherUseDateOR      bool   `gorm:"not null;default:false" json:"branch_setting_check_voucher_use_date_or"`

		// BranchSettingDefaultMemberTypeID *uuid.UUID  `gorm:"type:uuid;index" json:"branch_setting_default_member_type_id,omitempty"`
		// BranchSettingDefaultMemberType   *MemberType `gorm:"foreignKey:BranchSettingDefaultMemberTypeID;constraint:OnDelete:SET NULL;"`
	}
	BranchRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		MediaID       *uuid.UUID `json:"media_id,omitempty"`
		Type          string     `json:"type" validate:"required"`
		Name          string     `json:"name" validate:"required"`
		Email         string     `json:"email" validate:"required,email"`
		Description   *string    `json:"description,omitempty"`
		CountryCode   string     `json:"country_code" validate:"required"`
		ContactNumber *string    `json:"contact_number,omitempty"`
		Address       string     `json:"address" validate:"required"`
		Province      string     `json:"province" validate:"required"`
		City          string     `json:"city" validate:"required"`
		Region        string     `json:"region" validate:"required"`
		Barangay      string     `json:"barangay" validate:"required"`
		PostalCode    string     `json:"postal_code" validate:"required"`
		Latitude      *float64   `json:"latitude,omitempty"`
		Longitude     *float64   `json:"longitude,omitempty"`

		IsMainBranch bool `json:"is_main_branch,omitempty"`
	}

	BranchResponse struct {
		ID           uuid.UUID             `json:"id"`
		CreatedAt    string                `json:"created_at"`
		CreatedByID  uuid.UUID             `json:"created_by_id"`
		CreatedBy    *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt    string                `json:"updated_at"`
		UpdatedByID  uuid.UUID             `json:"updated_by_id"`
		UpdatedBy    *UserResponse         `json:"updated_by,omitempty"`
		Organization *OrganizationResponse `json:"organization,omitempty"`

		MediaID       *uuid.UUID     `json:"media_id,omitempty"`
		Media         *MediaResponse `json:"media,omitempty"`
		Type          string         `json:"type"`
		Name          string         `json:"name"`
		Email         string         `json:"email"`
		Description   *string        `json:"description,omitempty"`
		CountryCode   string         `json:"country_code"`
		ContactNumber *string        `json:"contact_number,omitempty"`
		Address       string         `json:"address"`
		Province      string         `json:"province"`
		City          string         `json:"city"`
		Region        string         `json:"region"`
		Barangay      string         `json:"barangay"`
		PostalCode    string         `json:"postal_code"`
		Latitude      *float64       `json:"latitude,omitempty"`
		Longitude     *float64       `json:"longitude,omitempty"`

		IsMainBranch bool `json:"is_main_branch,omitempty"`

		Footsteps           []*FootstepResponse           `json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReportResponse    `json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCodeResponse     `json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplateResponse `json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganizationResponse   `json:"user_organizations,omitempty"`

		BranchSettingWithdrawAllowUserInput bool   `json:"branch_setting_withdraw_allow_user_input"`
		BranchSettingWithdrawPrefix         string `json:"branch_setting_withdraw_prefix"`
		BranchSettingWithdrawORStart        int    `json:"branch_setting_withdraw_or_start"`
		BranchSettingWithdrawORCurrent      int    `json:"branch_setting_withdraw_or_current"`
		BranchSettingWithdrawOREnd          int    `json:"branch_setting_withdraw_or_end"`
		BranchSettingWithdrawIteration      int    `json:"branch_setting_withdraw_or_iteration"`
		BranchSettingWithdrawORUnique       bool   `json:"branch_setting_withdraw_or_unique"`
		BranchSettingWithdrawUseDateOR      bool   `json:"branch_setting_withdraw_use_date_or"`

		BranchSettingDepositAllowUserInput bool   `json:"branch_setting_deposit_allow_user_input"`
		BranchSettingDepositPrefix         string `json:"branch_setting_deposit_prefix"`
		BranchSettingDepositORStart        int    `json:"branch_setting_deposit_or_start"`
		BranchSettingDepositORCurrent      int    `json:"branch_setting_deposit_or_current"`
		BranchSettingDepositOREnd          int    `json:"branch_setting_deposit_or_end"`
		BranchSettingDepositORIteration    int    `json:"branch_setting_deposit_or_iteration"`
		BranchSettingDepositORUnique       bool   `json:"branch_setting_deposit_or_unique"`
		BranchSettingDepositUseDateOR      bool   `json:"branch_setting_deposit_use_date_or"`

		BranchSettingLoanAllowUserInput bool   `json:"branch_setting_loan_allow_user_input"`
		BranchSettingLoanPrefix         string `json:"branch_setting_loan_prefix"`
		BranchSettingLoanORStart        int    `json:"branch_setting_loan_or_start"`
		BranchSettingLoanORCurrent      int    `json:"branch_setting_loan_or_current"`
		BranchSettingLoanOREnd          int    `json:"branch_setting_loan_or_end"`
		BranchSettingLoanORIteration    int    `json:"branch_setting_loan_or_iteration"`
		BranchSettingLoanORUnique       bool   `json:"branch_setting_loan_or_unique"`
		BranchSettingLoanUseDateOR      bool   `json:"branch_setting_loan_use_date_or"`

		BranchSettingCheckVoucherAllowUserInput bool   `json:"branch_setting_check_voucher_allow_user_input"`
		BranchSettingCheckVoucherPrefix         string `json:"branch_setting_check_voucher_prefix"`
		BranchSettingCheckVoucherORStart        int    `json:"branch_setting_check_voucher_or_start"`
		BranchSettingCheckVoucherORCurrent      int    `json:"branch_setting_check_voucher_or_current"`
		BranchSettingCheckVoucherOREnd          int    `json:"branch_setting_check_voucher_or_end"`
		BranchSettingCheckVoucherORIteration    int    `json:"branch_setting_check_voucher_or_iteration"`
		BranchSettingCheckVoucherORUnique       bool   `json:"branch_setting_check_voucher_or_unique"`
		BranchSettingCheckVoucherUseDateOR      bool   `json:"branch_setting_check_voucher_use_date_or"`

		BranchSettingDefaultMemberTypeID *uuid.UUID          `json:"branch_setting_default_member_type_id,omitempty"`
		BranchSettingDefaultMemberType   *MemberTypeResponse `json:"branch_setting_default_member_type,omitempty"`
	}

	BranchSettingRequest struct {
		BranchSettingWithdrawAllowUserInput bool   `json:"branch_setting_withdraw_allow_user_input"`
		BranchSettingWithdrawPrefix         string `json:"branch_setting_withdraw_prefix" validate:"omitempty"`
		BranchSettingWithdrawORStart        int    `json:"branch_setting_withdraw_or_start" validate:"min=0"`
		BranchSettingWithdrawORCurrent      int    `json:"branch_setting_withdraw_or_current" validate:"min=0"`
		BranchSettingWithdrawOREnd          int    `json:"branch_setting_withdraw_or_end" validate:"min=0"`
		BranchSettingWithdrawIteration      int    `json:"branch_setting_withdraw_or_iteration" validate:"min=0"`
		BranchSettingWithdrawORUnique       bool   `json:"branch_setting_withdraw_or_unique"`
		BranchSettingWithdrawUseDateOR      bool   `json:"branch_setting_withdraw_use_date_or"`

		BranchSettingDepositAllowUserInput bool   `json:"branch_setting_deposit_allow_user_input"`
		BranchSettingDepositPrefix         string `json:"branch_setting_deposit_prefix" validate:"omitempty"`
		BranchSettingDepositORStart        int    `json:"branch_setting_deposit_or_start" validate:"min=0"`
		BranchSettingDepositORCurrent      int    `json:"branch_setting_deposit_or_current" validate:"min=0"`
		BranchSettingDepositOREnd          int    `json:"branch_setting_deposit_or_end" validate:"min=0"`
		BranchSettingDepositORIteration    int    `json:"branch_setting_deposit_or_iteration" validate:"min=0"`
		BranchSettingDepositORUnique       bool   `json:"branch_setting_deposit_or_unique"`
		BranchSettingDepositUseDateOR      bool   `json:"branch_setting_deposit_use_date_or"`

		BranchSettingLoanAllowUserInput bool   `json:"branch_setting_loan_allow_user_input"`
		BranchSettingLoanPrefix         string `json:"branch_setting_loan_prefix" validate:"omitempty"`
		BranchSettingLoanORStart        int    `json:"branch_setting_loan_or_start" validate:"min=0"`
		BranchSettingLoanORCurrent      int    `json:"branch_setting_loan_or_current" validate:"min=0"`
		BranchSettingLoanOREnd          int    `json:"branch_setting_loan_or_end" validate:"min=0"`
		BranchSettingLoanORIteration    int    `json:"branch_setting_loan_or_iteration" validate:"min=0"`
		BranchSettingLoanORUnique       bool   `json:"branch_setting_loan_or_unique"`
		BranchSettingLoanUseDateOR      bool   `json:"branch_setting_loan_use_date_or"`

		BranchSettingCheckVoucherAllowUserInput bool   `json:"branch_setting_check_voucher_allow_user_input"`
		BranchSettingCheckVoucherPrefix         string `json:"branch_setting_check_voucher_prefix" validate:"omitempty"`
		BranchSettingCheckVoucherORStart        int    `json:"branch_setting_check_voucher_or_start" validate:"min=0"`
		BranchSettingCheckVoucherORCurrent      int    `json:"branch_setting_check_voucher_or_current" validate:"min=0"`
		BranchSettingCheckVoucherOREnd          int    `json:"branch_setting_check_voucher_or_end" validate:"min=0"`
		BranchSettingCheckVoucherORIteration    int    `json:"branch_setting_check_voucher_or_iteration" validate:"min=0"`
		BranchSettingCheckVoucherORUnique       bool   `json:"branch_setting_check_voucher_or_unique"`
		BranchSettingCheckVoucherUseDateOR      bool   `json:"branch_setting_check_voucher_use_date_or"`

		BranchSettingDefaultMemberTypeID *uuid.UUID `json:"branch_setting_default_member_type_id,omitempty"`
	}
)

func (m *Model) Branch() {
	m.Migration = append(m.Migration, &Branch{})
	m.BranchManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Branch, BranchResponse, BranchRequest]{
		Preloads: []string{
			"Media",
			"CreatedBy",
			"UpdatedBy",
			"Footsteps",
			"GeneratedReports",
			"InvitationCodes",
			"PermissionTemplates",
			"UserOrganizations",
			"Organization",
			"Organization.Media",
			"Organization.CreatedBy",
			"Organization.Media",
			"Organization.CoverMedia",
			"BranchSettingDefaultMemberType",
		},
		Service: m.provider.Service,
		Resource: func(data *Branch) *BranchResponse {
			if data == nil {
				return nil
			}
			return &BranchResponse{
				ID:           data.ID,
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:  data.CreatedByID,
				CreatedBy:    m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:  data.UpdatedByID,
				UpdatedBy:    m.UserManager.ToModel(data.UpdatedBy),
				Organization: m.OrganizationManager.ToModel(data.Organization),

				MediaID:       data.MediaID,
				Media:         m.MediaManager.ToModel(data.Media),
				Type:          data.Type,
				Name:          data.Name,
				Email:         data.Email,
				Description:   data.Description,
				CountryCode:   data.CountryCode,
				ContactNumber: data.ContactNumber,
				Address:       data.Address,
				Province:      data.Province,
				City:          data.City,
				Region:        data.Region,
				Barangay:      data.Barangay,
				PostalCode:    data.PostalCode,
				Latitude:      data.Latitude,
				Longitude:     data.Longitude,

				IsMainBranch: data.IsMainBranch,

				Footsteps:           m.FootstepManager.ToModels(data.Footsteps),
				GeneratedReports:    m.GeneratedReportManager.ToModels(data.GeneratedReports),
				InvitationCodes:     m.InvitationCodeManager.ToModels(data.InvitationCodes),
				PermissionTemplates: m.PermissionTemplateManager.ToModels(data.PermissionTemplates),
				UserOrganizations:   m.UserOrganizationManager.ToModels(data.UserOrganizations),

				BranchSettingWithdrawAllowUserInput: data.BranchSettingWithdrawAllowUserInput,
				BranchSettingWithdrawPrefix:         data.BranchSettingWithdrawPrefix,
				BranchSettingWithdrawORStart:        data.BranchSettingWithdrawORStart,
				BranchSettingWithdrawORCurrent:      data.BranchSettingWithdrawORCurrent,
				BranchSettingWithdrawOREnd:          data.BranchSettingWithdrawOREnd,
				BranchSettingWithdrawIteration:      data.BranchSettingWithdrawIteration,
				BranchSettingWithdrawORUnique:       data.BranchSettingWithdrawORUnique,
				BranchSettingWithdrawUseDateOR:      data.BranchSettingWithdrawUseDateOR,

				BranchSettingDepositAllowUserInput: data.BranchSettingDepositAllowUserInput,
				BranchSettingDepositPrefix:         data.BranchSettingDepositPrefix,
				BranchSettingDepositORStart:        data.BranchSettingDepositORStart,
				BranchSettingDepositORCurrent:      data.BranchSettingDepositORCurrent,
				BranchSettingDepositOREnd:          data.BranchSettingDepositOREnd,
				BranchSettingDepositORIteration:    data.BranchSettingDepositORIteration,
				BranchSettingDepositORUnique:       data.BranchSettingDepositORUnique,
				BranchSettingDepositUseDateOR:      data.BranchSettingDepositUseDateOR,

				BranchSettingLoanAllowUserInput: data.BranchSettingLoanAllowUserInput,
				BranchSettingLoanPrefix:         data.BranchSettingLoanPrefix,
				BranchSettingLoanORStart:        data.BranchSettingLoanORStart,
				BranchSettingLoanORCurrent:      data.BranchSettingLoanORCurrent,
				BranchSettingLoanOREnd:          data.BranchSettingLoanOREnd,
				BranchSettingLoanORIteration:    data.BranchSettingLoanORIteration,
				BranchSettingLoanORUnique:       data.BranchSettingLoanORUnique,
				BranchSettingLoanUseDateOR:      data.BranchSettingLoanUseDateOR,

				BranchSettingCheckVoucherAllowUserInput: data.BranchSettingCheckVoucherAllowUserInput,
				BranchSettingCheckVoucherPrefix:         data.BranchSettingCheckVoucherPrefix,
				BranchSettingCheckVoucherORStart:        data.BranchSettingCheckVoucherORStart,
				BranchSettingCheckVoucherORCurrent:      data.BranchSettingCheckVoucherORCurrent,
				BranchSettingCheckVoucherOREnd:          data.BranchSettingCheckVoucherOREnd,
				BranchSettingCheckVoucherORIteration:    data.BranchSettingCheckVoucherORIteration,
				BranchSettingCheckVoucherORUnique:       data.BranchSettingCheckVoucherORUnique,
				BranchSettingCheckVoucherUseDateOR:      data.BranchSettingCheckVoucherUseDateOR,

				// BranchSettingDefaultMemberTypeID: data.BranchSettingDefaultMemberTypeID,
				// BranchSettingDefaultMemberType:   m.MemberTypeManager.ToModel(data.BranchSettingDefaultMemberType),
			}
		},
		Created: func(data *Branch) []string {
			return []string{
				"branch.create",
				fmt.Sprintf("branch.create.%s", data.ID),
				fmt.Sprintf("branch.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Branch) []string {
			return []string{
				"branch.update",
				fmt.Sprintf("branch.update.%s", data.ID),
				fmt.Sprintf("branch.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Branch) []string {
			return []string{
				"branch.delete",
				fmt.Sprintf("branch.delete.%s", data.ID),
				fmt.Sprintf("branch.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) GetBranchesByOrganization(context context.Context, organizationId uuid.UUID) ([]*Branch, error) {
	return m.BranchManager.Find(context, &Branch{OrganizationID: organizationId})
}

func (m *Model) GetBranchesByOrganizationCount(context context.Context, organizationId uuid.UUID) (int64, error) {
	return m.BranchManager.Count(context, &Branch{OrganizationID: organizationId})
}
