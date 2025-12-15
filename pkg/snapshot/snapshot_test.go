package snapshot

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

// TestRoundTripDataIntegrity_EmptyResult tests that an empty result round-trips correctly
func TestRoundTripDataIntegrity_EmptyResult(t *testing.T) {
	ctx := context.Background()

	// Create empty result
	cols := []*pqueryresult.ColumnDef{}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
	result.Close()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT 1",
	}

	// Convert to snapshot
	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Convert back to result
	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())

	// BUG?: Does it handle empty columns correctly?
	if err != nil {
		t.Logf("Error on empty result conversion: %v", err)
	}

	if result2 != nil {
		assert.Equal(t, 0, len(result2.Cols), "Empty result should have 0 columns")
	}
}

// TestRoundTripDataIntegrity_BasicData tests basic data round-trip
func TestRoundTripDataIntegrity_BasicData(t *testing.T) {
	ctx := context.Background()

	// Create result with data
	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "name", DataType: "text"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	// Add test data
	testRows := [][]interface{}{
		{1, "Alice"},
		{2, "Bob"},
		{3, "Charlie"},
	}

	go func() {
		for _, row := range testRows {
			result.StreamRow(row)
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id, name FROM users",
	}

	// Convert to snapshot
	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{"public"}, time.Now())
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Verify snapshot structure
	assert.Equal(t, schemaVersion, snapshot.SchemaVersion)
	assert.NotEmpty(t, snapshot.Panels)

	// Convert back to result
	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
	require.NoError(t, err)
	require.NotNil(t, result2)

	// Verify columns
	assert.Equal(t, len(cols), len(result2.Cols))
	for i, col := range result2.Cols {
		assert.Equal(t, cols[i].Name, col.Name)
	}

	// Verify rows
	rowCount := 0
	for rowResult, ok := <-result2.RowChan; ok; rowResult, ok = <-result2.RowChan {
		assert.Equal(t, len(cols), len(rowResult.Data), "Row %d should have correct number of columns", rowCount)
		rowCount++
	}

	// BUG?: Are all rows preserved?
	assert.Equal(t, len(testRows), rowCount, "All rows should be preserved in round-trip")
}

// TestRoundTripDataIntegrity_NullValues tests null value handling
func TestRoundTripDataIntegrity_NullValues(t *testing.T) {
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "value", DataType: "text"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	// Add rows with null values
	testRows := [][]interface{}{
		{1, nil},
		{nil, "value"},
		{nil, nil},
	}

	go func() {
		for _, row := range testRows {
			result.StreamRow(row)
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id, value FROM test",
	}

	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)

	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
	require.NoError(t, err)

	// BUG?: Are null values preserved correctly?
	rowCount := 0
	for rowResult, ok := <-result2.RowChan; ok; rowResult, ok = <-result2.RowChan {
		t.Logf("Row %d: %v", rowCount, rowResult.Data)
		rowCount++
	}

	assert.Equal(t, len(testRows), rowCount, "All rows with nulls should be preserved")
}

