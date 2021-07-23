package local_db

import (
	"database/sql"
	"encoding/json"
)

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
