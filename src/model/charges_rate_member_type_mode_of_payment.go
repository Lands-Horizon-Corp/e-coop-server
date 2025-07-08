package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Enum for charges_rate_member_type_enum
type ChargesRateMemberTypeEnum string

const (
	ChargesMemberTypeAll         ChargesRateMemberTypeEnum = "all"
	ChargesMemberTypeDaily       ChargesRateMemberTypeEnum = "daily"
	ChargesMemberTypeWeekly      ChargesRateMemberTypeEnum = "weekly"
	ChargesMemberTypeMonthly     ChargesRateMemberTypeEnum = "monthly"
	ChargesMemberTypeSemiMonthly ChargesRateMemberTypeEnum = "semi-monthly"
	ChargesMemberTypeQuarterly   ChargesRateMemberTypeEnum = "quarterly"
	ChargesMemberTypeSemiAnnual  ChargesRateMemberTypeEnum = "semi-annual"
	ChargesMemberTypeLumpsum     ChargesRateMemberTypeEnum = "lumpsum"
)

type (
	ChargesRateMemberTypeModeOfPayment struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_member_type_mode_of_payment"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_member_type_mode_of_payment"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberTypeID uuid.UUID   `gorm:"type:uuid;not null"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`

		ModeOfPayment ChargesRateMemberTypeEnum `gorm:"type:varchar(20);default:'all'"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	ChargesRateMemberTypeModeOfPaymentResponse struct {
		ID             uuid.UUID                 `json:"id"`
		CreatedAt      string                    `json:"created_at"`
		CreatedByID    uuid.UUID                 `json:"created_by_id"`
		CreatedBy      *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt      string                    `json:"updated_at"`
		UpdatedByID    uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy      *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID                 `json:"organization_id"`
		Organization   *OrganizationResponse     `json:"organization,omitempty"`
		BranchID       uuid.UUID                 `json:"branch_id"`
		Branch         *BranchResponse           `json:"branch,omitempty"`
		MemberTypeID   uuid.UUID                 `json:"member_type_id"`
		MemberType     *MemberTypeResponse       `json:"member_type,omitempty"`
		ModeOfPayment  ChargesRateMemberTypeEnum `json:"mode_of_payment"`
		Name           string                    `json:"name"`
		Description    string                    `json:"description"`
	}

	ChargesRateMemberTypeModeOfPaymentRequest struct {
		MemberTypeID  uuid.UUID                 `json:"member_type_id" validate:"required"`
		ModeOfPayment ChargesRateMemberTypeEnum `json:"mode_of_payment,omitempty"`
		Name          string                    `json:"name,omitempty"`
		Description   string                    `json:"description,omitempty"`
	}
)

func (m *Model) ChargesRateMemberTypeModeOfPayment() {
	m.Migration = append(m.Migration, &ChargesRateMemberTypeModeOfPayment{})
	m.ChargesRateMemberTypeModeOfPaymentManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ChargesRateMemberTypeModeOfPayment, ChargesRateMemberTypeModeOfPaymentResponse, ChargesRateMemberTypeModeOfPaymentRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberType",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateMemberTypeModeOfPayment) *ChargesRateMemberTypeModeOfPaymentResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateMemberTypeModeOfPaymentResponse{
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
				MemberTypeID:   data.MemberTypeID,
				MemberType:     m.MemberTypeManager.ToModel(data.MemberType),
				ModeOfPayment:  data.ModeOfPayment,
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *ChargesRateMemberTypeModeOfPayment) []string {
			return []string{
				"charges_rate_member_type_mode_of_payment.create",
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.create.%s", data.ID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateMemberTypeModeOfPayment) []string {
			return []string{
				"charges_rate_member_type_mode_of_payment.update",
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.update.%s", data.ID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateMemberTypeModeOfPayment) []string {
			return []string{
				"charges_rate_member_type_mode_of_payment.delete",
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_member_type_mode_of_payment.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) ChargesRateMemberTypeModeOfPaymentCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*ChargesRateMemberTypeModeOfPayment, error) {
	return m.ChargesRateMemberTypeModeOfPaymentManager.Find(context, &ChargesRateMemberTypeModeOfPayment{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
