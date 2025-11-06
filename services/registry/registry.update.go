package registry

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// UpdateByID - Handles ALL types and zero values (false, nil, 0, "", etc.)
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
	if err := r.copyAllFields(fields, existing, id); err != nil {
		return eris.Wrapf(err, "failed to copy fields for entity %s", id)
	}

	// Use Model().Where().Save() to ensure proper WHERE clause and hook firing
	if err := r.Client(context).Model(existing).Where("id = ?", id).Save(existing).Error; err != nil {
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

// copyAllFields - Comprehensive field copying that handles ALL Go types and zero values
func (r *Registry[TData, TResponse, TRequest]) copyAllFields(source, dest any, id uuid.UUID) error {
	sourceVal := reflect.ValueOf(source)
	destVal := reflect.ValueOf(dest)

	// Handle pointer dereferencing
	if sourceVal.Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	if destVal.Kind() == reflect.Ptr {
		destVal = destVal.Elem()
	}

	sourceType := sourceVal.Type()

	for i := 0; i < sourceVal.NumField(); i++ {
		field := sourceType.Field(i)

		// Skip unexported fields, ID field, and embedded structs
		if !field.IsExported() || field.Name == "ID" {
			continue
		}

		sourceField := sourceVal.Field(i)
		destField := destVal.FieldByName(field.Name)

		// Skip if destination field doesn't exist or can't be set
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}

		// Copy the field value - this handles ALL types and zero values
		if err := r.copyFieldValue(sourceField, destField, field.Name); err != nil {
			return eris.Wrapf(err, "failed to copy field %s", field.Name)
		}
	}

	// Ensure ID is set correctly
	idField := destVal.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(id))
	}

	return nil
}

// copyFieldValue - Handles copying of individual field values for ALL types
func (r *Registry[TData, TResponse, TRequest]) copyFieldValue(source, dest reflect.Value, fieldName string) error {
	// Ensure types are assignable
	if !source.Type().AssignableTo(dest.Type()) {
		// Try conversion if direct assignment fails
		if source.Type().ConvertibleTo(dest.Type()) {
			dest.Set(source.Convert(dest.Type()))
			return nil
		}
		return eris.Errorf("cannot assign %s to %s for field %s", source.Type(), dest.Type(), fieldName)
	}

	// Handle different kinds of values
	switch source.Kind() {
	case reflect.Ptr:
		// Handle pointer types (including *bool, *int, *string, etc.)
		if source.IsNil() {
			// Set to nil pointer
			dest.Set(reflect.Zero(dest.Type()))
		} else {
			// Create new pointer and copy the value
			if dest.IsNil() {
				dest.Set(reflect.New(dest.Type().Elem()))
			}
			return r.copyFieldValue(source.Elem(), dest.Elem(), fieldName)
		}
	case reflect.Interface:
		// Handle any types
		if source.IsNil() {
			dest.Set(reflect.Zero(dest.Type()))
		} else {
			dest.Set(source)
		}
	case reflect.Slice, reflect.Array:
		// Handle slices and arrays (including []byte, []string, etc.)
		if source.IsNil() {
			dest.Set(reflect.Zero(dest.Type()))
		} else {
			dest.Set(source)
		}
	case reflect.Map:
		// Handle maps (including map[string]any, etc.)
		if source.IsNil() {
			dest.Set(reflect.Zero(dest.Type()))
		} else {
			dest.Set(source)
		}
	case reflect.Struct:
		// Handle embedded structs (time.Time, custom structs, etc.)
		if source.Type() == dest.Type() {
			dest.Set(source)
		} else {
			// Recursive copy for different struct types
			for i := 0; i < source.NumField(); i++ {
				srcField := source.Field(i)
				destField := dest.FieldByName(source.Type().Field(i).Name)
				if destField.IsValid() && destField.CanSet() {
					if err := r.copyFieldValue(srcField, destField, fieldName+"."+source.Type().Field(i).Name); err != nil {
						return err
					}
				}
			}
		}
	case reflect.Chan, reflect.Func:
		// Handle channels and functions
		dest.Set(source)
	default:
		// Handle all primitive types: bool, int, int8, int16, int32, int64,
		// uint, uint8, uint16, uint32, uint64, uintptr, float32, float64,
		// complex64, complex128, string, and any other basic types
		dest.Set(source)
	}

	return nil
}

// UpdateByIDWithTx - Transaction version that also handles all types
func (r *Registry[TData, TResponse, TRequest]) UpdateByIDWithTx(
	context context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	fields *TData,
	preloads ...string,
) error {
	// First, fetch the existing record within transaction
	existing := new(TData)
	if err := tx.Where("id = ?", id).First(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to find entity %s in transaction", id)
	}

	// Copy ALL fields from fields to existing
	if err := r.copyAllFields(fields, existing, id); err != nil {
		return eris.Wrapf(err, "failed to copy fields for entity %s in transaction", id)
	}

	// Use Model().Where().Save() within transaction to ensure proper WHERE clause
	if err := tx.Model(existing).Where("id = ?", id).Save(existing).Error; err != nil {
		return eris.Wrapf(err, "failed to update entity %s in transaction", id)
	}

	// Reload with preloads within transaction
	if preloads == nil {
		preloads = r.preloads
	}
	reloadDb := tx.Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}

	if err := reloadDb.First(fields).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after update in transaction", id)
	}

	r.OnUpdate(context, fields)
	return nil
}
