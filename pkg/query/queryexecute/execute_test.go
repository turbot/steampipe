package queryexecute

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/query"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
	"github.com/turbot/steampipe/v2/pkg/test/mocks"
)

// TestExecuteQuery_SimpleQuery tests executing a simple SELECT query
func TestExecuteQuery_SimpleQuery(t *testing.T) {
	ctx := context.Background()

	// Create a mock client
	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			// Create a result with some test data
			cols := []*pqueryresult.ColumnDef{
				{Name: "column1", DataType: "TEXT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())

			// Stream some test rows
			go func() {
				result.StreamRow([]interface{}{"value1"})
				result.Close()
			}()

			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	// Create test InitData
	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{
				ExecuteSQL: "SELECT 1",
				Name:       "test_query",
			},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	// Set output format to table (avoid snapshot logic)
	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	// Execute the query
	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 0, rowErrors)
	assert.Len(t, mockClient.ExecuteCalls, 1)
	assert.Equal(t, "SELECT 1", mockClient.ExecuteCalls[0].SQL)
}

// TestExecuteQuery_WithArgs tests executing a query with arguments
func TestExecuteQuery_WithArgs(t *testing.T) {
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			cols := []*pqueryresult.ColumnDef{
				{Name: "name", DataType: "TEXT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{"test-bucket"})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{
				ExecuteSQL: "SELECT * FROM aws_s3_bucket WHERE name = $1",
				Args:       []any{"test-bucket"},
			},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	assert.NoError(t, err)
	assert.Equal(t, 0, rowErrors)
	assert.Len(t, mockClient.ExecuteCalls, 1)
	assert.Len(t, mockClient.ExecuteCalls[0].Args, 1)
	assert.Equal(t, "test-bucket", mockClient.ExecuteCalls[0].Args[0])
}

// TestExecuteQuery_QueryError tests handling of query execution errors
func TestExecuteQuery_QueryError(t *testing.T) {
	ctx := context.Background()

	expectedError := fmt.Errorf("syntax error at or near \"SELECTT\"")
	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			return nil, expectedError
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{
				ExecuteSQL: "SELECTT * FROM table",
			},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	defer viper.Reset()

	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, 0, rowErrors)
}

// TestExecuteQueries_MultipleQueries tests executing multiple queries in batch
func TestExecuteQueries_MultipleQueries(t *testing.T) {
	ctx := context.Background()

	executeCount := 0
	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			executeCount++
			cols := []*pqueryresult.ColumnDef{
				{Name: "result", DataType: "INT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{executeCount})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT 1"},
			{ExecuteSQL: "SELECT 2"},
			{ExecuteSQL: "SELECT 3"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	failures := executeQueries(ctx, initData)

	assert.Equal(t, 0, failures)
	assert.Equal(t, 3, executeCount)
	assert.Len(t, mockClient.ExecuteCalls, 3)
}

// TestExecuteQueries_WithFailures tests handling failures in batch execution
func TestExecuteQueries_WithFailures(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			callCount++
			// Second query fails
			if callCount == 2 {
				return nil, fmt.Errorf("query failed")
			}
			cols := []*pqueryresult.ColumnDef{
				{Name: "result", DataType: "INT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{callCount})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT 1"},
			{ExecuteSQL: "INVALID QUERY"},
			{ExecuteSQL: "SELECT 3"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	failures := executeQueries(ctx, initData)

	// One failure expected
	assert.Equal(t, 1, failures)
	// All three queries should be attempted
	assert.Equal(t, 3, callCount)
}

// TestRunBatchSession_Success tests a successful batch session
func TestRunBatchSession_Success(t *testing.T) {
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			cols := []*pqueryresult.ColumnDef{
				{Name: "result", DataType: "INT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{1})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
		GetCustomSearchPathFunc: func() []string {
			return nil
		},
	}

	loadedChan := make(chan struct{})
	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT 1"},
		},
		Loaded:    loadedChan,
		StartTime: time.Now(),
	}
	initData.Client = mockClient
	initData.Result = &db_common.InitResult{} // Initialize Result
	// Signal that loading is complete
	close(loadedChan)

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	failures, err := RunBatchSession(ctx, initData)

	assert.NoError(t, err)
	assert.Equal(t, 0, failures)
}

// TestRunBatchSession_InitError tests batch session with initialization error
func TestRunBatchSession_InitError(t *testing.T) {
	ctx := context.Background()

	expectedError := fmt.Errorf("initialization failed")
	loadedChan := make(chan struct{})

	initData := &query.InitData{
		Queries:   []*modconfig.ResolvedQuery{},
		Loaded:    loadedChan,
		StartTime: time.Now(),
	}
	initData.Result = &db_common.InitResult{} // Initialize Result before using it
	initData.Result.Error = expectedError
	// Signal that loading is complete (with error)
	close(loadedChan)

	failures, err := RunBatchSession(ctx, initData)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, 0, failures)
}

// TestRunBatchSession_NoQueries tests batch session with no queries
func TestRunBatchSession_NoQueries(t *testing.T) {
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		GetCustomSearchPathFunc: func() []string {
			return nil
		},
	}

	loadedChan := make(chan struct{})
	initData := &query.InitData{
		Queries:   []*modconfig.ResolvedQuery{},
		Loaded:    loadedChan,
		StartTime: time.Now(),
	}
	initData.Client = mockClient
	initData.Result = &db_common.InitResult{} // Initialize Result
	close(loadedChan)

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	defer viper.Reset()

	failures, err := RunBatchSession(ctx, initData)

	assert.NoError(t, err)
	assert.Equal(t, 0, failures)
}

// TestNeedSnapshot tests the needSnapshot function
func TestNeedSnapshot(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		share         bool
		snapshot      bool
		exportSet     bool
		expectedResult bool
	}{
		{
			name:           "snapshot output format",
			outputFormat:   pconstants.OutputFormatSnapshot,
			share:          false,
			snapshot:       false,
			exportSet:      false,
			expectedResult: true,
		},
		{
			name:           "sps output format",
			outputFormat:   pconstants.OutputFormatSteampipeSnapshotShort,
			share:          false,
			snapshot:       false,
			exportSet:      false,
			expectedResult: true,
		},
		{
			name:           "share flag set",
			outputFormat:   pconstants.OutputFormatTable,
			share:          true,
			snapshot:       false,
			exportSet:      false,
			expectedResult: true,
		},
		{
			name:           "snapshot flag set",
			outputFormat:   pconstants.OutputFormatTable,
			share:          false,
			snapshot:       true,
			exportSet:      false,
			expectedResult: true,
		},
		{
			name:           "export set",
			outputFormat:   pconstants.OutputFormatTable,
			share:          false,
			snapshot:       false,
			exportSet:      true,
			expectedResult: true,
		},
		{
			name:           "table format no flags",
			outputFormat:   pconstants.OutputFormatTable,
			share:          false,
			snapshot:       false,
			exportSet:      false,
			expectedResult: false,
		},
		{
			name:           "json format no flags",
			outputFormat:   pconstants.OutputFormatJSON,
			share:          false,
			snapshot:       false,
			exportSet:      false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(pconstants.ArgOutput, tt.outputFormat)
			viper.Set(pconstants.ArgShare, tt.share)
			viper.Set(pconstants.ArgSnapshot, tt.snapshot)
			if tt.exportSet {
				viper.Set(pconstants.ArgExport, []string{"sps"})
			}

			result := needSnapshot()

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestShowBlankLineBetweenResults tests the showBlankLineBetweenResults function
func TestShowBlankLineBetweenResults(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   string
		header         bool
		expectedResult bool
	}{
		{
			name:           "csv without header",
			outputFormat:   "csv",
			header:         false,
			expectedResult: false,
		},
		{
			name:           "csv with header",
			outputFormat:   "csv",
			header:         true,
			expectedResult: true,
		},
		{
			name:           "table format",
			outputFormat:   "table",
			header:         true,
			expectedResult: true,
		},
		{
			name:           "json format",
			outputFormat:   "json",
			header:         true,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(pconstants.ArgOutput, tt.outputFormat)
			viper.Set(pconstants.ArgHeader, tt.header)

			result := showBlankLineBetweenResults()

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestHandlePublishSnapshotError tests error handling for snapshot publishing
func TestHandlePublishSnapshotError(t *testing.T) {
	tests := []struct {
		name          string
		inputError    error
		expectedError string
	}{
		{
			name:          "payment required error",
			inputError:    fmt.Errorf("402 Payment Required"),
			expectedError: "maximum number of snapshots reached",
		},
		{
			name:          "other error",
			inputError:    fmt.Errorf("network error"),
			expectedError: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlePublishSnapshotError(tt.inputError)

			assert.Equal(t, tt.expectedError, result.Error())
		})
	}
}

// TestExecuteQuery_ContextCancellation tests query execution with context cancellation
func TestExecuteQuery_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			// Immediately cancel the context
			cancel()
			return nil, context.Canceled
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT * FROM slow_table"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	defer viper.Reset()

	err, _ := executeQuery(ctx, initData, initData.Queries[0])

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestExecuteQuery_JSONOutput tests query execution with JSON output format
func TestExecuteQuery_JSONOutput(t *testing.T) {
	t.Skip("TODO: This test hangs - needs investigation")
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			cols := []*pqueryresult.ColumnDef{
				{Name: "name", DataType: "TEXT"},
				{Name: "value", DataType: "INT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{"test", 42})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT name, value FROM test"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	assert.NoError(t, err)
	assert.Equal(t, 0, rowErrors)
	assert.Len(t, mockClient.ExecuteCalls, 1)
}

// TestExecuteQuery_CSVOutput tests query execution with CSV output format
func TestExecuteQuery_CSVOutput(t *testing.T) {
	t.Skip("TODO: This test hangs - needs investigation")
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			cols := []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "TEXT"},
				{Name: "col2", DataType: "TEXT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				result.StreamRow([]interface{}{"value1", "value2"})
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT col1, col2 FROM test"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, constants.OutputFormatCSV)
	viper.Set(pconstants.ArgHeader, true)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	assert.NoError(t, err)
	assert.Equal(t, 0, rowErrors)
}

// TestExecuteQuery_MultipleRows tests query execution returning multiple rows
func TestExecuteQuery_MultipleRows(t *testing.T) {
	ctx := context.Background()

	mockClient := &mocks.MockClient{
		ExecuteFunc: func(ctx context.Context, sql string, args ...any) (*pqueryresult.Result[queryresult.TimingResultStream], error) {
			cols := []*pqueryresult.ColumnDef{
				{Name: "id", DataType: "INT"},
			}
			result := pqueryresult.NewResult(cols, queryresult.NewTimingResultStream())
			go func() {
				for i := 1; i <= 10; i++ {
					result.StreamRow([]interface{}{i})
				}
				result.Close()
			}()
			return result, nil
		},
		GetRequiredSessionSearchPathFunc: func() []string {
			return []string{"public"}
		},
	}

	initData := &query.InitData{
		Queries: []*modconfig.ResolvedQuery{
			{ExecuteSQL: "SELECT id FROM test LIMIT 10"},
		},
		StartTime: time.Now(),
	}
	initData.Client = mockClient

	viper.Set(pconstants.ArgOutput, pconstants.OutputFormatTable)
	viper.Set(pconstants.ArgTiming, pconstants.ArgOff)
	defer viper.Reset()

	err, rowErrors := executeQuery(ctx, initData, initData.Queries[0])

	assert.NoError(t, err)
	assert.Equal(t, 0, rowErrors)
}
