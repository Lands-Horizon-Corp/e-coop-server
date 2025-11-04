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
	for _, f := range filters {
		switch f.Op {
		case OpEq:
			db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		case OpGt:
			db = db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
		case OpGte:
			db = db.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
		case OpLt:
			db = db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
		case OpLte:
			db = db.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
		case OpNe:
			db = db.Where(fmt.Sprintf("%s <> ?", f.Field), f.Value)
		case OpIn:
			db = db.Where(fmt.Sprintf("%s IN ?", f.Field), f.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN ?", f.Field), f.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			return 0, eris.Errorf("unsupported filter operation: %s", f.Op)
		}
	}

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
	for _, f := range filters {
		switch f.Op {
		case OpEq:
			db = db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
		case OpGt:
			db = db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
		case OpGte:
			db = db.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
		case OpLt:
			db = db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
		case OpLte:
			db = db.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
		case OpNe:
			db = db.Where(fmt.Sprintf("%s <> ?", f.Field), f.Value)
		case OpIn:
			db = db.Where(fmt.Sprintf("%s IN ?", f.Field), f.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN ?", f.Field), f.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", f.Field), f.Value)
		case OpILike:
			db = db.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", f.Field), f.Value)
		case OpIsNull:
			db = db.Where(fmt.Sprintf("%s IS NULL", f.Field))
		case OpNotNull:
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", f.Field))
		default:
			return 0, eris.Errorf("unsupported filter operation: %s", f.Op)
		}
	}

	var maxValue int
	if err := db.Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", field)).Scan(&maxValue).Error; err != nil {
		return 0, eris.Wrapf(err, "failed to get max value for field: %s", field)
	}

	return maxValue, nil
}
