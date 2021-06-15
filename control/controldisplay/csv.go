package controldisplay

import (
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/control/execute"
)

type CSVRenderer struct {
	columns *ResultColumns
}

func newGroupCsvRenderer() *CSVRenderer {
	return &CSVRenderer{}
}

func (r CSVRenderer) Render(tree *execute.ExecutionTree) [][]string {
	r.columns = newResultColumns(tree)
	return r.renderGroup(tree.Root)
}

func (r CSVRenderer) renderGroup(group *execute.ResultGroup) [][]string {
	var results [][]string
	for _, childGroup := range group.Groups {
		results = append(results, r.renderGroup(childGroup)...)
	}
	for _, run := range group.ControlRuns {
		results = append(results, r.renderControl(run, group)...)
	}
	return results
}

func (r CSVRenderer) renderControl(run *execute.ControlRun, group *execute.ResultGroup) [][]string {
	var res = make([][]string, len(run.Rows))

	groupColumns := r.columns.GroupColumns
	rowColumns := r.columns.ResultColumns

	for idx, row := range run.Rows {
		record := []string{}

		for _, groupColumn := range groupColumns {
			val, _ := helpers.GetFieldValueFromInterface(group, groupColumn.fieldName)
			record = append(record, typehelpers.ToString(val))
		}
		for _, rowColumn := range rowColumns {
			val, _ := helpers.GetFieldValueFromInterface(row, rowColumn.fieldName)
			record = append(record, typehelpers.ToString(val))
		}
		for _, fieldName := range r.columns.DimensionColumns {
			val, _ := helpers.GetFieldValueFromInterface(row, fieldName)
			if val == nil {
				val = ""
			}
			record = append(record, typehelpers.ToString(val))
		}
		tags := make(map[string]string)
		if run.Control.Tags != nil {
			tags = *run.Control.Tags
		}
		for _, prop := range r.columns.TagColumns {
			val := tags[prop]
			record = append(record, typehelpers.ToString(val))
		}

		res[idx] = record
	}
	return res
}
