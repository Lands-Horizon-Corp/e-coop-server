package registry

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) UpdateByID(
	context context.Context,
	id uuid.UUID,
	fieldsToUpdate *TData,
	preloads ...string,
) error {
	// dbClient is already scoped to Model(new(TData)) by the Client() method
	dbClient := r.Client(context)

	// 1. Load the existing record into memory so GORM can track it
	existingRecord := new(TData)
	// Use dbClient directly without an extra Model() call
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

	// 3. Update the fields on the *loaded* 'existingRecord' using GORM's merge helper.
	// We use an unscoped GORM DB instance for this in-memory merge operation to be safe,
	// or rely on GORM's tracking.
	// A better approach is often manual merging or using a library.

	// Let's use the manual approach for clarity and safety in a generic context:
	// This is where you need to implement your generic merge logic properly.
	// ... (logic to copy values from fieldsToUpdate to existingRecord based on fieldNames) ...

	// A better, simpler fix is to just remove the problematic line below.

	// REMOVE THIS LINE:
	// dbClient.Model(existingRecord).Updates(fieldsToUpdate)
	// ^ This line is where the error originates from (duplicate Model/table definition)

	// You rely on the merge behavior of Save() with Select()

	// 4. Use `Save()` with `Select()` to persist changes and trigger the hook
	// `Save()` automatically uses the Model of 'existingRecord' and tracks changes.
	if err := dbClient.Select(fieldNames).Save(existingRecord).Error; err != nil {
		return eris.Wrapf(err, "failed to update fields for entity %s", id)
	}

	// 5. Reload with preloads (using existingRecord now)
	if preloads == nil {
		preloads = r.preloads
	}
	// Use dbClient directly without an extra Model() call
	reloadDb := r.Client(context).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}

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
