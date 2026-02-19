package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MemberStatusPending   MemberStatus = "pending"
	MemberStatusForReview MemberStatus = "for review"
	MemberStatusVerified  MemberStatus = "verified"
	MemberStatusNotAllowd MemberStatus = "not allowed"
)

const (
	MemberMale   Sex = "male"
	MemberFemale Sex = "female"
)

type (
	Sex string

	MemberStatus string

	MemberProfile struct {
		ID                             uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt                      time.Time             `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID                    *uuid.UUID            `gorm:"type:uuid" json:"created_by,omitempty"`
		CreatedBy                      *User                 `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by_user,omitempty"`
		UpdatedAt                      time.Time             `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID                    *uuid.UUID            `gorm:"type:uuid" json:"updated_by,omitempty"`
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
		IsMutualFundMember             bool                  `gorm:"not null;default:true" json:"is_mutual_fund_member"`
		IsMicroFinanceMember           bool                  `gorm:"not null;default:true" json:"is_micro_finance_member"`
		FirstName                      string                `gorm:"type:varchar(255);not null" json:"first_name"`
		MiddleName                     string                `gorm:"type:varchar(255)" json:"middle_name,omitempty"`
		LastName                       string                `gorm:"type:varchar(255);not null" json:"last_name"`
		FullName                       string                `gorm:"type:varchar(255);not null;index:idx_full_name" json:"full_name"`
		Suffix                         string                `gorm:"type:varchar(50)" json:"suffix,omitempty"`
		BirthDate                      *time.Time            `gorm:"type:date;not null" json:"birthdate"`
		Status                         MemberStatus          `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
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
		BirthPlace                   string                         `gorm:"type:varchar(255)" json:"birth_place,omitempty"`

		Latitude  *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude *float64 `gorm:"type:double precision" json:"longitude,omitempty"`

		Sex Sex `gorm:"type:varchar(10);not null;default:'n/a'" json:"sex"`
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
		Status                         MemberStatus                  `json:"status"`
		Description                    string                        `json:"description"`
		Notes                          string                        `json:"notes"`
		ContactNumber                  string                        `json:"contact_number"`
		OldReferenceID                 string                        `json:"old_reference_id"`
		Passbook                       string                        `json:"passbook"`
		BusinessAddress                string                        `json:"business_address"`
		BusinessContactNumber          string                        `json:"business_contact_number"`
		CivilStatus                    string                        `json:"civil_status"`

		Latitude  *float64 `json:"latitude,omitempty"`
		Longitude *float64 `json:"longitude,omitempty"`

		QRCode *horizon.QRResult `json:"qr_code,omitempty"`

		MemberAddresses              []*MemberAddressResponse               `json:"member_addresses,omitempty"`
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
		BirthPlace                   string                                 `json:"birth_place"`
		Sex                          Sex                                    `json:"sex"`
	}

	MemberProfileRequest struct {
		OrganizationID                 uuid.UUID    `json:"organization_id" validate:"required"`
		BranchID                       uuid.UUID    `json:"branch_id" validate:"required"`
		MediaID                        *uuid.UUID   `json:"media_id,omitempty"`
		SignatureMediaID               *uuid.UUID   `json:"signature_media_id,omitempty"`
		UserID                         uuid.UUID    `json:"user_id" validate:"required"`
		MemberTypeID                   *uuid.UUID   `json:"member_type_id,omitempty"`
		MemberGroupID                  *uuid.UUID   `json:"member_group_id,omitempty"`
		MemberGenderID                 *uuid.UUID   `json:"member_gender_id,omitempty"`
		MemberDepartmentID             *uuid.UUID   `json:"member_department_id,omitempty"`
		MemberCenterID                 *uuid.UUID   `json:"member_center_id,omitempty"`
		MemberOccupationID             *uuid.UUID   `json:"member_occupation_id,omitempty"`
		MemberClassificationID         *uuid.UUID   `json:"member_classification_id,omitempty"`
		MemberVerifiedByEmployeeUserID *uuid.UUID   `json:"member_verified_by_employee_user_id,omitempty"`
		RecruitedByMemberProfileID     *uuid.UUID   `json:"recruited_by_member_profile_id,omitempty"`
		IsClosed                       bool         `json:"is_closed"`
		IsMutualFundMember             bool         `json:"is_mutual_fund_member"`
		IsMicroFinanceMember           bool         `json:"is_micro_finance_member"`
		FirstName                      string       `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName                     string       `json:"middle_name,omitempty"`
		LastName                       string       `json:"last_name" validate:"required,min=1,max=255"`
		FullName                       string       `json:"full_name" validate:"required,min=1,max=255"`
		Suffix                         string       `json:"suffix,omitempty"`
		BirthDate                      time.Time    `json:"birthdate" validate:"required"`
		Status                         MemberStatus `json:"status,omitempty"`
		Description                    string       `json:"description,omitempty"`
		Notes                          string       `json:"notes,omitempty"`
		ContactNumber                  string       `json:"contact_number,omitempty"`
		OldReferenceID                 string       `json:"old_reference_id,omitempty"`
		Passbook                       string       `json:"passbook,omitempty"`
		BusinessAddress                string       `json:"business_address,omitempty"`
		BusinessContactNumber          string       `json:"business_contact_number,omitempty"`
		CivilStatus                    string       `json:"civil_status,omitempty"`
		BirthPlace                     string       `json:"birth_place,omitempty"`
	}

	MemberProfileCoordinatesRequest struct {
		Latitude  float64 `json:"latitude" validate:"required"`
		Longitude float64 `json:"longitude" validate:"required"`
	}

	MemberProfilePersonalInfoRequest struct {
		FirstName      string     `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName     string     `json:"middle_name,omitempty" validate:"max=255"`
		LastName       string     `json:"last_name" validate:"required,min=1,max=255"`
		FullName       string     `json:"full_name,omitempty" validate:"max=255"`
		Suffix         string     `json:"suffix,omitempty" validate:"max=50"`
		MemberGenderID *uuid.UUID `json:"member_gender_id,omitempty"`
		BirthDate      *time.Time `json:"birthdate" validate:"required"`
		BirthPlace     string     `json:"birth_place,omitempty" validate:"max=255"`
		ContactNumber  string     `json:"contact_number,omitempty" validate:"max=255"`

		MediaID          *uuid.UUID `json:"media_id,omitempty"`
		SignatureMediaID *uuid.UUID `json:"signature_media_id,omitempty"`
		CivilStatus      string     `json:"civil_status" validate:"required,oneof=single married widowed separated divorced"` // Adjust the allowed values as needed

		MemberOccupationID    *uuid.UUID `json:"member_occupation_id,omitempty"`
		BusinessAddress       string     `json:"business_address,omitempty" validate:"max=255"`
		BusinessContactNumber string     `json:"business_contact_number,omitempty" validate:"max=255"`
		Notes                 string     `json:"notes,omitempty"`
		Description           string     `json:"description,omitempty"`
		Sex                   Sex        `json:"sex,omitempty" validate:"omitempty,oneof=male female n/a"`
	}

	MemberProfileMembershipInfoRequest struct {
		Passbook                   string       `json:"passbook,omitempty" validate:"max=255"`
		OldReferenceID             string       `json:"old_reference_id,omitempty" validate:"max=50"`
		Status                     MemberStatus `json:"status,omitempty" validate:"max=50"`
		MemberTypeID               *uuid.UUID   `json:"member_type_id,omitempty"`
		MemberGroupID              *uuid.UUID   `json:"member_group_id,omitempty"`
		MemberClassificationID     *uuid.UUID   `json:"member_classification_id,omitempty"`
		MemberCenterID             *uuid.UUID   `json:"member_center_id,omitempty"`
		RecruitedByMemberProfileID *uuid.UUID   `json:"recruited_by_member_profile_id,omitempty"`
		MemberDepartmentID         *uuid.UUID   `json:"member_department_id,omitempty"`
		IsMutualFundMember         bool         `json:"is_mutual_fund_member"`
		IsMicroFinanceMember       bool         `json:"is_micro_finance_member"`
	}

	MemberProfileAccountRequest struct {
		UserID *uuid.UUID `json:"user_id,omitempty"`
	}

	AccountInfo struct {
		Username string `json:"user_name" validate:"required,min=1,max=255"`
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
		BirthDate            *time.Time   `json:"birthdate" validate:"required"`
		ContactNumber        string       `json:"contact_number,omitempty" validate:"max=255"`
		CivilStatus          string       `json:"civil_status" validate:"required,oneof=single married widowed separated divorced"`
		MemberOccupationID   *uuid.UUID   `json:"member_occupation_id,omitempty"`
		Status               MemberStatus `json:"status" validate:"required,max=50"`
		IsMutualFundMember   bool         `json:"is_mutual_fund_member"`
		IsMicroFinanceMember bool         `json:"is_micro_finance_member"`
		MemberTypeID         *uuid.UUID   `json:"member_type_id"`
		AccountInfo          *AccountInfo `json:"new_user_info,omitempty" validate:"omitempty"`
		BirthPlace           string       `json:"birth_place,omitempty" validate:"max=255"`
		Sex                  Sex          `json:"sex,omitempty" validate:"omitempty,oneof=male female n/a"`
		PBAutoGenerated      bool         `json:"pb_auto_generated,omitempty"`
	}

	MemberProfileUserAccountRequest struct {
		Password      string     `json:"password,omitempty" validate:"omitempty,min=6,max=100"`
		Username      string     `json:"user_name" validate:"required,min=1,max=50"`
		FirstName     string     `json:"first_name" validate:"required,min=1,max=50"`
		LastName      string     `json:"last_name" validate:"required,min=1,max=50"`
		MiddleName    string     `json:"middle_name,omitempty" validate:"max=50"`
		FullName      string     `json:"full_name" validate:"required,min=1,max=150"`
		Suffix        string     `json:"suffix,omitempty" validate:"max=20"`
		Email         string     `json:"email" validate:"required,email,max=100"`
		ContactNumber string     `json:"contact_number" validate:"required,max=20"`
		BirthDate     *time.Time `json:"birthdate" validate:"required"`
	}

	MemberTypeCountResponse struct {
		MemberTypeID uuid.UUID          `json:"member_type_id"`
		MemberType   MemberTypeResponse `json:"member_type"`
		Count        int64              `json:"count"`
	}
	MemberProfileDashboardSummaryResponse struct {
		TotalMembers       int64                     `json:"total_members"`
		TotalMaleMembers   int64                     `json:"total_male_members"`
		TotalFemaleMembers int64                     `json:"total_female_members"`
		MemberTypeCounts   []MemberTypeCountResponse `json:"member_type_counts"`
	}
)

func (m *MemberProfile) Address() string {
	address := ""
	if len(m.MemberAddresses) > 0 {
		addr := m.MemberAddresses[0]

		var b strings.Builder
		write := func(s string) {
			if s == "" {
				return
			}
			if b.Len() > 0 {
				b.WriteString(", ")
			}
			b.WriteString(s)
		}

		write(addr.Label)
		write(addr.Address)
		write(addr.Barangay)
		write(addr.ProvinceState)
		write(addr.City)
		if addr.PostalCode != "" {
			write(addr.PostalCode)
		}
		if addr.CountryCode != "" {
			write(addr.CountryCode)
		}
		if addr.Landmark != "" {
			if b.Len() > 0 {
				b.WriteString(" ")
			}
			b.WriteString(fmt.Sprintf("(Landmark: %s)", addr.Landmark))
		}

		address = b.String()
	}
	return address
}
