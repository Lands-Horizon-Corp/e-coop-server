package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
)

func (r *Registry[TData, TResponse, TRequest]) NoPaginationWithFields(
	context context.Context,
	query string,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	filterRoot, _, _, err := parseStringQuery(query)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to parse string query on no pagination field")
	}
	if preloads == nil {
		preloads = r.preloads
	}
	filterRoot.Preload = preloads
	db := filter.ApplyPresetConditions(r.Client(context), fields)
	data, err := r.filtering.DataGormNoPage(db, filterRoot)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities on no pagination field")
	}
	return r.ToModels(data), nil
}
