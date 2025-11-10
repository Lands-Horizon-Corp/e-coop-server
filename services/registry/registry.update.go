package registry

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// UpdateByID
func (r *Registry[TData, TResponse, TRequest]) UpdateByID(
	context context.Context,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {

	if preloads == nil {
		preloads = r.preloads
	}
	// Perform update with explicit field selection using Update instead of Save
	if err := r.Client(context).Where("id = ?", id).UpdateColumns(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s", id)
	}
	// Reload with preloads
	reloadDb := r.Client(context).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update", id)
	}
	r.OnUpdate(context, fields)
	return nil
}

// UpdateByIDWithTx
func (r *Registry[TData, TResponse, TRequest]) UpdateByIDWithTx(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}

	// Store important field values before save to preserve them after reload
	v := reflect.ValueOf(fields).Elem()
	t := reflect.TypeOf(fields).Elem()
	preservedValues := make(map[string]interface{})

	// Identify foreign key fields that need to be preserved
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Preserve all UUID pointer fields (foreign keys) - both nil and non-nil values
		if field.Type.String() == "*uuid.UUID" {
			if fieldValue.IsNil() {
				preservedValues[field.Name] = (*uuid.UUID)(nil)
			} else {
				preservedValues[field.Name] = fieldValue.Interface()
			}
			fmt.Printf("DEBUG: Preserving %s = %v\n", field.Name, preservedValues[field.Name])
		}
	}

	// Perform update with Save to trigger hooks
	if err := tx.Save(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s in transaction", id)
	}

	// Reload with preloads
	reloadDb := tx.Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update in transaction", id)
	}

	// Restore preserved foreign key values after reload
	v = reflect.ValueOf(fields).Elem()
	for fieldName, value := range preservedValues {
		if field := v.FieldByName(fieldName); field.IsValid() && field.CanSet() {
			fmt.Printf("DEBUG: Restoring %s = %v\n", fieldName, value)
			field.Set(reflect.ValueOf(value))
		}
	}

	r.OnUpdate(context, fields)
	return nil
}
