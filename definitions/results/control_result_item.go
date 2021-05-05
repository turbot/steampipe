package results

import (
	"database/sql"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlStatus string

const (
	ControlOk      = "ok"
	ControlAlarm   = "alarm"
	ControlSkipped = "skipped"
	ControlInfo    = "info"
	ControlError   = "error"
)

// ControlResultItem :: the result of a control for a single resource
type ControlResultItem struct {
	Reason   string
	Resource string
	Status   ControlStatus
	Error    error
	// the parent control
	Control *modconfig.Control
}

func NewControlResultItem(control *modconfig.Control, row *RowResult, colTypes []*sql.ColumnType) (*ControlResultItem, error) {
	res := &ControlResultItem{Control: control}

	if row.Error != nil {
		return nil, res.Error
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
