package controlresult

import (
	"database/sql"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlStatus string

const (
	ControlOk    ControlStatus = "ok"
	ControlAlarm               = "alarm"
	ControlSkip                = "skip"
	ControlInfo                = "info"
	ControlError               = "error"
)

// ResultRow is the result of a control execution for a single resource
type ResultRow struct {
	Reason     string             `json:"reason"`
	Resource   string             `json:"resource"`
	Status     ControlStatus      `json:"status"`
	Dimensions map[string]string  `json:"dimensions"`
	Control    *modconfig.Control `json:"-"`
}

func NewResultRow(control *modconfig.Control, row *queryresult.RowResult, colTypes []*sql.ColumnType) (*ResultRow, error) {
	res := &ResultRow{Control: control}

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
			status := ControlStatus(typehelpers.ToString(row.Data[i]))
			//if !ok {
			//	return nil, fmt.Errorf("invalid control status '%v'", row.Data[i])
			//}
			res.Status = status
		}
	}
	return res, nil
}
