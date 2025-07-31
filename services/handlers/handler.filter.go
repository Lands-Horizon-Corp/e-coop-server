package handlers

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
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// --- Constants and Type Definitions ---

type FilterLogic string

const (
	FilterLogicAnd FilterLogic = "AND"
	FilterLogicOr  FilterLogic = "OR"
)

type DataType string

const (
	DataTypeText    DataType = "text"
	DataTypeDate    DataType = "date"
	DataTypeNumber  DataType = "number"
	DataTypeBoolean DataType = "boolean"
	DataTypeTime    DataType = "time"
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
	DataType DataType `json:"dataType"`
	Field    string   `json:"field"`
	Mode     string   `json:"mode"`
	Value    any      `json:"value"`
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

// --- Error Definitions ---

var (
	ErrInvalidFilterParam   = eris.New("invalid filter parameter")
	ErrInvalidSortParam     = eris.New("invalid sort parameter")
	ErrInvalidField         = eris.New("field not found in struct")
	ErrUnsupportedOperation = eris.New("unsupported operation for data type")
	ErrTypeConversion       = eris.New("type conversion error")
	ErrContextCancelled     = eris.New("operation cancelled by context")
	ErrInvalidValue         = eris.New("invalid value for operation")
)

// --- Global Cache and Pools ---

type cachedFieldInfo struct {
	Index    []int
	DataType reflect.Type
	IsPtr    bool
}

var (
	fieldCache = make(map[string]map[string]cachedFieldInfo)
	cacheMutex = &sync.RWMutex{}
	workerPool = sync.Pool{
		New: func() any { s := make([]int, 0, 1024); return &s },
	}
)

// --- Field Cache Functions ---

func getCachedField(t reflect.Type, fieldName string) (cachedFieldInfo, error) {
	typeName := t.String()

	cacheMutex.RLock()
	typeCache, typeExists := fieldCache[typeName]
	if typeExists {
		info, fieldExists := typeCache[fieldName]
		cacheMutex.RUnlock()
		if fieldExists {
			return info, nil
		}
	} else {
		cacheMutex.RUnlock()
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if typeCache, typeExists = fieldCache[typeName]; typeExists {
		if info, fieldExists := typeCache[fieldName]; fieldExists {
			return info, nil
		}
	} else {
		typeCache = make(map[string]cachedFieldInfo)
		fieldCache[typeName] = typeCache
	}

	info, found := findFieldRecursive(t, fieldName, nil)
	if found {
		typeCache[fieldName] = info
		return info, nil
	}

	return cachedFieldInfo{}, eris.Wrapf(ErrInvalidField, "field '%s' not found", fieldName)
}

func findFieldRecursive(t reflect.Type, fieldName string, basePath []int) (cachedFieldInfo, bool) {
	parts := strings.SplitN(fieldName, ".", 2)
	currentField := parts[0]

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tagValue := strings.Split(f.Tag.Get("json"), ",")[0]

		if tagValue != "" && strings.EqualFold(tagValue, currentField) {
			return handleFoundField(f, parts, basePath, i)
		}

		if strings.EqualFold(f.Name, currentField) {
			return handleFoundField(f, parts, basePath, i)
		}

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if info, found := findFieldRecursive(
				f.Type,
				fieldName,
				append(basePath, i),
			); found {
				return info, true
			}
		}
	}

	return cachedFieldInfo{}, false
}

func handleFoundField(f reflect.StructField, parts []string, basePath []int, fieldIndex int) (cachedFieldInfo, bool) {
	path := append(basePath, fieldIndex)
	fieldType := f.Type
	isPtr := false

	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
		isPtr = true
	}

	if len(parts) > 1 {
		if fieldType.Kind() == reflect.Struct {
			return findFieldRecursive(fieldType, parts[1], path)
		}
		return cachedFieldInfo{}, false
	}

	return cachedFieldInfo{
		Index:    path,
		DataType: fieldType,
		IsPtr:    isPtr,
	}, true
}

func getFieldValue(val reflect.Value, info cachedFieldInfo) reflect.Value {
	fieldVal := val
	for _, idx := range info.Index {
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				return reflect.Value{}
			}
			fieldVal = fieldVal.Elem()
		}
		fieldVal = fieldVal.Field(idx)
	}

	if info.IsPtr && fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return reflect.Value{}
		}
		return fieldVal.Elem()
	}
	return fieldVal
}

// --- Filtering Functions ---

