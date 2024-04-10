package queryresult

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"time"
)

type ScanMetadataRow struct {
	// the fields of this struct need to be public since these are populated by pgx using RowsToStruct
	Id           int64                    `db:"id" json:"-"`
	Connection   string                   `db:"connection" json:"connection"`
	Table        string                   `db:"table"  json:"table"`
	CacheHit     bool                     `db:"cache_hit"  json:"cache_hit"`
	RowsFetched  int64                    `db:"rows_fetched" json:"rows_fetched"`
	HydrateCalls int64                    `db:"hydrate_calls" json:"hydrate_calls"`
	StartTime    time.Time                `db:"start_time" json:"start_time"`
	Duration     float64                  `db:"duration" json:"duration"`
	Columns      []string                 `db:"columns" json:"columns"`
	Limit        *int64                   `db:"limit" json:"limit"`
	Quals        []*grpc.SerializableQual `db:"quals" json:"quals"`
}