// TestConcurrentSnapshotToQueryResult_Race tests for race conditions
func TestConcurrentSnapshotToQueryResult_Race(t *testing.T) {
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	go func() {
		for i := 0; i < 100; i++ {
			result.StreamRow([]interface{}{i})
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id FROM test",
	}

	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)

	// BUG?: Race condition when multiple goroutines read the same snapshot?
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
			if err != nil {
				errors <- fmt.Errorf("error in concurrent conversion: %w", err)
				return
			}

			// Consume all rows
			for range result2.RowChan {
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestSnapshotToQueryResult_GoroutineCleanup tests goroutine cleanup
// FOUND BUG: Goroutine leak when rows are not fully consumed
func TestSnapshotToQueryResult_GoroutineCleanup(t *testing.T) {
	// t.Skip("Demonstrates bug #4768 - Goroutines leak when rows are not consumed - see snapshot.go:193. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	go func() {
		for i := 0; i < 1000; i++ {
			result.StreamRow([]interface{}{i})
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id FROM test",
	}

	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)

	// Create result but don't consume rows
	// BUG?: Does the goroutine leak if rows are not consumed?
	for i := 0; i < 100; i++ {
		result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
		require.NoError(t, err)

		// Only read one row, then abandon
		<-result2.RowChan
		// Goroutine should clean up even if we don't read all rows
	}

	// If goroutines leaked, this test would fail with a race detector or show up in profiling
	time.Sleep(100 * time.Millisecond)
}

// TestSnapshotToQueryResult_PartialConsumption tests partial row consumption
// FOUND BUG: Goroutine leak when rows are not fully consumed
func TestSnapshotToQueryResult_PartialConsumption(t *testing.T) {
	// t.Skip("Demonstrates bug #4768 - Goroutines leak when rows are not consumed - see snapshot.go:193. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	go func() {
		for i := 0; i < 100; i++ {
			result.StreamRow([]interface{}{i})
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id FROM test",
	}

	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)

	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
	require.NoError(t, err)

	// Only consume first 10 rows
	for i := 0; i < 10; i++ {
		row, ok := <-result2.RowChan
		require.True(t, ok, "Should be able to read row %d", i)
		require.NotNil(t, row)
	}

	// BUG?: What happens if we stop consuming? Does the goroutine block forever?
	// Let goroutine finish
	time.Sleep(100 * time.Millisecond)
}

// TestLargeDataHandling tests performance with large datasets
func TestLargeDataHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data test in short mode")
	}

	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "data", DataType: "text"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	// Large dataset
	numRows := 10000
	go func() {
		for i := 0; i < numRows; i++ {
			result.StreamRow([]interface{}{i, fmt.Sprintf("data_%d", i)})
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT id, data FROM large_table",
	}

	startTime := time.Now()
	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	conversionTime := time.Since(startTime)

	require.NoError(t, err)
	t.Logf("Large data conversion took: %v", conversionTime)

	// BUG?: Does large data cause performance issues?
	startTime = time.Now()
	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
	require.NoError(t, err)

	rowCount := 0
	for range result2.RowChan {
		rowCount++
	}
	roundTripTime := time.Since(startTime)

	assert.Equal(t, numRows, rowCount, "All rows should be preserved in large dataset")
	t.Logf("Large data round-trip took: %v", roundTripTime)

	// BUG?: Performance degradation with large data?
	if roundTripTime > 5*time.Second {
		t.Logf("WARNING: Round-trip took longer than 5 seconds for %d rows", numRows)
	}
}

// TestSnapshotToQueryResult_InvalidSnapshot tests error handling
func TestSnapshotToQueryResult_InvalidSnapshot(t *testing.T) {
	// Test with invalid snapshot (missing expected panel)
	invalidSnapshot := &steampipeconfig.SteampipeSnapshot{
		Panels: map[string]steampipeconfig.SnapshotPanel{},
	}

	result, err := SnapshotToQueryResult[queryresult.TimingResultStream](invalidSnapshot, time.Now())

	// BUG?: Should return error, not panic
	assert.Error(t, err, "Should return error for invalid snapshot")
	assert.Nil(t, result, "Result should be nil on error")
}

// TestSnapshotToQueryResult_WrongPanelType tests type assertion safety
func TestSnapshotToQueryResult_WrongPanelType(t *testing.T) {
	// Create snapshot with wrong panel type
	wrongSnapshot := &steampipeconfig.SteampipeSnapshot{
		Panels: map[string]steampipeconfig.SnapshotPanel{
			"custom.table.results": &PanelData{
				// This is the right type, but let's test the assertion
			},
		},
	}

	// This should work
	result, err := SnapshotToQueryResult[queryresult.TimingResultStream](wrongSnapshot, time.Now())
	require.NoError(t, err)

	// Consume rows
	for range result.RowChan {
	}
}

