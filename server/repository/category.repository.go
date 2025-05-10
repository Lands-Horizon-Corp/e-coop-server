package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) CategoryList() ([]*model.Category, error) {
	var categorys []*model.Category
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&categorys).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list category")
	}
	return categorys, nil
}

func (r *Repository) CategoryCreate(data *model.Category) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create category")
	}
	r.publisher.CategoryOnCreate(data)
	return nil
}

func (r *Repository) CategoryUpdate(data *model.Category) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update category")
	}
	r.publisher.CategoryOnUpdate(data)
	return nil
}

func (r *Repository) CategoryDelete(data *model.Category) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete category")
	}
	r.publisher.CategoryOnDelete(data)
	return nil
}

func (r *Repository) CategoryGetByID(id uuid.UUID) (*model.Category, error) {
	var category model.Category
	if err := r.database.Client().First(&category, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find category with id: %s", id)
	}
	return &category, nil
}

func (r *Repository) CategoryUpdateCreateTransaction(tx *gorm.DB, data *model.Category) error {
	var existing model.Category
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create category in UpdateCreate")
		}
		r.publisher.CategoryOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update category in UpdateCreate")
		}
		r.publisher.CategoryOnUpdate(data)
	}

	return nil
}

func (r *Repository) CategoryUpdateCreate(data *model.Category) error {
	var existing model.Category
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create category in UpdateCreate")
		}
		r.publisher.CategoryOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update category in UpdateCreate")
		}
		r.publisher.CategoryOnUpdate(data)
	}
	return nil
}
