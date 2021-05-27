package controldisplay

import (
	"log"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/control/execute"
)

type GroupCsvRenderer struct {
	columns *ResultColumns
}

func newGroupCsvRenderer() *GroupCsvRenderer {
	return &GroupCsvRenderer{}
}

func (r GroupCsvRenderer) Render(tree *execute.ExecutionTree) [][]string {
	r.columns = newResultColumns(tree)
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

	groupColumns := getCsvColumns(*group)

	for idx, row := range run.Rows {
		record := []string{}
		rowColumns := getCsvColumns(*row)

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
