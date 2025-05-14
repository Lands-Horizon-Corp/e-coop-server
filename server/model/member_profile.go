package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	MemberProfile struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_profile"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_profile"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID           *uuid.UUID `gorm:"type:uuid"`
		User             *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID          *uuid.UUID `gorm:"type:uuid"`
		Media            *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		SignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature,omitempty"`

		MemberCenterID         *uuid.UUID            `gorm:"type:uuid"`
		MemberCenter           *MemberCenter         `gorm:"foreignKey:MemberCenterID;constraint:OnDelete:SET NULL;" json:"member_center,omitempty"`
		MemberClassificationID *uuid.UUID            `gorm:"type:uuid"`
		MemberClassification   *MemberClassification `gorm:"foreignKey:MemberClassificationID;constraint:OnDelete:SET NULL;" json:"member_classification,omitempty"`
		MemberGenderID         *uuid.UUID            `gorm:"type:uuid"`
		MemberGender           *MemberGender         `gorm:"foreignKey:MemberGenderID;constraint:OnDelete:SET NULL;" json:"member_gender,omitempty"`
		MemberGroupID          *uuid.UUID            `gorm:"type:uuid"`
		MemberGroup            *MemberGroup          `gorm:"foreignKey:MemberGroupID;constraint:OnDelete:SET NULL;" json:"member_group,omitempty"`
		MemberOccupationID     *uuid.UUID            `gorm:"type:uuid"`
		MemberOccupation       *MemberOccupation     `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:SET NULL;" json:"member_occupation,omitempty"`
		MemberTypeID           *uuid.UUID            `gorm:"type:uuid"`
		MemberType             *MemberType           `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		IsClosed             bool   `gorm:"not null;default:false"`
		IsMutualFundMember   bool   `gorm:"not null;default:false"`
		IsMicroFinanceMember bool   `gorm:"not null;default:false"`
		FirstName            string `gorm:"type:varchar(255);not null"`
		MiddleName           string `gorm:"type:varchar(255)"`
		LastName             string `gorm:"type:varchar(255);not null"`
		FullName             string `gorm:"type:varchar(255);not null"`
		Suffix               string `gorm:"type:varchar(50)"`
		Birthdate            *time.Time
		Status               string `gorm:"type:varchar(50);not null;default:'pending'"`

		Description           string `gorm:"type:text"`
		Notes                 string `gorm:"type:text"`
		ContactNumber         string `gorm:"type:varchar(255)"`
		OldReferenceID        string `gorm:"type:varchar(50)"`
		Passbook              string `gorm:"type:varchar(255)"`
		Occupation            string `gorm:"type:varchar(255)"`
		BusinessAddress       string `gorm:"type:varchar(255)"`
		BusinessContactNumber string `gorm:"type:varchar(255)"`
		CivilStatus           string `gorm:"type:varchar(255);not null;default:'single'"`
	}
	MemberProfileResponse struct {
		ID                     uuid.UUID                    `json:"id"`
		CreatedAt              string                       `json:"created_at"`
		CreatedByID            uuid.UUID                    `json:"created_by_id"`
		CreatedBy              *UserResponse                `json:"created_by,omitempty"`
		UpdatedAt              string                       `json:"updated_at"`
		UpdatedByID            uuid.UUID                    `json:"updated_by_id"`
		UpdatedBy              *UserResponse                `json:"updated_by,omitempty"`
		OrganizationID         uuid.UUID                    `json:"organization_id"`
		Organization           *OrganizationResponse        `json:"organization,omitempty"`
		BranchID               uuid.UUID                    `json:"branch_id"`
		Branch                 *BranchResponse              `json:"branch,omitempty"`
		UserID                 uuid.UUID                    `json:"user_id"`
		User                   UserResponse                 `json:"user,omitempty"`
		MediaID                uuid.UUID                    `json:"media_id"`
		Media                  MediaResponse                `json:"media,omitempty"`
		SignatureMediaID       uuid.UUID                    `json:"signature_media_id"`
		SignatureMedia         MediaResponse                `json:"signature_media,omitempty"`
		MemberCenterID         uuid.UUID                    `json:"member_center_id"`
		MemberCenter           MemberCenterResponse         `json:"member_center,omitempty"`
		MemberClassificationID uuid.UUID                    `json:"member_classification_id"`
		MemberClassification   MemberClassificationResponse `json:"member_classification,omitempty"`
		MemberGenderID         uuid.UUID                    `json:"member_gender_id"`
		MemberGender           MemberGenderResponse         `json:"member_gender,omitempty"`
		MemberGroupID          uuid.UUID                    `json:"member_group_id"`
		MemberGroup            MemberGroupResponse          `json:"member_group,omitempty"`
		MemberOccupationID     uuid.UUID                    `json:"member_occupation_id"`
		MemberOccupation       MemberOccupationResponse     `json:"member_occupation,omitempty"`
		MemberTypeID           uuid.UUID                    `json:"member_type_id,omitempty"`
		MemberType             MemberTypeResponse           `json:"member_tyoe,omitempty"`
		IsClosed               bool                         `json:"is_closed"`
		IsMutualFundMember     bool                         `json:"is_mutual_fund_member"`
		IsMicroFinanceMember   bool                         `json:"is_micro_finance_member"`
		FirstName              string                       `json:"first_name"`
		MiddleName             string                       `json:"middle_name"`
		LastName               string                       `json:"last_name"`
		FullName               string                       `json:"full_name"`
		Suffix                 string                       `json:"suffix"`
		Birthdate              string                       `json:"birthdate"`
		Status                 string                       `json:"status"`
		Description            string                       `json:"description"`
		Notes                  string                       `json:"notes"`
		ContactNumber          string                       `json:"contact_number"`
		OldReferenceID         string                       `json:"old_reference_id"`
		Passbook               string                       `json:"passbook"`
		Occupation             string                       `json:"occupation"`
		BusinessAddress        string                       `json:"business_address"`
		BusinessContactNumber  string                       `json:"business_contact_number"`
		CivilStatus            string                       `json:"civil_status"`
	}

	MemberProfileRequest struct {
		MediaID                *uuid.UUID `json:"media_id,omitempty" validate:"omitempty,uuid4"`
		SignatureMediaID       *uuid.UUID `json:"signature_media_id,omitempty" validate:"omitempty,uuid4"`
		MemberCenterID         *uuid.UUID `json:"member_center_id,omitempty" validate:"omitempty,uuid4"`
		MemberClassificationID *uuid.UUID `json:"member_classification_id,omitempty" validate:"omitempty,uuid4"`
		MemberGenderID         *uuid.UUID `json:"member_gender_id,omitempty" validate:"omitempty,uuid4"`
		MemberGroupID          *uuid.UUID `json:"member_group_id,omitempty" validate:"omitempty,uuid4"`
		MemberOccupationID     *uuid.UUID `json:"member_occupation_id,omitempty" validate:"omitempty,uuid4"`
		MemberTypeID           *uuid.UUID `json:"member_type_id,omitempty" validate:"omitempty,uuid4"`

		IsClosed             bool `json:"is_closed"`
		IsMutualFundMember   bool `json:"is_mutual_fund_member"`
		IsMicroFinanceMember bool `json:"is_micro_finance_member"`

		FirstName  string     `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName string     `json:"middle_name,omitempty" validate:"omitempty,max=255"`
		LastName   string     `json:"last_name"  validate:"required,min=1,max=255"`
		FullName   string     `json:"full_name"  validate:"required,min=1,max=255"`
		Suffix     string     `json:"suffix,omitempty" validate:"omitempty,max=50"`
		Birthdate  *time.Time `json:"birthdate,omitempty" validate:"omitempty"`
		Status     string     `json:"status" validate:"required,oneof=pending active inactive"`

		Description           string `json:"description,omitempty" validate:"omitempty"`
		Notes                 string `json:"notes,omitempty" validate:"omitempty"`
		ContactNumber         string `json:"contact_number,omitempty" validate:"omitempty,max=255"`
		OldReferenceID        string `json:"old_reference_id,omitempty" validate:"omitempty,max=50"`
		Passbook              string `json:"passbook,omitempty" validate:"omitempty,max=255"`
		Occupation            string `json:"occupation,omitempty" validate:"omitempty,max=255"`
		BusinessAddress       string `json:"business_address,omitempty" validate:"omitempty,max=255"`
		BusinessContactNumber string `json:"business_contact_number,omitempty" validate:"omitempty,max=255"`
		CivilStatus           string `json:"civil_status" validate:"required,oneof=single married widowed separated"`
	}

	MemberProfileCollection struct {
		Manager horizon_manager.CollectionManager[MemberProfile]
	}
)

