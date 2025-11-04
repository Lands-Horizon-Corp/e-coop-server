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

	for _, f := range filters {
		switch f.Op {
		case OpEq:
			tx = tx.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		case OpGt:
			tx = tx.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
		case OpGte:
			tx = tx.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
		case OpLt:
			tx = tx.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
		case OpLte:
			tx = tx.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
		case OpNe:
			tx = tx.Where(fmt.Sprintf("%s <> ?", f.Field), f.Value)
		case OpIn:
			tx = tx.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
		case OpNotIn:
			tx = tx.Where(fmt.Sprintf("%s NOT IN (?)", f.Field), f.Value)
		case OpLike:
			tx = tx.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			// Case-insensitive LIKE
			tx = tx.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			tx = tx.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			tx = tx.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			tx = tx.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}

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

	for _, f := range filters {
		switch f.Op {
		case OpEq:
			tx = tx.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		case OpGt:
			tx = tx.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
		case OpGte:
			tx = tx.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
		case OpLt:
			tx = tx.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
		case OpLte:
			tx = tx.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
		case OpNe:
			tx = tx.Where(fmt.Sprintf("%s <> ?", f.Field), f.Value)
		case OpIn:
			tx = tx.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
		case OpNotIn:
			tx = tx.Where(fmt.Sprintf("%s NOT IN (?)", f.Field), f.Value)
		case OpLike:
			tx = tx.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			// Case-insensitive LIKE
			tx = tx.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			tx = tx.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			tx = tx.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			tx = tx.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}

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

	// Apply filters
	for _, f := range filters {
		switch f.Op {
		case OpEq:
			db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		case OpGt:
			db = db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
		case OpGte:
			db = db.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
		case OpLt:
			db = db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
		case OpLte:
			db = db.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
		case OpNe:
			db = db.Where(fmt.Sprintf("%s <> ?", f.Field), f.Value)
		case OpIn:
			db = db.Where(fmt.Sprintf("%s IN ?", f.Field), f.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN ?", f.Field), f.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			return 0, eris.Errorf("unsupported filter operation: %s", f.Op)
		}
	}

	var maxValue int
	if err := db.Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", field)).Scan(&maxValue).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to get max value for field: %s with lock", field)
	}

	return maxValue, nil
}
