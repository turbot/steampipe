package queryresult

import "reflect"

// ColumnDef is a struct used to store column information from query results
type ColumnDef struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	ScanType reflect.Type
}
