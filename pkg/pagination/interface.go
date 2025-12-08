package pagination

import "time"

// Mode defines the type of comparison operation to perform
type Mode string

// mode constants define available comparison operations
const (
	ModeEqual       Mode = "equal"       // Exact match
	ModeNotEqual    Mode = "notEqual"    // Not equal
	ModeContains    Mode = "contains"    // Contains substring
	ModeNotContains Mode = "notContains" // Does not contain substring
	ModeStartsWith  Mode = "startsWith"  // Starts with prefix
	ModeEndsWith    Mode = "endsWith"    // Ends with suffix
	ModeIsEmpty     Mode = "isEmpty"     // Is empty or null
	ModeIsNotEmpty  Mode = "isNotEmpty"  // Is not empty
	ModeGT          Mode = "gt"          // Greater than
	ModeGTE         Mode = "gte"         // Greater than or equal
	ModeLT          Mode = "lt"          // Less than
	ModeLTE         Mode = "lte"         // Less than or equal
	ModeRange       Mode = "range"       // Between two values
	ModeBefore      Mode = "before"      // Before (date/time)
	ModeAfter       Mode = "after"       // After (date/time)
)

// DataType defines the data type being filtered
type DataType string

// data type constants define the type of data being filtered
const (
	DataTypeNumber DataType = "number" // Numeric values
	DataTypeText   DataType = "text"   // Text/string values
	DataTypeBool   DataType = "bool"   // Boolean values
	DataTypeDate   DataType = "date"   // Date values
	DataTypeTime   DataType = "time"   // Time values
)

type Logic string

const (
	LogicAnd Logic = "and"
	LogicOr  Logic = "or"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type FieldFilter struct {
	Field    string   `json:"field"`
	Value    any      `json:"value"`
	Mode     Mode     `json:"mode"`
	DataType DataType `json:"dataType"`
}

type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

type Root struct {
	FieldFilters []FieldFilter `json:"filters"`
	SortFields   []SortField   `json:"sortFields"`
	Logic        Logic         `json:"logic"`
	Preload      []string      `json:"preload"`
}

type Range struct {
	From any `json:"from"`
	To   any `json:"to"`
}

type PaginationResult[T any] struct {
	Data      []*T `json:"data"`
	TotalSize int  `json:"totalSize"`
	TotalPage int  `json:"totalPage"`
	PageIndex int  `json:"pageIndex"`
	PageSize  int  `json:"pageSize"`
}

type RangeNumber struct {
	From float64
	To   float64
}

type RangeDate struct {
	From time.Time
	To   time.Time
}
