package controldisplay

import (
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type CSVRenderer struct {
	columns *ResultColumns
}

func newGroupCsvRenderer() *CSVRenderer {
	return &CSVRenderer{}
}

func (r CSVRenderer) Render(tree *controlexecute.ExecutionTree) [][]string {
	r.columns = newResultColumns(tree)
	return r.renderGroup(tree.Root)
}

func (r CSVRenderer) renderGroup(group *controlexecute.ResultGroup) [][]string {
	var results [][]string
	for _, childGroup := range group.Groups {
		results = append(results, r.renderGroup(childGroup)...)
	}
	for _, run := range group.ControlRuns {
		results = append(results, r.renderControl(run, group)...)
	}
	return results
}

func (r CSVRenderer) renderControl(run *controlexecute.ControlRun, group *controlexecute.ResultGroup) [][]string {
	var res = make([][]string, len(run.Rows))

	groupColumns := r.columns.GroupColumns
	rowColumns := r.columns.ResultColumns

	for idx, row := range run.Rows {
		record := []string{}

		for _, groupColumn := range groupColumns {
			val, _ := helpers.GetNestedFieldValueFromInterface(group, groupColumn.fieldName)
			record = append(record, typehelpers.ToString(val))
		}
		for _, rowColumn := range rowColumns {
			val, _ := helpers.GetNestedFieldValueFromInterface(row, rowColumn.fieldName)
			record = append(record, typehelpers.ToString(val))
		}
		dimensions := r.resultDimensionMap(row)
		for _, dimensionKey := range r.columns.DimensionColumns {
			if value, found := dimensions[dimensionKey]; found {
				record = append(record, value)
			} else {
				record = append(record, "")
			}
		}
		tags := make(map[string]string)
		if run.Control.Tags != nil {
			tags = run.Control.Tags
		}
		for _, prop := range r.columns.TagColumns {
			val := tags[prop]
			record = append(record, typehelpers.ToString(val))
		}

		res[idx] = record
	}
	return res
}

func (r CSVRenderer) resultDimensionMap(row *controlexecute.ResultRow) map[string]string {
	dimensionMap := map[string]string{}
	for _, dimension := range row.Dimensions {
		dimensionMap[dimension.Key] = dimension.Value
	}
	return dimensionMap
}
