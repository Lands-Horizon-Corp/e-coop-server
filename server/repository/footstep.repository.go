package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type FootstepRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.FootstepBroadcast
}

func NewFootstepRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.FootstepBroadcast,
) (*FootstepRepository, error) {
	return &FootstepRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *FootstepRepository) List() ([]*collection.Footstep, error) {
	var footsteps []*collection.Footstep
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&footsteps).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list footstep")
	}
	return footsteps, nil
}

func (r *FootstepRepository) Create(data *collection.Footstep) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create footstep")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *FootstepRepository) Update(data *collection.Footstep) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update footstep")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *FootstepRepository) Delete(data *collection.Footstep) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete footstep")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *FootstepRepository) GetByID(id uuid.UUID) (*collection.Footstep, error) {
	var footstep collection.Footstep
	if err := r.database.Client().First(&footstep, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find footstep with id: %s", id)
	}
	return &footstep, nil
}

func (r *FootstepRepository) UpdateCreateTransaction(tx *gorm.DB, data *collection.Footstep) error {
	var existing collection.Footstep
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create footstep in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update footstep in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}

	return nil
}

func (r *FootstepRepository) UpdateCreate(data *collection.Footstep) error {
	var existing collection.Footstep
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create footstep in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update footstep in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}
	return nil
}