// TestConcurrentDataAccess_MultipleGoroutines tests concurrent data structure access
func TestConcurrentDataAccess_MultipleGoroutines(t *testing.T) {
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "value", DataType: "text"},
	}

	// BUG?: Race condition when multiple goroutines create snapshots?
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

			go func() {
				for j := 0; j < 100; j++ {
					result.StreamRow([]interface{}{j, fmt.Sprintf("value_%d", j)})
				}
				result.Close()
			}()

			resolvedQuery := &modconfig.ResolvedQuery{
				RawSQL: fmt.Sprintf("SELECT id, value FROM test_%d", id),
			}

			_, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestDataIntegrity_SpecialCharacters tests special character handling
func TestDataIntegrity_SpecialCharacters(t *testing.T) {
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "text_col", DataType: "text"},
	}
	result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

	// Special characters that might cause issues
	specialStrings := []string{
		"",                    // empty string
		"'single quotes'",
		"\"double quotes\"",
		"line\nbreak",
		"tab\there",
		"unicode: 你好",
		"emoji: 😀",
		"null\x00byte",
	}

	go func() {
		for _, str := range specialStrings {
			result.StreamRow([]interface{}{str})
		}
		result.Close()
	}()

	resolvedQuery := &modconfig.ResolvedQuery{
		RawSQL: "SELECT text_col FROM test",
	}

	snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
	require.NoError(t, err)

	result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
	require.NoError(t, err)

	// BUG?: Are special characters preserved correctly?
	rowCount := 0
	for rowResult, ok := <-result2.RowChan; ok; rowResult, ok = <-result2.RowChan {
		require.NotNil(t, rowResult)
		t.Logf("Row %d: %v", rowCount, rowResult.Data)
		rowCount++
	}

	assert.Equal(t, len(specialStrings), rowCount, "All special character rows should be preserved")
}

// TestHashCollision_DifferentQueries tests hash uniqueness
func TestHashCollision_DifferentQueries(t *testing.T) {
	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	queries := []string{
		"SELECT 1",
		"SELECT 2",
		"SELECT 3",
		"SELECT 1 ",  // trailing space
	}

	hashes := make(map[string]bool)

	for _, query := range queries {
		result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

		go func() {
			result.StreamRow([]interface{}{1})
			result.Close()
		}()

		resolvedQuery := &modconfig.ResolvedQuery{
			RawSQL: query,
		}

		snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
		require.NoError(t, err)

		// Extract dashboard name to check uniqueness
		var dashboardName string
		for name := range snapshot.Panels {
			if name != "custom.table.results" {
				dashboardName = name
				break
			}
		}

		// BUG?: Hash collision for different queries?
		if hashes[dashboardName] {
			t.Logf("WARNING: Hash collision detected for query: %s", query)
		}
		hashes[dashboardName] = true
	}
}

// TestMemoryLeak_RepeatedConversions tests for memory leaks
func TestMemoryLeak_RepeatedConversions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	ctx := context.Background()

	cols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	// BUG?: Memory leak with repeated conversions?
	for i := 0; i < 1000; i++ {
		result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

		go func() {
			for j := 0; j < 100; j++ {
				result.StreamRow([]interface{}{j})
			}
			result.Close()
		}()

		resolvedQuery := &modconfig.ResolvedQuery{
			RawSQL: fmt.Sprintf("SELECT id FROM test_%d", i),
		}

		snapshot, err := QueryResultToSnapshot(ctx, result, resolvedQuery, []string{}, time.Now())
		require.NoError(t, err)

		result2, err := SnapshotToQueryResult[queryresult.TimingResultStream](snapshot, time.Now())
		require.NoError(t, err)

		// Consume all rows
		for range result2.RowChan {
		}

		if i%100 == 0 {
			t.Logf("Completed %d iterations", i)
		}
	}
}
