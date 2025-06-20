package horizon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// --- Filter and Sort Constants and Types ---

type FilterLogic string

const (
	FilterLogicAnd FilterLogic = "AND"
	FilterLogicOr  FilterLogic = "OR"
)

const (
	FilterModeEqual       = "equal"
	FilterModeNotEqual    = "nequal"
	FilterModeContains    = "contains"
	FilterModeNotContains = "ncontains"
	FilterModeStartsWith  = "startswith"
	FilterModeEndsWith    = "endswith"
	FilterModeIsEmpty     = "isempty"
	FilterModeIsNotEmpty  = "isnotempty"
	FilterModeGT          = "gt"
	FilterModeGTE         = "gte"
	FilterModeLT          = "lt"
	FilterModeLTE         = "lte"
	FilterModeRange       = "range"
	FilterModeBefore      = "before"
	FilterModeAfter       = "after"
)

type Filter struct {
	DataType string `json:"dataType"`
	Field    string `json:"field"`
	Mode     string `json:"mode"`
	Value    any    `json:"value"`
}

type FilterRoot struct {
	Filters []Filter    `json:"filters"`
	Logic   FilterLogic `json:"logic"`
}

type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type PaginationResult[T any] struct {
	Data      []*T        `json:"data"`
	PageIndex int         `json:"pageIndex"`
	TotalPage int         `json:"totalPage"`
	PageSize  int         `json:"pageSize"`
	TotalSize int         `json:"totalSize"`
	Sort      []SortField `json:"sort"`
}

func valueToAnything(val reflect.Value) any {
	if !val.IsValid() {
		return nil
	}
	// Unwrap pointer or interface recursively
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	return val.Interface()
}

