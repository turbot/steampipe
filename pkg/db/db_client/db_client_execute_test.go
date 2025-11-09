package db_client

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
)

// TestPopulateRow tests the populateRow function
func TestPopulateRow(t *testing.T) {
	tests := map[string]struct {
		columnValues []interface{}
		cols         []*pqueryresult.ColumnDef
		expected     []interface{}
		expectError  bool
	}{
		"nil values": {
			columnValues: []interface{}{nil, nil, nil},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "TEXT"},
				{Name: "col2", DataType: "INTEGER"},
				{Name: "col3", DataType: "BOOLEAN"},
			},
			expected:    []interface{}{nil, nil, nil},
			expectError: false,
		},
		"simple string values": {
			columnValues: []interface{}{"test1", "test2", "test3"},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "TEXT"},
				{Name: "col2", DataType: "TEXT"},
				{Name: "col3", DataType: "TEXT"},
			},
			expected:    []interface{}{"test1", "test2", "test3"},
			expectError: false,
		},
		"mixed types": {
			columnValues: []interface{}{"text", int64(123), true},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "TEXT"},
				{Name: "col2", DataType: "BIGINT"},
				{Name: "col3", DataType: "BOOLEAN"},
			},
			expected:    []interface{}{"text", int64(123), true},
			expectError: false,
		},
		"text array": {
			columnValues: []interface{}{
				[]interface{}{"item1", "item2", "item3"},
			},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "_TEXT"},
			},
			expected:    []interface{}{"item1,item2,item3"},
			expectError: false,
		},
		"numeric type": {
			columnValues: []interface{}{
				pgtype.Numeric{Int: nil, Exp: 0, NaN: false, InfinityModifier: pgtype.Finite, Valid: true},
			},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "NUMERIC"},
			},
			expected:    []interface{}{float64(0)},
			expectError: false,
		},
		"time type": {
			columnValues: []interface{}{
				pgtype.Time{Microseconds: 36000000000, Valid: true}, // 10:00:00
			},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "TIME"},
			},
			expected:    []interface{}{"10:00:00"},
			expectError: false,
		},
		"interval type - days only": {
			columnValues: []interface{}{
				pgtype.Interval{Months: 0, Days: 5, Microseconds: 0, Valid: true},
			},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "INTERVAL"},
			},
			expected:    []interface{}{"5 days "},
			expectError: false,
		},
		"interval type - years and months": {
			columnValues: []interface{}{
				pgtype.Interval{Months: 14, Days: 0, Microseconds: 0, Valid: true}, // 1 year, 2 months
			},
			cols: []*pqueryresult.ColumnDef{
				{Name: "col1", DataType: "INTERVAL"},
			},
			expected:    []interface{}{"1 year 2 mons "},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := populateRow(tc.columnValues, tc.cols)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expected), len(result))
				for i, expected := range tc.expected {
					if expected == nil {
						assert.Nil(t, result[i])
					} else {
						assert.NotNil(t, result[i])
					}
				}
			}
		})
	}
}

// TestExecuteInSessionNilSession tests ExecuteInSession with nil session
func TestExecuteInSessionNilSession(t *testing.T) {
	client := &DbClient{}
	ctx := context.Background()

	result, err := client.ExecuteInSession(ctx, nil, nil, "SELECT 1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "nil session")
}
