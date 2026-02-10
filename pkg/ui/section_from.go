package ui

import (
	"fmt"
	"reflect"
)

func SectionFrom(title string, v any) Section {
	rows := []Row{}
	if v == nil {
		return Section{
			Title: title,
			Rows:  []Row{{Label: "Value", Value: "nil"}},
		}
	}
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return Section{
				Title: title,
				Rows:  []Row{{Label: "Value", Value: "nil"}},
			}
		}
		val = val.Elem()
		typ = typ.Elem()
	}
	if val.Kind() != reflect.Struct {
		return Section{
			Title: title,
			Rows:  []Row{{Label: "Value", Value: fmt.Sprint(v)}},
		}
	}
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		rows = append(rows, Row{
			Label: field.Name,
			Value: fmt.Sprint(val.Field(i).Interface()),
		})
	}
	return Section{
		Title: title,
		Rows:  rows,
	}
}
