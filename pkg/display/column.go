package display

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmetb/go-linq"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

// ColumnNames :: extract names from columns
func ColumnNames(columns []*sql.ColumnType) []string {
	var colNames []string
	linq.From(columns).SelectT(func(c *sql.ColumnType) string { return c.Name() }).ToSlice(&colNames)
	return colNames
}

// ColumnValuesAsString converts a slice of columns into strings
func ColumnValuesAsString(values []interface{}, columns []*sql.ColumnType) ([]string, error) {
	rowAsString := make([]string, len(columns))
	for idx, val := range values {
		val, err := ColumnValueAsString(val, columns[idx])
		if err != nil {
			return nil, err
		}
		rowAsString[idx] = val
	}
	return rowAsString, nil
}

// ColumnValueAsString converts column value to string
func ColumnValueAsString(val interface{}, colType *sql.ColumnType) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("%v", val)
		}
	}()

	if val == nil {
		return constants.NullString, nil
	}

	//log.Printf("[TRACE] ColumnValueAsString type %s", colType.DatabaseTypeName())
	// possible types for colType are defined in pq/oid/types.go
	switch colType.DatabaseTypeName() {
	case "JSON", "JSONB":
		bytes, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	case "TIMESTAMP", "DATE", "TIME", "INTERVAL":
		t, ok := val.(time.Time)
		if ok {
			return t.Format("2006-01-02 15:04:05"), nil
		}
		fallthrough
	case "NAME":
		result := string(val.([]uint8))
		return result, nil

	default:
		return typeHelpers.ToString(val), nil
	}

}

// segregate data types, ignore string conversion for certain data types :
// JSON, JSONB, BOOL and so on..
func ParseJSONOutputColumnValue(val interface{}, colType *sql.ColumnType) (interface{}, error) {
	if val == nil {
		return nil, nil
	}
	switch colType.DatabaseTypeName() {
	// we can revise/increment the list of DT's in future
	case "JSON", "JSONB", "BOOL", "INT2", "INT4", "INT8", "FLOAT8", "FLOAT4":
		return val, nil
	default:
		return ColumnValueAsString(val, colType)
	}
}
