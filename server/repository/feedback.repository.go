package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/server/model"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func (r *Repository) FeedbackList() ([]*model.Feedback, error) {
	var feedbacks []*model.Feedback
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&feedbacks).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list feedback")
	}
	return feedbacks, nil
}

func (r *Repository) FeedbackCreate(data *model.Feedback) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create feedback")
	}
	r.publisher.FeedbackOnCreate(data)
	return nil
}

func (r *Repository) FeedbackUpdate(data *model.Feedback) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update feedback")
	}
	r.publisher.FeedbackOnUpdate(data)
	return nil
}

func (r *Repository) FeedbackDelete(data *model.Feedback) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete feedback")
	}
	r.publisher.FeedbackOnDelete(data)
	return nil
}

func (r *Repository) FeedbackGetByID(id uuid.UUID) (*model.Feedback, error) {
	var feedback model.Feedback
	if err := r.database.Client().First(&feedback, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find feedback with id: %s", id)
	}
	return &feedback, nil
}

func (r *Repository) FeedbackUpdateCreateTransaction(tx *gorm.DB, data *model.Feedback) error {
	var existing model.Feedback
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create feedback in UpdateCreate")
		}
		r.publisher.FeedbackOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update feedback in UpdateCreate")
		}
		r.publisher.FeedbackOnUpdate(data)
	}

	return nil
}

func (r *Repository) FeedbackUpdateCreate(data *model.Feedback) error {
	var existing model.Feedback
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create feedback in UpdateCreate")
		}
		r.publisher.FeedbackOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update feedback in UpdateCreate")
		}
		r.publisher.FeedbackOnUpdate(data)
	}
	return nil
}
