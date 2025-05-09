package repository

import (
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type FeedbackRepository struct {
	database  *horizon.HorizonDatabase
	broadcast *broadcast.FeedbackBroadcast
}

func NewFeedbackRepository(
	database *horizon.HorizonDatabase,
	broadcast *broadcast.FeedbackBroadcast,
) (*FeedbackRepository, error) {
	return &FeedbackRepository{
		database:  database,
		broadcast: broadcast,
	}, nil
}

func (r *FeedbackRepository) List() ([]*collection.Feedback, error) {
	var feedbacks []*collection.Feedback
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&feedbacks).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list feedback")
	}
	return feedbacks, nil
}

func (r *FeedbackRepository) Create(data *collection.Feedback) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create feedback")
	}
	r.broadcast.OnCreate(data)
	return nil
}

func (r *FeedbackRepository) Update(data *collection.Feedback) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update feedback")
	}
	r.broadcast.OnUpdate(data)
	return nil
}

func (r *FeedbackRepository) Delete(data *collection.Feedback) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete feedback")
	}
	r.broadcast.OnDelete(data)
	return nil
}

func (r *FeedbackRepository) GetByID(id uuid.UUID) (*collection.Feedback, error) {
	var feedback collection.Feedback
	if err := r.database.Client().First(&feedback, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find feedback with id: %s", id)
	}
	return &feedback, nil
}

func (r *FeedbackRepository) UpdateCreateTransaction(tx *gorm.DB, data *collection.Feedback) error {
	var existing collection.Feedback
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create feedback in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update feedback in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}

	return nil
}

func (r *FeedbackRepository) UpdateCreate(data *collection.Feedback) error {
	var existing collection.Feedback
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create feedback in UpdateCreate")
		}
		r.broadcast.OnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update feedback in UpdateCreate")
		}
		r.broadcast.OnUpdate(data)
	}
	return nil
}
