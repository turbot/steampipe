package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type loadWorkspaceTest struct {
	source   string
	expected interface{}
}

var m3alias = "m3"

// the actual mod loading logic is tested more thoroughly in TestLoadMod (steampipeconfig/load_mod_test.go)
// this test is primarily to verify the namedQueryMap building
var testCasesLoadWorkspace = map[string]loadWorkspaceTest{
	"single mod": {
		source: "test_data/w_1",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				Name:  "w_1",
				Title: "workspace 1",
				ModDepends: []*modconfig.ModVersion{
					{Name: "github.com/turbot/m1", Version: "0.0.0"},
					{Name: "github.com/turbot/m2", Version: "0.0.0"},
				},
				Queries: []*modconfig.Query{
					{
						"localq1", "LocalQ1", "THIS IS LOCAL QUERY 1", ".tables",
					},
					{
						"localq2", "LocalQ2", "THIS IS LOCAL QUERY 2", ".inspect",
					},
				},
			},
			namedQueryMap: map[string]*modconfig.Query{
				"w_1.query.localq1": {
					"localq1", "LocalQ1", "THIS IS LOCAL QUERY 1", ".tables",
				},
				"query.localq1": {
					"localq1", "LocalQ1", "THIS IS LOCAL QUERY 1", ".tables",
				},
				"w_2.query.localq2": {
					"localq2", "LocalQ2", "THIS IS LOCAL QUERY 2", ".inspect",
				},
				"query.localq2": {
					"localq2", "LocalQ2", "THIS IS LOCAL QUERY 2", ".inspect",
				},
				"m1.query.q1": {
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
				"m2.query.q2": {
					"q2", "Q2", "THIS IS QUERY 2", "select 2",
				},
			},
		},
	},
	//"single_mod_with_ignored_directory": {
	//	source: "test_data/single_mod_with_ignored_directory",
	//	expected: &Workspace{Mod: &modconfig.Mod{
	//		Name:        "m1",
	//		Title:       "M1",
	//		Description: "THIS IS M1",
	//		Queries: []*modconfig.Query{
	//			{
	//				"q1", "Q1", "THIS IS QUERY 1", "select 1",
	//			},
	//			{
	//				"q2", "Q2", "THIS IS QUERY 2", "select 2",
	//			},
	//		},
	//	},
	//	},
	//},
	//"single_mod_with_ignored_sql_files": {
	//	source: "test_data/single_mod_with_ignored_sql_files",
	//	expected: &Workspace{Mod: &modconfig.Mod{
	//		Name:        "m1",
	//		Title:       "M1",
	//		Description: "THIS IS M1",
	//		Queries: []*modconfig.Query{
	//			{
	//				"q1", "Q1", "THIS IS QUERY 1", "select 1",
	//			},
	//		},
	//	}},
	//},
}

func TestLoadWorkspace(t *testing.T) {
	for name, test := range testCasesLoadWorkspace {

		workspacePath, err := filepath.Abs(test.source)
		workspace, err := Load(workspacePath)

		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			continue
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}

		if match, message := WorkspacesEqual(test.expected.(*Workspace), workspace); !match {
			t.Errorf("Test: '%s'' FAILED : %s", name, message)
		}
	}
}

func WorkspacesEqual(expected, actual *Workspace) (bool, string) {

	errors := []string{}
	if actual.Mod.String() != expected.Mod.String() {
		errors = append(errors, fmt.Sprintf("workspace mods do not match - expected \n\n%s\n\nbut got\n\n%s\n", expected.Mod.String(), actual.Mod.String()))
	}

	for name, expectedQuery := range expected.GetNamedQueryMap() {
		actualQuery, ok := actual.GetNamedQueryMap()[name]
		if ok {
			if expectedQuery.String() != actualQuery.String() {
				errors = append(errors, fmt.Sprintf("query %s expected\n\n%s\n\n, got\na\n%s\n\n", name, expectedQuery.String(), actualQuery.String()))
			}
		} else {
			errors = append(errors, fmt.Sprintf("mod map missing expected key %s", name))
		}
	}
	for name, _ := range actual.GetNamedQueryMap() {
		if _, ok := expected.GetNamedQueryMap()[name]; ok {
			errors = append(errors, fmt.Sprintf("unexpected query %s in query map", name))
		}
	}
	return len(errors) > 0, strings.Join(errors, "\n")
}
