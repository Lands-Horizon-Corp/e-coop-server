package core

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type (
	// POST /api/v1/kyc/personal-details
	KYCPersonalDetailsRequest struct {
		Username   string `json:"username" validate:"required,min=3,max=30,alphanum"`
		FirstName  string `json:"first_name" validate:"required,alpha"`
		MiddleName string `json:"middle_name" validate:"omitempty,alpha"`
		LastName   string `json:"last_name" validate:"required,alpha"`
		Gender     string `json:"gender" validate:"required,oneof=male female other"`
	}

	// POST /api/v1/kyc/security-details
	KYCSecurityDetailsRequest struct {
		Email string `json:"email" validate:"required,email"`

		FullName string `json:"full_name" validate:"required"`

		ContactNumber        string `json:"contact_number" validate:"required,e164"`
		Password             string `json:"password" validate:"required,min=8,max=50"`
		PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	}

	// /api/v1/kyc/verify-email
	KYCVerifyEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,max=50"`
		OTP      string `json:"otp" validate:"required,len=6,numeric"`
	}

	// /api/v1/kyc/verify-contact-number
	KYCVerifyContactNumberRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,e164"`
		Password      string `json:"password" validate:"required,min=8,max=50"`
		OTP           string `json:"otp" validate:"required,len=6,numeric"`
	}

	// /api/v1/kyc/verify-addresses
	KYCVerifyAddressesRequest struct {
		Label         string   `json:"label" validate:"required,min=1,max=255"`
		City          string   `json:"city" validate:"required,min=1,max=255"`
		CountryCode   string   `json:"country_code" validate:"required,min=1,max=5"`
		PostalCode    string   `json:"postal_code,omitempty" validate:"omitempty,max=255"`
		ProvinceState string   `json:"province_state,omitempty" validate:"omitempty,max=255"`
		Barangay      string   `json:"barangay,omitempty" validate:"omitempty,max=255"`
		Landmark      string   `json:"landmark,omitempty" validate:"omitempty,max=255"`
		Address       string   `json:"address" validate:"required,min=1,max=255"`
		Longitude     *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
		Latitude      *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	}

	// /api/v1/kyc/verify-government-benefits
	KYCVerifyGovernmentBenefitsRequest struct {
		FrontMediaID *uuid.UUID `json:"front_media_id,omitempty"`
		BackMediaID  *uuid.UUID `json:"back_media_id,omitempty"`
		CountryCode  string     `json:"country_code,omitempty"`
		Description  string     `json:"description,omitempty"`
		Name         string     `json:"name,omitempty"`
		Value        string     `json:"value" validate:"required,min=1,max=254"`
		ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
	}

	// /api/v1/kyc/face-recognize
	KYCFaceRecognizeRequest struct {
		File *multipart.FileHeader `form:"file" validate:"required"`
	}

	// POST /api/v1/kyc/selfie
	KYCSelfieRequest struct {
		File *multipart.FileHeader `form:"file" validate:"required"`
	}

	// POST /api/v1/kyc/resend-email-verification
	KYCResendEmailVerificationRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,max=50"`
		FullName string `json:"full_name" validate:"required,alpha_space"`
	}

	// POST /api/v1/kyc/resend-contact-number-verification
	KYCResendContactNumberVerificationRequest struct {
		ContactNumber string `json:"contact_number" validate:"required,e164"`
		Password      string `json:"password" validate:"required,min=8,max=50"`
		FullName      string `json:"full_name" validate:"required,alpha_space"`
	}

	// POST /api/v1/kyc/register
	KYCRegisterRequest struct {
		Username           string     `json:"username" validate:"required,min=3,max=30,alphanum"`
		FirstName          string     `json:"first_name" validate:"required,alpha"`
		MiddleName         string     `json:"middle_name" validate:"omitempty,alpha"`
		LastName           string     `json:"last_name" validate:"required,alpha"`
		FullName           string     `json:"full_name" validate:"required,alpha_space"`
		ContactNumber      string     `json:"contact_number" validate:"required,e164"`
		Suffix             string     `json:"suffix" validate:"omitempty,alpha"`
		MemberGenderID     uuid.UUID  `json:"member_gender_id" validate:"required,uuid"`
		CivilStatus        string     `json:"civil_status" validate:"required,oneof=single married widowed divorced separated"`
		MemberOccupationID *uuid.UUID `json:"member_occupation_id" validate:"omitempty,uuid"`
		Email              string     `json:"email" validate:"required,email"`
		Phone              string     `json:"phone" validate:"required,e164"`
		BirthDate          *time.Time `json:"birth_date" validate:"omitempty"`
		BranchID           *uuid.UUID `json:"branch_id" validate:"omitempty"`
		Password           string     `json:"password" validate:"required,min=8,max=50"`
		OldPassbook        string     `json:"old_passbook" validate:"omitempty"`

		PasswordConfirmation string                               `json:"password_confirmation" validate:"required,eqfield=Password"`
		Addresses            []KYCVerifyAddressesRequest          `json:"addresses" validate:"required,dive,required"`
		GovernmentBenefits   []KYCVerifyGovernmentBenefitsRequest `json:"government_benefits" validate:"required,dive,required"`
		SelfieMediaID        *uuid.UUID                           `json:"selfie_media_id" validate:"required"`
	}

	// POST /api/v1/kyc/login
	KYCLoginRequest struct {
		Key      string `json:"key" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
)
