package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Delete(
	context context.Context,
	id any,
) error {
	var entity TData
	if err := r.Client(context).
		Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).
		First(&entity).Error; err != nil {
		return fmt.Errorf("failed to find entity for delete: %w", err)
	}

	if err := r.Client(context).Delete(&entity).Error; err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	r.OnDelete(context, &entity)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) DeleteIncludeDeleted(
	ctx context.Context,
	id any,
) error {
	var entity TData
	if err := r.Client(ctx).
		Unscoped().
		Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).
		First(&entity).Error; err != nil {
		return fmt.Errorf("failed to find entity (including deleted): %w", err)
	}
	if err := r.Client(ctx).
		Unscoped().
		Delete(&entity).Error; err != nil {
		return fmt.Errorf("failed to permanently delete entity: %w", err)
	}
	r.OnDelete(ctx, &entity)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) DeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	id any,
) error {
	var entity TData
	if err := tx.Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).First(&entity).WithContext(context).Error; err != nil {
		return fmt.Errorf("failed to find entity for delete: %w", err)
	}

	if err := tx.Delete(&entity).Error; err != nil {
		return fmt.Errorf("failed to delete entity with transaction: %w", err)
	}

	r.OnDelete(context, &entity)
	return nil
}
func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	ctx context.Context,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		var entity TData
		if err := r.Client(ctx).
			Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).
			First(&entity).Error; err != nil {
			continue
		}
		if err := r.Client(ctx).Delete(&entity).Error; err != nil {
			continue
		}
		r.OnDelete(ctx, &entity)
	}
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	var entities []TData
	if err := tx.Where(fmt.Sprintf("%s IN ?", r.columnDefaultID), ids).Find(&entities).Error; err != nil {
		return fmt.Errorf("failed to find entities for bulk delete: %w", err)
	}

	if len(entities) != len(ids) {
		return fmt.Errorf("some entities not found for bulk delete")
	}
	if err := tx.Where(fmt.Sprintf("%s IN ?", r.columnDefaultID), ids).Delete(&entities).Error; err != nil {
		return fmt.Errorf("failed to bulk delete entities: %w", err)
	}
	for _, data := range entities {
		r.OnDelete(context, &data)
	}

	return nil
}
