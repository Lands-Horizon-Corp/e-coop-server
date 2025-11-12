package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

// FindWithSQLLock finds entities matching the provided SQL filters with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindWithSQLLock(
	context context.Context,
	tx *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData

	// Handle joins for related table filters
	joinMap := make(map[string]bool)
	for _, f := range filters {
		// Check if field references a relationship (contains dot)
		if strings.Contains(f.Field, ".") {
			parts := strings.Split(f.Field, ".")
			if len(parts) == 2 {
				relationName := strings.ToUpper(string(parts[0][0])) + parts[0][1:]
				if !joinMap[relationName] {
					tx = tx.Joins(relationName)
					joinMap[relationName] = true
				}
			}
		}
	}

	// Apply locking
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})

	tx = r.applySQLFilters(tx, filters)

	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}

	if len(sorts) > 0 {
		for _, s := range sorts {
			tx = tx.Order(fmt.Sprintf("%s %s", s.Field, s.Order))
		}
	} else {
		tx = tx.Order("updated_at DESC")
	}

	if err := tx.Find(&entities).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entities with %d filters and lock", len(filters))
	}
	return entities, nil
}

// FindOneWithSQLLock finds a single entity matching the provided SQL filters with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindOneWithSQLLock(
	context context.Context,
	tx *gorm.DB,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) (*TData, error) {
	var entity TData

	// Handle joins for related table filters
	joinMap := make(map[string]bool)
	for _, f := range filters {
		// Check if field references a relationship (contains dot)
		if strings.Contains(f.Field, ".") {
			parts := strings.Split(f.Field, ".")
			if len(parts) == 2 {
				relationName := strings.ToUpper(string(parts[0][0])) + parts[0][1:]
				if !joinMap[relationName] {
					tx = tx.Joins(relationName)
					joinMap[relationName] = true
				}
			}
		}
	}

	// Apply locking
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})

	tx = r.applySQLFilters(tx, filters)

	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}

	// Apply sorting
	if len(sorts) > 0 {
		for _, s := range sorts {
			tx = tx.Order(fmt.Sprintf("%s %s", s.Field, s.Order))
		}
	} else {
		tx = tx.Order("updated_at DESC")
	}

	if err := tx.First(&entity).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entity with %d filters and lock", len(filters))
	}
	return &entity, nil
}

// FindFilterQueryLock - finds filtered entities from query params with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindFilterQueryLock(
	context context.Context,
	tx *gorm.DB,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to parse query")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads

	// Apply lock to the transaction
	lockedTx := tx.Clauses(clause.Locking{Strength: "UPDATE"})

	data, err := r.filtering.DataGorm(lockedTx, filterRoot, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities with lock")
	}
	return data, nil
}

// GetByIDLock
func (r *Registry[TData, TResponse, TRequest]) GetByIDLock(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	preloads ...string,
) (*TData, error) {
	var entity TData
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entity by ID with lock")
	}
	return &entity, nil
}

// GetMaxLock retrieves the maximum value of a field with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) GetMaxLock(
	context context.Context,
	tx *gorm.DB,
	field string,
	filters []FilterSQL,
) (int, error) {
	db := tx.Clauses(clause.Locking{Strength: "UPDATE"})

	db = r.applySQLFilters(db, filters)
	var maxValue int
	if err := db.Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", field)).Scan(&maxValue).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to get max value for field: %s with lock", field)
	}

	return maxValue, nil
}
