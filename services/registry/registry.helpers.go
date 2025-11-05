package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// parseFilters decodes and validates filter parameters
func parseFilters(ctx echo.Context) (filter.Root, error) {
	filterParam := ctx.QueryParam("filter")
	fmt.Println("=== FILTER DEBUG ===")
	fmt.Println("Raw filter param:", filterParam)

	if filterParam == "" {
		fmt.Println("Filter param is empty, returning default AND logic")
		return filter.Root{Logic: filter.LogicAnd}, nil
	}

	filterDecodedRaw, err := url.QueryUnescape(filterParam)
	if err != nil {
		fmt.Println("Unescape failed:", err)
		return filter.Root{}, eris.Wrap(err, "unescaping failed")
	}
	fmt.Println("After unescape:", filterDecodedRaw)

	filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
	if err != nil {
		fmt.Println("Base64 decode failed:", err)
		return filter.Root{}, eris.Wrap(err, "base64 decoding failed")
	}
	fmt.Println("After base64 decode:", string(filterBytes))

	var filterRoot filter.Root
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		fmt.Println("JSON unmarshal failed:", err)
		return filter.Root{}, eris.Wrap(err, "JSON unmarshalling failed")
	}

	fmt.Println("Parsed filterRoot:", filterRoot)
	fmt.Println("FieldFilters count:", len(filterRoot.FieldFilters))
	fmt.Println("Logic:", filterRoot.Logic)
	if len(filterRoot.FieldFilters) > 0 {
		for i, f := range filterRoot.FieldFilters {
			fmt.Printf("Filter[%d]: Field=%s, Mode=%s, DataType=%s, Value=%v\n",
				i, f.Field, f.Mode, f.DataType, f.Value)
		}
	}

	if filterRoot.Logic == "" {
		filterRoot.Logic = filter.LogicAnd
	}
	fmt.Println("=== END FILTER DEBUG ===")
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
