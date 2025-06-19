package horizon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	FilterModeIn          = "in"
	FilterModeNotIn       = "notin"
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

// --- Helper Functions ---

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
		i, err := strconv.ParseInt(rv.String(), 10, 64)
		return i, err == nil
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
		f, err := strconv.ParseFloat(rv.String(), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

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

func convertToAnySlice(v any) []any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}
	return result
}

// --- Field Access ---

func getFieldValue(val reflect.Value, fieldPath string) reflect.Value {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}
		}
		val = val.Elem()
	}

	// Make fieldPath case-insensitive by lowering all parts
	parts := strings.Split(fieldPath, ".")
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	current := val
	for _, part := range parts {
		if current.Kind() != reflect.Struct {
			return reflect.Value{}
		}

		found := false
		t := current.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			tagName := strings.Split(tag, ",")[0]
			// Lowercase for case-insensitive comparison
			if (tagName != "" && strings.ToLower(tagName) == part) || strings.ToLower(f.Name) == part {
				current = current.Field(i)
				found = true
				break
			}
		}

		if !found {
			return reflect.Value{}
		}
	}
	return current
}

// --- Time Handling ---

var (
	timeFormats = []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC822,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04",
		"2006-01-02 15",
		"15:04:05",
		"15:04",
	}
)

func parseTimeWithLocation(val string, loc *time.Location) (time.Time, error) {
	// Try as unix timestamp
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return time.Unix(i, 0).In(loc), nil
	}

	for _, layout := range timeFormats {
		if t, err := time.ParseInLocation(layout, val, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time: %s", val)
}

func tryParseTime(val string) (time.Time, error) {
	return parseTimeWithLocation(val, time.UTC)
}

func tryParseTimeWithLocation(val string, loc *time.Location) (time.Time, error) {
	return parseTimeWithLocation(val, loc)
}

// --- Empty Check ---

func isEmpty(val reflect.Value) bool {
	if !val.IsValid() {
		return true
	}

	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return true
		}
		return isEmpty(val.Elem())
	}

	if val.CanInterface() {
		switch v := val.Interface().(type) {
		case time.Time:
			return v.IsZero()
		case *time.Time:
			return v == nil || v.IsZero()
		}
	}

	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Map, reflect.Slice, reflect.Array:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Struct:
		zero := reflect.Zero(val.Type()).Interface()
		return reflect.DeepEqual(val.Interface(), zero)
	}

	return false
}

// --- Comparison Functions ---

func compareValues(a, b any) int {
	// Handle time comparisons first
	if at, aok := a.(time.Time); aok {
		if bt, bok := b.(time.Time); bok {
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
			return 0
		}
	}

	// Numeric comparisons
	af, aok := toFloat64Safe(a)
	bf, bok := toFloat64Safe(b)
	if aok && bok {
		if af < bf {
			return -1
		}
		if af > bf {
			return 1
		}
		return 0
	}

	// Bool comparisons
	if ab, aok := a.(bool); aok {
		if bb, bok := b.(bool); bok {
			if !ab && bb {
				return -1
			}
			if ab && !bb {
				return 1
			}
			return 0
		}
	}

	// String comparisons (case-insensitive)
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	as = strings.ToLower(as)
	bs = strings.ToLower(bs)
	return strings.Compare(as, bs)
}

func compareForEquality(a, b any) bool {
	return compareValues(a, b) == 0
}

func compareForSortUniversal(a, b any) int {
	return compareValues(a, b)
}

// --- Filtering Logic ---

