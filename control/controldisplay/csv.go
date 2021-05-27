package controldisplay

import (
	"log"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/control/execute"
)

type GroupCsvRenderer struct {
	columns *execute.ResultColumns
}

func newGroupCsvRenderer(flatResults *execute.ResultColumns) *GroupCsvRenderer {
	return &GroupCsvRenderer{
		columns: flatResults,
	}
}

func (r GroupCsvRenderer) Render(tree *execute.ExecutionTree) [][]string {
	return r.renderGroup(tree.Root)
}

func (r GroupCsvRenderer) renderGroup(group *execute.ResultGroup) [][]string {
	log.Printf("[TRACE] begin group  csv render '%s'\n", group.GroupId)
	defer log.Printf("[TRACE] end table csv render'%s'\n", group.GroupId)
	var results [][]string
	for _, childGroup := range group.Groups {
		results = append(results, r.renderGroup(childGroup)...)
	}
	for _, run := range group.ControlRuns {
		results = append(results, r.renderControl(run, group)...)
	}
	return results
}

func (r GroupCsvRenderer) renderControl(run *execute.ControlRun, group *execute.ResultGroup) [][]string {
	var res = make([][]string, len(run.Rows))
	_, groupColumnsKeyOrder := execute.ResultGroup{}.CsvColumns()
	_, resultColumnsKeyOrder := execute.ResultRow{}.CsvColumns()

	for idx, row := range run.Rows {
		record := []string{}
		for _, groupColumnName := range groupColumnsKeyOrder {
			val, _ := helpers.GetFieldValueFromInterface(group, r.columns.GroupColumns[groupColumnName])
			record = append(record, typehelpers.ToString(val))
		}
		for _, resultColumnName := range resultColumnsKeyOrder {
			val, _ := helpers.GetFieldValueFromInterface(row, r.columns.ResultColumns[resultColumnName])
			record = append(record, typehelpers.ToString(val))
		}
		// for _, fieldName := range r.columns.ResultColumns {
		// 	if helpers.StringSliceContains(resultColumnsKeyOrder, fieldName) {
		// 		continue
		// 	}
		// 	val, _ := helpers.GetFieldValueFromInterface(row, fieldName)
		// 	if val == nil {
		// 		val = ""
		// 	}
		// 	record = append(record, typehelpers.ToString(val))
		// }
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
