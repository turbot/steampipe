package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// file operation
type operation uint32

const (
	create operation = 1 << iota
	update
	delete
)

type watcherTest struct {
	name      string
	operation operation
	path      string
	data      string
	expected  map[string]*modconfig.Query
}

// NOTE - these test cases are cumulative - the expected is based on the previous expected plus the current operation
var testCasesWatcher = []watcherTest{
	{
		name:      "add root sql file",
		operation: create,
		path:      "q1.sql",
		data:      "select 1",
		expected:  map[string]*modconfig.Query{"query.q1": {Name: "q1", SQL: "select 1"}},
	},
	{name: "add nested sql file",
		operation: create,
		path:      "queries/q1.sql",
		data:      "select 1",
		expected: map[string]*modconfig.Query{
			"query.q1":         {Name: "q1", SQL: "select 1"},
			"query.queries_q1": {Name: "queries_q1", SQL: "select 1"},
		},
	},
	{name: "update nested sql file",
		operation: update,
		path:      "queries/q1.sql",
		data:      "select 2",
		expected: map[string]*modconfig.Query{
			"query.q1":         {Name: "q1", SQL: "select 1"},
			"query.queries_q1": {Name: "queries_q1", SQL: "select 2"},
		},
	},
	{
		name:      "add deeply nested sql file",
		operation: create,
		path:      "queries/a/b/c/q10.sql",
		data:      "select 10",
		expected: map[string]*modconfig.Query{
			"query.q1":                 {Name: "q1", SQL: "select 1"},
			"query.queries_q1":         {Name: "queries_q1", SQL: "select 2"},
			"query.queries_a_b_c_q110": {Name: "queries_a_b_c_q110", SQL: "select 10"},
		},
	},
}

func TestWorkspaceFileWatcher(t *testing.T) {
	workspacePath, err := filepath.Abs(`test_data/watcher_test`)
	if err != nil {
		t.Fatalf("failed to build absolute config filepath from %s", workspacePath)
	}
	if err := os.RemoveAll(workspacePath); err != nil {
		t.Fatalf("failed to build initialise test directory")
	}
	if err := os.Mkdir(workspacePath, 0755); err != nil {
		t.Fatalf("failed to build initialise test directory")
	}

	os.Chdir(workspacePath)

	workspace, err := Load()
	if err != nil {
		t.Fatalf("failed to load workspace: %v", err)
	}
	queryMap := workspace.GetNamedQueryMap()
	if len(queryMap) != 0 {
		t.Fatalf("expected initial map to be empty but got %+v", queryMap)
	}

	for _, test := range testCasesWatcher {
		switch test.operation {
		case create, update:
			writeFile(test.path, test.data)

		case delete:
			deleteFile(test.path)

		}
		// now check the result
		queryMap = workspace.GetNamedQueryMap()
		if queryMapsEqual(queryMap, test.expected) {
			fmt.Printf("'%s' passed\n", test.name)
		} else {
			t.Fatalf("test '%s' failed: expected \n\n%+v\n\n got: \n\n%+v\n\n", test.name, test.expected, queryMap)
		}
	}
}

func writeFile(path, content string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return
		}

		_, err = f.WriteString(content)

		// wait for watcher to get event
		time.Sleep(500 * time.Millisecond)

		f.Close()
		wg.Done()
	}()
	wg.Wait()
}

func deleteFile(path string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		os.Remove(path)

		// wait for watcher to get event
		time.Sleep(100 * time.Millisecond)

		wg.Done()
	}()
	wg.Wait()
}

func queryMapsEqual(l, r map[string]*modconfig.Query) bool {
	if len(l) != len(r) {
		return false
	}

	for name, lquery := range l {
		rquery, ok := r[name]
		if !ok {
			return false
		}
		if !lquery.Equals(rquery) {
			return false
		}
	}
	return true
}
