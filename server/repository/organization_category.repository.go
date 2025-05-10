package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) OrganizationCategoryList() ([]*model.OrganizationCategory, error) {
	var organization_categorys []*model.OrganizationCategory
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&organization_categorys).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list organization_category")
	}
	return organization_categorys, nil
}

func (r *Repository) OrganizationCategoryCreate(data *model.OrganizationCategory) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create organization_category")
	}
	r.publisher.OrganizationCategoryOnCreate(data)
	return nil
}

func (r *Repository) OrganizationCategoryUpdate(data *model.OrganizationCategory) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update organization_category")
	}
	r.publisher.OrganizationCategoryOnUpdate(data)
	return nil
}

func (r *Repository) OrganizationCategoryDelete(data *model.OrganizationCategory) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete organization_category")
	}
	r.publisher.OrganizationCategoryOnDelete(data)
	return nil
}

func (r *Repository) OrganizationCategoryGetByID(id uuid.UUID) (*model.OrganizationCategory, error) {
	var organization_category model.OrganizationCategory
	if err := r.database.Client().First(&organization_category, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find organization_category with id: %s", id)
	}
	return &organization_category, nil
}

func (r *Repository) OrganizationCategoryUpdateCreateTransaction(tx *gorm.DB, data *model.OrganizationCategory) error {
	var existing model.OrganizationCategory
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization_category in UpdateCreate")
		}
		r.publisher.OrganizationCategoryOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization_category in UpdateCreate")
		}
		r.publisher.OrganizationCategoryOnUpdate(data)
	}

	return nil
}

func (r *Repository) OrganizationCategoryUpdateCreate(data *model.OrganizationCategory) error {
	var existing model.OrganizationCategory
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization_category in UpdateCreate")
		}
		r.publisher.OrganizationCategoryOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization_category in UpdateCreate")
		}
		r.publisher.OrganizationCategoryOnUpdate(data)
	}
	return nil
}
