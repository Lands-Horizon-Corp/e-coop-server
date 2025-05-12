package collection

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

type Collection struct {
	broadcast *horizon.HorizonBroadcast
	database  *horizon.HorizonDatabase
	model     *model.Model
}

func NewCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *model.Model,
) (*Collection, error) {
	return &Collection{
		broadcast: broadcast,
		database:  database,
		model:     model,
	}, nil
}
