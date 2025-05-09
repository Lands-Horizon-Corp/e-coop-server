package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	Footstep struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Description    string        `gorm:"type:varchar(2048)" json:"description,omitempty"`
		Activity       string        `gorm:"type:varchar(255);unsigned" json:"activity"`
		BranchID       *uuid.UUID    `gorm:"type:uuid"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:SET NULL;"`
		OrganizationID *uuid.UUID    `gorm:"type:uuid"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:SET NULL;"`
		UserID         *uuid.UUID    `gorm:"type:uuid"`
		User           *User         `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID        *uuid.UUID    `gorm:"type:uuid"`
		Media          *Media        `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		// Fields
		AccountType    string    `gorm:"type:varchar(11);unsigned" json:"account_type"`
		Module         string    `gorm:"type:varchar(255);unsigned" json:"module"`
		Latitude       *float64  `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
		Longitude      *float64  `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
		Timestamp      time.Time `gorm:"type:timestamp" json:"timestamp"`
		IsDeleted      bool      `gorm:"default:false" json:"is_deleted"`
		IPAddress      string    `gorm:"type:varchar(45)" json:"ip_address"`
		UserAgent      string    `gorm:"type:varchar(1000)" json:"user_agent"`
		Referer        string    `gorm:"type:varchar(1000)" json:"referer"`
		Location       string    `gorm:"type:varchar(255)" json:"location"`
		AcceptLanguage string    `gorm:"type:varchar(255)" json:"accept_language"`
	}

	// FootstepRequest defines the payload for creating a Footstep
	FootstepRequest struct {
		Description string `json:"description" validate:"required,min=1,max=2048"`
		Activity    string `json:"activity" validate:"required,min=1,max=255"`

		BranchID       *uuid.UUID `json:"branch_id,omitempty"`
		OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
		UserID         *uuid.UUID `json:"user_id,omitempty"`

		AccountType    string     `json:"account_type" validate:"required,min=1,max=11"`
		Module         string     `json:"module" validate:"required,min=1,max=255"`
		Latitude       *float64   `json:"latitude,omitempty"`
		Longitude      *float64   `json:"longitude,omitempty"`
		Timestamp      *time.Time `json:"timestamp,omitempty"` // optional in request
		IPAddress      string     `json:"ip_address,omitempty"`
		UserAgent      string     `json:"user_agent,omitempty"`
		Referer        string     `json:"referer,omitempty"`
		Location       string     `json:"location,omitempty"`
		AcceptLanguage string     `json:"accept_language,omitempty"`
	}

	// FootstepResponse defines the HTTP response for a Footstep
	FootstepResponse struct {
		ID             uuid.UUID             `json:"id"`
		Description    string                `json:"description"`
		Activity       string                `json:"activity"`
		BranchID       *uuid.UUID            `json:"branch_id,omitempty"`
		Branch         *BranchResponse       `json:"branch,omitempty"`
		OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		UserID         *uuid.UUID            `json:"user_id,omitempty"`
		User           *UserResponse         `json:"user,omitempty"`
		MediaID        *uuid.UUID            `gorm:"type:uuid"`
		Media          *MediaResponse        `json:"media,omitempty"`
		CreatedAt      string                `json:"created_at"`
		UpdatedAt      string                `json:"updated_at"`

		AccountType    string   `json:"account_type"`
		Module         string   `json:"module"`
		Latitude       *float64 `json:"latitude,omitempty"`
		Longitude      *float64 `json:"longitude,omitempty"`
		Timestamp      string   `json:"timestamp"`
		IsDeleted      bool     `json:"is_deleted"`
		IPAddress      string   `json:"ip_address"`
		UserAgent      string   `json:"user_agent"`
		Referer        string   `json:"referer"`
		Location       string   `json:"location"`
		AcceptLanguage string   `json:"accept_language"`
	}

	// FootstepCollection handles validation and model mapping
	FootstepCollection struct {
		validator *validator.Validate
		branchCol *BranchCollection
		orgCol    *OrganizationCollection
		mediaCol  *MediaCollection
		userCol   *UserCollection
	}
)

// NewFootstepCollection constructs a FootstepCollection
func NewFootstepCollection(
	branchCol *BranchCollection,
	orgCol *OrganizationCollection,
	mediaCol *MediaCollection,
	userCol *UserCollection,
) (*FootstepCollection, error) {
	return &FootstepCollection{
		validator: validator.New(),
		branchCol: branchCol,
		orgCol:    orgCol,
		mediaCol:  mediaCol,
		userCol:   userCol,
	}, nil
}

// ValidateCreate binds and validates a FootstepRequest
func (fc *FootstepCollection) ValidateCreate(c echo.Context) (*FootstepRequest, error) {
	req := new(FootstepRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := fc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

// ToModel maps a Footstep DB model to a FootstepResponse
func (fc *FootstepCollection) ToModel(f *Footstep) *FootstepResponse {
	if f == nil {
		return nil
	}
	resp := &FootstepResponse{
		ID:             f.ID,
		Description:    f.Description,
		Activity:       f.Activity,
		BranchID:       f.BranchID,
		Branch:         fc.branchCol.ToModel(f.Branch),
		OrganizationID: f.OrganizationID,
		Organization:   fc.orgCol.ToModel(f.Organization),
		UserID:         f.UserID,
		User:           fc.userCol.ToModel(f.User),
		MediaID:        f.MediaID,
		Media:          fc.mediaCol.ToModel(f.Media),
		CreatedAt:      f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      f.UpdatedAt.Format(time.RFC3339),

		AccountType:    f.AccountType,
		Module:         f.Module,
		Latitude:       f.Latitude,
		Longitude:      f.Longitude,
		Timestamp:      f.Timestamp.Format(time.RFC3339),
		IsDeleted:      f.IsDeleted,
		IPAddress:      f.IPAddress,
		UserAgent:      f.UserAgent,
		Referer:        f.Referer,
		Location:       f.Location,
		AcceptLanguage: f.AcceptLanguage,
	}
	return resp
}

// ToModels maps a slice of Footstep DB models to FootstepResponse
func (fc *FootstepCollection) ToModels(data []*Footstep) []*FootstepResponse {
	if data == nil {
		return []*FootstepResponse{}
	}
	var out []*FootstepResponse
	for _, f := range data {
		if m := fc.ToModel(f); m != nil {
			out = append(out, m)
		}
	}
	return out
}
