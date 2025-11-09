package queryresult

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// TestNewTimingResultStream tests the creation of a new TimingResultStream
func TestNewTimingResultStream(t *testing.T) {
	tests := map[string]struct {
		validate func(*testing.T, TimingResultStream)
	}{
		"creates_valid_stream": {
			validate: func(t *testing.T, stream TimingResultStream) {
				assert.NotNil(t, stream.Stream)
			},
		},
		"multiple_streams": {
			validate: func(t *testing.T, stream TimingResultStream) {
				stream2 := NewTimingResultStream()
				assert.NotNil(t, stream.Stream)
				assert.NotNil(t, stream2.Stream)
				// They should be different channels
				assert.NotEqual(t, stream.Stream, stream2.Stream)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stream := NewTimingResultStream()
			tc.validate(t, stream)
		})
	}
}

// TestTimingResultStream_SetAndGetTiming tests setting and getting timing
func TestTimingResultStream_SetAndGetTiming(t *testing.T) {
	tests := map[string]struct {
		result   *TimingResult
		validate func(*testing.T, any)
	}{
		"simple_timing": {
			result: &TimingResult{
				DurationMs:   100,
				RowsReturned: 50,
			},
			validate: func(t *testing.T, timing any) {
				result, ok := timing.(*TimingResult)
				assert.True(t, ok, "Should be a TimingResult")
				assert.Equal(t, int64(100), result.DurationMs)
				assert.Equal(t, int64(50), result.RowsReturned)
			},
		},
		"complete_timing": {
			result: &TimingResult{
				DurationMs:          500,
				RowsReturned:        100,
				ScanCount:           3,
				UncachedRowsFetched: 75,
				CachedRowsFetched:   25,
				HydrateCalls:        50,
				ConnectionCount:     2,
			},
			validate: func(t *testing.T, timing any) {
				result, ok := timing.(*TimingResult)
				assert.True(t, ok, "Should be a TimingResult")
				assert.Equal(t, int64(500), result.DurationMs)
				assert.Equal(t, int64(100), result.RowsReturned)
				assert.Equal(t, int64(3), result.ScanCount)
				assert.Equal(t, int64(75), result.UncachedRowsFetched)
				assert.Equal(t, int64(25), result.CachedRowsFetched)
				assert.Equal(t, int64(50), result.HydrateCalls)
				assert.Equal(t, int64(2), result.ConnectionCount)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stream := NewTimingResultStream()

			// Set timing in a goroutine to avoid blocking
			go func() {
				stream.SetTiming(tc.result)
			}()

			// Get timing
			timing := stream.GetTiming()
			assert.NotNil(t, timing)
			tc.validate(t, timing)
		})
	}
}

