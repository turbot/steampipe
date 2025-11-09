package queryresult

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// TestNewScanMetadataRow tests the creation of a new ScanMetadataRow
func TestNewScanMetadataRow(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 100 * time.Millisecond

	tests := map[string]struct {
		connection string
		table      string
		columns    []string
		quals      map[string]*proto.Quals
		startTime  time.Time
		duration   time.Duration
		limit      int64
		metadata   *proto.QueryMetadata
		validate   func(*testing.T, ScanMetadataRow)
	}{
		"simple_metadata": {
			connection: "aws",
			table:      "aws_s3_bucket",
			columns:    []string{"name", "region"},
			quals:      make(map[string]*proto.Quals),
			startTime:  startTime,
			duration:   duration,
			limit:      100,
			metadata:   nil,
			validate: func(t *testing.T, row ScanMetadataRow) {
				assert.Equal(t, "aws", row.Connection)
				assert.Equal(t, "aws_s3_bucket", row.Table)
				assert.Equal(t, []string{"name", "region"}, row.Columns)
				assert.Equal(t, startTime, row.StartTime)
				assert.Equal(t, int64(100), row.DurationMs)
				assert.NotNil(t, row.Limit)
				assert.Equal(t, int64(100), *row.Limit)
			},
		},
		"with_metadata": {
			connection: "aws",
			table:      "aws_ec2_instance",
			columns:    []string{"instance_id", "state"},
			quals:      make(map[string]*proto.Quals),
			startTime:  startTime,
			duration:   duration,
			limit:      50,
			metadata: &proto.QueryMetadata{
				CacheHit:     true,
				RowsFetched:  25,
				HydrateCalls: 10,
			},
			validate: func(t *testing.T, row ScanMetadataRow) {
				assert.True(t, row.CacheHit)
				assert.Equal(t, int64(25), row.RowsFetched)
				assert.Equal(t, int64(10), row.HydrateCalls)
			},
		},
		"unlimited_limit": {
			connection: "aws",
			table:      "aws_s3_bucket",
			columns:    []string{"name"},
			quals:      make(map[string]*proto.Quals),
			startTime:  startTime,
			duration:   duration,
			limit:      -1, // -1 means unlimited
			metadata:   nil,
			validate: func(t *testing.T, row ScanMetadataRow) {
				assert.Nil(t, row.Limit, "Limit should be nil for unlimited queries")
			},
		},
		"empty_columns": {
			connection: "test",
			table:      "test_table",
			columns:    []string{},
			quals:      make(map[string]*proto.Quals),
			startTime:  startTime,
			duration:   duration,
			limit:      100,
			metadata:   nil,
			validate: func(t *testing.T, row ScanMetadataRow) {
				assert.Empty(t, row.Columns)
			},
		},
		"zero_duration": {
			connection: "test",
			table:      "test_table",
			columns:    []string{"col1"},
			quals:      make(map[string]*proto.Quals),
			startTime:  startTime,
			duration:   0,
			limit:      100,
			metadata:   nil,
			validate: func(t *testing.T, row ScanMetadataRow) {
				assert.Equal(t, int64(0), row.DurationMs)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			row := NewScanMetadataRow(
				tc.connection,
				tc.table,
				tc.columns,
				tc.quals,
				tc.startTime,
				tc.duration,
				tc.limit,
				tc.metadata,
			)

			tc.validate(t, row)
		})
	}
}

