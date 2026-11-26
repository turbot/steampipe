package db_common

import (
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

type schemaIterationRows struct {
	pgx.Rows
	iterated bool
	err      error
}

func (r *schemaIterationRows) Next() bool {
	if r.iterated {
		return false
	}
	r.iterated = true
	return true
}

func (r *schemaIterationRows) Scan(dest ...any) error {
	values := []string{"example_table", "id", "", "NO", "integer", "int4", "public", "", ""}
	for i, value := range values {
		*dest[i].(*[]byte) = []byte(value)
	}
	return nil
}

func (r *schemaIterationRows) Err() error {
	return r.err
}

func TestGetSchemaRecordsFromRowsIterationError(t *testing.T) {
	iterationErr := errors.New("schema iteration interrupted")
	records, err := getSchemaRecordsFromRows(&schemaIterationRows{err: iterationErr})
	if !errors.Is(err, iterationErr) {
		t.Fatalf("got error %v, want %v", err, iterationErr)
	}
	if records != nil {
		t.Fatalf("got partial schema records after iteration error: %+v", records)
	}
}

func TestGetForeignSchemaNamesFromRows(t *testing.T) {
	rows := &fakeRows{values: [][]any{{"zeta"}, {constants.InternalSchema}, {"alpha"}, {constants.LegacyCommandSchema}}}
	names, err := getForeignSchemaNamesFromRows(rows)
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	if want := []string{"alpha", "zeta"}; !slices.Equal(names, want) {
		t.Fatalf("got %v, want %v", names, want)
	}
}

func TestGetForeignSchemaNamesFromRowsQueryError(t *testing.T) {
	queryErr := errors.New("canceling statement due to statement timeout")
	names, err := getForeignSchemaNamesFromRows(&fakeRows{values: [][]any{{"alpha"}}, err: queryErr})
	if !errors.Is(err, queryErr) {
		t.Fatalf("got error %v, want %v", err, queryErr)
	}
	if names != nil {
		t.Fatalf("got the schema names read before the error: %v", names)
	}
}

func TestGetForeignSchemaNamesFromRowsScanErrorClosesRows(t *testing.T) {
	rows := &fakeRows{values: [][]any{{"alpha"}, {42}, {"beta"}}}
	if _, err := getForeignSchemaNamesFromRows(rows); err == nil {
		t.Fatal("got no error for a row that cannot be scanned")
	}
	if !rows.closed {
		t.Fatal("the rows were left open after a scan error")
	}
}
