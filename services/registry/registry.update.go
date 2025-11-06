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
	fieldsToUpdate *TData, // Renamed for clarity: this is just the input data
	preloads ...string,
) error {
	dbClient := r.Client(context) // Pre-configured client with Model(new(TData))

	// 1. Load the existing record into memory so GORM can track it
	existingRecord := new(TData)
	if err := dbClient.Where("id = ?", id).First(existingRecord).Error; err != nil {
		return eris.Wrapf(err, "failed to find entity %s before update", id)
	}

	// 2. Determine which fields were intended to be updated (your original reflection logic)
	t := reflect.TypeOf(new(TData)).Elem()
	fieldNames := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "ID" {
			continue
		}
		fieldNames = append(fieldNames, field.Name)
	}

	// 3. Update the fields on the *loaded* 'existingRecord' using the input data.
	// NOTE: Merging generically requires more complex code (e.g., using a library like 'mergo' or deep reflection).
	// For this example, we will assume you can apply the values here.
	// GORM will automatically mark fields as "dirty" if they change from their original DB values.

	// A common way to merge generically without external libraries is:
	// valFields := reflect.ValueOf(fieldsToUpdate).Elem()
	// valExisting := reflect.ValueOf(existingRecord).Elem()
	// for _, name := range fieldNames {
	//     fieldValue := valFields.FieldByName(name)
	//     if fieldValue.IsValid() && !reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()) {
	//         valExisting.FieldByName(name).Set(fieldValue)
	//     }
	// }

	// GORM provides a built-in helper for this exact merge operation:
	// It copies non-zero values from `fieldsToUpdate` into `existingRecord`
	dbClient.Model(existingRecord).Updates(fieldsToUpdate)

	// 4. Use `Save()` with `Select()` to persist changes and trigger the hook
	// `Save()` checks the primary key on `existingRecord` and runs the BeforeUpdate hook.
	if err := dbClient.Select(fieldNames).Save(existingRecord).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s", id)
	}

	// 5. Reload with preloads (using existingRecord now)
	if preloads == nil {
		preloads = r.preloads
	}
	reloadDb := r.Client(context).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	// Update the original input pointer (fieldsToUpdate) if necessary for the caller
	if err := reloadDb.First(fieldsToUpdate).Error; err != nil {
		return eris.Wrapf(err, "failed to reload entity %s after field update", id)
	}

	r.OnUpdate(context, fieldsToUpdate)
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
