package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) NotificationList() ([]*model.Notification, error) {
	var notifications []*model.Notification
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list notification")
	}
	return notifications, nil
}

func (r *Repository) NotificationCreate(data *model.Notification) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create notification")
	}
	r.publisher.NotificationOnCreate(data)
	return nil
}

func (r *Repository) NotificationUpdate(data *model.Notification) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update notification")
	}
	r.publisher.NotificationOnUpdate(data)
	return nil
}

func (r *Repository) NotificationDelete(data *model.Notification) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete notification")
	}
	r.publisher.NotificationOnDelete(data)
	return nil
}

func (r *Repository) NotificationGetByID(id uuid.UUID) (*model.Notification, error) {
	var notification model.Notification
	if err := r.database.Client().First(&notification, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find notification with id: %s", id)
	}
	return &notification, nil
}

func (r *Repository) NotificationUpdateCreateTransaction(tx *gorm.DB, data *model.Notification) error {
	var existing model.Notification
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create notification in UpdateCreate")
		}
		r.publisher.NotificationOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update notification in UpdateCreate")
		}
		r.publisher.NotificationOnUpdate(data)
	}

	return nil
}

func (r *Repository) NotificationUpdateCreate(data *model.Notification) error {
	var existing model.Notification
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create notification in UpdateCreate")
		}
		r.publisher.NotificationOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update notification in UpdateCreate")
		}
		r.publisher.NotificationOnUpdate(data)
	}
	return nil
}

func (r *Repository) CountUnviewedNotifications() (int64, error) {
	var count int64
	if err := r.database.Client().
		Model(&model.Notification{}).
		Where("is_viewed = ?", false).
		Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count unviewed notifications")
	}
	return count, nil
}

func (r *Repository) CountUnviewedByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.database.Client().
		Model(&model.Notification{}).
		Where("recipient_user_id = ? AND is_viewed = FALSE", userID).
		Count(&count).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to count unviewed notifications for user %s", userID)
	}
	return count, nil
}

func (r *Repository) NotificationListByUserID(userID uuid.UUID) ([]*model.Notification, error) {
	var notes []*model.Notification
	if err := r.database.Client().
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to list notifications for user %s", userID)
	}
	return notes, nil
}
