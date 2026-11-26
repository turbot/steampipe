package db_local

import (
	"errors"
	"testing"
)

func TestCountClientsFromRows(t *testing.T) {
	rows := &fakeRows{values: [][]any{
		{"steampipe_client_query", 2},
		{"steampipe_client_system", 1},
		{"steampipe_service_plugin_manager", 3},
		{"psql", 4},
	}}
	counts, err := countClientsFromRows(rows)
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	if want := (ClientCount{TotalClients: 10, SteampipeClients: 2, PluginManagerClients: 3}); *counts != want {
		t.Fatalf("got %+v, want %+v", *counts, want)
	}
}

func TestCountClientsFromRowsQueryError(t *testing.T) {
	queryErr := errors.New("terminating connection due to administrator command")
	counts, err := countClientsFromRows(&fakeRows{values: [][]any{{"steampipe_client_query", 2}}, err: queryErr})
	if !errors.Is(err, queryErr) {
		t.Fatalf("got error %v, want %v", err, queryErr)
	}
	if counts != nil {
		t.Fatalf("got the counts of the rows read before the error: %+v", *counts)
	}
}
