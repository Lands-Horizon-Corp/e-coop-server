package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Timesheet struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_timesheet"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_timesheet"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		UserID uuid.UUID `gorm:"type:uuid"`
		User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:RESTRICT;" json:"user,omitempty"`

		MediaInID  *uuid.UUID `gorm:"type:uuid"`
		MediaIn    *Media     `gorm:"foreignKey:MediaInID;constraint:OnDelete:RESTRICT;" json:"media_in,omitempty"`
		MediaOutID *uuid.UUID `gorm:"type:uuid"`
		MediaOut   *Media     `gorm:"foreignKey:MediaOutID;constraint:OnDelete:RESTRICT;" json:"media_out,omitempty"`

		TimeIn  time.Time  `gorm:"not null;default:now()"`
		TimeOut *time.Time `gorm:"default:NULL"`
	}

	TimesheetResponse struct {
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
		UserID         uuid.UUID             `json:"user_id"`
		User           *UserResponse         `json:"user,omitempty"`
		MediaInID      *uuid.UUID            `json:"media_in_id,omitempty"`
		MediaIn        *MediaResponse        `json:"media_in,omitempty"`
		MediaOutID     *uuid.UUID            `json:"media_out_id,omitempty"`
		MediaOut       *MediaResponse        `json:"media_out,omitempty"`
		TimeIn         string                `json:"time_in"`
		TimeOut        *string               `json:"time_out,omitempty"`
	}

	TimesheetRequest struct {
		MediaID *uuid.UUID `json:"media_id,omitempty"`
	}
)

func TimesheetManager(service *horizon.HorizonService) *registry.Registry[Timesheet, TimesheetResponse, TimesheetRequest] {
	return registry.NewRegistry(registry.RegistryParams[Timesheet, TimesheetResponse, TimesheetRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Organization",
			"User",
			"User.Media",
			"MediaIn", "MediaOut",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Timesheet) *TimesheetResponse {
			if data == nil {
				return nil
			}
			var timeOutStr *string
			if data.TimeOut != nil {
				str := data.TimeOut.Format(time.RFC3339)
				timeOutStr = &str
			}
			return &TimesheetResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				UserID:         data.UserID,
				User:           UserManager(service).ToModel(data.User),
				MediaInID:      data.MediaInID,
				MediaIn:        MediaManager(service).ToModel(data.MediaIn),
				MediaOutID:     data.MediaOutID,
				MediaOut:       MediaManager(service).ToModel(data.MediaOut),
				TimeIn:         data.TimeIn.Format(time.RFC3339),
				TimeOut:        timeOutStr,
			}
		},

		Created: func(data *Timesheet) registry.Topics {
			return []string{
				"timesheet.create",
				fmt.Sprintf("timesheet.create.%s", data.ID),
				fmt.Sprintf("timesheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("timesheet.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("timesheet.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *Timesheet) registry.Topics {
			return []string{
				"timesheet.update",
				fmt.Sprintf("timesheet.update.%s", data.ID),
				fmt.Sprintf("timesheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("timesheet.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("timesheet.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *Timesheet) registry.Topics {
			return []string{
				"timesheet.delete",
				fmt.Sprintf("timesheet.delete.%s", data.ID),
				fmt.Sprintf("timesheet.delete.branch.%s", data.BranchID),
				fmt.Sprintf("timesheet.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("timesheet.delete.user.%s", data.UserID),
			}
		},
	})
}

func TimesheetCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}

func GetUserTimesheet(context context.Context, service *horizon.HorizonService, userID, organizationID, branchID uuid.UUID) ([]*Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "user_id", Op: query.ModeEqual, Value: userID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}

func TimeSheetActiveUsers(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "time_out", Op: query.ModeIsEmpty, Value: nil},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}
