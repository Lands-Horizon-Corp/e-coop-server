package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func UserManager(service *horizon.HorizonService) *registry.Registry[types.User, types.UserResponse, types.UserRegisterRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.User, types.UserResponse, types.UserRegisterRequest]{
		Preloads: []string{
			"Media",
			"SignatureMedia",
			"Footsteps",
			"Footsteps.Media",
			"GeneratedReports",
			"GeneratedReports.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.User) *types.UserResponse {
			context := context.Background()
			if data == nil {
				return nil
			}
			result, err := service.QR.EncodeQR(context, &types.QRUser{
				UserID:        data.ID.String(),
				Email:         data.Email,
				ContactNumber: data.ContactNumber,
				Username:      data.Username,
				Lastname:      *data.LastName,
				Firstname:     *data.FirstName,
				Middlename:    *data.MiddleName,
			}, "user-qr")
			if err != nil {
				return nil
			}
			return &types.UserResponse{
				ID:                data.ID,
				Birthdate:         data.Birthdate.Format("2006-01-02"),
				Username:          data.Username,
				Description:       data.Description,
				FirstName:         data.FirstName,
				MiddleName:        data.MiddleName,
				LastName:          data.LastName,
				Suffix:            data.Suffix,
				Email:             data.Email,
				IsEmailVerified:   data.IsEmailVerified,
				ContactNumber:     data.ContactNumber,
				IsContactVerified: data.IsContactVerified,
				QRCode:            result,
				FullName:          data.FullName,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),

				MediaID:          data.MediaID,
				Media:            MediaManager(service).ToModel(data.Media),
				SignatureMediaID: data.SignatureMediaID,
				SignatureMedia:   MediaManager(service).ToModel(data.SignatureMedia),
				Footsteps:        FootstepManager(service).ToModels(data.Footsteps),
				GeneratedReports: GeneratedReportManager(service).ToModels(data.GeneratedReports),
				Notifications:    NotificationManager(service).ToModels(data.Notification),

				UserOrganizations: UserOrganizationManager(service).ToModels(data.UserOrganizations),
			}
		},

		Created: func(data *types.User) registry.Topics {
			return []string{
				"user.create",
				fmt.Sprintf("user.create.%s", data.ID),
			}
		},
		Updated: func(data *types.User) registry.Topics {
			return []string{
				"user.update",
				fmt.Sprintf("user.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.User) registry.Topics {
			return []string{
				"user.delete",
				fmt.Sprintf("user.delete.%s", data.ID),
			}
		},
	})
}

func GetUserByContactNumber(context context.Context, service *horizon.HorizonService, contactNumber string) (*types.User, error) {
	return UserManager(service).FindOne(context, &types.User{ContactNumber: contactNumber})
}

func GetUserByEmail(context context.Context, service *horizon.HorizonService, email string) (*types.User, error) {
	return UserManager(service).FindOne(context, &types.User{Email: email})
}

func GetUserByUsername(context context.Context, service *horizon.HorizonService, userName string) (*types.User, error) {
	return UserManager(service).FindOne(context, &types.User{Username: userName})
}

func GetUserByIdentifier(context context.Context, service *horizon.HorizonService, identifier string) (*types.User, error) {
	if strings.Contains(identifier, "@") {
		if u, err := GetUserByEmail(context, service, identifier); err == nil {
			return u, nil
		}
	}
	numeric := strings.Trim(identifier, "+-0123456789")
	if numeric == "" {
		if u, err := GetUserByContactNumber(context, service, identifier); err == nil {
			return u, nil
		}
	}
	if u, err := GetUserByUsername(context, service, identifier); err == nil {
		return u, nil
	}
	return nil, eris.New("user not found by email, contact number, or username")
}