// TestTimingResult_Initialise tests initializing a TimingResult
func TestTimingResult_Initialise(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 100 * time.Millisecond

	tests := map[string]struct {
		summary  *QueryRowSummary
		scans    []*ScanMetadataRow
		validate func(*testing.T, *TimingResult)
	}{
		"empty_summary": {
			summary: NewQueryRowSummary(),
			scans:   []*ScanMetadataRow{},
			validate: func(t *testing.T, result *TimingResult) {
				assert.Equal(t, int64(0), result.ScanCount)
				assert.Equal(t, int64(0), result.ConnectionCount)
				assert.Equal(t, int64(0), result.UncachedRowsFetched)
				assert.Equal(t, int64(0), result.CachedRowsFetched)
				assert.Equal(t, int64(0), result.HydrateCalls)
				assert.Empty(t, result.Scans)
			},
		},
		"single_scan": {
			summary: func() *QueryRowSummary {
				s := NewQueryRowSummary()
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
			scans: func() []*ScanMetadataRow {
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
				return []*ScanMetadataRow{&scan}
			}(),
			validate: func(t *testing.T, result *TimingResult) {
				assert.Equal(t, int64(1), result.ScanCount)
				assert.Equal(t, int64(1), result.ConnectionCount)
				assert.Equal(t, int64(0), result.UncachedRowsFetched)
				assert.Equal(t, int64(50), result.CachedRowsFetched)
				assert.Equal(t, int64(10), result.HydrateCalls)
				assert.Len(t, result.Scans, 1)
			},
		},
		"multiple_scans_different_connections": {
			summary: func() *QueryRowSummary {
				s := NewQueryRowSummary()
				scans := []ScanMetadataRow{
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
				}
				for _, scan := range scans {
					s.Update(scan)
				}
				return s
			}(),
			scans: func() []*ScanMetadataRow {
				scan1 := NewScanMetadataRow(
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
				scan2 := NewScanMetadataRow(
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
				)
				return []*ScanMetadataRow{&scan1, &scan2}
			}(),
			validate: func(t *testing.T, result *TimingResult) {
				assert.Equal(t, int64(2), result.ScanCount)
				assert.Equal(t, int64(2), result.ConnectionCount)
				assert.Equal(t, int64(75), result.UncachedRowsFetched)
				assert.Equal(t, int64(50), result.CachedRowsFetched)
				assert.Equal(t, int64(35), result.HydrateCalls, "10 + 25")
				assert.Len(t, result.Scans, 2)
			},
		},
		"nil_scans": {
			summary: func() *QueryRowSummary {
				s := NewQueryRowSummary()
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
			scans: nil,
			validate: func(t *testing.T, result *TimingResult) {
				assert.Equal(t, int64(1), result.ScanCount)
				assert.Nil(t, result.Scans)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := &TimingResult{}
			result.Initialise(tc.summary, tc.scans)
			tc.validate(t, result)
		})
	}
}

// TestTimingResult_GetTiming tests the GetTiming method
func TestTimingResult_GetTiming(t *testing.T) {
	tests := map[string]struct {
		result   TimingResult
		validate func(*testing.T, any)
	}{
		"simple_result": {
			result: TimingResult{
				DurationMs:   100,
				RowsReturned: 50,
			},
			validate: func(t *testing.T, timing any) {
				result, ok := timing.(TimingResult)
				assert.True(t, ok, "Should be a TimingResult")
				assert.Equal(t, int64(100), result.DurationMs)
				assert.Equal(t, int64(50), result.RowsReturned)
			},
		},
		"complete_result": {
			result: TimingResult{
				DurationMs:          500,
				RowsReturned:        100,
				ScanCount:           3,
				UncachedRowsFetched: 75,
				CachedRowsFetched:   25,
				HydrateCalls:        50,
				ConnectionCount:     2,
			},
			validate: func(t *testing.T, timing any) {
				result, ok := timing.(TimingResult)
				assert.True(t, ok, "Should be a TimingResult")
				assert.Equal(t, int64(500), result.DurationMs)
				assert.Equal(t, int64(100), result.RowsReturned)
				assert.Equal(t, int64(3), result.ScanCount)
				assert.Equal(t, int64(75), result.UncachedRowsFetched)
				assert.Equal(t, int64(25), result.CachedRowsFetched)
				assert.Equal(t, int64(50), result.HydrateCalls)
				assert.Equal(t, int64(2), result.ConnectionCount)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			timing := tc.result.GetTiming()
			assert.NotNil(t, timing)
			tc.validate(t, timing)
		})
	}
}

// TestTimingResultStream_Buffering tests that the channel is buffered
func TestTimingResultStream_Buffering(t *testing.T) {
	stream := NewTimingResultStream()

	// Should be able to set without blocking (because channel is buffered)
	result := &TimingResult{
		DurationMs:   100,
		RowsReturned: 50,
	}
	stream.SetTiming(result)

	// Get timing
	timing := stream.GetTiming()
	assert.NotNil(t, timing)

	timingResult, ok := timing.(*TimingResult)
	assert.True(t, ok)
	assert.Equal(t, int64(100), timingResult.DurationMs)
	assert.Equal(t, int64(50), timingResult.RowsReturned)
}

// TestTimingResult_WithScans tests TimingResult with actual scan data
func TestTimingResult_WithScans(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := 100 * time.Millisecond

	scan1 := NewScanMetadataRow(
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
	)

	scan2 := NewScanMetadataRow(
		"aws",
		"aws_ec2_instance",
		[]string{"instance_id", "state"},
		make(map[string]*proto.Quals),
		startTime.Add(1*time.Second),
		duration,
		100,
		&proto.QueryMetadata{
			CacheHit:     false,
			RowsFetched:  100,
			HydrateCalls: 50,
		},
	)

	summary := NewQueryRowSummary()
	summary.Update(scan1)
	summary.Update(scan2)

	result := &TimingResult{
		DurationMs:   1100,
		RowsReturned: 150,
	}

	result.Initialise(summary, []*ScanMetadataRow{&scan1, &scan2})

	assert.Equal(t, int64(1100), result.DurationMs)
	assert.Equal(t, int64(150), result.RowsReturned)
	assert.Equal(t, int64(2), result.ScanCount)
	assert.Equal(t, int64(1), result.ConnectionCount)
	assert.Equal(t, int64(100), result.UncachedRowsFetched)
	assert.Equal(t, int64(50), result.CachedRowsFetched)
	assert.Equal(t, int64(60), result.HydrateCalls)
	assert.Len(t, result.Scans, 2)

	assert.Equal(t, "aws_s3_bucket", result.Scans[0].Table)
	assert.Equal(t, "aws_ec2_instance", result.Scans[1].Table)
}
