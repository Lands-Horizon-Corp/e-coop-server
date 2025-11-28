package registry

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Exists checks if at least one record exists matching the given filters.
// Uses SELECT EXISTS for optimal performance.
// Returns true if exists, false if not found, or error if query fails.
func (r *Registry[TData, TResponse, TRequest]) Exists(ctx context.Context, filters []FilterSQL) (bool, error) {
	query := r.Client(ctx).Model(new(TData))

	query = r.applySQLFilters(query, filters)

	// Most optimal: SELECT EXISTS (SELECT 1 FROM table WHERE ... LIMIT 1)
	var exists bool
	subQuery := query.Select("1").Limit(1)

	err := r.Client(ctx).Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence")
	}

	return exists, nil
}

// ExistsWithTx checks if at least one record exists using a transaction.
func (r *Registry[TData, TResponse, TRequest]) ExistsWithTx(ctx context.Context, tx *gorm.DB, filters []FilterSQL) (bool, error) {
	query := tx.WithContext(ctx).Model(new(TData))

	query = r.applySQLFilters(query, filters)
	var exists bool
	subQuery := query.Select("1").Limit(1)

	err := tx.Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence in transaction")
	}

	return exists, nil
}

// ExistsByID checks if a record exists by its ID.
// Most optimal for single ID lookups.
func (r *Registry[TData, TResponse, TRequest]) ExistsByID(ctx context.Context, id any) (bool, error) {
	var exists bool

	subQuery := r.Client(ctx).Model(new(TData)).Select("1").Where("id = ?", id).Limit(1)

	err := r.Client(ctx).Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence by ID")
	}

	return exists, nil
}

// ExistsByIDWithTx checks if a record exists by its ID using a transaction.
func (r *Registry[TData, TResponse, TRequest]) ExistsByIDWithTx(ctx context.Context, tx *gorm.DB, id any) (bool, error) {
	var exists bool

	subQuery := tx.WithContext(ctx).Model(new(TData)).Select("1").Where("id = ?", id).Limit(1)

	err := tx.Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence by ID in transaction")
	}

	return exists, nil
}

// ExistsIncludingDeleted checks if at least one record exists (including soft-deleted) matching the given filters.
func (r *Registry[TData, TResponse, TRequest]) ExistsIncludingDeleted(ctx context.Context, filters []FilterSQL) (bool, error) {
	query := r.Client(ctx).Unscoped().Model(new(TData))

	// Apply filters
	query = r.applySQLFilters(query, filters)

	var exists bool
	subQuery := query.Select("1").Limit(1)

	err := r.Client(ctx).Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence including deleted")
	}

	return exists, nil
}

// ExistsIncludingDeletedWithTx checks if at least one record exists (including soft-deleted) using a transaction.
func (r *Registry[TData, TResponse, TRequest]) ExistsIncludingDeletedWithTx(ctx context.Context, tx *gorm.DB, filters []FilterSQL) (bool, error) {
	query := tx.WithContext(ctx).Unscoped().Model(new(TData))

	query = r.applySQLFilters(query, filters)

	var exists bool
	subQuery := query.Select("1").Limit(1)

	err := tx.Raw("SELECT EXISTS (?)", subQuery).Scan(&exists).Error
	if err != nil {
		return false, eris.Wrap(err, "failed to check existence including deleted in transaction")
	}

	return exists, nil
}
