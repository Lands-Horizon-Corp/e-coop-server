package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type RegistryParams[TData any, TResponse any, TRequest any] struct {
	Service  *services.HorizonService
	Created  func(*TData) []string
	Updated  func(*TData) []string
	Deleted  func(*TData) []string
	Resource func(*TData) *TResponse
	Preloads []string
}

type Registry[TData any, TResponse any, TRequest any] struct {
	service   *services.HorizonService
	created   func(*TData) []string
	updated   func(*TData) []string
	deleted   func(*TData) []string
	resource  func(*TData) *TResponse
	filtering *filter.Handler[TData]
	preloads  []string
}

func NewRegistry[TData any, TResponse any, TRequest any](
	params RegistryParams[TData, TResponse, TRequest],
) *Registry[TData, TResponse, TRequest] {
	return &Registry[TData, TResponse, TRequest]{
		service:   params.Service,
		created:   params.Created,
		updated:   params.Updated,
		deleted:   params.Deleted,
		resource:  params.Resource,
		preloads:  params.Preloads,
		filtering: filter.NewFilter[TData](),
	}
}

func (r *Registry[TData, TResponse, TRequest]) Client(context context.Context) *gorm.DB {
	if r.service == nil || r.service.Database == nil {
		return nil
	}
	return r.service.Database.Client().WithContext(context).Model(new(TData))
}

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
	result, err := r.filtering.DataGorm(r.Client(context), root, pageIndex, pageSize)
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

// Find
// FindRaw
// FindOne
// FindOneRaw

// GetByID
// GetByIDRaw

// Create
// CreateWithTx
// Update
// UpdateWithTx
// Delete
// DeleteWithTx
// Count

// ToModel converts a data entity to its response representation.
func (c *Registry[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

// ToModels converts a slice of data entities to their response representations.
func (c *Registry[TData, TResponse, TRequest]) ToModels(data []*TData) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := c.ToModel(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}
