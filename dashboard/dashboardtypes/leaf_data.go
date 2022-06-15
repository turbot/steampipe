package dashboardtypes

import (
	"database/sql"

	"github.com/turbot/steampipe/query/queryresult"
)

type ColumnSchema struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

func NewLeafDataColumnType(sqlType *sql.ColumnType) *ColumnSchema {
	return &ColumnSchema{
		Name:     sqlType.Name(),
		DataType: sqlType.DatabaseTypeName(),
	}
}

type LeafData struct {
	Columns []*ColumnSchema          `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

func NewLeafData(result *queryresult.SyncQueryResult) *LeafData {
	leafData := &LeafData{
		Rows:    make([]map[string]interface{}, len(result.Rows)),
		Columns: make([]*ColumnSchema, len(result.ColTypes)),
	}

	for i, c := range result.ColTypes {
		leafData.Columns[i] = NewLeafDataColumnType(c)
	}
	for rowIdx, row := range result.Rows {
		rowData := make(map[string]interface{}, len(result.ColTypes))
		for i, data := range row.(*queryresult.RowResult).Data {
			columnName := leafData.Columns[i].Name
			rowData[columnName] = data
		}

		leafData.Rows[rowIdx] = rowData
	}
	return leafData
}
