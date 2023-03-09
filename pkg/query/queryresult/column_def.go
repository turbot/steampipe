package queryresult

import "reflect"

// ColumnDef is a struct used to store column information from query results
type ColumnDef struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	isScalar *bool
}

// IsScalar checks if the given value is a scalar value
// it also mutates the containing ColumnDef so that it doesn't have to reflect
// for all values in a column
func (c *ColumnDef) IsScalar(v any) bool {
	if c.isScalar == nil {
		var scalar bool
		switch reflect.ValueOf(v).Kind() {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			scalar = false
		default:
			scalar = true
		}
		c.isScalar = &scalar
	}
	return *c.isScalar
}
