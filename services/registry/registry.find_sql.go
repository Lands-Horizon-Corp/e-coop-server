package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
)

// FilterOp represents database filter operations for query conditions.
type FilterOp string

const (
	// OpEq represents the equals operation (=)
	OpEq FilterOp = "="
	// OpGt represents the greater than operation (>)
	OpGt FilterOp = ">"
	// OpGte represents the greater than or equal operation (>=)
	OpGte FilterOp = ">="
	// OpLt represents the less than operation (<)
	OpLt FilterOp = "<"
	// OpLte represents the less than or equal operation (<=)
	OpLte FilterOp = "<="
	// OpNe represents the not equals operation (<>)
	OpNe FilterOp = "<>"
	// OpIn represents the IN operation
	OpIn FilterOp = "IN"
	// OpNotIn represents the NOT IN operation
	OpNotIn FilterOp = "NOT IN"
	// OpLike represents the LIKE operation
	OpLike FilterOp = "LIKE"
	// OpILike represents the case-insensitive LIKE operation
	OpILike FilterOp = "ILIKE"
	// OpIsNull represents the IS NULL operation
	OpIsNull FilterOp = "IS NULL"
	// OpNotNull represents the IS NOT NULL operation
	OpNotNull FilterOp = "IS NOT NULL"
)

// Filter represents a database query filter with field, operation, and value.
type FilterSQL struct {
	Field string
	Op    FilterOp
	Value any
}

type FilterSortSQL struct {
	Field string
	Order filter.SortOrder
}

func (r *Registry[TData, TResponse, TRequest]) FindWithSQL(
	context context.Context,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := r.Client(context)

	// Handle joins for related table filters
	joinMap := make(map[string]bool)
	for _, f := range filters {
		// Check if field references a relationship (contains dot)
		if strings.Contains(f.Field, ".") {
			parts := strings.Split(f.Field, ".")
			if len(parts) == 2 {
				relationName := strings.ToUpper(string(parts[0][0])) + parts[0][1:]
				if !joinMap[relationName] {
					db = db.Joins(relationName)
					joinMap[relationName] = true
				}
			}
		}
	}
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
			db = db.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN (?)", f.Field), f.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			// Case-insensitive LIKE
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			db = db.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		for _, s := range sorts {
			db = db.Order(fmt.Sprintf("%s %s", s.Field, s.Order))
		}
	} else {
		db = db.Order("updated_at DESC")
	}

	if err := db.Find(&entities).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entities with %d filters", len(filters))
	}
	return entities, nil
}

// FindOneWithSQL finds a single entity matching the provided filters
func (r *Registry[TData, TResponse, TRequest]) FindOneWithSQL(
	context context.Context,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) (*TData, error) {
	var entity TData
	db := r.Client(context)

	// Handle joins for related table filters
	joinMap := make(map[string]bool)
	for _, f := range filters {
		// Check if field references a relationship (contains dot)
		if strings.Contains(f.Field, ".") {
			parts := strings.Split(f.Field, ".")
			if len(parts) == 2 {
				relationName := strings.ToUpper(string(parts[0][0])) + parts[0][1:]
				if !joinMap[relationName] {
					db = db.Joins(relationName)
					joinMap[relationName] = true
				}
			}
		}
	}

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
			db = db.Where(fmt.Sprintf("%s IN (?)", f.Field), f.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN (?)", f.Field), f.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			// Case-insensitive LIKE
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			db = db.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}

	if preloads == nil {
		preloads = r.preloads
	}

	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	// Apply sorting
	if len(sorts) > 0 {
		for _, s := range sorts {
			db = db.Order(fmt.Sprintf("%s %s", s.Field, s.Order))
		}
	} else {
		db = db.Order("updated_at DESC")
	}

	if err := db.First(&entity).Error; err != nil {
		return nil, eris.Wrapf(err, "failed to find entity with %d filters", len(filters))
	}
	return &entity, nil
}
