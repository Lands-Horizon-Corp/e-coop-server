package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
)

func (r *Registry[TData, TResponse, TRequest]) FilterFieldsCSV(
	context context.Context,
	query string,
	fields *TData,
	preloads ...string,
) ([]byte, error) {
	if preloads == nil {
		preloads = r.preloads
	}
	uuids, ok := parseUUIDArrayFromQuery(query)
	if ok && len(uuids) > 0 {
		return r.FilterByUUIDsCSV(context, uuids, preloads)
	}
	filterRoot, _, _, err := parseStringQuery(query)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to parse string query on no pagination field")
	}
	filterRoot.Preload = preloads
	db := filter.ApplyPresetConditions(r.Client(context), fields)
	data, err := r.filtering.GormNoPaginationCSV(db, filterRoot)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities on no pagination field")
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) FilterWithSQLString(
	context context.Context,
	query string,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	uuids, ok := parseUUIDArrayFromQuery(query)
	if ok && len(uuids) > 0 {
		return r.FilterByUUIDsCSV(context, uuids, preloads)
	}

	filterRoot, _, _, err := parseStringQuery(query)
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse string query")
	}

	if preloads == nil {
		preloads = r.preloads
	}
	db := r.Client(context)
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
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			db = db.Where(fmt.Sprintf("%s %s ?", f.Field, f.Op), f.Value)
		}
	}

	// Convert sorts to filter.SortField format and merge with query sorts
	filterSorts := make([]filter.SortField, len(sorts))
	for i, s := range sorts {
		filterSorts[i] = filter.SortField{
			Field: s.Field,
			Order: s.Order,
		}
	}
	if len(filterSorts) > 0 {
		filterRoot.SortFields = append(filterRoot.SortFields, filterSorts...)
	}

	filterRoot.Preload = preloads

	// Use the advanced GORM filtering without pagination
	data, err := r.filtering.GormNoPaginationCSV(
		db,
		filterRoot,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities")
	}

	return data, nil
}