func FilterSlice[T any](
	ctx context.Context,
	data []*T,
	filters []Filter,
	logic FilterLogic,
	batchSize int,
	maxWorkers int,
) ([]*T, error) {
	if len(filters) == 0 || len(data) == 0 {
		return data, nil
	}

	var t T
	itemType := reflect.TypeOf(t)
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	for _, f := range filters {
		if _, err := getCachedField(itemType, f.Field); err != nil {
			return nil, eris.Wrapf(err, "field '%s'", f.Field)
		}
	}

	batchCount := (len(data) + batchSize - 1) / batchSize
	results := make([][]*T, batchCount)
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for batch := range batchCount {
		start := batch * batchSize
		end := min(start+batchSize, len(data))
		if start >= end {
			continue
		}

		numWorkers := min(maxWorkers, (end-start+batchSize-1)/batchSize)
		chunkSize := (end - start + numWorkers - 1) / numWorkers

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			workerStart := start + w*chunkSize
			workerEnd := min(workerStart+chunkSize, end)

			go func(batchIdx, startIdx, endIdx int) {
				defer wg.Done()
				indicesPtr := workerPool.Get().(*[]int)
				indices := (*indicesPtr)[:0]
				defer func() {
					workerPool.Put(indicesPtr)
				}()

				for idx := startIdx; idx < endIdx; idx++ {
					if ctx.Err() != nil {
						return
					}

					val := reflect.ValueOf(data[idx]).Elem()
					match, err := checkItemMatch(val, filters, logic, itemType)
					if err != nil {
						select {
						case errChan <- eris.Wrapf(err, "item at index %d", idx):
							cancel()
						default:
						}
						return
					}

					if match {
						indices = append(indices, idx)
					}
				}

				batchResults := make([]*T, len(indices))
				for i, idx := range indices {
					batchResults[i] = data[idx]
				}
				results[batchIdx] = batchResults
			}(batch, workerStart, workerEnd)
		}
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, eris.Wrap(err, "filtering failed")
	}

	var finalResult []*T
	for _, batch := range results {
		finalResult = append(finalResult, batch...)
	}
	return finalResult, nil
}

