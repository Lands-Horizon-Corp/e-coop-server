package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) MediaList() ([]*model.Media, error) {
	var medias []*model.Media
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&medias).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list media")
	}
	return medias, nil
}

func (r *Repository) MediaCreate(data *model.Media) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create media")
	}
	r.publisher.MediaOnCreate(data)
	return nil
}

func (r *Repository) MediaUpdate(data *model.Media) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update media")
	}
	r.publisher.MediaOnUpdate(data)
	return nil
}

func (r *Repository) MediaDelete(data *model.Media) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete media")
	}
	r.publisher.MediaOnDelete(data)
	return nil
}

func (r *Repository) MediaGetByID(id uuid.UUID) (*model.Media, error) {
	var media model.Media
	if err := r.database.Client().First(&media, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find media with id: %s", id)
	}
	return &media, nil
}

func (r *Repository) MediaUpdateCreateTransaction(tx *gorm.DB, data *model.Media) error {
	var existing model.Media
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create media in UpdateCreate")
		}
		r.publisher.MediaOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update media in UpdateCreate")
		}
		r.publisher.MediaOnUpdate(data)
	}

	return nil
}

func (r *Repository) MediaUpdateCreate(data *model.Media) error {
	var existing model.Media
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create media in UpdateCreate")
		}
		r.publisher.MediaOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update media in UpdateCreate")
		}
		r.publisher.MediaOnUpdate(data)
	}
	return nil
}
