package repository

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/publisher"
)

type Repository struct {
	database  *horizon.HorizonDatabase
	publisher *publisher.Publisher
}

func NewRepository(
	database *horizon.HorizonDatabase,
	publisher *publisher.Publisher,
) (*Repository, error) {
	return &Repository{
		database:  database,
		publisher: publisher,
	}, nil
}