// TestScanMetadataRow_AsResultRow tests converting ScanMetadataRow to result row
func TestScanMetadataRow_AsResultRow(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 100 * time.Millisecond

	tests := map[string]struct {
		row      ScanMetadataRow
		validate func(*testing.T, map[string]any)
	}{
		"complete_row": {
			row: NewScanMetadataRow(
				"aws",
				"aws_s3_bucket",
				[]string{"name", "region"},
				make(map[string]*proto.Quals),
				startTime,
				duration,
				100,
				&proto.QueryMetadata{
					CacheHit:     true,
					RowsFetched:  50,
					HydrateCalls: 10,
				},
			),
			validate: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "aws", result["connection"])
				assert.Equal(t, "aws_s3_bucket", result["table"])
				assert.True(t, result["cache_hit"].(bool))
				assert.Equal(t, int64(50), result["rows_fetched"])
				assert.Equal(t, int64(10), result["hydrate_calls"])
				assert.Equal(t, startTime, result["start_time"])
				assert.Equal(t, int64(100), result["duration_ms"])
				assert.NotNil(t, result["columns"])
				// Quals may be nil for empty map
				_, hasQuals := result["quals"]
				assert.True(t, hasQuals, "quals key should exist")
				assert.Equal(t, int64(100), result["limit"])
			},
		},
		"nil_limit": {
			row: NewScanMetadataRow(
				"aws",
				"aws_s3_bucket",
				[]string{"name"},
				make(map[string]*proto.Quals),
				startTime,
				duration,
				-1, // unlimited
				nil,
			),
			validate: func(t *testing.T, result map[string]any) {
				// Limit should be explicitly nil for unlimited queries
				limit, exists := result["limit"]
				assert.True(t, exists, "limit key should exist")
				assert.Nil(t, limit, "limit should be nil for unlimited queries")
			},
		},
		"cache_miss": {
			row: NewScanMetadataRow(
				"aws",
				"aws_ec2_instance",
				[]string{"instance_id"},
				make(map[string]*proto.Quals),
				startTime,
				duration,
				100,
				&proto.QueryMetadata{
					CacheHit:     false,
					RowsFetched:  100,
					HydrateCalls: 50,
				},
			),
			validate: func(t *testing.T, result map[string]any) {
				assert.False(t, result["cache_hit"].(bool))
				assert.Equal(t, int64(100), result["rows_fetched"])
				assert.Equal(t, int64(50), result["hydrate_calls"])
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.row.AsResultRow()
			assert.NotNil(t, result)
			tc.validate(t, result)
		})
	}
}

// TestNewQueryRowSummary tests the creation of a new QueryRowSummary
func TestNewQueryRowSummary(t *testing.T) {
	tests := map[string]struct {
		validate func(*testing.T, *QueryRowSummary)
	}{
		"new_summary": {
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.NotNil(t, summary)
				assert.Equal(t, int64(0), summary.UncachedRowsFetched)
				assert.Equal(t, int64(0), summary.CachedRowsFetched)
				assert.Equal(t, int64(0), summary.HydrateCalls)
				assert.Equal(t, int64(0), summary.ScanCount)
				assert.Equal(t, int64(0), summary.ConnectionCount)
				assert.NotNil(t, summary.connections)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			summary := NewQueryRowSummary()
			tc.validate(t, summary)
		})
	}
}