func findFieldByTagOrName(val reflect.Value, fieldName string) reflect.Value {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			fmt.Printf("findFieldByTagOrName: field=%s, value=<nil pointer>\n", fieldName)
			return reflect.Value{}
		}
		val = val.Elem()
	}
	parts := strings.SplitN(fieldName, ".", 2)
	currentField := parts[0]

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		tagName := strings.Split(tag, ",")[0]
		if (tagName != "" && strings.EqualFold(tagName, currentField)) || strings.EqualFold(f.Name, currentField) {
			fieldVal := val.Field(i)

			if len(parts) == 1 {
				return fieldVal
			}
			// If pointer, check for nil before recursing
			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {

				return reflect.Value{}
			}
			return findFieldByTagOrName(fieldVal, parts[1])
		}
		if f.Anonymous && val.Field(i).Kind() == reflect.Struct {
			found := findFieldByTagOrName(val.Field(i), fieldName)
			if found.IsValid() {
				return found
			}
		}
	}
	fmt.Printf("findFieldByTagOrName: field=%s not found in type=%s\n", fieldName, val.Type().Name())
	return reflect.Value{}
}
func FilterSlice[T any](ctx context.Context, data []*T, filters []Filter, logic FilterLogic) []*T {
	if len(filters) == 0 {
		return data
	}
	var result []*T
	for _, item := range data {
		select {
		case <-ctx.Done():
			return result
		default:
		}
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		matches := logic == FilterLogicAnd
		for _, filter := range filters {
			fieldVal := findFieldByTagOrName(val, filter.Field)

			if !fieldVal.IsValid() {
				// Treat missing field as non-match for this filter
				if logic == FilterLogicAnd {
					matches = false
					break
				}
				// For OR, just skip to next filter
				continue
			}
			itemValue := fieldVal.Interface()
			match := false

			// --- TEXT DATA TYPE HANDLER ---
			if filter.DataType == "text" {
				itemStr := fmt.Sprintf("%v", itemValue)
				filterStr := fmt.Sprintf("%v", filter.Value)
				switch filter.Mode {
				case FilterModeEqual:
					match = strings.EqualFold(itemStr, filterStr)
				case FilterModeNotEqual:
					match = !strings.EqualFold(itemStr, filterStr)
				case FilterModeContains:
					match = strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
				case FilterModeNotContains:
					match = !strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
				case FilterModeStartsWith:
					match = strings.HasPrefix(strings.ToLower(itemStr), strings.ToLower(filterStr))
				case FilterModeEndsWith:
					match = strings.HasSuffix(strings.ToLower(itemStr), strings.ToLower(filterStr))
				default:
					match = false
				}
				if logic == FilterLogicAnd && !match {
					matches = false
					break
				}
				if logic == FilterLogicOr && match {
					matches = true
					break
				}
				continue
			}

			// --- DATE DATA TYPE HANDLER ---
			if filter.DataType == "date" && isWholeDaySupportedMode(filter.Mode) {
				valStr := fmt.Sprintf("%v", filter.Value)
				filterTime, filterErr := tryParseTime(valStr)
				var itemTime time.Time
				var itemErr error
				switch v := itemValue.(type) {
				case string:
					itemTime, itemErr = tryParseTime(v)
				case time.Time:
					itemTime = v
				case *time.Time:
					if v != nil {
						itemTime = *v
					} else {
						itemErr = fmt.Errorf("nil *time.Time")
					}
				default:
					itemErr = fmt.Errorf("not a date type")
				}
				if filterErr == nil && itemErr == nil {
					isWholeDay := filterTime.Hour() == 0 && filterTime.Minute() == 0 && filterTime.Second() == 0 && filterTime.Nanosecond() == 0
					dayStart := filterTime
					dayEnd := filterTime.Add(24 * time.Hour)
					switch filter.Mode {
					case FilterModeEqual:
						if isWholeDay {
							match = !itemTime.Before(dayStart) && itemTime.Before(dayEnd)
						} else {
							match = itemTime.Equal(filterTime)
						}
					case FilterModeNotEqual:
						if isWholeDay {
							match = itemTime.Before(dayStart) || !itemTime.Before(dayEnd)
						} else {
							match = !itemTime.Equal(filterTime)
						}
					case FilterModeGT, FilterModeAfter:
						if isWholeDay {
							match = itemTime.After(dayEnd.Add(-time.Nanosecond))
						} else {
							match = itemTime.After(filterTime)
						}
					case FilterModeGTE:
						if isWholeDay {
							match = !itemTime.Before(dayStart)
						} else {
							match = itemTime.Equal(filterTime) || itemTime.After(filterTime)
						}
					case FilterModeLT, FilterModeBefore:
						if isWholeDay {
							match = itemTime.Before(dayStart)
						} else {
							match = itemTime.Before(filterTime)
						}
					case FilterModeLTE:
						if isWholeDay {
							match = itemTime.Before(dayEnd)
						} else {
							match = itemTime.Equal(filterTime) || itemTime.Before(filterTime)
						}
					}
				} else {
					match = false
				}
				if logic == FilterLogicAnd && !match {
					matches = false
					break
				}
				if logic == FilterLogicOr && match {
					matches = true
					break
				}
				continue
			}

			// --- DEFAULT STRING/NUMBER/BOOL HANDLER ---
			switch filter.Mode {
			case FilterModeEqual, FilterModeNotEqual,
				FilterModeContains, FilterModeNotContains,
				FilterModeStartsWith, FilterModeEndsWith:
				match = compareStringModes(itemValue, filter.Value, filter.Mode)
			case FilterModeIsEmpty:
				match = isEmpty(fieldVal)
			case FilterModeIsNotEmpty:
				match = !isEmpty(fieldVal)
			case FilterModeGT, FilterModeGTE, FilterModeLT, FilterModeLTE, FilterModeRange,
				FilterModeBefore, FilterModeAfter:
				match = compareAdvanced(itemValue, filter.Mode, filter.Value)
			default:
				match = false
			}

			// --- BOOLEAN DATA TYPE HANDLER ---
			if filter.DataType == "boolean" {
				bv := toBool(itemValue)
				fv := toBool(filter.Value)
				switch filter.Mode {
				case FilterModeEqual:
					match = bv == fv
				case FilterModeNotEqual:
					match = bv != fv
				}
			}

			if logic == FilterLogicAnd && !match {
				matches = false
				break
			}
			if logic == FilterLogicOr && match {
				matches = true
				break
			}
		}
		if matches {
			result = append(result, item)
		}
	}
	return result
}
func isWholeDaySupportedMode(mode string) bool {
	switch mode {
	case FilterModeEqual, FilterModeNotEqual,
		FilterModeGT, FilterModeGTE, FilterModeLT, FilterModeLTE,
		FilterModeBefore, FilterModeAfter:
		return true
	}
	return false
}

