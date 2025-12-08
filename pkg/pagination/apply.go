package pagination

import (
	"strings"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"gorm.io/gorm"
)

func ApplyPresetConditions(db *gorm.DB, conditions any) *gorm.DB {
	if conditions == nil {
		return db
	}
	return db.Where(conditions)
}

func (f *Pagination[T]) fieldExists(db *gorm.DB, field string) bool {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil || stmt.Schema == nil {
		return false
	}
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		currentSchema := stmt.Schema
		for i, part := range parts {
			fObj := currentSchema.LookUpField(part)
			if fObj == nil {
				fObj = currentSchema.LookUpField(handlers.ToSnakeCase(part))
			}
			if fObj == nil {
				return false
			}
			if i < len(parts)-1 {
				rel, ok := currentSchema.Relationships.Relations[part]
				if !ok || rel.FieldSchema == nil {
					return false
				}
				currentSchema = rel.FieldSchema
			}
		}
		return true
	}
	if stmt.Schema.LookUpField(field) != nil {
		return true
	}
	if stmt.Schema.LookUpField(handlers.ToSnakeCase(field)) != nil {
		return true
	}
	if stmt.Schema.LookUpField(strings.ToLower(field)) != nil {
		return true
	}

	return false
}
