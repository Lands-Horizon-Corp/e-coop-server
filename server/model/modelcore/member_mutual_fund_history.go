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
	MemberMutualFundHistory struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_mutual_fund_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_mutual_fund_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Title       string  `gorm:"type:varchar(255)"`
		Amount      float64 `gorm:"type:decimal(20,6)"`
		Description string  `gorm:"type:text"`
	}

	// MemberMutualFundHistoryResponse represents the response structure for membermutualfundhistory data

	MemberMutualFundHistoryResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		UpdatedAt       string                 `json:"updated_at"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		Title           string                 `json:"title"`
		Amount          float64                `json:"amount"`
		Description     string                 `json:"description"`
	}

	// MemberMutualFundHistoryRequest represents the request structure for creating/updating membermutualfundhistory

	MemberMutualFundHistoryRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		Title           string    `json:"title" validate:"required,min=1,max=255"`
		Amount          float64   `json:"amount" validate:"required"`
		Description     string    `json:"description,omitempty"`
	}
)

func (m *ModelCore) memberMutualFundHistory() {
	m.Migration = append(m.Migration, &MemberMutualFundHistory{})
	m.MemberMutualFundHistoryManager = services.NewRepository(services.RepositoryParams[MemberMutualFundHistory, MemberMutualFundHistoryResponse, MemberMutualFundHistoryRequest]{
		Preloads: []string{"Organization", "Branch", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberMutualFundHistory) *MemberMutualFundHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberMutualFundHistoryResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				Title:           data.Title,
				Amount:          data.Amount,
				Description:     data.Description,
			}
		},

		Created: func(data *MemberMutualFundHistory) []string {
			return []string{
				"member_mutual_fund_history.create",
				fmt.Sprintf("member_mutual_fund_history.create.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberMutualFundHistory) []string {
			return []string{
				"member_mutual_fund_history.update",
				fmt.Sprintf("member_mutual_fund_history.update.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberMutualFundHistory) []string {
			return []string{
				"member_mutual_fund_history.delete",
				fmt.Sprintf("member_mutual_fund_history.delete.%s", data.ID),
				fmt.Sprintf("member_mutual_fund_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_mutual_fund_history.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberMutualFundHistoryCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*MemberMutualFundHistory, error) {
	return m.MemberMutualFundHistoryManager.Find(context, &MemberMutualFundHistory{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