// --- String Comparison Helper ---

func compareStringModes(itemValue, filterValue any, mode string) bool {
	// Try to compare by type first
	switch v := itemValue.(type) {
	case string:
		fs, _ := filterValue.(string)
		itemStr := v
		filterStr := fs
		switch mode {
		case FilterModeEqual:
			return strings.EqualFold(itemStr, filterStr)
		case FilterModeNotEqual:
			return !strings.EqualFold(itemStr, filterStr)
		case FilterModeContains:
			return strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
		case FilterModeNotContains:
			return !strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
		case FilterModeStartsWith:
			return strings.HasPrefix(strings.ToLower(itemStr), strings.ToLower(filterStr))
		case FilterModeEndsWith:
			return strings.HasSuffix(strings.ToLower(itemStr), strings.ToLower(filterStr))
		}
	case int, int8, int16, int32, int64:
		fi, ok := toInt64(filterValue)
		vi, _ := toInt64(v)
		switch mode {
		case FilterModeEqual:
			return ok && vi == fi
		case FilterModeNotEqual:
			return ok && vi != fi
		}
	case float32, float64:
		ff, ok := toFloat64Safe(filterValue)
		vf, _ := toFloat64Safe(v)
		switch mode {
		case FilterModeEqual:
			return ok && vf == ff
		case FilterModeNotEqual:
			return ok && vf != ff
		}
	case bool:
		// For booleans, handled in FilterSlice directly
	}
	// fallback: string compare
	itemStr := fmt.Sprintf("%v", itemValue)
	filterStr := fmt.Sprintf("%v", filterValue)
	switch mode {
	case FilterModeEqual:
		return strings.EqualFold(itemStr, filterStr)
	case FilterModeNotEqual:
		return !strings.EqualFold(itemStr, filterStr)
	case FilterModeContains:
		return strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
	case FilterModeNotContains:
		return !strings.Contains(strings.ToLower(itemStr), strings.ToLower(filterStr))
	case FilterModeStartsWith:
		return strings.HasPrefix(strings.ToLower(itemStr), strings.ToLower(filterStr))
	case FilterModeEndsWith:
		return strings.HasSuffix(strings.ToLower(itemStr), strings.ToLower(filterStr))
	}
	return false
}

// --- Sorting Function ---

func SortSlice[T any](ctx context.Context, data []*T, sortFields []SortField) []*T {
	sort.SliceStable(data, func(i, j int) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		aVal := reflect.ValueOf(data[i])
		bVal := reflect.ValueOf(data[j])
		if aVal.Kind() == reflect.Ptr {
			aVal = aVal.Elem()
		}
		if bVal.Kind() == reflect.Ptr {
			bVal = bVal.Elem()
		}

		for _, sortField := range sortFields {
			fieldA := aVal.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, sortField.Field)
			})
			fieldB := bVal.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, sortField.Field)
			})
			if !fieldA.IsValid() || !fieldB.IsValid() {
				continue
			}
			order := strings.ToLower(sortField.Order)
			va := fieldA.Interface()
			vb := fieldB.Interface()
			cmp := compareForSortUniversal(va, vb)
			if cmp == 0 {
				continue
			}
			if order == "desc" {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})
	return data
}

// --- Universal Sort Comparison for All Supported Types ---

