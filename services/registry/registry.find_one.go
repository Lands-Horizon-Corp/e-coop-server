package registry

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// FindOneLock - finds one entity with a row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindOneLock(
	context context.Context,
	tx *gorm.DB,
	fields *TData,
	preloads ...string,
) (*TData, error) {
	var entity TData
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(fields).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find one entity with lock")
	}
	return &entity, nil
}