// TestQueryRowSummary_Update tests updating QueryRowSummary with scan metadata
func TestQueryRowSummary_Update(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 100 * time.Millisecond

	tests := map[string]struct {
		scans    []ScanMetadataRow
		validate func(*testing.T, *QueryRowSummary)
	}{
		"single_cached_scan": {
			scans: []ScanMetadataRow{
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  50,
						HydrateCalls: 10,
					},
				),
			},
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.Equal(t, int64(50), summary.CachedRowsFetched)
				assert.Equal(t, int64(0), summary.UncachedRowsFetched)
				assert.Equal(t, int64(10), summary.HydrateCalls)
				assert.Equal(t, int64(1), summary.ScanCount)
				assert.Equal(t, int64(1), summary.ConnectionCount)
			},
		},
		"single_uncached_scan": {
			scans: []ScanMetadataRow{
				NewScanMetadataRow(
					"aws",
					"aws_ec2_instance",
					[]string{"instance_id"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     false,
						RowsFetched:  100,
						HydrateCalls: 50,
					},
				),
			},
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.Equal(t, int64(0), summary.CachedRowsFetched)
				assert.Equal(t, int64(100), summary.UncachedRowsFetched)
				assert.Equal(t, int64(50), summary.HydrateCalls)
				assert.Equal(t, int64(1), summary.ScanCount)
				assert.Equal(t, int64(1), summary.ConnectionCount)
			},
		},
		"multiple_scans_same_connection": {
			scans: []ScanMetadataRow{
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  50,
						HydrateCalls: 10,
					},
				),
				NewScanMetadataRow(
					"aws",
					"aws_ec2_instance",
					[]string{"instance_id"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     false,
						RowsFetched:  100,
						HydrateCalls: 50,
					},
				),
			},
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.Equal(t, int64(50), summary.CachedRowsFetched)
				assert.Equal(t, int64(100), summary.UncachedRowsFetched)
				assert.Equal(t, int64(60), summary.HydrateCalls)
				assert.Equal(t, int64(2), summary.ScanCount)
				assert.Equal(t, int64(1), summary.ConnectionCount, "Should count unique connections")
			},
		},
		"multiple_scans_different_connections": {
			scans: []ScanMetadataRow{
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  50,
						HydrateCalls: 10,
					},
				),
				NewScanMetadataRow(
					"gcp",
					"gcp_compute_instance",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     false,
						RowsFetched:  75,
						HydrateCalls: 25,
					},
				),
				NewScanMetadataRow(
					"azure",
					"azure_compute_virtual_machine",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  30,
						HydrateCalls: 5,
					},
				),
			},
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.Equal(t, int64(80), summary.CachedRowsFetched, "50 + 30")
				assert.Equal(t, int64(75), summary.UncachedRowsFetched)
				assert.Equal(t, int64(40), summary.HydrateCalls, "10 + 25 + 5")
				assert.Equal(t, int64(3), summary.ScanCount)
				assert.Equal(t, int64(3), summary.ConnectionCount, "Should count 3 unique connections")
			},
		},
		"mixed_cache_hits_and_misses": {
			scans: []ScanMetadataRow{
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  100,
						HydrateCalls: 0,
					},
				),
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     false,
						RowsFetched:  200,
						HydrateCalls: 100,
					},
				),
				NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  50,
						HydrateCalls: 10,
					},
				),
			},
			validate: func(t *testing.T, summary *QueryRowSummary) {
				assert.Equal(t, int64(150), summary.CachedRowsFetched, "100 + 50")
				assert.Equal(t, int64(200), summary.UncachedRowsFetched)
				assert.Equal(t, int64(110), summary.HydrateCalls, "0 + 100 + 10")
				assert.Equal(t, int64(3), summary.ScanCount)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			summary := NewQueryRowSummary()

			for _, scan := range tc.scans {
				summary.Update(scan)
			}

			tc.validate(t, summary)
		})
	}
}

// TestQueryRowSummary_AsResultRow tests converting QueryRowSummary to result row
func TestQueryRowSummary_AsResultRow(t *testing.T) {
	tests := map[string]struct {
		summary  *QueryRowSummary
		validate func(*testing.T, map[string]any)
	}{
		"empty_summary": {
			summary: NewQueryRowSummary(),
			validate: func(t *testing.T, result map[string]any) {
				assert.Equal(t, int64(0), result["uncached_rows_fetched"])
				assert.Equal(t, int64(0), result["cached_rows_fetched"])
				assert.Equal(t, int64(0), result["hydrate_calls"])
				assert.Equal(t, int64(0), result["scan_count"])
				assert.Equal(t, int64(0), result["connection_count"])
			},
		},
		"populated_summary": {
			summary: func() *QueryRowSummary {
				s := NewQueryRowSummary()
				startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				duration := 100 * time.Millisecond

				scan := NewScanMetadataRow(
					"aws",
					"aws_s3_bucket",
					[]string{"name"},
					make(map[string]*proto.Quals),
					startTime,
					duration,
					100,
					&proto.QueryMetadata{
						CacheHit:     true,
						RowsFetched:  50,
						HydrateCalls: 10,
					},
				)
				s.Update(scan)
				return s
			}(),
			validate: func(t *testing.T, result map[string]any) {
				assert.Equal(t, int64(50), result["cached_rows_fetched"])
				assert.Equal(t, int64(0), result["uncached_rows_fetched"])
				assert.Equal(t, int64(10), result["hydrate_calls"])
				assert.Equal(t, int64(1), result["scan_count"])
				assert.Equal(t, int64(1), result["connection_count"])
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.summary.AsResultRow()
			assert.NotNil(t, result)
			tc.validate(t, result)
		})
	}
}
