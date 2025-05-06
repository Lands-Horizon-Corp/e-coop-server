package repository

import (
	"horizon.com/server/horizon"
	"horizon.com/server/models/broadcast"
	"horizon.com/server/models/collection"

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