func (m *Model) MemberProfileModel(data *MemberProfile) *MemberProfileResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberProfile) *MemberProfileResponse {
		return &MemberProfileResponse{
			ID:                     data.ID,
			CreatedAt:              data.CreatedAt.Format(time.RFC3339),
			CreatedByID:            data.CreatedByID,
			CreatedBy:              m.UserModel(data.CreatedBy),
			UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:            data.UpdatedByID,
			UpdatedBy:              m.UserModel(data.UpdatedBy),
			OrganizationID:         data.OrganizationID,
			Organization:           m.OrganizationModel(data.Organization),
			BranchID:               data.BranchID,
			Branch:                 m.BranchModel(data.Branch),
			UserID:                 *data.UserID,
			User:                   *m.UserModel(data.User),
			MediaID:                *data.MediaID,
			Media:                  *m.MediaModel(data.Media),
			SignatureMediaID:       *data.SignatureMediaID,
			SignatureMedia:         *m.MediaModel(data.SignatureMedia),
			MemberCenterID:         *data.MemberCenterID,
			MemberCenter:           *m.MemberCenterModel(data.MemberCenter),
			MemberClassificationID: *data.MemberClassificationID,
			MemberClassification:   *m.MemberClassificationModel(data.MemberClassification),
			MemberGenderID:         *data.MemberGenderID,
			MemberGender:           *m.MemberGenderModel(data.MemberGender),
			MemberGroupID:          *data.MemberGroupID,
			MemberGroup:            *m.MemberGroupModel(data.MemberGroup),
			MemberOccupationID:     *data.MemberOccupationID,
			MemberOccupation:       *m.MemberOccupationModel(data.MemberOccupation),
			MemberTypeID:           *data.MemberTypeID,
			MemberType:             *m.MemberTypeModel(data.MemberType),
			IsClosed:               data.IsClosed,
			IsMutualFundMember:     data.IsMutualFundMember,
			IsMicroFinanceMember:   data.IsMicroFinanceMember,
			FirstName:              data.FirstName,
			MiddleName:             data.MiddleName,
			LastName:               data.LastName,
			FullName:               data.FullName,
			Suffix:                 data.Suffix,
			Birthdate:              data.Birthdate.Format(time.RFC3339),
			Status:                 data.Status,
			Description:            data.Description,
			Notes:                  data.Notes,
			ContactNumber:          data.ContactNumber,
			OldReferenceID:         data.OldReferenceID,
			Passbook:               data.Passbook,
			Occupation:             data.Occupation,
			BusinessAddress:        data.BusinessAddress,
			BusinessContactNumber:  data.BusinessContactNumber,
			CivilStatus:            data.CivilStatus,
		}
	})
}

