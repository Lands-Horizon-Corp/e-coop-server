package core

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func MediaManager(service *horizon.HorizonService) *registry.Registry[types.Media, types.MediaResponse, types.MediaRequest] {
	return registry.GetRegistry(
		registry.RegistryParams[types.Media, types.MediaResponse, types.MediaRequest]{
			Preloads: nil,
			Database: service.Database.Client(),
			Dispatch: func(topics registry.Topics, payload any) error {
				return service.Broker.Dispatch(topics, payload)
			},
			Resource: func(data *types.Media) *types.MediaResponse {
				if data == nil {
					return nil
				}
				temporary, err := service.Storage.GeneratePresignedURL(
					context.Background(),
					&horizon.Storage{
						StorageKey: data.StorageKey,
						BucketName: data.BucketName,
						FileName:   data.FileName,
					},
					time.Hour,
				)
				if err != nil {
					temporary = ""
				}
				return &types.MediaResponse{
					ID:          data.ID,
					CreatedAt:   data.CreatedAt.Format(time.RFC3339),
					UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
					FileName:    data.FileName,
					FileSize:    data.FileSize,
					FileType:    data.FileType,
					StorageKey:  data.StorageKey,
					Key:         data.Key,
					BucketName:  data.BucketName,
					Status:      data.Status,
					Progress:    data.Progress,
					DownloadURL: temporary,
				}
			},
			Created: func(data *types.Media) registry.Topics {
				return []string{"media.create", "media.create." + data.ID.String()}
			},
			Updated: func(data *types.Media) registry.Topics {
				return []string{"media.update", "media.update." + data.ID.String()}
			},
			Deleted: func(data *types.Media) registry.Topics {
				return []string{"media.delete", "media.delete." + data.ID.String()}
			},
		},
	)
}

func MediaDelete(context context.Context, service *horizon.HorizonService, mediaID uuid.UUID) error {
	if mediaID == uuid.Nil {
		return nil
	}
	media, err := MediaManager(service).GetByID(context, mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return nil
	}
	if err := MediaManager(service).Delete(context, media.ID); err != nil {
		return err
	}
	if err := service.Storage.DeleteFile(context, &horizon.Storage{
		FileName:   media.FileName,
		FileSize:   media.FileSize,
		FileType:   media.FileType,
		StorageKey: media.StorageKey,
		BucketName: media.BucketName,
		Status:     "delete",
	}); err != nil {
		return err
	}
	return nil

}
