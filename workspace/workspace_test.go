package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type loadWorkspaceTest struct {
	source   string
	expected interface{}
}

var toStringPointer = utils.ToStringPointer

var m3alias = "m3"

// the actual mod loading logic is tested more thoroughly in TestLoadMod (steampipeconfig/load_mod_test.go)
// this test is primarily to verify the QueryMap building
var testCasesLoadWorkspace = map[string]loadWorkspaceTest{
	"single mod": {
		source: "test_data/w_1",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				Name:  toStringPointer("w_1"),
				Title: toStringPointer("workspace 1"),
				//ModDepends: []*modconfig.ModVersion{
				//	{Name: "github.com/turbot/m1", Version: "0.0.0"},
				//	{Name: "github.com/turbot/m2", Version: "0.0.0"},
				//},
				Queries: map[string]*modconfig.Query{
					"localq1": {
						Name: toStringPointer("localq1"), Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
					},
					"localq2": {
						Name: toStringPointer("localq2"), Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
					},
				},
			},
			QueryMap: map[string]*modconfig.Query{
				"w_1.query.localq1": {
					Name: toStringPointer("localq1"), Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
				},
				"query.localq1": {
					Name: toStringPointer("localq1"), Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
				},
				"w_2.query.localq2": {
					Name: toStringPointer("localq2"), Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
				},
				"query.localq2": {
					Name: toStringPointer("localq2"), Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
				},
				"m1.query.q1": {
					toStringPointer("q1"), toStringPointer("Q1"), toStringPointer("THIS IS QUERY 1"), toStringPointer("select 1"),
				},
				"m2.query.q2": {
					toStringPointer("q2"), toStringPointer("Q2"), toStringPointer("THIS IS QUERY 2"), toStringPointer("select 2"),
				},
			},
		},
	},
	"single_mod_with_ignored_directory": {
		source: "test_data/single_mod_with_ignored_directory",
		expected: &Workspace{Mod: &modconfig.Mod{
			Name:        toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					toStringPointer("q1"), toStringPointer("Q1"), toStringPointer("THIS IS QUERY 1"), toStringPointer("select 1"),
				},
				"q2": {
					toStringPointer("q2"), toStringPointer("Q2"), toStringPointer("THIS IS QUERY 2"), toStringPointer("select 2"),
				},
			},
		},
		},
	},
	"single_mod_with_ignored_sql_files": {
		source: "test_data/single_mod_with_ignored_sql_files",
		expected: &Workspace{Mod: &modconfig.Mod{
			Name:        toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					toStringPointer("q1"), toStringPointer("Q1"), toStringPointer("THIS IS QUERY 1"), toStringPointer("select 1"),
				},
			},
		}},
	},
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
