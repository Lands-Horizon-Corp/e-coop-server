package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func TimesheetManager(service *horizon.HorizonService) *registry.Registry[types.Timesheet, types.TimesheetResponse, types.TimesheetRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Timesheet, types.TimesheetResponse, types.TimesheetRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"User",
			"User.Media",
			"MediaIn", "MediaOut",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Timesheet) *types.TimesheetResponse {
			if data == nil {
				return nil
			}
			var timeOutStr *string
			if data.TimeOut != nil {
				str := data.TimeOut.Format(time.RFC3339)
				timeOutStr = &str
			}
			return &types.TimesheetResponse{
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

		Created: func(data *types.Timesheet) registry.Topics {
			return []string{
				"timesheet.create",
				fmt.Sprintf("timesheet.create.%s", data.ID),
				fmt.Sprintf("timesheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("timesheet.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("timesheet.create.user.%s", data.UserID),
			}
		},
		Updated: func(data *types.Timesheet) registry.Topics {
			return []string{
				"timesheet.update",
				fmt.Sprintf("timesheet.update.%s", data.ID),
				fmt.Sprintf("timesheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("timesheet.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("timesheet.update.user.%s", data.UserID),
			}
		},
		Deleted: func(data *types.Timesheet) registry.Topics {
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

func TimesheetCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}

func GetUserTimesheet(context context.Context, service *horizon.HorizonService, userID, organizationID, branchID uuid.UUID) ([]*types.Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "user_id", Op: query.ModeEqual, Value: userID},
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}

func TimeSheetActiveUsers(context context.Context, service *horizon.HorizonService, organizationID, branchID uuid.UUID) ([]*types.Timesheet, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "time_out", Op: query.ModeIsEmpty, Value: nil},
	}

	return TimesheetManager(service).ArrFind(context, filters, nil)
}
