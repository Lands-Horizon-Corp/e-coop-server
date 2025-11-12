package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
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

// ===============================
// UNSCOPED METHODS (INCLUDING DELETED RECORDS)
// ===============================

// GetByIDIncludingDeleted - gets entity by ID including soft-deleted ones using Unscoped()
func (r *Registry[TData, TResponse, TRequest]) GetByIDIncludingDeleted(
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
	if err := db.Unscoped().Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to get entity by ID including deleted")
	}
	return &entity, nil
}

// GetByIDIncludingDeletedRaw - gets entity by ID including soft-deleted ones and returns response model
func (r *Registry[TData, TResponse, TRequest]) GetByIDIncludingDeletedRaw(
	context context.Context,
	id uuid.UUID,
	preloads ...string,
) (*TResponse, error) {
	data, err := r.GetByIDIncludingDeleted(context, id, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get raw entity by ID including deleted")
	}
	return r.ToModel(data), nil
}
