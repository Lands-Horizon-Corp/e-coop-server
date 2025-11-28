package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/gorm"
)

type RegistryParams[TData any, TResponse any, TRequest any] struct {
	Service   *services.HorizonService
	Created   func(*TData) []string
	Updated   func(*TData) []string
	Deleted   func(*TData) []string
	Resource  func(*TData) *TResponse
	Headers   []string
	Preloads  []string
	Threshold int
}

type Registry[TData any, TResponse any, TRequest any] struct {
	service   *services.HorizonService
	created   func(*TData) []string
	updated   func(*TData) []string
	deleted   func(*TData) []string
	resource  func(*TData) *TResponse
	filtering *filter.Handler[TData]
	preloads  []string
	headers   []string
	threshold int
}

func NewRegistry[TData any, TResponse any, TRequest any](
	params RegistryParams[TData, TResponse, TRequest],
) *Registry[TData, TResponse, TRequest] {
	maxDepth := 1
	return &Registry[TData, TResponse, TRequest]{
		service:  params.Service,
		created:  params.Created,
		updated:  params.Updated,
		deleted:  params.Deleted,
		resource: params.Resource,
		preloads: params.Preloads,
		headers:  params.Headers,
		filtering: filter.NewFilter[TData](
			filter.GolangFilteringConfig{
				MaxDepth: &maxDepth,
			},
		),
		threshold: params.Threshold,
	}
}

func (r *Registry[TData, TResponse, TRequest]) Client(context context.Context) *gorm.DB {
	if r.service == nil || r.service.Database == nil {
		return nil
	}
	return r.service.Database.Client().WithContext(context).Model(new(TData))
}
