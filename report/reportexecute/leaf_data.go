package reportexecute

import (
	"database/sql"

	"github.com/turbot/steampipe/query/queryresult"
)

type LeafDataColumnType struct {
	Name     string `json:"name"`
	DataType string `json:"data_type_name"`
}

func NewLeafDataColumnType(sqlType *sql.ColumnType) *LeafDataColumnType {
	return &LeafDataColumnType{
		Name:     sqlType.Name(),
		DataType: sqlType.DatabaseTypeName(),
	}
}

type LeafData struct {
	Columns []*LeafDataColumnType    `json:"columns"`
	Items   []map[string]interface{} `json:"items"`
}

func NewLeafData(result *queryresult.SyncQueryResult) *LeafData {
	leafData := &LeafData{
		Items:   make([]map[string]interface{}, len(result.Rows)),
		Columns: make([]*LeafDataColumnType, len(result.ColTypes)),
	}

	for i, c := range result.ColTypes {
		leafData.Columns[i] = NewLeafDataColumnType(c)
	}
	for rowIdx, row := range result.Rows {
		rowData := make(map[string]interface{}, len(result.ColTypes))
		for columnIdx, columnVal := range row.(*queryresult.RowResult).Data {
			rowData[leafData.Columns[columnIdx].Name] = columnVal
		}
		leafData.Items[rowIdx] = rowData
	}
	return leafData
}
