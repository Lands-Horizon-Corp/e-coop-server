package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
)

// List
func (r *Registry[TData, TResponse, TRequest]) List(
	context context.Context,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := r.Client(context)
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list entities")
	}
	return entities, nil
}

// ListRaw
func (r *Registry[TData, TResponse, TRequest]) ListRaw(context context.Context,
	preloads ...string) ([]*TResponse, error) {
	data, err := r.List(context, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to list raw entities")
	}
	return r.ToModels(data), nil
}

// ListFilter
func (r *Registry[TData, TResponse, TRequest]) ListFilter(
	context context.Context,
	root filter.Root,
	pageIndex int, pageSize int,
) (*filter.PaginationResult[TData], error) {
	result, err := r.filtering.DataGorm(r.Client(context), root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to list filtered entities")
	}
	return result, nil
}

// ListFilterRaw
func (r *Registry[TData, TResponse, TRequest]) ListFilterRaw(
	context context.Context,
	root filter.Root,
	pageIndex int, pageSize int,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.filtering.Hybrid(r.Client(context), r.threshold, root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to list filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}

// ListFilterQuery

// ListFilterQuery
