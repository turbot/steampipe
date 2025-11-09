package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestValidateQueryArgs_InteractiveModeWithSnapshot tests validation for interactive mode
func TestValidateQueryArgs_InteractiveModeWithSnapshot(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// No args = interactive mode, with snapshot flag should fail
	viper.Set(pconstants.ArgSnapshot, true)

	err := validateQueryArgs(ctx, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share snapshots in interactive mode")

	viper.Reset()
}

// TestValidateQueryArgs_InteractiveModeWithShare tests validation for interactive mode with share
func TestValidateQueryArgs_InteractiveModeWithShare(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// No args = interactive mode, with share flag should fail
	viper.Set(pconstants.ArgShare, true)

	err := validateQueryArgs(ctx, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share snapshots in interactive mode")

	viper.Reset()
}

// TestValidateQueryArgs_InteractiveModeWithExport tests validation for interactive mode with export
func TestValidateQueryArgs_InteractiveModeWithExport(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// No args = interactive mode, with export should fail
	viper.Set(pconstants.ArgExport, []string{"sps"})

	err := validateQueryArgs(ctx, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot export query results in interactive mode")

	viper.Reset()
}

// TestValidateQueryArgs_BatchModeValid tests validation for valid batch mode
func TestValidateQueryArgs_BatchModeValid(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// With args = batch mode, snapshot should be valid
	viper.Set(pconstants.ArgSnapshot, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// If not authenticated, this will fail with authentication error
	// This is expected and correct behavior - skip the test
	if err != nil && strings.Contains(err.Error(), "Not authenticated") {
		t.Skip("Skipping test - requires Turbot Pipes authentication")
	}

	// Otherwise should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_InvalidOutputFormat tests validation for invalid output format
func TestValidateQueryArgs_InvalidOutputFormat(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, "invalid-format")

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")

	viper.Reset()
}

// TestValidateQueryArgs_ValidOutputFormats tests validation for all valid output formats
func TestValidateQueryArgs_ValidOutputFormats(t *testing.T) {
	validFormats := []string{
		constants.OutputFormatLine,
		constants.OutputFormatCSV,
		constants.OutputFormatTable,
		constants.OutputFormatJSON,
		constants.OutputFormatSnapshot,
		constants.OutputFormatSnapshotShort,
		constants.OutputFormatNone,
	}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			ctx := context.Background()
			viper.Reset()

			viper.Set(pconstants.ArgOutput, format)

			err := validateQueryArgs(ctx, []string{"SELECT 1"})

			// Should not error due to output format
			assert.NoError(t, err)

			viper.Reset()
		})
	}
}

// TestValidateQueryArgs_InteractiveModeNoFlags tests interactive mode without problematic flags
func TestValidateQueryArgs_InteractiveModeNoFlags(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Interactive mode (no args) with valid output format
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{})

	// Should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_BatchModeWithExport tests batch mode with export
func TestValidateQueryArgs_BatchModeWithExport(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Batch mode with export should be valid
	viper.Set(pconstants.ArgExport, []string{"sps"})
	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// Should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_EmptyOutputFormat tests with empty output format
func TestValidateQueryArgs_EmptyOutputFormat(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Empty output format should be invalid
	viper.Set(pconstants.ArgOutput, "")

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")

	viper.Reset()
}

// TestValidateQueryArgs_MultipleQueries tests validation with multiple queries
func TestValidateQueryArgs_MultipleQueries(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{"SELECT 1", "SELECT 2", "SELECT 3"})

	// Multiple queries should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_SnapshotWithTableOutput tests snapshot flag with table output
func TestValidateQueryArgs_SnapshotWithTableOutput(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgSnapshot, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// If not authenticated, this will fail with authentication error
	// This is expected and correct behavior - skip the test
	if err != nil && strings.Contains(err.Error(), "Not authenticated") {
		t.Skip("Skipping test - requires Turbot Pipes authentication")
	}

	// Otherwise should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_ShareWithJSONOutput tests share flag with JSON output
func TestValidateQueryArgs_ShareWithJSONOutput(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgShare, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// If not authenticated, this will fail with authentication error
	// This is expected and correct behavior - skip the test
	if err != nil && strings.Contains(err.Error(), "Not authenticated") {
		t.Skip("Skipping test - requires Turbot Pipes authentication")
	}

	// Otherwise should be valid
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_OutputFormatCaseSensitivity tests output format case handling
func TestValidateQueryArgs_OutputFormatCaseSensitivity(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		shouldError bool
	}{
		{
			name:        "lowercase table",
			format:      "table",
			shouldError: false,
		},
		{
			name:        "uppercase table",
			format:      "TABLE",
			shouldError: true, // Constants are lowercase
		},
		{
			name:        "mixed case",
			format:      "Table",
			shouldError: true,
		},
		{
			name:        "lowercase json",
			format:      "json",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			viper.Reset()

			viper.Set(pconstants.ArgOutput, tt.format)

			err := validateQueryArgs(ctx, []string{"SELECT 1"})

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			viper.Reset()
		})
	}
}

// TestValidateQueryArgs_SnapshotFormatOutput tests snapshot format as output
func TestValidateQueryArgs_SnapshotFormatOutput(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatSnapshot)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_SpsFormatOutput tests sps format as output
func TestValidateQueryArgs_SpsFormatOutput(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatSnapshotShort)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_NoneFormatOutput tests none format as output
func TestValidateQueryArgs_NoneFormatOutput(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatNone)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_CSVFormat tests CSV output format
func TestValidateQueryArgs_CSVFormat(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatCSV)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_LineFormat tests line output format
func TestValidateQueryArgs_LineFormat(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatLine)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_MultipleExportFormats tests multiple export formats
func TestValidateQueryArgs_MultipleExportFormats(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgExport, []string{"sps", "json"})
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// Should be valid (validation happens elsewhere for valid export formats)
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_InteractiveModeMultipleIssues tests multiple validation issues
func TestValidateQueryArgs_InteractiveModeMultipleIssues(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Interactive mode with both snapshot and export
	viper.Set(pconstants.ArgSnapshot, true)
	viper.Set(pconstants.ArgExport, []string{"sps"})

	err := validateQueryArgs(ctx, []string{})

	// Should fail on first check (snapshot in interactive mode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share snapshots in interactive mode")

	viper.Reset()
}

// TestValidateQueryArgs_OnlyShareFlag tests only share flag without query
func TestValidateQueryArgs_OnlyShareFlag(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Share flag in interactive mode
	viper.Set(pconstants.ArgShare, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share snapshots in interactive mode")

	viper.Reset()
}

// TestValidateQueryArgs_LongQuery tests validation with a long query
func TestValidateQueryArgs_LongQuery(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Create a long query string
	longQuery := "SELECT " + string(make([]byte, 10000))
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{longQuery})

	// Should be valid (length doesn't matter for validation)
	assert.NoError(t, err)

	viper.Reset()
}

// TestValidateQueryArgs_SpecialCharactersInQuery tests validation with special characters
func TestValidateQueryArgs_SpecialCharactersInQuery(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	specialQueries := []string{
		"SELECT 'test with quotes'",
		"SELECT * FROM table WHERE name = 'O''Brien'",
		"SELECT \"column-with-dash\"",
		"SELECT /* comment */ 1",
	}

	for _, query := range specialQueries {
		err := validateQueryArgs(ctx, []string{query})
		assert.NoError(t, err)
	}

	viper.Reset()
}

// TestQueryCmd_Creation tests that queryCmd can be created without panic
func TestQueryCmd_Creation(t *testing.T) {
	assert.NotPanics(t, func() {
		cmd := queryCmd()
		assert.NotNil(t, cmd)
		assert.Equal(t, "query", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})
}

// TestQueryCmd_Flags tests that query command has expected flags
func TestQueryCmd_Flags(t *testing.T) {
	cmd := queryCmd()

	// Check for existence of important flags
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgOutput))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgTiming))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgHeader))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgSeparator))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgSearchPath))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgSnapshot))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgShare))
	assert.NotNil(t, cmd.Flags().Lookup(pconstants.ArgExport))
}

