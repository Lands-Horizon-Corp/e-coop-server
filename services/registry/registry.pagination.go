package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

func (r *Registry[TData, TResponse, TRequest]) PaginationWithFields(
	context context.Context,
	ctx echo.Context,

	fields *TData,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to parse query")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads
	db := filter.ApplyPresetConditions(r.Client(context), fields)
	data, err := r.filtering.DataGorm(
		db,
		filterRoot,
		pageIndex, pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

func (r *Registry[TData, TResponse, TRequest]) Pagination(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to parse query")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads
	data, err := r.filtering.Hybrid(
		r.Client(context),
		r.threshold,
		filterRoot,
		pageIndex, pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

func (r *Registry[TData, TResponse, TRequest]) PaginationData(
	context context.Context,
	ctx echo.Context,
	data []*TData,
) (*filter.PaginationResult[TResponse], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to parse query")
	}
	result, err := r.filtering.DataQuery(
		data,
		filterRoot,
		pageIndex, pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to retrieve paginated data")
	}
	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(result.Data),
		TotalSize: result.TotalSize,
		TotalPage: result.TotalPage,
		PageIndex: result.PageIndex,
		PageSize:  result.PageSize,
	}, nil
}

// PaginationWithSQL provides advanced pagination with SQL filters using the sophisticated GORM filtering
func (r *Registry[TData, TResponse, TRequest]) PaginationWithSQL(
	context context.Context,
	ctx echo.Context,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) (*filter.PaginationResult[TResponse], error) {
	// Parse pagination parameters from context
	_, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to parse query")
	}

	// Use default preloads if none provided
	if preloads == nil {
		preloads = r.preloads
	}

	// Start with base database connection
	db := r.Client(context)

	// Apply filters as preset conditions
	db = r.applySQLFilters(db, filters)

	// Convert sorts to filter.SortField format
	filterSorts := make([]filter.SortField, len(sorts))
	for i, sort := range sorts {
		filterSorts[i] = filter.SortField{
			Field: sort.Field,
			Order: sort.Order,
		}
	}

	// Create filter root with sorts and preloads
	filterRoot := filter.Root{
		SortFields: filterSorts,
		Preload:    preloads,
	}

	// Use the advanced GORM filtering with preset conditions
	data, err := r.filtering.DataGorm(
		db,
		filterRoot,
		pageIndex,
		pageSize,
	)
	if err != nil {
		return &filter.PaginationResult[TResponse]{}, eris.Wrap(err, "failed to find filtered entities")
	}

	return &filter.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}
