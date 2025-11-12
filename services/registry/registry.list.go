package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
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
func (r *Registry[TData, TResponse, TRequest]) ListRaw(
	context context.Context,
	preloads ...string,
) ([]*TResponse, error) {
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
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	if preloads == nil {
		preloads = r.preloads
	}
	root.Preload = preloads
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
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	if preloads == nil {
		preloads = r.preloads
	}
	root.Preload = preloads
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
func (r *Registry[TData, TResponse, TRequest]) ListFilterQuery(
	context context.Context,
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
	data, err := r.filtering.DataGorm(r.Client(context), filterRoot, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to list filtered entities")
	}
	return data, nil
}

// ListFilterQueryRaw
func (r *Registry[TData, TResponse, TRequest]) ListFilterQueryRaw(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	data, err := r.ListFilterQuery(context, ctx, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to list filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

// ===============================
// UNSCOPED METHODS (INCLUDING DELETED RECORDS)
// ===============================

// ListIncludingDeleted - lists all entities including soft-deleted ones using Unscoped()
func (r *Registry[TData, TResponse, TRequest]) ListIncludingDeleted(
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
	if err := db.Unscoped().Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list entities including deleted")
	}
	return entities, nil
}

// ListIncludingDeletedRaw - lists all entities including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) ListIncludingDeletedRaw(
	context context.Context,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ListIncludingDeleted(context, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to list raw entities including deleted")
	}
	return r.ToModels(data), nil
}

// ListOnlyDeleted - lists only soft-deleted entities
func (r *Registry[TData, TResponse, TRequest]) ListOnlyDeleted(
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
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to list only deleted entities")
	}
	return entities, nil
}

// ListOnlyDeletedRaw - lists only soft-deleted entities and returns response models
func (r *Registry[TData, TResponse, TRequest]) ListOnlyDeletedRaw(
	context context.Context,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ListOnlyDeleted(context, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to list only deleted raw entities")
	}
	return r.ToModels(data), nil
}

// ListFilterIncludingDeleted - lists filtered entities including soft-deleted ones
func (r *Registry[TData, TResponse, TRequest]) ListFilterIncludingDeleted(
	context context.Context,
	root filter.Root,
	pageIndex int, pageSize int,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	if preloads == nil {
		preloads = r.preloads
	}
	root.Preload = preloads

	// Create unscoped client
	unscopedClient := r.Client(context).Unscoped()

	result, err := r.filtering.DataGorm(unscopedClient, root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to list filtered entities including deleted")
	}
	return result, nil
}

// ListFilterIncludingDeletedRaw - lists filtered entities including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) ListFilterIncludingDeletedRaw(
	context context.Context,
	root filter.Root,
	pageIndex int, pageSize int,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.ListFilterIncludingDeleted(context, root, pageIndex, pageSize, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to list filtered entities including deleted")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}

// ListFilterQueryIncludingDeleted - lists filtered entities from query params including soft-deleted ones
func (r *Registry[TData, TResponse, TRequest]) ListFilterQueryIncludingDeleted(
	context context.Context,
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

	// Create unscoped client
	unscopedClient := r.Client(context).Unscoped()

	data, err := r.filtering.DataGorm(unscopedClient, filterRoot, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to list filtered entities including deleted")
	}
	return data, nil
}

// ListFilterQueryIncludingDeletedRaw - lists filtered entities from query params including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) ListFilterQueryIncludingDeletedRaw(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	data, err := r.ListFilterQueryIncludingDeleted(context, ctx, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to list filtered entities including deleted")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}
