package registry

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// GetMax retrieves the maximum value of a field based on filters
func (r *Registry[TData, TResponse, TRequest]) GetMax(
	context context.Context,
	field string,
	filters []FilterSQL,
) (int, error) {
	db := r.Client(context)

	// Apply filters
	db = r.applySQLFilters(db, filters)

	var maxValue int
	if err := db.Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", field)).Scan(&maxValue).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to get max value for field: %s", field)
	}

	return maxValue, nil
}

// GetMaxWithTx retrieves the maximum value of a field based on filters within a transaction
func (r *Registry[TData, TResponse, TRequest]) GetMaxWithTx(
	context context.Context,
	tx *gorm.DB,
	field string,
	filters []FilterSQL,
) (int, error) {
	db := tx

	// Apply filters
	db = r.applySQLFilters(db, filters)

	var maxValue int
	if err := db.Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", field)).Scan(&maxValue).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to get max value for field: %s", field)
	}

	return maxValue, nil
}
