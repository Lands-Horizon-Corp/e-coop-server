package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ComakerMemberProfile struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_comaker_member_profile" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_comaker_member_profile" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_loan_transaction_comaker" json:"loan_transaction_id"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null" json:"member_profile_id"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Amount      float64 `gorm:"type:decimal;not null" json:"amount"`
		Description string  `gorm:"type:text" json:"description"`
		MonthsCount int     `gorm:"type:int;default:0" json:"months_count"`
		YearCount   int     `gorm:"type:int;default:0" json:"year_count"`
	}

	ComakerMemberProfileResponse struct {
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

		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`

		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		MonthsCount int     `json:"months_count"`
		YearCount   int     `json:"year_count"`
	}

	ComakerMemberProfileRequest struct {
		ID                *uuid.UUID `json:"id,omitempty"`
		LoanTransactionID uuid.UUID  `json:"loan_transaction_id" validate:"required"`
		MemberProfileID   uuid.UUID  `json:"member_profile_id" validate:"required"`
		Amount            float64    `json:"amount" validate:"required,min=0"`
		Description       string     `json:"description,omitempty"`
		MonthsCount       int        `json:"months_count,omitempty"`
		YearCount         int        `json:"year_count,omitempty"`
	}
)

func (m *ModelCore) comakerMemberProfile() {
	m.Migration = append(m.Migration, &ComakerMemberProfile{})
	m.ComakerMemberProfileManager = services.NewRepository(services.RepositoryParams[ComakerMemberProfile, ComakerMemberProfileResponse, ComakerMemberProfileRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "LoanTransaction", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *ComakerMemberProfile) *ComakerMemberProfileResponse {
			if data == nil {
				return nil
			}
			return &ComakerMemberProfileResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager.ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager.ToModel(data.Branch),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager.ToModel(data.LoanTransaction),
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     m.MemberProfileManager.ToModel(data.MemberProfile),
				Amount:            data.Amount,
				Description:       data.Description,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},
		Created: func(data *ComakerMemberProfile) []string {
			return []string{
				"comaker_member_profile.create",
				fmt.Sprintf("comaker_member_profile.create.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.create.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.create.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Updated: func(data *ComakerMemberProfile) []string {
			return []string{
				"comaker_member_profile.update",
				fmt.Sprintf("comaker_member_profile.update.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.update.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.update.loan_transaction.%s", data.LoanTransactionID),
			}
		},
		Deleted: func(data *ComakerMemberProfile) []string {
			return []string{
				"comaker_member_profile.delete",
				fmt.Sprintf("comaker_member_profile.delete.%s", data.ID),
				fmt.Sprintf("comaker_member_profile.delete.branch.%s", data.BranchID),
				fmt.Sprintf("comaker_member_profile.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("comaker_member_profile.delete.loan_transaction.%s", data.LoanTransactionID),
			}
		},
	})
}

func (m *ModelCore) comakerMemberProfileCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*ComakerMemberProfile, error) {
	return m.ComakerMemberProfileManager.Find(context, &ComakerMemberProfile{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *ModelCore) comakerMemberProfileByLoanTransaction(context context.Context, loanTransactionId uuid.UUID) ([]*ComakerMemberProfile, error) {
	return m.ComakerMemberProfileManager.Find(context, &ComakerMemberProfile{
		LoanTransactionID: loanTransactionId,
	})
}
