package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// delete
func (r *Registry[TData, TResponse, TRequest]) Delete(
	context context.Context,
	id uuid.UUID,
) error {
	db := r.Client(context)
	var entity TData
	if err := db.Delete(&entity, id).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity")
	}
	return nil
}

// deleteWithTx
func (r *Registry[TData, TResponse, TRequest]) DeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) error {
	var entity TData
	if err := tx.Delete(&entity, id).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity with transaction")
	}
	return nil
}

// bulkDelete
func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	context context.Context,
	ids []uuid.UUID,
) error {
	db := r.Client(context)
	var entity []*TData
	if err := db.Delete(entity, ids).Error; err != nil {
		return eris.Wrap(err, "failed to bulk delete entities")
	}
	return nil
}

// bulkDeleteWithTx
func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	ids []uuid.UUID,
) error {
	var entity []*TData
	if err := tx.Delete(&entity, ids).Error; err != nil {
		return eris.Wrap(err, "failed to bulk delete entities with transaction")
	}
	return nil
}
