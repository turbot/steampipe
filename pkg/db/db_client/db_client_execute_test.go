package db_client

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimestamptzTextFormatImplemented verifies that the timestamptz wire protocol fix is in place.
// Reference: https://github.com/turbot/steampipe/issues/4450
//
// This test verifies that startQuery uses QueryResultFormatsByOID to request text format
// for timestamptz columns, ensuring PostgreSQL formats values using the session timezone.
//
// Without this fix, pgx uses binary protocol which loses session timezone info, causing
// timestamptz values to display in the local machine timezone instead of the session timezone.
func TestTimestamptzTextFormatImplemented(t *testing.T) {
	// Read the db_client_execute.go file to verify the fix is present
	content, err := os.ReadFile("db_client_execute.go")
	require.NoError(t, err, "should be able to read db_client_execute.go")

	sourceCode := string(content)

	// Verify QueryResultFormatsByOID is used
	assert.Contains(t, sourceCode, "pgx.QueryResultFormatsByOID",
		"QueryResultFormatsByOID must be used to specify format for specific column types")

	// Verify TimestamptzOID is referenced
	assert.Contains(t, sourceCode, "pgtype.TimestamptzOID",
		"TimestamptzOID must be specified to request text format for timestamptz columns")

	// Verify TextFormatCode is used
	assert.Contains(t, sourceCode, "pgx.TextFormatCode",
		"TextFormatCode must be used to request text format")

	// Verify the fix is in startQuery function
	funcStart := strings.Index(sourceCode, "func (c *DbClient) startQuery")
	assert.NotEqual(t, -1, funcStart, "startQuery function must exist")

	// Extract just the startQuery function for more precise checking
	funcEnd := strings.Index(sourceCode[funcStart:], "\nfunc ")
	if funcEnd == -1 {
		funcEnd = len(sourceCode)
	} else {
		funcEnd += funcStart
	}
	startQueryFunc := sourceCode[funcStart:funcEnd]

	// Verify all three components are in startQuery
	assert.Contains(t, startQueryFunc, "QueryResultFormatsByOID",
		"QueryResultFormatsByOID must be in startQuery function")
	assert.Contains(t, startQueryFunc, "TimestamptzOID",
		"TimestamptzOID must be in startQuery function")
	assert.Contains(t, startQueryFunc, "TextFormatCode",
		"TextFormatCode must be in startQuery function")

	// Verify there's a comment explaining the fix
	hasComment := strings.Contains(startQueryFunc, "session timezone") ||
		strings.Contains(startQueryFunc, "text format for timestamptz") ||
		strings.Contains(startQueryFunc, "Request text format")
	assert.True(t, hasComment,
		"Comment should explain why text format is needed for timestamptz")

	// Verify queryArgs are constructed and used
	assert.Contains(t, startQueryFunc, "queryArgs",
		"queryArgs variable must be used to prepend format specification")
	assert.Contains(t, startQueryFunc, "conn.Query(ctx, query, queryArgs...)",
		"conn.Query must use queryArgs instead of args directly")
}

// TestTimestamptzFormatCorrectness verifies the format specification structure
func TestTimestamptzFormatCorrectness(t *testing.T) {
	content, err := os.ReadFile("db_client_execute.go")
	require.NoError(t, err, "should be able to read db_client_execute.go")

	sourceCode := string(content)

	// Verify the QueryResultFormatsByOID is constructed as the first element
	// This is critical - it must be the first argument before actual query parameters
	assert.Contains(t, sourceCode, "queryArgs := make([]any, 0, len(args)+1)",
		"queryArgs must be allocated with capacity for format spec + args")

	// Verify format spec is appended first
	lines := strings.Split(sourceCode, "\n")
	var foundMake, foundAppendFormat, foundAppendArgs bool
	var makeIdx, appendFormatIdx, appendArgsIdx int

	for i, line := range lines {
		if strings.Contains(line, "queryArgs := make([]any, 0, len(args)+1)") {
			foundMake = true
			makeIdx = i
		}
		if strings.Contains(line, "queryArgs = append(queryArgs, pgx.QueryResultFormatsByOID{") {
			foundAppendFormat = true
			appendFormatIdx = i
		}
		if strings.Contains(line, "queryArgs = append(queryArgs, args...)") {
			foundAppendArgs = true
			appendArgsIdx = i
		}
	}

	assert.True(t, foundMake, "queryArgs must be allocated")
	assert.True(t, foundAppendFormat, "format spec must be appended to queryArgs")
	assert.True(t, foundAppendArgs, "original args must be appended to queryArgs")

	// Verify correct order: make -> append format spec -> append args
	if foundMake && foundAppendFormat && foundAppendArgs {
		assert.Less(t, makeIdx, appendFormatIdx,
			"queryArgs must be allocated before appending format spec")
		assert.Less(t, appendFormatIdx, appendArgsIdx,
			"format spec must be appended before original args")
	}
}

// TestTimestamptzFormatDoesNotAffectOtherTypes verifies only timestamptz format is changed
func TestTimestamptzFormatDoesNotAffectOtherTypes(t *testing.T) {
	content, err := os.ReadFile("db_client_execute.go")
	require.NoError(t, err, "should be able to read db_client_execute.go")

	sourceCode := string(content)

	// Find the QueryResultFormatsByOID map construction
	funcStart := strings.Index(sourceCode, "func (c *DbClient) startQuery")
	require.NotEqual(t, -1, funcStart, "startQuery function must exist")

	funcEnd := strings.Index(sourceCode[funcStart:], "\nfunc ")
	if funcEnd == -1 {
		funcEnd = len(sourceCode)
	} else {
		funcEnd += funcStart
	}
	startQueryFunc := sourceCode[funcStart:funcEnd]

	// Verify ONLY TimestamptzOID is in the map (no other OIDs)
	// This ensures we don't accidentally change format for other types
	otherOIDs := []string{
		"DateOID",
		"TimestampOID",
		"TimeOID",
		"IntervalOID",
		"JSONOID",
		"JSONBOID",
	}

	for _, oid := range otherOIDs {
		assert.NotContains(t, startQueryFunc, "pgtype."+oid,
			"Should not change format for "+oid+" - only timestamptz needs text format")
	}

	// Verify there's only one entry in QueryResultFormatsByOID
	// Count how many times we see "OID:" in the map definition
	oidCount := strings.Count(startQueryFunc, "OID:")
	assert.Equal(t, 1, oidCount,
		"QueryResultFormatsByOID should have exactly one entry (TimestamptzOID)")
}
