package pagination

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"gorm.io/gorm"
)

func (f *Pagination[T]) query(
	db *gorm.DB,
	filterRoot Root,
) *gorm.DB {
	query := db.Model(new(T))
	query = autoJoinRelatedTables(query, filterRoot.FieldFilters, filterRoot.SortFields)
	if len(filterRoot.Preload) > 0 {
		for _, preloadField := range filterRoot.Preload {
			query = query.Preload(preloadField)
		}
	}
	if len(filterRoot.FieldFilters) > 0 {
		query = f.applysGorm(query, filterRoot)
	}
	hasNestedFields := false
	for _, filter := range filterRoot.FieldFilters {
		if strings.Contains(filter.Field, ".") {
			hasNestedFields = true
			break
		}
	}
	if !hasNestedFields {
		for _, sortField := range filterRoot.SortFields {
			if strings.Contains(sortField.Field, ".") {
				hasNestedFields = true
				break
			}
		}
	}
	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			if !strings.Contains(sortField.Field, ".") && !f.fieldExists(db, sortField.Field) {
				continue
			}
			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			field := sortField.Field
			if strings.Contains(field, ".") {
				parts := strings.Split(field, ".")
				if len(parts) >= 2 {
					parts[0] = handlers.ToPascalCase(parts[0])
					field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
					for i := 2; i < len(parts); i++ {
						field = fmt.Sprintf(`%s."%s"`, field, parts[i])
					}
				}
			} else if mainTableName != "" {
				field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
			}
			query = query.Order(fmt.Sprintf("%s %s", field, order))
		}
	} else {
		if mainTableName != "" {
			query = query.Order(fmt.Sprintf(`"%s"."created_at" DESC`, mainTableName))
		} else {
			query = query.Order("created_at DESC")
		}
	}
	return query
}

func (f *Pagination[T]) applysGorm(db *gorm.DB, filterRoot Root) *gorm.DB {
	if len(filterRoot.FieldFilters) == 0 {
		return db
	}
	hasNestedFields := false
	for _, filter := range filterRoot.FieldFilters {
		if strings.Contains(filter.Field, ".") {
			hasNestedFields = true
			break
		}
	}
	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}
	if filterRoot.Logic == LogicAnd {
		for _, filter := range filterRoot.FieldFilters {
			if strings.Contains(filter.Field, ".") || f.fieldExists(db, filter.Field) {
				condition, values := f.buildConditionWithTableName(filter, mainTableName)
				if condition != "" {
					db = db.Where(condition, values...)
				}
			}
		}
	} else {
		var orConditions []string
		var orValues []any
		for _, filter := range filterRoot.FieldFilters {
			if strings.Contains(filter.Field, ".") || f.fieldExists(db, filter.Field) {
				condition, values := f.buildConditionWithTableName(filter, mainTableName)
				if condition != "" {
					orConditions = append(orConditions, condition)
					orValues = append(orValues, values...)
				}
			}
		}
		if len(orConditions) > 0 {
			db = db.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}
	return db
}

func (f *Pagination[T]) buildConditionWithTableName(filter FieldFilter, mainTableName string) (string, []any) {
	field := filter.Field
	value := filter.Value
	isNestedField := strings.Contains(field, ".")
	if isNestedField {
		parts := strings.Split(field, ".")
		if len(parts) >= 2 {
			parts[0] = handlers.ToPascalCase(parts[0])
			field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
			for i := 2; i < len(parts); i++ {
				field = fmt.Sprintf(`%s."%s"`, field, parts[i])
			}
		}
	} else if mainTableName != "" {
		field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
	}
	switch filter.DataType {
	case DataTypeNumber:
		return f.buildNumberCondition(field, filter.Mode, value)
	case DataTypeText:
		return f.buildTextCondition(field, filter.Mode, value)
	case DataTypeBool:
		return f.buildBoolCondition(field, filter.Mode, value)
	case DataTypeDate:
		return f.buildDateCondition(field, filter.Mode, value)
	case DataTypeTime:
		return f.buildTimeCondition(field, filter.Mode, value)
	default:
		return "", nil
	}

}

func (f *Pagination[T]) buildNumberCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s = ?", field), []any{num}
	case ModeNotEqual:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s != ?", field), []any{num}
	case ModeGT:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s > ?", field), []any{num}
	case ModeGTE:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s >= ?", field), []any{num}
	case ModeLT:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s < ?", field), []any{num}
	case ModeLTE:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s <= ?", field), []any{num}
	case ModeRange:
		rangeVal, err := parseRangeNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rangeVal.From, rangeVal.To}
	}
	return "", nil
}

