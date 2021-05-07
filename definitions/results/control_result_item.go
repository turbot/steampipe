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
	// the parent control
	Control *modconfig.Control
}

func NewControlResultItem(control *modconfig.Control, row *RowResult, colTypes []*sql.ColumnType) (*ControlResultItem, error) {
	res := &ControlResultItem{Control: control}

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
