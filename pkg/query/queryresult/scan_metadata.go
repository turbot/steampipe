package queryresult

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"time"
)

type ScanMetadataRow struct {
	// the fields of this struct need to be public since these are populated by pgx using RowsToStruct
	Connection   string                  `db:"connection,optional" json:"connection"`
	Table        string                  `db:"table"  json:"table"`
	CacheHit     bool                    `db:"cache_hit"  json:"cache_hit"`
	RowsFetched  int64                   `db:"rows_fetched" json:"rows_fetched"`
	HydrateCalls int64                   `db:"hydrate_calls" json:"hydrate_calls"`
	StartTime    time.Time               `db:"start_time" json:"start_time"`
	DurationMs   int64                   `db:"duration_ms" json:"duration_ms"`
	Columns      []string                `db:"columns" json:"columns"`
	Limit        *int64                  `db:"limit" json:"limit,omitempty"`
	Quals        []grpc.SerializableQual `db:"quals" json:"quals,omitempty"`
}

func NewScanMetadataRow(connection string, table string, columns []string, quals map[string]*proto.Quals, startTime time.Time, diration time.Duration, limit int64, m *proto.QueryMetadata) ScanMetadataRow {
	res := ScanMetadataRow{
		Connection: connection,
		Table:      table,
		StartTime:  startTime,
		DurationMs: diration.Milliseconds(),
		Columns:    columns,
		Quals:      grpc.QualMapToSerializableSlice(quals),
	}
	if limit == -1 {
		res.Limit = nil
	} else {
		res.Limit = &limit
	}
	if m != nil {
		res.CacheHit = m.CacheHit
		res.RowsFetched = m.RowsFetched
		res.HydrateCalls = m.HydrateCalls
	}
	return res
}

// AsResultRow returns the ScanMetadata as a map[string]interface which can be returned as a query result
func (m ScanMetadataRow) AsResultRow() map[string]any {
	res := map[string]any{
		"connection":    m.Connection,
		"table":         m.Table,
		"cache_hit":     m.CacheHit,
		"rows_fetched":  m.RowsFetched,
		"hydrate_calls": m.HydrateCalls,
		"start_time":    m.StartTime,
		"duration_ms":   m.DurationMs,
		"columns":       m.Columns,
		"quals":         m.Quals,
	}
	// explicitly set limit to nil if needed (otherwise postgres returns `1`)
	if m.Limit != nil {
		res["limit"] = *m.Limit
	} else {
		res["limit"] = nil // Explicitly set nil
	}
	return res
}

type QueryRowSummary struct {
	UncachedRowsFetched int64 `db:"uncached_rows_fetched" json:"uncached_rows_fetched"`
	CachedRowsFetched   int64 `db:"cached_rows_fetched" json:"cached_rows_fetched"`
	HydrateCalls        int64 `db:"hydrate_calls" json:"hydrate_calls"`
	ScanCount           int64 `db:"scan_count" json:"scan_count"`
	ConnectionCount     int64 `db:"connection_count" json:"connection_count"`
	// map connections to the scans
	connections map[string]struct{}
}

func NewQueryRowSummary() *QueryRowSummary {
	return &QueryRowSummary{
		connections: make(map[string]struct{}),
	}
}
func (s *QueryRowSummary) AsResultRow() map[string]any {
	res := map[string]any{
		"uncached_rows_fetched": s.UncachedRowsFetched,
		"cached_rows_fetched":   s.CachedRowsFetched,
		"hydrate_calls":         s.HydrateCalls,
		"scan_count":            s.ScanCount,
		"connection_count":      s.ConnectionCount,
	}

	return res
}

func (s *QueryRowSummary) Update(m ScanMetadataRow) {
	if m.CacheHit {
		s.CachedRowsFetched += m.RowsFetched
	} else {
		s.UncachedRowsFetched += m.RowsFetched
	}
	s.HydrateCalls += m.HydrateCalls
	s.ScanCount++
	s.connections[m.Connection] = struct{}{}
	s.ConnectionCount = int64(len(s.connections))
}
