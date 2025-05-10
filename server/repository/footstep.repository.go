package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/server/model"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Repository) FootstepList() ([]*model.Footstep, error) {
	var footsteps []*model.Footstep
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&footsteps).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list footstep")
	}
	return footsteps, nil
}

func (r *Repository) FootstepCreate(data *model.Footstep) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create footstep")
	}
	r.publisher.FootstepOnCreate(data)
	return nil
}

func (r *Repository) FootstepUpdate(data *model.Footstep) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update footstep")
	}
	r.publisher.FootstepOnUpdate(data)
	return nil
}

func (r *Repository) FootstepDelete(data *model.Footstep) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete footstep")
	}
	r.publisher.FootstepOnDelete(data)
	return nil
}

func (r *Repository) FootstepetByID(id uuid.UUID) (*model.Footstep, error) {
	var footstep model.Footstep
	if err := r.database.Client().First(&footstep, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find footstep with id: %s", id)
	}
	return &footstep, nil
}

func (r *Repository) FootstepUpdateCreateTransaction(tx *gorm.DB, data *model.Footstep) error {
	var existing model.Footstep
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create footstep in UpdateCreate")
		}
		r.publisher.FootstepOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update footstep in UpdateCreate")
		}
		r.publisher.FootstepOnUpdate(data)
	}

	return nil
}

func (r *Repository) FootstepUpdateCreate(data *model.Footstep) error {
	var existing model.Footstep
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create footstep in UpdateCreate")
		}
		r.publisher.FootstepOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update footstep in UpdateCreate")
		}
		r.publisher.FootstepOnUpdate(data)
	}
	return nil
}
