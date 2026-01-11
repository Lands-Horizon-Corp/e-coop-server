package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) Tabular(
	context context.Context,
	filter TData,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalTabular(r.Client(context), filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) RequestTabular(
	context context.Context,
	echocontext echo.Context,
	filter *TData,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalRequestTabular(r.Client(context), echocontext, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StringTabular(
	context context.Context,
	filterValue string,
	filter *TData,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalStringTabular(r.Client(context), filterValue, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrTabular(
	context context.Context,

	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrTabular(r.Client(context), filters, sorts, r.tabular, r.preload(preloads...)...)
}
func (r *Registry[TData, TResponse, TRequest]) ArrRequestTabular(
	context context.Context,
	echocontext echo.Context,
	extraFilters []query.ArrFilterSQL,
	extraSorts []query.ArrFilterSortSQL,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrRequestTabular(r.Client(context), echocontext, extraFilters, extraSorts, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrStringTabular(
	context context.Context,
	filterValue string,
	extraFilters []query.ArrFilterSQL,
	extraSorts []query.ArrFilterSortSQL,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrStringTabular(r.Client(context), filterValue, extraFilters, extraSorts, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredTabular(
	context context.Context,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredTabular(r.Client(context), filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredRequestTabular(
	context context.Context,
	echocontext echo.Context,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredRequestTabular(r.Client(context), echocontext, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredStringTabular(
	context context.Context,
	filterValue string,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredStringTabular(r.Client(context), filterValue, filter, r.tabular, r.preload(preloads...)...)
}
