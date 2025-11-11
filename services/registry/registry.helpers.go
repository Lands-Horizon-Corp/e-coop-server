package registry

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// parseFilters decodes and validates filter parameters
func parseFilters(ctx echo.Context) (filter.Root, error) {
	filterParam := ctx.QueryParam("filter")
	if filterParam == "" {
		return filter.Root{Logic: filter.LogicAnd}, nil
	}
	filterDecodedRaw, err := url.QueryUnescape(filterParam)
	if err != nil {
		return filter.Root{}, eris.Wrap(err, "unescaping failed")
	}
	filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
	if err != nil {
		return filter.Root{}, eris.Wrap(err, "base64 decoding failed")
	}
	var filterRoot filter.Root
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return filter.Root{}, eris.Wrap(err, "JSON unmarshalling failed")
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = filter.LogicAnd
	}
	return filterRoot, nil
}

// parseSort decodes and validates sort parameters
func parseSort(ctx echo.Context) ([]filter.SortField, error) {
	sortParam := ctx.QueryParam("sort")
	if sortParam == "" {
		return nil, nil
	}
	sortDecodedRaw, err := url.QueryUnescape(sortParam)
	if err != nil {
		return nil, eris.Wrap(err, "unescaping failed")
	}
	sortBytes, err := base64.StdEncoding.DecodeString(sortDecodedRaw)
	if err != nil {
		return nil, eris.Wrap(err, "base64 decoding failed")
	}
	var sortFields []filter.SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, eris.Wrap(err, "JSON unmarshalling failed")
	}
	for i, sf := range sortFields {
		order := strings.ToUpper(string(sf.Order))
		if order != "ASC" && order != "DESC" {
			sortFields[i].Order = "ASC"
		} else {
			sortFields[i].Order = filter.SortOrder(order)
		}
	}
	return sortFields, nil
}

// parsePageSize extracts and validates page size parameter
func parsePageSize(ctx echo.Context) (int, error) {
	pageSize, err := strconv.Atoi(ctx.QueryParam("pageSize"))
	if err != nil {
		return 0, eris.Wrap(err, "invalid pageSize parameter")
	}
	return pageSize, nil
}

// parsePageIndex extracts and validates page index parameter
func parsePageIndex(ctx echo.Context) (int, error) {
	pageIndex, err := strconv.Atoi(ctx.QueryParam("pageIndex"))
	if err != nil {
		return 0, eris.Wrap(err, "invalid pageIndex parameter")
	}
	return pageIndex, nil
}

// parseQuery extracts filter, sort, page index, and page size from the request context
func parseQuery(ctx echo.Context) (filter.Root, int, int, error) {
	filterRoot, err := parseFilters(ctx)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "filter processing failed")
	}
	sortFields, err := parseSort(ctx)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "sort processing failed")
	}
	filterRoot.SortFields = sortFields
	pageIndex, err := parsePageIndex(ctx)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "pageIndex processing failed")
	}
	pageSize, err := parsePageSize(ctx)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "pageSize processing failed")
	}

	return filterRoot, pageIndex, pageSize, nil
}

func parsesStringFilters(query string) (filter.Root, error) {
	filterBytes, err := base64.StdEncoding.DecodeString(query)
	if err != nil {
		return filter.Root{}, eris.Wrap(err, "base64 decoding failed")
	}
	var filterRoot filter.Root
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return filter.Root{}, eris.Wrap(err, "JSON unmarshalling failed")
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = filter.LogicAnd
	}
	return filterRoot, nil
}

// parseSort decodes and validates sort parameters
func parseStringSort(query string) ([]filter.SortField, error) {
	sortBytes, err := base64.StdEncoding.DecodeString(query)
	if err != nil {
		return nil, eris.Wrap(err, "base64 decoding failed")
	}
	var sortFields []filter.SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, eris.Wrap(err, "JSON unmarshalling failed")
	}
	for i, sf := range sortFields {
		order := strings.ToUpper(string(sf.Order))
		if order != "ASC" && order != "DESC" {
			sortFields[i].Order = "ASC"
		} else {
			sortFields[i].Order = filter.SortOrder(order)
		}
	}
	return sortFields, nil
}

// parsePageIndex extracts and validates page index parameter
func parseStringPageIndex(query string) (int, error) {
	pageIndex, err := strconv.Atoi(query)
	if err != nil {
		return 0, eris.Wrap(err, "invalid pageIndex parameter")
	}
	return pageIndex, nil
}

// parsePageSize extracts and validates page size parameter
func parseStringPageSize(query string) (int, error) {
	pageSize, err := strconv.Atoi(query)
	if err != nil {
		return 0, eris.Wrap(err, "invalid pageSize parameter")
	}
	return pageSize, nil
}

func parseStringQuery(query string) (filter.Root, int, int, error) {
	filterRoot, err := parsesStringFilters(query)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "filter processing failed")
	}
	sortFields, err := parseStringSort(query)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "sort processing failed")
	}
	filterRoot.SortFields = sortFields
	pageIndex, err := parseStringPageIndex(query)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "pageIndex processing failed")
	}
	pageSize, err := parseStringPageSize(query)
	if err != nil {
		return filter.Root{}, 0, 0, eris.Wrap(err, "pageSize processing failed")
	}
	return filterRoot, pageIndex, pageSize, nil
}

func parseUUIDArrayFromQuery(query string) ([]uuid.UUID, bool) {
	if query == "" {
		return nil, false
	}
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "[") && strings.HasSuffix(query, "]") {
		query = strings.Trim(query, "[]")
	}
	values := strings.Split(query, ",")
	if len(values) == 0 {
		return nil, false
	}
	var uuids []uuid.UUID
	for _, value := range values {
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		parsedUUID, err := uuid.Parse(value)
		if err != nil {
			return nil, false
		}

		uuids = append(uuids, parsedUUID)
	}
	return uuids, true
}

// applySQLFilters safely applies FilterSQL conditions to the database query
func (r *Registry[TData, TResponse, TRequest]) applySQLFilters(db *gorm.DB, filters []FilterSQL) *gorm.DB {
	for _, f := range filters {

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
