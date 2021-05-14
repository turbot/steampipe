package execute

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

const (
	ControlOk    = "ok"
	ControlAlarm = "alarm"
	ControlSkip  = "skip"
	ControlInfo  = "info"
	ControlError = "error"
)

// ResultRow is the result of a control execution for a single resource
type ResultRow struct {
	Reason     string             `json:"reason"`
	Resource   string             `json:"resource"`
	Status     string             `json:"status"`
	Dimensions map[string]string  `json:"dimensions"`
	Control    *modconfig.Control `json:"-"`
}

// AddDimension checks whether a column value is a scalar type, and if so adds it to the Dimensions map
func (r ResultRow) AddDimension(c *sql.ColumnType, val interface{}) {
	switch c.ScanType().Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return
	default:
		r.Dimensions[c.Name()] = typehelpers.ToString(val)
	}
}

func NewResultRow(control *modconfig.Control, row *queryresult.RowResult, colTypes []*sql.ColumnType) (*ResultRow, error) {
	res := &ResultRow{
		Control:    control,
		Dimensions: make(map[string]string),
	}

	// was there a SQL error _executing the control
	// Note: this is different from the contrrol state being 'error'
	if row.Error != nil {
		return nil, row.Error
	}

	for i, c := range colTypes {
		switch c.Name() {
		case "reason":
			res.Reason = typehelpers.ToString(row.Data[i])
		case "resource":
			res.Resource = typehelpers.ToString(row.Data[i])
		case "status":
			status := typehelpers.ToString(row.Data[i])
			if !IsValidControlStatus(status) {
				return nil, fmt.Errorf("invalid control status '%s'", status)
			}
			res.Status = status
		default:
			// if this is a scalar type, add to dimensions
			res.AddDimension(c, row.Data[i])
		}
	}
	return res, nil
}

func IsValidControlStatus(status string) bool {
	return helpers.StringSliceContains([]string{ControlOk, ControlAlarm, ControlInfo, ControlError, ControlSkip}, status)
}
