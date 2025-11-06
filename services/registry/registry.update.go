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
	if preloads == nil {
		preloads = r.preloads
	}

	// First, get the existing record
	existing := new(TData)
	if err := r.Client(context).Where("id = ?", id).First(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to find entity %s", id)
	}

	// Use reflection to copy fields from input to existing record
	existingValue := reflect.ValueOf(existing).Elem()
	fieldsValue := reflect.ValueOf(fields).Elem()
	fieldsType := reflect.TypeOf(fields).Elem()

	for i := 0; i < fieldsType.NumField(); i++ {
		field := fieldsType.Field(i)
		if field.Name == "ID" || field.Name == "CreatedAt" || field.Name == "CreatedByID" {
			continue // Skip immutable fields
		}

		fieldValue := fieldsValue.Field(i)
		existingField := existingValue.Field(i)

		if existingField.CanSet() && fieldValue.IsValid() {
			existingField.Set(fieldValue)
		}
	}

	// Use Save() instead of Updates() to trigger hooks
	if err := r.Client(context).Save(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to update entity %s", id)
	}

	// Reload with preloads
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
