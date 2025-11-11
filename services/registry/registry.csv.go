package registry

import (
	"context"
	"reflect"
	"strings"
	"unicode"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) FilterFieldsCSV(
	context context.Context,
	query string,
	fields *TData,
	preloads ...string,
) ([]byte, error) {
	if preloads == nil {
		preloads = r.preloads
	}
	uuids, ok := parseUUIDArrayFromQuery(query)
	if ok && len(uuids) > 0 {
		return r.FilterByUUIDsCSV(context, uuids, preloads)
	}
	filterRoot, _, _, err := parseStringQuery(query)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to parse string query on no pagination field")
	}
	filterRoot.Preload = preloads
	db := filter.ApplyPresetConditions(r.Client(context), fields)
	data, err := r.filtering.GormNoPaginationCSV(db, filterRoot)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities on no pagination field")
	}
	return data, nil
}

// applySQLFilters safely applies FilterSQL conditions to the database query
func (r *Registry[TData, TResponse, TRequest]) applySQLFilters(db *gorm.DB, filters []FilterSQL) *gorm.DB {
	// Get allowed field names from the TData struct
	allowedFields := r.getAllowedFields()

	for _, f := range filters {
		// Validate field name against whitelist
		if !r.isValidField(f.Field, allowedFields) {
			continue // Skip invalid field names
		}

		// Use GORM's safe query methods instead of string formatting
		switch f.Op {
		case OpEq:
			db = db.Where(f.Field+" = ?", f.Value)
		case OpGt:
			db = db.Where(f.Field+" > ?", f.Value)
		case OpGte:
			db = db.Where(f.Field+" >= ?", f.Value)
		case OpLt:
			db = db.Where(f.Field+" < ?", f.Value)
		case OpLte:
			db = db.Where(f.Field+" <= ?", f.Value)
		case OpNe:
			db = db.Where(f.Field+" <> ?", f.Value)
		case OpIn:
			db = db.Where(f.Field+" IN (?)", f.Value)
		case OpNotIn:
			db = db.Where(f.Field+" NOT IN (?)", f.Value)
		case OpLike:
			db = db.Where(f.Field+" LIKE ?", f.Value)
		case OpILike:
			db = db.Where("LOWER("+f.Field+") LIKE LOWER(?)", f.Value)
		case OpIsNull:
			db = db.Where(f.Field + " IS NULL")
		case OpNotNull:
			db = db.Where(f.Field + " IS NOT NULL")
		}
	}
	return db
}

// getAllowedFields extracts valid field names from the TData struct
func (r *Registry[TData, TResponse, TRequest]) getAllowedFields() map[string]bool {
	var data TData
	allowedFields := make(map[string]bool)

	dataType := reflect.TypeOf(data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Get GORM column name or use field name
		if gormTag := field.Tag.Get("gorm"); gormTag != "" {
			// Parse GORM tag for column name
			if strings.Contains(gormTag, "column:") {
				parts := strings.Split(gormTag, ";")
				for _, part := range parts {
					if strings.HasPrefix(part, "column:") {
						columnName := strings.TrimPrefix(part, "column:")
						allowedFields[columnName] = true
						break
					}
				}
			}
		}

		// Also add the struct field name (snake_case converted)
		fieldName := r.toSnakeCase(field.Name)
		allowedFields[fieldName] = true
	}

	return allowedFields
}

// isValidField checks if a field name is in the whitelist
func (r *Registry[TData, TResponse, TRequest]) isValidField(field string, allowedFields map[string]bool) bool {
	return allowedFields[field]
}

// toSnakeCase converts CamelCase to snake_case
func (r *Registry[TData, TResponse, TRequest]) toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// Updated FilterWithSQLString method
func (r *Registry[TData, TResponse, TRequest]) FilterWithSQLString(
	context context.Context,
	query string,
	filters []FilterSQL,
	sorts []FilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	uuids, ok := parseUUIDArrayFromQuery(query)
	if ok && len(uuids) > 0 {
		return r.FilterByUUIDsCSV(context, uuids, preloads)
	}

	filterRoot, _, _, err := parseStringQuery(query)
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse string query")
	}

	if preloads == nil {
		preloads = r.preloads
	}

	// Start with base client and apply SQL filters safely
	db := r.Client(context)
	db = r.applySQLFilters(db, filters)

	// Convert sorts to filter.SortField format and merge with query sorts
	filterSorts := make([]filter.SortField, len(sorts))
	allowedFields := r.getAllowedFields()

	for i, s := range sorts {
		// Validate sort field
		if r.isValidField(s.Field, allowedFields) {
			filterSorts[i] = filter.SortField{
				Field: s.Field,
				Order: s.Order,
			}
		}
	}

	if len(filterSorts) > 0 {
		filterRoot.SortFields = append(filterRoot.SortFields, filterSorts...)
	}

	filterRoot.Preload = preloads

	// Use the advanced GORM filtering without pagination
	data, err := r.filtering.GormNoPaginationCSV(db, filterRoot)
	if err != nil {
		return nil, eris.Wrap(err, "failed to find filtered entities")
	}

	return data, nil
}
