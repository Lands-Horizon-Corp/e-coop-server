package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) PermissionTemplateList() ([]*model.PermissionTemplate, error) {
	var permission_templates []*model.PermissionTemplate
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&permission_templates).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list permission_template")
	}
	return permission_templates, nil
}

func (r *Repository) PermissionTemplateCreate(data *model.PermissionTemplate) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create permission_template")
	}
	r.publisher.PermissionTemplateOnCreate(data)
	return nil
}

func (r *Repository) PermissionTemplateUpdate(data *model.PermissionTemplate) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update permission_template")
	}
	r.publisher.PermissionTemplateOnUpdate(data)
	return nil
}

func (r *Repository) PermissionTemplateDelete(data *model.PermissionTemplate) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete permission_template")
	}
	r.publisher.PermissionTemplateOnDelete(data)
	return nil
}

func (r *Repository) PermissionTemplateGetByID(id uuid.UUID) (*model.PermissionTemplate, error) {
	var permission_template model.PermissionTemplate
	if err := r.database.Client().First(&permission_template, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find permission_template with id: %s", id)
	}
	return &permission_template, nil
}

func (r *Repository) PermissionTemplateUpdateCreateTransaction(tx *gorm.DB, data *model.PermissionTemplate) error {
	var existing model.PermissionTemplate
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create permission_template in UpdateCreate")
		}
		r.publisher.PermissionTemplateOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update permission_template in UpdateCreate")
		}
		r.publisher.PermissionTemplateOnUpdate(data)
	}

	return nil
}

func (r *Repository) PermissionTemplateUpdateCreate(data *model.PermissionTemplate) error {
	var existing model.PermissionTemplate
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create permission_template in UpdateCreate")
		}
		r.publisher.PermissionTemplateOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update permission_template in UpdateCreate")
		}
		r.publisher.PermissionTemplateOnUpdate(data)
	}
	return nil
}