// TestQueryCmd_DefaultValues tests default flag values
func TestQueryCmd_DefaultValues(t *testing.T) {
	cmd := queryCmd()

	// Test default values
	headerFlag := cmd.Flags().Lookup(pconstants.ArgHeader)
	assert.NotNil(t, headerFlag)
	assert.Equal(t, "true", headerFlag.DefValue)

	separatorFlag := cmd.Flags().Lookup(pconstants.ArgSeparator)
	assert.NotNil(t, separatorFlag)
	assert.Equal(t, ",", separatorFlag.DefValue)

	inputFlag := cmd.Flags().Lookup(pconstants.ArgInput)
	assert.NotNil(t, inputFlag)
	assert.Equal(t, "true", inputFlag.DefValue)
}

// TestQueryCmd_AcceptsArbitraryArgs tests that query command accepts arbitrary args
func TestQueryCmd_AcceptsArbitraryArgs(t *testing.T) {
	cmd := queryCmd()

	// ArbitraryArgs should be set
	assert.NotNil(t, cmd.Args)
}

// TestValidateQueryArgs_BatchModeWithAllValidFormats tests all valid formats in batch mode
func TestValidateQueryArgs_BatchModeWithAllValidFormats(t *testing.T) {
	ctx := context.Background()

	validCombinations := []struct {
		format string
		args   []string
	}{
		{constants.OutputFormatTable, []string{"SELECT 1"}},
		{constants.OutputFormatJSON, []string{"SELECT 1"}},
		{constants.OutputFormatCSV, []string{"SELECT * FROM table"}},
		{constants.OutputFormatLine, []string{"SELECT * FROM table"}},
		{constants.OutputFormatSnapshot, []string{"SELECT 1"}},
		{constants.OutputFormatSnapshotShort, []string{"SELECT 1"}},
		{constants.OutputFormatNone, []string{"SELECT 1"}},
	}

	for _, combo := range validCombinations {
		t.Run(combo.format, func(t *testing.T) {
			viper.Reset()
			viper.Set(pconstants.ArgOutput, combo.format)

			err := validateQueryArgs(ctx, combo.args)

			assert.NoError(t, err)

			viper.Reset()
		})
	}
}

