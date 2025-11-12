package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// FindLock - finds entities with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindLock(
	context context.Context,
	tx *gorm.DB,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(fields).Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities with lock")
	}
	return entities, nil
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

// FindFilterLock - finds filtered entities with row-level lock (FOR UPDATE)
func (r *Registry[TData, TResponse, TRequest]) FindFilterLock(
	context context.Context,
	tx *gorm.DB,
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

	// Apply lock to the transaction
	lockedTx := tx.Clauses(clause.Locking{Strength: "UPDATE"})

	result, err := r.filtering.DataGorm(lockedTx, root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities with lock")
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
	if preloads == nil {
		preloads = r.preloads
	}
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

// ===============================
// UNSCOPED METHODS (INCLUDING DELETED RECORDS)
// ===============================

// FindIncludingDeleted - finds entities including soft-deleted ones using Unscoped()
func (r *Registry[TData, TResponse, TRequest]) FindIncludingDeleted(
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
	if err := db.Unscoped().Where(fields).Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find entities including deleted")
	}
	return entities, nil
}

// FindIncludingDeletedRaw - finds entities including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) FindIncludingDeletedRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.FindIncludingDeleted(context, fields, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find raw entities including deleted")
	}
	return r.ToModels(data), nil
}

// FindOnlyDeleted - finds only soft-deleted entities
func (r *Registry[TData, TResponse, TRequest]) FindOnlyDeleted(
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
	if err := db.Unscoped().Where(fields).Where("deleted_at IS NOT NULL").Find(&entities).Error; err != nil {
		return nil, eris.Wrap(err, "failed to find only deleted entities")
	}
	return entities, nil
}

// FindOnlyDeletedRaw - finds only soft-deleted entities and returns response models
func (r *Registry[TData, TResponse, TRequest]) FindOnlyDeletedRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.FindOnlyDeleted(context, fields, preloads...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find only deleted raw entities")
	}
	return r.ToModels(data), nil
}

// FindFilterIncludingDeleted - finds filtered entities including soft-deleted ones
func (r *Registry[TData, TResponse, TRequest]) FindFilterIncludingDeleted(
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

	// Create unscoped client
	unscopedClient := r.Client(context).Unscoped()

	result, err := r.filtering.DataGorm(unscopedClient, root, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities including deleted")
	}
	return result, nil
}

// FindFilterIncludingDeletedRaw - finds filtered entities including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) FindFilterIncludingDeletedRaw(
	context context.Context,
	field *TData,
	root filter.Root,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.FindFilterIncludingDeleted(context, field, root, pageIndex, pageSize, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities including deleted")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}

// FindFilterQueryIncludingDeleted - finds filtered entities from query params including soft-deleted ones
func (r *Registry[TData, TResponse, TRequest]) FindFilterQueryIncludingDeleted(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TData], error) {
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to parse query")
	}
	filterRoot.Preload = preloads

	// Create unscoped client
	unscopedClient := r.Client(context).Unscoped()

	data, err := r.filtering.DataGorm(unscopedClient, filterRoot, pageIndex, pageSize)
	if err != nil {
		return &filter.PaginationResult[TData]{}, eris.Wrap(err, "failed to find filtered entities including deleted")
	}
	return data, nil
}

// FindFilterQueryIncludingDeletedRaw - finds filtered entities from query params including soft-deleted ones and returns response models
func (r *Registry[TData, TResponse, TRequest]) FindFilterQueryIncludingDeletedRaw(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	result, err := r.FindFilterQueryIncludingDeleted(context, ctx, preloads...)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities including deleted")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}
