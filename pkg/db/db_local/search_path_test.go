package db_local

import (
	"errors"
	"slices"
	"testing"
)

func TestUserSearchPathQueries(t *testing.T) {
	rows := &fakeRows{values: [][]any{{"alice"}, {"root"}, {"bob"}}}
	queries, err := userSearchPathQueries(rows, []string{`"public"`, `"steampipe_internal"`})
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	want := []string{
		"LOCK TABLE pg_user IN SHARE ROW EXCLUSIVE MODE;",
		`ALTER USER "alice" SET SEARCH_PATH TO "public","steampipe_internal";`,
		`ALTER USER "bob" SET SEARCH_PATH TO "public","steampipe_internal";`,
	}
	if !slices.Equal(queries, want) {
		t.Fatalf("got %q, want %q", queries, want)
	}
}

func TestUserSearchPathQueriesQueryError(t *testing.T) {
	queryErr := errors.New("canceling statement due to statement timeout")
	queries, err := userSearchPathQueries(&fakeRows{values: [][]any{{"alice"}}, err: queryErr}, []string{`"public"`})
	if !errors.Is(err, queryErr) {
		t.Fatalf("got error %v, want %v", err, queryErr)
	}
	if queries != nil {
		t.Fatalf("got statements for the users read before the error: %q", queries)
	}
}

func TestUserSearchPathQueriesScanErrorClosesRows(t *testing.T) {
	rows := &fakeRows{values: [][]any{{"alice"}, {42}, {"bob"}}}
	if _, err := userSearchPathQueries(rows, []string{`"public"`}); err == nil {
		t.Fatal("got no error for a row that cannot be scanned")
	}
	if !rows.closed {
		t.Fatal("the rows were left open after a scan error")
	}
}
