package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) OrganizationDailyUsageList() ([]*model.OrganizationDailyUsage, error) {
	var organization_daily_usages []*model.OrganizationDailyUsage
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&organization_daily_usages).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list organization_daily_usage")
	}
	return organization_daily_usages, nil
}

func (r *Repository) OrganizationDailyUsageCreate(data *model.OrganizationDailyUsage) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create organization_daily_usage")
	}
	r.publisher.OrganizationDailyUsageOnCreate(data)
	return nil
}

func (r *Repository) OrganizationDailyUsageUpdate(data *model.OrganizationDailyUsage) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update organization_daily_usage")
	}
	r.publisher.OrganizationDailyUsageOnUpdate(data)
	return nil
}

func (r *Repository) OrganizationDailyUsageDelete(data *model.OrganizationDailyUsage) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete organization_daily_usage")
	}
	r.publisher.OrganizationDailyUsageOnDelete(data)
	return nil
}

func (r *Repository) OrganizationDailyUsageGetByID(id uuid.UUID) (*model.OrganizationDailyUsage, error) {
	var organization_daily_usage model.OrganizationDailyUsage
	if err := r.database.Client().First(&organization_daily_usage, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find organization_daily_usage with id: %s", id)
	}
	return &organization_daily_usage, nil
}

func (r *Repository) OrganizationDailyUsageUpdateCreateTransaction(tx *gorm.DB, data *model.OrganizationDailyUsage) error {
	var existing model.OrganizationDailyUsage
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization_daily_usage in UpdateCreate")
		}
		r.publisher.OrganizationDailyUsageOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization_daily_usage in UpdateCreate")
		}
		r.publisher.OrganizationDailyUsageOnUpdate(data)
	}

	return nil
}

func (r *Repository) OrganizationDailyUsageUpdateCreate(data *model.OrganizationDailyUsage) error {
	var existing model.OrganizationDailyUsage
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create organization_daily_usage in UpdateCreate")
		}
		r.publisher.OrganizationDailyUsageOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update organization_daily_usage in UpdateCreate")
		}
		r.publisher.OrganizationDailyUsageOnUpdate(data)
	}
	return nil
}
