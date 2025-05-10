package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) OrganizationList() ([]*model.Organization, error) {
	var organizations []*model.Organization
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&organizations).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list organization")
	}
	return organizations, nil
}

func (r *Repository) OrganizationCreate(data *model.Organization) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create organization")
	}
	r.publisher.OrganizationOnCreate(data)
	return nil
}

func (r *Repository) OrganizationUpdate(data *model.Organization) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update organization")
	}
	r.publisher.OrganizationOnUpdate(data)
	return nil
}

func (r *Repository) OrganizationDelete(data *model.Organization) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete organization")
	}
	r.publisher.OrganizationOnDelete(data)
	return nil
}

func (r *Repository) OrganizationGetByID(id uuid.UUID) (*model.Organization, error) {
	var organization model.Organization
	if err := r.database.Client().First(&organization, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find organization with id: %s", id)
	}
	return &organization, nil
}

func (r *Repository) OrganizationUpdateCreateTransaction(tx *gorm.DB, data *model.Organization) error {
	var existing model.Organization
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization in UpdateCreate")
		}
		r.publisher.OrganizationOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization in UpdateCreate")
		}
		r.publisher.OrganizationOnUpdate(data)
	}

	return nil
}

func (r *Repository) OrganizationUpdateCreate(data *model.Organization) error {
	var existing model.Organization
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization in UpdateCreate")
		}
		r.publisher.OrganizationOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization in UpdateCreate")
		}
		r.publisher.OrganizationOnUpdate(data)
	}
	return nil
}

func (r *Repository) OrganizationDeleteTransaction(tx *gorm.DB, data *model.Organization) error {
	// Check if the organization exists in the database
	var existing model.Organization
	err := tx.First(&existing, "id = ?", data.ID).Error
	if err != nil {
		return eris.Wrapf(err, "organization with id %s not found for deletion", data.ID)
	}

	// Proceed to delete the organization if it exists
	if err := tx.Delete(&existing).Error; err != nil {
		return eris.Wrap(err, "failed to delete organization within transaction")
	}

	// Notify publisher about the organization deletion
	r.publisher.OrganizationOnDelete(&existing)

	return nil
}
