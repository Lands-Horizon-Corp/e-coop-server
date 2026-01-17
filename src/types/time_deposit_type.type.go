package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	TimeDepositType struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_type"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_type"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Header1  int `gorm:"default:30"`
		Header2  int `gorm:"default:60"`
		Header3  int `gorm:"default:90"`
		Header4  int `gorm:"default:120"`
		Header5  int `gorm:"default:150"`
		Header6  int `gorm:"default:180"`
		Header7  int `gorm:"default:210"`
		Header8  int `gorm:"default:240"`
		Header9  int `gorm:"default:300"`
		Header10 int `gorm:"default:330"`
		Header11 int `gorm:"default:360"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text"`

		PreMature     int     `gorm:"default:0"`
		PreMatureRate float64 `gorm:"type:decimal;default:0"`
		Excess        float64 `gorm:"type:decimal;default:0"`

		TimeDepositComputations          []*TimeDepositComputation          `gorm:"foreignKey:TimeDepositTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"time_deposit_computations,omitempty"`
		TimeDepositComputationPreMatures []*TimeDepositComputationPreMature `gorm:"foreignKey:TimeDepositTypeID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"time_deposit_computation_pre_matures,omitempty"`
	}

	TimeDepositTypeResponse struct {
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
		CurrencyID     uuid.UUID             `json:"currency_id"`
		Currency       *CurrencyResponse     `json:"currency,omitempty"`

		Header1  int `json:"header_1"`
		Header2  int `json:"header_2"`
		Header3  int `json:"header_3"`
		Header4  int `json:"header_4"`
		Header5  int `json:"header_5"`
		Header6  int `json:"header_6"`
		Header7  int `json:"header_7"`
		Header8  int `json:"header_8"`
		Header9  int `json:"header_9"`
		Header10 int `json:"header_10"`
		Header11 int `json:"header_11"`

		Name          string  `json:"name"`
		Description   string  `json:"description"`
		PreMature     int     `json:"pre_mature"`
		PreMatureRate float64 `json:"pre_mature_rate"`
		Excess        float64 `json:"excess"`

		TimeDepositComputations          []*TimeDepositComputationResponse          `json:"time_deposit_computations,omitempty"`
		TimeDepositComputationPreMatures []*TimeDepositComputationPreMatureResponse `json:"time_deposit_computation_pre_matures,omitempty"`
	}

	TimeDepositTypeRequest struct {
		Name          string    `json:"name" validate:"required,min=1,max=255"`
		Description   string    `json:"description,omitempty"`
		CurrencyID    uuid.UUID `json:"currency_id" validate:"required"`
		PreMature     int       `json:"pre_mature,omitempty"`
		PreMatureRate float64   `json:"pre_mature_rate,omitempty"`
		Excess        float64   `json:"excess,omitempty"`

		Header1  int `json:"header_1,omitempty"`
		Header2  int `json:"header_2,omitempty"`
		Header3  int `json:"header_3,omitempty"`
		Header4  int `json:"header_4,omitempty"`
		Header5  int `json:"header_5,omitempty"`
		Header6  int `json:"header_6,omitempty"`
		Header7  int `json:"header_7,omitempty"`
		Header8  int `json:"header_8,omitempty"`
		Header9  int `json:"header_9,omitempty"`
		Header10 int `json:"header_10,omitempty"`
		Header11 int `json:"header_11,omitempty"`

		TimeDepositComputations          []*TimeDepositComputationRequest          `json:"time_deposit_computations,omitempty"`
		TimeDepositComputationPreMatures []*TimeDepositComputationPreMatureRequest `json:"time_deposit_computation_pre_matures,omitempty"`

		TimeDepositComputationsDeleted          uuid.UUIDs `json:"time_deposit_computations_deleted,omitempty"`
		TimeDepositComputationPreMaturesDeleted uuid.UUIDs `json:"time_deposit_computation_pre_matures_deleted,omitempty"`
	}
)
