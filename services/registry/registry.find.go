package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// Find
func (r *Registry[TData, TResponse, TRequest]) Find(
	context context.Context,
	fields *TData,
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
	if err := db.Where(fields).Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities")
	}
	return entities, nil
}

// FindRaw
func (r *Registry[TData, TResponse, TRequest]) FindRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.Find(context, fields, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find raw entities")
	}
	return r.ToModels(data), nil
}

// FindFilter
func (r *Registry[TData, TResponse, TRequest]) FindFilter(
	context context.Context,
	field *TData,

	root filter.Root,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	if preloads == nil {
		preloads = r.preloads
	}
	root.Preload = preloads
	result, err := r.filtering.DataGorm(r.Client(context), root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return result, nil
}

// FindFilterRaw
func (r *Registry[TData, TResponse, TRequest]) FindFilterRaw(
	context context.Context,
	field *TData,

	root filter.Root,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.FindFilter(context, field, root, pageIndex, pageSize, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}

// FindFilterQuery
func (r *Registry[TData, TResponse, TRequest]) FindFilterQuery(
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
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return data, nil
}

// FindFilterQueryRaw
func (r *Registry[TData, TResponse, TRequest]) FindFilterQueryRaw(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.FindFilterQuery(context, ctx, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}


