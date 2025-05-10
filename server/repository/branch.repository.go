package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) BranchList() ([]*model.Branch, error) {
	var branchs []*model.Branch
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&branchs).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list branch")
	}
	return branchs, nil
}

func (r *Repository) BranchCreate(data *model.Branch) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create branch")
	}
	r.publisher.BranchOnCreate(data)
	return nil
}

func (r *Repository) BranchUpdate(data *model.Branch) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update branch")
	}
	r.publisher.BranchOnUpdate(data)
	return nil
}

func (r *Repository) BranchDelete(data *model.Branch) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete branch")
	}
	r.publisher.BranchOnDelete(data)
	return nil
}

func (r *Repository) BranchGetByID(id uuid.UUID) (*model.Branch, error) {
	var branch model.Branch
	if err := r.database.Client().First(&branch, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find branch with id: %s", id)
	}
	return &branch, nil
}

func (r *Repository) BranchUpdateCreateTransaction(tx *gorm.DB, data *model.Branch) error {
	var existing model.Branch
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create branch in UpdateCreate")
		}
		r.publisher.BranchOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update branch in UpdateCreate")
		}
		r.publisher.BranchOnUpdate(data)
	}

	return nil
}

func (r *Repository) BranchUpdateCreate(data *model.Branch) error {
	var existing model.Branch
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create branch in UpdateCreate")
		}
		r.publisher.BranchOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update branch in UpdateCreate")
		}
		r.publisher.BranchOnUpdate(data)
	}
	return nil
}
