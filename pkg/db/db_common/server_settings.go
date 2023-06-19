package db_common

import (
	"time"
)

type ServerSettings struct {
	StartTime        time.Time `db:"start_time"`
	SteampipeVersion string    `db:"steampipe_version"`
	FdwVersion       string    `db:"fdw_version"`
	CacheMaxTtl      int       `db:"cache_max_ttl"`
	CacheMaxSizeMb   int       `db:"cache_max_size_mb"`
	CacheEnabled     bool      `db:"cache_enabled"`
}
