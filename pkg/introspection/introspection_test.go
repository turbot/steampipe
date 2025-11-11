package introspection

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// =============================================================================
// SQL INJECTION TESTS - CRITICAL SECURITY TESTS
// =============================================================================

// TestGetSetConnectionStateSql_SQLInjection tests for SQL injection vulnerability
// BUG FOUND: The 'state' parameter is directly interpolated into SQL string
// allowing SQL injection attacks
func TestGetSetConnectionStateSql_SQLInjection(t *testing.T) {
	// t.Skip("Demonstrates bug #4748 - CRITICAL SQL injection vulnerability in GetSetConnectionStateSql. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	tests := []struct {
		name          string
		connectionName string
		state         string
		expectInSQL   string // What we expect to find if vulnerable
		shouldNotContain string // What should not be in safe SQL
	}{
		{
			name:          "SQL injection via single quote escape",
			connectionName: "test_conn",
			state:         "ready'; DROP TABLE steampipe_connection; --",
			expectInSQL:   "DROP TABLE",
			shouldNotContain: "",
		},
		{
			name:          "SQL injection via comment injection",
			connectionName: "test_conn",
			state:         "ready' OR '1'='1",
			expectInSQL:   "OR '1'='1",
			shouldNotContain: "",
		},
		{
			name:          "SQL injection via union attack",
			connectionName: "test_conn",
			state:         "ready' UNION SELECT * FROM pg_user --",
			expectInSQL:   "UNION SELECT",
			shouldNotContain: "",
		},
		{
			name:          "SQL injection via semicolon terminator",
			connectionName: "test_conn",
			state:         "ready'; DELETE FROM steampipe_connection WHERE name='victim'; --",
			expectInSQL:   "DELETE FROM",
			shouldNotContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSetConnectionStateSql(tt.connectionName, tt.state)
			require.NotEmpty(t, result, "Expected queries to be returned")

			// Check if malicious SQL is present in the generated query
			sql := result[0].Query
			if strings.Contains(sql, tt.expectInSQL) {
				t.Errorf("SQL INJECTION VULNERABILITY DETECTED!\nMalicious payload found in SQL: %s\nFull SQL: %s",
					tt.expectInSQL, sql)
			}

			// The state should be parameterized, not interpolated
			// Count the number of parameters - should be 2 ($1 for state, $2 for name)
			// But currently only has 1 ($1 for name)
			paramCount := strings.Count(sql, "$")
			if paramCount < 2 {
				t.Errorf("State parameter is not parameterized! Only found %d parameters, expected at least 2", paramCount)
			}
		})
	}
}

// TestGetConnectionStateErrorSql_ConstantUsage verifies that constants are used
// (not direct interpolation of user input)
func TestGetConnectionStateErrorSql_ConstantUsage(t *testing.T) {
	connectionName := "test_conn"
	err := errors.New("test error")

	result := GetConnectionStateErrorSql(connectionName, err)
	require.NotEmpty(t, result)

	sql := result[0].Query
	args := result[0].Args

	// Should have 2 args: error message and connection name
	assert.Len(t, args, 2, "Expected 2 parameterized arguments")
	assert.Equal(t, err.Error(), args[0], "First arg should be error message")
	assert.Equal(t, connectionName, args[1], "Second arg should be connection name")

	// The constant should be embedded (which is safe as it's not user input)
	assert.Contains(t, sql, constants.ConnectionStateError)
}

// =============================================================================
// NIL/EMPTY INPUT TESTS
// =============================================================================

func TestGetConnectionStateErrorSql_EmptyConnectionName(t *testing.T) {
	// Empty connection name should not panic
	result := GetConnectionStateErrorSql("", errors.New("test error"))
	require.NotEmpty(t, result)
	assert.Equal(t, "", result[0].Args[1])
}

func TestGetSetConnectionStateSql_EmptyInputs(t *testing.T) {
	tests := []struct {
		name          string
		connectionName string
		state         string
	}{
		{"empty connection name", "", "ready"},
		{"empty state", "test", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := GetSetConnectionStateSql(tt.connectionName, tt.state)
			require.NotEmpty(t, result)
		})
	}
}