func compareForSortUniversal(a, b any) int {
	// Bool direct
	if ab, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			if ab == bb {
				return 0
			}
			if !ab && bb {
				return -1
			}
			return 1
		}
	}

	// Try as time.Time
	if at, ok := a.(time.Time); ok {
		if bt, ok := b.(time.Time); ok {
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
			return 0
		}
	}

	// Number compare only if both are numbers
	af, aok := toFloat64Safe(a)
	bf, bok := toFloat64Safe(b)
	if aok && bok {
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}
		return 0
	}

	// String fallback (case-insensitive)
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return strings.Compare(strings.ToLower(as), strings.ToLower(bs))
}

func toInt64(val any) (int64, bool) {
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	case reflect.String:
		var i int64
		n, err := fmt.Sscanf(rv.String(), "%d", &i)
		return i, err == nil && n == 1
	default:
		return 0, false
	}
}

func toFloat64Safe(val any) (float64, bool) {
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	case reflect.String:
		var f float64
		n, err := fmt.Sscanf(rv.String(), "%f", &f)
		return f, err == nil && n == 1
	default:
		return 0, false
	}
}

// --- Advanced Comparison for Filters ---

func compareAdvanced(itemValue any, mode string, filterValue any) bool {
	switch v := itemValue.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		itemFloat, _ := toFloat64Safe(v)
		switch mode {
		case FilterModeGT:
			fv, ok := toFloat64Safe(filterValue)
			return ok && itemFloat > fv
		case FilterModeGTE:
			fv, ok := toFloat64Safe(filterValue)
			return ok && itemFloat >= fv
		case FilterModeLT:
			fv, ok := toFloat64Safe(filterValue)
			return ok && itemFloat < fv
		case FilterModeLTE:
			fv, ok := toFloat64Safe(filterValue)
			return ok && itemFloat <= fv
		case FilterModeRange:
			if arr, ok := filterValue.([]any); ok && len(arr) == 2 {
				min, minok := toFloat64Safe(arr[0])
				max, maxok := toFloat64Safe(arr[1])
				return minok && maxok && itemFloat >= min && itemFloat <= max
			} else if arr, ok := filterValue.([]interface{}); ok && len(arr) == 2 {
				min, minok := toFloat64Safe(arr[0])
				max, maxok := toFloat64Safe(arr[1])
				return minok && maxok && itemFloat >= min && itemFloat <= max
			}
		}
	case string:
		t, err := tryParseTime(v)
		if err == nil {
			switch mode {
			case FilterModeBefore:
				ft, err := tryParseTime(fmt.Sprintf("%v", filterValue))
				return err == nil && t.Before(ft)
			case FilterModeAfter:
				ft, err := tryParseTime(fmt.Sprintf("%v", filterValue))
				return err == nil && t.After(ft)
			case FilterModeRange:
				if arr, ok := filterValue.([]any); ok && len(arr) == 2 {
					start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
					end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
					return err1 == nil && err2 == nil && !t.Before(start) && !t.After(end)
				} else if arr, ok := filterValue.([]interface{}); ok && len(arr) == 2 {
					start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
					end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
					return err1 == nil && err2 == nil && !t.Before(start) && !t.After(end)
				}
			case FilterModeGTE:
				ft, err := tryParseTime(fmt.Sprintf("%v", filterValue))
				return err == nil && (t.Equal(ft) || t.After(ft))
			case FilterModeLTE:
				ft, err := tryParseTime(fmt.Sprintf("%v", filterValue))
				return err == nil && (t.Equal(ft) || t.Before(ft))
			}
		}
	case time.Time:
		ft, err := tryParseTime(fmt.Sprintf("%v", filterValue))
		switch mode {
		case FilterModeBefore:
			return err == nil && v.Before(ft)
		case FilterModeAfter:
			return err == nil && v.After(ft)
		case FilterModeGTE:
			return err == nil && (v.Equal(ft) || v.After(ft))
		case FilterModeLTE:
			return err == nil && (v.Equal(ft) || v.Before(ft))
		case FilterModeRange:
			if arr, ok := filterValue.([]any); ok && len(arr) == 2 {
				start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
				end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
				return err1 == nil && err2 == nil && !v.Before(start) && !v.After(end)
			} else if arr, ok := filterValue.([]interface{}); ok && len(arr) == 2 {
				start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
				end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
				return err1 == nil && err2 == nil && !v.Before(start) && !v.After(end)
			}
		}
	case bool:
		fb := toBool(filterValue)
		switch mode {
		case FilterModeEqual:
			return v == fb
		case FilterModeNotEqual:
			return v != fb
		}
	}
	return false
}

