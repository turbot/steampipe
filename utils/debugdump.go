package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	typeHelpers "github.com/turbot/go-kit/types"

	"github.com/spf13/viper"
)

// functions specifically used for Debugging purposes.
func DebugDumpJSON(msg string, d interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	os.Stdout.WriteString(msg)
	enc.Encode(d)
}

func DebugDumpViper() {
	fmt.Println(strings.Repeat("*", 80))
	for _, vKey := range viper.AllKeys() {
		fmt.Printf("%-30s => %v\n", vKey, viper.Get(vKey))
	}
	fmt.Println(strings.Repeat("*", 80))
}

func DebugDumpRows(rows *sql.Rows) {
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		// we do not need to stream because
		// defer takes care of it!
		return
	}
	cols, err := rows.Columns()
	if err != nil {
		// we do not need to stream because
		// defer takes care of it!
		return
	}
	fmt.Println(cols)
	fmt.Println("---------------------------------------")
	for rows.Next() {
		row, _ := readRow(rows, cols, colTypes)
		rowAsString, _ := columnValuesAsString(row, colTypes)
		fmt.Println(rowAsString)
	}
}

func readRow(rows *sql.Rows, cols []string, colTypes []*sql.ColumnType) ([]interface{}, error) {
	// slice of interfaces to receive the row data
	columnValues := make([]interface{}, len(cols))
	// make a slice of pointers to the result to pass to scan
	resultPtrs := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i := range columnValues {
		resultPtrs[i] = &columnValues[i]
	}
	rows.Scan(resultPtrs...)

	return populateRow(columnValues, colTypes), nil
}

func populateRow(columnValues []interface{}, colTypes []*sql.ColumnType) []interface{} {
	result := make([]interface{}, len(columnValues))
	for i, columnValue := range columnValues {
		if columnValue != nil {
			colType := colTypes[i]
			dbType := colType.DatabaseTypeName()
			switch dbType {
			case "JSON", "JSONB":
				var val interface{}
				if err := json.Unmarshal(columnValue.([]byte), &val); err != nil {
					// what???
					// TODO how to handle error
				}
				result[i] = val
			default:
				result[i] = columnValue
			}
		}
	}
	return result
}

// columnValuesAsString converts a slice of columns into strings
func columnValuesAsString(values []interface{}, columns []*sql.ColumnType) ([]string, error) {
	rowAsString := make([]string, len(columns))
	for idx, val := range values {
		val, err := columnValueAsString(val, columns[idx])
		if err != nil {
			return nil, err
		}
		rowAsString[idx] = val
	}
	return rowAsString, nil
}

// columnValueAsString converts column value to string
func columnValueAsString(val interface{}, colType *sql.ColumnType) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("%v", val)
		}
	}()

	if val == nil {
		return "<null>", nil
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
