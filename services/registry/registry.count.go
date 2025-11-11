package registry

import (
	"context"
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
	db = r.applySQLFilters(db, filters)

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
