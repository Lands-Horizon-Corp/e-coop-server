package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/server/model"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Repository) ContactUsList() ([]*model.ContactUs, error) {
	var contact_uss []*model.ContactUs
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&contact_uss).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list contact_us")
	}
	return contact_uss, nil
}

func (r *Repository) ContactUsCreate(data *model.ContactUs) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create contact_us")
	}
	r.publisher.ContactUsOnCreate(data)
	return nil
}

func (r *Repository) ContactUsUpdate(data *model.ContactUs) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update contact_us")
	}
	r.publisher.ContactUsOnUpdate(data)
	return nil
}

func (r *Repository) ContactUsDelete(data *model.ContactUs) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete contact_us")
	}
	r.publisher.ContactUsOnDelete(data)
	return nil
}

func (r *Repository) ContactUsGetByID(id uuid.UUID) (*model.ContactUs, error) {
	var contact_us model.ContactUs
	if err := r.database.Client().First(&contact_us, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find contact_us with id: %s", id)
	}
	return &contact_us, nil
}

func (r *Repository) ContactUsUpdateCreateTransaction(tx *gorm.DB, data *model.ContactUs) error {
	var existing model.ContactUs
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create contact_us in UpdateCreate")
		}
		r.publisher.ContactUsOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update contact_us in UpdateCreate")
		}
		r.publisher.ContactUsOnUpdate(data)
	}

	return nil
}

func (r *Repository) ContactUsUpdateCreate(data *model.ContactUs) error {
	var existing model.ContactUs
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create contact_us in UpdateCreate")
		}
		r.publisher.ContactUsOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update contact_us in UpdateCreate")
		}
		r.publisher.ContactUsOnUpdate(data)
	}
	return nil
}
