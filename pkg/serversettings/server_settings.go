package serversettings

import (
	"time"
)

type ServerSettings struct {
	StartTime        time.Time `setting_key:"start_time" json:"start_time"`
	SteampipeVersion string    `setting_key:"steampipe_version" json:"steampipe_version"`
	FdwVersion       string    `setting_key:"fdw_version" json:"fdw_version"`
	CacheMaxTtl      int       `setting_key:"cache_max_ttl" json:"cache_max_ttl"`
	CacheMaxSizeMb   int       `setting_key:"cache_max_size_mb" json:"cache_max_size_mb"`
	CacheEnabled     bool      `setting_key:"cache_enabled" json:"cache_enabled"`
}
