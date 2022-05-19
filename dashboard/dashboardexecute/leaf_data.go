package dashboardexecute

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
	Columns []*LeafDataColumnType `json:"columns"`
	Rows    [][]interface{}       `json:"rows"`
}

func NewLeafData(result *queryresult.SyncQueryResult) *LeafData {
	leafData := &LeafData{
		Rows:    make([][]interface{}, len(result.Rows)),
		Columns: make([]*LeafDataColumnType, len(result.ColTypes)),
	}

	for i, c := range result.ColTypes {
		leafData.Columns[i] = NewLeafDataColumnType(c)
	}
	for rowIdx, row := range result.Rows {
		rowData := make([]interface{}, len(result.ColTypes))
		copy(rowData, row.(*queryresult.RowResult).Data)
		leafData.Rows[rowIdx] = rowData
	}
	return leafData
}
