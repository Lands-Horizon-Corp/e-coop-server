package ui

import (
	"fmt"
	"log"
	"reflect"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Row struct {
	Label string
	Value string
}

func RenderRow(t Theme, r Row) string {
	return fmt.Sprintf(
		"%s %s",
		t.Label.Render(r.Label+":"),
		t.Value.Render(r.Value),
	)
}

func RenderTitle(t Theme, title string) string {
	return t.Title.Render(title)
}

func RenderBox(t Theme, content ...string) string {
	return t.Box.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

func renderStruct(t Theme, v reflect.Value, typ reflect.Type) string {
	rows := []string{
		RenderTitle(t, "📦 "+typ.Name()),
	}

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)

		// skip unexported
		if field.PkgPath != "" {
			continue
		}

		value := fmt.Sprint(v.Field(i).Interface())

		rows = append(rows, RenderRow(t, Row{
			Label: field.Name,
			Value: value,
		}))
	}

	return RenderBox(t, rows...)
}

func renderMap(t Theme, v reflect.Value) string {
	rows := []string{
		RenderTitle(t, "🗺️ Map"),
	}

	for _, key := range v.MapKeys() {
		rows = append(rows, RenderRow(t, Row{
			Label: fmt.Sprint(key.Interface()),
			Value: fmt.Sprint(v.MapIndex(key).Interface()),
		}))
	}

	return RenderBox(t, rows...)
}

func renderSlice(t Theme, v reflect.Value) string {
	out := ""

	for i := 0; i < v.Len(); i++ {
		out += RenderAny(t, v.Index(i).Interface()) + "\n"
	}

	return out
}

func RenderAny(t Theme, v any) string {
	if v == nil {
		return RenderBox(t, RenderRow(t, Row{"Value", "nil"}))
	}
	if s, ok := v.(fmt.Stringer); ok {
		return RenderBox(t, RenderRow(t, Row{"Value", s.String()}))
	}

	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return RenderBox(t, RenderRow(t, Row{"Value", "nil"}))
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	switch val.Kind() {

	case reflect.Struct:
		return renderStruct(t, val, typ)

	case reflect.Map:
		return renderMap(t, val)

	case reflect.Slice, reflect.Array:
		return renderSlice(t, val)

	default:
		return RenderBox(t, RenderRow(t, Row{"Value", fmt.Sprint(v)}))
	}
}

func PrintEndpoints(title string, endpoints map[string]string) {
	theme := DefaultTheme()

	headerStyle := theme.Title.
		Bold(true).
		Align(lipgloss.Center) // header styling

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1) // normal cell style

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")) // color for URL column

	t := table.New().
		Headers("Service", "URL").
		Border(lipgloss.RoundedBorder()).
		BorderStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == -1: // ✅ HEADER ROW
				return headerStyle
			case col == 1: // URL column
				return urlStyle
			default:
				return cellStyle
			}
		})

	for name, url := range endpoints {
		t.Row(name, url)
	}

	log.Println(
		"\n\n",
		theme.Box.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				theme.Title.Render(title),
				t.Render(),
			),
		),
	)
}
