package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
)

// Count
func (r *Registry[TData, TResponse, TRequest]) Count(
	context context.Context,
	fields *TData,
) (int64, error) {
	var count int64
	if err := r.Client(context).Where(fields).Count(&count).Error; err != nil {
		return 0, eris.Wrap(err, "failed to count entities")
	}
	return count, nil
}

// CountWithSQL
func (r *Registry[TData, TResponse, TRequest]) CountWithSQL(
	context context.Context,
	filters []FilterSQL,
) (int64, error) {
	var count int64
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

	if err := db.Model(new(TData)).Count(&count).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to count entities with %d filters", len(filters))
	}
	return count, nil
}

// CountWithFilter
func (r *Registry[TData, TResponse, TRequest]) CountWithFilter(
	context context.Context,
	field *TData,

	root filter.Root,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (int, error) {
	count, err := r.FindFilter(context, field, root, pageIndex, pageSize, preloads...)
	if err != nil {
		return 0, eris.Wrap(err, "failed to count entities with filter")
	}
	return len(count.Data), nil
}