// --- Try Parse Time Helper ---

func tryParseTime(val string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",     // ISO8601
		"2006-01-02T15:04:05.000Z07:00", // ISO8601 with ms
		"2006-01-02",
		"2006-01-02 15:04:05",
		"15:04:05",
	}
	// support unix timestamp as int (seconds)
	if i, err := toInt64(val); err {
		tm := time.Unix(i, 0)
		return tm, nil
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", val)
}

// --- Helper for Empty Check ---

func isEmpty(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Map, reflect.Slice, reflect.Array:
		return val.IsNil() || val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Struct:
		zero := reflect.Zero(val.Type()).Interface()
		return reflect.DeepEqual(val.Interface(), zero)
	}
	return false
}

// --- Helper for Boolean ---

func toBool(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case int, int8, int16, int32, int64:
		f, _ := toFloat64Safe(v)
		return f != 0
	case float32, float64:
		f, _ := toFloat64Safe(v)
		return f != 0
	}
	return false
}

// --- Pagination ---

func PaginateSlice[T any](ctx context.Context, data []*T, pageIndex, pageSize int) []*T {
	if pageSize <= 0 {
		return data
	}
	start := pageIndex * pageSize
	if start >= len(data) {
		return []*T{}
	}
	end := start + pageSize
	if end > len(data) {
		end = len(data)
	}
	return data[start:end]
}

func Pagination[T any](
	context context.Context,
	ctx echo.Context,
	data []*T,
) PaginationResult[T] {
	filterParam := ctx.QueryParam("filter")
	pageIndexParam := ctx.QueryParam("pageIndex")
	pageSizeParam := ctx.QueryParam("pageSize")
	sortParam := ctx.QueryParam("sort")

	// --- FIXED: decode directly to FilterRoot
	var filterRoot FilterRoot
	if filterParam != "" {
		filterDecodedRaw, _ := url.QueryUnescape(filterParam)
		filterBytes, _ := base64.StdEncoding.DecodeString(filterDecodedRaw)
		_ = json.Unmarshal(filterBytes, &filterRoot)
	}

	pageIndex := 0
	if pageIndexParam != "" {
		if val, err := strconv.Atoi(pageIndexParam); err == nil {
			pageIndex = val
		}
	}
	pageSize := 0
	if pageSizeParam != "" {
		if val, err := strconv.Atoi(pageSizeParam); err == nil {
			pageSize = val
		}
	}
	var sortDecoded []SortField
	if sortParam != "" {
		sortDecodedRaw, _ := url.QueryUnescape(sortParam)
		sortBytes, _ := base64.StdEncoding.DecodeString(sortDecodedRaw)
		_ = json.Unmarshal(sortBytes, &sortDecoded)
	}

	originalTotalSize := len(data)
	sorted := SortSlice(context, data, sortDecoded)
	filtered := FilterSlice(context, sorted, filterRoot.Filters, filterRoot.Logic)
	paged := PaginateSlice(context, filtered, pageIndex, pageSize)

	totalPage := 0
	if pageSize > 0 {
		totalPage = (originalTotalSize + pageSize - 1) / pageSize
	}

	return PaginationResult[T]{
		Data:      paged,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		TotalSize: originalTotalSize,
		Sort:      sortDecoded,
		TotalPage: totalPage,
	}
}
