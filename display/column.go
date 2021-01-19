package display

import (
	"database/sql"
	"encoding/json"
	"fmt"

	typeHelpers "github.com/turbot/go-kit/types"

	"log"
	"time"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"

	"github.com/ahmetb/go-linq"
)

// ColumnNames :: extract names from columns
func ColumnNames(columns []*sql.ColumnType) []string {
	var colNames []string
	linq.From(columns).SelectT(func(c *sql.ColumnType) string { return c.Name() }).ToSlice(&colNames)
	return colNames
}

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

// ColumnValueAsString :: convert column value to string
func ColumnValueAsString(val interface{}, colType *sql.ColumnType) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("%v", val)
		}
	}()

	if val == nil {
		return cmdconfig.Viper().GetString(constants.ArgNullString), nil
	}

	log.Printf("[TRACE] ColumnValueAsString type %s", colType.DatabaseTypeName())
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
