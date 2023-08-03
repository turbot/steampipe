package db_client

type ScanMetadataRow struct {
	id           int64 `db:"id"`
	rowsFetched  int64 `db:"rows_fetched"`
	cacheHit     bool  `db:"cache_hit"`
	hydrateCalls int64 `db:"hydrate_calls"`
}
