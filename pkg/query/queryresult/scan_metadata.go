package queryresult

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"time"
)

type ScanMetadataRow struct {
	// the fields of this struct need to be public since these are populated by pgx using RowsToStruct
	Id           int64         `db:"id" json:"id,omitempty"`
	Table        string        `db:"table"  json:"table"`
	CacheHit     bool          `db:"cache_hit"  json:"cache_hit"`
	RowsFetched  int64         `db:"rows_fetched" json:"rows_fetched"`
	HydrateCalls int64         `db:"hydrate_calls" json:"hydrate_calls"`
	StartTime    time.Time     `db:"start_time" json:"start_time"`
	Duration     float64       `db:"duration" json:"duration"`
	Columns      []string      `db:"columns" json:"columns"`
	Limit        *int64        `db:"limit" json:"limit"`
	Quals        []*proto.Qual `db:"-" json:"quals"`
}

//
//{Name: "id", Type: proto.ColumnType_INT},
//{Name: "table", Type: proto.ColumnType_STRING},
//{Name: "cache_hit", Type: proto.ColumnType_BOOL},
//{Name: "rows_fetched", Type: proto.ColumnType_INT},
//{Name: "hydrate_calls", Type: proto.ColumnType_INT},
//{Name: "start_time", Type: proto.ColumnType_TIMESTAMP},
//{Name: "duration", Type: proto.ColumnType_DOUBLE},
//{Name: "columns", Type: proto.ColumnType_JSON},
//{Name: "limit", Type: proto.ColumnType_INT},
//{Name: "quals", Type: proto.ColumnType_STRING},
