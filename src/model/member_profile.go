package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	MemberProfile struct {
		ID                             uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt                      time.Time             `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID                    uuid.UUID             `gorm:"type:uuid" json:"created_by,omitempty"`
		CreatedBy                      *User                 `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by_user,omitempty"`
		UpdatedAt                      time.Time             `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID                    uuid.UUID             `gorm:"type:uuid" json:"updated_by,omitempty"`
		UpdatedBy                      *User                 `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by_user,omitempty"`
		DeletedAt                      gorm.DeletedAt        `gorm:"index" json:"deleted_at"`
		DeletedByID                    *uuid.UUID            `gorm:"type:uuid" json:"deleted_by,omitempty"`
		DeletedBy                      *User                 `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by_user,omitempty"`
		OrganizationID                 uuid.UUID             `gorm:"type:uuid;not null;index:idx_organization_branch_member_profile" json:"organization_id"`
		Organization                   *Organization         `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID                       uuid.UUID             `gorm:"type:uuid;not null;index:idx_organization_branch_member_profile" json:"branch_id"`
		Branch                         *Branch               `gorm:"foreignKey:BranchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"branch,omitempty"`
		MediaID                        *uuid.UUID            `gorm:"type:uuid" json:"media_id,omitempty"`
		Media                          *Media                `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"media,omitempty"`
		SignatureMediaID               *uuid.UUID            `gorm:"type:uuid" json:"signature_media_id,omitempty"`
		SignatureMedia                 *Media                `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"signature_media,omitempty"`
		UserID                         *uuid.UUID            `gorm:"type:uuid" json:"user_id,omitempty"`
		User                           *User                 `gorm:"foreignKey:UserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"user,omitempty"`
		MemberTypeID                   *uuid.UUID            `gorm:"type:uuid" json:"member_type_id,omitempty"`
		MemberType                     *MemberType           `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`
		MemberGroupID                  *uuid.UUID            `gorm:"type:uuid" json:"member_group_id,omitempty"`
		MemberGroup                    *MemberGroup          `gorm:"foreignKey:MemberGroupID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_group,omitempty"`
		MemberGenderID                 *uuid.UUID            `gorm:"type:uuid" json:"member_gender_id,omitempty"`
		MemberGender                   *MemberGender         `gorm:"foreignKey:MemberGenderID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_gender,omitempty"`
		MemberDepartmentID             *uuid.UUID            `gorm:"type:uuid" json:"member_department_id,omitempty"`
		MemberDepartment               *MemberDepartment     `gorm:"foreignKey:MemberDepartmentID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_department,omitempty"`
		MemberCenterID                 *uuid.UUID            `gorm:"type:uuid" json:"member_center_id,omitempty"`
		MemberCenter                   *MemberCenter         `gorm:"foreignKey:MemberCenterID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_center,omitempty"`
		MemberOccupationID             *uuid.UUID            `gorm:"type:uuid" json:"member_occupation_id,omitempty"`
		MemberOccupation               *MemberOccupation     `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_occupation,omitempty"`
		MemberClassificationID         *uuid.UUID            `gorm:"type:uuid" json:"member_classification_id,omitempty"`
		MemberClassification           *MemberClassification `gorm:"foreignKey:MemberClassificationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_classification,omitempty"`
		MemberVerifiedByEmployeeUserID *uuid.UUID            `gorm:"type:uuid" json:"member_verified_by_employee_user_id,omitempty"`
		MemberVerifiedByEmployeeUser   *User                 `gorm:"foreignKey:MemberVerifiedByEmployeeUserID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"member_verified_by_employee_user,omitempty"`
		RecruitedByMemberProfileID     *uuid.UUID            `gorm:"type:uuid" json:"recruited_by_member_profile_id,omitempty"`
		RecruitedByMemberProfile       *MemberProfile        `gorm:"foreignKey:RecruitedByMemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"recruited_by_member_profile,omitempty"`
		IsClosed                       bool                  `gorm:"not null;default:false" json:"is_closed"`
		IsMutualFundMember             bool                  `gorm:"not null;default:false" json:"is_mutual_fund_member"`
		IsMicroFinanceMember           bool                  `gorm:"not null;default:false" json:"is_micro_finance_member"`
		FirstName                      string                `gorm:"type:varchar(255);not null" json:"first_name"`
		MiddleName                     string                `gorm:"type:varchar(255)" json:"middle_name,omitempty"`
		LastName                       string                `gorm:"type:varchar(255);not null" json:"last_name"`
		FullName                       string                `gorm:"type:varchar(255);not null;index:idx_full_name" json:"full_name"`
		Suffix                         string                `gorm:"type:varchar(50)" json:"suffix,omitempty"`
		BirthDate                      time.Time             `gorm:"type:date;not null" json:"birthdate"`
		Status                         string                `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
		Description                    string                `gorm:"type:text" json:"description,omitempty"`
		Notes                          string                `gorm:"type:text" json:"notes,omitempty"`
		ContactNumber                  string                `gorm:"type:varchar(255)" json:"contact_number,omitempty"`
		OldReferenceID                 string                `gorm:"type:varchar(50)" json:"old_reference_id,omitempty"`
		Passbook                       string                `gorm:"type:varchar(255)" json:"passbook,omitempty"`
		Occupation                     string                `gorm:"type:varchar(255)" json:"occupation,omitempty"`
		BusinessAddress                string                `gorm:"type:varchar(255)" json:"business_address,omitempty"`
		BusinessContactNumber          string                `gorm:"type:varchar(255)" json:"business_contact_number,omitempty"`
		CivilStatus                    string                `gorm:"type:varchar(50);not null;default:'single'" json:"civil_status"`

		RecruitedMembers             []*MemberProfile               `gorm:"foreignKey:RecruitedByMemberProfileID" json:"recruited_members,omitempty"`
		MemberAddresses              []*MemberAddress               `gorm:"foreignKey:MemberProfileID" json:"member_addresses,omitempty"`
		MemberAssets                 []*MemberAsset                 `gorm:"foreignKey:MemberProfileID" json:"member_assets,omitempty"`
		MemberIncomes                []*MemberIncome                `gorm:"foreignKey:MemberProfileID" json:"member_incomes,omitempty"`
		MemberExpenses               []*MemberExpense               `gorm:"foreignKey:MemberProfileID" json:"member_expenses,omitempty"`
		MemberGovernmentBenefits     []*MemberGovernmentBenefit     `gorm:"foreignKey:MemberProfileID" json:"member_government_benefits,omitempty"`
		MemberJointAccounts          []*MemberJointAccount          `gorm:"foreignKey:MemberProfileID" json:"member_joint_accounts,omitempty"`
		MemberRelativeAccounts       []*MemberRelativeAccount       `gorm:"foreignKey:MemberProfileID" json:"member_relative_accounts,omitempty"`
		MemberEducationalAttainments []*MemberEducationalAttainment `gorm:"foreignKey:MemberProfileID" json:"member_educational_attainments,omitempty"`
		MemberContactReferences      []*MemberContactReference      `gorm:"foreignKey:MemberProfileID" json:"member_contact_references,omitempty"`
		MemberCloseRemarks           []*MemberCloseRemark           `gorm:"foreignKey:MemberProfileID" json:"member_close_remarks,omitempty"`
	}
	MemberProfileResponse struct {
		ID                             uuid.UUID                     `json:"id"`
		CreatedAt                      string                        `json:"created_at"`
		CreatedByID                    uuid.UUID                     `json:"created_by_id"`
		CreatedBy                      *UserResponse                 `json:"created_by,omitempty"`
		UpdatedAt                      string                        `json:"updated_at"`
		UpdatedByID                    uuid.UUID                     `json:"updated_by_id"`
		UpdatedBy                      *UserResponse                 `json:"updated_by,omitempty"`
		OrganizationID                 uuid.UUID                     `json:"organization_id"`
		Organization                   *OrganizationResponse         `json:"organization,omitempty"`
		BranchID                       uuid.UUID                     `json:"branch_id"`
		Branch                         *BranchResponse               `json:"branch,omitempty"`
		MediaID                        *uuid.UUID                    `json:"media_id,omitempty"`
		Media                          *MediaResponse                `json:"media,omitempty"`
		SignatureMediaID               *uuid.UUID                    `json:"signature_media_id,omitempty"`
		SignatureMedia                 *MediaResponse                `json:"signature_media,omitempty"`
		UserID                         *uuid.UUID                    `json:"user_id,omitempty"`
		User                           *UserResponse                 `json:"user,omitempty"`
		MemberTypeID                   *uuid.UUID                    `json:"member_type_id,omitempty"`
		MemberType                     *MemberTypeResponse           `json:"member_type,omitempty"`
		MemberGroupID                  *uuid.UUID                    `json:"member_group_id,omitempty"`
		MemberGroup                    *MemberGroupResponse          `json:"member_group,omitempty"`
		MemberGenderID                 *uuid.UUID                    `json:"member_gender_id,omitempty"`
		MemberGender                   *MemberGenderResponse         `json:"member_gender,omitempty"`
		MemberDepartmentID             *uuid.UUID                    `json:"member_department_id,omitempty"`
		MemberDepartment               *MemberDepartmentResponse     `json:"member_department,omitempty"`
		MemberCenterID                 *uuid.UUID                    `json:"member_center_id,omitempty"`
		MemberCenter                   *MemberCenterResponse         `json:"member_center,omitempty"`
		MemberOccupationID             *uuid.UUID                    `json:"member_occupation_id,omitempty"`
		MemberOccupation               *MemberOccupationResponse     `json:"member_occupation,omitempty"`
		MemberClassificationID         *uuid.UUID                    `json:"member_classification_id,omitempty"`
		MemberClassification           *MemberClassificationResponse `json:"member_classification,omitempty"`
		MemberVerifiedByEmployeeUserID *uuid.UUID                    `json:"member_verified_by_employee_user_id,omitempty"`
		MemberVerifiedByEmployeeUser   *UserResponse                 `json:"member_verified_by_employee_user,omitempty"`
		RecruitedByMemberProfileID     *uuid.UUID                    `json:"recruited_by_member_profile_id,omitempty"`
		RecruitedByMemberProfile       *MemberProfileResponse        `json:"recruited_by_member_profile,omitempty"`
		IsClosed                       bool                          `json:"is_closed"`
		IsMutualFundMember             bool                          `json:"is_mutual_fund_member"`
		IsMicroFinanceMember           bool                          `json:"is_micro_finance_member"`
		FirstName                      string                        `json:"first_name"`
		MiddleName                     string                        `json:"middle_name"`
		LastName                       string                        `json:"last_name"`
		FullName                       string                        `json:"full_name"`
		Suffix                         string                        `json:"suffix"`
		BirthDate                      string                        `json:"birthdate"`
		Status                         string                        `json:"status"`
		Description                    string                        `json:"description"`
		Notes                          string                        `json:"notes"`
		ContactNumber                  string                        `json:"contact_number"`
		OldReferenceID                 string                        `json:"old_reference_id"`
		Passbook                       string                        `json:"passbook"`
		BusinessAddress                string                        `json:"business_address"`
		BusinessContactNumber          string                        `json:"business_contact_number"`
		CivilStatus                    string                        `json:"civil_status"`

		QRCode *horizon.QRResult `json:"qr_code,omitempty"`

		MemberAddresses              []*MemberAddressReponse                `json:"member_addresses,omitempty"`
		MemberAssets                 []*MemberAssetResponse                 `json:"member_assets,omitempty"`
		MemberIncomes                []*MemberIncomeResponse                `json:"member_incomes,omitempty"`
		MemberExpenses               []*MemberExpenseResponse               `json:"member_expenses,omitempty"`
		MemberGovernmentBenefits     []*MemberGovernmentBenefitResponse     `json:"member_government_benefits,omitempty"`
		MemberJointAccounts          []*MemberJointAccountResponse          `json:"member_joint_accounts,omitempty"`
		MemberRelativeAccounts       []*MemberRelativeAccountResponse       `json:"member_relative_accounts,omitempty"`
		MemberEducationalAttainments []*MemberEducationalAttainmentResponse `json:"member_educational_attainments,omitempty"`
		MemberContactReferences      []*MemberContactReferenceResponse      `json:"member_contact_references,omitempty"`
		MemberCloseRemarks           []*MemberCloseRemarkResponse           `json:"member_close_remarks,omitempty"`
		RecruitedMembers             []*MemberProfileResponse               `json:"recruited_members,omitempty"`
	}

	MemberProfileRequest struct {
		OrganizationID                 uuid.UUID  `json:"organization_id" validate:"required"`
		BranchID                       uuid.UUID  `json:"branch_id" validate:"required"`
		MediaID                        *uuid.UUID `json:"media_id,omitempty"`
		SignatureMediaID               *uuid.UUID `json:"signature_media_id,omitempty"`
		UserID                         uuid.UUID  `json:"user_id" validate:"required"`
		MemberTypeID                   *uuid.UUID `json:"member_type_id,omitempty"`
		MemberGroupID                  *uuid.UUID `json:"member_group_id,omitempty"`
		MemberGenderID                 *uuid.UUID `json:"member_gender_id,omitempty"`
		MemberDepartmentID             *uuid.UUID `json:"member_department_id,omitempty"`
		MemberCenterID                 *uuid.UUID `json:"member_center_id,omitempty"`
		MemberOccupationID             *uuid.UUID `json:"member_occupation_id,omitempty"`
		MemberClassificationID         *uuid.UUID `json:"member_classification_id,omitempty"`
		MemberVerifiedByEmployeeUserID *uuid.UUID `json:"member_verified_by_employee_user_id,omitempty"`
		RecruitedByMemberProfileID     *uuid.UUID `json:"recruited_by_member_profile_id,omitempty"`
		IsClosed                       bool       `json:"is_closed"`
		IsMutualFundMember             bool       `json:"is_mutual_fund_member"`
		IsMicroFinanceMember           bool       `json:"is_micro_finance_member"`
		FirstName                      string     `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName                     string     `json:"middle_name,omitempty"`
		LastName                       string     `json:"last_name" validate:"required,min=1,max=255"`
		FullName                       string     `json:"full_name" validate:"required,min=1,max=255"`
		Suffix                         string     `json:"suffix,omitempty"`
		BirthDate                      time.Time  `json:"birthdate" validate:"required"`
		Status                         string     `json:"status,omitempty"`
		Description                    string     `json:"description,omitempty"`
		Notes                          string     `json:"notes,omitempty"`
		ContactNumber                  string     `json:"contact_number,omitempty"`
		OldReferenceID                 string     `json:"old_reference_id,omitempty"`
		Passbook                       string     `json:"passbook,omitempty"`
		BusinessAddress                string     `json:"business_address,omitempty"`
		BusinessContactNumber          string     `json:"business_contact_number,omitempty"`
		CivilStatus                    string     `json:"civil_status,omitempty"`
	}

	MemberProfilePersonalInfoRequest struct {
		FirstName      string     `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName     string     `json:"middle_name,omitempty" validate:"max=255"`
		LastName       string     `json:"last_name" validate:"required,min=1,max=255"`
		FullName       string     `json:"full_name,omitempty" validate:"max=255"`
		Suffix         string     `json:"suffix,omitempty" validate:"max=50"`
		MemberGenderID *uuid.UUID `json:"member_gender_id,omitempty"`
		BirthDate      time.Time  `json:"birthdate" validate:"required"`
		ContactNumber  string     `json:"contact_number,omitempty" validate:"max=255"`

		MediaID          *uuid.UUID `json:"media_id,omitempty"`
		SignatureMediaID *uuid.UUID `json:"signature_media_id,omitempty"`
		CivilStatus      string     `json:"civil_status" validate:"required,oneof=single married widowed separated divorced"` // Adjust the allowed values as needed

		MemberOccupationID    *uuid.UUID `json:"member_occupation_id,omitempty"`
		BusinessAddress       string     `json:"business_address,omitempty" validate:"max=255"`
		BusinessContactNumber string     `json:"business_contact_number,omitempty" validate:"max=255"`
		Notes                 string     `json:"notes,omitempty"`
		Description           string     `json:"description,omitempty"`
	}

	MemberProfileMembershipInfoRequest struct {
		Passbook                   string     `json:"passbook,omitempty" validate:"max=255"`
		OldReferenceID             string     `json:"old_reference_id,omitempty" validate:"max=50"`
		Status                     string     `json:"status,omitempty" validate:"max=50"`
		MemberTypeID               *uuid.UUID `json:"member_type_id,omitempty"`
		MemberGroupID              *uuid.UUID `json:"member_group_id,omitempty"`
		MemberClassificationID     *uuid.UUID `json:"member_classification_id,omitempty"`
		MemberCenterID             *uuid.UUID `json:"member_center_id,omitempty"`
		RecruitedByMemberProfileID *uuid.UUID `json:"recruited_by_member_profile_id,omitempty"`
		MemberDepartmentID         *uuid.UUID `json:"member_department_id,omitempty"`
		IsMutualFundMember         bool       `json:"is_mutual_fund_member"`
		IsMicroFinanceMember       bool       `json:"is_micro_finance_member"`
	}

	MemberProfileAccountRequest struct {
		UserID *uuid.UUID `json:"user_id,omitempty"`
	}

	MemberProfileMediasRequest struct {
		MediaID          *uuid.UUID `json:"media_id,omitempty"`
		SignatureMediaID *uuid.UUID `json:"signature_media_id,omitempty"`
	}

	AccountInfo struct {
		UserName string `json:"user_name" validate:"required,min=1,max=255"`
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required,min=6,max=128"`
	}

	MemberProfileQuickCreateRequest struct {
		OldReferenceID       string       `json:"old_reference_id,omitempty" validate:"max=50"`
		Passbook             string       `json:"passbook,omitempty" validate:"max=255"`
		OrganizationID       uuid.UUID    `json:"organization_id" validate:"required"`
		BranchID             uuid.UUID    `json:"branch_id" validate:"required"`
		FirstName            string       `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName           string       `json:"middle_name,omitempty" validate:"max=255"`
		LastName             string       `json:"last_name" validate:"required,min=1,max=255"`
		FullName             string       `json:"full_name,omitempty" validate:"max=255"`
		Suffix               string       `json:"suffix,omitempty" validate:"max=50"`
		MemberGenderID       *uuid.UUID   `json:"member_gender_id,omitempty"`
		BirthDate            time.Time    `json:"birthdate" validate:"required"`
		ContactNumber        string       `json:"contact_number,omitempty" validate:"max=255"`
		CivilStatus          string       `json:"civil_status" validate:"required,oneof=single married widowed separated divorced"` // adjust allowed values as needed
		MemberOccupationID   *uuid.UUID   `json:"member_occupation_id,omitempty"`
		Status               string       `json:"status" validate:"required,max=50"`
		IsMutualFundMember   bool         `json:"is_mutual_fund_member"`
		IsMicroFinanceMember bool         `json:"is_micro_finance_member"`
		MemberTypeID         *uuid.UUID   `json:"member_type_id"`
		AccountInfo          *AccountInfo `json:"new_user_info,omitempty" validate:"omitempty"`
	}

	MemberProfileUserAccountRequest struct {
		Password      string    `json:"password,omitempty" validate:"omitempty,min=6,max=100"`
		UserName      string    `json:"user_name" validate:"required,min=1,max=50"`
		FirstName     string    `json:"first_name" validate:"required,min=1,max=50"`
		LastName      string    `json:"last_name" validate:"required,min=1,max=50"`
		MiddleName    string    `json:"middle_name,omitempty" validate:"max=50"`
		FullName      string    `json:"full_name" validate:"required,min=1,max=150"`
		Suffix        string    `json:"suffix,omitempty" validate:"max=20"`
		Email         string    `json:"email" validate:"required,email,max=100"`
		ContactNumber string    `json:"contact_number" validate:"required,max=20"`
		BirthDate     time.Time `json:"birthdate" validate:"required"`
	}
)

func (m *Model) MemberProfile() {
	m.Migration = append(m.Migration, &MemberProfile{})
	m.MemberProfileManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberProfile, MemberProfileResponse, MemberProfileRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Branch", "Organization",
			"Branch.Media", "Organization.Media",
			"Media", "SignatureMedia",
			"User",
			"User.Media",
			"MemberType", "MemberGroup", "MemberGender", "MemberCenter",
			"MemberOccupation", "MemberClassification", "MemberVerifiedByEmployeeUser",
			"MemberVerifiedByEmployeeUser.Media",

			"MemberAddresses",
			"MemberAssets", "MemberAssets.Media",
			"MemberIncomes", "MemberIncomes.Media",
			"MemberExpenses",
			"MemberGovernmentBenefits", "MemberGovernmentBenefits.FrontMedia", "MemberGovernmentBenefits.BackMedia",
			"MemberJointAccounts", "MemberJointAccounts.PictureMedia", "MemberJointAccounts.SignatureMedia",

			"MemberRelativeAccounts", "MemberRelativeAccounts.RelativeMemberProfile", "MemberRelativeAccounts.RelativeMemberProfile.Media",
			"RecruitedByMemberProfile", "RecruitedByMemberProfile.Media",
			"RecruitedMembers", "RecruitedMembers.Media",
			"MemberEducationalAttainments",
			"MemberContactReferences",
			"MemberCloseRemarks",
			"MemberDepartment",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberProfile) *MemberProfileResponse {
			context := context.Background()
			if data == nil {
				return nil
			}

			result, err := m.provider.Service.QR.EncodeQR(context, &QRMemberProfile{
				FirstName:       data.FirstName,
				LastName:        data.LastName,
				MiddleName:      data.MiddleName,
				ContactNumber:   data.ContactNumber,
				MemberProfileID: data.ID.String(),
				BranchID:        data.BranchID.String(),
				OrganizationID:  data.OrganizationID.String(),
			}, "member-qr")
			if err != nil {
				return nil
			}
			return &MemberProfileResponse{
				ID:                             data.ID,
				CreatedAt:                      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                    data.CreatedByID,
				CreatedBy:                      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                    data.UpdatedByID,
				UpdatedBy:                      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:                 data.OrganizationID,
				Organization:                   m.OrganizationManager.ToModel(data.Organization),
				BranchID:                       data.BranchID,
				Branch:                         m.BranchManager.ToModel(data.Branch),
				MediaID:                        data.MediaID,
				Media:                          m.MediaManager.ToModel(data.Media),
				SignatureMediaID:               data.SignatureMediaID,
				SignatureMedia:                 m.MediaManager.ToModel(data.SignatureMedia),
				UserID:                         data.UserID,
				User:                           m.UserManager.ToModel(data.User),
				MemberTypeID:                   data.MemberTypeID,
				MemberType:                     m.MemberTypeManager.ToModel(data.MemberType),
				MemberGroupID:                  data.MemberGroupID,
				MemberGroup:                    m.MemberGroupManager.ToModel(data.MemberGroup),
				MemberGenderID:                 data.MemberGenderID,
				MemberGender:                   m.MemberGenderManager.ToModel(data.MemberGender),
				MemberCenterID:                 data.MemberCenterID,
				MemberCenter:                   m.MemberCenterManager.ToModel(data.MemberCenter),
				MemberOccupationID:             data.MemberOccupationID,
				MemberOccupation:               m.MemberOccupationManager.ToModel(data.MemberOccupation),
				MemberClassificationID:         data.MemberClassificationID,
				MemberClassification:           m.MemberClassificationManager.ToModel(data.MemberClassification),
				MemberVerifiedByEmployeeUserID: data.MemberVerifiedByEmployeeUserID,
				MemberVerifiedByEmployeeUser:   m.UserManager.ToModel(data.MemberVerifiedByEmployeeUser),
				RecruitedByMemberProfileID:     data.RecruitedByMemberProfileID,
				RecruitedByMemberProfile:       m.MemberProfileManager.ToModel(data.RecruitedByMemberProfile),
				IsClosed:                       data.IsClosed,
				IsMutualFundMember:             data.IsMutualFundMember,
				IsMicroFinanceMember:           data.IsMicroFinanceMember,
				FirstName:                      data.FirstName,
				MiddleName:                     data.MiddleName,
				LastName:                       data.LastName,
				FullName:                       data.FullName,
				Suffix:                         data.Suffix,
				BirthDate:                      data.BirthDate.Format(time.RFC3339),
				Status:                         data.Status,
				Description:                    data.Description,
				Notes:                          data.Notes,
				ContactNumber:                  data.ContactNumber,
				OldReferenceID:                 data.OldReferenceID,
				Passbook:                       data.Passbook,
				BusinessAddress:                data.BusinessAddress,
				BusinessContactNumber:          data.BusinessContactNumber,
				CivilStatus:                    data.CivilStatus,
				QRCode:                         result,
				MemberAddresses:                m.MemberAddressManager.ToModels(data.MemberAddresses),
				MemberAssets:                   m.MemberAssetManager.ToModels(data.MemberAssets),
				MemberIncomes:                  m.MemberIncomeManager.ToModels(data.MemberIncomes),
				MemberExpenses:                 m.MemberExpenseManager.ToModels(data.MemberExpenses),
				MemberGovernmentBenefits:       m.MemberGovernmentBenefitManager.ToModels(data.MemberGovernmentBenefits),
				MemberJointAccounts:            m.MemberJointAccountManager.ToModels(data.MemberJointAccounts),
				MemberRelativeAccounts:         m.MemberRelativeAccountManager.ToModels(data.MemberRelativeAccounts),
				MemberEducationalAttainments:   m.MemberEducationalAttainmentManager.ToModels(data.MemberEducationalAttainments),
				MemberContactReferences:        m.MemberContactReferenceManager.ToModels(data.MemberContactReferences),
				MemberCloseRemarks:             m.MemberCloseRemarkManager.ToModels(data.MemberCloseRemarks),
				RecruitedMembers:               m.MemberProfileManager.ToModels(data.RecruitedMembers),
				MemberDepartmentID:             data.MemberDepartmentID,
				MemberDepartment:               m.MemberDepartmentManager.ToModel(data.MemberDepartment),
			}
		},

		Created: func(data *MemberProfile) []string {
			return []string{
				"member_profile.create",
				fmt.Sprintf("member_profile.create.%s", data.ID),
				fmt.Sprintf("member_profile.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberProfile) []string {
			return []string{
				"member_profile.update",
				fmt.Sprintf("member_profile.update.%s", data.ID),
				fmt.Sprintf("member_profile.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberProfile) []string {
			return []string{
				"member_profile.delete",
				fmt.Sprintf("member_profile.delete.%s", data.ID),
				fmt.Sprintf("member_profile.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberProfileCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberProfile, error) {
	return m.MemberProfileManager.Find(context, &MemberProfile{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *Model) MemberProfileDelete(context context.Context, tx *gorm.DB, memberProfileId uuid.UUID) error {
	// Delete MemberEducationalAttainment records
	memberEducationalAttainments, err := m.MemberEducationalAttainmentManager.Find(context, &MemberEducationalAttainment{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberEducationalAttainments {
		if err := m.MemberEducationalAttainmentManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberCloseRemarks, err := m.MemberCloseRemarkManager.Find(context, &MemberCloseRemark{
		MemberProfileID: &memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberCloseRemarks {
		if err := m.MemberCloseRemarkManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberAddress records
	memberAddresses, err := m.MemberAddressManager.Find(context, &MemberAddress{
		MemberProfileID: &memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberAddresses {
		if err := m.MemberAddressManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberContactReference records
	memberContactReferences, err := m.MemberContactReferenceManager.Find(context, &MemberContactReference{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberContactReferences {
		if err := m.MemberContactReferenceManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberAsset records
	memberAssets, err := m.MemberAssetManager.Find(context, &MemberAsset{
		MemberProfileID: &memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberAssets {
		if err := m.MemberAssetManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberIncome records
	memberIncomes, err := m.MemberIncomeManager.Find(context, &MemberIncome{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberIncomes {
		if err := m.MemberIncomeManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberExpenses, err := m.MemberExpenseManager.Find(context, &MemberExpense{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberExpenses {
		if err := m.MemberExpenseManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberGovernmentBenefits, err := m.MemberGovernmentBenefitManager.Find(context, &MemberGovernmentBenefit{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGovernmentBenefits {
		if err := m.MemberGovernmentBenefitManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberJointAccounts, err := m.MemberJointAccountManager.Find(context, &MemberJointAccount{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberJointAccounts {
		if err := m.MemberJointAccountManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberRelativeAccounts, err := m.MemberRelativeAccountManager.Find(context, &MemberRelativeAccount{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberRelativeAccounts {
		if err := m.MemberRelativeAccountManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberGenderHistory records
	memberGenderHistories, err := m.MemberGenderHistoryManager.Find(context, &MemberGenderHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGenderHistories {
		if err := m.MemberGenderHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberCenterHistory records
	memberCenterHistories, err := m.MemberCenterHistoryManager.Find(context, &MemberCenterHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberCenterHistories {
		if err := m.MemberCenterHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberTypeHistory records
	memberTypeHistories, err := m.MemberTypeHistoryManager.Find(context, &MemberTypeHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberTypeHistories {
		if err := m.MemberTypeHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberClassificationHistory records
	memberClassificationHistories, err := m.MemberClassificationHistoryManager.Find(context, &MemberClassificationHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberClassificationHistories {
		if err := m.MemberClassificationHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberOccupationHistory records
	memberOccupationHistories, err := m.MemberOccupationHistoryManager.Find(context, &MemberOccupationHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberOccupationHistories {
		if err := m.MemberOccupationHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	// Delete MemberGroupHistory records
	memberGroupHistories, err := m.MemberGroupHistoryManager.Find(context, &MemberGroupHistory{
		MemberProfileID: memberProfileId,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGroupHistories {
		if err := m.MemberGroupHistoryManager.DeleteByIDWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	return m.MemberProfileManager.DeleteByIDWithTx(context, tx, memberProfileId)
}

func (m *Model) MemberProfileFindUserByID(ctx context.Context, userId uuid.UUID, orgId uuid.UUID, branchId uuid.UUID) (*MemberProfile, error) {
	return m.MemberProfileManager.FindOne(ctx, &MemberProfile{
		UserID:         &userId,
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
func (m *Model) MemberProfileDestroy(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	memberProfile, err := m.MemberProfileManager.GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "failed to get MemberProfile by ID: %s", id)
	}
	memberAddresses, err := m.MemberAddressManager.Find(ctx, &MemberAddress{
		MemberProfileID: &memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member addresses")
	}
	for _, memberAddress := range memberAddresses {
		if err := m.MemberAddressManager.DeleteByIDWithTx(ctx, tx, memberAddress.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member address: %s", memberAddress.ID)
		}
	}

	memberAssets, err := m.MemberAssetManager.Find(ctx, &MemberAsset{
		MemberProfileID: &memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member assets")
	}
	for _, memberAsset := range memberAssets {
		if err := m.MemberAssetManager.DeleteByIDWithTx(ctx, tx, memberAsset.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member asset: %s", memberAsset.ID)
		}
	}

	memberIncomes, err := m.MemberIncomeManager.Find(ctx, &MemberIncome{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member incomes")
	}
	for _, memberIncome := range memberIncomes {
		if err := m.MemberIncomeManager.DeleteByIDWithTx(ctx, tx, memberIncome.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member income: %s", memberIncome.ID)
		}
	}

	memberExpenses, err := m.MemberExpenseManager.Find(ctx, &MemberExpense{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member expenses")
	}
	for _, memberExpense := range memberExpenses {
		if err := m.MemberExpenseManager.DeleteByIDWithTx(ctx, tx, memberExpense.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member expense: %s", memberExpense.ID)
		}
	}

	memberBenefits, err := m.MemberGovernmentBenefitManager.Find(ctx, &MemberGovernmentBenefit{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member government benefits")
	}
	for _, memberBenefit := range memberBenefits {
		if err := m.MemberGovernmentBenefitManager.DeleteByIDWithTx(ctx, tx, memberBenefit.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member government benefit: %s", memberBenefit.ID)
		}
	}
	memberJointAccounts, err := m.MemberJointAccountManager.Find(ctx, &MemberJointAccount{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member joint accounts")
	}
	for _, memberJointAccount := range memberJointAccounts {
		if err := m.MemberJointAccountManager.DeleteByIDWithTx(ctx, tx, memberJointAccount.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member joint account: %s", memberJointAccount.ID)
		}
	}

	memberRelativeAccounts, err := m.MemberRelativeAccountManager.Find(ctx, &MemberRelativeAccount{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member relative accounts")
	}
	for _, memberRelativeAccount := range memberRelativeAccounts {
		if err := m.MemberRelativeAccountManager.DeleteByIDWithTx(ctx, tx, memberRelativeAccount.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member relative account: %s", memberRelativeAccount.ID)
		}
	}
	memberEducations, err := m.MemberEducationalAttainmentManager.Find(ctx, &MemberEducationalAttainment{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member educational attainments")
	}
	for _, memberEducation := range memberEducations {
		if err := m.MemberEducationalAttainmentManager.DeleteByIDWithTx(ctx, tx, memberEducation.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member educational attainment: %s", memberEducation.ID)
		}
	}
	memberContacts, err := m.MemberContactReferenceManager.Find(ctx, &MemberContactReference{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member contact references")
	}
	for _, memberContact := range memberContacts {
		if err := m.MemberContactReferenceManager.DeleteByIDWithTx(ctx, tx, memberContact.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member contact reference: %s", memberContact.ID)
		}
	}
	if err := m.MemberProfileDelete(ctx, tx, memberProfile.ID); err != nil {
		return eris.Wrapf(err, "failed to delete member profile: %s", memberProfile.ID)
	}
	return err
}
