package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) GeneratedReportList() ([]*model.GeneratedReport, error) {
	var generated_reports []*model.GeneratedReport
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&generated_reports).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list generated_report")
	}
	return generated_reports, nil
}

func (r *Repository) GeneratedReportCreate(data *model.GeneratedReport) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create generated_report")
	}
	r.publisher.GeneratedReportOnCreate(data)
	return nil
}

func (r *Repository) GeneratedReportUpdate(data *model.GeneratedReport) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update generated_report")
	}
	r.publisher.GeneratedReportOnUpdate(data)
	return nil
}

func (r *Repository) GeneratedReportDelete(data *model.GeneratedReport) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete generated_report")
	}
	r.publisher.GeneratedReportOnDelete(data)
	return nil
}

func (r *Repository) GeneratedReportGetByID(id uuid.UUID) (*model.GeneratedReport, error) {
	var generated_report model.GeneratedReport
	if err := r.database.Client().First(&generated_report, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find generated_report with id: %s", id)
	}
	return &generated_report, nil
}

func (r *Repository) GeneratedReportUpdateCreateTransaction(tx *gorm.DB, data *model.GeneratedReport) error {
	var existing model.GeneratedReport
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create generated_report in UpdateCreate")
		}
		r.publisher.GeneratedReportOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update generated_report in UpdateCreate")
		}
		r.publisher.GeneratedReportOnUpdate(data)
	}

	return nil
}

func (r *Repository) GeneratedReportUpdateCreate(data *model.GeneratedReport) error {
	var existing model.GeneratedReport
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create generated_report in UpdateCreate")
		}
		r.publisher.GeneratedReportOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update generated_report in UpdateCreate")
		}
		r.publisher.GeneratedReportOnUpdate(data)
	}
	return nil
}
