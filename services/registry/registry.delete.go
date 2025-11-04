package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Delete removes a single entity by ID
func (r *Registry[TData, TResponse, TRequest]) Delete(
	context context.Context,
	id uuid.UUID,
) error {
	return r.service.Database.StartTransactionWithContext(context, func(tx *gorm.DB) error {
		return r.DeleteWithTx(context, tx, id)
	})
}

// DeleteWithTx removes a single entity by ID within a transaction
func (r *Registry[TData, TResponse, TRequest]) DeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) error {
	var entity TData
	if err := tx.First(&entity, id).Error; err != nil {
		return eris.Wrap(err, "failed to find entity for delete")
	}
	if err := tx.Delete(&entity).Error; err != nil {
		return eris.Wrap(err, "failed to delete entity with transaction")
	}
	r.OnDelete(context, &entity)
	return nil
}

// BulkDelete removes multiple entities by IDs
func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	context context.Context,
	ids uuid.UUIDs,
) error {
	return r.service.Database.StartTransactionWithContext(context, func(tx *gorm.DB) error {
		return r.BulkDeleteWithTx(context, tx, ids)
	})
}

// BulkDeleteWithTx removes multiple entities by IDs within a transaction
func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	ids uuid.UUIDs,
) error {
	if len(ids) == 0 {
		return nil
	}

	var entities []TData
	// First, fetch the entities to be deleted (validates they exist)
	if err := tx.Find(&entities, ids).Error; err != nil {
		return eris.Wrap(err, "failed to find entities for bulk delete")
	}

	// Check if we found all requested entities
	if len(entities) != len(ids) {
		return eris.New("some entities not found for bulk delete")
	}

	// Then delete them
	if err := tx.Delete(&entities).Error; err != nil {
		return eris.Wrap(err, "failed to bulk delete entities")
	}

	// Trigger OnDelete callbacks for each deleted entity
	for _, data := range entities {
		r.OnDelete(context, &data)
	}

	return nil
}
