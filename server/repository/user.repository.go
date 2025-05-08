package repository

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type UserRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.UserBroadcast
}

func NewUserRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.UserBroadcast,
) (*UserRepository, error) {
	return &UserRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *UserRepository) Create(data *collection.User) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create user")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *UserRepository) Update(data *collection.User) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update user")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *UserRepository) Delete(data *collection.User) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete user")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*collection.User, error) {
	var user collection.User
	if err := r.database.Client().First(&user, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user with id: %s", id)
	}
	return &user, nil
}

func (r *UserRepository) List() ([]*collection.User, error) {
	var users []*collection.User
	if err := r.database.Client().Find(&users).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list user")
	}
	return users, nil
}