func checkItemMatch(
	val reflect.Value,
	filters []Filter,
	logic FilterLogic,
	itemType reflect.Type,
) (bool, error) {
	switch logic {
	case FilterLogicAnd:
		for _, filter := range filters {
			match, err := evaluateFilter(val, filter, itemType)
			if err != nil {
				return false, eris.Wrapf(err, "filter on field '%s'", filter.Field)
			}
			if !match {
				return false, nil
			}
		}
		return true, nil

	case FilterLogicOr:
		for _, filter := range filters {
			match, err := evaluateFilter(val, filter, itemType)
			if err != nil {
				return false, eris.Wrapf(err, "filter on field '%s'", filter.Field)
			}
			if match {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, eris.Wrapf(ErrUnsupportedOperation, "logic '%s'", logic)
	}
}

func evaluateFilter(val reflect.Value, filter Filter, itemType reflect.Type) (bool, error) {
	info, err := getCachedField(itemType, filter.Field)
	if eris.Is(err, ErrInvalidField) && filter.Mode == FilterModeIsEmpty {
		return true, nil
	} else if err != nil {
		return false, err
	}

	fieldVal := getFieldValue(val, info)
	if !fieldVal.IsValid() {
		return filter.Mode == FilterModeIsEmpty, nil
	}

	switch filter.DataType {
	case DataTypeText:
		return compareString(fieldVal, filter.Value, filter.Mode)
	case DataTypeNumber:
		return compareNumber(fieldVal, filter.Mode, filter.Value)
	case DataTypeBoolean:
		return compareBool(fieldVal, filter.Value, filter.Mode)
	case DataTypeDate, DataTypeTime:
		return compareTime(fieldVal, filter.Value, filter.Mode, filter.DataType == DataTypeDate)
	default:
		return compareGeneric(fieldVal, filter.Mode, filter.Value)
	}
}

// --- Comparison Functions ---

func compareString(val reflect.Value, filterValue any, mode string) (bool, error) {
	itemStr := strings.ToLower(fmt.Sprintf("%v", val.Interface()))
	filterStr := strings.ToLower(fmt.Sprintf("%v", filterValue))

	switch mode {
	case FilterModeEqual:
		return itemStr == filterStr, nil
	case FilterModeNotEqual:
		return itemStr != filterStr, nil
	case FilterModeContains:
		return strings.Contains(itemStr, filterStr), nil
	case FilterModeNotContains:
		return !strings.Contains(itemStr, filterStr), nil
	case FilterModeStartsWith:
		return strings.HasPrefix(itemStr, filterStr), nil
	case FilterModeEndsWith:
		return strings.HasSuffix(itemStr, filterStr), nil
	case FilterModeIsEmpty:
		return itemStr == "", nil
	case FilterModeIsNotEmpty:
		return itemStr != "", nil
	default:
		return false, eris.Wrapf(ErrUnsupportedOperation, "mode '%s' for text", mode)
	}
}

func compareBool(val reflect.Value, filterValue any, mode string) (bool, error) {
	itemBool := toBool(val.Interface())
	filterBool := toBool(filterValue)

	switch mode {
	case FilterModeEqual:
		return itemBool == filterBool, nil
	case FilterModeNotEqual:
		return itemBool != filterBool, nil
	default:
		return false, eris.Wrapf(ErrUnsupportedOperation, "mode '%s' for boolean", mode)
	}
}

func compareTime(val reflect.Value, filterValue any, mode string, dateOnly bool) (bool, error) {
	itemTime, err := toTime(val.Interface())
	if err != nil {
		return false, eris.Wrap(err, "item time conversion")
	}

	filterTime, err := toTime(filterValue)
	if err != nil {
		if mode == FilterModeIsEmpty {
			return itemTime.IsZero(), nil
		}
		return false, eris.Wrap(err, "filter time conversion")
	}

	if dateOnly {
		itemTime = truncateToDate(itemTime)
		filterTime = truncateToDate(filterTime)
	}

	switch mode {
	case FilterModeEqual:
		return itemTime.Equal(filterTime), nil
	case FilterModeNotEqual:
		return !itemTime.Equal(filterTime), nil
	case FilterModeAfter, FilterModeGT:
		return itemTime.After(filterTime), nil
	case FilterModeBefore, FilterModeLT:
		return itemTime.Before(filterTime), nil
	case FilterModeGTE:
		return itemTime.After(filterTime) || itemTime.Equal(filterTime), nil
	case FilterModeLTE:
		return itemTime.Before(filterTime) || itemTime.Equal(filterTime), nil
	case FilterModeRange:
		times, ok := filterValue.([]any)
		if !ok || len(times) != 2 {
			return false, eris.Wrap(ErrInvalidValue, "range requires two values")
		}

		start, err1 := toTime(times[0])
		end, err2 := toTime(times[1])
		if err1 != nil || err2 != nil {
			return false, eris.Wrap(ErrTypeConversion, "range values must be times")
		}

		return (itemTime.Equal(start) || itemTime.After(start)) &&
			(itemTime.Equal(end) || itemTime.Before(end)), nil
	default:
		return false, eris.Wrapf(ErrUnsupportedOperation, "mode '%s' for time", mode)
	}
}

func compareNumber(val reflect.Value, mode string, filterValue any) (bool, error) {
	itemNum, err := toFloat64(val.Interface())
	if err != nil {
		return false, eris.Wrap(err, "item number conversion")
	}

	switch mode {
	case FilterModeEqual:
		filterNum, err := toFloat64(filterValue)
		return itemNum == filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeNotEqual:
		filterNum, err := toFloat64(filterValue)
		return itemNum != filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeGT:
		filterNum, err := toFloat64(filterValue)
		return itemNum > filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeGTE:
		filterNum, err := toFloat64(filterValue)
		return itemNum >= filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeLT:
		filterNum, err := toFloat64(filterValue)
		return itemNum < filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeLTE:
		filterNum, err := toFloat64(filterValue)
		return itemNum <= filterNum, eris.Wrap(err, "filter number conversion")
	case FilterModeRange:
		nums, ok := filterValue.([]any)
		if !ok || len(nums) != 2 {
			return false, eris.Wrap(ErrInvalidValue, "range requires two values")
		}

		min, err1 := toFloat64(nums[0])
		max, err2 := toFloat64(nums[1])
		if err1 != nil || err2 != nil {
			return false, eris.Wrap(ErrTypeConversion, "range values must be numbers")
		}

		return itemNum >= min && itemNum <= max, nil
	default:
		return false, eris.Wrapf(ErrUnsupportedOperation, "mode '%s' for number", mode)
	}
}

func compareGeneric(val reflect.Value, mode string, filterValue any) (bool, error) {
	switch mode {
	case FilterModeIsEmpty:
		return isEmpty(val), nil
	case FilterModeIsNotEmpty:
		return !isEmpty(val), nil
	default:
		return compareNumber(val, mode, filterValue)
	}
}

// --- Sorting Functions ---

func SortSlice[T any](
	ctx context.Context,
	data []*T,
	sortFields []SortField,
	batchSize int,
	maxWorkers int,
) ([]*T, error) {
	if len(sortFields) == 0 || len(data) < 2 {
		return data, nil
	}

	var t T
	itemType := reflect.TypeOf(t)
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	type sortableItem struct {
		original *T
		keys     []any
	}

	batchCount := (len(data) + batchSize - 1) / batchSize
	sortableData := make([][]sortableItem, batchCount)
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for batch := range batchCount {
		start := batch * batchSize
		end := min(start+batchSize, len(data))
		if start >= end {
			continue
		}

		wg.Add(1)
		go func(batchIdx, startIdx, endIdx int) {
			defer wg.Done()
			batchData := make([]sortableItem, endIdx-startIdx)

			for i := startIdx; i < endIdx; i++ {
				if ctx.Err() != nil {
					return
				}

				val := reflect.ValueOf(data[i]).Elem()
				keys := make([]any, len(sortFields))

				for j, sf := range sortFields {
					info, err := getCachedField(itemType, sf.Field)
					if err != nil {
						select {
						case errChan <- eris.Wrapf(err, "sort field '%s'", sf.Field):
							cancel()
						default:
						}
						return
					}

					fieldVal := getFieldValue(val, info)
					if fieldVal.IsValid() {
						keys[j] = fieldVal.Interface()
					} else {
						keys[j] = nil
					}
				}

				batchData[i-startIdx] = sortableItem{
					original: data[i],
					keys:     keys,
				}
			}

			sortableData[batchIdx] = batchData
		}(batch, start, end)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, eris.Wrap(err, "preparing sort data failed")
	}

	var finalSortable []sortableItem
	for _, batch := range sortableData {
		finalSortable = append(finalSortable, batch...)
	}

	sort.SliceStable(finalSortable, func(i, j int) bool {
		for k, sf := range sortFields {
			cmp := compareValues(finalSortable[i].keys[k], finalSortable[j].keys[k])
			if cmp != 0 {
				if strings.ToUpper(sf.Order) == "DESC" {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})

	for i := range finalSortable {
		data[i] = finalSortable[i].original
	}

	return data, nil
}

// --- Utility Functions ---

func compareValues(a, b any) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	switch aVal := a.(type) {
	case string:
		bVal, ok := b.(string)
		if !ok {
			return -1
		}
		return strings.Compare(strings.ToLower(aVal), strings.ToLower(bVal))
	case float64:
		bVal, ok := b.(float64)
		if !ok {
			return -1
		}
		if aVal < bVal {
			return -1
		}
		if aVal > bVal {
			return 1
		}
		return 0
	case bool:
		bVal, ok := b.(bool)
		if !ok {
			return -1
		}
		if aVal == bVal {
			return 0
		}
		if !aVal && bVal {
			return -1
		}
		return 1
	case time.Time:
		bVal, ok := b.(time.Time)
		if !ok {
			return -1
		}
		if aVal.Before(bVal) {
			return -1
		}
		if aVal.After(bVal) {
			return 1
		}
		return 0
	default:
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		return strings.Compare(aStr, bStr)
	}
}

func toFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, eris.Wrapf(err, "value '%s'", v)
	default:
		return 0, eris.Wrapf(ErrTypeConversion, "value '%v' type %T", val, val)
	}
}

func toTime(val any) (time.Time, error) {
	switch v := val.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v != nil {
			return *v, nil
		}
		return time.Time{}, eris.Wrap(ErrInvalidValue, "nil time pointer")
	case string:
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05",
			"2006-01-02",
			"15:04:05",
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
			time.RFC822,
			time.RFC822Z,
			time.RFC850,
			time.RFC1123,
			time.RFC1123Z,
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, eris.Wrapf(ErrTypeConversion, "value '%s'", v)
	default:
		return time.Time{}, eris.Wrapf(ErrTypeConversion, "value '%v' type %T", val, val)
	}
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func toBool(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	default:
		return false
	}
}

func isEmpty(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return val.IsZero()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PaginateSlice returns a paginated subset of the data
func PaginateSlice[T any](data []*T, pageIndex, pageSize int) []*T {
	if pageSize <= 0 || len(data) == 0 {
		return []*T{}
	}

	start := pageIndex * pageSize
	if start >= len(data) {
		return []*T{}
	}

	end := min(start+pageSize, len(data))
	return data[start:end]
}

// parseFilters decodes and validates filter parameters
func parseFilters(ctx echo.Context) (FilterRoot, error) {
	filterParam := ctx.QueryParam("filter")
	if filterParam == "" {
		return FilterRoot{Logic: FilterLogicAnd}, nil
	}

	filterDecodedRaw, err := url.QueryUnescape(filterParam)
	if err != nil {
		return FilterRoot{}, eris.Wrap(ErrInvalidFilterParam, "unescaping failed")
	}

	filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
	if err != nil {
		return FilterRoot{}, eris.Wrap(ErrInvalidFilterParam, "base64 decoding failed")
	}

	var filterRoot FilterRoot
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return FilterRoot{}, eris.Wrap(ErrInvalidFilterParam, "JSON unmarshalling failed")
	}

	if filterRoot.Logic == "" {
		filterRoot.Logic = FilterLogicAnd
	}

	return filterRoot, nil
}

// parseSort decodes and validates sort parameters
func parseSort(ctx echo.Context) ([]SortField, error) {
	sortParam := ctx.QueryParam("sort")
	if sortParam == "" {
		return nil, nil
	}

	sortDecodedRaw, err := url.QueryUnescape(sortParam)
	if err != nil {
		return nil, eris.Wrap(ErrInvalidSortParam, "unescaping failed")
	}

	sortBytes, err := base64.StdEncoding.DecodeString(sortDecodedRaw)
	if err != nil {
		return nil, eris.Wrap(ErrInvalidSortParam, "base64 decoding failed")
	}

	var sortFields []SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, eris.Wrap(ErrInvalidSortParam, "JSON unmarshalling failed")
	}

	// Normalize sort orders
	for i, sf := range sortFields {
		order := strings.ToUpper(sf.Order)
		if order != "ASC" && order != "DESC" {
			sortFields[i].Order = "ASC"
		} else {
			sortFields[i].Order = order
		}
	}

	return sortFields, nil
}

