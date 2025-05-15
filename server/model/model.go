package model

import (
	"github.com/go-playground/validator"
	"horizon.com/server/horizon"
)

type (
	QRMemberProfile struct {
		Firstname       string `json:"first_name"`
		Lastname        string `json:"last_name"`
		Middlename      string `json:"middle_name"`
		ContactNumber   string `json:"contact_number"`
		MemberProfileID string `json:"member_profile_id"`
		BranchID        string `json:"branch_id"`
		OrganizationID  string `json:"organization_id"`
		Email           string `json:"email"`
	}
	QRInvitationCode struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"user_type"`
		Code           string `json:"code"`
		CurrentUse     int    `json:"current_use"`
		Description    string `json:"description"`
	}

	QRUser struct {
		UserID        string `json:"user_id"`
		Email         string `json:"email"`
		ContactNumber string `json:"contact_number"`
		Username      string `json:"user_name"`
		Name          string `json:"name"`
		Lastname      string `json:"last_name"`
		Firstname     string `json:"firs_tname"`
		Middlename    string `json:"middle_name"`
	}
)
type Model struct {
	validator *validator.Validate
	storage   *horizon.HorizonStorage
	qr        *horizon.HorizonQR
}

func NewModel(
	storage *horizon.HorizonStorage,
	qr *horizon.HorizonQR,
) (*Model, error) {
	return &Model{
		validator: validator.New(),
		storage:   storage,
		qr:        qr,
	}, nil
}