func TestGetDeleteConnectionStateSql_EmptyName(t *testing.T) {
	result := GetDeleteConnectionStateSql("")
	require.NotEmpty(t, result)
	assert.Equal(t, "", result[0].Args[0])
}

func TestGetUpsertConnectionStateSql_NilFields(t *testing.T) {
	// Test with minimal connection state (some fields nil/empty)
	cs := &steampipeconfig.ConnectionState{
		ConnectionName: "test",
		State:         "ready",
		// Other fields left as zero values
	}

	result := GetUpsertConnectionStateSql(cs)
	require.NotEmpty(t, result)
	assert.Len(t, result[0].Args, 15)
}

func TestGetNewConnectionStateFromConnectionInsertSql_MinimalConnection(t *testing.T) {
	// Test with minimal connection
	conn := &modconfig.SteampipeConnection{
		Name:   "test",
		Plugin: "test_plugin",
	}

	result := GetNewConnectionStateFromConnectionInsertSql(conn)
	require.NotEmpty(t, result)
	assert.Len(t, result[0].Args, 14)
}

// =============================================================================
// SPECIAL CHARACTERS AND EDGE CASES
// =============================================================================

func TestGetSetConnectionStateSql_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name          string
		connectionName string
		state         string
	}{
		{"unicode in connection name", "test_😀_conn", "ready"},
		{"quotes in connection name", "test'conn\"name", "ready"},
		{"newlines in connection name", "test\nconn", "ready"},
		{"backslashes", "test\\conn\\name", "ready"},
		{"null bytes (truncated by Go)", "test\x00conn", "ready"},
		{"very long connection name", strings.Repeat("a", 10000), "ready"},
		{"state with newlines", "test", "ready\nmalicious"},
		{"state with quotes", "test", "ready'\"state"},
		{"state with backslashes", "test", "ready\\state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := GetSetConnectionStateSql(tt.connectionName, tt.state)
			require.NotEmpty(t, result)

			// Verify the connection name is parameterized (in args, not query string)
			sql := result[0].Query
			assert.NotContains(t, sql, tt.connectionName,
				"Connection name should be parameterized, not in SQL string")
		})
	}
}

func TestGetConnectionStateErrorSql_SpecialCharactersInError(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
	}{
		{"quotes in error", "error with 'quotes' and \"double quotes\""},
		{"newlines in error", "error\nwith\nnewlines"},
		{"unicode in error", "error with 😀 emoji"},
		{"very long error", strings.Repeat("error ", 10000)},
		{"null bytes", "error\x00with\x00nulls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConnectionStateErrorSql("test", errors.New(tt.errMsg))
			require.NotEmpty(t, result)

			// Error message should be parameterized
			assert.Equal(t, tt.errMsg, result[0].Args[0])
		})
	}
}

func TestGetDeleteConnectionStateSql_SpecialCharacters(t *testing.T) {
	maliciousNames := []string{
		"'; DROP TABLE connections; --",
		"test' OR '1'='1",
		"test\"; DELETE FROM connections; --",
		strings.Repeat("a", 10000),
	}

	for _, name := range maliciousNames {
		result := GetDeleteConnectionStateSql(name)
		require.NotEmpty(t, result)

		// Name should be in args, not in SQL string
		assert.Equal(t, name, result[0].Args[0])
		assert.NotContains(t, result[0].Query, name,
			"Malicious name should be parameterized")
	}
}

// =============================================================================
// PLUGIN TABLE SQL TESTS
// =============================================================================

func TestGetPluginTableCreateSql_ValidSQL(t *testing.T) {
	result := GetPluginTableCreateSql()

	// Basic validation
	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "CREATE TABLE IF NOT EXISTS")
	assert.Contains(t, result.Query, constants.InternalSchema)
	assert.Contains(t, result.Query, constants.PluginInstanceTable)

	// Check for proper column definitions
	assert.Contains(t, result.Query, "plugin_instance TEXT")
	assert.Contains(t, result.Query, "plugin TEXT NOT NULL")
	assert.Contains(t, result.Query, "version TEXT")
}

