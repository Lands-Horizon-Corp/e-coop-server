package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

// Findone
func (r *Registry[TData, TResponse, TRequest]) FindOne(
	context context.Context,
	fields *TData,
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
	if err := db.Where(fields).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find one entity")
	}
	return &entity, nil
}

// FindOneRaw
func (r *Registry[TData, TResponse, TRequest]) FindOneRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	data, err := r.FindOne(context, fields, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find one raw entity")
	}
	return r.ToModel(data), nil
}

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