// Pagination handles filtering, sorting, and pagination
func Pagination[T any](
	ctx context.Context,
	echoCtx echo.Context,
	data []*T,
	batchSize int,
	maxWorkers int,
) (PaginationResult[T], error) {
	result := PaginationResult[T]{
		PageIndex: 0,
		PageSize:  30,
	}

	// Parse pagination parameters
	if pageIndex, err := strconv.Atoi(echoCtx.QueryParam("pageIndex")); err == nil {
		result.PageIndex = pageIndex
	}
	if pageSize, err := strconv.Atoi(echoCtx.QueryParam("pageSize")); err == nil && pageSize > 0 {
		result.PageSize = pageSize
	}

	// Process filters
	filterRoot, err := parseFilters(echoCtx)
	if err != nil {
		return result, eris.Wrap(err, "filter processing failed")
	}

	// Process sorting
	sortFields, err := parseSort(echoCtx)
	if err != nil {
		return result, eris.Wrap(err, "sort processing failed")
	}
	result.Sort = sortFields

	// Apply filtering
	filtered, err := FilterSlice(ctx, data, filterRoot.Filters, filterRoot.Logic, batchSize, maxWorkers)
	if err != nil {
		return result, eris.Wrap(err, "filtering failed")
	}
	result.TotalSize = len(filtered)

	// Apply sorting
	sorted, err := SortSlice(ctx, filtered, sortFields, batchSize, maxWorkers)
	if err != nil {
		return result, eris.Wrap(err, "sorting failed")
	}

	// Apply pagination
	result.Data = PaginateSlice(sorted, result.PageIndex, result.PageSize)

	// Calculate total pages
	if result.PageSize > 0 {
		result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize
	} else {
		result.TotalPage = 1
	}

	return result, nil
}

// FilterAndSortSlice filters and sorts data based on query parameters without pagination
func FilterAndSortSlice[T any](
	ctx context.Context,
	echoCtx echo.Context,
	data []*T,
	batchSize int,
	maxWorkers int,
) ([]*T, error) {
	filterRoot, err := parseFilters(echoCtx)
	if err != nil {
		return data, eris.Wrap(err, "filter processing failed")
	}
	filtered, err := FilterSlice(ctx, data, filterRoot.Filters, filterRoot.Logic, batchSize, maxWorkers)
	if err != nil {
		return data, eris.Wrap(err, "filtering failed")
	}
	sortFields, err := parseSort(echoCtx)
	if err != nil {
		return data, eris.Wrap(err, "sort processing failed")
	}
	sorted, err := SortSlice(ctx, filtered, sortFields, batchSize, maxWorkers)
	if err != nil {
		return data, eris.Wrap(err, "sorting failed")
	}

	// Apply limit based on pageSize query parameter
	pageSizeStr := echoCtx.QueryParam("pageSize")
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if len(sorted) > pageSize {
				sorted = sorted[:pageSize]
			}
		}
	}

	return sorted, nil
}