func TestGetPluginTablePopulateSql_AllFields(t *testing.T) {
	memoryMaxMb := 512
	fileName := "/path/to/plugin.spc"
	startLine := 10
	endLine := 20

	p := &plugin.Plugin{
		Plugin:   "test_plugin",
		Version:  "1.0.0",
		Instance: "test_instance",
		MemoryMaxMb: &memoryMaxMb,
		FileName: &fileName,
		StartLineNumber: &startLine,
		EndLineNumber: &endLine,
	}

	result := GetPluginTablePopulateSql(p)

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "INSERT INTO")
	assert.Len(t, result.Args, 8)
	assert.Equal(t, p.Plugin, result.Args[0])
	assert.Equal(t, p.Version, result.Args[1])
}

func TestGetPluginTablePopulateSql_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name   string
		plugin *plugin.Plugin
	}{
		{
			"quotes in plugin name",
			&plugin.Plugin{
				Plugin: "test'plugin\"name",
				Version: "1.0.0",
			},
		},
		{
			"very long version string",
			&plugin.Plugin{
				Plugin: "test",
				Version: strings.Repeat("1.0.", 1000),
			},
		},
		{
			"unicode in fields",
			&plugin.Plugin{
				Plugin: "test_😀",
				Version: "v1.0.0-beta",
				Instance: "instance_with_特殊字符",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := GetPluginTablePopulateSql(tt.plugin)
			assert.NotEmpty(t, result.Query)
			assert.NotEmpty(t, result.Args)
		})
	}
}

func TestGetPluginTableDropSql_ValidSQL(t *testing.T) {
	result := GetPluginTableDropSql()

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "DROP TABLE IF EXISTS")
	assert.Contains(t, result.Query, constants.InternalSchema)
	assert.Contains(t, result.Query, constants.PluginInstanceTable)
}

func TestGetPluginTableGrantSql_ValidSQL(t *testing.T) {
	result := GetPluginTableGrantSql()

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "GRANT SELECT ON TABLE")
	assert.Contains(t, result.Query, constants.DatabaseUsersRole)
}

// =============================================================================
// PLUGIN COLUMN TABLE SQL TESTS
// =============================================================================

func TestGetPluginColumnTableCreateSql_ValidSQL(t *testing.T) {
	result := GetPluginColumnTableCreateSql()

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "CREATE TABLE IF NOT EXISTS")
	assert.Contains(t, result.Query, "plugin TEXT NOT NULL")
	assert.Contains(t, result.Query, "table_name TEXT NOT NULL")
	assert.Contains(t, result.Query, "name TEXT NOT NULL")
}

func TestGetPluginColumnTablePopulateSql_AllFieldTypes(t *testing.T) {
	tests := []struct {
		name         string
		columnSchema *proto.ColumnDefinition
		expectError  bool
	}{
		{
			"basic column",
			&proto.ColumnDefinition{
				Name:        "test_col",
				Type:        proto.ColumnType_STRING,
				Description: "test description",
			},
			false,
		},
		{
			"column with quotes in description",
			&proto.ColumnDefinition{
				Name:        "test_col",
				Type:        proto.ColumnType_STRING,
				Description: "description with 'quotes' and \"double quotes\"",
			},
			false,
		},
		{
			"column with unicode",
			&proto.ColumnDefinition{
				Name:        "test_😀_col",
				Type:        proto.ColumnType_STRING,
				Description: "Unicode: 你好 мир",
			},
			false,
		},
		{
			"column with very long description",
			&proto.ColumnDefinition{
				Name:        "test_col",
				Type:        proto.ColumnType_STRING,
				Description: strings.Repeat("Very long description. ", 1000),
			},
			false,
		},
		{
			"empty column name",
			&proto.ColumnDefinition{
				Name: "",
				Type: proto.ColumnType_STRING,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetPluginColumnTablePopulateSql(
				"test_plugin",
				"test_table",
				tt.columnSchema,
				nil,
				nil,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.Query)
				assert.Contains(t, result.Query, "INSERT INTO")
			}
		})
	}
}

