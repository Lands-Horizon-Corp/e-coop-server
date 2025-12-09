package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
)

func (r *Registry[TData, TResponse, TRequest]) NormalTabular(
	context context.Context,
	filter TData,
	getter func(data *TData) map[string]any,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalTabular(r.Client(context), filter, getter, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrTabular(
	context context.Context,
	getter func(data *TData) map[string]any,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrTabular(r.Client(context), getter, filters, sorts, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredTabular(
	context context.Context,
	filter query.StructuredFilter,
	getter func(data *TData) map[string]any,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredTabular(r.Client(context), filter, getter, r.preload(preloads...)...)
}

// func (r *Registry[TData, TResponse, TRequest]) FilterFieldsCSV(
// 	context context.Context,
// 	query string,
// 	fields *TData,
// 	preloads ...string,
// ) ([]byte, error) {
// 	if preloads == nil {
// 		preloads = r.preloads
// 	}
// 	uuids, ok := parseUUIDArrayFromQuery(query)
// 	if ok && len(uuids) > 0 {
// 		return r.FilterByUUIDsCSV(context, uuids, preloads)
// 	}
// 	filterRoot, _, _, err := parseStringQuery(query)
// 	if err != nil {
// 		return nil, eris.Wrapf(err, "failed to parse string query on no pagination field")
// 	}
// 	filterRoot.Preload = preloads
// 	db := filter.ApplyPresetConditions(r.Client(context), fields)
// 	data, err := r.filtering.GormNoPaginationCSV(db, filterRoot)
// 	if err != nil {
// 		return nil, eris.Wrap(err, "failed to find filtered entities on no pagination field")
// 	}
// 	return data, nil
// }
