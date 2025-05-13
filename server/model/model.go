package model

import (
	"github.com/go-playground/validator"
	"horizon.com/server/horizon"
)

type (
	QRMemberProfile struct {
		Firstname       string `json:"firstname"`
		Lastname        string `json:"lastname"`
		Middlename      string `json:"middlename"`
		ContactNumber   string `json:"contact_number"`
		MemberProfileID string `json:"member_profile_id"`
		BranchID        string `json:"branch_id"`
		OrganizationID  string `json:"organization_id"`
		Email           string `json:"email"`
	}
	QRInvitationCode struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"UserType"`
		Code           string `json:"Code"`
		CurrentUse     int    `json:"CurrentUse"`
		Description    string `json:"Description"`
	}

	QRUser struct {
		UserID        string `json:"user_id"`
		Email         string `json:"email"`
		ContactNumber string `json:"contact_number"`
		Username      string `json:"user_name"`
		Name          string `json:"name"`
		Lastname      string `json:"lastname"`
		Firstname     string `json:"firstname"`
		Middlename    string `json:"middlename"`
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