func TestGetPluginColumnTablePopulateSql_SQLInjectionAttempts(t *testing.T) {
	maliciousInputs := []struct {
		name      string
		pluginName string
		tableName  string
		columnName string
	}{
		{
			"malicious plugin name",
			"plugin'; DROP TABLE steampipe_plugin_column; --",
			"table",
			"column",
		},
		{
			"malicious table name",
			"plugin",
			"table'; DELETE FROM steampipe_plugin_column; --",
			"column",
		},
		{
			"malicious column name",
			"plugin",
			"table",
			"col' OR '1'='1",
		},
	}

	for _, tt := range maliciousInputs {
		t.Run(tt.name, func(t *testing.T) {
			columnSchema := &proto.ColumnDefinition{
				Name: tt.columnName,
				Type: proto.ColumnType_STRING,
			}

			result, err := GetPluginColumnTablePopulateSql(
				tt.pluginName,
				tt.tableName,
				columnSchema,
				nil,
				nil,
			)

			require.NoError(t, err)

			// All inputs should be parameterized
			sql := result.Query
			assert.NotContains(t, sql, "DROP TABLE", "SQL injection detected!")
			assert.NotContains(t, sql, "DELETE FROM", "SQL injection detected!")

			// Verify inputs are in args, not in SQL string
			assert.Equal(t, tt.pluginName, result.Args[0])
			assert.Equal(t, tt.tableName, result.Args[1])
			assert.Equal(t, tt.columnName, result.Args[2])
		})
	}
}

func TestGetPluginColumnTableDeletePluginSql_SpecialCharacters(t *testing.T) {
	maliciousPlugins := []string{
		"plugin'; DROP TABLE steampipe_plugin_column; --",
		"plugin' OR '1'='1",
		strings.Repeat("p", 10000),
	}

	for _, plugin := range maliciousPlugins {
		result := GetPluginColumnTableDeletePluginSql(plugin)

		assert.NotEmpty(t, result.Query)
		assert.Contains(t, result.Query, "DELETE FROM")
		assert.Equal(t, plugin, result.Args[0], "Plugin name should be parameterized")
		assert.NotContains(t, result.Query, plugin, "Plugin name should not be in SQL string")
	}
}

// =============================================================================
// RATE LIMITER TABLE SQL TESTS
// =============================================================================

func TestGetRateLimiterTableCreateSql_ValidSQL(t *testing.T) {
	result := GetRateLimiterTableCreateSql()

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "CREATE TABLE IF NOT EXISTS")
	assert.Contains(t, result.Query, constants.InternalSchema)
	assert.Contains(t, result.Query, constants.RateLimiterDefinitionTable)
	assert.Contains(t, result.Query, "name TEXT")
	assert.Contains(t, result.Query, "\"where\" TEXT") // 'where' is a SQL keyword, should be quoted
}

func TestGetRateLimiterTablePopulateSql_AllFields(t *testing.T) {
	bucketSize := int64(100)
	fillRate := float32(10.5)
	maxConcurrency := int64(5)
	where := "some condition"
	fileName := "/path/to/file.spc"
	startLine := 1
	endLine := 10

	rl := &plugin.RateLimiter{
		Name:           "test_limiter",
		Plugin:         "test_plugin",
		PluginInstance: "test_instance",
		Source:         "config",
		Status:         "active",
		BucketSize:     &bucketSize,
		FillRate:       &fillRate,
		MaxConcurrency: &maxConcurrency,
		Where:          &where,
		FileName:       &fileName,
		StartLineNumber: &startLine,
		EndLineNumber:   &endLine,
	}

	result := GetRateLimiterTablePopulateSql(rl)

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "INSERT INTO")
	assert.Len(t, result.Args, 13)
	assert.Equal(t, rl.Name, result.Args[0])
	assert.Equal(t, rl.FillRate, result.Args[6])
}