func (f *Pagination[T]) buildTextCondition(field string, mode Mode, value any) (string, []any) {
	if mode == ModeRange {
		rangeVal, ok := value.(Range)
		if !ok {
			return "", nil
		}
		fromStr, err := parseText(rangeVal.From)
		if err != nil {
			return "", nil
		}
		toStr, err := parseText(rangeVal.To)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{fromStr, toStr}
	}
	str, err := parseText(value)
	if err != nil {
		return "", nil
	}
	switch mode {
	case ModeEqual:
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{str}
	case ModeNotEqual:
		return fmt.Sprintf("LOWER(%s) != LOWER(?)", field), []any{str}
	case ModeContains:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case ModeNotContains:
		return fmt.Sprintf("LOWER(%s) NOT LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case ModeStartsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{str + "%"}
	case ModeEndsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str}
	case ModeIsEmpty:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", field, field), []any{}
	case ModeIsNotEmpty:
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", field, field), []any{}
	case ModeGT:
		return fmt.Sprintf("%s > ?", field), []any{str}
	case ModeGTE, ModeAfter:
		return fmt.Sprintf("%s >= ?", field), []any{str}
	case ModeLT, ModeBefore:
		return fmt.Sprintf("%s < ?", field), []any{str}
	case ModeLTE:
		return fmt.Sprintf("%s <= ?", field), []any{str}
	}
	return "", nil
}

func (f *Pagination[T]) buildBoolCondition(field string, mode Mode, value any) (string, []any) {
	boolVal, err := parseBool(value)
	if err != nil {
		return "", nil
	}
	switch mode {
	case ModeEqual:
		return fmt.Sprintf("%s = ?", field), []any{boolVal}
	case ModeNotEqual:
		return fmt.Sprintf("%s != ?", field), []any{boolVal}
	}
	return "", nil
}

func (f *Pagination[T]) buildDateCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s = ?", field), []any{t}
		}
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{startOfDay, endOfDay}
	case ModeNotEqual:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s != ?", field), []any{t}
		}
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		return fmt.Sprintf("(%s < ? OR %s > ?)", field, field), []any{startOfDay, endOfDay}
	case ModeGTE:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s >= ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s >= ?", field), []any{startOfDay}
		}
	case ModeLT:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s < ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s < ?", field), []any{startOfDay}
		}
	case ModeLTE:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s <= ?", field), []any{t}
		} else {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s <= ?", field), []any{endOfDay}
		}
	case ModeBefore:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s < ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s < ?", field), []any{startOfDay}
		}
	case ModeAfter:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s > ?", field), []any{t}
		} else {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s > ?", field), []any{endOfDay}
		}
	case ModeRange:
		rangeVal, err := parseRangeDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTimeFrom := hasTimeComponent(rangeVal.From)
		hasTimeTo := hasTimeComponent(rangeVal.To)

		if hasTimeFrom && hasTimeTo {
			// Both dates have time components, use exact timestamps
			return fmt.Sprintf("%s >= ? AND %s <= ?", field, field), []any{rangeVal.From, rangeVal.To}
		} else {
			// Date-only range: include entire days from start of From day to end of To day
			startOfFromDay := time.Date(rangeVal.From.Year(), rangeVal.From.Month(), rangeVal.From.Day(), 0, 0, 0, 0, rangeVal.From.Location())
			endOfToDay := time.Date(rangeVal.To.Year(), rangeVal.To.Month(), rangeVal.To.Day(), 23, 59, 59, 999999999, rangeVal.To.Location())
			return fmt.Sprintf("%s >= ? AND %s <= ?", field, field), []any{startOfFromDay, endOfToDay}
		}
	}
	return "", nil
}

func (f *Pagination[T]) buildTimeCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) = ?", field), []any{timeStr}
	case ModeNotEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) != ?", field), []any{timeStr}
	case ModeGT:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) > ?", field), []any{timeStr}
	case ModeGTE, ModeAfter:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) >= ?", field), []any{timeStr}
	case ModeLT, ModeBefore:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) < ?", field), []any{timeStr}
	case ModeLTE:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) <= ?", field), []any{timeStr}
	case ModeRange:
		rangeVal, err := parseRangeTime(value)
		if err != nil {
			return "", nil
		}
		fromStr := rangeVal.From.Format("15:04:05")
		toStr := rangeVal.To.Format("15:04:05")
		return fmt.Sprintf("time(%s) BETWEEN ? AND ?", field), []any{fromStr, toStr}
	}
	return "", nil
}
