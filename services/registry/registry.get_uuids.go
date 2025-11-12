package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

// GetByUUIDs retrieves entities by UUID array with preloads
func (r *Registry[TData, TResponse, TRequest]) GetByUUIDs(
	context context.Context,
	query string,
	preloads ...string,
) ([]*TResponse, bool, error) {
	uuids, ok := parseUUIDArrayFromQuery(query)
	if !ok {
		return nil, false, nil
	}

	db := r.Client(context)
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	var data []*TData
	err := db.Where("id IN ?", uuids).Find(&data).Error
	if err != nil {
		return nil, true, eris.Wrap(err, "failed to find entities by UUIDs")
	}

	return r.ToModels(data), true, nil
}

func (r *Registry[TData, TResponse, TRequest]) FilterByUUIDsCSV(
	context context.Context,
	uuids []uuid.UUID,
	preloads []string,
) ([]byte, error) {
	fieldFilters := []filter.FieldFilter{}
	for _, id := range uuids {
		fieldFilters = append(fieldFilters, filter.FieldFilter{
			Field:    "id",
			Value:    id,
			Mode:     filter.ModeEqual,
			DataType: filter.DataTypeText,
		})
	}
	customFilter := filter.Root{
		FieldFilters: fieldFilters,
		Logic:        filter.LogicOr,
		Preload:      preloads,
	}
	data, err := r.filtering.GormNoPaginationCSV(r.Client(context), customFilter)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities by UUIDs")
	}
	return data, nil
}