// TestValidateQueryArgs_ExportInInteractiveMode tests export in interactive mode fails
func TestValidateQueryArgs_ExportInInteractiveMode(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Empty args = interactive mode
	viper.Set(pconstants.ArgExport, []string{"sps"})
	viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)

	err := validateQueryArgs(ctx, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot export")

	viper.Reset()
}

// TestValidateQueryArgs_ValidBatchModeScenarios tests various valid batch scenarios
func TestValidateQueryArgs_ValidBatchModeScenarios(t *testing.T) {
	scenarios := []struct {
		name   string
		setup  func()
		args   []string
		requiresAuth bool
	}{
		{
			name: "single query table output",
			setup: func() {
				viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)
			},
			args: []string{"SELECT 1"},
			requiresAuth: false,
		},
		{
			name: "multiple queries json output",
			setup: func() {
				viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)
			},
			args: []string{"SELECT 1", "SELECT 2"},
			requiresAuth: false,
		},
		{
			name: "query with export",
			setup: func() {
				viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)
				viper.Set(pconstants.ArgExport, []string{"sps"})
			},
			args: []string{"SELECT 1"},
			requiresAuth: false,
		},
		{
			name: "query with snapshot",
			setup: func() {
				viper.Set(pconstants.ArgOutput, constants.OutputFormatTable)
				viper.Set(pconstants.ArgSnapshot, true)
			},
			args: []string{"SELECT 1"},
			requiresAuth: true,
		},
		{
			name: "query with share",
			setup: func() {
				viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)
				viper.Set(pconstants.ArgShare, true)
			},
			args: []string{"SELECT 1"},
			requiresAuth: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := context.Background()
			viper.Reset()
			scenario.setup()

			err := validateQueryArgs(ctx, scenario.args)

			// If this scenario requires auth and we get an auth error, skip
			if scenario.requiresAuth && err != nil && strings.Contains(err.Error(), "Not authenticated") {
				t.Skip("Skipping test - requires Turbot Pipes authentication")
			}

			assert.NoError(t, err)

			viper.Reset()
		})
	}
}

// TestValidateQueryArgs_AuthenticationRequired tests that authentication errors are properly handled
func TestValidateQueryArgs_AuthenticationRequired(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Set snapshot without authentication token
	viper.Set(pconstants.ArgSnapshot, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)
	// Ensure no token is set
	viper.Set(pconstants.ArgPipesToken, "")

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// Should get an authentication error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not authenticated")

	viper.Reset()
}

// TestValidateQueryArgs_ShareRequiresAuthentication tests that share flag requires authentication
func TestValidateQueryArgs_ShareRequiresAuthentication(t *testing.T) {
	ctx := context.Background()
	viper.Reset()

	// Set share without authentication token
	viper.Set(pconstants.ArgShare, true)
	viper.Set(pconstants.ArgOutput, constants.OutputFormatJSON)
	// Ensure no token is set
	viper.Set(pconstants.ArgPipesToken, "")

	err := validateQueryArgs(ctx, []string{"SELECT 1"})

	// Should get an authentication error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not authenticated")

	viper.Reset()
}
