package publisher

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

type Publisher struct {
	broadcast *horizon.HorizonBroadcast
	model     *model.Model
}

func NewPublisher(
	model *model.Model,
	broadcast *horizon.HorizonBroadcast,
) (*Publisher, error) {
	return &Publisher{
		model:     model,
		broadcast: broadcast,
	}, nil
}
