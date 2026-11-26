package db_common

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
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
