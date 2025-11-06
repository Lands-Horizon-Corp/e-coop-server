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

	// Copy non-zero fields from fields to existing
	// This ensures hooks will fire when using Save()
	val := reflect.ValueOf(fields).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Name == "ID" {
			continue
		}
		existingField := reflect.ValueOf(existing).Elem().FieldByName(field.Name)
		if !existingField.CanSet() {
			continue
		}
		newValue := val.Field(i)
		if !newValue.IsZero() {
			existingField.Set(newValue)
		}
	}

	// Use Save() instead of Updates() to ensure hooks fire
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
