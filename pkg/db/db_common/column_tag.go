package db_common

import (
	"reflect"
	"strings"
)

// ColumnTag is a struct used to display column info in introspection tables
type ColumnTag struct {
	Column string
	// the introspected go type
	ColumnType string
}

func newColumnTag(field reflect.StructField) (*ColumnTag, bool) {
	columnTag, ok := field.Tag.Lookup(TagColumn)
	if !ok {
		return nil, false
	}
	split := strings.Split(columnTag, ",")
	if len(split) != 2 {
		return nil, false
	}
	column := split[0]
	columnType := split[1]
	return &ColumnTag{column, columnType}, true
}