func (m *Model) MemberProfileValidate(ctx echo.Context) (*MemberProfileRequest, error) {
	return horizon_manager.Validate[MemberProfileRequest](ctx, m.validator)
}

func (m *Model) MemberProfileModels(data []*MemberProfile) []*MemberProfileResponse {
	return horizon_manager.ToModels(data, m.MemberProfileModel)
}

func NewMemberProfileCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberProfileCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberProfile) ([]string, any) {
			return []string{
				fmt.Sprintf("member_profile.create.%s", data.ID),
				fmt.Sprintf("member_profile.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_center.create.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		func(data *MemberProfile) ([]string, any) {
			return []string{
				"member_profile.update",
				fmt.Sprintf("member_profile.update.%s", data.ID),
				fmt.Sprintf("member_profile.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_profile.update.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		func(data *MemberProfile) ([]string, any) {
			return []string{
				"member_profile.delete",
				fmt.Sprintf("member_profile.delete.%s", data.ID),
				fmt.Sprintf("member_profile.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_profile.delete.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"User",
			"Media",
			"SignatureMedia",
			"MemberCenter",
			"MemberClassification",
			"MemberGender",
			"MemberGroup",
			"MemberOccupation",
			"MemberType",
		},
	)
	return &MemberProfileCollection{
		Manager: manager,
	}, nil
}

// member-group/branch/:branch_id
func (fc *MemberProfileCollection) ListByBranch(branchID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		BranchID: branchID,
	})
}

// member-group/organization/:organization_id
func (fc *MemberProfileCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		OrganizationID: organizationID,
	})
}

// member-group/organization/:organization_id/branch/:branch_id
func (fc *MemberProfileCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
