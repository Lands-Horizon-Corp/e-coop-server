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
	LoanComakerMember struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_comaker_member"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_comaker_member"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID   uuid.UUID        `gorm:"type:uuid;not null"`
		MemberProfile     *MemberProfile   `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Description string  `gorm:"type:text"`
		Amount      float64 `gorm:"type:decimal"`
		MonthsCount int     `gorm:"type:int"`
		YearCount   float64 `gorm:"type:decimal"`
	}

	LoanComakerMemberResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		MemberProfileID   uuid.UUID                `json:"member_profile_id"`
		MemberProfile     *MemberProfileResponse   `json:"member_profile,omitempty"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		Description       string                   `json:"description"`
		Amount            float64                  `json:"amount"`
		MonthsCount       int                      `json:"months_count"`
		YearCount         float64                  `json:"year_count"`
	}

	LoanComakerMemberRequest struct {
		MemberProfileID   uuid.UUID `json:"member_profile_id" validate:"required"`
		LoanTransactionID uuid.UUID `json:"loan_transaction_id" validate:"required"`
		Description       string    `json:"description,omitempty"`
		Amount            float64   `json:"amount,omitempty"`
		MonthsCount       int       `json:"months_count,omitempty"`
		YearCount         float64   `json:"year_count,omitempty"`
	}
)

func (m *ModelCore) loanComakerMember() {
	m.Migration = append(m.Migration, &LoanComakerMember{})
	m.LoanComakerMemberManager = services.NewRepository(services.RepositoryParams[
		LoanComakerMember, LoanComakerMemberResponse, LoanComakerMemberRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"MemberProfile", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanComakerMember) *LoanComakerMemberResponse {
			if data == nil {
				return nil
			}
			return &LoanComakerMemberResponse{
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
				MemberProfileID:   data.MemberProfileID,
				MemberProfile:     m.MemberProfileManager.ToModel(data.MemberProfile),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager.ToModel(data.LoanTransaction),
				Description:       data.Description,
				Amount:            data.Amount,
				MonthsCount:       data.MonthsCount,
				YearCount:         data.YearCount,
			}
		},

		Created: func(data *LoanComakerMember) []string {
			return []string{
				"loan_comaker_member.create",
				fmt.Sprintf("loan_comaker_member.create.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanComakerMember) []string {
			return []string{
				"loan_comaker_member.update",
				fmt.Sprintf("loan_comaker_member.update.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanComakerMember) []string {
			return []string{
				"loan_comaker_member.delete",
				fmt.Sprintf("loan_comaker_member.delete.%s", data.ID),
				fmt.Sprintf("loan_comaker_member.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_comaker_member.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) LoanComakerMemberCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanComakerMember, error) {
	return m.LoanComakerMemberManager.Find(context, &LoanComakerMember{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