func TestGetRateLimiterTablePopulateSql_SQLInjection(t *testing.T) {
	tests := []struct {
		name string
		rl   *plugin.RateLimiter
	}{
		{
			"malicious name",
			&plugin.RateLimiter{
				Name:   "limiter'; DROP TABLE steampipe_rate_limiter; --",
				Plugin: "plugin",
			},
		},
		{
			"malicious plugin",
			&plugin.RateLimiter{
				Name:   "limiter",
				Plugin: "plugin' OR '1'='1",
			},
		},
		{
			"malicious where clause",
			func() *plugin.RateLimiter {
				where := "'; DELETE FROM steampipe_rate_limiter; --"
				return &plugin.RateLimiter{
					Name:   "limiter",
					Plugin: "plugin",
					Where:  &where,
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRateLimiterTablePopulateSql(tt.rl)

			sql := result.Query
			// Verify no SQL injection keywords are in the generated SQL
			assert.NotContains(t, sql, "DROP TABLE", "SQL injection detected!")
			assert.NotContains(t, sql, "DELETE FROM", "SQL injection detected!")

			// All fields should be parameterized (not in SQL string directly)
			// The malicious parts should not be in the SQL
			if strings.Contains(tt.rl.Name, "DROP TABLE") {
				assert.NotContains(t, sql, "limiter'; DROP TABLE", "Name should be parameterized")
			}
			if strings.Contains(tt.rl.Plugin, "OR '1'='1") {
				assert.NotContains(t, sql, "OR '1'='1", "Plugin should be parameterized")
			}
			if tt.rl.Where != nil && strings.Contains(*tt.rl.Where, "DELETE FROM") {
				assert.NotContains(t, sql, "DELETE FROM", "Where should be parameterized")
			}
		})
	}
}

func TestGetRateLimiterTablePopulateSql_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
		rl   *plugin.RateLimiter
	}{
		{
			"unicode in name",
			&plugin.RateLimiter{
				Name:   "limiter_😀_test",
				Plugin: "plugin",
			},
		},
		{
			"quotes in fields",
			func() *plugin.RateLimiter {
				where := "condition with 'quotes'"
				return &plugin.RateLimiter{
					Name:   "test'limiter\"name",
					Plugin: "plugin'test",
					Where:  &where,
				}
			}(),
		},
		{
			"very long fields",
			func() *plugin.RateLimiter {
				where := strings.Repeat("condition ", 1000)
				return &plugin.RateLimiter{
					Name:   strings.Repeat("a", 10000),
					Plugin: "plugin",
					Where:  &where,
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := GetRateLimiterTablePopulateSql(tt.rl)
			assert.NotEmpty(t, result.Query)
			assert.NotEmpty(t, result.Args)
		})
	}
}

func TestGetRateLimiterTableGrantSql_ValidSQL(t *testing.T) {
	result := GetRateLimiterTableGrantSql()

	assert.NotEmpty(t, result.Query)
	assert.Contains(t, result.Query, "GRANT SELECT ON TABLE")
	assert.Contains(t, result.Query, constants.DatabaseUsersRole)
}

// =============================================================================
// HELPER FUNCTION TESTS
// =============================================================================

func TestGetConnectionStateQueries_ReturnsMultipleQueries(t *testing.T) {
	queryFormat := "SELECT * FROM %s.%s WHERE name=$1"
	args := []any{"test_conn"}

	result := getConnectionStateQueries(queryFormat, args)

	// Should return 2 queries (one for new table, one for legacy)
	assert.Len(t, result, 2)

	// Both should have the same args
	assert.Equal(t, args, result[0].Args)
	assert.Equal(t, args, result[1].Args)

	// Queries should reference different tables
	assert.Contains(t, result[0].Query, constants.ConnectionTable)
	assert.Contains(t, result[1].Query, constants.LegacyConnectionStateTable)
}

// =============================================================================
// EDGE CASE: VERY LONG IDENTIFIERS
// =============================================================================

func TestVeryLongIdentifiers(t *testing.T) {
	longName := strings.Repeat("a", 10000)

	t.Run("very long connection name", func(t *testing.T) {
		result := GetSetConnectionStateSql(longName, "ready")
		require.NotEmpty(t, result)
		// Should be in args, not cause buffer issues
		// Args order: state (args[0]), connectionName (args[1])
		assert.Equal(t, longName, result[0].Args[1])
	})

	t.Run("very long state", func(t *testing.T) {
		result := GetSetConnectionStateSql("test", longName)
		require.NotEmpty(t, result)
		// Note: This will expose the injection vulnerability if state is in SQL string
	})
}
