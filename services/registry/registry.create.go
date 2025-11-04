package registry

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Create
func (r *Registry[TData, TResponse, TRequest]) Create(
	context context.Context,
	data *TData,
) error {
	db := r.Client(context)
	if err := db.Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create entity")
	}
	return nil
}

// CreateWithTx
func (r *Registry[TData, TResponse, TRequest]) CreateWithTx(
	context context.Context,
	tx *gorm.DB,
	data *TData,
) error {
	if err := tx.Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create entity with transaction")
	}
	return nil
}

// CreateMany
func (r *Registry[TData, TResponse, TRequest]) CreateMany(
	context context.Context,
	data []*TData,
) error {
	db := r.Client(context)
	if err := db.Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create entities")
	}
	return nil
}

// CreateManyWithTx
func (r *Registry[TData, TResponse, TRequest]) CreateManyWithTx(
	context context.Context,
	tx *gorm.DB,
	data []*TData,
) error {
	if err := tx.Create(data).Error; err != nil {
		return eris.Wrap(err, "failed to create entities with transaction")
	}
	return nil
}
