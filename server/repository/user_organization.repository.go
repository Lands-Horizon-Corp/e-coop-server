package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) UserOrganizationList() ([]*model.UserOrganization, error) {
	var user_organizations []*model.UserOrganization
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&user_organizations).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list user_organization")
	}
	return user_organizations, nil
}

func (r *Repository) UserOrganizationCreate(data *model.UserOrganization) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create user_organization")
	}
	r.publisher.UserOrganizationOnCreate(data)
	return nil
}

func (r *Repository) UserOrganizationUpdate(data *model.UserOrganization) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update user_organization")
	}
	r.publisher.UserOrganizationOnUpdate(data)
	return nil
}

func (r *Repository) UserOrganizationDelete(data *model.UserOrganization) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete user_organization")
	}
	r.publisher.UserOrganizationOnDelete(data)
	return nil
}

func (r *Repository) UserOrganizationGetByID(id uuid.UUID) (*model.UserOrganization, error) {
	var user_organization model.UserOrganization
	if err := r.database.Client().First(&user_organization, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find user_organization with id: %s", id)
	}
	return &user_organization, nil
}

func (r *Repository) UserOrganizationUpdateCreateTransaction(tx *gorm.DB, data *model.UserOrganization) error {
	var existing model.UserOrganization
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user_organization in UpdateCreate")
		}
		r.publisher.UserOrganizationOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user_organization in UpdateCreate")
		}
		r.publisher.UserOrganizationOnUpdate(data)
	}

	return nil
}

func (r *Repository) UserOrganizationUpdateCreate(data *model.UserOrganization) error {
	var existing model.UserOrganization
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create user_organization in UpdateCreate")
		}
		r.publisher.UserOrganizationOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update user_organization in UpdateCreate")
		}
		r.publisher.UserOrganizationOnUpdate(data)
	}
	return nil
}

func (r *Repository) UserOrganizationsCount(organizationID uuid.UUID) (int64, error) {
	var count int64
	if err := r.database.Client().Model(&model.UserOrganization{}).
		Where("organization_id = ?", organizationID).
		Count(&count).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to count user organizations for organization ID: %s", organizationID)
	}
	return count, nil
}
