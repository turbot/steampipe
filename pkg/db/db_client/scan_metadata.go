package db_client

type ScanMetadataRow struct {
	// the fields of this struct need to be public since these are populated by pgx using RowsToStruct
	Id           int64 `db:"id"`
	RowsFetched  int64 `db:"rows_fetched"`
	CacheHit     bool  `db:"cache_hit"`
	HydrateCalls int64 `db:"hydrate_calls"`
}
