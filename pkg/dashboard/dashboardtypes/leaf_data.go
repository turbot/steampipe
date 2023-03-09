package dashboardtypes

import (
	"github.com/turbot/steampipe/pkg/query/queryresult"
)

type LeafData struct {
	Columns []*queryresult.ColumnDef `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

func NewLeafData(result *queryresult.SyncQueryResult) *LeafData {
	leafData := &LeafData{
		Rows:    make([]map[string]interface{}, len(result.Rows)),
		Columns: result.Cols,
	}

	for rowIdx, row := range result.Rows {
		rowData := make(map[string]interface{}, len(result.Cols))
		for i, data := range row.(*queryresult.RowResult).Data {
			columnName := leafData.Columns[i].Name
			rowData[columnName] = data
		}

		leafData.Rows[rowIdx] = rowData
	}
	return leafData
}
