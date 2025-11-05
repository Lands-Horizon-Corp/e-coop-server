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
