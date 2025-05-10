package repository

import (
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"horizon.com/server/server/model"
)

func (r *Repository) SubscriptionPlanList() ([]*model.SubscriptionPlan, error) {
	var subscription_plans []*model.SubscriptionPlan
	if err := r.database.Client().
		Order("created_at DESC").
		Find(&subscription_plans).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list subscription_plan")
	}
	return subscription_plans, nil
}

func (r *Repository) SubscriptionPlanCreate(data *model.SubscriptionPlan) error {
	if err := r.database.Client().Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create subscription_plan")
	}
	r.publisher.SubscriptionPlanOnCreate(data)
	return nil
}

func (r *Repository) SubscriptionPlanUpdate(data *model.SubscriptionPlan) error {
	if err := r.database.Client().Save(data).Error; err != nil {
		return eris.Wrap(err, "failed to update subscription_plan")
	}
	r.publisher.SubscriptionPlanOnUpdate(data)
	return nil
}

func (r *Repository) SubscriptionPlanDelete(data *model.SubscriptionPlan) error {
	if err := r.database.Client().Delete(data).Error; err != nil {
		return eris.Wrap(err, "failed to delete subscription_plan")
	}
	r.publisher.SubscriptionPlanOnDelete(data)
	return nil
}

func (r *Repository) SubscriptionPlanGetByID(id uuid.UUID) (*model.SubscriptionPlan, error) {
	var subscription_plan model.SubscriptionPlan
	if err := r.database.Client().First(&subscription_plan, "id = ?", id).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find subscription_plan with id: %s", id)
	}
	return &subscription_plan, nil
}

func (r *Repository) SubscriptionPlanUpdateCreateTransaction(tx *gorm.DB, data *model.SubscriptionPlan) error {
	var existing model.SubscriptionPlan
	err := tx.First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := tx.Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create subscription_plan in UpdateCreate")
		}
		r.publisher.SubscriptionPlanOnCreate(data)
	} else {
		if err := tx.Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update subscription_plan in UpdateCreate")
		}
		r.publisher.SubscriptionPlanOnUpdate(data)
	}

	return nil
}

func (r *Repository) SubscriptionPlanUpdateCreate(data *model.SubscriptionPlan) error {
	var existing model.SubscriptionPlan
	err := r.database.Client().First(&existing, "id = ?", data.ID).Error

	if err != nil {
		if err := r.database.Client().Create(data).Error; err != nil {
			return eris.Wrap(err, "failed to create subscription_plan in UpdateCreate")
		}
		r.publisher.SubscriptionPlanOnCreate(data)
	} else {
		if err := r.database.Client().Save(data).Error; err != nil {
			return eris.Wrap(err, "failed to update subscription_plan in UpdateCreate")
		}
		r.publisher.SubscriptionPlanOnUpdate(data)
	}
	return nil
}
