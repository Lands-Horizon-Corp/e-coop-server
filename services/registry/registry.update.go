package registry

import (
	"context"
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
	// First, fetch the existing record
	existing := new(TData)
	if err := r.Client(context).Where("id = ?", id).First(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to find entity %s", id)
	}

	// Copy ALL fields from fields to existing (including zero values)
	// This allows updating to false, null, empty string, etc.
	val := reflect.ValueOf(fields).Elem()
	existingVal := reflect.ValueOf(existing).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Name == "ID" {
			continue
		}
		existingField := existingVal.FieldByName(field.Name)
		if !existingField.CanSet() {
			continue
		}
		newValue := val.Field(i)

		// Remove the IsZero() check to allow false, nil, empty values
		existingField.Set(newValue)
	}

	// Ensure the ID is set on existing record for Save() to work properly
	idField := existingVal.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(id))
	}

	// Use Save() without WHERE clause since we have the ID set
	if err := r.Client(context).Save(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to update entity %s", id)
	}

	// Reload with preloads
	if preloads == nil {
		preloads = r.preloads
	}
	reloadDb := r.Client(context).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}

	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after update", id)
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
	t := reflect.TypeOf(new(TData)).Elem()
	fieldNames := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "ID" {
			continue
		}
		fieldNames = append(fieldNames, field.Name)
	}
	// Perform update with explicit field selection
	db := tx.Model(new(TData)).Where("id = ?", id).Select(fieldNames).Updates(fields)
	if err := db.Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s in transaction", id)
	}
	reloadDb := tx.Model(new(TData)).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update in transaction", id)
	}
	r.OnUpdate(context, fields)
	return nil
}