func FilterSlice[T any](ctx context.Context, data []*T, filters []Filter, logic FilterLogic) []*T {
	if len(filters) == 0 {
		return data
	}

	result := make([]*T, 0, len(data))
	for _, item := range data {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		matches := logic == FilterLogicAnd
	FilterLoop:
		for _, filter := range filters {
			fieldVal := getFieldValue(val, filter.Field)
			if !fieldVal.IsValid() {
				if logic == FilterLogicAnd {
					matches = false
					break FilterLoop
				}
				continue
			}

			itemValue := fieldVal.Interface()
			match := false

			// Handle special data types first
			switch filter.DataType {
			case "date", "datetime":
				if isDateSupportedMode(filter.Mode) {
					match = handleDateFilter(itemValue, filter)
					break
				}
			case "boolean":
				if isBoolSupportedMode(filter.Mode) {
					match = handleBoolFilter(itemValue, filter)
					break
				}
			}

			if !match {
				switch filter.Mode {
				case FilterModeIn:
					match = handleInFilter(itemValue, filter.Value)
				case FilterModeNotIn:
					match = !handleInFilter(itemValue, filter.Value)
				case FilterModeEqual, FilterModeNotEqual,
					FilterModeContains, FilterModeNotContains,
					FilterModeStartsWith, FilterModeEndsWith:
					match = handleStringFilter(itemValue, filter.Value, filter.Mode)
				case FilterModeIsEmpty:
					match = isEmpty(fieldVal)
				case FilterModeIsNotEmpty:
					match = !isEmpty(fieldVal)
				case FilterModeGT, FilterModeGTE, FilterModeLT, FilterModeLTE, FilterModeRange,
					FilterModeBefore, FilterModeAfter:
					match = handleAdvancedFilter(itemValue, filter.Mode, filter.Value)
				}
			}

			if logic == FilterLogicAnd && !match {
				matches = false
				break FilterLoop
			}
			if logic == FilterLogicOr && match {
				matches = true
				break FilterLoop
			}
		}

		if matches {
			result = append(result, item)
		}
	}

	return result
}

func isDateSupportedMode(mode string) bool {
	switch mode {
	case FilterModeEqual, FilterModeNotEqual,
		FilterModeGT, FilterModeGTE, FilterModeLT, FilterModeLTE,
		FilterModeBefore, FilterModeAfter, FilterModeRange,
		FilterModeIsEmpty, FilterModeIsNotEmpty:
		return true
	}
	return false
}

func isBoolSupportedMode(mode string) bool {
	switch mode {
	case FilterModeEqual, FilterModeNotEqual,
		FilterModeIsEmpty, FilterModeIsNotEmpty:
		return true
	}
	return false
}

func handleDateFilter(itemValue any, filter Filter) bool {
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
			itemErr = errors.New("nil *time.Time")
		}
	default:
		itemErr = errors.New("not a date type")
	}

	if filterErr != nil || itemErr != nil {
		return false
	}

	isWholeDay := filterTime.Hour() == 0 && filterTime.Minute() == 0 &&
		filterTime.Second() == 0 && filterTime.Nanosecond() == 0
	dayStart := filterTime
	dayEnd := filterTime.Add(24 * time.Hour)

	switch filter.Mode {
	case FilterModeEqual:
		if isWholeDay {
			return !itemTime.Before(dayStart) && itemTime.Before(dayEnd)
		}
		return itemTime.Equal(filterTime)
	case FilterModeNotEqual:
		if isWholeDay {
			return itemTime.Before(dayStart) || !itemTime.Before(dayEnd)
		}
		return !itemTime.Equal(filterTime)
	case FilterModeGT, FilterModeAfter:
		if isWholeDay {
			return itemTime.After(dayEnd.Add(-time.Nanosecond))
		}
		return itemTime.After(filterTime)
	case FilterModeGTE:
		if isWholeDay {
			return !itemTime.Before(dayStart)
		}
		return itemTime.Equal(filterTime) || itemTime.After(filterTime)
	case FilterModeLT, FilterModeBefore:
		if isWholeDay {
			return itemTime.Before(dayStart)
		}
		return itemTime.Before(filterTime)
	case FilterModeLTE:
		if isWholeDay {
			return itemTime.Before(dayEnd)
		}
		return itemTime.Equal(filterTime) || itemTime.Before(filterTime)
	case FilterModeRange:
		if arr, ok := filter.Value.([]any); ok && len(arr) == 2 {
			start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
			end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
			if err1 == nil && err2 == nil {
				return !itemTime.Before(start) && !itemTime.After(end)
			}
		}
	case FilterModeIsEmpty:
		return itemTime.IsZero()
	case FilterModeIsNotEmpty:
		return !itemTime.IsZero()
	}

	return false
}

func handleBoolFilter(itemValue any, filter Filter) bool {
	bv := toBool(itemValue)
	fv := toBool(filter.Value)

	switch filter.Mode {
	case FilterModeEqual:
		return bv == fv
	case FilterModeNotEqual:
		return bv != fv
	case FilterModeIsEmpty:
		return !bv
	case FilterModeIsNotEmpty:
		return bv
	}

	return false
}

