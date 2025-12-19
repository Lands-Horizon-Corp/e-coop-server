package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
)

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStr(
	context context.Context,
	filterValue string,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStr(r.Client(context), filterValue, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationNormalStr(
	context context.Context,
	filterValue string,
	filter *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationNormalStr(r.Client(context), filterValue, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination normal data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationArrayStr(
	context context.Context,
	filterValue string,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationArrayStr(r.Client(context), filterValue, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination array data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStructuredStr(
	context context.Context,
	filterValue string,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStructuredStr(r.Client(context), filterValue, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination structured data: %w", err)
	}
	return r.ToModels(data), nil
}
