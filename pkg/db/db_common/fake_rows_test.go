package db_common

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// fakeRows is a pgx.Rows that returns values, then reports err from Err, the way pgx reports a query that
// fails while it runs (a statement timeout, a cancel, a lost connection)
type fakeRows struct {
	pgx.Rows
	values [][]any
	err    error
	next   int
	closed bool
}

func (r *fakeRows) Next() bool {
	if r.closed || r.next >= len(r.values) {
		// pgx closes the rows when Next returns false
		r.closed = true
		return false
	}
	r.next++
	return true
}

func (r *fakeRows) Scan(dest ...any) error {
	for i, v := range r.values[r.next-1] {
		switch d := dest[i].(type) {
		case *string:
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("cannot scan %T into a string", v)
			}
			*d = s
		default:
			return fmt.Errorf("fakeRows cannot scan into %T", dest[i])
		}
	}
	return nil
}

func (r *fakeRows) Err() error { return r.err }

func (r *fakeRows) Close() { r.closed = true }