func handleInFilter(itemValue any, filterValue any) bool {
	slice := convertToAnySlice(filterValue)
	if slice == nil {
		return false
	}

	for _, v := range slice {
		if compareForEquality(itemValue, v) {
			return true
		}
	}
	return false
}

func handleStringFilter(itemValue, filterValue any, mode string) bool {
	itemStr := fmt.Sprintf("%v", itemValue)
	filterStr := fmt.Sprintf("%v", filterValue)

	// Lowercase both for case-insensitive comparison
	itemStr = strings.ToLower(itemStr)
	filterStr = strings.ToLower(filterStr)

	switch mode {
	case FilterModeEqual:
		return itemStr == filterStr
	case FilterModeNotEqual:
		return itemStr != filterStr
	case FilterModeContains:
		return strings.Contains(itemStr, filterStr)
	case FilterModeNotContains:
		return !strings.Contains(itemStr, filterStr)
	case FilterModeStartsWith:
		return strings.HasPrefix(itemStr, filterStr)
	case FilterModeEndsWith:
		return strings.HasSuffix(itemStr, filterStr)
	}
	return false
}

func handleAdvancedFilter(itemValue any, mode string, filterValue any) bool {
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
			if arr := convertToAnySlice(filterValue); len(arr) == 2 {
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
				if arr := convertToAnySlice(filterValue); len(arr) == 2 {
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
		if err != nil {
			return false
		}
		switch mode {
		case FilterModeBefore:
			return v.Before(ft)
		case FilterModeAfter:
			return v.After(ft)
		case FilterModeGTE:
			return v.Equal(ft) || v.After(ft)
		case FilterModeLTE:
			return v.Equal(ft) || v.Before(ft)
		case FilterModeRange:
			if arr := convertToAnySlice(filterValue); len(arr) == 2 {
				start, err1 := tryParseTime(fmt.Sprintf("%v", arr[0]))
				end, err2 := tryParseTime(fmt.Sprintf("%v", arr[1]))
				return err1 == nil && err2 == nil && !v.Before(start) && !v.After(end)
			}
		}
	}
	return false
}

// --- Sorting Logic ---

func SortSlice[T any](ctx context.Context, data []*T, sortFields []SortField) []*T {
	if len(sortFields) == 0 {
		return data
	}

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
			fieldA := getFieldValue(aVal, sortField.Field)
			fieldB := getFieldValue(bVal, sortField.Field)
			if !fieldA.IsValid() || !fieldB.IsValid() {
				continue
			}

			va := fieldA.Interface()
			vb := fieldB.Interface()
			cmp := compareForSortUniversal(va, vb)
			if cmp == 0 {
				continue
			}

			order := strings.ToLower(sortField.Order)
			if order == "desc" {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})
	return data
}

// --- Pagination Logic ---

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
	ctx context.Context,
	echoCtx echo.Context,
	data []*T,
) PaginationResult[T] {
	filterParam := echoCtx.QueryParam("filter")
	pageIndexParam := echoCtx.QueryParam("pageIndex")
	pageSizeParam := echoCtx.QueryParam("pageSize")
	sortParam := echoCtx.QueryParam("sort")

	// Parse filter
	var filterRoot FilterRoot
	if filterParam != "" {
		filterDecodedRaw, _ := url.QueryUnescape(filterParam)
		filterBytes, _ := base64.StdEncoding.DecodeString(filterDecodedRaw)
		_ = json.Unmarshal(filterBytes, &filterRoot)
	}

	// Parse pagination
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

	// Parse sort
	var sortDecoded []SortField
	if sortParam != "" {
		sortDecodedRaw, _ := url.QueryUnescape(sortParam)
		sortBytes, _ := base64.StdEncoding.DecodeString(sortDecodedRaw)
		_ = json.Unmarshal(sortBytes, &sortDecoded)
	}

	// Apply filtering first (most efficient)
	filtered := FilterSlice(ctx, data, filterRoot.Filters, filterRoot.Logic)
	filteredSize := len(filtered)

	// Then apply sorting
	sorted := SortSlice(ctx, filtered, sortDecoded)

	// Finally paginate
	paged := PaginateSlice(ctx, sorted, pageIndex, pageSize)

	// Calculate total pages
	totalPage := 0
	if pageSize > 0 {
		totalPage = (filteredSize + pageSize - 1) / pageSize
	}

	return PaginationResult[T]{
		Data:      paged,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		TotalSize: filteredSize,
		Sort:      sortDecoded,
		TotalPage: totalPage,
	}
}
