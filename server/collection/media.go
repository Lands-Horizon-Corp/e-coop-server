package collection

import (
	"fmt"

	"horizon.com/server/horizon"
	"horizon.com/server/server/manager"
	"horizon.com/server/server/model"
)

type MediaCollection struct {
	Manager manager.CollectionManager[model.Media]
}

func NewMediaCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	mod *model.Model,
) (*MediaCollection, error) {
	manager := manager.NewcollectionManager(
		database,
		broadcast,
		func(data *model.Media) ([]string, any) {
			return []string{"media.create", fmt.Sprintf("media.create.%s", data.ID)}, mod.MediaModel(data)
		},
		func(data *model.Media) ([]string, any) {
			return []string{"media.updated", fmt.Sprintf("media.update.%s", data.ID)}, mod.MediaModel(data)
		},
		func(data *model.Media) ([]string, any) {
			return []string{"media.deleted", fmt.Sprintf("media.delete.%s", data.ID)}, mod.MediaModel(data)
		},
	)
	return &MediaCollection{
		Manager: manager,
	}, nil
}
