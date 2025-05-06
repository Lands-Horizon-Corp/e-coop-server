package repository

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type FeedbackRepository struct {
	database          *horizon.HorizonDatabase
	feedbackBroadcast *broadcast.FeedbackBroadcast
}

func NewFeedbackRepository(
	database *horizon.HorizonDatabase,
	feedbackBroadcast *broadcast.FeedbackBroadcast,
) (*FeedbackRepository, error) {
	return &FeedbackRepository{
		database:          database,
		feedbackBroadcast: feedbackBroadcast,
	}, nil
}

func (r *FeedbackRepository) Create(data *collection.Feedback) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create feedback")
	}
	r.feedbackBroadcast.OnCreate(data)
	return nil
}

func (r *FeedbackRepository) Update(data *collection.Feedback) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update feedback")
	}
	r.feedbackBroadcast.OnUpdate(data)
	return nil
}

func (r *FeedbackRepository) Delete(data *collection.Feedback) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete feedback")
	}
	r.feedbackBroadcast.OnDelete(data)
	return nil
}

func (r *FeedbackRepository) GetByID(id uuid.UUID) (*collection.Feedback, error) {
	var feedback collection.Feedback
	if err := r.database.Client().First(&feedback, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find feedback with id: %s", id)
	}
	return &feedback, nil
}

func (r *FeedbackRepository) List() ([]*collection.Feedback, error) {
	var feedbacks []*collection.Feedback
	if err := r.database.Client().Find(&feedbacks).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list feedback")
	}
	return feedbacks, nil
}
