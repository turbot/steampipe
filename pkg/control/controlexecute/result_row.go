package controlexecute

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

type ResultRows []*ResultRow

// ToLeafData converts the result rows to snapshot data format
func (r ResultRows) ToLeafData(dimensionSchema map[string]*queryresult.ColumnDef) *dashboardtypes.LeafData {
	var res = &dashboardtypes.LeafData{
		Columns: []*queryresult.ColumnDef{
			{Name: "reason", DataType: "TEXT"},
			{Name: "resource", DataType: "TEXT"},
			{Name: "status", DataType: "TEXT"},
		},
		Rows: make([]map[string]interface{}, len(r)),
	}
	for _, d := range dimensionSchema {
		res.Columns = append(res.Columns, d)
	}
	for i, row := range r {
		res.Rows[i] = map[string]interface{}{
			"reason":   row.Reason,
			"resource": row.Resource,
			"status":   row.Status,
		}
		// flatten dimensions
		for _, d := range row.Dimensions {
			res.Rows[i][d.Key] = d.Value
		}
	}
	return res
}

// ResultRow is the result of a control execution for a single resource
type ResultRow struct {
	// reason for the status
	Reason string `json:"reason" csv:"reason"`
	// resource name
	Resource string `json:"resource" csv:"resource"`
	// status of the row (ok, info, alarm, error, skip)
	Status string `json:"status" csv:"status"`
	// dimensions for this row
	Dimensions []Dimension `json:"dimensions"`
	// parent control run
	Run *ControlRun `json:"-"`
	// source control
	Control *modconfig.Control `json:"-" csv:"control_id:UnqualifiedName,control_title:Title,control_description:Description"`
}

// GetDimensionValue returns the value for a dimension key. Returns an empty string with 'false' if not found
func (r *ResultRow) GetDimensionValue(key string) string {
	for _, dim := range r.Dimensions {
		if dim.Key == key {
			return dim.Value
		}
	}
	return ""
}

// AddDimension checks whether a column value is a scalar type, and if so adds it to the Dimensions map
func (r *ResultRow) AddDimension(c *queryresult.ColumnDef, val interface{}) {
	r.Dimensions = append(r.Dimensions, Dimension{
		Key:     c.Name,
		Value:   typehelpers.ToString(val),
		SqlType: c.DataType,
	})
}

func NewResultRow(run *ControlRun, row *queryresult.RowResult, cols []*queryresult.ColumnDef) (*ResultRow, error) {
	// validate the required columns exist in the result
	if err := validateColumns(cols); err != nil {
		return nil, err
	}
	res := &ResultRow{
		Run:     run,
		Control: run.Control,
	}

	// was there a SQL error _executing the control
	// Note: this is different from the control state being 'error'
	if row.Error != nil {
		return nil, row.Error
	}

	for i, c := range cols {
		switch c.Name {
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
			val := row.Data[i]
			// isScalar may mutate the ColumnDef struct by lazily populating the internal isScalar property
			if c.IsScalar(val) {
				res.AddDimension(c, val)
			}
		}
	}
	return res, nil
}

func IsValidControlStatus(status string) bool {
	return helpers.StringSliceContains([]string{constants.ControlOk, constants.ControlAlarm, constants.ControlInfo, constants.ControlError, constants.ControlSkip}, status)
}

func validateColumns(cols []*queryresult.ColumnDef) error {
	requiredColumns := []string{"reason", "resource", "status"}
	var missingColumns []string
	for _, col := range requiredColumns {
		if !columnTypesContainsColumn(col, cols) {
			missingColumns = append(missingColumns, col)
		}
	}
	if len(missingColumns) > 0 {
		return fmt.Errorf("control result is missing required %s: %v", utils.Pluralize("column", len(missingColumns)), missingColumns)
	}
	return nil
}

func columnTypesContainsColumn(col string, colTypes []*queryresult.ColumnDef) bool {
	for _, ct := range colTypes {
		if ct.Name == col {
			return true
		}
	}
	return false
}
