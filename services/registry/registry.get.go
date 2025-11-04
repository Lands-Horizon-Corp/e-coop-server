package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetByID
func (r *Registry[TData, TResponse, TRequest]) GetByID(
	context context.Context,
	id uuid.UUID,
	preloads ...string,
) (*TData, error) {
	var entity TData
	db := r.Client(context)
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entity by ID")
	}
	return &entity, nil
}

// GetByIDRaw
func (r *Registry[TData, TResponse, TRequest]) GetByIDRaw(
	context context.Context,
	id uuid.UUID,
	preloads ...string,
) (*TResponse, error) {
	data, err := r.GetByID(context, id, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get raw entity by ID")
	}
	return r.ToModel(data), nil
}

// GetByIDLock
func (r *Registry[TData, TResponse, TRequest]) GetByIDLock(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	preloads ...string,
) (*TData, error) {
	var entity TData
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entity by ID with lock")
	}
	return &entity, nil
}
