package backend

import (
	"github.com/turbot/steampipe/pkg/query/queryresult"
)

type RowReader interface {
	Read(columnValues []interface{}, cols []*queryresult.ColumnDef) ([]any, error)
}

func RowReaderFactory(backend DBClientBackendType) RowReader {
	var reader RowReader
	switch backend {
	case PostgresDBClientBackend:
		reader = &PgxRowReader{}
	case SqliteDBClientBackend:
		reader = &SqliteRowReader{}
	default:

	}
	return reader
}
