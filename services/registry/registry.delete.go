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
	r.OnDelete(context, &entity)
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
	r.OnDelete(context, &entity)
	return nil
}

// bulkDelete
func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	context context.Context,
	ids []uuid.UUID,
) error {
	db := r.Client(context)
	var entities []TData
	// First, fetch the entities to be deleted
	if err := db.Find(&entities, ids).Error; err != nil {
		return eris.Wrap(err, "failed to find entities for bulk delete")
	}
	// Then delete them
	if err := db.Delete(&entities, ids).Error; err != nil {
		return eris.Wrap(err, "failed to bulk delete entities")
	}
	for _, data := range entities {
		r.OnDelete(context, &data)
	}
	return nil
}

// bulkDeleteWithTx
func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	ids []uuid.UUID,
) error {
	var entities []TData
	// First, fetch the entities to be deleted
	if err := tx.Find(&entities, ids).Error; err != nil {
		return eris.Wrap(err, "failed to find entities for bulk delete")
	}
	// Then delete them
	if err := tx.Delete(&entities, ids).Error; err != nil {
		return eris.Wrap(err, "failed to bulk delete entities")
	}
	for _, data := range entities {
		r.OnDelete(context, &data)
	}
	return nil
}
