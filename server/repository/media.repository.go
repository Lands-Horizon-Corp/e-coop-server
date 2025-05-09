package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
)

type MediaRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.MediaBroadcast
}

func NewMediaRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.MediaBroadcast,
) (*MediaRepository, error) {
	return &MediaRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *MediaRepository) List() ([]*collection.Media, error) {
	var medias []*collection.Media
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&medias).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list media")
	}
	return medias, nil
}

func (r *MediaRepository) Create(data *collection.Media) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create media")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *MediaRepository) Update(data *collection.Media) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update media")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *MediaRepository) Delete(data *collection.Media) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete media")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *MediaRepository) GetByID(id uuid.UUID) (*collection.Media, error) {
	var media collection.Media
	if err := r.database.Client().First(&media, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find media with id: %s", id)
	}
	return &media, nil
}

func (r *MediaRepository) UpdateCreateTransaction(tx *gorm.DB, data *collection.Media) error {
	var existing collection.Media
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create media in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update media in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}

	return nil
}

func (r *MediaRepository) UpdateCreate(data *collection.Media) error {
	var existing collection.Media
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create media in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update media in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}
	return nil
}
