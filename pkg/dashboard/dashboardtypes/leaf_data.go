package dashboardtypes

import (
	"github.com/turbot/steampipe/pkg/query/queryresult"
)

type ColumnSchema struct {
	Name          string          `json:"name"`
	DataType      string          `json:"data_type"`
	SqlColumnType *sql.ColumnType `json:"-"`
}

func NewLeafDataColumnType(sqlType *sql.ColumnType) *ColumnSchema {
	return &ColumnSchema{
		Name:          sqlType.Name(),
		DataType:      sqlType.DatabaseTypeName(),
		SqlColumnType: sqlType,
	}
}

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
