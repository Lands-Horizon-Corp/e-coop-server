package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// UpdateByID
func (r *Registry[TData, TResponse, TRequest]) UpdateByID(
	context context.Context,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {

	if preloads == nil {
		preloads = r.preloads
	}
	// Perform update with explicit field selection
	if err := r.service.Database.Client().WithContext(context).Where("id = ?", id).Save(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s", id)
	}
	// Reload with preloads
	reloadDb := r.Client(context).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update", id)
	}
	r.OnUpdate(context, fields)
	return nil
}

// UpdateByIDWithTx
func (r *Registry[TData, TResponse, TRequest]) UpdateByIDWithTx(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}
	// Perform update with explicit field selection
	if err := tx.Where("id = ?", id).Save(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s in transaction", id)
	}
	// Reload with preloads
	reloadDb := tx.Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update in transaction", id)
	}
	r.OnUpdate(context, fields)
	return nil
}
