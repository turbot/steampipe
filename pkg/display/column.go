package display

import (
	"encoding/json"
	"fmt"
	"time"

	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

// ColumnValuesAsString converts a slice of columns into strings
func ColumnValuesAsString(values []interface{}, columnTypes []string) ([]string, error) {
	rowAsString := make([]string, len(columnTypes))
	for idx, val := range values {
		val, err := ColumnValueAsString(val, columnTypes[idx])
		if err != nil {
			return nil, err
		}
		rowAsString[idx] = val
	}
	return rowAsString, nil
}

// ColumnValueAsString converts column value to string
func ColumnValueAsString(val interface{}, colType string) (result string, err error) {
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
	// TODO KAI
	switch colType { //.DatabaseTypeName() {
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
func ParseJSONOutputColumnValue(val interface{}, colType string) (interface{}, error) {
	if val == nil {
		return nil, nil
	}
	// TODO KAI
	switch colType { //.DatabaseTypeName() {
	// we can revise/increment the list of DT's in future
	case "JSON", "JSONB", "BOOL", "INT2", "INT4", "INT8", "FLOAT8", "FLOAT4":
		return val, nil
	default:
		return ColumnValueAsString(val, colType)
	}
}
