package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type ContactUsRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.ContactUsBroadcast
}

func NewContactUsRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.ContactUsBroadcast,
) (*ContactUsRepository, error) {
	return &ContactUsRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *ContactUsRepository) List() ([]*collection.ContactUs, error) {
	var contact_uss []*collection.ContactUs
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&contact_uss).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list contact_us")
	}
	return contact_uss, nil
}

func (r *ContactUsRepository) Create(data *collection.ContactUs) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create contact_us")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *ContactUsRepository) Update(data *collection.ContactUs) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update contact_us")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *ContactUsRepository) Delete(data *collection.ContactUs) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete contact_us")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *ContactUsRepository) GetByID(id uuid.UUID) (*collection.ContactUs, error) {
	var contact_us collection.ContactUs
	if err := r.database.Client().First(&contact_us, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find contact_us with id: %s", id)
	}
	return &contact_us, nil
}

func (r *ContactUsRepository) UpdateCreateTransaction(tx *gorm.DB, data *collection.ContactUs) error {
	var existing collection.ContactUs
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create contact_us in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update contact_us in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}

	return nil
}

func (r *ContactUsRepository) UpdateCreate(data *collection.ContactUs) error {
	var existing collection.ContactUs
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create contact_us in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update contact_us in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}
	return nil
}
